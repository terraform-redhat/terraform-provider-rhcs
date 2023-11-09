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

package oidcconfig

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type RosaOidcConfigResource struct {
	oidcConfigClient *cmv1.OidcConfigsClient
	clustersClient   *cmv1.ClustersClient
}

var _ resource.ResourceWithConfigure = &RosaOidcConfigResource{}
var _ resource.ResourceWithImportState = &RosaOidcConfigResource{}

func New() resource.Resource {
	return &RosaOidcConfigResource{}
}

func (o *RosaOidcConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rosa_oidc_config"
}

func (o *RosaOidcConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OIDC config",
		Attributes: map[string]schema.Attribute{
			"managed": schema.BoolAttribute{
				Description: "Indicates whether it is a Red Hat managed or unmanaged (Customer hosted) OIDC configuration, for the cluster's OIDC provider.",
				Required:    true,
			},
			"secret_arn": schema.StringAttribute{
				Description: "Indicates for unmanaged OIDC config, the secret ARN",
				Optional:    true,
			},
			"issuer_url": schema.StringAttribute{
				Description: "The bucket/issuer URL",
				Optional:    true,
				Computed:    true,
			},
			"installer_role_arn": schema.StringAttribute{
				Description: "AWS STS Role ARN for cluster install (with get-secrets permission in the attached policy)",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The OIDC config ID",
				Computed:    true,
			},
			"thumbprint": schema.StringAttribute{
				Description: "SHA1-hash value of the root CA of the issuer URL",
				Computed:    true,
			},
			"oidc_endpoint_url": schema.StringAttribute{
				Description: "OIDC Endpoint URL",
				Computed:    true,
			},
		},
	}
	return
}

func (o *RosaOidcConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	o.oidcConfigClient = connection.ClustersMgmt().V1().OidcConfigs()
	o.clustersClient = connection.ClustersMgmt().V1().Clusters()
}

func (o *RosaOidcConfigResource) Create(ctx context.Context, request resource.CreateRequest,
	response *resource.CreateResponse) {
	// Get the plan:
	state := &RosaOidcConfigState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	managed := state.Managed.ValueBool()
	var oidcConfig *cmv1.OidcConfig
	var err error
	if managed {
		if (!state.SecretARN.IsUnknown() && !state.SecretARN.IsNull()) ||
			(!state.IssuerUrl.IsUnknown() && !state.IssuerUrl.IsNull()) ||
			(!state.InstallerRoleARN.IsUnknown() && !state.InstallerRoleARN.IsNull()) {
			response.Diagnostics.AddError(
				"Attribute's values are not supported for managed OIDC Configuration",
				fmt.Sprintf(
					"In order to create managed OIDC Configuration, "+
						"the attributes' values of `secret_arn`, `issuer_url` and `installer_role_arn` should be empty",
				),
			)
			return
		}
		oidcConfig, err = cmv1.NewOidcConfig().Managed(true).Build()
		if err != nil {
			response.Diagnostics.AddError(
				"There was a problem building the managed OIDC Configuration",
				fmt.Sprintf(
					"There was a problem building the managed OIDC Configuration: %v", err,
				),
			)
			return
		}
	} else {
		if state.SecretARN.IsUnknown() || state.SecretARN.IsNull() ||
			state.IssuerUrl.IsUnknown() || state.IssuerUrl.IsNull() ||
			state.InstallerRoleARN.IsUnknown() || state.InstallerRoleARN.IsNull() {
			response.Diagnostics.AddError(
				"There is a missing parameter for unmanaged OIDC Configuration",
				fmt.Sprintf(
					"There is a missing parameter for unmanaged OIDC Configuration. "+
						"Please provide values for all those attributes `secret_arn`, `issuer_url` and `installer_role_arn`",
				),
			)
			return
		}
		oidcConfig, err = cmv1.NewOidcConfig().
			Managed(false).
			SecretArn(state.SecretARN.ValueString()).
			IssuerUrl(state.IssuerUrl.ValueString()).
			InstallerRoleArn(state.InstallerRoleARN.ValueString()).
			Build()

		if err != nil {
			response.Diagnostics.AddError(
				"There was a problem building the unmanaged OIDC Configuration",
				fmt.Sprintf(
					"There was a problem building the unmanaged OIDC Configuration: %v", err,
				),
			)
			return
		}
	}

	object, err := o.oidcConfigClient.Add().Body(oidcConfig).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"There was a problem registering the OIDC Configuration",
			fmt.Sprintf(
				"There was a problem registering the OIDC Configuration: %v", err,
			),
		)
		return
	}

	oidcConfig = object.Body()

	// Save the state:
	o.populateState(ctx, oidcConfig, state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (o *RosaOidcConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get the current state:
	state := &RosaOidcConfigState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the oidc config:
	get, err := o.oidcConfigClient.OidcConfig(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("oidc config (%s) not found, removing from state",
			state.ID.ValueString(),
		))
		response.State.RemoveResource(ctx)
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Cannot find OIDC config",
			fmt.Sprintf(
				"Cannot find OIDC config with ID %s, %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}

	object := get.Body()

	// Save the state:
	o.populateState(ctx, object, state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (o *RosaOidcConfigResource) Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse) {
	response.Diagnostics.AddError(
		"Update methode is not supported for that resource",
		fmt.Sprintf(
			"Update methode is not supported for that resource",
		),
	)
	return
}

func (o *RosaOidcConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// Get the state:
	state := &RosaOidcConfigState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the oidc config:
	get, err := o.oidcConfigClient.OidcConfig(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Cannot find OIDC config",
			fmt.Sprintf(
				"Cannot find OIDC config with ID %s, %v",
				state.ID.ValueString(), err,
			),
		)
		return
	}

	oidcConfig := get.Body()

	// check if there is a cluster using the oidc endpoint:
	hasClusterUsingOidcConfig, err := o.hasAClusterUsingOidcEndpointUrl(ctx, oidcConfig.IssuerUrl())
	if err != nil {
		response.Diagnostics.AddError(
			"There was a problem checking if any clusters are using OIDC config",
			fmt.Sprintf(
				"There was a problem checking if any clusters are using OIDC config '%s' : %v",
				oidcConfig.IssuerUrl(), err,
			),
		)
		return
	}
	if hasClusterUsingOidcConfig {
		response.Diagnostics.AddError(
			"here are clusters using OIDC config, can't delete the configuration",
			fmt.Sprintf(
				"here are clusters using OIDC config '%s', can't delete the configuration",
				oidcConfig.IssuerUrl(),
			),
		)
		return
	}

	err = o.deleteOidcConfig(ctx, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"There was a problem deleting the OIDC config",
			fmt.Sprintf(
				"There was a problem deleting the OIDC config '%s' : %v",
				oidcConfig.IssuerUrl(), err,
			),
		)
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx)
}

func (o *RosaOidcConfigResource) hasAClusterUsingOidcEndpointUrl(ctx context.Context, issuerUrl string) (bool, error) {
	query := fmt.Sprintf(
		"aws.sts.oidc_endpoint_url = '%s'", issuerUrl,
	)
	request := o.clustersClient.List().Search(query)
	page := 1
	response, err := request.Page(page).SendContext(ctx)
	if err != nil {
		return false, err
	}
	if response.Total() > 0 {
		return true, nil
	}
	return false, nil
}

func (o *RosaOidcConfigResource) deleteOidcConfig(ctx context.Context, id string) error {
	_, err := o.oidcConfigClient.
		OidcConfig(id).
		Delete().
		SendContext(ctx)
	return err
}

func (o *RosaOidcConfigResource) ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(
		ctx,
		path.Root("id"),
		request,
		response,
	)
}

// populateState copies the data from the API object to the Terraform state.
func (o *RosaOidcConfigResource) populateState(ctx context.Context, object *cmv1.OidcConfig, state *RosaOidcConfigState) {
	if id, ok := object.GetID(); ok {
		state.ID = types.StringValue(id)
	}

	if managed, ok := object.GetManaged(); ok {
		state.Managed = types.BoolValue(managed)
	}

	issuerUrl, ok := object.GetIssuerUrl()
	if ok && issuerUrl != "" {
		state.IssuerUrl = types.StringValue(issuerUrl)
	}

	installerRoleArn, ok := object.GetInstallerRoleArn()
	if ok && installerRoleArn != "" {
		state.InstallerRoleARN = types.StringValue(installerRoleArn)
	}

	secretArn, ok := object.GetSecretArn()
	if ok && secretArn != "" {
		state.SecretARN = types.StringValue(secretArn)
	}

	oidcEndpointURL := issuerUrl
	if strings.HasPrefix(oidcEndpointURL, "https://") {
		oidcEndpointURL = strings.TrimPrefix(oidcEndpointURL, "https://")
	}
	state.OIDCEndpointURL = types.StringValue(oidcEndpointURL)

	thumbprint, err := common.GetThumbprint(issuerUrl, common.DefaultHttpClient{})
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("cannot get thumbprint, with error: %v", err))
		state.Thumbprint = types.StringValue("")
	} else {
		state.Thumbprint = types.StringValue(thumbprint)
	}

}
