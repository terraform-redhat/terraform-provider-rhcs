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
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type RosaOidcConfigResourceType struct {
}

type RosaOidcConfigResource struct {
	oidcConfigClient *cmv1.OidcConfigsClient
	clustersClient   *cmv1.ClustersClient
}

func (t *RosaOidcConfigResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "OIDC config",
		Attributes: map[string]tfsdk.Attribute{
			"managed": {
				Description: "Indicates whether it is a Red Hat managed or unmanaged (Customer hosted***REMOVED*** OIDC Configuration",
				Type:        types.BoolType,
				Required:    true,
	***REMOVED***,
			"secret_arn": {
				Description: "Indicates for unmanaged OIDC config, the secret ARN",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"issuer_url": {
				Description: "The bucket URL",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"installer_role_arn": {
				Description: "STS Role ARN with get secrets permission",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"id": {
				Description: "The OIDC config ID",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"thumbprint": {
				Description: "SHA1-hash value of the root CA of the issuer URL",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"oidc_endpoint_url": {
				Description: "OIDC Endpoint URL",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *RosaOidcConfigResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation: use it directly when needed.
	parent := p.(*Provider***REMOVED***

	// Get the oidcConfigClient:
	oidcConfigClient := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.OidcConfigs(***REMOVED***
	// Get the clustersClient:
	clustersClient := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Create the resource:
	result = &RosaOidcConfigResource{
		oidcConfigClient: oidcConfigClient,
		clustersClient:   clustersClient,
	}

	return
}

func (r *RosaOidcConfigResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	// Get the plan:
	state := &RosaOidcConfigState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	managed := state.Managed.Value
	var oidcConfig *cmv1.OidcConfig
	var err error
	if managed {
		if (!state.SecretARN.Unknown && !state.SecretARN.Null***REMOVED*** ||
			(!state.IssuerUrl.Unknown && !state.IssuerUrl.Null***REMOVED*** ||
			(!state.InstallerRoleARN.Unknown && !state.InstallerRoleARN.Null***REMOVED*** {
			response.Diagnostics.AddError(
				"Attribute's values are not supported for managed OIDC Configuration",
				fmt.Sprintf(
					"In order to create managed OIDC Configuration, "+
						"the attributes' values of `secret_arn`, `issuer_url` and `installer_role_arn` should be empty",
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		oidcConfig, err = cmv1.NewOidcConfig(***REMOVED***.Managed(true***REMOVED***.Build(***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(
				"There was a problem building the managed OIDC Configuration",
				fmt.Sprintf(
					"There was a problem building the managed OIDC Configuration: %v", err,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
	} else {
		if state.SecretARN.Unknown || state.SecretARN.Null ||
			state.IssuerUrl.Unknown || state.IssuerUrl.Null ||
			state.InstallerRoleARN.Unknown || state.InstallerRoleARN.Null {
			response.Diagnostics.AddError(
				"There is a missing parameter for unmanaged OIDC Configuration",
				fmt.Sprintf(
					"There is a missing parameter for unmanaged OIDC Configuration. "+
						"Please provide values for all those attributes `secret_arn`, `issuer_url` and `installer_role_arn`",
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
		oidcConfig, err = cmv1.NewOidcConfig(***REMOVED***.
			Managed(false***REMOVED***.
			SecretArn(state.SecretARN.Value***REMOVED***.
			IssuerUrl(state.IssuerUrl.Value***REMOVED***.
			InstallerRoleArn(state.InstallerRoleARN.Value***REMOVED***.
			Build(***REMOVED***

		if err != nil {
			response.Diagnostics.AddError(
				"There was a problem building the unmanaged OIDC Configuration",
				fmt.Sprintf(
					"There was a problem building the unmanaged OIDC Configuration: %v", err,
				***REMOVED***,
			***REMOVED***
			return
***REMOVED***
	}

	object, err := r.oidcConfigClient.Add(***REMOVED***.Body(oidcConfig***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"There was a problem registering the OIDC Configuration",
			fmt.Sprintf(
				"There was a problem registering the OIDC Configuration: %v", err,
			***REMOVED***,
		***REMOVED***
		return
	}

	oidcConfig = object.Body(***REMOVED***

	// Save the state:
	r.populateState(ctx, oidcConfig, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *RosaOidcConfigResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	// Get the current state:
	state := &RosaOidcConfigState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the oidc config:
	get, err := r.oidcConfigClient.OidcConfig(state.ID.Value***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find OIDC config",
			fmt.Sprintf(
				"Can't find OIDC config with ID %s, %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object := get.Body(***REMOVED***

	// Save the state:
	r.populateState(ctx, object, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *RosaOidcConfigResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
	response.Diagnostics.AddError(
		"Update methode is not supported for that resource",
		fmt.Sprintf(
			"Update methode is not supported for that resource",
		***REMOVED***,
	***REMOVED***
	return
}

func (r *RosaOidcConfigResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	// Get the state:
	state := &RosaOidcConfigState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the oidc config:
	get, err := r.oidcConfigClient.OidcConfig(state.ID.Value***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find OIDC config",
			fmt.Sprintf(
				"Can't find OIDC config with ID %s, %v",
				state.ID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	oidcConfig := get.Body(***REMOVED***

	// check if there is a cluster using the oidc endpoint:
	hasClusterUsingOidcConfig, err := r.hasAClusterUsingOidcEndpointUrl(ctx, oidcConfig.IssuerUrl(***REMOVED******REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"There was a problem checking if any clusters are using OIDC config",
			fmt.Sprintf(
				"There was a problem checking if any clusters are using OIDC config '%s' : %v",
				oidcConfig.IssuerUrl(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	if hasClusterUsingOidcConfig {
		response.Diagnostics.AddError(
			"here are clusters using OIDC config, can't delete the configuration",
			fmt.Sprintf(
				"here are clusters using OIDC config '%s', can't delete the configuration",
				oidcConfig.IssuerUrl(***REMOVED***,
			***REMOVED***,
		***REMOVED***
		return
	}

	err = r.deleteOidcConfig(ctx, state.ID.Value***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"There was a problem deleting the OIDC config",
			fmt.Sprintf(
				"There was a problem deleting the OIDC config '%s' : %v",
				oidcConfig.IssuerUrl(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *RosaOidcConfigResource***REMOVED*** hasAClusterUsingOidcEndpointUrl(ctx context.Context, issuerUrl string***REMOVED*** (bool, error***REMOVED*** {
	query := fmt.Sprintf(
		"aws.sts.oidc_endpoint_url = '%s'", issuerUrl,
	***REMOVED***
	request := r.clustersClient.List(***REMOVED***.Search(query***REMOVED***
	page := 1
	response, err := request.Page(page***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		return false, err
	}
	if response.Total(***REMOVED*** > 0 {
		return true, nil
	}
	return false, nil
}

func (r *RosaOidcConfigResource***REMOVED*** deleteOidcConfig(ctx context.Context, id string***REMOVED*** error {
	_, err := r.oidcConfigClient.
		OidcConfig(id***REMOVED***.
		Delete(***REMOVED***.
		SendContext(ctx***REMOVED***
	return err
}

func (r *RosaOidcConfigResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	tfsdk.ResourceImportStatePassthroughID(
		ctx,
		tftypes.NewAttributePath(***REMOVED***.WithAttributeName("id"***REMOVED***,
		request,
		response,
	***REMOVED***
}

// populateState copies the data from the API object to the Terraform state.
func (r *RosaOidcConfigResource***REMOVED*** populateState(ctx context.Context, object *cmv1.OidcConfig, state *RosaOidcConfigState***REMOVED*** {
	state.ID = types.String{
		Value: object.ID(***REMOVED***,
	}
	state.Managed = types.Bool{
		Value: object.Managed(***REMOVED***,
	}

	issuerUrl, ok := object.GetIssuerUrl(***REMOVED***
	if ok && issuerUrl != "" {
		state.IssuerUrl = types.String{
			Value: issuerUrl,
***REMOVED***
	}

	installerRoleArn, ok := object.GetInstallerRoleArn(***REMOVED***
	if ok && installerRoleArn != "" {
		state.InstallerRoleARN = types.String{
			Value: installerRoleArn,
***REMOVED***
	}

	secretArn, ok := object.GetSecretArn(***REMOVED***
	if ok && secretArn != "" {
		state.SecretARN = types.String{
			Value: secretArn,
***REMOVED***
	}
	oidcEndpointURL := issuerUrl
	if strings.HasPrefix(oidcEndpointURL, "https://"***REMOVED*** {
		oidcEndpointURL = strings.TrimPrefix(oidcEndpointURL, "https://"***REMOVED***
	}
	state.OIDCEndpointURL = types.String{
		Value: oidcEndpointURL,
	}

	thumbprint, err := getThumbprint(issuerUrl, DefaultHttpClient{}***REMOVED***
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("cannot get thumbprint, with error: %v", err***REMOVED******REMOVED***
		state.Thumbprint = types.String{
			Value: "",
***REMOVED***
	} else {
		state.Thumbprint = types.String{
			Value: thumbprint,
***REMOVED***
	}

}
