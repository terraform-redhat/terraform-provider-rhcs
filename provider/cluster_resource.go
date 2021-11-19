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
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
			"multi_az": {
				Description: "Indicates if the cluster should be deployed to " +
					"multiple availability zones. Default value is 'false'.",
				Type:     types.BoolType,
				Optional: true,
				Computed: true,
	***REMOVED***,
			"properties": {
				Description: "User defined properties.",
				Type: types.MapType{
					ElemType: types.StringType,
		***REMOVED***,
				Optional: true,
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
	r.logger.Info(ctx, "Get plan"***REMOVED***
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

	// Set the computed attributes:
	state.ID = types.String{
		Value: object.ID(***REMOVED***,
	}
	state.MultiAZ = types.Bool{
		Value: object.MultiAZ(***REMOVED***,
	}
	state.APIURL = types.String{
		Value: object.API(***REMOVED***.URL(***REMOVED***,
	}
	state.ConsoleURL = types.String{
		Value: object.Console(***REMOVED***.URL(***REMOVED***,
	}
	state.State = types.String{
		Value: string(object.State(***REMOVED******REMOVED***,
	}

	// Save the state:
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
			fmt.Sprintf(
				"can't find cluster with identifier '%s'",
				state.ID.Value,
			***REMOVED***,
			err.Error(***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Copy the cluster data into the state:
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
	state.State = types.String{
		Value: string(object.State(***REMOVED******REMOVED***,
	}
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *ClusterResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
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
	tfsdk.ResourceImportStatePassthroughID(
		ctx,
		tftypes.NewAttributePath(***REMOVED***.WithAttributeName("id"***REMOVED***,
		request,
		response,
	***REMOVED***
}
