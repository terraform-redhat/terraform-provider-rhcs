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
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/cluster"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterautoscaler"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/classic"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/hcp"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterwaiter"
	defaultingress "github.com/terraform-redhat/terraform-provider-rhcs/provider/defaultingress/classic"
	hcpingress "github.com/terraform-redhat/terraform-provider-rhcs/provider/defaultingress/hcp"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/dnsdomain"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/group"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/groupmembership"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/identityprovider"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/info"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/kubeletconfig"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/machine_types"
	machinepool "github.com/terraform-redhat/terraform-provider-rhcs/provider/machinepool/classic"
	nodepool "github.com/terraform-redhat/terraform-provider-rhcs/provider/machinepool/hcp"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/ocm_policies"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/oidcconfig"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/oidcconfiginput"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/rosa_operator_roles"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/tuningconfigs"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/versions"
)

// Provider is the implementation of the Provider.
type Provider struct{}

var _ tfprovider.Provider = &Provider{}

// Config contains the configuration of the provider.
type Config struct {
	URL          types.String `tfsdk:"url"`
	TokenURL     types.String `tfsdk:"token_url"`
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
				Description: fmt.Sprintf("URL sets the base URL of the API gateway. The default is `%s`", sdk.DefaultURL),
				Optional:    true,
			},
			"token_url": tfpschema.StringAttribute{
				Description: fmt.Sprintf("TokenURL returns the URL that the connection is using request OpenID access tokens. The default value is '%s'", sdk.DefaultTokenURL),
				Optional:    true,
			},
			"token": tfpschema.StringAttribute{
				Description: "Access or refresh token that is " +
					"generated from https://console.redhat.com/openshift/token/rosa.",
				Optional:  true,
				Sensitive: true,
			},
			"client_id": tfpschema.StringAttribute{
				Description: fmt.Sprintf("OpenID client identifier. The default value is '%s'.", sdk.DefaultClientID),
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

func (p *Provider) getAttrValueOrConfig(attr types.String, envSuffix string) (string, bool) {
	if !attr.IsNull() {
		return attr.ValueString(), true
	}
	if value, ok := os.LookupEnv(fmt.Sprintf("RHCS_%s", envSuffix)); ok {
		return value, true
	}
	return "", false
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
	if url, ok := p.getAttrValueOrConfig(config.URL, "URL"); ok {
		builder.URL(url)
	}
	if tokenURL, ok := p.getAttrValueOrConfig(config.TokenURL, "TOKEN_URL"); ok {
		builder.TokenURL(tokenURL)
	}
	if token, ok := p.getAttrValueOrConfig(config.Token, "TOKEN"); ok {
		builder.Tokens(token)
	}
	clientID, clientIdExists := p.getAttrValueOrConfig(config.ClientID, "CLIENT_ID")
	clientSecret, clientSecretExists := p.getAttrValueOrConfig(config.ClientSecret, "CLIENT_SECRET")
	if clientIdExists && clientSecretExists {
		builder.Client(clientID, clientSecret)
	}
	if trustedCAs, ok := p.getAttrValueOrConfig(config.TrustedCAs, "TRUSTED_CAS"); ok {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(trustedCAs)) {
			resp.Diagnostics.AddError(
				"the value of 'trusted_cas' doesn't contain any certificate",
				"",
			)
			return
		}
		builder.TrustedCAs(pool)
	}
	if !config.Insecure.IsNull() {
		builder.Insecure(config.Insecure.ValueBool())
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
		classic.New,
		identityprovider.New,
		cluster.New,
		clusterautoscaler.New,
		defaultingress.New,
		kubeletconfig.New,
		hcp.New,
		nodepool.New,
		hcpingress.New,
		tuningconfigs.New,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cloudprovider.New,
		group.New,
		machine_types.New,
		ocm_policies.New,
		rosa_operator_roles.New,
		versions.New,
		info.New,
		classic.NewDataSource,
		machinepool.NewDatasource,
		hcp.NewDataSource,
		nodepool.NewDatasource,
	}
}
