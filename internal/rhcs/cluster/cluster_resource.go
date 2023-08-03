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

package cluster

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	clusterschema2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/cluster/clusterschema"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/errors"
)

func ResourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: clusterschema2.ClusterFields(),
	}
}

func resourceClusterCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	clusterState := clusterFromResourceData(resourceData)
	object, err := CreateClusterObject(clusterState)
	if err != nil {
		if err != nil {
			return diag.Errorf(
				fmt.Sprintf(
					"Can't build cluster with name '%s': %v",
					resourceData.Get("name").(string), err,
				))
		}
	}
	add, err := clusterCollection.Add().Body(object).SendContext(ctx)
	if err != nil {
		return diag.Errorf(
			fmt.Sprintf(
				"Can't create cluster with name '%s': %v",
				resourceData.Get("name").(string), err,
			))
	}
	object = add.Body()
	// Wait till the cluster is ready unless explicitly disabled:
	if err = common2.WaitTillClusterIsReadyOrFail(ctx, clusterCollection, object.ID()); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't poll cluster state",
				Detail: fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					object.ID(), err,
				),
			}}
	}
	// Save the state:
	resourceData.SetId(object.ID())
	clusterToResourceData(object, resourceData)
	return
}

func clusterToResourceData(object *cmv1.Cluster, resourceData *schema.ResourceData) {
	resourceData.SetId(object.ID())
	resourceData.Set("product", object.Product().ID())
	resourceData.Set("name", object.Name())
	resourceData.Set("cloud_provider", object.CloudProvider().ID())
	resourceData.Set("cloud_region", object.Region().ID())

	resourceData.Set("multi_az", object.MultiAZ())
	resourceData.Set("properties", object.Properties())
	resourceData.Set("api_url", object.API().URL())
	resourceData.Set("console_url", object.Console().URL())
	resourceData.Set("compute_nodes", object.Nodes().Compute())
	resourceData.Set("compute_machine_type", object.Nodes().ComputeMachineType().ID())
	resourceData.Set("ccs_enabled", object.CCS().Enabled())
	resourceData.Set("state", object.State())

	awsAccountID, ok := object.AWS().GetAccountID()
	if ok {
		resourceData.Set("aws_account_id", awsAccountID)
	}

	awsAccessKeyID, ok := object.AWS().GetAccessKeyID()
	if ok {
		resourceData.Set("aws_access_key_id", awsAccessKeyID)
	}

	awsSecretAccessKey, ok := object.AWS().GetSecretAccessKey()
	if ok {
		resourceData.Set("aws_secret_access_key", awsSecretAccessKey)
	}

	awsPrivateLink, ok := object.AWS().GetPrivateLink()
	if ok {
		resourceData.Set("aws_private_link", awsPrivateLink)
	}

	machineCIDR, ok := object.Network().GetMachineCIDR()
	if ok {
		resourceData.Set("machine_cidr", machineCIDR)

	}

	serviceCIDR, ok := object.Network().GetServiceCIDR()
	if ok {
		resourceData.Set("service_cidr", serviceCIDR)
	}

	podCIDR, ok := object.Network().GetPodCIDR()
	if ok {
		resourceData.Set("pod_cidr", podCIDR)
	}

	hostPrefix, ok := object.Network().GetHostPrefix()
	if ok {
		resourceData.Set("host_prefix", hostPrefix)
	}

	version, ok := object.Version().GetID()
	if ok {
		resourceData.Set("version", version)
	}

	resourceData.Set("proxy", clusterschema2.FlatProxy(object, resourceData))

	if subnetIds, ok := object.AWS().GetSubnetIDs(); ok && len(subnetIds) > 0 {
		resourceData.Set("aws_subnet_ids", subnetIds)
	}

	if azs, ok := object.Nodes().GetAvailabilityZones(); ok && len(azs) > 0 {
		resourceData.Set("availability_zones", azs)
	}
}

func clusterFromResourceData(resourceData *schema.ResourceData) *clusterschema2.ClusterState {
	result := &clusterschema2.ClusterState{
		Product:       resourceData.Get("product").(string),
		Name:          resourceData.Get("name").(string),
		CloudProvider: resourceData.Get("cloud_provider").(string),
		CloudRegion:   resourceData.Get("cloud_region").(string),
		ID:            resourceData.Id(),
	}

	// optional attributes
	result.APIURL = common2.GetOptionalString(resourceData, "api_url")
	result.AWSAccessKeyID = common2.GetOptionalString(resourceData, "aws_access_key_id")
	result.AWSAccountID = common2.GetOptionalString(resourceData, "aws_account_id")
	result.AWSSecretAccessKey = common2.GetOptionalString(resourceData, "aws_secret_access_key")
	result.AWSSubnetIDs = common2.GetOptionalListOfValueStringsFromResourceData(resourceData, "aws_subnet_ids")
	result.AWSPrivateLink = common2.GetOptionalBool(resourceData, "aws_private_link")
	result.CCSEnabled = common2.GetOptionalBool(resourceData, "ccs_enabled")
	result.ComputeMachineType = common2.GetOptionalString(resourceData, "compute_machine_type")
	result.ConsoleURL = common2.GetOptionalString(resourceData, "console_url")
	result.ComputeNodes = common2.GetOptionalInt(resourceData, "compute_nodes")
	result.HostPrefix = common2.GetOptionalInt(resourceData, "host_prefix")
	result.MachineCIDR = common2.GetOptionalString(resourceData, "machine_cidr")
	result.MultiAZ = common2.GetOptionalBool(resourceData, "multi_az")
	result.AvailabilityZones = common2.GetOptionalListOfValueStringsFromResourceData(resourceData, "availability_zones")
	result.PodCIDR = common2.GetOptionalString(resourceData, "pod_cidr")
	result.ServiceCIDR = common2.GetOptionalString(resourceData, "service_cidr")
	result.Properties = common2.GetOptionalMapStringFromResourceData(resourceData, "properties")
	result.Proxy = clusterschema2.ExpandProxyFromResourceData(resourceData)
	result.Version = common2.GetOptionalString(resourceData, "version")
	result.Wait = common2.GetOptionalBool(resourceData, "wait")

	if state := common2.GetOptionalString(resourceData, "state"); state != nil {
		result.State = *state
	}

	return result
}

func CreateClusterObject(state *clusterschema2.ClusterState) (*cmv1.Cluster, error) {
	// Create the cluster:
	builder := cmv1.NewCluster()
	builder.Name(state.Name)
	builder.CloudProvider(cmv1.NewCloudProvider().ID(state.CloudProvider))
	builder.Product(cmv1.NewProduct().ID(state.Product))
	builder.Region(cmv1.NewCloudRegion().ID(state.CloudRegion))

	if state.MultiAZ != nil {
		builder.MultiAZ(*state.MultiAZ)
	}

	nodes := cmv1.NewClusterNodes()
	if state.ComputeNodes != nil {
		nodes.Compute(*state.ComputeNodes)
	}
	if state.ComputeMachineType != nil {
		nodes.ComputeMachineType(
			cmv1.NewMachineType().ID(*state.ComputeMachineType),
		)
	}
	if !nodes.Empty() {
		builder.Nodes(nodes)
	}
	ccs := cmv1.NewCCS()
	if state.CCSEnabled != nil {
		ccs.Enabled(*state.CCSEnabled)
	}
	if !ccs.Empty() {
		builder.CCS(ccs)
	}
	aws := cmv1.NewAWS()
	if state.AWSAccountID != nil {
		aws.AccountID(*state.AWSAccountID)
	}
	if state.AWSAccessKeyID != nil {
		aws.AccessKeyID(*state.AWSAccessKeyID)
	}
	if state.AWSSecretAccessKey != nil {
		aws.SecretAccessKey(*state.AWSSecretAccessKey)
	}
	if state.AWSPrivateLink != nil {
		aws.PrivateLink(*state.AWSPrivateLink)
		api := cmv1.NewClusterAPI()
		if *state.AWSPrivateLink {
			api.Listening(cmv1.ListeningMethodInternal)
		}
		builder.API(api)
	}

	if !aws.Empty() {
		builder.AWS(aws)
	}
	network := cmv1.NewNetwork()
	if state.MachineCIDR != nil {
		network.MachineCIDR(*state.MachineCIDR)
	}
	if state.ServiceCIDR != nil {
		network.ServiceCIDR(*state.ServiceCIDR)
	}
	if state.PodCIDR != nil {
		network.PodCIDR(*state.PodCIDR)
	}
	if state.HostPrefix != nil {
		network.HostPrefix(*state.HostPrefix)
	}
	if !network.Empty() {
		builder.Network(network)
	}
	if state.Version != nil {
		builder.Version(cmv1.NewVersion().ID(*state.Version))
	}

	if state.Properties != nil {
		builder.Properties(state.Properties)
	}

	proxy := cmv1.NewProxy()
	if state.Proxy != nil {
		proxy.HTTPProxy(*state.Proxy.HttpProxy)
		proxy.HTTPSProxy(*state.Proxy.HttpsProxy)
		builder.Proxy(proxy)
	}

	if state.AvailabilityZones != nil && len(state.AvailabilityZones) > 0 {
		nodes.AvailabilityZones(state.AvailabilityZones...)
	}

	if state.AWSSubnetIDs != nil {
		aws.SubnetIDs(state.AWSSubnetIDs...)
	}

	object, err := builder.Build()

	return object, err
}

func resourceClusterRead(ctx context.Context, resourceData *schema.ResourceData,
	meta interface{}) (diags diag.Diagnostics) {
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	// Find the cluster:
	get, err := clusterCollection.Cluster(resourceData.Id()).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		resourceID := resourceData.Id()
		summary := fmt.Sprintf("cluster (%s) not found, removing from state", resourceID)
		tflog.Warn(ctx, summary)
		resourceData.SetId("")
		return []diag.Diagnostic{
			{
				Severity: diag.Warning,
				Summary:  summary,
				Detail: fmt.Sprintf(
					"cluster (%s) not found, removing from state",
					resourceID),
			}}
	} else if err != nil {
		return diag.Errorf(
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				resourceData.Id(), err,
			))
	}
	object := get.Body()

	clusterToResourceData(object, resourceData)
	return
}

func resourceClusterUpdate(ctx context.Context, resourceData *schema.ResourceData,
	meta interface{}) (diags diag.Diagnostics) {
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()

	// Send request to update the cluster:
	builder := cmv1.NewCluster()
	if resourceData.HasChange("compute_nodes") {
		var nodes *cmv1.ClusterNodesBuilder
		_, newV := resourceData.GetChange("compute_nodes")
		nodes.Compute(newV.(int))
		builder.Nodes(nodes)
	}

	patch, err := builder.Build()
	if err != nil {
		return diag.Errorf(
			fmt.Sprintf(
				"Can't build patch for cluster with identifier '%s': %v",
				resourceData.Id(), err,
			))
	}
	update, err := clusterCollection.Cluster(resourceData.Id()).Update().
		Body(patch).
		SendContext(ctx)
	if err != nil {
		return diag.Errorf(
			fmt.Sprintf(
				"Can't update cluster with identifier '%s': %v",
				resourceData.Id(), err,
			))
	}

	object := update.Body()
	clusterToResourceData(object, resourceData)
	return
}

func resourceClusterDelete(ctx context.Context, resourceData *schema.ResourceData,
	meta interface{}) (diags diag.Diagnostics) {
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()

	// Send the request to delete the cluster:
	resource := clusterCollection.Cluster(resourceData.Id())
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		return diag.Errorf(
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				resourceData.Id(), err,
			))
	}

	// Wait till the cluster has been effectively deleted:
	wait := common2.GetOptionalBool(resourceData, "wait")
	if wait != nil && *wait {
		pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		_, err := resource.Poll().
			Interval(30 * time.Second).
			Status(http.StatusNotFound).
			StartContext(pollCtx)
		sdkErr, ok := err.(*errors.Error)
		if ok && sdkErr.Status() == http.StatusNotFound {
			err = nil
		}
		if err != nil {
			return diag.Errorf(
				fmt.Sprintf(
					"Can't poll deletion of cluster with identifier '%s': %v",
					resourceData.Id(), err,
				),
			)
		}
	}

	resourceData.SetId("")
	return
}
