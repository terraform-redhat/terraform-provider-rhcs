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
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfpschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	"github.com/terraform-redhat/terraform-provider-rhcs/logging"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/cloudprovider"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterwaiter"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/dnsdomain"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/group"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/groupmembership"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/machinepool"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/machine_types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/oidcconfig"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/oidcconfiginput"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/versions"
)

// Provider is the implementation of the Provider.
type Provider struct{}

var _ tfprovider.Provider = &Provider{}

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
func New() tfprovider.Provider {
	return &Provider{}
}

func (p *Provider) Metadata(ctx context.Context, req tfprovider.MetadataRequest, resp *tfprovider.MetadataResponse) {
	resp.TypeName = "rhcs"
	resp.Version = build.Version
}

// Provider creates the schema for the provider.
func (p *Provider) Schema(ctx context.Context, req tfprovider.SchemaRequest, resp *tfprovider.SchemaResponse) {
	resp.Schema = tfpschema.Schema{
		Attributes: map[string]tfpschema.Attribute{
			"url": tfpschema.StringAttribute{
				Description: "URL of the API server.",
				Optional:    true,
			},
			"token_url": tfpschema.StringAttribute{
				Description: "OpenID token URL.",
				Optional:    true,
			},
			"user": tfpschema.StringAttribute{
				Description: "User name.",
				Optional:    true,
			},
			"password": tfpschema.StringAttribute{
				Description: "User password.",
				Optional:    true,
				Sensitive:   true,
			},
			"token": tfpschema.StringAttribute{
				Description: "Access or refresh token that is " +
					"generated from https://console.redhat.com/openshift/token/rosa.",
				Optional:  true,
				Sensitive: true,
			},
			"client_id": tfpschema.StringAttribute{
				Description: "OpenID client identifier.",
				Optional:    true,
			},
			"client_secret": tfpschema.StringAttribute{
				Description: "OpenID client secret.",
				Optional:    true,
				Sensitive:   true,
			},
			"trusted_cas": tfpschema.StringAttribute{
				Description: "PEM encoded certificates of authorities that will " +
					"be trusted. If this is not explicitly specified, then " +
					"the provider will trust the certificate authorities " +
					"trusted by default by the system.",
				Optional: true,
			},
			"insecure": tfpschema.BoolAttribute{
				Description: "When set to 'true' enables insecure communication " +
					"with the server. This disables verification of TLS " +
					"certificates and host names, and it is not recommended " +
					"for production environments.",
				Optional: true,
			},
		},
	}
}

// configure is the configuration function of the provider. It is responsible for checking the
// connection parameters and creating the connection that will be used by the resources.
func (p *Provider) Configure(ctx context.Context, req tfprovider.ConfigureRequest,
	resp *tfprovider.ConfigureResponse) {
	// Retrieve the provider configuration:
	var config Config
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The plugin infrastructure redirects the log package output so that it is sent to the main
	// Terraform process, so if we want to have the logs of the SDK redirected we need to use
	// the log package as well.
	logger := logging.New()

	// Create the builder:
	builder := sdk.NewConnectionBuilder()
	builder.Logger(logger)
	builder.Agent(fmt.Sprintf("OCM-TF/%s-%s", build.Version, build.Commit))

	// Copy the settings:
	if !config.URL.IsNull() {
		builder.URL(config.URL.ValueString())
	} else {
		url, ok := os.LookupEnv("RHCS_URL")
		if ok {
			builder.URL(url)
		}
	}
	if !config.TokenURL.IsNull() {
		builder.TokenURL(config.TokenURL.ValueString())
	}
	if !config.User.IsNull() && !config.Password.IsNull() {
		builder.User(config.User.ValueString(), config.Password.ValueString())
	}
	if !config.Token.IsNull() {
		builder.Tokens(config.Token.ValueString())
	} else {
		token, ok := os.LookupEnv("RHCS_TOKEN")
		if ok {
			builder.Tokens(token)
		}
	}
	if !config.ClientID.IsNull() && !config.ClientSecret.IsNull() {
		builder.Client(config.ClientID.ValueString(), config.ClientSecret.ValueString())
	}
	if !config.Insecure.IsNull() {
		builder.Insecure(config.Insecure.ValueBool())
	}
	if !config.TrustedCAs.IsNull() {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(config.TrustedCAs.ValueString())) {
			resp.Diagnostics.AddError(
				"the value of 'trusted_cas' doesn't contain any certificate",
				"",
			)
			return
		}
		builder.TrustedCAs(pool)
	}

	// Create the connection:
	connection, err := builder.BuildContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	// Save the connection:
	resp.DataSourceData = connection
	resp.ResourceData = connection
}

// Resources returns the resources supported by the provider.
func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		clusterwaiter.New,
		dnsdomain.New,
		groupmembership.New,
		machinepool.New,
		oidcconfig.New,
		oidcconfiginput.New,
		// TODO uncomment this after ClusterRosaClassic resource is fixed
		// clusterwaiter.NewClusterRosaClassicResource,
	}
}

// func (p *Provider) Resources(ctx context.Context) (result map[string]resource.Resource,
// 	diags diag.Diagnostics) {
// 	result = map[string]resource.Resource{
// 		"rhcs_cluster":                &ClusterResourceType{},
// 		"rhcs_cluster_rosa_classic":   &ClusterRosaClassicResourceType{},
// 		"rhcs_group_membership":       &GroupMembershipResourceType{},
// 		"rhcs_identity_provider":      &IdentityProviderResourceType{},
// 		"rhcs_machine_pool":           &MachinePoolResourceType{},
// 		"rhcs_cluster_wait":           &ClusterWaiterResourceType{},
// 		"rhcs_rosa_oidc_config_input": &RosaOidcConfigInputResourceType{},
// 		"rhcs_rosa_oidc_config":       &RosaOidcConfigResourceType{},
// 		"rhcs_dns_domain":             &DNSDomainResourceType{},
// 	}
// 	return
// }

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cloudprovider.New,
		versions.New,
		group.New,
		machine_types.New,
	}
}

// GetDataSources returns the data sources supported by the provider.
// func (p *Provider) DataSources(ctx context.Context) (result map[string]datasource.DataSource,
// 	diags diag.Diagnostics) {
// 	result = map[string]datasource.DataSource{
// 		"rhcs_cloud_providers":     &CloudProvidersDataSourceType{},
// 		"rhcs_rosa_operator_roles": &RosaOperatorRolesDataSourceType{},
// 		"rhcs_policies":            &OcmPoliciesDataSourceType{},
// 		"rhcs_groups":              &GroupsDataSourceType{},
// 		"rhcs_machine_types":       &MachineTypesDataSourceType{},
// 		"rhcs_versions":            &VersionsDataSourceType{},
// 	}
// 	return
// }
