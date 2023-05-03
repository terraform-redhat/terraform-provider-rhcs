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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/openshift-online/ocm-sdk-go/logging"
	rosaoidcconfig "github.com/openshift/rosa/pkg/helper/oidc_config"
***REMOVED***

type RosaOidcConfigInputResourceType struct {
	logger logging.Logger
}

type RosaOidcConfigInputResource struct {
	logger logging.Logger
}

func (t *RosaOidcConfigInputResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "OIDC config input resources' names",
		Attributes: map[string]tfsdk.Attribute{
			"region": {
				Description: "Unique identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"bucket_name": {
				Description: "The S3 bucket name",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"discovery_doc": {
				Description: "The discovery document string file",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"jwks": {
				Description: "Json web key set string file",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"private_key": {
				Description: "RSA private key",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"private_key_file_name": {
				Description: "The private key file name",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"private_key_secret_name": {
				Description: "The secret name that store the private key",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"issuer_url": {
				Description: "The issuer URL",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *RosaOidcConfigInputResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Create the resource:
	result = &RosaOidcConfigInputResource{
		logger: parent.logger,
	}

	return
}

func (r *RosaOidcConfigInputResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	// Get the plan:
	var state RosaOidcConfigInputState
	diags := request.Plan.Get(ctx, &state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	region := state.Region.Value
	oidcConfigInput, err := rosaoidcconfig.BuildOidcConfigInput("", region***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't generate oidc config input object",
			fmt.Sprintf(
				"Can't generate oidc config input object: %v",
				err,
			***REMOVED***,
		***REMOVED***
		return
	}
	state.BucketName = types.String{
		Value: oidcConfigInput.BucketName,
	}
	state.IssuerUrl = types.String{
		Value: oidcConfigInput.IssuerUrl,
	}
	state.PrivateKey = types.String{
		Value: string(oidcConfigInput.PrivateKey[:]***REMOVED***,
	}
	state.PrivateKeyFileName = types.String{
		Value: oidcConfigInput.PrivateKeyFilename,
	}
	state.DiscoveryDoc = types.String{
		Value: oidcConfigInput.DiscoveryDocument,
	}
	state.Jwks = types.String{
		Value: string(oidcConfigInput.Jwks[:]***REMOVED***,
	}
	state.PrivateKeySecretName = types.String{
		Value: oidcConfigInput.PrivateKeySecretName,
	}

	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *RosaOidcConfigInputResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	// Do Nothing
}

func (r *RosaOidcConfigInputResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
	response.Diagnostics.AddError(
		"Update methode is not supported for that resource",
		fmt.Sprintf(
			"Update methode is not supported for that resource",
		***REMOVED***,
	***REMOVED***
	return
}

func (r *RosaOidcConfigInputResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *RosaOidcConfigInputResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	// Do Nothing
}
