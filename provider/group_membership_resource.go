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
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type GroupMembershipResourceType struct {
}

type GroupMembershipResource struct {
	collection *cmv1.ClustersClient
}

func (t *GroupMembershipResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "Manages user group membership.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"group": {
				Description: "Identifier of the group.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"id": {
				Description: "Identifier of the membership.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"user": {
				Description: "user name.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *GroupMembershipResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation: use it directly when needed.
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Create the resource:
	result = &GroupMembershipResource{
		collection: collection,
	}

	return
}

func (r *GroupMembershipResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	// Get the plan:
	state := &GroupMembershipState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
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

	// Create the membership:
	builder := cmv1.NewUser(***REMOVED***
	builder.ID(state.User.Value***REMOVED***
	object, err := builder.Build(***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build group membership",
			fmt.Sprintf(
				"Can't build group membership for cluster '%s' and group '%s': %v",
				state.Cluster.Value, state.Group.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	collection := resource.Groups(***REMOVED***.Group(state.Group.Value***REMOVED***.Users(***REMOVED***
	add, err := collection.Add(***REMOVED***.Body(object***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't create group membership",
			fmt.Sprintf(
				"Can't create group membership for cluster '%s' and group '%s': %v",
				state.Cluster.Value, state.Group.Value, err,
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

func (r *GroupMembershipResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	// Get the current state:
	state := &GroupMembershipState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the group membership:
	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***.Groups(***REMOVED***.Group(state.Group.Value***REMOVED***.
		Users(***REMOVED***.
		User(state.ID.Value***REMOVED***
	get, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find group membership",
			fmt.Sprintf(
				"Can't find user group membership identifier '%s' for "+
					"cluster '%s' and group '%s': %v",
				state.ID.Value, state.Cluster.Value, state.Group.Value, err,
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

func (r *GroupMembershipResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
}

func (r *GroupMembershipResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	// Get the state:
	state := &GroupMembershipState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete group membership:
	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***.Groups(***REMOVED***.Group(state.Group.Value***REMOVED***.
		Users(***REMOVED***.
		User(state.ID.Value***REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete group membership",
			fmt.Sprintf(
				"Can't delete group membership with identifier '%s' for "+
					"cluster '%s' and group '%s': %v",
				state.ID.Value, state.Cluster.Value, state.Group.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *GroupMembershipResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	tfsdk.ResourceImportStatePassthroughID(
		ctx,
		tftypes.NewAttributePath(***REMOVED***.WithAttributeName("id"***REMOVED***,
		request,
		response,
	***REMOVED***
}

// populateState copies the data from the API object to the Terraform state.
func (r *GroupMembershipResource***REMOVED*** populateState(object *cmv1.User, state *GroupMembershipState***REMOVED*** {
	state.ID = types.String{
		Value: object.ID(***REMOVED***,
	}
	state.User = types.String{
		Value: object.ID(***REMOVED***,
	}
}
