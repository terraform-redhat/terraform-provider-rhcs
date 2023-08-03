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

package machinepool

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"net/http"
	"regexp"
	"strings"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func ResourceMachinePool() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMachinePoolCreate,
		ReadContext:   resourceMachinePoolRead,
		UpdateContext: resourceMachinePoolUpdate,
		DeleteContext: resourceMachinePoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceMachinePoolImport,
		},
		Schema: MachinePoolFields(),
	}
}

var machinepoolNameRE = regexp.MustCompile(
	`^[a-z]([-a-z0-9]*[a-z0-9])?$`,
)

func resourceMachinePoolCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin Creating")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	machinePoolState := machinePoolFromResourceData(resourceData)

	// Wait till the cluster is ready:
	if err := common2.WaitTillClusterIsReadyOrFail(ctx, clusterCollection, machinePoolState.Cluster); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't poll cluster state",
				Detail: fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					machinePoolState.Cluster, err),
			}}
	}

	// Create the machine pool:
	builder := cmv1.NewMachinePool().ID(machinePoolState.ID).InstanceType(machinePoolState.MachineType)
	builder.ID(machinePoolState.Name)

	if err := setSpotInstances(machinePoolState, builder); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't build machine pool",
				Detail: fmt.Sprintf(
					"Can't build machine pool for cluster '%s: %v'", machinePoolState.Cluster, err,
				),
			}}
	}

	isMultiAZPool, err := validateAZConfig(machinePoolState, clusterCollection)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't build machine pool",
				Detail: fmt.Sprintf(
					"Can't build machine pool for cluster '%s': %v",
					machinePoolState.Cluster, err,
				),
			}}
	}
	if !common2.IsStringAttributeEmpty(machinePoolState.AvailabilityZone) {
		builder.AvailabilityZones(*machinePoolState.AvailabilityZone)
	}
	if !common2.IsStringAttributeEmpty(machinePoolState.SubnetID) {
		builder.Subnets(*machinePoolState.SubnetID)
	}

	computeNodeEnabled := false
	autoscalingEnabled, errMsg := getAutoscaling(machinePoolState, builder)
	if errMsg != "" {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't build machine pool",
				Detail: fmt.Sprintf(
					"Can't build machine pool for cluster '%s, %s'", machinePoolState.Cluster, errMsg,
				),
			}}
	}

	if machinePoolState.Replicas != nil {
		computeNodeEnabled = true
		if isMultiAZPool && *machinePoolState.Replicas%3 != 0 {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Can't build machine pool",
					Detail: fmt.Sprintf(
						"Can't build machine pool for cluster '%s', replicas must be a multiple of 3",
						machinePoolState.Cluster,
					),
				}}
		}
		builder.Replicas(*machinePoolState.Replicas)
	}
	if (!autoscalingEnabled && !computeNodeEnabled) || (autoscalingEnabled && computeNodeEnabled) {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't build machine pool",
				Detail: fmt.Sprintf(
					"Can't build machine pool for cluster '%s', please provide a value for either the 'replicas' or 'autoscaling_enabled' parameter. It is mandatory to include at least one of these parameters in the resource plan.",
					machinePoolState.Cluster,
				),
			}}
	}

	if machinePoolState.Taints != nil && len(machinePoolState.Taints) > 0 {
		taintBuilders := taintBuilder(machinePoolState.Taints)
		builder.Taints(taintBuilders...)
	}

	if machinePoolState.Labels != nil {
		builder.Labels(machinePoolState.Labels)
	}

	object, err := builder.Build()
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't build machine pool",
				Detail: fmt.Sprintf(
					"Can't build machine pool for cluster '%s': %v",
					machinePoolState.Cluster, err,
				),
			}}
	}

	collection := clusterCollection.Cluster(machinePoolState.Cluster).MachinePools()
	add, err := collection.Add().Body(object).SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't create machine pool",
				Detail: fmt.Sprintf(
					"Can't create machine pool for cluster '%s': %v",
					machinePoolState.Cluster, err,
				),
			}}
	}
	object = add.Body()
	machinePoolToResourceData(object, resourceData, machinePoolState)
	return nil
}

func taintBuilder(taints []Taint) []*cmv1.TaintBuilder {
	var taintBuilders []*cmv1.TaintBuilder
	for _, taint := range taints {
		taintBuilders = append(taintBuilders, cmv1.NewTaint().Key(taint.Key).Value(taint.Value).Effect(taint.ScheduleType))
	}
	return taintBuilders
}

func resourceMachinePoolRead(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin Creating")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	machinePoolState := machinePoolFromResourceData(resourceData)

	// Find the machine pool:
	resource := clusterCollection.Cluster(machinePoolState.Cluster).
		MachinePools().
		MachinePool(machinePoolState.ID)
	get, err := resource.Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf(fmt.Sprintf("machine pool (%s) of cluster (%s) not found, removing from state",
			machinePoolState.ID, machinePoolState.Cluster,
		)))
		resourceData.SetId("")
		return
	} else if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Failed to fetch machine pool",
				Detail: fmt.Sprintf(
					"Failed to fetch machine pool with identifier %s for cluster %s. Response code: %v",
					machinePoolState.ID, machinePoolState.Cluster, get.Status(),
				),
			}}
	}
	object := get.Body()
	machinePoolToResourceData(object, resourceData, machinePoolState)
	return nil
}

func resourceMachinePoolUpdate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin Creating")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	machinePoolState := machinePoolFromResourceData(resourceData)

	// Find the machine pool:
	resource := clusterCollection.Cluster(machinePoolState.Cluster).
		MachinePools().
		MachinePool(machinePoolState.ID)
	_, err := resource.Get().SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't find machine pool",
				Detail: fmt.Sprintf(
					"Can't find machine pool with identifier '%s' for "+
						"cluster '%s': %v",
					machinePoolState.ID, machinePoolState.Cluster, err,
				),
			}}
	}

	mpBuilder := cmv1.NewMachinePool().ID(machinePoolState.ID)

	if resourceData.HasChange("machine_type") {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't update machine pool",
				Detail: fmt.Sprintf(
					"Can't update machine pool for cluster '%s', machine type cannot be updated",
					machinePoolState.Cluster,
				),
			}}
	}
	computeNodesEnabled := false

	if machinePoolState.Replicas != nil {
		computeNodesEnabled = true
		mpBuilder.Replicas(*machinePoolState.Replicas)
	}

	autoscalingEnabled, errMsg := getAutoscaling(machinePoolState, mpBuilder)
	if errMsg != "" {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't update machine pool",
				Detail: fmt.Sprintf(
					"Can't update machine pool for cluster '%s, %s ", machinePoolState.Cluster, errMsg,
				),
			}}
	}

	if (autoscalingEnabled && computeNodesEnabled) || (!autoscalingEnabled && !computeNodesEnabled) {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't update machine pool",
				Detail: fmt.Sprintf(
					"Can't update machine pool for cluster '%s: either autoscaling or compute nodes should be enabled", machinePoolState.Cluster,
				),
			}}
	}

	if resourceData.HasChange("labels") {
		_, newV := resourceData.GetChange("labels")
		if newV != nil && len(newV.(map[string]interface{})) > 0 {
			labels := common2.ExpandStringValueMap(newV.(map[string]interface{}))
			mpBuilder.Labels(labels)
		}

	}

	if resourceData.HasChange("taints") {
		_, newV := resourceData.GetChange("taints")
		if newV != nil {
			taints := ExpandTaintsFromInterface(newV)
			taintBuilders := taintBuilder(taints)
			mpBuilder.Taints(taintBuilders...)
		}
	}

	machinePool, err := mpBuilder.Build()
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't update machine pool",
				Detail: fmt.Sprintf(
					"Can't update machine pool for cluster '%s: %v ", machinePoolState.Cluster, err,
				),
			}}
	}
	update, err := clusterCollection.Cluster(machinePoolState.Cluster).
		MachinePools().
		MachinePool(machinePoolState.ID).Update().Body(machinePool).SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Failed to update machine pool",
				Detail: fmt.Sprintf(
					"Failed to update machine pool '%s'  on cluster '%s': %v",
					machinePoolState.ID, machinePoolState.Cluster, err,
				),
			}}
	}

	object := update.Body()

	machinePoolToResourceData(object, resourceData, machinePoolState)
	return nil
}

func resourceMachinePoolDelete(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin Creating")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	machinePoolState := machinePoolFromResourceData(resourceData)

	resource := clusterCollection.Cluster(machinePoolState.Cluster).
		MachinePools().
		MachinePool(machinePoolState.ID)

	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't delete machine pool",
				Detail: fmt.Sprintf(
					"Can't delete machine pool with identifier '%s' for "+
						"cluster '%s': %v",
					machinePoolState.ID, machinePoolState.Cluster, err,
				),
			}}
	}

	// Remove the state:
	resourceData.SetId("")

	return nil
}

func resourceMachinePoolImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// To import a machine pool, we need to know the cluster ID and the machine pool ID
	fields := strings.Split(d.Id(), ",")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		return nil, fmt.Errorf("invalid import identifier (%s), "+
			"Machine pool to import should be specified as <cluster_id>,<machine_pool_id>", d.Id())
	}

	clusterID := fields[0]
	machinePoolID := fields[1]

	d.SetId(machinePoolID)
	d.Set("cluster", clusterID)

	return []*schema.ResourceData{d}, nil
}

func machinePoolFromResourceData(resourceData *schema.ResourceData) *MachinePoolState {
	machinePool := MachinePoolState{
		Cluster:     resourceData.Get("cluster").(string),
		MachineType: resourceData.Get("machine_type").(string),
		Name:        resourceData.Get("name").(string),
		ID:          resourceData.Id(),
	}

	machinePool.Replicas = common2.GetOptionalInt(resourceData, "replicas")
	machinePool.UseSpotInstances = common2.GetOptionalBool(resourceData, "use_spot_instances")
	machinePool.MaxSpotPrice = common2.GetOptionalFloat(resourceData, "max_spot_price")
	machinePool.Taints = ExpandTaintsFromResourceData(resourceData)
	machinePool.Labels = common2.GetOptionalMapStringFromResourceData(resourceData, "labels")
	machinePool.MultiAvailabilityZone = common2.GetOptionalBool(resourceData, "multi_availability_zone")
	machinePool.AvailabilityZone = common2.GetOptionalString(resourceData, "availability_zone")
	machinePool.SubnetID = common2.GetOptionalString(resourceData, "subnet_id")
	machinePool.AutoScalingEnabled = common2.GetOptionalBool(resourceData, "autoscaling_enabled")

	// in the new sdk V2 there is no concept of null and if the value was set to zero it will not be returned
	// the validation if the attribute was set is part of the setting the `RequiredWith` flag on the attribute
	if machinePool.AutoScalingEnabled != nil && *machinePool.AutoScalingEnabled {
		machinePool.MinReplicas = common2.Pointer(resourceData.Get("min_replicas").(int))
		machinePool.MaxReplicas = common2.Pointer(resourceData.Get("max_replicas").(int))
	}

	return &machinePool
}

func machinePoolToResourceData(object *cmv1.MachinePool, resourceData *schema.ResourceData, machinePoolState *MachinePoolState) {
	resourceData.SetId(object.ID())
	resourceData.Set("name", object.ID())

	if getAWS, ok := object.GetAWS(); ok {
		if spotMarketOptions, ok := getAWS.GetSpotMarketOptions(); ok {
			resourceData.Set("use_spot_instances", true)
			if spotMarketOptions.MaxPrice() != 0 {
				resourceData.Set("max_spot_price", float64(spotMarketOptions.MaxPrice()))
			}
		}
	}

	autoscaling, ok := object.GetAutoscaling()
	if ok {
		var minReplicas, maxReplicas int
		resourceData.Set("autoscaling_enabled", true)
		minReplicas, ok = autoscaling.GetMinReplicas()
		if ok {
			resourceData.Set("min_replicas", minReplicas)
		}
		maxReplicas, ok = autoscaling.GetMaxReplicas()
		if ok {
			resourceData.Set("max_replicas", maxReplicas)
		}
	}

	instanceType, ok := object.GetInstanceType()
	if ok {
		resourceData.Set("machine_type", instanceType)
	}

	replicas, ok := object.GetReplicas()
	if ok {
		resourceData.Set("replicas", replicas)
	}

	resourceData.Set("taints", FlatTaints(object))

	labels := object.Labels()
	if len(labels) > 0 {
		resourceData.Set("labels", labels)
	}

	if machinePoolState.MultiAvailabilityZone != nil {
		resourceData.Set("multi_availability_zone", *machinePoolState.MultiAvailabilityZone)
	}
	if machinePoolState.AvailabilityZone != nil {
		resourceData.Set("availability_zone", *machinePoolState.AvailabilityZone)
	}
	if machinePoolState.SubnetID != nil {
		resourceData.Set("subnet_id", *machinePoolState.SubnetID)
	}
	return
}

// Validate the machine pool's settings that pertain to availability zones.
// Returns whether the machine pool is/will be multi-AZ.
func validateAZConfig(state *MachinePoolState, clusterCollection *cmv1.ClustersClient) (bool, error) {
	resp, err := clusterCollection.Cluster(state.Cluster).Get().Send()
	if err != nil {
		return false, fmt.Errorf("failed to get information for cluster %s: %v", state.Cluster, err)
	}
	cluster := resp.Body()
	isMultiAZCluster := cluster.MultiAZ()
	clusterAZs := cluster.Nodes().AvailabilityZones()
	clusterSubnets := cluster.AWS().SubnetIDs()

	if isMultiAZCluster {
		// Can't set both availability_zone and subnet_id
		if !common2.IsStringAttributeEmpty(state.AvailabilityZone) && !common2.IsStringAttributeEmpty(state.SubnetID) {
			return false, fmt.Errorf("availability_zone and subnet_id are mutually exclusive")
		}

		// multi_availability_zone setting must be consistent with availability_zone and subnet_id
		azOrSubnet := !common2.IsStringAttributeEmpty(state.AvailabilityZone) || !common2.IsStringAttributeEmpty(state.SubnetID)
		if state.MultiAvailabilityZone != nil {
			if azOrSubnet && *state.MultiAvailabilityZone {
				return false, fmt.Errorf("multi_availability_zone must be False when availability_zone or subnet_id is set")
			}
		} else {
			state.MultiAvailabilityZone = common2.Pointer(!azOrSubnet)
		}
	} else { // not a multi-AZ cluster
		if !common2.IsStringAttributeEmpty(state.AvailabilityZone) {
			return false, fmt.Errorf("availability_zone can only be set for multi-AZ clusters")
		}
		if !common2.IsStringAttributeEmpty(state.SubnetID) {
			return false, fmt.Errorf("subnet_id can only be set for multi-AZ clusters")
		}
		if state.MultiAvailabilityZone != nil && *state.MultiAvailabilityZone {
			return false, fmt.Errorf("multi_availability_zone can only be set for multi-AZ clusters")
		}
		state.MultiAvailabilityZone = common2.Pointer(false)
	}

	// Ensure that the machine pool's AZ and subnet are valid for the cluster
	// If subnet is set, we make sure it's valid for the cluster, but we don't default it if not set
	if !common2.IsStringAttributeEmpty(state.SubnetID) {
		inClusterSubnet := false
		for _, subnet := range clusterSubnets {
			if subnet == *state.SubnetID {
				inClusterSubnet = true
				break
			}
		}
		if !inClusterSubnet {
			return false, fmt.Errorf("subnet_id %s is not valid for cluster %s", *state.SubnetID, state.Cluster)
		}
	}

	// If AZ is set, we make sure it's valid for the cluster. If not set and neither is subnet, we default it to the 1st AZ in the cluster
	if !common2.IsStringAttributeEmpty(state.AvailabilityZone) {
		inClusterAZ := false
		for _, az := range clusterAZs {
			if az == *state.AvailabilityZone {
				inClusterAZ = true
				break
			}
		}
		if !inClusterAZ {
			return false, fmt.Errorf("availability_zone %s is not valid for cluster %s", *state.AvailabilityZone, state.Cluster)
		}
	} else {
		if len(clusterAZs) > 0 && !*state.MultiAvailabilityZone && isMultiAZCluster && common2.IsStringAttributeEmpty(state.SubnetID) {
			state.AvailabilityZone = common2.Pointer(clusterAZs[0])
		}
	}

	return *state.MultiAvailabilityZone, nil
}

func setSpotInstances(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder) error {
	useSpotInstances := state.UseSpotInstances != nil && *state.UseSpotInstances
	isSpotMaxPriceSet := state.MaxSpotPrice != nil

	if isSpotMaxPriceSet && !useSpotInstances {
		return errors.New("Can't set max price when not using spot instances (set \"use_spot_instances\" to true)")
	}

	if useSpotInstances {
		if isSpotMaxPriceSet && *state.MaxSpotPrice <= 0 {
			return errors.New("To use Spot instances, you must set \"max_spot_price\" with positive value")
		}

		awsMachinePool := cmv1.NewAWSMachinePool()
		spotMarketOptions := cmv1.NewAWSSpotMarketOptions()
		if isSpotMaxPriceSet {
			spotMarketOptions.MaxPrice(*state.MaxSpotPrice)
		}
		awsMachinePool.SpotMarketOptions(spotMarketOptions)
		mpBuilder.AWS(awsMachinePool)
	}

	return nil
}

func getAutoscaling(state *MachinePoolState, mpBuilder *cmv1.MachinePoolBuilder) (
	autoscalingEnabled bool, errMsg string) {
	autoscalingEnabled = false
	if state.AutoScalingEnabled != nil && *state.AutoScalingEnabled {
		autoscalingEnabled = true

		autoscaling := cmv1.NewMachinePoolAutoscaling()
		if state.MaxReplicas != nil {
			autoscaling.MaxReplicas(*state.MaxReplicas)
		} else {
			return false, "when enabling autoscaling, should set value for maxReplicas"
		}
		if state.MinReplicas != nil {
			autoscaling.MinReplicas(*state.MinReplicas)
		} else {
			return false, "when enabling autoscaling, should set value for minReplicas"
		}
		if !autoscaling.Empty() {
			mpBuilder.Autoscaling(autoscaling)
		}
	} else {
		if state.MaxReplicas != nil || state.MinReplicas != nil {
			return false, "when disabling autoscaling, can't set min_replicas and/or max_replicas"
		}
	}

	return autoscalingEnabled, ""
}
