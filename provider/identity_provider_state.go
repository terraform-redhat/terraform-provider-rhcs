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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/idps"
***REMOVED***

type IdentityProviderState struct {
	Cluster  types.String                 `tfsdk:"cluster"`
	ID       types.String                 `tfsdk:"id"`
	Name     types.String                 `tfsdk:"name"`
	HTPasswd *HTPasswdIdentityProvider    `tfsdk:"htpasswd"`
	Gitlab   *GitlabIdentityProvider      `tfsdk:"gitlab"`
	Github   *idps.GithubIdentityProvider `tfsdk:"github"`
	LDAP     *LDAPIdentityProvider        `tfsdk:"ldap"`
	OpenID   *OpenIDIdentityProvider      `tfsdk:"openid"`
}

type HTPasswdIdentityProvider struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type GitlabIdentityProvider struct {
	CA           types.String `tfsdk:"ca"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	URL          types.String `tfsdk:"url"`
}

type LDAPIdentityProvider struct {
	BindDN       types.String                    `tfsdk:"bind_dn"`
	BindPassword types.String                    `tfsdk:"bind_password"`
	CA           types.String                    `tfsdk:"ca"`
	Insecure     types.Bool                      `tfsdk:"insecure"`
	URL          types.String                    `tfsdk:"url"`
	Attributes   *LDAPIdentityProviderAttributes `tfsdk:"attributes"`
}

type LDAPIdentityProviderAttributes struct {
	EMail             []string `tfsdk:"email"`
	ID                []string `tfsdk:"id"`
	Name              []string `tfsdk:"name"`
	PreferredUsername []string `tfsdk:"preferred_username"`
}

type OpenIDIdentityProvider struct {
	CA                       types.String                  `tfsdk:"ca"`
	Claims                   *OpenIDIdentityProviderClaims `tfsdk:"claims"`
	ClientID                 types.String                  `tfsdk:"client_id"`
	ClientSecret             types.String                  `tfsdk:"client_secret"`
	ExtraScopes              []string                      `tfsdk:"extra_scopes"`
	ExtraAuthorizeParameters map[string]string             `tfsdk:"extra_authorize_parameters"`
	Issuer                   types.String                  `tfsdk:"issuer"`
}

type OpenIDIdentityProviderClaims struct {
	EMail             []string `tfsdk:"email"`
	Groups            []string `tfsdk:"groups"`
	Name              []string `tfsdk:"name"`
	PreferredUsername []string `tfsdk:"preferred_username"`
}
