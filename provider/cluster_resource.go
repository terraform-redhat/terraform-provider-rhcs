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
***REMOVED***
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

type ClusterResourceType struct {
}

type ClusterResource struct {
	logger     logging.Logger
	collection *cmv1.ClustersClient
}

func (t *ClusterResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "OpenShift managed cluster.",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Description: "Unique identifier of the cluster.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"product": {
				Description: "Product ID OSD or Rosa",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"name": {
				Description: "Name of the cluster.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"cloud_provider": {
				Description: "Cloud provider identifier, for example 'aws'.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"cloud_region": {
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"sts": {
				Description: "STS Configuration",
				Attributes:  stsResource(***REMOVED***,
				Optional:    true,
	***REMOVED***,
			"multi_az": {
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"properties": {
				Description: "User defined properties.",
				Type: types.MapType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
				Computed: true,
	***REMOVED***,
			"api_url": {
				Description: "URL of the API server.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"console_url": {
				Description: "URL of the console.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"compute_nodes": {
				Description: "Number of compute nodes of the cluster.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"compute_machine_type": {
				Description: "Identifier of the machine type used by the compute nodes, " +
					"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
					"source to find the possible values.",
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"ccs_enabled": {
				Description: "Enables customer cloud subscription.",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"aws_account_id": {
				Description: "Identifier of the AWS account.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"aws_access_key_id": {
				Description: "Identifier of the AWS access key.",
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"aws_secret_access_key": {
				Description: "AWS access key.",
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"aws_subnet_ids": {
				Description: "aws subnet ids",
				Type: types.ListType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
	***REMOVED***,
			"aws_private_link": {
				Description: "aws subnet ids",
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"availability_zones": {
				Description: "availability zones",
				Type: types.ListType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
	***REMOVED***,
			"machine_cidr": {
				Description: "Block of IP addresses for nodes.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"proxy": {
				Description: "proxy",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"http_proxy": {
						Description: "http proxy",
						Type:        types.StringType,
						Required:    true,
			***REMOVED***,
					"https_proxy": {
						Description: "https proxy",
						Type:        types.StringType,
						Required:    true,
			***REMOVED***,
					"no_proxy": {
						Description: "no proxy",
						Type:        types.StringType,
						Optional:    true,
			***REMOVED***,
		***REMOVED******REMOVED***,
				Optional: true,
	***REMOVED***,
			"service_cidr": {
				Description: "Block of IP addresses for services.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"pod_cidr": {
				Description: "Block of IP addresses for pods.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"host_prefix": {
				Description: "Length of the prefix of the subnet assigned to each node.",
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"version": {
				Description: "Identifier of the version of OpenShift, for example 'openshift-v4.1.0'.",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"state": {
				Description: "State of the cluster.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"wait": {
				Description: "Wait till the cluster is ready.",
				Type:        types.BoolType,
				Optional:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *ClusterResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Create the resource:
	result = &ClusterResource{
		logger:     parent.logger,
		collection: collection,
	}

	return
}

func (r *ClusterResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	// Get the plan:
	state := &ClusterState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Create the cluster:
	builder := cmv1.NewCluster(***REMOVED***
	builder.Name(state.Name.Value***REMOVED***
	builder.CloudProvider(cmv1.NewCloudProvider(***REMOVED***.ID(state.CloudProvider.Value***REMOVED******REMOVED***
	builder.Product(cmv1.NewProduct(***REMOVED***.ID(state.Product.Value***REMOVED******REMOVED***
	builder.Region(cmv1.NewCloudRegion(***REMOVED***.ID(state.CloudRegion.Value***REMOVED******REMOVED***
	if !state.MultiAZ.Unknown && !state.MultiAZ.Null {
		builder.MultiAZ(state.MultiAZ.Value***REMOVED***
	}
	if !state.Properties.Unknown && !state.Properties.Null {
		properties := map[string]string{}
		for k, v := range state.Properties.Elems {
			properties[k] = v.(types.String***REMOVED***.Value
***REMOVED***
		builder.Properties(properties***REMOVED***
	}
	nodes := cmv1.NewClusterNodes(***REMOVED***
	if !state.ComputeNodes.Unknown && !state.ComputeNodes.Null {
		nodes.Compute(int(state.ComputeNodes.Value***REMOVED******REMOVED***
	}
	if !state.ComputeMachineType.Unknown && !state.ComputeMachineType.Null {
		nodes.ComputeMachineType(
			cmv1.NewMachineType(***REMOVED***.ID(state.ComputeMachineType.Value***REMOVED***,
		***REMOVED***
	}

	if !state.AvailabilityZones.Unknown && !state.AvailabilityZones.Null {
		azs := make([]string, 0***REMOVED***
		for _, e := range state.AvailabilityZones.Elems {
			azs = append(azs, e.(types.String***REMOVED***.Value***REMOVED***
***REMOVED***
		nodes.AvailabilityZones(azs...***REMOVED***
	}

	if !nodes.Empty(***REMOVED*** {
		builder.Nodes(nodes***REMOVED***
	}
	ccs := cmv1.NewCCS(***REMOVED***
	if !state.CCSEnabled.Unknown && !state.CCSEnabled.Null {
		ccs.Enabled(state.CCSEnabled.Value***REMOVED***
	}
	if !ccs.Empty(***REMOVED*** {
		builder.CCS(ccs***REMOVED***
	}
	aws := cmv1.NewAWS(***REMOVED***
	if !state.AWSAccountID.Unknown && !state.AWSAccountID.Null {
		aws.AccountID(state.AWSAccountID.Value***REMOVED***
	}
	if !state.AWSAccessKeyID.Unknown && !state.AWSAccessKeyID.Null {
		aws.AccessKeyID(state.AWSAccessKeyID.Value***REMOVED***
	}
	if !state.AWSSecretAccessKey.Unknown && !state.AWSSecretAccessKey.Null {
		aws.SecretAccessKey(state.AWSSecretAccessKey.Value***REMOVED***
	}
	if !state.AWSPrivateLink.Unknown && !state.AWSPrivateLink.Null {
		aws.PrivateLink((state.AWSPrivateLink.Value***REMOVED******REMOVED***
		api := cmv1.NewClusterAPI(***REMOVED***
		api.Listening(cmv1.ListeningMethodInternal***REMOVED***
		builder.API(api***REMOVED***
	}

	sts := cmv1.NewSTS(***REMOVED***
	if state.Sts != nil {
		sts.RoleARN(state.Sts.RoleARN.Value***REMOVED***
		sts.SupportRoleARN(state.Sts.SupportRoleArn.Value***REMOVED***
		instanceIamRoles := cmv1.NewInstanceIAMRoles(***REMOVED***
		instanceIamRoles.MasterRoleARN(state.Sts.InstanceIAMRoles.MasterRoleARN.Value***REMOVED***
		instanceIamRoles.WorkerRoleARN(state.Sts.InstanceIAMRoles.WorkerRoleARN.Value***REMOVED***
		sts.InstanceIAMRoles(instanceIamRoles***REMOVED***

		operatorRoles := make([]*cmv1.OperatorIAMRoleBuilder, 0***REMOVED***
		for _, operatorRole := range state.Sts.OperatorIAMRoles {
			r := cmv1.NewOperatorIAMRole(***REMOVED***
			r.Name(operatorRole.Name.Value***REMOVED***
			r.Namespace(operatorRole.Namespace.Value***REMOVED***
			r.RoleARN(operatorRole.RoleARN.Value***REMOVED***
			operatorRoles = append(operatorRoles, r***REMOVED***
***REMOVED***
		sts.OperatorIAMRoles(operatorRoles...***REMOVED***
		aws.STS(sts***REMOVED***
	}

	if !state.AWSSubnetIDs.Unknown && !state.AWSSubnetIDs.Null {
		subnetIds := make([]string, 0***REMOVED***
		for _, e := range state.AWSSubnetIDs.Elems {
			subnetIds = append(subnetIds, e.(types.String***REMOVED***.Value***REMOVED***
***REMOVED***
		aws.SubnetIDs(subnetIds...***REMOVED***
	}

	if !aws.Empty(***REMOVED*** {
		builder.AWS(aws***REMOVED***
	}
	network := cmv1.NewNetwork(***REMOVED***
	if !state.MachineCIDR.Unknown && !state.MachineCIDR.Null {
		network.MachineCIDR(state.MachineCIDR.Value***REMOVED***
	}
	if !state.ServiceCIDR.Unknown && !state.ServiceCIDR.Null {
		network.ServiceCIDR(state.ServiceCIDR.Value***REMOVED***
	}
	if !state.PodCIDR.Unknown && !state.PodCIDR.Null {
		network.PodCIDR(state.PodCIDR.Value***REMOVED***
	}
	if !state.HostPrefix.Unknown && !state.HostPrefix.Null {
		network.HostPrefix(int(state.HostPrefix.Value***REMOVED******REMOVED***
	}
	if !network.Empty(***REMOVED*** {
		builder.Network(network***REMOVED***
	}
	if !state.Version.Unknown && !state.Version.Null {
		builder.Version(cmv1.NewVersion(***REMOVED***.ID(state.Version.Value***REMOVED******REMOVED***
	}

	proxy := cmv1.NewProxy(***REMOVED***
	if state.Proxy != nil {
		proxy.HTTPProxy(state.Proxy.HttpProxy.Value***REMOVED***
		proxy.HTTPSProxy(state.Proxy.HttpsProxy.Value***REMOVED***
		builder.Proxy(proxy***REMOVED***
	}

	object, err := builder.Build(***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster",
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.Value, err,
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
				state.Name.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	// Wait till the cluster is ready unless explicitly disabled:
	wait := state.Wait.Unknown || state.Wait.Null || state.Wait.Value
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
	r.populateState(object, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	// Get the current state:
	state := &ClusterState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the cluster:
	get, err := r.collection.Cluster(state.ID.Value***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.ID.Value, err,
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

func (r *ClusterResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
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
	compute, ok := shouldPatchInt(state.ComputeNodes, plan.ComputeNodes***REMOVED***
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
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	update, err := r.collection.Cluster(state.ID.Value***REMOVED***.Update(***REMOVED***.
		Body(patch***REMOVED***.
		SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster with identifier '%s': %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := update.Body(***REMOVED***

	// Update the state:
	r.populateState(object, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	// Get the state:
	state := &ClusterState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete the cluster:
	resource := r.collection.Cluster(state.ID.Value***REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete cluster",
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Wait till the cluster has been effectively deleted:
	if state.Wait.Unknown || state.Wait.Null || state.Wait.Value {
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
					state.ID.Value, err,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
	}

	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *ClusterResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	// Try to retrieve the object:
	get, err := r.collection.Cluster(request.ID***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				request.ID, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Save the state:
	state := &ClusterState{}
	r.populateState(object, state***REMOVED***
	diags := response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

// populateState copies the data from the API object to the Terraform state.
func (r *ClusterResource***REMOVED*** populateState(object *cmv1.Cluster, state *ClusterState***REMOVED*** {
	state.ID = types.String{
		Value: object.ID(***REMOVED***,
	}
	state.Product = types.String{
		Value: object.Product(***REMOVED***.ID(***REMOVED***,
	}
	state.Name = types.String{
		Value: object.Name(***REMOVED***,
	}
	state.CloudProvider = types.String{
		Value: object.CloudProvider(***REMOVED***.ID(***REMOVED***,
	}
	state.CloudRegion = types.String{
		Value: object.Region(***REMOVED***.ID(***REMOVED***,
	}
	state.MultiAZ = types.Bool{
		Value: object.MultiAZ(***REMOVED***,
	}
	state.Properties = types.Map{
		ElemType: types.StringType,
		Elems:    map[string]attr.Value{},
	}
	for k, v := range object.Properties(***REMOVED*** {
		state.Properties.Elems[k] = types.String{
			Value: v,
***REMOVED***
	}
	state.APIURL = types.String{
		Value: object.API(***REMOVED***.URL(***REMOVED***,
	}
	state.ConsoleURL = types.String{
		Value: object.Console(***REMOVED***.URL(***REMOVED***,
	}
	state.ComputeNodes = types.Int64{
		Value: int64(object.Nodes(***REMOVED***.Compute(***REMOVED******REMOVED***,
	}
	state.ComputeMachineType = types.String{
		Value: object.Nodes(***REMOVED***.ComputeMachineType(***REMOVED***.ID(***REMOVED***,
	}

	azs, ok := object.Nodes(***REMOVED***.GetAvailabilityZones(***REMOVED***
	if ok {
		state.AvailabilityZones.Elems = make([]attr.Value, 0***REMOVED***
		for _, az := range azs {
			state.AvailabilityZones.Elems = append(state.AvailabilityZones.Elems, types.String{
				Value: az,
	***REMOVED******REMOVED***
***REMOVED***
	}

	state.CCSEnabled = types.Bool{
		Value: object.CCS(***REMOVED***.Enabled(***REMOVED***,
	}
	//The API does not return account id
	awsAccountID, ok := object.AWS(***REMOVED***.GetAccountID(***REMOVED***
	if ok {
		state.AWSAccountID = types.String{
			Value: awsAccountID,
***REMOVED***
	}
	awsAccessKeyID, ok := object.AWS(***REMOVED***.GetAccessKeyID(***REMOVED***
	if ok {
		state.AWSAccessKeyID = types.String{
			Value: awsAccessKeyID,
***REMOVED***
	} else {
		state.AWSAccessKeyID = types.String{
			Null: true,
***REMOVED***
	}
	awsSecretAccessKey, ok := object.AWS(***REMOVED***.GetSecretAccessKey(***REMOVED***
	if ok {
		state.AWSSecretAccessKey = types.String{
			Value: awsSecretAccessKey,
***REMOVED***
	} else {
		state.AWSSecretAccessKey = types.String{
			Null: true,
***REMOVED***
	}
	awsPrivateLink, ok := object.AWS(***REMOVED***.GetPrivateLink(***REMOVED***
	if ok {
		state.AWSPrivateLink = types.Bool{
			Value: awsPrivateLink,
***REMOVED***
	} else {
		state.AWSPrivateLink = types.Bool{
			Null: true,
***REMOVED***
	}
	sts, ok := object.AWS(***REMOVED***.GetSTS(***REMOVED***
	if ok {
		state.Sts = &Sts{}
		state.Sts.OIDCEndpointURL = types.String{
			Value: sts.OIDCEndpointURL(***REMOVED***,
***REMOVED***
		state.Sts.RoleARN = types.String{
			Value: sts.RoleARN(***REMOVED***,
***REMOVED***
		state.Sts.SupportRoleArn = types.String{
			Value: sts.SupportRoleARN(***REMOVED***,
***REMOVED***
		state.Sts.InstanceIAMRoles.MasterRoleARN = types.String{
			Value: sts.InstanceIAMRoles(***REMOVED***.MasterRoleARN(***REMOVED***,
***REMOVED***
		state.Sts.InstanceIAMRoles.WorkerRoleARN = types.String{
			Value: sts.InstanceIAMRoles(***REMOVED***.WorkerRoleARN(***REMOVED***,
***REMOVED***

		for _, operatorRole := range sts.OperatorIAMRoles(***REMOVED*** {
			r := OperatorIAMRole{
				Name: types.String{
					Value: operatorRole.Name(***REMOVED***,
		***REMOVED***,
				Namespace: types.String{
					Value: operatorRole.Namespace(***REMOVED***,
		***REMOVED***,
				RoleARN: types.String{
					Value: operatorRole.RoleARN(***REMOVED***,
		***REMOVED***,
	***REMOVED***
			state.Sts.OperatorIAMRoles = append(state.Sts.OperatorIAMRoles, r***REMOVED***
***REMOVED***
	}

	subnetIds, ok := object.AWS(***REMOVED***.GetSubnetIDs(***REMOVED***
	if ok {
		state.AWSSubnetIDs.Elems = make([]attr.Value, 0***REMOVED***
		for _, subnetId := range subnetIds {
			state.AWSSubnetIDs.Elems = append(state.AWSSubnetIDs.Elems, types.String{
				Value: subnetId,
	***REMOVED******REMOVED***
***REMOVED***
	}

	proxy, ok := object.GetProxy(***REMOVED***
	if ok {
		state.Proxy.HttpProxy = types.String{
			Value: proxy.HTTPProxy(***REMOVED***,
***REMOVED***
		state.Proxy.HttpsProxy = types.String{
			Value: proxy.HTTPSProxy(***REMOVED***,
***REMOVED***
	}
	machineCIDR, ok := object.Network(***REMOVED***.GetMachineCIDR(***REMOVED***
	if ok {
		state.MachineCIDR = types.String{
			Value: machineCIDR,
***REMOVED***
	} else {
		state.MachineCIDR = types.String{
			Null: true,
***REMOVED***
	}
	serviceCIDR, ok := object.Network(***REMOVED***.GetServiceCIDR(***REMOVED***
	if ok {
		state.ServiceCIDR = types.String{
			Value: serviceCIDR,
***REMOVED***
	} else {
		state.ServiceCIDR = types.String{
			Null: true,
***REMOVED***
	}
	podCIDR, ok := object.Network(***REMOVED***.GetPodCIDR(***REMOVED***
	if ok {
		state.PodCIDR = types.String{
			Value: podCIDR,
***REMOVED***
	} else {
		state.PodCIDR = types.String{
			Null: true,
***REMOVED***
	}
	hostPrefix, ok := object.Network(***REMOVED***.GetHostPrefix(***REMOVED***
	if ok {
		state.HostPrefix = types.Int64{
			Value: int64(hostPrefix***REMOVED***,
***REMOVED***
	} else {
		state.HostPrefix = types.Int64{
			Null: true,
***REMOVED***
	}
	version, ok := object.Version(***REMOVED***.GetID(***REMOVED***
	if ok {
		state.Version = types.String{
			Value: version,
***REMOVED***
	} else {
		state.Version = types.String{
			Null: true,
***REMOVED***
	}
	state.State = types.String{
		Value: string(object.State(***REMOVED******REMOVED***,
	}
}
