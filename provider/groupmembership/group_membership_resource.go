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

package groupmembership

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

type GroupMembershipResource struct {
	collection *cmv1.ClustersClient
}

var _ resource.ResourceWithConfigure = &GroupMembershipResource{}
var _ resource.ResourceWithImportState = &GroupMembershipResource{}

func New(***REMOVED*** resource.Resource {
	return &GroupMembershipResource{}
}

func (g *GroupMembershipResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_group_membership"
}

func (g *GroupMembershipResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "Manages user group membership.",
		DeprecationMessage: fmt.Sprintf(
			"using group membership as a resource is deprecated"***REMOVED***,
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
	***REMOVED***,
			"group": schema.StringAttribute{
				Description: "Identifier of the group.",
				Required:    true,
	***REMOVED***,
			"id": schema.StringAttribute{
				Description: "Identifier of the membership.",
				Computed:    true,
	***REMOVED***,
			"user": schema.StringAttribute{
				Description: "user name.",
				Required:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (g *GroupMembershipResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
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

	g.collection = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
}

func (g *GroupMembershipResource***REMOVED*** Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse***REMOVED*** {
	// Get the plan:
	state := &GroupMembershipState{}
	diags := req.Plan.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Wait till the cluster is ready:
	err := common.WaitTillClusterReady(ctx, g.collection, state.Cluster.ValueString(***REMOVED******REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't poll cluster state",
			fmt.Sprintf(
				"Can't poll state of cluster with identifier '%s': %v",
				state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Create the membership:
	builder := cmv1.NewUser(***REMOVED***
	builder.ID(state.User.ValueString(***REMOVED******REMOVED***
	object, err := builder.Build(***REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't build group membership",
			fmt.Sprintf(
				"Can't build group membership for cluster '%s' and group '%s': %v",
				state.Cluster.ValueString(***REMOVED***, state.Group.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	collection := g.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.Groups(***REMOVED***.Group(state.Group.ValueString(***REMOVED******REMOVED***.Users(***REMOVED***
	add, err := collection.Add(***REMOVED***.Body(object***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't create group membership",
			fmt.Sprintf(
				"Can't create group membership for cluster '%s' and group '%s': %v",
				state.Cluster.ValueString(***REMOVED***, state.Group.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	// Save the state:
	g.populateState(object, state***REMOVED***
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (g *GroupMembershipResource***REMOVED*** Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse***REMOVED*** {
	// Get the current state:
	state := &GroupMembershipState{}
	diags := req.State.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the group membership:
	obj := g.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.Groups(***REMOVED***.Group(state.Group.ValueString(***REMOVED******REMOVED***.
		Users(***REMOVED***.
		User(state.ID.ValueString(***REMOVED******REMOVED***
	get, err := obj.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't find group membership",
			fmt.Sprintf(
				"Can't find user group membership identifier '%s' for "+
					"cluster '%s' and group '%s': %v",
				state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, state.Group.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Save the state:
	g.populateState(object, state***REMOVED***

	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (g *GroupMembershipResource***REMOVED*** Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse***REMOVED*** {
	// Until we support. return an informative error
	resp.Diagnostics.AddError("Can't update group membership", "Update is currently not supported."***REMOVED***
}

func (g *GroupMembershipResource***REMOVED*** Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse***REMOVED*** {
	// Get the state:
	state := &GroupMembershipState{}
	diags := req.State.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete group membership:
	obj := g.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.Groups(***REMOVED***.Group(state.Group.ValueString(***REMOVED******REMOVED***.
		Users(***REMOVED***.
		User(state.ID.ValueString(***REMOVED******REMOVED***
	_, err := obj.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		resp.Diagnostics.AddError(
			"Can't delete group membership",
			fmt.Sprintf(
				"Can't delete group membership with identifier '%s' for "+
					"cluster '%s' and group '%s': %v",
				state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, state.Group.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Remove the state:
	resp.State.RemoveResource(ctx***REMOVED***
}

func (g *GroupMembershipResource***REMOVED*** ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse***REMOVED*** {
	resource.ImportStatePassthroughID(ctx, path.Root("id"***REMOVED***, req, resp***REMOVED***
}

// populateState copies the data from the API object to the Terraform state.
func (g *GroupMembershipResource***REMOVED*** populateState(object *cmv1.User, state *GroupMembershipState***REMOVED*** {
	if id, ok := object.GetID(***REMOVED***; ok {
		state.ID = types.StringValue(id***REMOVED***
		state.User = types.StringValue(id***REMOVED***
	}
}
