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
	"regexp"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfpschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hfrest "github.com/openshift-online/rosa-hyperfleet-api/clientset/rest"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"
	"github.com/terraform-redhat/terraform-provider-rhcs/logging"
	classicAutoscaler "github.com/terraform-redhat/terraform-provider-rhcs/provider/autoscaler/classic"
	hcpAutoscaler "github.com/terraform-redhat/terraform-provider-rhcs/provider/autoscaler/hcp"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/breakglasscredential"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/cloudprovider"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/cluster"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/classic"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterrosa/hcp"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/clusterwaiter"
	defaultingress "github.com/terraform-redhat/terraform-provider-rhcs/provider/defaultingress/classic"
	hcpingress "github.com/terraform-redhat/terraform-provider-rhcs/provider/defaultingress/hcp"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/dnsdomain"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/group"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/groupmembership"
	hyperfleethcp "github.com/terraform-redhat/terraform-provider-rhcs/provider/hyperfleet"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/identityprovider"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/imagemirror"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/info"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/kubeletconfig"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/logforwarder"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/machine_types"
	machinepool "github.com/terraform-redhat/terraform-provider-rhcs/provider/machinepool/classic"
	nodepool "github.com/terraform-redhat/terraform-provider-rhcs/provider/machinepool/hcp"
	classicStsPolicies "github.com/terraform-redhat/terraform-provider-rhcs/provider/ocm_policies/classic"
	hcpStsPolicies "github.com/terraform-redhat/terraform-provider-rhcs/provider/ocm_policies/hcp"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/ocmrole"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/oidcconfig"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/oidcconfiginput"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/providerdata"
	classicOperatorRoles "github.com/terraform-redhat/terraform-provider-rhcs/provider/rosa_operator_roles/classic"
	hcpOperatorRoles "github.com/terraform-redhat/terraform-provider-rhcs/provider/rosa_operator_roles/hcp"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/trusted_ip_addresses"
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
	RefreshToken types.String `tfsdk:"refresh_token"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	TrustedCAs   types.String `tfsdk:"trusted_cas"`
	Insecure     types.Bool   `tfsdk:"insecure"`

	// Hyperfleet (Platform API v2) configuration. All fields are optional; when
	// hyperfleet_url is absent the provider operates in OCM-only mode.
	HyperfleetURL types.String `tfsdk:"hyperfleet_url"`
	AWSAccountID  types.String `tfsdk:"aws_account_id"`
	AWSCallerARN  types.String `tfsdk:"aws_caller_arn"`
	AWSRegion     types.String `tfsdk:"aws_region"`
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
			"refresh_token": tfpschema.StringAttribute{
				Description: "Refresh token that is generated from `rosa login`.",
				Optional:    true,
				Sensitive:   true,
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
			"hyperfleet_url": tfpschema.StringAttribute{
				Description: "Base URL of the Hyperfleet Platform API v2 endpoint " +
					"(e.g. `https://abc123.execute-api.us-east-1.amazonaws.com`). " +
					"When set the provider enables hyperfleet resources. " +
					"Can also be set via the `RHCS_HYPERFLEET_URL` environment variable.",
				Optional: true,
			},
			"aws_account_id": tfpschema.StringAttribute{
				Description: "AWS account ID used to namespace Hyperfleet API calls. " +
					"Required when `hyperfleet_url` is set. " +
					"Typically sourced from `data.aws_caller_identity.current.account_id`. " +
					"Can also be set via the `RHCS_AWS_ACCOUNT_ID` environment variable.",
				Optional: true,
			},
			"aws_caller_arn": tfpschema.StringAttribute{
				Description: "ARN of the AWS caller identity forwarded to the Hyperfleet API " +
					"as an informational header. Optional when `hyperfleet_url` is set. " +
					"Typically sourced from `data.aws_caller_identity.current.arn`. " +
					"Can also be set via the `RHCS_AWS_CALLER_ARN` environment variable.",
				Optional: true,
			},
			"aws_region": tfpschema.StringAttribute{
				Description: "AWS region used for SigV4 request signing against the " +
					"Hyperfleet Platform API. Optional: when absent the region is derived " +
					"from the `hyperfleet_url` hostname. " +
					"Can also be set via the `RHCS_AWS_REGION` environment variable.",
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
	if refreshToken, ok := p.getAttrValueOrConfig(config.RefreshToken, "REFRESH_TOKEN"); ok {
		builder.Tokens(refreshToken)
	}
	clientID, clientIdExists := p.getAttrValueOrConfig(config.ClientID, "CLIENT_ID")
	clientSecret, _ := p.getAttrValueOrConfig(config.ClientSecret, "CLIENT_SECRET")
	if clientIdExists {
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

	shared := &providerdata.ProviderSharedData{}

	// Build the OCM connection. When hyperfleet_url is set without OCM credentials
	// the connection build may fail — that is acceptable for hyperfleet-only usage.
	connection, err := builder.BuildContext(ctx)
	if err != nil {
		if config.HyperfleetURL.IsNull() || config.HyperfleetURL.ValueString() == "" {
			// No hyperfleet fallback — OCM is required.
			resp.Diagnostics.AddError(err.Error(), "")
			return
		}
		// Hyperfleet-only mode: OCM credentials absent but hyperfleet_url is set.
	} else {
		shared.OCMConnection = connection
	}

	// Build the hyperfleet clientset when hyperfleet_url is configured.
	if hfURL, ok := p.getAttrValueOrConfig(config.HyperfleetURL, "HYPERFLEET_URL"); ok {
		accountID, hasAccountID := p.getAttrValueOrConfig(config.AWSAccountID, "AWS_ACCOUNT_ID")
		if !hasAccountID {
			resp.Diagnostics.AddError(
				"Missing required provider attribute",
				"'aws_account_id' is required when 'hyperfleet_url' is set. "+
					"Set it explicitly or via the RHCS_AWS_ACCOUNT_ID environment variable.",
			)
			return
		}

		awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to load AWS configuration", err.Error())
			return
		}

		callerARN, _ := p.getAttrValueOrConfig(config.AWSCallerARN, "AWS_CALLER_ARN")
		region, _ := p.getAttrValueOrConfig(config.AWSRegion, "AWS_REGION")

		// Derive the platform region from the hyperfleet URL (execute-api hostname).
		platformRegion := regionFromHyperfleetURL(hfURL)
		if platformRegion != "" && region != "" && region != platformRegion {
			resp.Diagnostics.AddError(
				"AWS region mismatch",
				fmt.Sprintf(
					"'aws_region' (%q) does not match the region embedded in 'hyperfleet_url' (%q). "+
						"Set 'aws_region' to %q or update 'hyperfleet_url' to point to the %q endpoint.",
					region, hfURL, platformRegion, region,
				),
			)
			return
		}
		if region == "" {
			region = platformRegion
		}

		hfClient, err := hyperfleet.NewForConfig(&hfrest.Config{
			Host:      hfURL,
			Region:    region,
			AccountID: accountID,
			CallerARN: callerARN,
			AWSConfig: awsCfg,
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to create Hyperfleet client", err.Error())
			return
		}

		shared.HyperfleetClient = hfClient
		shared.HyperfleetAccountID = accountID
		shared.HyperfleetCallerARN = callerARN
	}

	resp.DataSourceData = shared
	resp.ResourceData = shared
}

// Resources returns the resources supported by the provider.
func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		clusterwaiter.New,
		dnsdomain.New,
		groupmembership.New,
		imagemirror.New,
		machinepool.New,
		oidcconfig.New,
		oidcconfiginput.New,
		classic.New,
		identityprovider.New,
		cluster.New,
		classicAutoscaler.New,
		defaultingress.New,
		kubeletconfig.New,
		hcp.New,
		hyperfleethcp.New,
		hyperfleethcp.NewNodePool,
		nodepool.New,
		hcpingress.New,
		tuningconfigs.New,
		hcpAutoscaler.New,
		breakglasscredential.New,
		logforwarder.New,
		ocmrole.New,
	}
}

func (p *Provider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cloudprovider.New,
		group.New,
		machine_types.New,
		classicStsPolicies.New,
		classicOperatorRoles.New,
		versions.New,
		info.New,
		classic.NewDataSource,
		machinepool.NewDatasource,
		hcp.NewDataSource,
		nodepool.NewDatasource,
		hcpOperatorRoles.New,
		hcpStsPolicies.New,
		trusted_ip_addresses.New,
		imagemirror.NewDataSource,
		logforwarder.NewDataSource,
	}
}

// awsRegionRE matches standard and GovCloud AWS region names embedded in a URL
// hostname (e.g. us-east-1, ap-southeast-1, us-gov-east-1).
var awsRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

// regionFromHyperfleetURL extracts the AWS region from a Platform API execute-api
// URL (e.g. https://<id>.execute-api.<region>.amazonaws.com/…).
// Returns "" when no recognisable region segment is found.
func regionFromHyperfleetURL(rawURL string) string {
	return awsRegionRE.FindString(rawURL)
}
