/*
Copyright (c) 2026 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logforwarder

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type LogForwarderResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
}

var _ resource.Resource = &LogForwarderResource{}
var _ resource.ResourceWithConfigure = &LogForwarderResource{}
var _ resource.ResourceWithImportState = &LogForwarderResource{}

func New() resource.Resource {
	return &LogForwarderResource{}
}

func (r *LogForwarderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_forwarder"
}

func (r *LogForwarderResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&logForwarderConfigValidator{},
	}
}

type logForwarderConfigValidator struct{}

func (v *logForwarderConfigValidator) Description(_ context.Context) string {
	return "At least one of 'applications' or 'groups' must be specified and non-empty"
}

func (v *logForwarderConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *logForwarderConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config LogForwarder
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation if values are unknown (e.g., during plan)
	if config.Applications.IsUnknown() || config.Groups.IsUnknown() {
		return
	}

	hasApplications := !config.Applications.IsNull() && len(config.Applications.Elements()) > 0
	hasGroups := !config.Groups.IsNull() && len(config.Groups.Elements()) > 0

	if !hasApplications && !hasGroups {
		resp.Diagnostics.AddError(
			"Missing required configuration",
			"At least one of 'applications' or 'groups' must be specified with non-empty values",
		)
	}
}

func (r *LogForwarderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages log forwarder configuration for a cluster",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
				},
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the log forwarder.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"s3": schema.SingleNestedAttribute{
				Description: "S3 configuration for log forwarding destination.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"bucket_name": schema.StringAttribute{
						Description: "The name of the S3 bucket.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"bucket_prefix": schema.StringAttribute{
						Description: "The prefix to use for objects stored in the S3 bucket.",
						Optional:    true,
						Computed:    true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.MatchRoot("s3"),
						path.MatchRoot("cloudwatch"),
					),
				},
			},
			"cloudwatch": schema.SingleNestedAttribute{
				Description: "CloudWatch configuration for log forwarding destination.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"log_group_name": schema.StringAttribute{
						Description: "The name of the CloudWatch log group.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"log_distribution_role_arn": schema.StringAttribute{
						Description: "The ARN of the IAM role for log distribution.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.MatchRoot("s3"),
						path.MatchRoot("cloudwatch"),
					),
				},
			},
			"applications": schema.ListAttribute{
				Description: "List of additional applications to forward logs for.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"groups": schema.ListNestedAttribute{
				Description: "List of log forwarder groups.",
				Optional:    true,
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The identifier of the log forwarder group.",
							Required:    true,
						},
						"version": schema.StringAttribute{
							Description: "The version of the log forwarder group.",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *LogForwarderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.collection = connection.ClustersMgmt().V1().Clusters()
	r.clusterWait = common.NewClusterWait(r.collection, connection)
}

func (r *LogForwarderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &LogForwarder{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := plan.Cluster.ValueString()

	waitTimeoutInMinutes := int64(60)
	_, err := r.clusterWait.WaitForClusterToBeReady(ctx, clusterId, waitTimeoutInMinutes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot poll cluster state",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				clusterId, err,
			),
		)
		return
	}

	logForwarder, err := r.buildLogForwarderFromState(ctx, plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to build log forwarder",
			fmt.Sprintf("Failed to build log forwarder for cluster '%s': %v", clusterId, err),
		)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating log forwarder", map[string]interface{}{
		"cluster": clusterId,
	})

	logForwardersClient := r.collection.Cluster(clusterId).ControlPlane().LogForwarders()
	createResp, err := logForwardersClient.Add().Body(logForwarder).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create log forwarder",
			fmt.Sprintf(
				"Failed to create log forwarder for cluster '%s': %v",
				clusterId, err,
			),
		)
		return
	}

	err = r.populateState(ctx, createResp.Body(), plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to populate log forwarder state",
			fmt.Sprintf("Failed to populate log forwarder state: %v", err),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LogForwarderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &LogForwarder{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := state.Cluster.ValueString()
	logForwarderId := state.ID.ValueString()

	tflog.Debug(ctx, "Reading log forwarder", map[string]interface{}{
		"cluster":       clusterId,
		"log_forwarder": logForwarderId,
	})

	logForwardersClient := r.collection.Cluster(clusterId).ControlPlane().LogForwarders()
	getResp, err := logForwardersClient.LogForwarder(logForwarderId).Get().SendContext(ctx)
	if err != nil {
		if getResp != nil && getResp.Status() == http.StatusNotFound {
			tflog.Warn(ctx, "Log forwarder not found, removing from state", map[string]interface{}{
				"cluster":       clusterId,
				"log_forwarder": logForwarderId,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to fetch log forwarder",
			fmt.Sprintf(
				"Failed to fetch log forwarder '%s' for cluster '%s': %v",
				logForwarderId, clusterId, err,
			),
		)
		return
	}

	err = r.populateState(ctx, getResp.Body(), state, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to populate log forwarder state",
			fmt.Sprintf("Failed to populate log forwarder state: %v", err),
		)
		return
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *LogForwarderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	state := &LogForwarder{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan := &LogForwarder{}
	diags = req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	common.ValidateStateAndPlanEquals(state.Cluster, plan.Cluster, "cluster", &diags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := plan.Cluster.ValueString()
	logForwarderId := state.ID.ValueString()

	logForwarder, err := r.buildLogForwarderFromState(ctx, plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to build log forwarder",
			fmt.Sprintf("Failed to build log forwarder update for cluster '%s': %v", clusterId, err),
		)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating log forwarder", map[string]interface{}{
		"cluster":       clusterId,
		"log_forwarder": logForwarderId,
	})

	logForwardersClient := r.collection.Cluster(clusterId).ControlPlane().LogForwarders()
	updateResp, err := logForwardersClient.LogForwarder(logForwarderId).Update().Body(logForwarder).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update log forwarder",
			fmt.Sprintf(
				"Failed to update log forwarder '%s' for cluster '%s': %v",
				logForwarderId, clusterId, err,
			),
		)
		return
	}

	err = r.populateState(ctx, updateResp.Body(), plan, &resp.Diagnostics)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to populate log forwarder state",
			fmt.Sprintf("Failed to populate log forwarder state: %v", err),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LogForwarderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &LogForwarder{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := state.Cluster.ValueString()
	logForwarderId := state.ID.ValueString()

	tflog.Debug(ctx, "Deleting log forwarder", map[string]interface{}{
		"cluster":       clusterId,
		"log_forwarder": logForwarderId,
	})

	logForwardersClient := r.collection.Cluster(clusterId).ControlPlane().LogForwarders()
	_, err := logForwardersClient.LogForwarder(logForwarderId).Delete().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete log forwarder",
			fmt.Sprintf(
				"Failed to delete log forwarder '%s' for cluster '%s': %v",
				logForwarderId, clusterId, err,
			),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *LogForwarderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "begin importstate()")

	parts := strings.Split(req.ID, ",")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Expected import identifier with format: cluster_id,log_forwarder_id",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// buildLogForwarderFromState builds an OCM LogForwarder object from the Terraform state
func (r *LogForwarderResource) buildLogForwarderFromState(ctx context.Context, state *LogForwarder, diags *diag.Diagnostics) (*cmv1.LogForwarder, error) {
	builder := cmv1.NewLogForwarder()

	if !state.S3.IsNull() && !state.S3.IsUnknown() {
		s3Config := &S3Config{}
		d := state.S3.As(ctx, s3Config, basetypes.ObjectAsOptions{})
		if d.HasError() {
			diags.Append(d...)
			return nil, fmt.Errorf("failed to parse S3 configuration")
		}

		s3Builder := cmv1.NewLogForwarderS3Config()
		s3Builder.BucketName(s3Config.BucketName.ValueString())
		if !s3Config.BucketPrefix.IsNull() && !s3Config.BucketPrefix.IsUnknown() {
			s3Builder.BucketPrefix(s3Config.BucketPrefix.ValueString())
		}
		builder.S3(s3Builder)
	}

	if !state.CloudWatch.IsNull() && !state.CloudWatch.IsUnknown() {
		cloudWatchConfig := &CloudWatchConfig{}
		d := state.CloudWatch.As(ctx, cloudWatchConfig, basetypes.ObjectAsOptions{})
		if d.HasError() {
			diags.Append(d...)
			return nil, fmt.Errorf("failed to parse CloudWatch configuration")
		}

		cloudWatchBuilder := cmv1.NewLogForwarderCloudWatchConfig()
		cloudWatchBuilder.LogGroupName(cloudWatchConfig.LogGroupName.ValueString())
		cloudWatchBuilder.LogDistributionRoleArn(cloudWatchConfig.LogDistributionRoleArn.ValueString())
		builder.Cloudwatch(cloudWatchBuilder)
	}

	if !state.Applications.IsNull() && !state.Applications.IsUnknown() {
		applications := common.OptionalList(state.Applications)
		builder.Applications(applications...)
	}

	if !state.Groups.IsNull() && !state.Groups.IsUnknown() {
		var groupsList []LogForwarderGroup
		d := state.Groups.ElementsAs(ctx, &groupsList, false)
		if d.HasError() {
			diags.Append(d...)
			return nil, fmt.Errorf("failed to parse groups")
		}

		groupBuilders := make([]*cmv1.LogForwarderGroupBuilder, 0, len(groupsList))
		for _, group := range groupsList {
			groupBuilder := cmv1.NewLogForwarderGroup()
			groupBuilder.ID(group.ID.ValueString())
			if !group.Version.IsNull() && !group.Version.IsUnknown() {
				groupBuilder.Version(group.Version.ValueString())
			}
			groupBuilders = append(groupBuilders, groupBuilder)
		}
		builder.Groups(groupBuilders...)
	}

	logForwarder, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build log forwarder: %v", err)
	}

	return logForwarder, nil
}

// populateState populates the Terraform state from an OCM LogForwarder object
func (r *LogForwarderResource) populateState(ctx context.Context, logForwarder *cmv1.LogForwarder, state *LogForwarder, diags *diag.Diagnostics) error {
	state.ID = types.StringValue(logForwarder.ID())

	if s3Config, ok := logForwarder.GetS3(); ok {
		s3State := &S3Config{
			BucketName: types.StringValue(s3Config.BucketName()),
		}
		if bucketPrefix, ok := s3Config.GetBucketPrefix(); ok {
			s3State.BucketPrefix = types.StringValue(bucketPrefix)
		} else {
			s3State.BucketPrefix = types.StringValue("")
		}

		s3Object, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"bucket_name":   types.StringType,
			"bucket_prefix": types.StringType,
		}, s3State)
		if d.HasError() {
			diags.Append(d...)
			return fmt.Errorf("failed to convert S3 config to object")
		}
		state.S3 = s3Object
	} else {
		state.S3 = types.ObjectNull(map[string]attr.Type{
			"bucket_name":   types.StringType,
			"bucket_prefix": types.StringType,
		})
	}

	if cloudWatchConfig, ok := logForwarder.GetCloudwatch(); ok {
		cloudWatchState := &CloudWatchConfig{
			LogGroupName:           types.StringValue(cloudWatchConfig.LogGroupName()),
			LogDistributionRoleArn: types.StringValue(cloudWatchConfig.LogDistributionRoleArn()),
		}

		cloudWatchObject, d := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"log_group_name":            types.StringType,
			"log_distribution_role_arn": types.StringType,
		}, cloudWatchState)
		if d.HasError() {
			diags.Append(d...)
			return fmt.Errorf("failed to convert CloudWatch config to object")
		}
		state.CloudWatch = cloudWatchObject
	} else {
		state.CloudWatch = types.ObjectNull(map[string]attr.Type{
			"log_group_name":            types.StringType,
			"log_distribution_role_arn": types.StringType,
		})
	}

	if applications, ok := logForwarder.GetApplications(); ok && len(applications) > 0 {
		appList, err := common.StringArrayToList(applications)
		if err != nil {
			return fmt.Errorf("failed to convert applications to list: %v", err)
		}
		state.Applications = appList
	} else {
		state.Applications = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if groups, ok := logForwarder.GetGroups(); ok && len(groups) > 0 {
		groupsList := make([]LogForwarderGroup, 0, len(groups))
		for _, group := range groups {
			groupState := LogForwarderGroup{
				ID: types.StringValue(group.ID()),
			}
			if version, ok := group.GetVersion(); ok {
				groupState.Version = types.StringValue(version)
			} else {
				groupState.Version = types.StringNull()
			}
			groupsList = append(groupsList, groupState)
		}

		groupsListValue, d := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":      types.StringType,
				"version": types.StringType,
			},
		}, groupsList)
		if d.HasError() {
			diags.Append(d...)
			return fmt.Errorf("failed to convert groups to list")
		}
		state.Groups = groupsListValue
	} else {
		state.Groups = types.ListValueMust(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":      types.StringType,
				"version": types.StringType,
			},
		}, []attr.Value{})
	}

	return nil
}
