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

package oidcconfig

***REMOVED***
	"context"
***REMOVED***
***REMOVED***
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

type RosaOidcConfigResource struct {
	oidcConfigClient *cmv1.OidcConfigsClient
	clustersClient   *cmv1.ClustersClient
}

var _ resource.ResourceWithConfigure = &RosaOidcConfigResource{}
var _ resource.ResourceWithImportState = &RosaOidcConfigResource{}

func New(***REMOVED*** resource.Resource {
	return &RosaOidcConfigResource{}
}

func (o *RosaOidcConfigResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_rosa_oidc_config"
}

func (o *RosaOidcConfigResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "Manages OIDC config",
		Attributes: map[string]schema.Attribute{
			"managed": schema.BoolAttribute{
				Description: "Indicates whether it is a Red Hat managed or unmanaged (Customer hosted***REMOVED*** OIDC configuration, for the cluster's OIDC provider.",
				Required:    true,
	***REMOVED***,
			"secret_arn": schema.StringAttribute{
				Description: "Indicates for unmanaged OIDC config, the secret ARN",
				Optional:    true,
	***REMOVED***,
			"issuer_url": schema.StringAttribute{
				Description: "The bucket/issuer URL",
				Optional:    true,
				Computed:    true,
	***REMOVED***,
			"installer_role_arn": schema.StringAttribute{
				Description: "AWS STS Role ARN for cluster install (with get-secrets permission in the attached policy***REMOVED***",
				Optional:    true,
	***REMOVED***,
			"id": schema.StringAttribute{
				Description: "The OIDC config ID",
				Computed:    true,
	***REMOVED***,
			"thumbprint": schema.StringAttribute{
				Description: "SHA1-hash value of the root CA of the issuer URL",
				Computed:    true,
	***REMOVED***,
			"oidc_endpoint_url": schema.StringAttribute{
				Description: "OIDC Endpoint URL",
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (o *RosaOidcConfigResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
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

	o.oidcConfigClient = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.OidcConfigs(***REMOVED***
	o.clustersClient = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
}

func (o *RosaOidcConfigResource***REMOVED*** Create(ctx context.Context, request resource.CreateRequest,
	response *resource.CreateResponse***REMOVED*** {
	// Get the plan:
	state := &RosaOidcConfigState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	managed := state.Managed.ValueBool(***REMOVED***
	var oidcConfig *cmv1.OidcConfig
	var err error
	if managed {
		if (!state.SecretARN.IsUnknown(***REMOVED*** && !state.SecretARN.IsNull(***REMOVED******REMOVED*** ||
			(!state.IssuerUrl.IsUnknown(***REMOVED*** && !state.IssuerUrl.IsNull(***REMOVED******REMOVED*** ||
			(!state.InstallerRoleARN.IsUnknown(***REMOVED*** && !state.InstallerRoleARN.IsNull(***REMOVED******REMOVED*** {
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
		if state.SecretARN.IsUnknown(***REMOVED*** || state.SecretARN.IsNull(***REMOVED*** ||
			state.IssuerUrl.IsUnknown(***REMOVED*** || state.IssuerUrl.IsNull(***REMOVED*** ||
			state.InstallerRoleARN.IsUnknown(***REMOVED*** || state.InstallerRoleARN.IsNull(***REMOVED*** {
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
			SecretArn(state.SecretARN.ValueString(***REMOVED******REMOVED***.
			IssuerUrl(state.IssuerUrl.ValueString(***REMOVED******REMOVED***.
			InstallerRoleArn(state.InstallerRoleARN.ValueString(***REMOVED******REMOVED***.
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

	object, err := o.oidcConfigClient.Add(***REMOVED***.Body(oidcConfig***REMOVED***.SendContext(ctx***REMOVED***
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
	o.populateState(ctx, oidcConfig, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (o *RosaOidcConfigResource***REMOVED*** Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse***REMOVED*** {
	// Get the current state:
	state := &RosaOidcConfigState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the oidc config:
	get, err := o.oidcConfigClient.OidcConfig(state.ID.ValueString(***REMOVED******REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil && get.Status(***REMOVED*** == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("oidc config (%s***REMOVED*** not found, removing from state",
			state.ID.ValueString(***REMOVED***,
		***REMOVED******REMOVED***
		response.State.RemoveResource(ctx***REMOVED***
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Cannot find OIDC config",
			fmt.Sprintf(
				"Cannot find OIDC config with ID %s, %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	object := get.Body(***REMOVED***

	// Save the state:
	o.populateState(ctx, object, state***REMOVED***
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (o *RosaOidcConfigResource***REMOVED*** Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse***REMOVED*** {
	response.Diagnostics.AddError(
		"Update methode is not supported for that resource",
		fmt.Sprintf(
			"Update methode is not supported for that resource",
		***REMOVED***,
	***REMOVED***
	return
}

func (o *RosaOidcConfigResource***REMOVED*** Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse***REMOVED*** {
	// Get the state:
	state := &RosaOidcConfigState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the oidc config:
	get, err := o.oidcConfigClient.OidcConfig(state.ID.ValueString(***REMOVED******REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Cannot find OIDC config",
			fmt.Sprintf(
				"Cannot find OIDC config with ID %s, %v",
				state.ID.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	oidcConfig := get.Body(***REMOVED***

	// check if there is a cluster using the oidc endpoint:
	hasClusterUsingOidcConfig, err := o.hasAClusterUsingOidcEndpointUrl(ctx, oidcConfig.IssuerUrl(***REMOVED******REMOVED***
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

	err = o.deleteOidcConfig(ctx, state.ID.ValueString(***REMOVED******REMOVED***
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

func (o *RosaOidcConfigResource***REMOVED*** hasAClusterUsingOidcEndpointUrl(ctx context.Context, issuerUrl string***REMOVED*** (bool, error***REMOVED*** {
	query := fmt.Sprintf(
		"aws.sts.oidc_endpoint_url = '%s'", issuerUrl,
	***REMOVED***
	request := o.clustersClient.List(***REMOVED***.Search(query***REMOVED***
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

func (o *RosaOidcConfigResource***REMOVED*** deleteOidcConfig(ctx context.Context, id string***REMOVED*** error {
	_, err := o.oidcConfigClient.
		OidcConfig(id***REMOVED***.
		Delete(***REMOVED***.
		SendContext(ctx***REMOVED***
	return err
}

func (o *RosaOidcConfigResource***REMOVED*** ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse***REMOVED*** {
	resource.ImportStatePassthroughID(
		ctx,
		path.Root("id"***REMOVED***,
		request,
		response,
	***REMOVED***
}

// populateState copies the data from the API object to the Terraform state.
func (o *RosaOidcConfigResource***REMOVED*** populateState(ctx context.Context, object *cmv1.OidcConfig, state *RosaOidcConfigState***REMOVED*** {
	if id, ok := object.GetID(***REMOVED***; ok {
		state.ID = types.StringValue(id***REMOVED***
	}

	if managed, ok := object.GetManaged(***REMOVED***; ok {
		state.Managed = types.BoolValue(managed***REMOVED***
	}

	issuerUrl, ok := object.GetIssuerUrl(***REMOVED***
	if ok && issuerUrl != "" {
		state.IssuerUrl = types.StringValue(issuerUrl***REMOVED***
	}

	installerRoleArn, ok := object.GetInstallerRoleArn(***REMOVED***
	if ok && installerRoleArn != "" {
		state.InstallerRoleARN = types.StringValue(installerRoleArn***REMOVED***
	}

	secretArn, ok := object.GetSecretArn(***REMOVED***
	if ok && secretArn != "" {
		state.SecretARN = types.StringValue(secretArn***REMOVED***
	}

	oidcEndpointURL := issuerUrl
	if strings.HasPrefix(oidcEndpointURL, "https://"***REMOVED*** {
		oidcEndpointURL = strings.TrimPrefix(oidcEndpointURL, "https://"***REMOVED***
	}
	state.OIDCEndpointURL = types.StringValue(oidcEndpointURL***REMOVED***

	thumbprint, err := common.GetThumbprint(issuerUrl, common.DefaultHttpClient{}***REMOVED***
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("cannot get thumbprint, with error: %v", err***REMOVED******REMOVED***
		state.Thumbprint = types.StringValue(""***REMOVED***
	} else {
		state.Thumbprint = types.StringValue(thumbprint***REMOVED***
	}

}
