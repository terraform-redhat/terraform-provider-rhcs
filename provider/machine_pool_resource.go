/*
Copyright (c) 2021 Red Hat, Inc.

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

package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type MachinePoolResourceType struct {
}

var machinepoolNameRE = regexp.MustCompile(
	`^[a-z]([-a-z0-9]*[a-z0-9])?$`,
)

type MachinePoolResource struct {
	collection *cmv1.ClustersClient
}

func (t *MachinePoolResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "Machine pool.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"id": {
				Description: "Unique identifier of the machine pool.",
				Type:        types.StringType,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.UseStateForUnknown(),
				},
			},
			"name": {
				Description: "Name of the machine pool.Must consist of lower-case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character.",
				Type:        types.StringType,
				Required:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"machine_type": {
				Description: "Identifier of the machine type used by the nodes, " +
					"for example `r5.xlarge`. Use the `rhcs_machine_types` data " +
					"source to find the possible values.",
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"replicas": {
				Description: "The number of machines of the pool",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"use_spot_instances": {
				Description: "Use Spot Instances.",
				Type:        types.BoolType,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"max_spot_price": {
				Description: "Max Spot price.",
				Type:        types.Float64Type,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(),
				},
			},
			"autoscaling_enabled": {
				Description: "Enables autoscaling. This variable requires you to set a maximum and minimum replicas range using the `max_replicas` and `min_replicas` variables.",
				Type:        types.BoolType,
				Optional:    true,
			},
			"min_replicas": {
				Description: "The minimum number of replicas for autoscaling.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"max_replicas": {
				Description: "The maximum number of replicas for autoscaling functionality.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"taints": {
				Description: "Taints for machine pool. Format should be a comma-separated " +
					"list of 'key=value:ScheduleType'. This list will overwrite any modifications " +
					"made to node taints on an ongoing basis.\n",
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"key": {
						Description: "Taints key",
						Type:        types.StringType,
						Required:    true,
					},
					"value": {
						Description: "Taints value",
						Type:        types.StringType,
						Required:    true,
					},
					"schedule_type": {
						Description: "Taints schedule type",
						Type:        types.StringType,
						Required:    true,
					},
				}, tfsdk.ListNestedAttributesOptions{},
				),
				Optional: true,
			},
			"labels": {
				Description: "Labels for the machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis.",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"multi_availability_zone": {
				Description: "Create a multi-AZ machine pool for a multi-AZ cluster (default true)",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"availability_zone": {
				Description: "Select availability zone to create a single AZ machine pool for a multi-AZ cluster",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
			"subnet_id": {
				Description: "Select subnet to create a single AZ machine pool for BYOVPC cluster",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.UseStateForUnknown(),
					ValueCannotBeChangedModifier(),
				},
			},
		},
	}
	return
}

func (t *MachinePoolResourceType) NewResource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.Resource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation: use it directly when needed.
	parent := p.(*Provider)

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt().V1().Clusters()

	// Create the resource:
	result = &MachinePoolResource{
		collection: collection,
	}

	return
}

func (r *MachinePoolResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	// Get the plan:
	state := &MachinePoolState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	machinepoolName := state.Name.Value
	if !machinepoolNameRE.MatchString(machinepoolName) {
		response.Diagnostics.AddError(
			"Can't create machine pool: ",
			fmt.Sprintf("Can't create machine pool for cluster '%s' with name '%s'. Expected a valid value for 'name' matching %s",
				state.Cluster.Value, state.Name.Value, machinepoolNameRE,
			),
		)
		return
	}

	// Wait till the cluster is ready:
	resource := r.collection.Cluster(state.Cluster.Value)
	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()
	_, err := resource.Poll().
		Interval(30 * time.Second).
		Predicate(func(get *cmv1.ClusterGetResponse) bool {
			return get.Body().State() == cmv1.ClusterStateReady
		}).
		StartContext(pollCtx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't poll cluster state",
			fmt.Sprintf(
				"Can't poll state of cluster with identifier '%s': %v",
				state.Cluster.Value, err,
			),
		)
		return
	}

	// Create the machine pool:
	builder := cmv1.NewMachinePool().ID(state.ID.Value).InstanceType(state.MachineType.Value)
	builder.ID(state.Name.Value)

	if err := setSpotInstances(state, builder); err != nil {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s: %v'", state.Cluster.Value, err,
			),
		)
		return
	}

	isMultiAZPool, err := r.validateAZConfig(state)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s': %v",
				state.Cluster.Value, err,
			),
		)
		return
	}
	if !common.IsStringAttributeEmpty(state.AvailabilityZone) {
		builder.AvailabilityZones(state.AvailabilityZone.Value)
	}
	if !common.IsStringAttributeEmpty(state.SubnetID) {
		builder.Subnets(state.SubnetID.Value)
	}

	autoscalingEnabled := false
	computeNodeEnabled := false
	autoscalingEnabled, errMsg := getAutoscaling(state, builder)
	if errMsg != "" {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s, %s'", state.Cluster.Value, errMsg,
			),
		)
		return
	}

	if !state.Replicas.Unknown && !state.Replicas.Null {
		computeNodeEnabled = true
		if isMultiAZPool && state.Replicas.Value%3 != 0 {
			response.Diagnostics.AddError(
				"Can't build machine pool",
				fmt.Sprintf(
					"Can't build machine pool for cluster '%s', replicas must be a multiple of 3",
					state.Cluster.Value,
				),
			)
			return
		}
		builder.Replicas(int(state.Replicas.Value))
	}
	if (!autoscalingEnabled && !computeNodeEnabled) || (autoscalingEnabled && computeNodeEnabled) {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s', please provide a value for either the 'replicas' or 'autoscaling_enabled' parameter. It is mandatory to include at least one of these parameters in the resource plan.",
				state.Cluster.Value,
			),
		)
		return
	}

	if state.Taints != nil && len(state.Taints) > 0 {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range state.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint().Key(taint.Key.Value).Value(taint.Value.Value).Effect(taint.ScheduleType.Value))
		}
		builder.Taints(taintBuilders...)
	}

	if !state.Labels.Unknown && !state.Labels.Null {
		labels := map[string]string{}
		for k, v := range state.Labels.Elems {
			labels[k] = v.(types.String).Value
		}
		builder.Labels(labels)
	}

	object, err := builder.Build()
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s': %v",
				state.Cluster.Value, err,
			),
		)
		return
	}

	collection := resource.MachinePools()
	add, err := collection.Add().Body(object).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't create machine pool",
			fmt.Sprintf(
				"Can't create machine pool for cluster '%s': %v",
				state.Cluster.Value, err,
			),
		)
		return
	}
	object = add.Body()

	// Save the state:
	r.populateState(object, state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *MachinePoolResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse) {
	// Get the current state:
	state := &MachinePoolState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the machine pool:
	resource := r.collection.Cluster(state.Cluster.Value).
		MachinePools().
		MachinePool(state.ID.Value)
	get, err := resource.Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("machine pool (%s) of cluster (%s) not found, removing from state",
			state.ID.Value, state.Cluster.Value,
		))
		response.State.RemoveResource(ctx)
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Failed to fetch machine pool",
			fmt.Sprintf(
				"Failed to fetch machine pool with identifier %s for cluster %s. Response code: %v",
				state.ID.Value, state.Cluster.Value, get.Status(),
			),
		)
		return
	}

	object := get.Body()

	// Save the state:
	r.populateState(object, state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *MachinePoolResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse) {
	var diags diag.Diagnostics

	// Get the state:
	state := &MachinePoolState{}
	diags = request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &MachinePoolState{}
	diags = request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resource := r.collection.Cluster(state.Cluster.Value).
		MachinePools().
		MachinePool(state.ID.Value)
	_, err := resource.Get().SendContext(ctx)

	if err != nil {
		response.Diagnostics.AddError(
			"Can't find machine pool",
			fmt.Sprintf(
				"Can't find machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			),
		)
		return
	}

	mpBuilder := cmv1.NewMachinePool().ID(state.ID.Value)

	_, ok := common.ShouldPatchString(state.MachineType, plan.MachineType)
	if ok {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s', machine type cannot be updated",
				state.Cluster.Value,
			),
		)
		return
	}

	computeNodesEnabled := false
	autoscalingEnabled := false

	if !plan.Replicas.Unknown && !plan.Replicas.Null {
		computeNodesEnabled = true
		mpBuilder.Replicas(int(plan.Replicas.Value))

	}

	autoscalingEnabled, errMsg := getAutoscaling(plan, mpBuilder)
	if errMsg != "" {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s, %s ", state.Cluster.Value, errMsg,
			),
		)
		return
	}

	if (autoscalingEnabled && computeNodesEnabled) || (!autoscalingEnabled && !computeNodesEnabled) {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s: either autoscaling or compute nodes should be enabled", state.Cluster.Value,
			),
		)
		return
	}

	patchLabels, shouldPatchLabels := common.ShouldPatchMap(state.Labels, plan.Labels)
	if shouldPatchLabels {
		labels := map[string]string{}
		for k, v := range patchLabels.Elems {
			labels[k] = v.(types.String).Value
		}
		mpBuilder.Labels(labels)
	}

	if shouldPatchTaints(state.Taints, plan.Taints) {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range plan.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint().Key(taint.Key.Value).Value(taint.Value.Value).Effect(taint.ScheduleType.Value))
		}
		mpBuilder.Taints(taintBuilders...)
	}

	machinePool, err := mpBuilder.Build()
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s: %v ", state.Cluster.Value, err,
			),
		)
		return
	}
	update, err := r.collection.Cluster(state.Cluster.Value).
		MachinePools().
		MachinePool(state.ID.Value).Update().Body(machinePool).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to update machine pool",
			fmt.Sprintf(
				"Failed to update machine pool '%s'  on cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			),
		)
		return
	}

	object := update.Body()

	// update the autoscaling enabled with the plan value (important for nil and false cases)
	state.AutoScalingEnabled = plan.AutoScalingEnabled
	// update the Replicas with the plan value (important for nil and zero value cases)
	state.Replicas = plan.Replicas

	// Save the state:
	r.populateState(object, state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

// Validate the machine pool's settings that pertain to availability zones.
// Returns whether the machine pool is/will be multi-AZ.
func (r *MachinePoolResource) validateAZConfig(state *MachinePoolState) (bool, error) {
	resp, err := r.collection.Cluster(state.Cluster.Value).Get().Send()
	if err != nil {
		return false, fmt.Errorf("failed to get information for cluster %s: %v", state.Cluster.Value, err)
	}
	cluster := resp.Body()
	isMultiAZCluster := cluster.MultiAZ()
	clusterAZs := cluster.Nodes().AvailabilityZones()
	clusterSubnets := cluster.AWS().SubnetIDs()

	if isMultiAZCluster {
		// Can't set both availability_zone and subnet_id
		if !common.IsStringAttributeEmpty(state.AvailabilityZone) && !common.IsStringAttributeEmpty(state.SubnetID) {
			return false, fmt.Errorf("availability_zone and subnet_id are mutually exclusive")
		}

		// multi_availability_zone setting must be consistent with availability_zone and subnet_id
		azOrSubnet := !common.IsStringAttributeEmpty(state.AvailabilityZone) || !common.IsStringAttributeEmpty(state.SubnetID)
		if !state.MultiAvailabilityZone.Null && !state.MultiAvailabilityZone.Unknown {
			if azOrSubnet && state.MultiAvailabilityZone.Value {
				return false, fmt.Errorf("multi_availability_zone must be False when availability_zone or subnet_id is set")
			}
		} else {
			state.MultiAvailabilityZone = types.Bool{Value: !azOrSubnet}
		}
	} else { // not a multi-AZ cluster
		if !common.IsStringAttributeEmpty(state.AvailabilityZone) {
			return false, fmt.Errorf("availability_zone can only be set for multi-AZ clusters")
		}
		if !common.IsStringAttributeEmpty(state.SubnetID) {
			return false, fmt.Errorf("subnet_id can only be set for multi-AZ clusters")
		}
		if !state.MultiAvailabilityZone.Null && !state.MultiAvailabilityZone.Unknown && state.MultiAvailabilityZone.Value {
			return false, fmt.Errorf("multi_availability_zone can only be set for multi-AZ clusters")
		}
		state.MultiAvailabilityZone = types.Bool{Value: false}
	}

	// Ensure that the machine pool's AZ and subnet are valid for the cluster
	// If subnet is set, we make sure it's valid for the cluster, but we don't default it if not set
	if !common.IsStringAttributeEmpty(state.SubnetID) {
		inClusterSubnet := false
		for _, subnet := range clusterSubnets {
			if subnet == state.SubnetID.Value {
				inClusterSubnet = true
				break
			}
		}
		if !inClusterSubnet {
			return false, fmt.Errorf("subnet_id %s is not valid for cluster %s", state.SubnetID.Value, state.Cluster.Value)
		}
	} else {
		state.SubnetID = types.String{Null: true}
	}
	// If AZ is set, we make sure it's valid for the cluster. If not set and neither is subnet, we default it to the 1st AZ in the cluster
	if !common.IsStringAttributeEmpty(state.AvailabilityZone) {
		inClusterAZ := false
		for _, az := range clusterAZs {
			if az == state.AvailabilityZone.Value {
				inClusterAZ = true
				break
			}
		}
		if !inClusterAZ {
			return false, fmt.Errorf("availability_zone %s is not valid for cluster %s", state.AvailabilityZone.Value, state.Cluster.Value)
		}
	} else {
		if len(clusterAZs) > 0 && !state.MultiAvailabilityZone.Value && isMultiAZCluster && common.IsStringAttributeEmpty(state.SubnetID) {
			state.AvailabilityZone = types.String{Value: clusterAZs[0]}
		} else {
			state.AvailabilityZone = types.String{Null: true}
		}
	}
	return state.MultiAvailabilityZone.Value, nil
}

func setSpotInstances(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder) error {
	useSpotInstances := !state.UseSpotInstances.Unknown && !state.UseSpotInstances.Null && state.UseSpotInstances.Value
	isSpotMaxPriceSet := !state.MaxSpotPrice.Unknown && !state.MaxSpotPrice.Null

	if isSpotMaxPriceSet && !useSpotInstances {
		return errors.New("Can't set max price when not using spot instances (set \"use_spot_instances\" to true)")
	}

	if useSpotInstances {
		if isSpotMaxPriceSet && state.MaxSpotPrice.Value <= 0 {
			return errors.New("To use Spot instances, you must set \"max_spot_price\" with positive value")
		}

		awsMachinePool := cmv1.NewAWSMachinePool()
		spotMarketOptions := cmv1.NewAWSSpotMarketOptions()
		if isSpotMaxPriceSet {
			spotMarketOptions.MaxPrice(float64(state.MaxSpotPrice.Value))
		}
		awsMachinePool.SpotMarketOptions(spotMarketOptions)
		mpBuilder.AWS(awsMachinePool)
	}

	return nil
}

func getAutoscaling(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder) (
	autoscalingEnabled bool, errMsg string) {
	autoscalingEnabled = false
	if !state.AutoScalingEnabled.Unknown && !state.AutoScalingEnabled.Null && state.AutoScalingEnabled.Value {
		autoscalingEnabled = true

		autoscaling := cmv1.NewMachinePoolAutoscaling()
		if !state.MaxReplicas.Unknown && !state.MaxReplicas.Null {
			autoscaling.MaxReplicas(int(state.MaxReplicas.Value))
		} else {
			return false, "when enabling autoscaling, should set value for maxReplicas"
		}
		if !state.MinReplicas.Unknown && !state.MinReplicas.Null {
			autoscaling.MinReplicas(int(state.MinReplicas.Value))
		} else {
			return false, "when enabling autoscaling, should set value for minReplicas"
		}
		if !autoscaling.Empty() {
			mpBuilder.Autoscaling(autoscaling)
		}
	} else {
		if (!state.MaxReplicas.Unknown && !state.MaxReplicas.Null) || (!state.MinReplicas.Unknown && !state.MinReplicas.Null) {
			return false, "when disabling autoscaling, can't set min_replicas and/or max_replicas"
		}
	}

	return autoscalingEnabled, ""
}

func (r *MachinePoolResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse) {
	// Get the state:
	state := &MachinePoolState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send the request to delete the machine pool:
	resource := r.collection.Cluster(state.Cluster.Value).
		MachinePools().
		MachinePool(state.ID.Value)
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete machine pool",
			fmt.Sprintf(
				"Can't delete machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			),
		)
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx)
}

func (r *MachinePoolResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	// To import a machine pool, we need to know the cluster ID and the machine pool name
	fields := strings.Split(request.ID, ",")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		response.Diagnostics.AddError(
			"Invalid import identifier",
			"Machine pool to import should be specified as <cluster_id>,<machine_pool_name>",
		)
		return
	}
	clusterID := fields[0]
	machinePoolName := fields[1]
	response.Diagnostics.Append(response.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("cluster"), clusterID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), machinePoolName)...)
}

// populateState copies the data from the API object to the Terraform state.
func (r *MachinePoolResource) populateState(object *cmv1.MachinePool, state *MachinePoolState) {
	state.ID = types.String{
		Value: object.ID(),
	}
	state.Name = types.String{
		Value: object.ID(),
	}

	if getAWS, ok := object.GetAWS(); ok {
		if spotMarketOptions, ok := getAWS.GetSpotMarketOptions(); ok {
			state.UseSpotInstances = types.Bool{Value: true}
			if spotMarketOptions.MaxPrice() != 0 {
				state.MaxSpotPrice = types.Float64{
					Value: float64(spotMarketOptions.MaxPrice()),
				}
			}
		}
	}

	autoscaling, ok := object.GetAutoscaling()
	if ok {
		var minReplicas, maxReplicas int
		state.AutoScalingEnabled = types.Bool{Value: true}
		minReplicas, ok = autoscaling.GetMinReplicas()
		if ok {
			state.MinReplicas = types.Int64{
				Value: int64(minReplicas),
			}
		}
		maxReplicas, ok = autoscaling.GetMaxReplicas()
		if ok {
			state.MaxReplicas = types.Int64{
				Value: int64(maxReplicas),
			}
		}
	} else {
		state.MaxReplicas = types.Int64{Null: true}
		state.MinReplicas = types.Int64{Null: true}
	}

	instanceType, ok := object.GetInstanceType()
	if ok {
		{
			state.MachineType = types.String{
				Value: instanceType,
			}
		}
	}

	replicas, ok := object.GetReplicas()
	if ok {
		state.Replicas = types.Int64{
			Value: int64(replicas),
		}
	}

	taints := object.Taints()
	if len(taints) > 0 {
		state.Taints = make([]Taints, len(taints))
		for i, taint := range taints {
			state.Taints[i] = Taints{
				Key:          types.String{Value: taint.Key()},
				Value:        types.String{Value: taint.Value()},
				ScheduleType: types.String{Value: taint.Effect()},
			}
		}
	} else {
		state.Taints = nil
	}

	labels := object.Labels()
	if len(labels) > 0 {
		state.Labels = types.Map{
			ElemType: types.StringType,
			Elems:    map[string]attr.Value{},
		}
		for k, v := range labels {
			state.Labels.Elems[k] = types.String{
				Value: v,
			}
		}
	} else {
		state.Labels.Null = true
	}
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
