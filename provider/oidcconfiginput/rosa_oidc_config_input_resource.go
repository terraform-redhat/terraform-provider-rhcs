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
package oidcconfiginput

***REMOVED***
	"context"
***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	rosaoidcconfig "github.com/openshift/rosa/pkg/helper/oidc_config"
***REMOVED***

type RosaOidcConfigInputResource struct {
}

var _ resource.ResourceWithConfigure = &RosaOidcConfigInputResource{}
var _ resource.ResourceWithImportState = &RosaOidcConfigInputResource{}

func New(***REMOVED*** resource.Resource {
	return &RosaOidcConfigInputResource{}
}

func (o *RosaOidcConfigInputResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_rosa_oidc_config_input"
}

func (o *RosaOidcConfigInputResource***REMOVED*** Schema(ctx context.Context, request resource.SchemaRequest,
	response *resource.SchemaResponse***REMOVED*** {
	response.Schema = schema.Schema{
		Description: "OIDC config input resources' names",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Required:    true,
	***REMOVED***,
			"bucket_name": schema.StringAttribute{
				Description: "The S3 bucket name",
				Computed:    true,
	***REMOVED***,
			"discovery_doc": schema.StringAttribute{
				Description: "The discovery document string file",
				Computed:    true,
	***REMOVED***,
			"jwks": schema.StringAttribute{
				Description: "JSON web key set string file",
				Computed:    true,
	***REMOVED***,
			"private_key": schema.StringAttribute{
				Description: "RSA private key",
				Computed:    true,
	***REMOVED***,
			"private_key_file_name": schema.StringAttribute{
				Description: "The private key file name",
				Computed:    true,
	***REMOVED***,
			"private_key_secret_name": schema.StringAttribute{
				Description: "The secret name that stores the private key",
				Computed:    true,
	***REMOVED***,
			"issuer_url": schema.StringAttribute{
				Description: "The issuer URL",
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (o *RosaOidcConfigInputResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
	// Do nothing
}

func (o *RosaOidcConfigInputResource***REMOVED*** Create(ctx context.Context, request resource.CreateRequest,
	response *resource.CreateResponse***REMOVED*** {
	// Get the plan:
	var state RosaOidcConfigInputState
	diags := request.Plan.Get(ctx, &state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	region := state.Region.ValueString(***REMOVED***
	oidcConfigInput, err := rosaoidcconfig.BuildOidcConfigInput("", region***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Cannot generate oidc config input object",
			fmt.Sprintf(
				"Cannot generate oidc config input object: %v",
				err,
			***REMOVED***,
		***REMOVED***
		return
	}
	state.BucketName = types.StringValue(oidcConfigInput.BucketName***REMOVED***
	state.IssuerUrl = types.StringValue(oidcConfigInput.IssuerUrl***REMOVED***
	state.PrivateKey = types.StringValue(string(oidcConfigInput.PrivateKey[:]***REMOVED******REMOVED***
	state.PrivateKeyFileName = types.StringValue(oidcConfigInput.PrivateKeyFilename***REMOVED***
	state.DiscoveryDoc = types.StringValue(oidcConfigInput.DiscoveryDocument***REMOVED***

	state.Jwks = types.StringValue(string(oidcConfigInput.Jwks[:]***REMOVED******REMOVED***

	state.PrivateKeySecretName = types.StringValue(oidcConfigInput.PrivateKeySecretName***REMOVED***

	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (o *RosaOidcConfigInputResource***REMOVED*** Read(ctx context.Context, request resource.ReadRequest,
	response *resource.ReadResponse***REMOVED*** {
	// Do Nothing
}

func (o *RosaOidcConfigInputResource***REMOVED*** Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse***REMOVED*** {
	response.Diagnostics.AddError(
		"Update methode is not supported for that resource",
		fmt.Sprintf(
			"Update methode is not supported for that resource",
		***REMOVED***,
	***REMOVED***
	return
}

func (o *RosaOidcConfigInputResource***REMOVED*** Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse***REMOVED*** {
	response.State.RemoveResource(ctx***REMOVED***
}

func (o *RosaOidcConfigInputResource***REMOVED*** ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse***REMOVED*** {
	// Do Nothing

}
