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

package cluster

***REMOVED***
	"context"
***REMOVED***
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"
***REMOVED***
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/errors"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

type ClusterResource struct {
	collection *cmv1.ClustersClient
}

var _ resource.ResourceWithConfigure = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

func New(***REMOVED*** resource.Resource {
	return &ClusterResource{}
}

func (r *ClusterResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "OpenShift managed cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Computed:    true,
	***REMOVED***,
			"product": schema.StringAttribute{
				Description: "Product ID OSD or Rosa",
				Required:    true,
	***REMOVED***,
			"name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Required:    true,
	***REMOVED***,
			"cloud_provider": schema.StringAttribute{
				Description: "Cloud provider identifier, for example 'aws'.",
				Required:    true,
	***REMOVED***,
			"cloud_region": schema.StringAttribute{
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Required:    true,
	***REMOVED***,
			"multi_az": schema.BoolAttribute{
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"properties": schema.MapAttribute{
				Description: "User defined properties.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"api_url": schema.StringAttribute{
				Description: "URL of the API server.",
				Computed:    true,
	***REMOVED***,
			"console_url": schema.StringAttribute{
				Description: "URL of the console.",
				Computed:    true,
	***REMOVED***,
			"compute_nodes": schema.Int64Attribute{
				Description: "Number of compute nodes of the cluster.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"compute_machine_type": schema.StringAttribute{
				Description: "Identifier of the machine type used by the compute nodes, " +
					"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
					"source to find the possible values.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"ccs_enabled": schema.BoolAttribute{
				Description: "Enables customer cloud subscription.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"aws_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account.",
				Optional:    true,
	***REMOVED***,
			"aws_access_key_id": schema.StringAttribute{
				Description: "Identifier of the AWS access key.",
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"aws_secret_access_key": schema.StringAttribute{
				Description: "AWS access key.",
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"aws_subnet_ids": schema.ListAttribute{
				Description: "AWS subnet IDs.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"aws_private_link": schema.BoolAttribute{
				Description: "Provides private connectivity between VPCs, AWS services, and your on-premises networks, without exposing your traffic to the public internet.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(***REMOVED***,
					boolplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"availability_zones": schema.ListAttribute{
				Description: "Availability zones.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(***REMOVED***,
					listplanmodifier.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"machine_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for nodes.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"proxy": schema.SingleNestedAttribute{
				Description: "proxy",
				Attributes:  proxy.ProxyResource(***REMOVED***,
				Optional:    true,
				Validators:  []validator.Object{proxy.ProxyValidator(***REMOVED***},
	***REMOVED***,
			"service_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for services.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"pod_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for pods.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"host_prefix": schema.Int64Attribute{
				Description: "Length of the prefix of the subnet assigned to each node.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"version": schema.StringAttribute{
				Description: "Identifier of the version of OpenShift, for example 'openshift-v4.1.0'.",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"state": schema.StringAttribute{
				Description: "State of the cluster.",
				Computed:    true,
	***REMOVED***,
			"wait": schema.BoolAttribute{
				Description: "Wait till the cluster is ready.",
				Optional:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (r *ClusterResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	connection, ok := req.ProviderData.(*sdk.Connection***REMOVED***
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData***REMOVED***,
		***REMOVED***
		return
	}

	r.collection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
	return
}

func createClusterObject(ctx context.Context,
	state *ClusterState, diags diag.Diagnostics***REMOVED*** (*cmv1.Cluster, error***REMOVED*** {
	// Create the cluster:
	builder := cmv1.NewCluster(***REMOVED***
	builder.Name(state.Name.ValueString(***REMOVED******REMOVED***
	builder.CloudProvider(cmv1.NewCloudProvider(***REMOVED***.ID(state.CloudProvider.ValueString(***REMOVED******REMOVED******REMOVED***
	builder.Product(cmv1.NewProduct(***REMOVED***.ID(state.Product.ValueString(***REMOVED******REMOVED******REMOVED***
	builder.Region(cmv1.NewCloudRegion(***REMOVED***.ID(state.CloudRegion.ValueString(***REMOVED******REMOVED******REMOVED***
	if common.HasValue(state.MultiAZ***REMOVED*** {
		builder.MultiAZ(state.MultiAZ.ValueBool(***REMOVED******REMOVED***
	}
	if common.HasValue(state.Properties***REMOVED*** {
		properties := map[string]string{}
		propertiesElements, err := common.OptionalMap(ctx, state.Properties***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		for k, v := range propertiesElements {
			properties[k] = v
***REMOVED***
		builder.Properties(properties***REMOVED***
	}
	nodes := cmv1.NewClusterNodes(***REMOVED***
	if common.HasValue(state.ComputeNodes***REMOVED*** {
		nodes.Compute(int(state.ComputeNodes.ValueInt64(***REMOVED******REMOVED******REMOVED***
	}
	if common.HasValue(state.ComputeMachineType***REMOVED*** {
		nodes.ComputeMachineType(
			cmv1.NewMachineType(***REMOVED***.ID(state.ComputeMachineType.ValueString(***REMOVED******REMOVED***,
		***REMOVED***
	}

	if common.HasValue(state.AvailabilityZones***REMOVED*** {
		availabilityZones, err := common.StringListToArray(ctx, state.AvailabilityZones***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		nodes.AvailabilityZones(availabilityZones...***REMOVED***
	}

	if !nodes.Empty(***REMOVED*** {
		builder.Nodes(nodes***REMOVED***
	}
	ccs := cmv1.NewCCS(***REMOVED***
	if common.HasValue(state.CCSEnabled***REMOVED*** {
		ccs.Enabled(state.CCSEnabled.ValueBool(***REMOVED******REMOVED***
	}
	if !ccs.Empty(***REMOVED*** {
		builder.CCS(ccs***REMOVED***
	}
	aws := cmv1.NewAWS(***REMOVED***
	if common.HasValue(state.AWSAccountID***REMOVED*** {
		aws.AccountID(state.AWSAccountID.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.AWSAccessKeyID***REMOVED*** {
		aws.AccessKeyID(state.AWSAccessKeyID.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.AWSSecretAccessKey***REMOVED*** {
		aws.SecretAccessKey(state.AWSSecretAccessKey.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.AWSPrivateLink***REMOVED*** {
		aws.PrivateLink(state.AWSPrivateLink.ValueBool(***REMOVED******REMOVED***
		api := cmv1.NewClusterAPI(***REMOVED***
		if state.AWSPrivateLink.ValueBool(***REMOVED*** {
			api.Listening(cmv1.ListeningMethodInternal***REMOVED***
***REMOVED***
		builder.API(api***REMOVED***
	}

	if common.HasValue(state.AWSSubnetIDs***REMOVED*** {
		subnetIds, err := common.StringListToArray(ctx, state.AWSSubnetIDs***REMOVED***
		if err != nil {
			return nil, err
***REMOVED***
		aws.SubnetIDs(subnetIds...***REMOVED***
	}

	if !aws.Empty(***REMOVED*** {
		builder.AWS(aws***REMOVED***
	}
	network := cmv1.NewNetwork(***REMOVED***
	if common.HasValue(state.MachineCIDR***REMOVED*** {
		network.MachineCIDR(state.MachineCIDR.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.ServiceCIDR***REMOVED*** {
		network.ServiceCIDR(state.ServiceCIDR.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.PodCIDR***REMOVED*** {
		network.PodCIDR(state.PodCIDR.ValueString(***REMOVED******REMOVED***
	}
	if common.HasValue(state.HostPrefix***REMOVED*** {
		network.HostPrefix(int(state.HostPrefix.ValueInt64(***REMOVED******REMOVED******REMOVED***
	}
	if !network.Empty(***REMOVED*** {
		builder.Network(network***REMOVED***
	}
	if common.HasValue(state.Version***REMOVED*** {
		builder.Version(cmv1.NewVersion(***REMOVED***.ID(state.Version.ValueString(***REMOVED******REMOVED******REMOVED***
	}

	proxyObj := cmv1.NewProxy(***REMOVED***
	if state.Proxy != nil {
		proxyObj.HTTPProxy(state.Proxy.HttpProxy.ValueString(***REMOVED******REMOVED***
		proxyObj.HTTPSProxy(state.Proxy.HttpsProxy.ValueString(***REMOVED******REMOVED***
		builder.Proxy(proxyObj***REMOVED***
	}

	object, err := builder.Build(***REMOVED***

	return object, err
}

func (r *ClusterResource***REMOVED*** Create(ctx context.Context, request resource.CreateRequest,
	response *resource.CreateResponse***REMOVED*** {
	// Get the plan:
	state := &ClusterState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	object, err := createClusterObject(ctx, state, diags***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster",
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	add, err := r.collection.Add(***REMOVED***.Body(object***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't create cluster",
			fmt.Sprintf(
				"Can't create cluster with name '%s': %v",
				state.Name.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	// Wait till the cluster is ready unless explicitly disabled:
	wait := state.Wait.IsUnknown(***REMOVED*** || state.Wait.IsNull(***REMOVED*** || state.Wait.ValueBool(***REMOVED***
	ready := object.State(***REMOVED*** == cmv1.ClusterStateReady
	if wait && !ready {
		pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour***REMOVED***
		defer cancel(***REMOVED***
		_, err := r.collection.Cluster(object.ID(***REMOVED******REMOVED***.Poll(***REMOVED***.
			Interval(30 * time.Second***REMOVED***.
			Predicate(func(get *cmv1.ClusterGetResponse***REMOVED*** bool {
				object = get.Body(***REMOVED***
				return object.State(***REMOVED*** == cmv1.ClusterStateReady
	***REMOVED******REMOVED***.
			StartContext(pollCtx***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
				"Can't poll cluster state",
				fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					object.ID(***REMOVED***, err,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
	}

	// Save the state:
	err = populateClusterState(object, state***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			***REMOVED***,
		***REMOVED***
		return
	}
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterResource***REMOVED*** Read(ctx context.Context, request resource.ReadRequest,
	response *resource.ReadResponse***REMOVED*** {
	// Get the current state:
	state := &ClusterState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the cluster:
	get, err := r.collection.Cluster(state.ID.ValueString(***REMOVED******REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Save the state:
	err = populateClusterState(object, state***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			***REMOVED***,
		***REMOVED***
		return
	}
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterResource***REMOVED*** Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse***REMOVED*** {
	var diags diag.Diagnostics

	// Get the state:
	state := &ClusterState{}
	diags = request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Get the plan:
	plan := &ClusterState{}
	diags = request.Plan.Get(ctx, plan***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send request to update the cluster:
	builder := cmv1.NewCluster(***REMOVED***
	var nodes *cmv1.ClusterNodesBuilder
	compute, ok := common.ShouldPatchInt(state.ComputeNodes, plan.ComputeNodes***REMOVED***
	if ok {
		nodes.Compute(int(compute***REMOVED******REMOVED***
	}
	if !nodes.Empty(***REMOVED*** {
		builder.Nodes(nodes***REMOVED***
	}
	patch, err := builder.Build(***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster patch",
			fmt.Sprintf(
				"Can't build patch for cluster with identifier '%s': %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	update, err := r.collection.Cluster(state.ID.ValueString(***REMOVED******REMOVED***.Update(***REMOVED***.
		Body(patch***REMOVED***.
		SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster with identifier '%s': %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := update.Body(***REMOVED***

	// Update the state:
	err = populateClusterState(object, state***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			***REMOVED***,
		***REMOVED***
		return
	}
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterResource***REMOVED*** Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse***REMOVED*** {
	// Get the state:
	state := &ClusterState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete the cluster:
	resource := r.collection.Cluster(state.ID.ValueString(***REMOVED******REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete cluster",
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Wait till the cluster has been effectively deleted:
	if state.Wait.IsUnknown(***REMOVED*** || state.Wait.IsNull(***REMOVED*** || state.Wait.ValueBool(***REMOVED*** {
		pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute***REMOVED***
		defer cancel(***REMOVED***
		_, err := resource.Poll(***REMOVED***.
			Interval(30 * time.Second***REMOVED***.
			Status(http.StatusNotFound***REMOVED***.
			StartContext(pollCtx***REMOVED***
		sdkErr, ok := err.(*errors.Error***REMOVED***
		if ok && sdkErr.Status(***REMOVED*** == http.StatusNotFound {
			err = nil
***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
				"Can't poll cluster deletion",
				fmt.Sprintf(
					"Can't poll deletion of cluster with identifier '%s': %v",
					state.ID.ValueString(***REMOVED***, err,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
	}

	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *ClusterResource***REMOVED*** ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse***REMOVED*** {
	resource.ImportStatePassthroughID(ctx, path.Root("id"***REMOVED***, request, response***REMOVED***
}

// populateClusterState copies the data from the API object to the Terraform state.
func populateClusterState(object *cmv1.Cluster, state *ClusterState***REMOVED*** error {
	state.ID = types.StringValue(object.ID(***REMOVED******REMOVED***

	object.API(***REMOVED***
	state.Product = types.StringValue(object.Product(***REMOVED***.ID(***REMOVED******REMOVED***
	state.Name = types.StringValue(object.Name(***REMOVED******REMOVED***
	state.CloudProvider = types.StringValue(object.CloudProvider(***REMOVED***.ID(***REMOVED******REMOVED***
	state.CloudRegion = types.StringValue(object.Region(***REMOVED***.ID(***REMOVED******REMOVED***
	state.MultiAZ = types.BoolValue(object.MultiAZ(***REMOVED******REMOVED***

	mapValue, err := common.ConvertStringMapToMapType(object.Properties(***REMOVED******REMOVED***
	if err != nil {
		return err
	} else {
		state.Properties = mapValue
	}

	state.APIURL = types.StringValue(object.API(***REMOVED***.URL(***REMOVED******REMOVED***
	state.ConsoleURL = types.StringValue(object.Console(***REMOVED***.URL(***REMOVED******REMOVED***
	state.ComputeNodes = types.Int64Value(int64(object.Nodes(***REMOVED***.Compute(***REMOVED******REMOVED******REMOVED***
	state.ComputeMachineType = types.StringValue(object.Nodes(***REMOVED***.ComputeMachineType(***REMOVED***.ID(***REMOVED******REMOVED***

	azs, ok := object.Nodes(***REMOVED***.GetAvailabilityZones(***REMOVED***
	if ok {
		listValue, err := common.StringArrayToList(azs***REMOVED***
		if err != nil {
			return err
***REMOVED*** else {
			state.AvailabilityZones = listValue
***REMOVED***
	}

	state.CCSEnabled = types.BoolValue(object.CCS(***REMOVED***.Enabled(***REMOVED******REMOVED***
	//The API does not return account id
	awsAccountID, ok := object.AWS(***REMOVED***.GetAccountID(***REMOVED***
	if ok {
		state.AWSAccountID = types.StringValue(awsAccountID***REMOVED***
	}
	awsAccessKeyID, ok := object.AWS(***REMOVED***.GetAccessKeyID(***REMOVED***
	if ok {
		state.AWSAccessKeyID = types.StringValue(awsAccessKeyID***REMOVED***
	}

	awsSecretAccessKey, ok := object.AWS(***REMOVED***.GetSecretAccessKey(***REMOVED***
	if ok {
		state.AWSSecretAccessKey = types.StringValue(awsSecretAccessKey***REMOVED***
	}
	awsPrivateLink, ok := object.AWS(***REMOVED***.GetPrivateLink(***REMOVED***
	if ok {
		state.AWSPrivateLink = types.BoolValue(awsPrivateLink***REMOVED***
	} else {
		state.AWSPrivateLink = types.BoolValue(true***REMOVED***
	}

	subnetIds, ok := object.AWS(***REMOVED***.GetSubnetIDs(***REMOVED***
	if ok {
		awsSubnetIds, err := common.StringArrayToList(subnetIds***REMOVED***
		if err != nil {
			return err
***REMOVED***
		state.AWSSubnetIDs = awsSubnetIds
	}

	proxyObj, ok := object.GetProxy(***REMOVED***
	if ok {
		state.Proxy.HttpProxy = types.StringValue(proxyObj.HTTPProxy(***REMOVED******REMOVED***
		state.Proxy.HttpsProxy = types.StringValue(proxyObj.HTTPSProxy(***REMOVED******REMOVED***
	}

	machineCIDR, ok := object.Network(***REMOVED***.GetMachineCIDR(***REMOVED***
	if ok {
		state.MachineCIDR = types.StringValue(machineCIDR***REMOVED***
	} else {
		state.MachineCIDR = types.StringNull(***REMOVED***
	}
	serviceCIDR, ok := object.Network(***REMOVED***.GetServiceCIDR(***REMOVED***
	if ok {
		state.ServiceCIDR = types.StringValue(serviceCIDR***REMOVED***
	} else {
		state.ServiceCIDR = types.StringNull(***REMOVED***
	}
	podCIDR, ok := object.Network(***REMOVED***.GetPodCIDR(***REMOVED***
	if ok {
		state.PodCIDR = types.StringValue(podCIDR***REMOVED***
	} else {
		state.PodCIDR = types.StringNull(***REMOVED***
	}
	hostPrefix, ok := object.Network(***REMOVED***.GetHostPrefix(***REMOVED***
	if ok {
		state.HostPrefix = types.Int64Value(int64(hostPrefix***REMOVED******REMOVED***
	} else {
		state.HostPrefix = types.Int64Null(***REMOVED***
	}
	version, ok := object.Version(***REMOVED***.GetID(***REMOVED***
	if ok {
		state.Version = types.StringValue(version***REMOVED***
	} else {
		state.Version = types.StringNull(***REMOVED***
	}
	state.State = types.StringValue(string(object.State(***REMOVED******REMOVED******REMOVED***

	return nil
}
