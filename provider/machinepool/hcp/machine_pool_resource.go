/*
Copyright (c) 2024 Red Hat, Inc.

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

package hcp

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	diskValidator "github.com/openshift-online/ocm-common/pkg/machinepool/validations"
	ocmUtils "github.com/openshift-online/ocm-common/pkg/ocm/utils"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	rosa "github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/machinepool/hcp/upgrade"
)

var nodePoolNameRE = regexp.MustCompile(
	`^[a-z]([-a-z0-9]*[a-z0-9])?$`,
)

// This is a magic name to trigger special handling for the cluster's default
// machine pool
// TODO: this should be in ocm-common repo
var standardNodePoolRegex = regexp.MustCompile(
	"^workers?(-[0-9]+)?$",
)

type HcpMachinePoolResource struct {
	clusterCollection *cmv1.ClustersClient
	versionCollection *cmv1.VersionsClient
	clusterWait       common.ClusterWait
}

var _ resource.ResourceWithConfigure = &HcpMachinePoolResource{}
var _ resource.ResourceWithImportState = &HcpMachinePoolResource{}
var _ resource.ResourceWithConfigValidators = &HcpMachinePoolResource{}

func New() resource.Resource {
	return &HcpMachinePoolResource{}
}

func (r *HcpMachinePoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hcp_machine_pool"
}

func (r *HcpMachinePoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Machine pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the machine pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine pool. Must consist of lower-case alphanumeric characters or '-', start and end with an alphanumeric character. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "name may not be empty/blank string"),
				},
			},
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "cluster ID may not be empty/blank string"),
				},
			},
			"replicas": schema.Int64Attribute{
				Description: "The number of machines in the pool. " +
					"Single zone clusters need at least 2 nodes, " +
					"multizone clusters need at least 3 nodes. " +
					"The maximum is 250 for cluster versions prior to 4.14.0-0.a, " +
					"and 500 for cluster versions 4.14.0-0.a and later.",
				Optional: true,
			},
			"autoscaling": schema.SingleNestedAttribute{
				Description: "Basic autoscaling options",
				Attributes:  AutoscalingResource(),
				Required:    true,
			},
			"taints": schema.ListNestedAttribute{
				Description: "Taints for a machine pool. Format should be a comma-separated " +
					"list of 'key=value'. This list will overwrite any modifications " +
					"made to node taints on an ongoing basis.\n",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Taints key",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Taints value",
							Required:    true,
						},
						"schedule_type": schema.StringAttribute{
							Description: "Taints schedule type",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("NoSchedule", "PreferNoSchedule", "NoExecute"),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				Optional: true,
			},
			"labels": schema.MapAttribute{
				Description: "Labels for the machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
				},
			},
			"availability_zone": schema.StringAttribute{
				Description: "Select the availability zone in which to create a single AZ machine pool for a multi-AZ cluster. " + common.ValueCannotBeChangedStringDescription,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Description: "Select the subnet in which to create a single AZ machine pool for BYO-VPC cluster. " + common.ValueCannotBeChangedStringDescription,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`.*\S.*`), "subnet ID may not be empty/blank string"),
				},
			},
			"status": schema.SingleNestedAttribute{
				Description: "HCP replica status",
				Attributes:  NodePoolStatusResource(),
				Computed:    true,
				Default:     nil,
			},
			"aws_node_pool": schema.SingleNestedAttribute{
				Description: "AWS settings for node pool",
				Attributes:  AwsNodePoolResource(),
				Required:    true,
			},
			"tuning_configs": schema.ListAttribute{
				Description: "A list of tuning configs attached to the pool.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"kubelet_configs": schema.StringAttribute{
				Description: "Name of the kubelet config applied to the machine pool. A single kubelet config is allowed. Kubelet config must already exist.",
				Optional:    true,
			},
			"auto_repair": schema.BoolAttribute{
				Description: "Indicates use of autor repair for the pool",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "Desired version of OpenShift for the machine pool, for example '4.11.0'. If version is greater than the currently running version, an upgrade will be scheduled.",
				Optional:    true,
			},
			"current_version": schema.StringAttribute{
				Description: "The currently running version of OpenShift on the machine pool, for example '4.11.0'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"upgrade_acknowledgements_for": schema.StringAttribute{
				Description: "Indicates acknowledgement of agreements required to upgrade the cluster version between" +
					" minor versions (e.g. a value of \"4.12\" indicates acknowledgement of any agreements required to " +
					"upgrade to OpenShift 4.12.z from 4.11 or before).",
				Optional: true,
			},
			"ignore_deletion_error": schema.BoolAttribute{
				Description: "Indicates to the provider to disregard API errors when deleting the machine pool." +
					" This will remove the resource from the management file, but not necessirely delete the underlying pool in case it errors." +
					" Setting this to true can bypass issues when destroying the cluster resource alongside the pool resource in the same management file." +
					" This is not recommended to be set in other use cases",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
		},
	}
}

func (r *HcpMachinePoolResource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(path.MatchRoot("autoscaling").AtName("min_replicas"), path.MatchRoot("autoscaling").AtName("max_replicas")),
		resourcevalidator.Conflicting(path.MatchRoot("replicas"), path.MatchRoot("autoscaling").AtName("min_replicas")),
		resourcevalidator.Conflicting(path.MatchRoot("replicas"), path.MatchRoot("autoscaling").AtName("max_replicas")),
		resourcevalidator.Conflicting(path.MatchRoot("aws_node_pool").AtName("capacity_reservation_id"), path.MatchRoot("autoscaling").AtName("min_replicas")),
		resourcevalidator.Conflicting(path.MatchRoot("aws_node_pool").AtName("capacity_reservation_id"), path.MatchRoot("autoscaling").AtName("max_replicas")),
	}
}

func (r *HcpMachinePoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.clusterCollection = connection.ClustersMgmt().V1().Clusters()
	r.versionCollection = connection.ClustersMgmt().V1().Versions()
	r.clusterWait = common.NewClusterWait(r.clusterCollection, connection)
}

func (r *HcpMachinePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get the plan:
	plan := &HcpMachinePoolState{}
	diags := req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	nodePoolName := plan.Name.ValueString()
	if !nodePoolNameRE.MatchString(nodePoolName) {
		resp.Diagnostics.AddError(
			"Cannot create machine pool: ",
			fmt.Sprintf("Cannot create machine pool for cluster '%s' with name '%s'. Expected a valid value for 'name' matching %s",
				plan.Cluster.ValueString(), plan.Name.ValueString(), nodePoolNameRE,
			),
		)
		return
	}

	// Wait till the cluster is ready:
	clusterObject, err := r.clusterWait.WaitForClusterToBeReady(ctx, plan.Cluster.ValueString(), 60)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot poll cluster state",
			fmt.Sprintf(
				"Cannot poll state of cluster with identifier '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	// The default machine pool is created automatically when the cluster is created.
	// We want to import it instead of creating it.
	if standardNodePoolRegex.MatchString(nodePoolName) {
		r.magicImport(ctx, plan, resp)
		return
	}

	// Create the machine pool:
	builder := cmv1.NewNodePool().ID(plan.ID.ValueString())
	builder.ID(plan.Name.ValueString())

	if plan.AWSNodePool != nil {
		awsNodePoolBuilder := cmv1.NewAWSNodePool()
		awsNodePoolBuilder.InstanceType(plan.AWSNodePool.InstanceType.ValueString())
		awsTags, err := common.OptionalMap(ctx, plan.AWSNodePool.Tags)
		if err != nil {
			return
		}
		if len(awsTags) > 0 {
			awsNodePoolBuilder.Tags(awsTags)
		}

		if common.HasValue(plan.AWSNodePool.AdditionalSecurityGroupIds) {
			additionalSecurityGroupIds, err := common.StringListToArray(ctx, plan.AWSNodePool.AdditionalSecurityGroupIds)
			if err != nil {
				resp.Diagnostics.AddError(
					"Cannot convert Additional Security Groups to slice",
					fmt.Sprintf(
						"Cannot convert Additional Security Groups to slice for cluster '%s: %v'", plan.Cluster.ValueString(), err,
					),
				)
				return
			}
			awsNodePoolBuilder.AdditionalSecurityGroupIds(additionalSecurityGroupIds...)
		}

		if common.IsStringAttributeUnknownOrEmpty(plan.AWSNodePool.Ec2MetadataHttpTokens) {
			plan.AWSNodePool.Ec2MetadataHttpTokens = types.StringValue(ec2.HttpTokensStateOptional)
		}
		awsNodePoolBuilder.Ec2MetadataHttpTokens(cmv1.Ec2MetadataHttpTokens(plan.AWSNodePool.Ec2MetadataHttpTokens.ValueString()))

		if workerDiskSize := common.OptionalInt64(plan.AWSNodePool.DiskSize); workerDiskSize != nil {
			err := diskValidator.ValidateNodePoolRootDiskSize(int(*workerDiskSize))
			if err != nil {
				resp.Diagnostics.AddError(
					"Cannot build machine pool",
					err.Error(),
				)
				return
			}
			awsNodePoolBuilder.RootVolume(cmv1.NewAWSVolume().Size(int(*workerDiskSize)))
		}

		if !common.IsStringAttributeUnknownOrEmpty(plan.AWSNodePool.CapacityReservationId) {
			capacityReservationBuilder := cmv1.NewAWSCapacityReservation()
			capacityReservationBuilder.Id(plan.AWSNodePool.CapacityReservationId.ValueString())
			awsNodePoolBuilder.CapacityReservation(capacityReservationBuilder)
		}

		builder.AWSNodePool(awsNodePoolBuilder)
	}

	if !common.IsStringAttributeUnknownOrEmpty(plan.SubnetID) {
		builder.Subnet(plan.SubnetID.ValueString())
	}

	autoscalingEnabled := false
	computeNodeEnabled := false
	autoscalingEnabled, errMsg := getAutoscaling(plan, builder)
	if errMsg != "" {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s, %s'", plan.Cluster.ValueString(), errMsg,
			),
		)
		return
	}

	if common.HasValue(plan.Replicas) {
		computeNodeEnabled = true
		builder.Replicas(int(plan.Replicas.ValueInt64()))
	}
	if !autoscalingEnabled && !computeNodeEnabled {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s', please provide a value for 'replicas' when 'autoscaling.enabled' is set to 'false'.",
				plan.Cluster.ValueString(),
			),
		)
		return
	}

	if autoscalingEnabled && computeNodeEnabled {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s', please do not provide a value for 'replicas' when 'autoscaling.enabled' is set to 'true'.",
				plan.Cluster.ValueString(),
			),
		)
	}

	if plan.Taints != nil && len(plan.Taints) > 0 {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range plan.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint().
				Key(taint.Key.ValueString()).
				Value(taint.Value.ValueString()).
				Effect(taint.ScheduleType.ValueString()))
		}
		builder.Taints(taintBuilders...)
	}

	if common.HasValue(plan.Labels) {
		labels := map[string]string{}
		for k, v := range plan.Labels.Elements() {
			labels[k] = v.(types.String).ValueString()
		}
		builder.Labels(labels)
	}

	if common.HasValue(plan.TuningConfigs) {
		tuningConfigs, err := common.StringListToArray(ctx, plan.TuningConfigs)
		if err != nil {
			resp.Diagnostics.AddError(
				"Cannot build machine pool",
				fmt.Sprintf(
					"Cannot build tuning configs for machine pool pool for cluster '%s': %v",
					plan.Cluster.ValueString(), err,
				),
			)
			return
		}
		if tuningConfigs != nil {
			builder.TuningConfigs(tuningConfigs...)
		}
	}

	if !common.IsStringAttributeUnknownOrEmpty(plan.KubeletConfigs) {
		builder.KubeletConfigs(plan.KubeletConfigs.ValueString())
	}

	if common.HasValue(plan.AutoRepair) {
		builder.AutoRepair(common.BoolWithTrueDefault(plan.AutoRepair))
	}

	if common.HasValue(plan.Version) {
		vBuilder := cmv1.NewVersion()
		vBuilder.ID(ocmUtils.CreateVersionId(plan.Version.ValueString(), clusterObject.Version().ChannelGroup()))
		vBuilder.ChannelGroup(clusterObject.Version().ChannelGroup())
		builder.Version(vBuilder)
	}

	object, err := builder.Build()
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot build machine pool",
			fmt.Sprintf(
				"Cannot build machine pool for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}

	collection := r.clusterCollection.Cluster(clusterObject.ID()).NodePools()
	add, err := collection.Add().Body(object).
		Parameter("fetchUserTagsOnly", true).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cannot create machine pool",
			fmt.Sprintf(
				"Cannot create machine pool for cluster '%s': %v",
				plan.Cluster.ValueString(), err,
			),
		)
		return
	}
	object = add.Body()

	// Save the state:
	err = populateState(ctx, object, plan, clusterObject)
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't populate machine pool state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// This handles the "magic" import of the default machine pool, allowing the
// user to include it in their config w/o having to specifically `terraform
// import` it.
func (r *HcpMachinePoolResource) magicImport(ctx context.Context, plan *HcpMachinePoolState, resp *resource.CreateResponse) {
	nodePoolName := plan.Name.ValueString()
	state := &HcpMachinePoolState{
		ID:      types.StringValue(nodePoolName),
		Cluster: plan.Cluster,
		Name:    types.StringValue(nodePoolName),
	}
	plan.ID = types.StringValue(nodePoolName)
	adjustInitialStateToPlan(state, plan)

	notFound, diags := readState(ctx, state, r.clusterCollection)
	if notFound {
		// We disallow creating a machine pool with the default name. This
		// case can only happen if the default machine pool was deleted and
		// the user tries to recreate it.
		diags.AddError(
			"Can't create machine pool",
			fmt.Sprintf(
				"Can't create machine pool for cluster '%s': "+
					"the default machine pool '%s' was deleted and a new machine pool with that name may not be created. "+
					"Please use a different name.",
				plan.Cluster.ValueString(),
				nodePoolName,
			),
		)
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = r.doUpdate(ctx, state, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *HcpMachinePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get the current state:
	state := &HcpMachinePoolState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notFound, diags := readState(ctx, state, r.clusterCollection)
	if notFound {
		// If we can't find the machine pool, it was deleted. Remove if from the
		// state and don't return an error so the TF apply() will automatically
		// recreate it.
		tflog.Warn(ctx, fmt.Sprintf("machine pool (%s) of cluster (%s) not found, removing from state",
			state.ID.ValueString(), state.Cluster.ValueString(),
		))
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func fetchCluster(ctx context.Context, state *HcpMachinePoolState, collection *cmv1.ClustersClient, diags *diag.Diagnostics) *cmv1.Cluster {
	clusterResource := collection.Cluster(state.Cluster.ValueString())
	getCluster, err := clusterResource.Get().SendContext(ctx)
	if err != nil {
		if getCluster.Status() == http.StatusNotFound {
			tflog.Warn(ctx, fmt.Sprintf("cluster '%s' not found, clearing state",
				state.Cluster.ValueString(),
			))
			return nil
		}
		diags.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.Cluster.ValueString(), err,
			),
		)
		return nil
	}
	return getCluster.Body()
}

func readState(ctx context.Context, state *HcpMachinePoolState, collection *cmv1.ClustersClient) (poolNotFound bool, diags diag.Diagnostics) {
	diags = diag.Diagnostics{}

	clusterObject := fetchCluster(ctx, state, collection, &diags)
	if clusterObject == nil {
		return
	}

	nodePoolResource := collection.Cluster(state.Cluster.ValueString()).
		NodePools().
		NodePool(state.ID.ValueString())
	getNp, err := nodePoolResource.Get().
		Parameter("fetchUserTagsOnly", true).SendContext(ctx)
	if err != nil {
		if getNp.Status() == http.StatusNotFound {
			poolNotFound = true
			return
		}
		diags.AddError(
			"Failed to fetch machine pool",
			fmt.Sprintf(
				"Failed to fetch machine pool with identifier %s for cluster %s. Response code: %v, err: %v",
				state.ID.ValueString(), state.Cluster.ValueString(), getNp.Status(), err,
			),
		)
		return
	}

	npObject := getNp.Body()
	err = populateState(ctx, npObject, state, clusterObject)
	if err != nil {
		diags.AddError(
			"Can't populate machine pool state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}
	return
}

func validateNoImmutableAttChange(state, plan *HcpMachinePoolState) diag.Diagnostics {
	diags := diag.Diagnostics{}
	validateStateAndPlanEquals(state.Cluster, plan.Cluster, "cluster", &diags)
	validateStateAndPlanEquals(state.Name, plan.Name, "name", &diags)
	validateStateAndPlanEquals(state.SubnetID, plan.SubnetID, "aws_node_pool.subnet_id", &diags)
	if state.AWSNodePool != nil && plan.AWSNodePool != nil {
		validateStateAndPlanEquals(state.AWSNodePool.InstanceProfile, plan.AWSNodePool.InstanceProfile, "aws_node_pool.instance_profile", &diags)
		validateStateAndPlanEquals(state.AWSNodePool.Tags, plan.AWSNodePool.Tags, "aws_node_pool.tags", &diags)
		validateStateAndPlanEquals(state.AWSNodePool.InstanceType, plan.AWSNodePool.InstanceType,
			"aws_node_pool.instance_type", &diags)
		validateStateAndPlanEquals(state.AWSNodePool.AdditionalSecurityGroupIds, plan.AWSNodePool.AdditionalSecurityGroupIds,
			"aws_node_pool.additional_security_group_ids", &diags)
		validateStateAndPlanEquals(state.AWSNodePool.Ec2MetadataHttpTokens, plan.AWSNodePool.Ec2MetadataHttpTokens,
			"aws_node_pool.ec2_metadata_http_tokens", &diags)
		validateStateAndPlanEquals(state.AWSNodePool.DiskSize, plan.AWSNodePool.DiskSize, "aws_node_pool.disk_size", &diags)
		validateStateAndPlanEquals(state.AWSNodePool.CapacityReservationId, plan.AWSNodePool.CapacityReservationId, "aws_node_pool.capacity_reservation_id", &diags)
	}
	return diags
}

func validateStateAndPlanEquals(stateAttr attr.Value, planAttr attr.Value, attrName string, diags *diag.Diagnostics) {
	// Its possible to have here unknown attributes
	// Relevant only for optional computed attributes in resource create
	// Check this because this function also used in "magicImport" function
	if !common.HasValue(planAttr) {
		return
	}
	common.ValidateStateAndPlanEquals(stateAttr, planAttr, attrName, diags)
}

func (r *HcpMachinePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get the state:
	state := &HcpMachinePoolState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &HcpMachinePoolState{}
	diags = req.Plan.Get(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = r.doUpdate(ctx, state, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *HcpMachinePoolResource) doUpdate(ctx context.Context, state *HcpMachinePoolState, plan *HcpMachinePoolState) diag.Diagnostics {
	//assert no changes on specific attributes
	diags := validateNoImmutableAttChange(state, plan)
	if diags.HasError() {
		return diags
	}

	clusterObject := fetchCluster(ctx, plan, r.clusterCollection, &diags)
	if clusterObject == nil {
		return diags
	}

	resource := r.clusterCollection.Cluster(state.Cluster.ValueString()).
		NodePools().
		NodePool(state.ID.ValueString())
	_, err := resource.Get().SendContext(ctx)

	if err != nil {
		diags.AddError(
			"Cannot find machine pool",
			fmt.Sprintf(
				"Cannot find machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return diags
	}

	// Schedule a cluster upgrade if a newer version is requested
	if err := r.upgradeMachinePoolIfNeeded(ctx, state, plan); err != nil {
		diags.AddError(
			"Can't upgrade machine pool",
			fmt.Sprintf("Can't upgrade machine pool version with identifier: `%s`, %v", state.ID.ValueString(), err),
		)
		return diags
	}

	if !state.AWSNodePool.CapacityReservationId.IsNull() && reflect.DeepEqual(state.AWSNodePool.CapacityReservationId, plan.AWSNodePool.CapacityReservationId) {
		diags.AddError(
			"Can't update machine pool",
			fmt.Sprintf("Capacity Reservation ID cannot be modified after it's creation (old value: '%s', "+
				"new value: '%s')", state.AWSNodePool.CapacityReservationId, plan.AWSNodePool.CapacityReservationId),
		)
		return diags
	}

	npBuilder := cmv1.NewNodePool().ID(state.ID.ValueString())

	if state.AWSNodePool != nil && plan.AWSNodePool != nil {
		awsNodePoolBuilder := cmv1.NewAWSNodePool()
		// FIXME: even though we do not necessarily need to patch OCM CS API enforces this value cannot be empty
		awsNodePoolBuilder.InstanceType(plan.AWSNodePool.InstanceType.ValueString())
		npBuilder.AWSNodePool(awsNodePoolBuilder)
	}

	if patchSubnet, should := common.ShouldPatchString(state.SubnetID, plan.SubnetID); should {
		npBuilder.Subnet(patchSubnet)
	}

	computeNodesEnabled := false
	autoscalingEnabled := false

	if common.HasValue(plan.Replicas) {
		computeNodesEnabled = true
		npBuilder.Replicas(int(plan.Replicas.ValueInt64()))
	}

	autoscalingEnabled, errMsg := getAutoscaling(plan, npBuilder)
	if errMsg != "" {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s, %s ", state.Cluster.ValueString(), errMsg,
			),
		)
		return diags
	}

	if (autoscalingEnabled && computeNodesEnabled) || (!autoscalingEnabled && !computeNodesEnabled) {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s: either replicas should be set or autoscaling enabled", state.Cluster.ValueString(),
			),
		)
		return diags
	}

	if patchAutoRepair, ok := common.ShouldPatchBool(state.AutoRepair, plan.AutoRepair); ok {
		npBuilder.AutoRepair(patchAutoRepair)
	}

	patchLabels, shouldPatchLabels := common.ShouldPatchMap(state.Labels, plan.Labels)
	if shouldPatchLabels {
		labels := map[string]string{}
		for k, v := range patchLabels.Elements() {
			labels[k] = v.(types.String).ValueString()
		}
		npBuilder.Labels(labels)
	}

	if shouldPatchTaints(state.Taints, plan.Taints) {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range plan.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint().
				Key(taint.Key.ValueString()).
				Value(taint.Value.ValueString()).
				Effect(taint.ScheduleType.ValueString()))
		}
		npBuilder.Taints(taintBuilders...)
	}

	patchTuningConfigs, shouldPatchTuningConfigs := common.ShouldPatchList(state.TuningConfigs, plan.TuningConfigs)
	if shouldPatchTuningConfigs {
		tuningConfigs, err := common.StringListToArray(ctx, patchTuningConfigs)
		if err != nil {
			diags.AddError(
				"Cannot update machine pool",
				fmt.Sprintf(
					"Cannot update tuning configs for machine pool pool for cluster '%s': %v",
					plan.Cluster.ValueString(), err,
				),
			)
			return diags
		}
		if tuningConfigs != nil {
			npBuilder.TuningConfigs(tuningConfigs...)
		}
	}

	patchKubeletConfigs, shouldPatchKubeletConfigs := common.ShouldPatchString(state.KubeletConfigs, plan.KubeletConfigs)
	if common.IsStringAttributeUnknownOrEmpty(plan.KubeletConfigs) {
		npBuilder.KubeletConfigs([]string{}...)
	} else if shouldPatchKubeletConfigs {
		npBuilder.KubeletConfigs(patchKubeletConfigs)
	}

	nodePool, err := npBuilder.Build()
	if err != nil {
		diags.AddError(
			"Cannot update machine pool",
			fmt.Sprintf(
				"Cannot update machine pool for cluster '%s: %v ", state.Cluster.ValueString(), err,
			),
		)
		return diags
	}
	update, err := r.clusterCollection.Cluster(state.Cluster.ValueString()).
		NodePools().
		NodePool(state.ID.ValueString()).Update().
		Parameter("fetchUserTagsOnly", true).Body(nodePool).SendContext(ctx)
	if err != nil {
		diags.AddError(
			"Failed to update machine pool",
			fmt.Sprintf(
				"Failed to update machine pool '%s'  on cluster '%s': %v",
				state.ID.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return diags
	}

	object := update.Body()

	adjustInitialStateToPlan(state, plan)

	// Save the state:
	err = populateState(ctx, object, state, clusterObject)
	if err != nil {
		diags.AddError(
			"Can't populate machine pool state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return diags
	}
	return diags
}

func adjustInitialStateToPlan(state, plan *HcpMachinePoolState) {
	// update some values the plan value (important for nil and false cases)
	if state.AutoScaling == nil {
		state.AutoScaling = new(AutoScaling)
	}
	state.AutoScaling.Enabled = plan.AutoScaling.Enabled

	state.Replicas = plan.Replicas
	state.NodePoolStatus = plan.NodePoolStatus
	state.Version = plan.Version
	state.IgnoreDeletionError = plan.IgnoreDeletionError

	if state.AWSNodePool == nil {
		state.AWSNodePool = new(AWSNodePool)
	}
	if common.HasValue(plan.AWSNodePool.Tags) {
		state.AWSNodePool.Tags = plan.AWSNodePool.Tags
	}

	if common.HasValue(plan.TuningConfigs) {
		state.TuningConfigs = plan.TuningConfigs
	}

	if plan.KubeletConfigs.ValueString() == "" {
		state.KubeletConfigs = plan.KubeletConfigs
	}
}

// Upgrades the cluster if the desired (plan) version is greater than the
// current version
func (r *HcpMachinePoolResource) upgradeMachinePoolIfNeeded(ctx context.Context, state, plan *HcpMachinePoolState) error {
	if common.IsStringAttributeUnknownOrEmpty(plan.Version) || common.IsStringAttributeUnknownOrEmpty(state.CurrentVersion) {
		// No version information, nothing to do
		tflog.Debug(ctx, "Insufficient cluster version information to determine if upgrade should be performed.")
		return nil
	}

	tflog.Debug(ctx, "HCP Machine Pool versions",
		map[string]interface{}{
			"current_version": state.CurrentVersion.ValueString(),
			"plan-version":    plan.Version.ValueString(),
			"state-version":   state.Version.ValueString(),
		})

	// See if the user has changed the requested version for this run
	requestedVersionChanged := true
	if !common.IsStringAttributeUnknownOrEmpty(plan.Version) && !common.IsStringAttributeUnknownOrEmpty(state.Version) {
		if plan.Version.ValueString() == state.Version.ValueString() {
			requestedVersionChanged = false
		}
	}

	// Check the versions to see if we need to upgrade
	currentVersion, err := semver.NewVersion(state.CurrentVersion.ValueString())
	if err != nil {
		return fmt.Errorf("failed to parse current cluster version: %v", err)
	}
	desiredVersion, err := semver.NewVersion(plan.Version.ValueString())
	if err != nil {
		return fmt.Errorf("failed to parse desired cluster version: %v", err)
	}
	if currentVersion.GreaterThan(desiredVersion) {
		tflog.Debug(ctx, "No cluster version upgrade needed.")
		if requestedVersionChanged {
			// User changed the version they want, but actual is higher. We
			// don't support downgrades.
			return fmt.Errorf("cluster version is already above the requested version")
		}
		return nil
	}
	cancelingUpgradeOnly := desiredVersion.Equal(currentVersion)

	if !cancelingUpgradeOnly {
		if err = r.validateUpgrade(ctx, state, plan); err != nil {
			return err
		}
	}

	// Fetch existing upgrade policies
	upgrades, err := upgrade.GetScheduledUpgrades(ctx, r.clusterCollection,
		state.Cluster.ValueString(), state.ID.ValueString())
	if err != nil {
		return fmt.Errorf("failed to get upgrade policies: %v", err)
	}

	// Stop if an upgrade is already in progress
	correctUpgradePending, err := upgrade.CheckAndCancelUpgrades(
		ctx, r.clusterCollection, upgrades, desiredVersion)
	if err != nil {
		return err
	}

	// Schedule a new upgrade
	if !correctUpgradePending && !cancelingUpgradeOnly {
		ackString := plan.UpgradeAcksFor.ValueString()
		if err = scheduleUpgrade(ctx, r.clusterCollection,
			state.Cluster.ValueString(), state.ID.ValueString(), desiredVersion, ackString); err != nil {
			return err
		}
	}

	state.Version = plan.Version
	state.UpgradeAcksFor = plan.UpgradeAcksFor
	return nil
}

func (r *HcpMachinePoolResource) validateUpgrade(ctx context.Context, state, plan *HcpMachinePoolState) error {
	availableVersions, err := upgrade.GetAvailableUpgradeVersions(
		ctx, r.clusterCollection, r.versionCollection, state.Cluster.ValueString(), state.ID.ValueString())
	if err != nil {
		return fmt.Errorf("failed to get available upgrades: %v", err)
	}
	trimmedDesiredVersion := strings.TrimPrefix(plan.Version.ValueString(), rosa.VersionPrefix)
	desiredVersion, err := semver.NewVersion(trimmedDesiredVersion)
	if err != nil {
		return fmt.Errorf("failed to parse desired version: %v", err)
	}
	found := false
	for _, v := range availableVersions {
		sem, err := semver.NewVersion(v.RawID())
		if err != nil {
			return fmt.Errorf("failed to parse available upgrade version: %v", err)
		}
		if desiredVersion.Equal(sem) {
			found = true
			break
		}
	}
	if !found {
		avail := []string{}
		for _, v := range availableVersions {
			avail = append(avail, v.RawID())
		}
		return fmt.Errorf("desired version (%s) is not in the list of available upgrades (%v)", desiredVersion, avail)
	}

	return nil
}

// Ensure user has acked upgrade gates and schedule the upgrade
func scheduleUpgrade(ctx context.Context, client *cmv1.ClustersClient,
	clusterID string, machinePoolId string, desiredVersion *semver.Version, userAckString string) error {
	// Gate agreements are checked when the upgrade is scheduled, resulting
	// in an error return. ROSA cli does this by scheduling once w/ dryRun
	// to look for un-acked agreements.
	clusterClient := client.Cluster(clusterID)
	upgradePoliciesClient := clusterClient.NodePools().NodePool(machinePoolId).UpgradePolicies()
	gates, description, err := upgrade.CheckMissingAgreements(desiredVersion.String(), clusterID, upgradePoliciesClient)
	if err != nil {
		return fmt.Errorf("failed to check for missing upgrade agreements: %v", err)
	}
	// User ack is required if we have any non-STS-only gates
	userAckRequired := false
	for _, gate := range gates {
		if !gate.STSOnly() {
			userAckRequired = true
		}
	}
	targetMinorVersion := getOcmVersionMinor(desiredVersion.String())
	if userAckRequired && userAckString != targetMinorVersion { // User has not acknowledged mandatory gates, stop here.
		return fmt.Errorf("%s\nTo acknowledge these items, please add \"upgrade_acknowledgements_for = %s\""+
			" and re-apply the changes", description, targetMinorVersion)
	}

	// Ack all gates to OCM
	for _, gate := range gates {
		gateID := gate.ID()
		tflog.Debug(ctx, "Acknowledging version gate", map[string]interface{}{"gateID": gateID})
		gateAgreementsClient := clusterClient.GateAgreements()
		err := upgrade.AckVersionGate(gateAgreementsClient, gateID)
		if err != nil {
			return fmt.Errorf("failed to acknowledge version gate '%s' for cluster '%s': %v",
				gateID, clusterID, err)
		}
	}

	// Schedule an upgrade
	tenMinFromNow := time.Now().UTC().Add(10 * time.Minute)
	newPolicy, err := cmv1.NewNodePoolUpgradePolicy().
		ScheduleType(cmv1.ScheduleTypeManual).
		Version(desiredVersion.String()).
		NextRun(tenMinFromNow).
		Build()
	if err != nil {
		return fmt.Errorf("failed to create upgrade policy: %v", err)
	}
	_, err = upgradePoliciesClient.
		Add().
		Body(newPolicy).
		SendContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to schedule upgrade: %v", err)
	}
	return nil
}

// TODO: move to ocm commons
func getOcmVersionMinor(ver string) string {
	version, err := semver.NewVersion(ver)
	if err != nil {
		segments := strings.Split(ver, ".")
		return fmt.Sprintf("%s.%s", segments[0], segments[1])
	}
	segments := version.Segments()
	return fmt.Sprintf("%d.%d", segments[0], segments[1])
}

func getAutoscaling(state *HcpMachinePoolState, mpBuilder *cmv1.NodePoolBuilder) (
	autoscalingEnabled bool, errMsg string) {
	autoscalingEnabled = false
	if common.HasValue(state.AutoScaling.Enabled) &&
		state.AutoScaling.Enabled.ValueBool() {
		autoscalingEnabled = true

		autoscaling := cmv1.NewNodePoolAutoscaling()
		if common.HasValue(state.AutoScaling.MaxReplicas) {
			autoscaling.MaxReplica(int(state.AutoScaling.MaxReplicas.ValueInt64()))
		} else {
			return false, "when enabling autoscaling, should set value for maxReplicas"
		}
		if common.HasValue(state.AutoScaling.MinReplicas) {
			autoscaling.MinReplica(int(state.AutoScaling.MinReplicas.ValueInt64()))
		} else {
			return false, "when enabling autoscaling, should set value for minReplicas"
		}
		if !autoscaling.Empty() {
			mpBuilder.Autoscaling(autoscaling)
		}
	} else {
		if common.HasValue(state.AutoScaling.MaxReplicas) ||
			common.HasValue(state.AutoScaling.MinReplicas) {
			return false, "when disabling autoscaling, cannot set min_replicas and/or max_replicas"
		}
	}

	return autoscalingEnabled, ""
}

func (r *HcpMachinePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get the state:
	state := &HcpMachinePoolState{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Send the request to delete the machine pool:
	resource := r.clusterCollection.Cluster(state.Cluster.ValueString()).
		NodePools().
		NodePool(state.ID.ValueString())
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		if common.BoolWithFalseDefault(state.IgnoreDeletionError) {
			resp.Diagnostics.AddWarning(
				"Cannot delete machine pool",
				fmt.Sprintf(
					"An error occurred deleting the pool,"+
						" because ignore deletion error is set it will still be removed from the terraform state. Reason: %v",
					err,
				),
			)
			// Remove the state:
			resp.State.RemoveResource(ctx)
			return
		}
		// We can't delete the pool, see if it's the last one:
		numPools, err2 := r.countPools(ctx, state.Cluster.ValueString())
		if numPools == 1 && err2 == nil {
			// It's the last one, issue warning instead of error
			resp.Diagnostics.AddWarning(
				"Cannot delete machine pool",
				fmt.Sprintf(
					"Cannot delete the last machine pool for cluster '%s'. "+
						"ROSA HCP clusters must have at least one machine pool. "+
						"It is being removed from the Terraform state only. "+
						"To resume managing this machine pool, import it again. "+
						"It will be automatically deleted when the cluster is deleted.",
					state.Cluster.ValueString(),
				),
			)
			// Remove the state:
			resp.State.RemoveResource(ctx)
			return
		} else {
			// Wasn't the last one, return error
			resp.Diagnostics.AddError(
				"Cannot delete machine pool",
				fmt.Sprintf(
					"Cannot delete machine pool with identifier '%s' for "+
						"cluster '%s': %v",
					state.ID.ValueString(), state.Cluster.ValueString(), err,
				),
			)
			return
		}
	}

	// Everything went fine deleting the resource in OCM
	// Remove the state:
	resp.State.RemoveResource(ctx)
}

// countPools returns the number of machine pools in the given cluster
func (r *HcpMachinePoolResource) countPools(ctx context.Context, clusterID string) (int, error) {
	resource := r.clusterCollection.Cluster(clusterID).NodePools()
	resp, err := resource.List().SendContext(ctx)
	if err != nil {
		return 0, err
	}
	return resp.Size(), nil
}

func (r *HcpMachinePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// To import a machine pool, we need to know the cluster ID and the machine pool ID
	fields := strings.Split(req.ID, ",")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			"Machine pool to import should be specified as <cluster_id>,<machine_pool_id>",
		)
		return
	}
	clusterID := fields[0]
	nodePoolId := fields[1]
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), nodePoolId)...)
}

// populateState copies the data from the API object to the Terraform state.
func populateState(ctx context.Context, object *cmv1.NodePool, state *HcpMachinePoolState, cluster *cmv1.Cluster) error {
	state.ID = types.StringValue(object.ID())
	state.Name = types.StringValue(object.ID())

	if awsNodePool, ok := object.GetAWSNodePool(); ok {
		if state.AWSNodePool == nil {
			state.AWSNodePool = new(AWSNodePool)
		}
		if instanceType, ok := awsNodePool.GetInstanceType(); ok {
			state.AWSNodePool.InstanceType = types.StringValue(instanceType)
		}
		if instanceProfile, ok := awsNodePool.GetInstanceProfile(); ok {
			state.AWSNodePool.InstanceProfile = types.StringValue(instanceProfile)
		}
		if state.AWSNodePool.Tags.IsUnknown() || state.AWSNodePool.Tags.IsNull() {
			state.AWSNodePool.Tags = types.MapNull(types.StringType)
		}
		if awsTags, ok := awsNodePool.GetTags(); ok {
			filteredAwsTags, err := filterClusterTagsNotPresentInNpInput(ctx, state, cluster, awsTags)
			if err != nil {
				return err
			}
			if len(filteredAwsTags) > 0 {
				mapValue, err := common.ConvertStringMapToMapType(filteredAwsTags)
				if err != nil {
					return err
				}
				state.AWSNodePool.Tags = mapValue
			}
		}
		if additionalSecurityGroupIds, ok := awsNodePool.GetAdditionalSecurityGroupIds(); ok {
			additionalSecurityGroupsList, err := common.StringArrayToList(additionalSecurityGroupIds)
			if err != nil {
				return err
			}
			state.AWSNodePool.AdditionalSecurityGroupIds = additionalSecurityGroupsList
		} else {
			state.AWSNodePool.AdditionalSecurityGroupIds = types.ListNull(types.StringType)
		}
		if httpTokensState, ok := awsNodePool.GetEc2MetadataHttpTokens(); ok {
			state.AWSNodePool.Ec2MetadataHttpTokens = types.StringValue(ec2.HttpTokensStateOptional)
			if httpTokensState != "" {
				state.AWSNodePool.Ec2MetadataHttpTokens = types.StringValue(string(httpTokensState))
			}
		}

		state.AWSNodePool.DiskSize = types.Int64Null()
		if rootVolume, ok := awsNodePool.GetRootVolume(); ok {
			if size, ok := rootVolume.GetSize(); ok {
				state.AWSNodePool.DiskSize = types.Int64Value(int64(size))
			}
		}

		if capacityReservation, ok := awsNodePool.GetCapacityReservation(); ok {
			if capacityReservationId, ok := capacityReservation.GetId(); ok {
				state.AWSNodePool.CapacityReservationId = types.StringValue(capacityReservationId)
			}
		} else {
			state.AWSNodePool.CapacityReservationId = types.StringNull()
		}
	}

	autoscaling, ok := object.GetAutoscaling()
	if ok {
		if state.AutoScaling == nil {
			state.AutoScaling = new(AutoScaling)
		}
		var minReplicas, maxReplicas int
		state.AutoScaling.Enabled = types.BoolValue(true)
		minReplicas, ok = autoscaling.GetMinReplica()
		if ok {
			state.AutoScaling.MinReplicas = types.Int64Value(int64(minReplicas))
		}
		maxReplicas, ok = autoscaling.GetMaxReplica()
		if ok {
			state.AutoScaling.MaxReplicas = types.Int64Value(int64(maxReplicas))
		}
	} else {
		if state.AutoScaling == nil {
			state.AutoScaling = new(AutoScaling)
		}
		state.AutoScaling.Enabled = types.BoolValue(false)
		state.AutoScaling.MinReplicas = types.Int64Null()
		state.AutoScaling.MaxReplicas = types.Int64Null()
	}

	if replicas, ok := object.GetReplicas(); ok {
		state.Replicas = types.Int64Value(int64(replicas))
	}

	taints := object.Taints()
	if len(taints) > 0 {
		state.Taints = make([]Taints, len(taints))
		for i, taint := range taints {
			state.Taints[i] = Taints{
				Key:          types.StringValue(taint.Key()),
				Value:        types.StringValue(taint.Value()),
				ScheduleType: types.StringValue(taint.Effect()),
			}
		}
	} else {
		state.Taints = nil
	}

	labels := object.Labels()
	if len(labels) > 0 {
		// XXX: We should be checking error here, but we don't have a way to return the error
		state.Labels, _ = common.ConvertStringMapToMapType(labels)
	} else {
		state.Labels = types.MapNull(types.StringType)
	}

	state.SubnetID = types.StringValue(object.Subnet())
	state.AvailabilityZone = types.StringValue(object.AvailabilityZone())

	if object.Status() != nil {
		state.NodePoolStatus = flattenNodePoolStatus(int64(object.Status().CurrentReplicas()), object.Status().Message())
	} else if state.NodePoolStatus.IsUnknown() {
		state.NodePoolStatus = nodePoolStatusNull()
	}

	if !common.HasValue(state.TuningConfigs) {
		state.TuningConfigs = types.ListNull(types.StringType)
	}
	if len(object.TuningConfigs()) > 0 {
		tuningConfigsList, err := common.StringArrayToList(object.TuningConfigs())
		if err != nil {
			return err
		}
		state.TuningConfigs = tuningConfigsList
	}

	if len(object.KubeletConfigs()) > 0 {
		state.KubeletConfigs = types.StringValue(object.KubeletConfigs()[0])
	} else if len(state.KubeletConfigs.ValueString()) != 0 {
		state.KubeletConfigs = types.StringNull()
	}

	if object.Version() != nil {
		version, ok := object.Version().GetID()
		// If we're using a non-default channel group, it will have been appended to
		// the version ID. Remove it before saving state.
		version = strings.TrimPrefix(version, rosa.VersionPrefix)
		if ok {
			state.CurrentVersion = types.StringValue(version)
		} else {
			state.CurrentVersion = types.StringNull()
		}
	}

	state.AutoRepair = types.BoolValue(object.AutoRepair())
	return nil
}

func filterClusterTagsNotPresentInNpInput(ctx context.Context, state *HcpMachinePoolState, cluster *cmv1.Cluster, awsTags map[string]string) (map[string]string, error) {
	if len(awsTags) == 0 {
		return awsTags, nil
	}
	if cluster.AWS() == nil || len(cluster.AWS().Tags()) == 0 {
		return awsTags, nil
	}
	filteredTags := make(map[string]string, len(awsTags))
	for k, v := range awsTags {
		filteredTags[k] = v
	}
	currentNpTfTags, err := common.OptionalMap(ctx, state.AWSNodePool.Tags)
	if err != nil {
		return filteredTags, err
	}
	clusterTags := cluster.AWS().Tags()
	for k := range clusterTags {
		if _, ok := currentNpTfTags[k]; !ok {
			delete(filteredTags, k)
		}
	}
	return filteredTags, nil
}

func shouldPatchTaints(a, b []Taints) bool {
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return true
	}
	if len(a) != len(b) {
		return true
	}
	for i := range a {
		if !a[i].Key.Equal(b[i].Key) || !a[i].Value.Equal(b[i].Value) || !a[i].ScheduleType.Equal(b[i].ScheduleType) {
			return true
		}
	}
	return false
}
