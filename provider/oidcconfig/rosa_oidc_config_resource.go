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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type RosaOidcConfigResourceType struct {
}

type RosaOidcConfigResource struct {
	oidcConfigClient *cmv1.OidcConfigsClient
	clustersClient   *cmv1.ClustersClient
}

func (t *RosaOidcConfigResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "OIDC config",
		Attributes: map[string]tfsdk.Attribute{
			"managed": {
				Description: "Indicates whether it is a Red Hat managed or unmanaged (Customer hosted) OIDC Configuration",
				Type:        types.BoolType,
				Required:    true,
			},
			"secret_arn": {
				Description: "Indicates for unmanaged OIDC config, the secret ARN",
				Type:        types.StringType,
				Optional:    true,
			},
			"issuer_url": {
				Description: "The bucket URL",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"installer_role_arn": {
				Description: "STS Role ARN with get secrets permission",
				Type:        types.StringType,
				Optional:    true,
			},
			"id": {
				Description: "The OIDC config ID",
				Type:        types.StringType,
				Computed:    true,
			},
			"thumbprint": {
				Description: "SHA1-hash value of the root CA of the issuer URL",
				Type:        types.StringType,
				Computed:    true,
			},
			"oidc_endpoint_url": {
				Description: "OIDC Endpoint URL",
				Type:        types.StringType,
				Computed:    true,
			},
		},
	}
	return
}

func (t *RosaOidcConfigResourceType) NewResource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.Resource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation: use it directly when needed.
	parent := p.(*Provider)

	// Get the oidcConfigClient:
	oidcConfigClient := parent.connection.ClustersMgmt().V1().OidcConfigs()
	// Get the clustersClient:
	clustersClient := parent.connection.ClustersMgmt().V1().Clusters()

	// Create the resource:
	result = &RosaOidcConfigResource{
		oidcConfigClient: oidcConfigClient,
		clustersClient:   clustersClient,
	}

	return
}

func (r *RosaOidcConfigResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	// Get the plan:
	state := &RosaOidcConfigState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	managed := state.Managed.Value
	var oidcConfig *cmv1.OidcConfig
	var err error
	if managed {
		if (!state.SecretARN.Unknown && !state.SecretARN.Null) ||
			(!state.IssuerUrl.Unknown && !state.IssuerUrl.Null) ||
			(!state.InstallerRoleARN.Unknown && !state.InstallerRoleARN.Null) {
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
		if state.SecretARN.Unknown || state.SecretARN.Null ||
			state.IssuerUrl.Unknown || state.IssuerUrl.Null ||
			state.InstallerRoleARN.Unknown || state.InstallerRoleARN.Null {
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
			SecretArn(state.SecretARN.Value).
			IssuerUrl(state.IssuerUrl.Value).
			InstallerRoleArn(state.InstallerRoleARN.Value).
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

	object, err := r.oidcConfigClient.Add().Body(oidcConfig).SendContext(ctx)
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
	r.populateState(ctx, oidcConfig, state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *RosaOidcConfigResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse) {
	// Get the current state:
	state := &RosaOidcConfigState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the oidc config:
	get, err := r.oidcConfigClient.OidcConfig(state.ID.Value).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("oidc config (%s) not found, removing from state",
			state.ID.Value,
		))
		response.State.RemoveResource(ctx)
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Can't find OIDC config",
			fmt.Sprintf(
				"Can't find OIDC config with ID %s, %v",
				state.ID.Value, err,
			),
		)
		return
	}

	object := get.Body()

	// Save the state:
	r.populateState(ctx, object, state)
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *RosaOidcConfigResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse) {
	response.Diagnostics.AddError(
		"Update methode is not supported for that resource",
		fmt.Sprintf(
			"Update methode is not supported for that resource",
		),
	)
	return
}

func (r *RosaOidcConfigResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse) {
	// Get the state:
	state := &RosaOidcConfigState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the oidc config:
	get, err := r.oidcConfigClient.OidcConfig(state.ID.Value).Get().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find OIDC config",
			fmt.Sprintf(
				"Can't find OIDC config with ID %s, %v",
				state.ID.Value, err,
			),
		)
		return
	}

	oidcConfig := get.Body()

	// check if there is a cluster using the oidc endpoint:
	hasClusterUsingOidcConfig, err := r.hasAClusterUsingOidcEndpointUrl(ctx, oidcConfig.IssuerUrl())
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

	err = r.deleteOidcConfig(ctx, state.ID.Value)
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

func (r *RosaOidcConfigResource) hasAClusterUsingOidcEndpointUrl(ctx context.Context, issuerUrl string) (bool, error) {
	query := fmt.Sprintf(
		"aws.sts.oidc_endpoint_url = '%s'", issuerUrl,
	)
	request := r.clustersClient.List().Search(query)
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

func (r *RosaOidcConfigResource) deleteOidcConfig(ctx context.Context, id string) error {
	_, err := r.oidcConfigClient.
		OidcConfig(id).
		Delete().
		SendContext(ctx)
	return err
}

func (r *RosaOidcConfigResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(
		ctx,
		tftypes.NewAttributePath().WithAttributeName("id"),
		request,
		response,
	)
}

// populateState copies the data from the API object to the Terraform state.
func (r *RosaOidcConfigResource) populateState(ctx context.Context, object *cmv1.OidcConfig, state *RosaOidcConfigState) {
	state.ID = types.String{
		Value: object.ID(),
	}
	state.Managed = types.Bool{
		Value: object.Managed(),
	}

	issuerUrl, ok := object.GetIssuerUrl()
	if ok && issuerUrl != "" {
		state.IssuerUrl = types.String{
			Value: issuerUrl,
		}
	}

	installerRoleArn, ok := object.GetInstallerRoleArn()
	if ok && installerRoleArn != "" {
		state.InstallerRoleARN = types.String{
			Value: installerRoleArn,
		}
	}

	secretArn, ok := object.GetSecretArn()
	if ok && secretArn != "" {
		state.SecretARN = types.String{
			Value: secretArn,
		}
	}
	oidcEndpointURL := issuerUrl
	if strings.HasPrefix(oidcEndpointURL, "https://") {
		oidcEndpointURL = strings.TrimPrefix(oidcEndpointURL, "https://")
	}
	state.OIDCEndpointURL = types.String{
		Value: oidcEndpointURL,
	}

	thumbprint, err := getThumbprint(issuerUrl, DefaultHttpClient{})
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("cannot get thumbprint, with error: %v", err))
		state.Thumbprint = types.String{
			Value: "",
		}
	} else {
		state.Thumbprint = types.String{
			Value: thumbprint,
		}
	}

}
