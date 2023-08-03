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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/oidcconfig/oidcconfigscehma"
	"net/http"
	"strings"
)

func ResourceOidcConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOidcConfigCreate,
		ReadContext:   resourceOidcConfigRead,
		UpdateContext: nil,
		DeleteContext: resourceOidcConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: oidcconfigscehma.OidcConfigFields(),
	}
}

func resourceOidcConfigCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin create()")
	// Get the oidc client:
	oidcConfigClient := meta.(*sdk.Connection).ClustersMgmt().V1().OidcConfigs()

	oidcConfigState := oidcConfigFromResourceData(resourceData)
	var oidcConfig *cmv1.OidcConfig
	var err error
	if oidcConfigState.Managed {
		if !common2.IsStringAttributeEmpty(oidcConfigState.SecretARN) ||
			!common2.IsStringAttributeEmpty(oidcConfigState.IssuerUrl) ||
			!common2.IsStringAttributeEmpty(oidcConfigState.InstallerRoleARN) {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Attribute's values are not supported for managed OIDC Configuration",
					Detail: fmt.Sprintf(
						"In order to create managed OIDC Configuration, " +
							"the attributes' values of `secret_arn`, `issuer_url` and `installer_role_arn` should be empty",
					),
				}}
		}

		oidcConfig, err = cmv1.NewOidcConfig().Managed(true).Build()
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "There was a problem building the managed OIDC Configuration",
					Detail: fmt.Sprintf(
						"There was a problem building the managed OIDC Configuration: %v", err,
					),
				}}
		}
	} else {
		if common2.IsStringAttributeEmpty(oidcConfigState.SecretARN) ||
			common2.IsStringAttributeEmpty(oidcConfigState.IssuerUrl) ||
			common2.IsStringAttributeEmpty(oidcConfigState.InstallerRoleARN) {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "There is a missing parameter for unmanaged OIDC Configuration",
					Detail: fmt.Sprintf(
						"There is a missing parameter for unmanaged OIDC Configuration. " +
							"Please provide values for all those attributes `secret_arn`, `issuer_url` and `installer_role_arn`",
					),
				}}
		}
		oidcConfig, err = cmv1.NewOidcConfig().
			Managed(false).
			SecretArn(*oidcConfigState.SecretARN).
			IssuerUrl(*oidcConfigState.IssuerUrl).
			InstallerRoleArn(*oidcConfigState.InstallerRoleARN).
			Build()

		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "There was a problem building the unmanaged OIDC Configuration",
					Detail: fmt.Sprintf(
						"There was a problem building the unmanaged OIDC Configuration: %v", err,
					),
				}}
		}
	}

	object, err := oidcConfigClient.Add().Body(oidcConfig).SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "There was a problem registering the OIDC Configuration",
				Detail: fmt.Sprintf(
					"There was a problem registering the OIDC Configuration: %v", err,
				),
			}}
	}

	oidcConfig = object.Body()
	oidcConfigToResourceData(ctx, oidcConfig, resourceData)
	return nil
}

func resourceOidcConfigRead(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin read()")
	// Get the OIDC config client:
	oidcConfigClient := meta.(*sdk.Connection).ClustersMgmt().V1().OidcConfigs()

	oidcConfigState := oidcConfigFromResourceData(resourceData)
	// Find the oidc config:
	get, err := oidcConfigClient.OidcConfig(oidcConfigState.ID).Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf(fmt.Sprintf("OIDC config (%s) not found, removing from state",
			oidcConfigState.ID,
		)))
		resourceData.SetId("")
		return
	} else if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't find OIDC config",
				Detail: fmt.Sprintf(
					"Can't find OIDC config with ID %s, %v",
					oidcConfigState.ID, err,
				),
			}}
	}

	oidcConfig := get.Body()
	oidcConfigToResourceData(ctx, oidcConfig, resourceData)
	return nil
}

func resourceOidcConfigUpdate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin update()")
	return []diag.Diagnostic{
		{
			Severity: diag.Error,
			Summary:  "Update methode is not supported for that resource",
			Detail: fmt.Sprintf(
				"Update methode is not supported for that resource",
			),
		}}
}

func resourceOidcConfigDelete(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin delete()")
	// Get the cluster collection
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	// Get the OIDC config client:
	oidcConfigClient := meta.(*sdk.Connection).ClustersMgmt().V1().OidcConfigs()

	oidcConfigState := oidcConfigFromResourceData(resourceData)
	// Find the oidc config:
	get, err := oidcConfigClient.OidcConfig(oidcConfigState.ID).Get().SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't find OIDC config",
				Detail: fmt.Sprintf(
					"Can't find OIDC config with ID %s, %v",
					oidcConfigState.ID, err,
				),
			}}
	}

	oidcConfig := get.Body()

	// check if there is a cluster using the oidc endpoint:
	hasClusterUsingOidcConfig, err := hasAClusterUsingOidcEndpointUrl(ctx, oidcConfig.IssuerUrl(), clusterCollection)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "There was a problem checking if any clusters are using OIDC config",
				Detail: fmt.Sprintf(
					"There was a problem checking if any clusters are using OIDC config '%s' : %v",
					oidcConfig.IssuerUrl(), err,
				),
			}}
	}

	if hasClusterUsingOidcConfig {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "here are clusters using OIDC config, can't delete the configuration",
				Detail: fmt.Sprintf(
					"here are clusters using OIDC config '%s', can't delete the configuration",
					oidcConfig.IssuerUrl(),
				),
			}}
	}

	err = deleteOidcConfig(ctx, oidcConfigState.ID, oidcConfigClient)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "There was a problem deleting the OIDC config",
				Detail: fmt.Sprintf(
					"There was a problem deleting the OIDC config '%s' : %v",
					oidcConfig.IssuerUrl(), err,
				),
			}}
	}

	resourceData.SetId("")
	return
}

func oidcConfigToResourceData(ctx context.Context, object *cmv1.OidcConfig, resourceData *schema.ResourceData) {
	resourceData.SetId(object.ID())

	resourceData.Set("managed", object.Managed())

	if installerRoleArn, ok := object.GetInstallerRoleArn(); ok && installerRoleArn != "" {
		resourceData.Set("installer_role_arn", installerRoleArn)
	}

	if secretArn, ok := object.GetSecretArn(); ok && secretArn != "" {
		resourceData.Set("secret_arn", secretArn)
	}

	issuerUrl := object.IssuerUrl()
	resourceData.Set("issuer_url", issuerUrl)
	oidcEndpointURL := issuerUrl
	if strings.HasPrefix(oidcEndpointURL, "https://") {
		oidcEndpointURL = strings.TrimPrefix(oidcEndpointURL, "https://")
	}
	resourceData.Set("oidc_endpoint_url", oidcEndpointURL)

	thumbprint, err := common2.GetThumbprint(issuerUrl, common2.DefaultHttpClient{})
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("cannot get thumbprint, with error: %v", err))
		resourceData.Set("thumbprint", "")
	} else {
		resourceData.Set("thumbprint", thumbprint)
	}
}
func oidcConfigFromResourceData(resourceData *schema.ResourceData) *oidcconfigscehma.RosaOidcConfigState {
	result := &oidcconfigscehma.RosaOidcConfigState{
		Managed: resourceData.Get("managed").(bool),
	}

	result.SecretARN = common2.GetOptionalString(resourceData, "secret_arn")
	result.IssuerUrl = common2.GetOptionalString(resourceData, "issuer_url")
	result.InstallerRoleARN = common2.GetOptionalString(resourceData, "installer_role_arn")

	result.ID = resourceData.Id()
	if thumbprint := common2.GetOptionalString(resourceData, "thumbprint"); thumbprint != nil {
		result.Thumbprint = *thumbprint
	}

	if oidcEnpointUrl := common2.GetOptionalString(resourceData, "oidc_endpoint_url"); oidcEnpointUrl != nil {
		result.OIDCEndpointURL = *oidcEnpointUrl
	}
	return result
}

func hasAClusterUsingOidcEndpointUrl(ctx context.Context, issuerUrl string, clustersClient *cmv1.ClustersClient) (bool, error) {
	query := fmt.Sprintf(
		"aws.sts.oidc_endpoint_url = '%s'", issuerUrl,
	)
	request := clustersClient.List().Search(query)
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

func deleteOidcConfig(ctx context.Context, id string, oidcConfigClient *cmv1.OidcConfigsClient) error {
	_, err := oidcConfigClient.
		OidcConfig(id).
		Delete().
		SendContext(ctx)
	return err
}
