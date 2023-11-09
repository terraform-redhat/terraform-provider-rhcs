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
package oidcconfiginput

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	rosaoidcconfig "github.com/openshift/rosa/pkg/helper/oidc_config"
)

type RosaOidcConfigInputResource struct {
}

var _ resource.ResourceWithConfigure = &RosaOidcConfigInputResource{}
var _ resource.ResourceWithImportState = &RosaOidcConfigInputResource{}

func New() resource.Resource {
	return &RosaOidcConfigInputResource{}
}

func (o *RosaOidcConfigInputResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rosa_oidc_config_input"
}

func (o *RosaOidcConfigInputResource) Schema(ctx context.Context, request resource.SchemaRequest,
	response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "OIDC config input resources' names",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "Unique identifier of the cluster.",
				Required:    true,
			},
			"bucket_name": schema.StringAttribute{
				Description: "The S3 bucket name",
				Computed:    true,
			},
			"discovery_doc": schema.StringAttribute{
				Description: "The discovery document string file",
				Computed:    true,
			},
			"jwks": schema.StringAttribute{
				Description: "JSON web key set string file",
				Computed:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "RSA private key",
				Computed:    true,
			},
			"private_key_file_name": schema.StringAttribute{
				Description: "The private key file name",
				Computed:    true,
			},
			"private_key_secret_name": schema.StringAttribute{
				Description: "The secret name that stores the private key",
				Computed:    true,
			},
			"issuer_url": schema.StringAttribute{
				Description: "The issuer URL",
				Computed:    true,
			},
		},
	}
	return
}

func (o *RosaOidcConfigInputResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Do nothing
}

func (o *RosaOidcConfigInputResource) Create(ctx context.Context, request resource.CreateRequest,
	response *resource.CreateResponse) {
	// Get the plan:
	var state RosaOidcConfigInputState
	diags := request.Plan.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	region := state.Region.ValueString()
	oidcConfigInput, err := rosaoidcconfig.BuildOidcConfigInput("", region)
	if err != nil {
		response.Diagnostics.AddError(
			"Cannot generate oidc config input object",
			fmt.Sprintf(
				"Cannot generate oidc config input object: %v",
				err,
			),
		)
		return
	}
	state.BucketName = types.StringValue(oidcConfigInput.BucketName)
	state.IssuerUrl = types.StringValue(oidcConfigInput.IssuerUrl)
	state.PrivateKey = types.StringValue(string(oidcConfigInput.PrivateKey[:]))
	state.PrivateKeyFileName = types.StringValue(oidcConfigInput.PrivateKeyFilename)
	state.DiscoveryDoc = types.StringValue(oidcConfigInput.DiscoveryDocument)

	state.Jwks = types.StringValue(string(oidcConfigInput.Jwks[:]))

	state.PrivateKeySecretName = types.StringValue(oidcConfigInput.PrivateKeySecretName)

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (o *RosaOidcConfigInputResource) Read(ctx context.Context, request resource.ReadRequest,
	response *resource.ReadResponse) {
	// Do Nothing
}

func (o *RosaOidcConfigInputResource) Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse) {
	response.Diagnostics.AddError(
		"Update methode is not supported for that resource",
		fmt.Sprintf(
			"Update methode is not supported for that resource",
		),
	)
	return
}

func (o *RosaOidcConfigInputResource) Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse) {
	response.State.RemoveResource(ctx)
}

func (o *RosaOidcConfigInputResource) ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse) {
	// Do Nothing

}
