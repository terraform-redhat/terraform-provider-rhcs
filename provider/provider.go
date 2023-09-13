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
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfpschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	"github.com/terraform-redhat/terraform-provider-rhcs/logging"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/cloudprovider"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/cluster"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosaclassic"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterwaiter"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/dnsdomain"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/group"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/groupmembership"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/identityprovider"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/machine_types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/machinepool"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/ocm_policies"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/oidcconfig"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/oidcconfiginput"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/rosa_operator_roles"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/versions"
***REMOVED***

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
func New(***REMOVED*** tfprovider.Provider {
	return &Provider{}
}

func (p *Provider***REMOVED*** Metadata(ctx context.Context, req tfprovider.MetadataRequest, resp *tfprovider.MetadataResponse***REMOVED*** {
	resp.TypeName = "rhcs"
	resp.Version = build.Version
}

// Provider creates the schema for the provider.
func (p *Provider***REMOVED*** Schema(ctx context.Context, req tfprovider.SchemaRequest, resp *tfprovider.SchemaResponse***REMOVED*** {
	resp.Schema = tfpschema.Schema{
		Attributes: map[string]tfpschema.Attribute{
			"url": tfpschema.StringAttribute{
				Description: "URL of the API server.",
				Optional:    true,
	***REMOVED***,
			"token_url": tfpschema.StringAttribute{
				Description: "OpenID token URL.",
				Optional:    true,
	***REMOVED***,
			"user": tfpschema.StringAttribute{
				Description: "User name.",
				Optional:    true,
	***REMOVED***,
			"password": tfpschema.StringAttribute{
				Description: "User password.",
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"token": tfpschema.StringAttribute{
				Description: "Access or refresh token that is " +
					"generated from https://console.redhat.com/openshift/token/rosa.",
				Optional:  true,
				Sensitive: true,
	***REMOVED***,
			"client_id": tfpschema.StringAttribute{
				Description: "OpenID client identifier.",
				Optional:    true,
	***REMOVED***,
			"client_secret": tfpschema.StringAttribute{
				Description: "OpenID client secret.",
				Optional:    true,
				Sensitive:   true,
	***REMOVED***,
			"trusted_cas": tfpschema.StringAttribute{
				Description: "PEM encoded certificates of authorities that will " +
					"be trusted. If this is not explicitly specified, then " +
					"the provider will trust the certificate authorities " +
					"trusted by default by the system.",
				Optional: true,
	***REMOVED***,
			"insecure": tfpschema.BoolAttribute{
				Description: "When set to 'true' enables insecure communication " +
					"with the server. This disables verification of TLS " +
					"certificates and host names, and it is not recommended " +
					"for production environments.",
				Optional: true,
	***REMOVED***,
***REMOVED***,
	}
}

// configure is the configuration function of the provider. It is responsible for checking the
// connection parameters and creating the connection that will be used by the resources.
func (p *Provider***REMOVED*** Configure(ctx context.Context, req tfprovider.ConfigureRequest,
	resp *tfprovider.ConfigureResponse***REMOVED*** {
	// Retrieve the provider configuration:
	var config Config
	diags := req.Config.Get(ctx, &config***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
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
			resp.Diagnostics.AddError(
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
		resp.Diagnostics.AddError(err.Error(***REMOVED***, ""***REMOVED***
		return
	}

	// Save the connection:
	resp.DataSourceData = connection
	resp.ResourceData = connection
}

// Resources returns the resources supported by the provider.
func (p *Provider***REMOVED*** Resources(ctx context.Context***REMOVED*** []func(***REMOVED*** resource.Resource {
	return []func(***REMOVED*** resource.Resource{
		clusterwaiter.New,
		dnsdomain.New,
		groupmembership.New,
		machinepool.New,
		oidcconfig.New,
		oidcconfiginput.New,
		clusterrosaclassic.New,
		identityprovider.New,
		cluster.New,
	}
}

func (p *Provider***REMOVED*** DataSources(ctx context.Context***REMOVED*** []func(***REMOVED*** datasource.DataSource {
	return []func(***REMOVED*** datasource.DataSource{
		cloudprovider.New,
		group.New,
		machine_types.New,
		ocm_policies.New,
		rosa_operator_roles.New,
		versions.New,
	}
}
