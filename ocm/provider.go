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

package ocm

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	"github.com/openshift-online/ocm-sdk-go/logging"
)

// Provider creates the schema for the provider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			urlKey: {
				Description: "URL of the API server.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     sdk.DefaultURL,
			},
			tokenURLKey: {
				Description: "OpenID token URL.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     sdk.DefaultTokenURL,
			},
			userKey: {
				Description: "User name.",
				Type:        schema.TypeString,
				Optional:    true,
				ConflictsWith: []string{
					clientIDKey,
					clientSecretKey,
					tokenKey,
				},
			},
			passwordKey: {
				Description: "User password.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ConflictsWith: []string{
					clientIDKey,
					clientSecretKey,
					tokenKey,
				},
			},
			tokenKey: {
				Description: "Access or refresh token.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("OCM_TOKEN", nil),
				ConflictsWith: []string{
					clientIDKey,
					clientSecretKey,
					passwordKey,
					userKey,
				},
			},
			clientIDKey: {
				Description: "OpenID client identifier.",
				Type:        schema.TypeString,
				Optional:    true,
				ConflictsWith: []string{
					passwordKey,
					tokenKey,
					userKey,
				},
			},
			clientSecretKey: {
				Description: "OpenID client secret.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				ConflictsWith: []string{
					passwordKey,
					tokenKey,
					userKey,
				},
			},
			trustedCAsKey: {
				Description: "PEM encoded certificates of authorities that will " +
					"be trusted. If this isn't explicitly specified then " +
					"the provider will trust the certificate authorities " +
					"trusted by default by the system.",
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			insecureKey: {
				Description: "When set to 'true' enables insecure communication " +
					"with the server. This disables verification of TLS " +
					"certificates and host names and it isn't recommended " +
					"for production environments.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"ocm_cluster": resourceCluster(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ocm_cloud_providers": dataSourceCloudProviders(),
		},
		ConfigureContextFunc: configure,
	}
}

// configure is the configuration function of the provider. It is responsible for checking the
// connection parameters and creating the connection that will be used by the resources.
func configure(ctx context.Context, data *schema.ResourceData) (config interface{},
	result diag.Diagnostics) {
	// Determine the log level used by the SDK from the environment variables used by Terraform:
	logLevel := os.Getenv("TF_LOG_PROVIDER")
	if logLevel == "" {
		logLevel = os.Getenv("TF_LOG")
	}
	if logLevel == "" {
		logLevel = logLevelInfo
	}

	// The plugin infrastructure redirects the log package output so that it is sent to the main
	// Terraform process, so if we want to have the logs of the SDK redirected we need to use
	// the log package as well.
	logger, err := logging.NewGoLoggerBuilder().
		Debug(logLevel == logLevelDebug).
		Info(logLevel == logLevelInfo).
		Warn(logLevel == logLevelWarn).
		Error(logLevel == logLevelError).
		Build()
	if err != nil {
		result = diag.FromErr(err)
		return
	}

	// Create the builder:
	builder := sdk.NewConnectionBuilder()
	builder.Logger(logger)

	// Copy the settings:
	urlValue, ok := data.GetOk(urlKey)
	if ok {
		builder.URL(urlValue.(string))
	}
	tokenURLValue, ok := data.GetOk(tokenURLKey)
	if ok {
		builder.TokenURL(tokenURLValue.(string))
	}
	userValue, userOk := data.GetOk(userKey)
	passwordValue, passwordOk := data.GetOk(passwordKey)
	if userOk || passwordOk {
		builder.User(userValue.(string), passwordValue.(string))
	}
	tokenValue, ok := data.GetOk(tokenKey)
	if ok {
		builder.Tokens(tokenValue.(string))
	}
	clientIDValue, clientIDOk := data.GetOk(clientIDKey)
	clientSecretValue, clientSecretOk := data.GetOk(clientSecretKey)
	if clientIDOk || clientSecretOk {
		builder.Client(clientIDValue.(string), clientSecretValue.(string))
	}
	insecureValue, ok := data.GetOk(insecureKey)
	if ok {
		builder.Insecure(insecureValue.(bool))
	}
	trustedCAs, ok := data.GetOk(trustedCAsKey)
	if ok {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(trustedCAs.(string))) {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Detail: fmt.Sprintf(
					"the value of '%s' doesn't contain any certificate",
					trustedCAsKey,
				),
			})
		}
		builder.TrustedCAs(pool)
	}

	// Create the connection:
	connection, err := builder.BuildContext(ctx)
	if err != nil {
		result = diag.FromErr(err)
		return
	}
	config = connection

	return
}

// Log levels:
const (
	logLevelDebug = "DEBUG"
	logLevelInfo  = "INFO"
	logLevelWarn  = "WARN"
	logLevelError = "ERROR"
)
