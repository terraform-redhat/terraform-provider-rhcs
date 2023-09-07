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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
func New(***REMOVED*** tfprovider.Provider {
	return &Provider{}
}

// Provider creates the schema for the provider.
func (p *Provider***REMOVED*** Schema(ctx context.Context***REMOVED*** (schema tfschema.Schema, diags diag.Diagnostics***REMOVED*** {
	schema = tfschema.Schema{
		Attributes: map[string]tfschema.Attribute{
			"url": tfschema.StringAttribute{
				Description: "URL of the API server.",
				Optional:    true,
	***REMOVED***,
			"token_url": tfschema.StringAttribute{
				Description: "OpenID token URL.",
				Optional:    true,
	***REMOVED***,
			"user": tfschema.StringAttribute{
				Description: "User name.",
				Optional:    true,
	***REMOVED***,
			"password": tfschema.StringAttribute{
				Description: "User password.",
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"token": tfschema.StringAttribute{
				Description: "Access or refresh token that is " +
					"generated from https://console.redhat.com/openshift/token/rosa.",
				Optional:  true,
				Sensitive: true,
	***REMOVED***,
			"client_id": tfschema.StringAttribute{
				Description: "OpenID client identifier.",
				Optional:    true,
	***REMOVED***,
			"client_secret": tfschema.StringAttribute{
				Description: "OpenID client secret.",
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"trusted_cas": tfschema.StringAttribute{
				Description: "PEM encoded certificates of authorities that will " +
					"be trusted. If this is not explicitly specified, then " +
					"the provider will trust the certificate authorities " +
					"trusted by default by the system.",
				Optional: true,
	***REMOVED***,
			"insecure": tfschema.BoolAttribute{
				Description: "When set to 'true' enables insecure communication " +
					"with the server. This disables verification of TLS " +
					"certificates and host names, and it is not recommended " +
					"for production environments.",
				Optional: true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

// configure is the configuration function of the provider. It is responsible for checking the
// connection parameters and creating the connection that will be used by the resources.
func (p *Provider***REMOVED*** Configure(ctx context.Context, request tfprovider.ConfigureRequest,
	response *tfprovider.ConfigureResponse***REMOVED*** {
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
	if !config.URL.IsNull(***REMOVED*** {
		builder.URL(config.URL.ValueString(***REMOVED******REMOVED***
	} else {
		url, ok := os.LookupEnv("RHCS_URL"***REMOVED***
		if ok {
			builder.URL(url***REMOVED***
***REMOVED***
	}
	if !config.TokenURL.IsNull(***REMOVED*** {
		builder.TokenURL(config.TokenURL.ValueString(***REMOVED******REMOVED***
	}
	if !config.User.IsNull(***REMOVED*** && !config.Password.IsNull(***REMOVED*** {
		builder.User(config.User.ValueString(***REMOVED***, config.Password.ValueString(***REMOVED******REMOVED***
	}
	if !config.Token.IsNull(***REMOVED*** {
		builder.Tokens(config.Token.ValueString(***REMOVED******REMOVED***
	} else {
		token, ok := os.LookupEnv("RHCS_TOKEN"***REMOVED***
		if ok {
			builder.Tokens(token***REMOVED***
***REMOVED***
	}
	if !config.ClientID.IsNull(***REMOVED*** && !config.ClientSecret.IsNull(***REMOVED*** {
		builder.Client(config.ClientID.ValueString(***REMOVED***, config.ClientSecret.ValueString(***REMOVED******REMOVED***
	}
	if !config.Insecure.IsNull(***REMOVED*** {
		builder.Insecure(config.Insecure.ValueBool(***REMOVED******REMOVED***
	}
	if !config.TrustedCAs.IsNull(***REMOVED*** {
		pool := x509.NewCertPool(***REMOVED***
		if !pool.AppendCertsFromPEM([]byte(config.TrustedCAs.ValueString(***REMOVED******REMOVED******REMOVED*** {
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

// Resources returns the resources supported by the provider.
func (p *Provider***REMOVED*** Resources(ctx context.Context***REMOVED*** []func(***REMOVED*** resource.Resource {
	return []func(***REMOVED*** resource.Resource{}
}

func (p *Provider***REMOVED*** Resources(ctx context.Context***REMOVED*** (result map[string]resource.Resource,
	diags diag.Diagnostics***REMOVED*** {
	result = map[string]resource.Resource{
		"rhcs_cluster":                &ClusterResourceType{},
		"rhcs_cluster_rosa_classic":   &ClusterRosaClassicResourceType{},
		"rhcs_group_membership":       &GroupMembershipResourceType{},
		"rhcs_identity_provider":      &IdentityProviderResourceType{},
		"rhcs_machine_pool":           &MachinePoolResourceType{},
		"rhcs_cluster_wait":           &ClusterWaiterResourceType{},
		"rhcs_rosa_oidc_config_input": &RosaOidcConfigInputResourceType{},
		"rhcs_rosa_oidc_config":       &RosaOidcConfigResourceType{},
		"rhcs_dns_domain":             &DNSDomainResourceType{},
	}
	return
}

// GetDataSources returns the data sources supported by the provider.
func (p *Provider***REMOVED*** DataSources(ctx context.Context***REMOVED*** (result map[string]datasource.DataSource,
	diags diag.Diagnostics***REMOVED*** {
	result = map[string]datasource.DataSource{
		"rhcs_cloud_providers":     &CloudProvidersDataSourceType{},
		"rhcs_rosa_operator_roles": &RosaOperatorRolesDataSourceType{},
		"rhcs_policies":            &OcmPoliciesDataSourceType{},
		"rhcs_groups":              &GroupsDataSourceType{},
		"rhcs_machine_types":       &MachineTypesDataSourceType{},
		"rhcs_versions":            &VersionsDataSourceType{},
	}
	return
}
