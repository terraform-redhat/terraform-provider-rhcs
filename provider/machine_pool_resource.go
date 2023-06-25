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
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"

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
			},
			"id": {
				Description: "Unique identifier of the machine pool.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the machine pool.Must consist of lower-case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character.",
				Type:        types.StringType,
				Required:    true,
			},
			"machine_type": {
				Description: "Identifier of the machine type used by the nodes, " +
					"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
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
				Description: "Enables autoscaling.",
				Type:        types.BoolType,
				Optional:    true,
			},
			"min_replicas": {
				Description: "Min replicas.",
				Type:        types.Int64Type,
				Optional:    true,
			},
			"max_replicas": {
				Description: "Max replicas.",
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
				Description: "Labels for machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis..",
				Type: types.MapType{
					ElemType: types.StringType,
				},
				Optional: true,
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

	_, errMsg := getSpotInstances(state, builder)
	if errMsg != "" {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s, %s'", state.Cluster.Value, errMsg,
			),
		)
		return
	}

	autoscalingEnabled := false
	computeNodeEnabled := false
	autoscalingEnabled, errMsg = getAutoscaling(state, builder)
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
		builder.Replicas(int(state.Replicas.Value))
	}
	if (!autoscalingEnabled && !computeNodeEnabled) || (autoscalingEnabled && computeNodeEnabled) {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s', should hold either Autoscaling or Compute nodes",
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

func getSpotInstances(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder) (
	useSpotInstances bool, errMsg string) {
	useSpotInstances = false

	if !state.UseSpotInstances.Unknown && !state.UseSpotInstances.Null && state.UseSpotInstances.Value {
		useSpotInstances = true

		awsMachinePool := cmv1.NewAWSMachinePool()
		spotMarketOptions := cmv1.NewAWSSpotMarketOptions()
		if !state.MaxSpotPrice.Unknown && !state.MaxSpotPrice.Null {
			spotMarketOptions.MaxPrice(float64(state.MaxSpotPrice.Value))
		}
		awsMachinePool.SpotMarketOptions(spotMarketOptions)

		if !awsMachinePool.Empty() {
			mpBuilder.AWS(awsMachinePool)
		}
	} else {
		if !state.MaxSpotPrice.Unknown && !state.MaxSpotPrice.Null {
			return false, "when not using aws spot instances, can't set max_spot_price"
		}
	}

	return useSpotInstances, ""
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
	tfsdk.ResourceImportStatePassthroughID(
		ctx,
		tftypes.NewAttributePath().WithAttributeName("id"),
		request,
		response,
	)
}

// populateState copies the data from the API object to the Terraform state.
func (r *MachinePoolResource) populateState(object *cmv1.MachinePool, state *MachinePoolState) {
	state.ID = types.String{
		Value: object.ID(),
	}
	state.Name = types.String{
		Value: object.ID(),
	}

	getAWS, ok := object.GetAWS()
	if ok {
		state.UseSpotInstances = types.Bool{Value: true}
		spotMarketOptions, ok := getAWS.GetSpotMarketOptions()
		if ok {
			if spotMarketOptions.MaxPrice() != 0 {
				state.MaxSpotPrice = types.Float64{
					Value: float64(spotMarketOptions.MaxPrice()),
				}
			} else {
				state.MaxSpotPrice.Null = true
			}
		}
	} else {
		state.UseSpotInstances.Null = true
		state.MaxSpotPrice.Null = true
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
		state.MaxReplicas.Null = true
		state.MinReplicas.Null = true
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
	if labels != nil {
		state.Labels = types.Map{
			ElemType: types.StringType,
			Elems:    map[string]attr.Value{},
		}
		for k, v := range labels {
			state.Labels.Elems[k] = types.String{
				Value: v,
			}
		}

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
