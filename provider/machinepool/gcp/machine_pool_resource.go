/*
Copyright (c) 2025 Red Hat, Inc.

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

package gcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// MachinePoolResource implements the rhcs_machine_pool resource for OSD-GCP clusters.
type MachinePoolResource struct {
	collection *cmv1.ClustersClient
}

var _ resource.Resource = &MachinePoolResource{}
var _ resource.ResourceWithConfigure = &MachinePoolResource{}
var _ resource.ResourceWithImportState = &MachinePoolResource{}
var _ resource.ResourceWithConfigValidators = &MachinePoolResource{}

// New creates a new OSD-GCP machine pool resource.
func New() resource.Resource {
	return &MachinePoolResource{}
}

func (r *MachinePoolResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	// Mirrors the HCP precedent (rhcs_hcp_machine_pool); the GCP variant uses
	// its own TypeName since two resources cannot share rhcs_machine_pool.
	resp.TypeName = req.ProviderTypeName + "_gcp_machine_pool"
}

func (r *MachinePoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Machine pool for an OSD cluster on GCP.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the machine pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the machine pool. Cannot be 'worker' or 'workers-*' (reserved for the default worker " +
					"pool, which is not yet manageable as a machine pool resource).",
				Required: true,
				Validators: []validator.String{
					machinePoolNameValidator{},
				},
			},
			"instance_type": schema.StringAttribute{
				Description: "GCP machine type (e.g., custom-4-16384).",
				Required:    true,
			},
			"replicas": schema.Int64Attribute{
				Description: "Number of replicas (when not using autoscaling).",
				Optional:    true,
			},
			"availability_zones": schema.ListAttribute{
				Description: "GCP availability zones. When not specified, OCM assigns a default zone; this value is populated " +
					"from the API after create.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"labels": schema.MapAttribute{
				Description: "Kubernetes labels.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"taints": schema.ListNestedAttribute{
				Description: "Taints for the machine pool.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":    schema.StringAttribute{Required: true},
						"value":  schema.StringAttribute{Required: true},
						"effect": schema.StringAttribute{Required: true},
					},
				},
			},
			"root_volume_size": schema.Int64Attribute{
				Description: "Root volume size in GiB.",
				Optional:    true,
			},
			"autoscaling": schema.SingleNestedAttribute{
				Description: "Autoscaling configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"min_replicas": schema.Int64Attribute{Required: true},
					"max_replicas": schema.Int64Attribute{Required: true},
				},
			},
			"gcp": schema.SingleNestedAttribute{
				Description: "GCP-specific machine pool options.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"secure_boot": schema.BoolAttribute{
						Description: "Enable Shielded VM Secure Boot.",
						Optional:    true,
					},
				},
			},
		},
	}
}

// machinePoolNameValidator rejects reserved names for the default worker pool.
type machinePoolNameValidator struct{}

func (machinePoolNameValidator) Description(_ context.Context) string {
	return "name cannot be 'worker' or 'workers-*' (reserved for the default worker pool)"
}

func (machinePoolNameValidator) MarkdownDescription(ctx context.Context) string {
	return "name cannot be 'worker' or 'workers-*' (reserved for the default worker pool)"
}

func (machinePoolNameValidator) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	name := req.ConfigValue.ValueString()
	if name == "worker" || strings.HasPrefix(name, "workers-") {
		resp.Diagnostics.AddError(
			"Invalid machine pool name",
			"The name 'worker' and names matching 'workers-*' are reserved for the cluster's "+
				"default worker pool, which is not yet manageable as a machine pool resource. "+
				"Choose a different name.",
		)
	}
}

// machinePoolConfigValidator enforces cross-attribute rules.
type machinePoolConfigValidator struct{}

func (machinePoolConfigValidator) Description(_ context.Context) string {
	return "validates autoscaling vs replicas mutual exclusivity and secure_boot with bare metal"
}

func (machinePoolConfigValidator) MarkdownDescription(ctx context.Context) string {
	return "validates autoscaling vs replicas mutual exclusivity and secure_boot with bare metal"
}

func (v *machinePoolConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var config MachinePoolState
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	secureBoot := false
	if !config.GCP.IsNull() && !config.GCP.IsUnknown() {
		if sb, ok := config.GCP.Attributes()["secure_boot"].(types.Bool); ok && !sb.IsNull() && !sb.IsUnknown() &&
			sb.ValueBool() {
			secureBoot = true
		}
	}
	instanceType := config.InstanceType.ValueString()
	if secureBoot && strings.HasSuffix(instanceType, "-metal") {
		resp.Diagnostics.AddError(
			"Invalid machine pool configuration",
			"Secure Boot is not supported on bare metal instance types (e.g. *-metal).",
		)
	}
}

func (r *MachinePoolResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(
			path.MatchRoot("autoscaling").AtName("min_replicas"),
			path.MatchRoot("autoscaling").AtName("max_replicas"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("replicas"),
			path.MatchRoot("autoscaling").AtName("min_replicas"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("replicas"),
			path.MatchRoot("autoscaling").AtName("max_replicas"),
		),
		&machinePoolConfigValidator{},
	}
}

func (r *MachinePoolResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	r.collection = conn.ClustersMgmt().V1().Clusters()
}

func (r *MachinePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MachinePoolState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpObj, err := r.buildMachinePoolObject(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("failed to build machine pool", err.Error())
		return
	}

	collection := r.collection.Cluster(plan.ClusterID.ValueString()).MachinePools()
	addResp, err := collection.Add().Body(mpObj).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create machine pool", err.Error())
		return
	}
	obj := addResp.Body()

	r.populateState(obj, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MachinePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MachinePoolState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.collection.Cluster(state.ClusterID.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString()).
		Get().
		SendContext(ctx)
	if err != nil {
		if getResp != nil && getResp.Status() == http.StatusNotFound {
			tflog.Warn(ctx, "machine pool not found, removing from state", map[string]interface{}{
				"cluster_id": state.ClusterID.ValueString(),
				"id":         state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to get machine pool", err.Error())
		return
	}
	obj := getResp.Body()

	r.populateState(obj, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MachinePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MachinePoolState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mpObj, err := r.buildMachinePoolObject(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("failed to build machine pool", err.Error())
		return
	}

	_, err = r.collection.Cluster(plan.ClusterID.ValueString()).
		MachinePools().
		MachinePool(plan.ID.ValueString()).
		Update().
		Body(mpObj).
		SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to update machine pool", err.Error())
		return
	}

	getResp, err := r.collection.Cluster(plan.ClusterID.ValueString()).
		MachinePools().
		MachinePool(plan.ID.ValueString()).
		Get().
		SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to get machine pool after update", err.Error())
		return
	}
	r.populateState(getResp.Body(), &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MachinePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MachinePoolState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	delResp, err := r.collection.Cluster(state.ClusterID.ValueString()).
		MachinePools().
		MachinePool(state.ID.ValueString()).
		Delete().
		SendContext(ctx)
	if err != nil {
		// Idempotent: a 404 means the pool (or the cluster itself) is
		// already gone. Treat that as success rather than blocking destroy.
		if delResp != nil && delResp.Status() == http.StatusNotFound {
			tflog.Warn(ctx, "machine pool already gone, treating delete as successful", map[string]interface{}{
				"cluster_id": state.ClusterID.ValueString(),
				"id":         state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to delete machine pool", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *MachinePoolResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *MachinePoolResource) buildMachinePoolObject(
	ctx context.Context,
	s *MachinePoolState,
) (*cmv1.MachinePool, error) {
	builder := cmv1.NewMachinePool().
		ID(s.Name.ValueString()).
		InstanceType(s.InstanceType.ValueString())

	hasAutoscaling := false
	if !s.Autoscaling.IsNull() && !s.Autoscaling.IsUnknown() {
		attrs := s.Autoscaling.Attributes()
		minVal, okMin := attrs["min_replicas"].(types.Int64)
		maxVal, okMax := attrs["max_replicas"].(types.Int64)
		if okMin && okMax && !minVal.IsNull() && !minVal.IsUnknown() && !maxVal.IsNull() && !maxVal.IsUnknown() {
			autoscaling := cmv1.NewMachinePoolAutoscaling().
				MinReplicas(int(minVal.ValueInt64())).
				MaxReplicas(int(maxVal.ValueInt64()))
			builder.Autoscaling(autoscaling)
			hasAutoscaling = true
		}
	}
	if !hasAutoscaling {
		replicas := 3
		if !s.Replicas.IsNull() && !s.Replicas.IsUnknown() {
			replicas = int(s.Replicas.ValueInt64())
		}
		builder.Replicas(replicas)
	}

	if !s.AvailabilityZones.IsNull() && !s.AvailabilityZones.IsUnknown() {
		azs, convErr := common.StringListToArray(ctx, s.AvailabilityZones)
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert availability_zones: %w", convErr)
		}
		builder.AvailabilityZones(azs...)
	}
	if !s.Labels.IsNull() && !s.Labels.IsUnknown() {
		labels := make(map[string]string)
		for k, v := range s.Labels.Elements() {
			if str, ok := v.(types.String); ok {
				labels[k] = str.ValueString()
			}
		}
		builder.Labels(labels)
	}
	if !s.Taints.IsNull() && !s.Taints.IsUnknown() {
		var taintBuilders []*cmv1.TaintBuilder
		for _, elem := range s.Taints.Elements() {
			obj, ok := elem.(types.Object)
			if !ok {
				continue
			}
			attrs := obj.Attributes()
			keyVal, okK := attrs["key"].(types.String)
			valVal, okV := attrs["value"].(types.String)
			effVal, okE := attrs["effect"].(types.String)
			if okK && okV && okE {
				taintBuilders = append(
					taintBuilders,
					cmv1.NewTaint().Key(keyVal.ValueString()).Value(valVal.ValueString()).Effect(effVal.ValueString()),
				)
			}
		}
		if len(taintBuilders) > 0 {
			builder.Taints(taintBuilders...)
		}
	}
	if !s.RootVolumeSize.IsNull() && s.RootVolumeSize.ValueInt64() > 0 {
		builder.RootVolume(cmv1.NewRootVolume().GCP(cmv1.NewGCPVolume().Size(int(s.RootVolumeSize.ValueInt64()))))
	}
	if !s.GCP.IsNull() && !s.GCP.IsUnknown() {
		if sb, ok := s.GCP.Attributes()["secure_boot"].(types.Bool); ok && !sb.IsNull() && !sb.IsUnknown() &&
			sb.ValueBool() {
			builder.GCP(cmv1.NewGCPMachinePool().SecureBoot(true))
		}
	}

	return builder.Build()
}

func (r *MachinePoolResource) populateState(obj *cmv1.MachinePool, state *MachinePoolState) {
	state.ID = types.StringValue(obj.ID())
	state.Name = types.StringValue(obj.ID())
	state.InstanceType = types.StringValue(obj.InstanceType())
	state.Replicas = types.Int64Value(int64(obj.Replicas()))
	if obj.Autoscaling() != nil {
		a := obj.Autoscaling()
		state.Autoscaling, _ = types.ObjectValue(autoscalingAttrTypes, map[string]attr.Value{
			"min_replicas": types.Int64Value(int64(a.MinReplicas())),
			"max_replicas": types.Int64Value(int64(a.MaxReplicas())),
		})
	} else {
		state.Autoscaling = types.ObjectNull(autoscalingAttrTypes)
	}
	if len(obj.AvailabilityZones()) > 0 {
		azList, _ := types.ListValueFrom(context.Background(), types.StringType, obj.AvailabilityZones())
		state.AvailabilityZones = azList
	} else {
		// Optional+Computed: must return a known value after apply. Use empty list when API returns nil/empty.
		state.AvailabilityZones = types.ListValueMust(types.StringType, []attr.Value{})
	}
	if len(obj.Labels()) > 0 {
		labelMap, _ := types.MapValueFrom(context.Background(), types.StringType, obj.Labels())
		state.Labels = labelMap
	}
	if obj.GCP() != nil {
		state.GCP, _ = types.ObjectValue(gcpAttrTypes, map[string]attr.Value{
			"secure_boot": types.BoolValue(obj.GCP().SecureBoot()),
		})
	} else {
		state.GCP = types.ObjectNull(gcpAttrTypes)
	}
}
