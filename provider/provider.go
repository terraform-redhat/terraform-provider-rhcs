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
	"crypto/x509"
***REMOVED***
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	"github.com/terraform-redhat/terraform-provider-rhcs/logging"
***REMOVED***

// Provider is the implementation of the Provider.
type Provider struct {
	connection *sdk.Connection
}

// Config contains the configuration of the provider.
type Config struct {
	URL          types.String `tfsdk:"url"`
	TokenURL     types.String `tfsdk:"token_url"`
	User         types.String `tfsdk:"user"`
	Password     types.String `tfsdk:"password"`
	Token        types.String `tfsdk:"token"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	TrustedCAs   types.String `tfsdk:"trusted_cas"`
	Insecure     types.Bool   `tfsdk:"insecure"`
}

// New creates the provider.
func New(***REMOVED*** tfsdk.Provider {
	return &Provider{}
}

// Provider creates the schema for the provider.
func (p *Provider***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (schema tfsdk.Schema, diags diag.Diagnostics***REMOVED*** {
	schema = tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"url": {
				Description: "URL of the API server.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"token_url": {
				Description: "OpenID token URL.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"user": {
				Description: "User name.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"password": {
				Description: "User password.",
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"token": {
				Description: "Access or refresh token that is " +
					"generated from https://console.redhat.com/openshift/token/rosa.",
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
	***REMOVED***,
			"client_id": {
				Description: "OpenID client identifier.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"client_secret": {
				Description: "OpenID client secret.",
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"trusted_cas": {
				Description: "PEM encoded certificates of authorities that will " +
					"be trusted. If this is not explicitly specified, then " +
					"the provider will trust the certificate authorities " +
					"trusted by default by the system.",
				Type:     types.StringType,
				Optional: true,
	***REMOVED***,
			"insecure": {
				Description: "When set to 'true' enables insecure communication " +
					"with the server. This disables verification of TLS " +
					"certificates and host names, and it is not recommended " +
					"for production environments.",
				Type:     types.BoolType,
				Optional: true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

// configure is the configuration function of the provider. It is responsible for checking the
// connection parameters and creating the connection that will be used by the resources.
func (p *Provider***REMOVED*** Configure(ctx context.Context, request tfsdk.ConfigureProviderRequest,
	response *tfsdk.ConfigureProviderResponse***REMOVED*** {
	// Retrieve the provider configuration:
	var config Config
	diags := request.Config.Get(ctx, &config***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// The plugin infrastructure redirects the log package output so that it is sent to the main
	// Terraform process, so if we want to have the logs of the SDK redirected we need to use
	// the log package as well.
	logger := logging.New(***REMOVED***

	// Create the builder:
	builder := sdk.NewConnectionBuilder(***REMOVED***
	builder.Logger(logger***REMOVED***
	builder.Agent(fmt.Sprintf("OCM-TF/%s-%s", build.Version, build.Commit***REMOVED******REMOVED***

	// Copy the settings:
	if !config.URL.Null {
		builder.URL(config.URL.Value***REMOVED***
	} else {
		url, ok := os.LookupEnv("OCM_URL"***REMOVED***
		if ok {
			builder.URL(url***REMOVED***
***REMOVED***
	}
	if !config.TokenURL.Null {
		builder.TokenURL(config.TokenURL.Value***REMOVED***
	}
	if !config.User.Null && !config.Password.Null {
		builder.User(config.User.Value, config.Password.Value***REMOVED***
	}
	if !config.Token.Null {
		builder.Tokens(config.Token.Value***REMOVED***
	} else {
		token, ok := os.LookupEnv("OCM_TOKEN"***REMOVED***
		if ok {
			builder.Tokens(token***REMOVED***
***REMOVED***
	}
	if !config.ClientID.Null && !config.ClientSecret.Null {
		builder.Client(config.ClientID.Value, config.ClientSecret.Value***REMOVED***
	}
	if !config.Insecure.Null {
		builder.Insecure(config.Insecure.Value***REMOVED***
	}
	if !config.TrustedCAs.Null {
		pool := x509.NewCertPool(***REMOVED***
		if !pool.AppendCertsFromPEM([]byte(config.TrustedCAs.Value***REMOVED******REMOVED*** {
			response.Diagnostics.AddError(
				"the value of 'trusted_cas' doesn't contain any certificate",
				"",
			***REMOVED***
			return
***REMOVED***
		builder.TrustedCAs(pool***REMOVED***
	}

	// Create the connection:
	connection, err := builder.BuildContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(err.Error(***REMOVED***, ""***REMOVED***
		return
	}

	// Save the connection:
	p.connection = connection
}

// GetResources returns the resources supported by the provider.
func (p *Provider***REMOVED*** GetResources(ctx context.Context***REMOVED*** (result map[string]tfsdk.ResourceType,
	diags diag.Diagnostics***REMOVED*** {
	result = map[string]tfsdk.ResourceType{
		"rhcs_cluster":                &ClusterResourceType{},
		"rhcs_cluster_rosa_classic":   &ClusterRosaClassicResourceType{},
		"rhcs_group_membership":       &GroupMembershipResourceType{},
		"rhcs_identity_provider":      &IdentityProviderResourceType{},
		"rhcs_machine_pool":           &MachinePoolResourceType{},
		"rhcs_cluster_wait":           &ClusterWaiterResourceType{},
		"rhcs_rosa_oidc_config_input": &RosaOidcConfigInputResourceType{},
		"rhcs_rosa_oidc_config":       &RosaOidcConfigResourceType{},
	}
	return
}

// GetDataSources returns the data sources supported by the provider.
func (p *Provider***REMOVED*** GetDataSources(ctx context.Context***REMOVED*** (result map[string]tfsdk.DataSourceType,
	diags diag.Diagnostics***REMOVED*** {
	result = map[string]tfsdk.DataSourceType{
		"rhcs_cloud_providers":     &CloudProvidersDataSourceType{},
		"rhcs_rosa_operator_roles": &RosaOperatorRolesDataSourceType{},
		"rhcs_policies":            &OcmPoliciesDataSourceType{},
		"rhcs_groups":              &GroupsDataSourceType{},
		"rhcs_machine_types":       &MachineTypesDataSourceType{},
		"rhcs_versions":            &VersionsDataSourceType{},
	}
	return
}
