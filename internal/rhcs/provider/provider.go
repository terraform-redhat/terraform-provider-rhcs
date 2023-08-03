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
	"crypto/x509"
	"fmt"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/cloudprovider"
	cluster2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/cluster"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/dnsdomain"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/groupmembership"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/groups"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/idps"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/machinepool"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/machinetypes"
	oidcconfig2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/oidcconfig"
	rolesandpolicies2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/rolesandpolicies"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/versions"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	"github.com/terraform-redhat/terraform-provider-rhcs/logging"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Description: "URL of the API server.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"token_url": {
				Description: "OpenID token URL.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"user": {
				Description: "User name.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"password": {
				Description: "User password.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"token": {
				Description: "Access or refresh token that is " +
					"generated from https://console.redhat.com/openshift/token/rosa.",
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"client_id": {
				Description: "OpenID client identifier.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"client_secret": {
				Description: "OpenID client secret.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"trusted_cas": {
				Description: "PEM encoded certificates of authorities that will " +
					"be trusted. If this is not explicitly specified, then " +
					"the clusterservice will trust the certificate authorities " +
					"trusted by default by the system.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"insecure": {
				Description: "When set to 'true' enables insecure communication " +
					"with the server. This disables verification of TLS " +
					"certificates and host names, and it is not recommended " +
					"for production environments.",
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
		DataSourcesMap: DatasourceMap(),
		ResourcesMap:   ResourceMap(),
	}
	p.ConfigureContextFunc = func(ctx context.Context, resourceData *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(ctx, resourceData, terraformVersion)
	}
	return p
}

func DatasourceMap() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"rhcs_cloud_providers":     cloudprovider.CloudProvidersDataSource(),
		"rhcs_rosa_operator_roles": rolesandpolicies2.OperatorRolesDataSource(),
		"rhcs_policies":            rolesandpolicies2.OcmPoliciesDataSource(),
		"rhcs_groups":              groups.GroupsDataSource(),
		"rhcs_machine_types":       machinetypes.MachineTypesDataSource(),
		"rhcs_versions":            versions.VersionsDataSourc(),
	}
}
func ResourceMap() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		"rhcs_cluster":                cluster2.ResourceCluster(),
		"rhcs_cluster_rosa_classic":   cluster2.ResourceClusterRosaClassic(),
		"rhcs_group_membership":       groupmembership.ResourceGroupMembership(),
		"rhcs_identity_provider":      idps.ResourceIdentityProvider(),
		"rhcs_machine_pool":           machinepool.ResourceMachinePool(),
		"rhcs_cluster_wait":           cluster2.ResourceClusterWaiter(),
		"rhcs_rosa_oidc_config_input": oidcconfig2.ResourceOidcConfigInput(),
		"rhcs_rosa_oidc_config":       oidcconfig2.ResourceOidcConfig(),
		"rhcs_dns_domain":             dnsdomain.ResourceDNSDomain(),
	}
}

func providerConfigure(ctx context.Context, resourceData *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// The plugin infrastructure redirects the log package output so that it is sent to the main
	// Terraform process, so if we want to have the logs of the SDK redirected we need to use
	// the log package as well.
	logger := logging.New()

	// Create the builder:
	builder := sdk.NewConnectionBuilder()
	builder.Logger(logger)
	builder.Agent(fmt.Sprintf("OCM-TF/%s-%s", build.Version, build.Commit))

	if url, ok := resourceData.GetOk("url"); ok {
		builder.URL(url.(string))
	} else {
		ocmURL, ok := os.LookupEnv("OCM_URL")
		if ok {
			builder.URL(ocmURL)
		}
	}

	if tokenURL, ok := resourceData.GetOk("token_url"); ok {
		builder.TokenURL(tokenURL.(string))
	}

	if user, ok := resourceData.GetOk("user"); ok {
		if password, ok := resourceData.GetOk("password"); ok {
			builder.User(user.(string), password.(string))
		}
	}

	if token, ok := resourceData.GetOk("token"); ok {
		builder.Tokens(token.(string))
	} else {
		ocmToken, ok := os.LookupEnv("OCM_TOKEN")
		if ok {
			builder.Tokens(ocmToken)
		}
	}

	if clientID, ok := resourceData.GetOk("client_id"); ok {
		if clientSecret, ok := resourceData.GetOk("client_secret"); ok {
			builder.Client(clientID.(string), clientSecret.(string))
		}
	}

	if insecure, ok := resourceData.GetOk("insecure"); ok {
		builder.Insecure(insecure.(bool))
	}

	if trustedCAs, ok := resourceData.GetOk("trusted_cas"); ok {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(trustedCAs.(string))) {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create OCM client",
				Detail:   "the value of 'trusted_cas' doesn't contain any certificate",
			})
			return nil, diags
		}
		builder.TrustedCAs(pool)
	}
	// Create the connection:
	connection, err := builder.BuildContext(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create OCM client",
			Detail:   err.Error(),
		})
		return nil, diags
	}

	return connection, diags
}
