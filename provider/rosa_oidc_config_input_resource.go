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
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	rosaoidcconfig "github.com/openshift/rosa/pkg/helper/oidc_config"
)

type RosaOidcConfigInputResourceType struct {
}

type RosaOidcConfigInputResource struct {
}

func (t *RosaOidcConfigInputResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "OIDC config input resources' names",
		Attributes: map[string]tfsdk.Attribute{
			"region": {
				Description: "Unique identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
			},
			"bucket_name": {
				Description: "The S3 bucket name",
				Type:        types.StringType,
				Computed:    true,
			},
			"discovery_doc": {
				Description: "The discovery document string file",
				Type:        types.StringType,
				Computed:    true,
			},
			"jwks": {
				Description: "Json web key set string file",
				Type:        types.StringType,
				Computed:    true,
			},
			"private_key": {
				Description: "RSA private key",
				Type:        types.StringType,
				Computed:    true,
			},
			"private_key_file_name": {
				Description: "The private key file name",
				Type:        types.StringType,
				Computed:    true,
			},
			"private_key_secret_name": {
				Description: "The secret name that store the private key",
				Type:        types.StringType,
				Computed:    true,
			},
			"issuer_url": {
				Description: "The issuer URL",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}
	return
}

func (t *RosaOidcConfigInputResourceType) NewResource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.Resource, diags diag.Diagnostics) {
	// Create the resource:
	result = &RosaOidcConfigInputResource{}

	return
}

func (r *RosaOidcConfigInputResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	// Get the plan:
	var state RosaOidcConfigInputState
	diags := request.Plan.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	region := state.Region.Value
	oidcConfigInput, err := rosaoidcconfig.BuildOidcConfigInput("", region)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't generate oidc config input object",
			fmt.Sprintf(
				"Can't generate oidc config input object: %v",
				err,
			),
		)
		return
	}
	state.BucketName = types.String{
		Value: oidcConfigInput.BucketName,
	}
	state.IssuerUrl = types.String{
		Value: oidcConfigInput.IssuerUrl,
	}
	state.PrivateKey = types.String{
		Value: string(oidcConfigInput.PrivateKey[:]),
	}
	state.PrivateKeyFileName = types.String{
		Value: oidcConfigInput.PrivateKeyFilename,
	}
	state.DiscoveryDoc = types.String{
		Value: oidcConfigInput.DiscoveryDocument,
	}
	state.Jwks = types.String{
		Value: string(oidcConfigInput.Jwks[:]),
	}
	state.PrivateKeySecretName = types.String{
		Value: oidcConfigInput.PrivateKeySecretName,
	}

	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *RosaOidcConfigInputResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse) {
	// Do Nothing
}

func (r *RosaOidcConfigInputResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse) {
	response.Diagnostics.AddError(
		"Update methode is not supported for that resource",
		fmt.Sprintf(
			"Update methode is not supported for that resource",
		),
	)
	return
}

func (r *RosaOidcConfigInputResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse) {
	response.State.RemoveResource(ctx)
}

func (r *RosaOidcConfigInputResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	// Do Nothing
}
