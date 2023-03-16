/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
	"context"
***REMOVED***
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

type MachinePoolResourceType struct {
	logger logging.Logger
}

var machinepoolNameRE = regexp.MustCompile(
	`^[a-z]([-a-z0-9]*[a-z0-9]***REMOVED***?$`,
***REMOVED***

type MachinePoolResource struct {
	logger     logging.Logger
	collection *cmv1.ClustersClient
}

func (t *MachinePoolResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "Machine pool.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"id": {
				Description: "Unique identifier of the machine pool.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"name": {
				Description: "Name of the machine pool.Must consist of lower-case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"machine_type": {
				Description: "Identifier of the machine type used by the nodes, " +
					"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
					"source to find the possible values.",
				Type:     types.StringType,
				Required: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(t.logger***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"replicas": {
				Description: "The number of machines of the pool",
				Type:        types.Int64Type,
				Optional:    true,
	***REMOVED***,
			"use_spot_instances": {
				Description: "Use Spot Instances. Applicable only for AWS.",
				Type:        types.BoolType,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(t.logger***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"max_spot_price": {
				Description: "Max Spot price. Applicable only for AWS.",
				Type:        types.Float64Type,
				Optional:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					ValueCannotBeChangedModifier(t.logger***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"autoscaling_enabled": {
				Description: "Enables autoscaling.",
				Type:        types.BoolType,
				Optional:    true,
	***REMOVED***,
			"min_replicas": {
				Description: "Min replicas.",
				Type:        types.Int64Type,
				Optional:    true,
	***REMOVED***,
			"max_replicas": {
				Description: "Max replicas.",
				Type:        types.Int64Type,
				Optional:    true,
	***REMOVED***,
			"taints": {
				Description: "Taints for machine pool. Format should be a comma-separated " +
					"list of 'key=value:ScheduleType'. This list will overwrite any modifications " +
					"made to node taints on an ongoing basis.\n",
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"key": {
						Description: "Taints key",
						Type:        types.StringType,
						Required:    true,
			***REMOVED***,
					"value": {
						Description: "Taints value",
						Type:        types.StringType,
						Required:    true,
			***REMOVED***,
					"schedule_type": {
						Description: "Taints schedule type",
						Type:        types.StringType,
						Required:    true,
			***REMOVED***,
		***REMOVED***, tfsdk.ListNestedAttributesOptions{},
				***REMOVED***,
				Optional: true,
	***REMOVED***,
			"labels": {
				Description: "Labels for machine pool. Format should be a comma-separated list of 'key = value'." +
					" This list will overwrite any modifications made to node labels on an ongoing basis..",
				Type: types.MapType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *MachinePoolResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation: use it directly when needed.
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Create the resource:
	result = &MachinePoolResource{
		logger:     parent.logger,
		collection: collection,
	}

	return
}

func (r *MachinePoolResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	// Get the plan:
	state := &MachinePoolState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	machinepoolName := state.Name.Value
	if !machinepoolNameRE.MatchString(machinepoolName***REMOVED*** {
		response.Diagnostics.AddError(
			"Can't create machine pool: ",
			fmt.Sprintf("Can't create machine pool for cluster '%s' with name '%s'. Expected a valid value for 'name' matching %s",
				state.Cluster.Value, state.Name.Value, machinepoolNameRE,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Wait till the cluster is ready:
	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***
	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour***REMOVED***
	defer cancel(***REMOVED***
	_, err := resource.Poll(***REMOVED***.
		Interval(30 * time.Second***REMOVED***.
		Predicate(func(get *cmv1.ClusterGetResponse***REMOVED*** bool {
			return get.Body(***REMOVED***.State(***REMOVED*** == cmv1.ClusterStateReady
***REMOVED******REMOVED***.
		StartContext(pollCtx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't poll cluster state",
			fmt.Sprintf(
				"Can't poll state of cluster with identifier '%s': %v",
				state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Create the machine pool:
	builder := cmv1.NewMachinePool(***REMOVED***.ID(state.ID.Value***REMOVED***.InstanceType(state.MachineType.Value***REMOVED***
	builder.ID(state.Name.Value***REMOVED***

	var errMsg string
	_, errMsg = getSpotinstances(state, builder***REMOVED***
	if errMsg != "" {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s, %s'", state.Cluster.Value, errMsg,
			***REMOVED***,
		***REMOVED***
		return
	}

	autoscalingEnabled := false
	computeNodeEnabled := false
	autoscalingEnabled, errMsg = getAutoscaling(state, builder***REMOVED***
	if errMsg != "" {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s, %s'", state.Cluster.Value, errMsg,
			***REMOVED***,
		***REMOVED***
		return
	}

	if !state.Replicas.Unknown && !state.Replicas.Null {
		computeNodeEnabled = true
		builder.Replicas(int(state.Replicas.Value***REMOVED******REMOVED***
	}
	if (!autoscalingEnabled && !computeNodeEnabled***REMOVED*** || (autoscalingEnabled && computeNodeEnabled***REMOVED*** {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s', should hold either Autoscailing or Compute nodes",
				state.Cluster.Value,
			***REMOVED***,
		***REMOVED***
		return
	}

	if state.Taints != nil && len(state.Taints***REMOVED*** > 0 {
		var taintBuilders []*cmv1.TaintBuilder
		for _, taint := range state.Taints {
			taintBuilders = append(taintBuilders, cmv1.NewTaint(***REMOVED***.Key(taint.Key.Value***REMOVED***.Value(taint.Value.Value***REMOVED***.Effect(taint.ScheduleType.Value***REMOVED******REMOVED***
***REMOVED***
		builder.Taints(taintBuilders...***REMOVED***
	}

	if !state.Labels.Unknown && !state.Labels.Null {
		labels := map[string]string{}
		for k, v := range state.Labels.Elems {
			labels[k] = v.(types.String***REMOVED***.Value
***REMOVED***
		builder.Labels(labels***REMOVED***
	}

	object, err := builder.Build(***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build machine pool",
			fmt.Sprintf(
				"Can't build machine pool for cluster '%s': %v",
				state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	collection := resource.MachinePools(***REMOVED***
	add, err := collection.Add(***REMOVED***.Body(object***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't create machine pool",
			fmt.Sprintf(
				"Can't create machine pool for cluster '%s': %v",
				state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	// Save the state:
	r.populateState(object, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *MachinePoolResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	// Get the current state:
	state := &MachinePoolState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the machine pool:
	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***.
		MachinePools(***REMOVED***.
		MachinePool(state.ID.Value***REMOVED***
	get, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find machine pool",
			fmt.Sprintf(
				"Can't find machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Save the state:
	r.populateState(object, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *MachinePoolResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
	var diags diag.Diagnostics

	// Get the state:
	state := &MachinePoolState{}
	diags = request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Get the plan:
	plan := &MachinePoolState{}
	diags = request.Plan.Get(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***.
		MachinePools(***REMOVED***.
		MachinePool(state.ID.Value***REMOVED***
	_, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***

	if err != nil {
		response.Diagnostics.AddError(
			"Can't find machine pool",
			fmt.Sprintf(
				"Can't find machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	mpBuilder := cmv1.NewMachinePool(***REMOVED***.ID(state.ID.Value***REMOVED***

	_, ok := shouldPatchString(state.MachineType, plan.MachineType***REMOVED***
	if ok {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s', machine type cannot be updated",
				state.Cluster.Value,
			***REMOVED***,
		***REMOVED***
		return
	}

	var errMsg string
	_, errMsg = getSpotinstances(plan, mpBuilder***REMOVED***
	if errMsg != "" {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s, %s ", state.Cluster.Value, errMsg,
			***REMOVED***,
		***REMOVED***
		return
	}

	computeNodesEnabled := false
	autoscalingEnabled := false

	if !plan.Replicas.Unknown && !plan.Replicas.Null {
		computeNodesEnabled = true
		mpBuilder.Replicas(int(plan.Replicas.Value***REMOVED******REMOVED***

	}

	autoscalingEnabled, errMsg = getAutoscaling(plan, mpBuilder***REMOVED***
	if errMsg != "" {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s, %s ", state.Cluster.Value, errMsg,
			***REMOVED***,
		***REMOVED***
		return
	}

	if (autoscalingEnabled && computeNodesEnabled***REMOVED*** || (!autoscalingEnabled && !computeNodesEnabled***REMOVED*** {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s: either autoscaling or compute nodes should be enabled", state.Cluster.Value,
			***REMOVED***,
		***REMOVED***
		return
	}

	machinePool, err := mpBuilder.Build(***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update machine pool",
			fmt.Sprintf(
				"Can't update machine pool for cluster '%s: %v ", state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	update, err := r.collection.Cluster(state.Cluster.Value***REMOVED***.
		MachinePools(***REMOVED***.
		MachinePool(state.ID.Value***REMOVED***.Update(***REMOVED***.Body(machinePool***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to update machine pool",
			fmt.Sprintf(
				"Failed to update machine pool '%s'  on cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object := update.Body(***REMOVED***

	// update the spot instances enabled with the plan value
	state.UseSpotInstances = plan.UseSpotInstances
	// update the autoscaling enabled with the plan value (important for nil and false cases***REMOVED***
	state.AutoScalingEnabled = plan.AutoScalingEnabled
	// update the Replicas with the plan value (important for nil and zero value cases***REMOVED***
	state.Replicas = plan.Replicas

	// Save the state:
	r.populateState(object, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func getSpotinstances(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder***REMOVED*** (
	useSpotInstances bool, errMsg string***REMOVED*** {
	useSpotInstances = false

	if !state.UseSpotInstances.Unknown && !state.UseSpotInstances.Null && state.UseSpotInstances.Value {
		useSpotInstances = true

		awsMachinePool := cmv1.NewAWSMachinePool(***REMOVED***
		spotMarketOptions := cmv1.NewAWSSpotMarketOptions(***REMOVED***
		if !state.MaxSpotPrice.Unknown && !state.MaxSpotPrice.Null {
			spotMarketOptions.MaxPrice(float64(state.MaxSpotPrice.Value***REMOVED******REMOVED***
***REMOVED***
		awsMachinePool.SpotMarketOptions(spotMarketOptions***REMOVED***

		if !awsMachinePool.Empty(***REMOVED*** {
			mpBuilder.AWS(awsMachinePool***REMOVED***
***REMOVED***
	} else {
		if !state.MaxSpotPrice.Unknown && !state.MaxSpotPrice.Null {
			return false, "when not using aws spot instances, can't set max_spot_price"
***REMOVED***
	}

	return useSpotInstances, ""
}

func getAutoscaling(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder***REMOVED*** (
	autoscalingEnabled bool, errMsg string***REMOVED*** {
	autoscalingEnabled = false
	if !state.AutoScalingEnabled.Unknown && !state.AutoScalingEnabled.Null && state.AutoScalingEnabled.Value {
		autoscalingEnabled = true

		autoscaling := cmv1.NewMachinePoolAutoscaling(***REMOVED***
		if !state.MaxReplicas.Unknown && !state.MaxReplicas.Null {
			autoscaling.MaxReplicas(int(state.MaxReplicas.Value***REMOVED******REMOVED***
***REMOVED*** else {
			return false, "when enabling autoscaling, should set value for maxReplicas"
***REMOVED***
		if !state.MinReplicas.Unknown && !state.MinReplicas.Null {
			autoscaling.MinReplicas(int(state.MinReplicas.Value***REMOVED******REMOVED***
***REMOVED*** else {
			return false, "when enabling autoscaling, should set value for minReplicas"
***REMOVED***
		if !autoscaling.Empty(***REMOVED*** {
			mpBuilder.Autoscaling(autoscaling***REMOVED***
***REMOVED***
	} else {
		if (!state.MaxReplicas.Unknown && !state.MaxReplicas.Null***REMOVED*** || (!state.MinReplicas.Unknown && !state.MinReplicas.Null***REMOVED*** {
			return false, "when disabling autoscaling, can't set min_replicas and/or max_replicas"
***REMOVED***
	}

	return autoscalingEnabled, ""
}

func (r *MachinePoolResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	// Get the state:
	state := &MachinePoolState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete the machine pool:
	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***.
		MachinePools(***REMOVED***.
		MachinePool(state.ID.Value***REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete machine pool",
			fmt.Sprintf(
				"Can't delete machine pool with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *MachinePoolResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	tfsdk.ResourceImportStatePassthroughID(
		ctx,
		tftypes.NewAttributePath(***REMOVED***.WithAttributeName("id"***REMOVED***,
		request,
		response,
	***REMOVED***
}

// populateState copies the data from the API object to the Terraform state.
func (r *MachinePoolResource***REMOVED*** populateState(object *cmv1.MachinePool, state *MachinePoolState***REMOVED*** {
	state.ID = types.String{
		Value: object.ID(***REMOVED***,
	}
	state.Name = types.String{
		Value: object.ID(***REMOVED***,
	}

	getAWS, ok := object.GetAWS(***REMOVED***
	if ok {
		state.UseSpotInstances = types.Bool{Value: true}
		spotMarketOptions, ok := getAWS.GetSpotMarketOptions(***REMOVED***
		if ok {
			if spotMarketOptions.MaxPrice(***REMOVED*** != 0 {
				state.MaxSpotPrice = types.Float64{
					Value: float64(spotMarketOptions.MaxPrice(***REMOVED******REMOVED***,
		***REMOVED***
	***REMOVED*** else {
				state.MaxSpotPrice.Null = true
	***REMOVED***
***REMOVED***
	} else {
		state.UseSpotInstances = types.Bool{Value: false}
		state.MaxSpotPrice.Null = true
	}

	autoscaling, ok := object.GetAutoscaling(***REMOVED***
	if ok {
		var minReplicas, maxReplicas int
		state.AutoScalingEnabled = types.Bool{Value: true}
		minReplicas, ok = autoscaling.GetMinReplicas(***REMOVED***
		if ok {
			state.MinReplicas = types.Int64{
				Value: int64(minReplicas***REMOVED***,
	***REMOVED***
***REMOVED***
		maxReplicas, ok = autoscaling.GetMaxReplicas(***REMOVED***
		if ok {
			state.MaxReplicas = types.Int64{
				Value: int64(maxReplicas***REMOVED***,
	***REMOVED***
***REMOVED***
	} else {
		state.MaxReplicas.Null = true
		state.MinReplicas.Null = true
	}

	instanceType, ok := object.GetInstanceType(***REMOVED***
	if ok {
		{
			state.MachineType = types.String{
				Value: instanceType,
	***REMOVED***
***REMOVED***
	}

	replicas, ok := object.GetReplicas(***REMOVED***
	if ok {
		state.Replicas = types.Int64{
			Value: int64(replicas***REMOVED***,
***REMOVED***
	}

	taints := object.Taints(***REMOVED***
	if len(taints***REMOVED*** > 0 {
		state.Taints = make([]Taints, len(taints***REMOVED******REMOVED***
		for i, taint := range taints {
			state.Taints[i] = Taints{
				Key:          types.String{Value: taint.Key(***REMOVED***},
				Value:        types.String{Value: taint.Value(***REMOVED***},
				ScheduleType: types.String{Value: taint.Effect(***REMOVED***},
	***REMOVED***
***REMOVED***

	}

	labels := object.Labels(***REMOVED***
	if labels != nil {
		state.Labels = types.Map{
			ElemType: types.StringType,
			Elems:    map[string]attr.Value{},
***REMOVED***
		for k, v := range labels {
			state.Labels.Elems[k] = types.String{
				Value: v,
	***REMOVED***
***REMOVED***

	}

}
