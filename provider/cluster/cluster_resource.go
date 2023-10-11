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
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/proxy"
	"net/http"
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
)

type ClusterResource struct {
	collection *cmv1.ClustersClient
}

var _ resource.ResourceWithConfigure = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

func New() resource.Resource {
	return &ClusterResource{}
}

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "OpenShift managed cluster.",
		DeprecationMessage: fmt.Sprintf(
			"using cluster as a resource is deprecated; consider using the cluster_rosa_classic resource instead"),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Computed:    true,
			},
			"product": schema.StringAttribute{
				Description: "Product ID OSD or Rosa",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Required:    true,
			},
			"cloud_provider": schema.StringAttribute{
				Description: "Cloud provider identifier, for example 'aws'.",
				Required:    true,
			},
			"cloud_region": schema.StringAttribute{
				Description: "Cloud region identifier, for example 'us-east-1'.",
				Required:    true,
			},
			"multi_az": schema.BoolAttribute{
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"properties": schema.MapAttribute{
				Description: "User defined properties.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"api_url": schema.StringAttribute{
				Description: "URL of the API server.",
				Computed:    true,
			},
			"console_url": schema.StringAttribute{
				Description: "URL of the console.",
				Computed:    true,
			},
			"compute_nodes": schema.Int64Attribute{
				Description: "Number of compute nodes of the cluster.",
				Optional:    true,
				Computed:    true,
			},
			"compute_machine_type": schema.StringAttribute{
				Description: "Identifier of the machine type used by the compute nodes, " +
					"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
					"source to find the possible values.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ccs_enabled": schema.BoolAttribute{
				Description: "Enables customer cloud subscription.",
				Optional:    true,
				Computed:    true,
			},
			"aws_account_id": schema.StringAttribute{
				Description: "Identifier of the AWS account.",
				Optional:    true,
			},
			"aws_access_key_id": schema.StringAttribute{
				Description: "Identifier of the AWS access key.",
				Optional:    true,
				Sensitive:   true,
			},
			"aws_secret_access_key": schema.StringAttribute{
				Description: "AWS access key.",
				Optional:    true,
				Sensitive:   true,
			},
			"aws_subnet_ids": schema.ListAttribute{
				Description: "AWS subnet IDs.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"aws_private_link": schema.BoolAttribute{
				Description: "Provides private connectivity between VPCs, AWS services, and your on-premises networks, without exposing your traffic to the public internet.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"availability_zones": schema.ListAttribute{
				Description: "Availability zones.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					listplanmodifier.RequiresReplace(),
				},
			},
			"machine_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for nodes.",
				Optional:    true,
				Computed:    true,
			},
			"proxy": schema.SingleNestedAttribute{
				Description: "proxy",
				Attributes:  proxy.ProxyResource(),
				Optional:    true,
				Validators:  []validator.Object{proxy.ProxyValidator()},
			},
			"service_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for services.",
				Optional:    true,
				Computed:    true,
			},
			"pod_cidr": schema.StringAttribute{
				Description: "Block of IP addresses for pods.",
				Optional:    true,
				Computed:    true,
			},
			"host_prefix": schema.Int64Attribute{
				Description: "Length of the prefix of the subnet assigned to each node.",
				Optional:    true,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Identifier of the version of OpenShift, for example 'openshift-v4.1.0'.",
				Optional:    true,
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "State of the cluster.",
				Computed:    true,
			},
			"wait": schema.BoolAttribute{
				Description: "Wait till the cluster is ready.",
				Optional:    true,
			},
		},
	}
	return
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

	r.collection = connection.ClustersMgmt().V1().Clusters()
	return
}

func createClusterObject(ctx context.Context,
	state *ClusterState, diags diag.Diagnostics) (*cmv1.Cluster, error) {
	// Create the cluster:
	builder := cmv1.NewCluster()
	builder.Name(state.Name.ValueString())
	builder.CloudProvider(cmv1.NewCloudProvider().ID(state.CloudProvider.ValueString()))
	builder.Product(cmv1.NewProduct().ID(state.Product.ValueString()))
	builder.Region(cmv1.NewCloudRegion().ID(state.CloudRegion.ValueString()))
	if common.HasValue(state.MultiAZ) {
		builder.MultiAZ(state.MultiAZ.ValueBool())
	}
	if common.HasValue(state.Properties) {
		properties := map[string]string{}
		propertiesElements, err := common.OptionalMap(ctx, state.Properties)
		if err != nil {
			return nil, err
		}
		for k, v := range propertiesElements {
			properties[k] = v
		}
		builder.Properties(properties)
	}
	nodes := cmv1.NewClusterNodes()
	if common.HasValue(state.ComputeNodes) {
		nodes.Compute(int(state.ComputeNodes.ValueInt64()))
	}
	if common.HasValue(state.ComputeMachineType) {
		nodes.ComputeMachineType(
			cmv1.NewMachineType().ID(state.ComputeMachineType.ValueString()),
		)
	}

	if common.HasValue(state.AvailabilityZones) {
		availabilityZones, err := common.StringListToArray(ctx, state.AvailabilityZones)
		if err != nil {
			return nil, err
		}
		nodes.AvailabilityZones(availabilityZones...)
	}

	if !nodes.Empty() {
		builder.Nodes(nodes)
	}
	ccs := cmv1.NewCCS()
	if common.HasValue(state.CCSEnabled) {
		ccs.Enabled(state.CCSEnabled.ValueBool())
	}
	if !ccs.Empty() {
		builder.CCS(ccs)
	}
	aws := cmv1.NewAWS()
	if common.HasValue(state.AWSAccountID) {
		aws.AccountID(state.AWSAccountID.ValueString())
	}
	if common.HasValue(state.AWSAccessKeyID) {
		aws.AccessKeyID(state.AWSAccessKeyID.ValueString())
	}
	if common.HasValue(state.AWSSecretAccessKey) {
		aws.SecretAccessKey(state.AWSSecretAccessKey.ValueString())
	}
	if common.HasValue(state.AWSPrivateLink) {
		aws.PrivateLink(state.AWSPrivateLink.ValueBool())
		api := cmv1.NewClusterAPI()
		if state.AWSPrivateLink.ValueBool() {
			api.Listening(cmv1.ListeningMethodInternal)
		}
		builder.API(api)
	}

	if common.HasValue(state.AWSSubnetIDs) {
		subnetIds, err := common.StringListToArray(ctx, state.AWSSubnetIDs)
		if err != nil {
			return nil, err
		}
		aws.SubnetIDs(subnetIds...)
	}

	if !aws.Empty() {
		builder.AWS(aws)
	}
	network := cmv1.NewNetwork()
	if common.HasValue(state.MachineCIDR) {
		network.MachineCIDR(state.MachineCIDR.ValueString())
	}
	if common.HasValue(state.ServiceCIDR) {
		network.ServiceCIDR(state.ServiceCIDR.ValueString())
	}
	if common.HasValue(state.PodCIDR) {
		network.PodCIDR(state.PodCIDR.ValueString())
	}
	if common.HasValue(state.HostPrefix) {
		network.HostPrefix(int(state.HostPrefix.ValueInt64()))
	}
	if !network.Empty() {
		builder.Network(network)
	}
	if common.HasValue(state.Version) {
		builder.Version(cmv1.NewVersion().ID(state.Version.ValueString()))
	}

	proxyObj := cmv1.NewProxy()
	if state.Proxy != nil {
		proxyObj.HTTPProxy(state.Proxy.HttpProxy.ValueString())
		proxyObj.HTTPSProxy(state.Proxy.HttpsProxy.ValueString())
		builder.Proxy(proxyObj)
	}

	object, err := builder.Build()

	return object, err
}

func (r *ClusterResource) Create(ctx context.Context, request resource.CreateRequest,
	response *resource.CreateResponse) {
	// Get the plan:
	state := &ClusterState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	object, err := createClusterObject(ctx, state, diags)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster",
			fmt.Sprintf(
				"Can't build cluster with name '%s': %v",
				state.Name.ValueString(), err,
			),
		)
		return
	}

	add, err := r.collection.Add().Body(object).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't create cluster",
			fmt.Sprintf(
				"Can't create cluster with name '%s': %v",
				state.Name.ValueString(), err,
			),
		)
		return
	}
	object = add.Body()

	// Wait till the cluster is ready unless explicitly disabled:
	wait := state.Wait.IsUnknown() || state.Wait.IsNull() || state.Wait.ValueBool()
	ready := object.State() == cmv1.ClusterStateReady
	if wait && !ready {
		pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour)
		defer cancel()
		_, err := r.collection.Cluster(object.ID()).Poll().
			Interval(30 * time.Second).
			Predicate(func(get *cmv1.ClusterGetResponse) bool {
				object = get.Body()
				return object.State() == cmv1.ClusterStateReady
			}).
			StartContext(pollCtx)
		if err != nil {
			response.Diagnostics.AddError(
				"Can't poll cluster state",
				fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					object.ID(), err,
				),
			)
			return
		}
	}

	// Save the state:
	err = populateClusterState(object, state)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterResource) Read(ctx context.Context, request resource.ReadRequest,
	response *resource.ReadResponse) {
	// Get the current state:
	state := &ClusterState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the cluster:
	get, err := r.collection.Cluster(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}
	object := get.Body()

	// Save the state:
	err = populateClusterState(object, state)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterResource) Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse) {
	var diags diag.Diagnostics

	// Get the state:
	state := &ClusterState{}
	diags = request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the plan:
	plan := &ClusterState{}
	diags = request.Plan.Get(ctx, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send request to update the cluster:
	builder := cmv1.NewCluster()
	var nodes *cmv1.ClusterNodesBuilder
	compute, ok := common.ShouldPatchInt(state.ComputeNodes, plan.ComputeNodes)
	if ok {
		nodes.Compute(int(compute))
	}
	if !nodes.Empty() {
		builder.Nodes(nodes)
	}
	patch, err := builder.Build()
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build cluster patch",
			fmt.Sprintf(
				"Can't build patch for cluster with identifier '%s': %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}
	update, err := r.collection.Cluster(state.ID.ValueString()).Update().
		Body(patch).
		SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't update cluster",
			fmt.Sprintf(
				"Can't update cluster with identifier '%s': %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}
	object := update.Body()

	// Update the state:
	err = populateClusterState(object, state)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't populate cluster state",
			fmt.Sprintf(
				"Received error %v", err,
			),
		)
		return
	}
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *ClusterResource) Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse) {
	// Get the state:
	state := &ClusterState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send the request to delete the cluster:
	resource := r.collection.Cluster(state.ID.ValueString())
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete cluster",
			fmt.Sprintf(
				"Can't delete cluster with identifier '%s': %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}

	// Wait till the cluster has been effectively deleted:
	if state.Wait.IsUnknown() || state.Wait.IsNull() || state.Wait.ValueBool() {
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
			response.Diagnostics.AddError(
				"Can't poll cluster deletion",
				fmt.Sprintf(
					"Can't poll deletion of cluster with identifier '%s': %v",
					state.ID.ValueString(), err,
				),
			)
			return
		}
	}

	// Remove the state:
	response.State.RemoveResource(ctx)
}

func (r *ClusterResource) ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// populateClusterState copies the data from the API object to the Terraform state.
func populateClusterState(object *cmv1.Cluster, state *ClusterState) error {
	state.ID = types.StringValue(object.ID())

	object.API()
	state.Product = types.StringValue(object.Product().ID())
	state.Name = types.StringValue(object.Name())
	state.CloudProvider = types.StringValue(object.CloudProvider().ID())
	state.CloudRegion = types.StringValue(object.Region().ID())
	state.MultiAZ = types.BoolValue(object.MultiAZ())

	mapValue, err := common.ConvertStringMapToMapType(object.Properties())
	if err != nil {
		return err
	} else {
		state.Properties = mapValue
	}

	state.APIURL = types.StringValue(object.API().URL())
	state.ConsoleURL = types.StringValue(object.Console().URL())
	state.ComputeNodes = types.Int64Value(int64(object.Nodes().Compute()))
	state.ComputeMachineType = types.StringValue(object.Nodes().ComputeMachineType().ID())

	azs, ok := object.Nodes().GetAvailabilityZones()
	if ok {
		listValue, err := common.StringArrayToList(azs)
		if err != nil {
			return err
		} else {
			state.AvailabilityZones = listValue
		}
	}

	state.CCSEnabled = types.BoolValue(object.CCS().Enabled())
	//The API does not return account id
	awsAccountID, ok := object.AWS().GetAccountID()
	if ok {
		state.AWSAccountID = types.StringValue(awsAccountID)
	}
	awsAccessKeyID, ok := object.AWS().GetAccessKeyID()
	if ok {
		state.AWSAccessKeyID = types.StringValue(awsAccessKeyID)
	}

	awsSecretAccessKey, ok := object.AWS().GetSecretAccessKey()
	if ok {
		state.AWSSecretAccessKey = types.StringValue(awsSecretAccessKey)
	}
	awsPrivateLink, ok := object.AWS().GetPrivateLink()
	if ok {
		state.AWSPrivateLink = types.BoolValue(awsPrivateLink)
	} else {
		state.AWSPrivateLink = types.BoolValue(true)
	}

	subnetIds, ok := object.AWS().GetSubnetIDs()
	if ok {
		awsSubnetIds, err := common.StringArrayToList(subnetIds)
		if err != nil {
			return err
		}
		state.AWSSubnetIDs = awsSubnetIds
	}

	proxyObj, ok := object.GetProxy()
	if ok {
		state.Proxy.HttpProxy = types.StringValue(proxyObj.HTTPProxy())
		state.Proxy.HttpsProxy = types.StringValue(proxyObj.HTTPSProxy())
	}

	machineCIDR, ok := object.Network().GetMachineCIDR()
	if ok {
		state.MachineCIDR = types.StringValue(machineCIDR)
	} else {
		state.MachineCIDR = types.StringNull()
	}
	serviceCIDR, ok := object.Network().GetServiceCIDR()
	if ok {
		state.ServiceCIDR = types.StringValue(serviceCIDR)
	} else {
		state.ServiceCIDR = types.StringNull()
	}
	podCIDR, ok := object.Network().GetPodCIDR()
	if ok {
		state.PodCIDR = types.StringValue(podCIDR)
	} else {
		state.PodCIDR = types.StringNull()
	}
	hostPrefix, ok := object.Network().GetHostPrefix()
	if ok {
		state.HostPrefix = types.Int64Value(int64(hostPrefix))
	} else {
		state.HostPrefix = types.Int64Null()
	}
	version, ok := object.Version().GetID()
	if ok {
		state.Version = types.StringValue(version)
	} else {
		state.Version = types.StringNull()
	}
	state.State = types.StringValue(string(object.State()))

	return nil
}
