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

package idps

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validMappingMethods = []string{"claim", "add", "generate", "lookup"} // Default is @ index 0
var DefaultMappingMethod = validMappingMethods[0]

func IdentityProviderFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster": {
			Description: "Identifier of the cluster.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"name": {
			Description: "Name of the identity clusterservice.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"mapping_method": {
			Description:      "Specifies how new identities are mapped to users when they log in. Options are [add claim generate lookup] (default 'claim')",
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(validMappingMethods, false)),
		},
		"htpasswd": {
			Description: "Details of the 'htpasswd' identity clusterservice.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: HtpasswdSchema(),
			},
			Optional: true,
		},
		"gitlab": {
			Description: "Details of the Gitlab identity clusterservice.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: GitlabSchema(),
			},
			Optional: true,
		},
		"github": {
			Description: "Details of the Github identity clusterservice.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: GithubSchema(),
			},
			Optional: true,
		},
		"google": {
			Description: "Details of the Google identity clusterservice.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: GoogleSchema(),
			},
			Optional: true,
		},
		"ldap": {
			Description: "Details of the LDAP identity clusterservice.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: LDAPSchema(),
			},
			Optional: true,
		},
		"openid": {
			Description: "Details of the OpenID identity clusterservice.",
			Type:        schema.TypeList,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: OpenidSchema(),
			},
			Optional: true,
		},
	}
}

type IdentityProviderState struct {
	// required
	Cluster string `tfsdk:"cluster"`
	Name    string `tfsdk:"name"`

	//optional
	MappingMethod *string                   `tfsdk:"mapping_method"`
	HTPasswd      *HTPasswdIdentityProvider `tfsdk:"htpasswd"`
	Gitlab        *GitlabIdentityProvider   `tfsdk:"gitlab"`
	Github        *GithubIdentityProvider   `tfsdk:"github"`
	Google        *GoogleIdentityProvider   `tfsdk:"google"`
	LDAP          *LDAPIdentityProvider     `tfsdk:"ldap"`
	OpenID        *OpenIDIdentityProvider   `tfsdk:"openid"`

	// computed
	ID string `tfsdk:"id"`
}
