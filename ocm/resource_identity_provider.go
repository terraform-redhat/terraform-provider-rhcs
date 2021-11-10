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
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func resourceIdentityProvider() *schema.Resource {
	return &schema.Resource{
		Description: "Creates an identity provider.",
		Schema: map[string]*schema.Schema{
			clusterIDKey: {
				Description: "Identifier of the cluster.",
				Type:        schema.TypeString,
				Required:    true,
			},
			nameKey: {
				Description: "Name of the identity provider.",
				Type:        schema.TypeString,
				Required:    true,
			},
			htpasswdKey: {
				Description: "Details for the 'htpassw' identity provider.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ConflictsWith: []string{
					ldapKey,
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						userKey: {
							Description: "User name.",
							Type:        schema.TypeString,
							Required:    true,
						},
						passwordKey: {
							Description: "User password.",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			ldapKey: {
				Description: "Details for the LDAP identity provider.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ConflictsWith: []string{
					htpasswdKey,
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						attributesKey: {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									emailKey: {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									idKey: {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									nameKey: {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									preferredUsernameKey: {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						bindDNKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						bindPasswordKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						caKey: {
							Type:     schema.TypeString,
							Optional: true,
						},
						insecureKey: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						urlKey: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
		CreateContext: resourceIdentityProviderCreate,
		ReadContext:   resourceIdentityProviderRead,
		UpdateContext: resourceIdentityProviderUpdate,
		DeleteContext: resourceIdentityProviderDelete,
	}
}

func resourceIdentityProviderCreate(ctx context.Context, data *schema.ResourceData,
	config interface{}) (result diag.Diagnostics) {
	// Get the connection:
	connection := config.(*sdk.Connection)

	// Check if the identity provider already exists. If it does exist then we don't need to do
	// anything else.
	var ip *cmv1.IdentityProvider
	ip, result = resourceIdentityProviderLookup(ctx, connection, data)
	if result.HasError() {
		return
	}

	// Currently we need to wait till the cluster is ready because the server explicitly rejects
	// requests to create identity providers when the cluster isn't ready yet:
	clusterID := data.Get(clusterIDKey).(string)
	clusterResource := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID)
	clusterResource.Poll().
		Interval(1 * time.Minute).
		Predicate(func(getResponse *cmv1.ClusterGetResponse) bool {
			cluster := getResponse.Body()
			return cluster.State() == cmv1.ClusterStateReady
		}).
		StartContext(ctx)

	// If the identity provider doesn't exist yet then try to create it:
	if ip == nil {
		ip, result = resourceIdentityProviderRender(data)
		if result.HasError() {
			return
		}
		ipsResource := clusterResource.IdentityProviders()
		addResponse, err := ipsResource.Add().
			Body(ip).
			SendContext(ctx)
		if err != nil {
			result = diag.FromErr(err)
			return
		}
		ip = addResponse.Body()
	}

	// Copy the identity provider data:
	result = resourceIdentityProviderParse(ip, data)
	return
}

func resourceIdentityProviderRead(ctx context.Context, data *schema.ResourceData,
	config interface{}) (result diag.Diagnostics) {
	// Get the connection:
	connection := config.(*sdk.Connection)

	// Try to find the identity provider:
	var ip *cmv1.IdentityProvider
	ip, result = resourceIdentityProviderLookup(ctx, connection, data)
	if result.HasError() {
		return
	}

	// If there is no matching identity provider the mark it for creation:
	if ip == nil {
		data.SetId("")
		return
	}

	// Parse the identity provider data:
	result = resourceIdentityProviderParse(ip, data)
	if result.HasError() {
		return
	}

	return
}

func resourceIdentityProviderUpdate(ctx context.Context, data *schema.ResourceData,
	config interface{}) (result diag.Diagnostics) {
	return
}

func resourceIdentityProviderDelete(ctx context.Context, data *schema.ResourceData,
	config interface{}) (result diag.Diagnostics) {
	// Get the connection:
	connection := config.(*sdk.Connection)

	// Send the request to delete the identity provider:
	clusterID := data.Get(clusterIDKey).(string)
	ipID := data.Id()
	clusterResource := connection.ClustersMgmt().V1().Clusters().Cluster(clusterID)
	ipResource := clusterResource.IdentityProviders().IdentityProvider(ipID)
	deleteResponse, err := ipResource.Delete().SendContext(ctx)
	if deleteResponse != nil && deleteResponse.Status() == http.StatusNotFound {
		return
	}
	if err != nil {
		result = diag.FromErr(err)
		return
	}

	return
}

// resourceIdentityProviderRender converts the internal representation of an identity provider into
// corresponding SDK identity provider object.
func resourceIdentityProviderRender(data *schema.ResourceData) (ip *cmv1.IdentityProvider,
	result diag.Diagnostics) {
	// Prepare the builder and set the basic attributes:
	builder := cmv1.NewIdentityProvider()
	var value interface{}
	var ok bool
	value, ok = data.GetOk(nameKey)
	if ok {
		builder.Name(value.(string))
	}

	// Attributes for the `htpasswd` type:
	value, ok = data.GetOk(htpasswdKey)
	if ok {
		htpasswdList := value.([]interface{})
		htpasswdMap := htpasswdList[0].(map[string]interface{})
		htpasswdBuilder := cmv1.NewHTPasswdIdentityProvider()
		value, ok = htpasswdMap[userKey]
		if ok {
			htpasswdBuilder.Username(value.(string))
		}
		value, ok = htpasswdMap[passwordKey]
		if ok {
			htpasswdBuilder.Password(value.(string))
		}
		builder.Type(cmv1.IdentityProviderType("HTPasswdIdentityProvider"))
		builder.Htpasswd(htpasswdBuilder)
	}

	// Attributes for the `ldap` type:
	value, ok = data.GetOk(ldapKey)
	if ok {
		ldapList := value.([]interface{})
		ldapMap := ldapList[0].(map[string]interface{})
		ldapBuilder := cmv1.NewLDAPIdentityProvider()
		value, ok = ldapMap[attributesKey]
		if ok {
			attributesList := value.([]interface{})
			attributesMap := attributesList[0].(map[string]interface{})
			attributesBuilder := cmv1.NewLDAPAttributes()
			value, ok = attributesMap[emailKey]
			if ok {
				values := value.([]interface{})
				texts := make([]string, len(values))
				for i, value := range values {
					texts[i] = value.(string)
				}
				attributesBuilder.Email(texts...)
			}
			value, ok = attributesMap[idKey]
			if ok {
				values := value.([]interface{})
				texts := make([]string, len(values))
				for i, value := range values {
					texts[i] = value.(string)
				}
				attributesBuilder.ID(texts...)
			}
			value, ok = attributesMap[nameKey]
			if ok {
				values := value.([]interface{})
				texts := make([]string, len(values))
				for i, value := range values {
					texts[i] = value.(string)
				}
				attributesBuilder.Name(texts...)
			}
			value, ok = attributesMap[preferredUsernameKey]
			if ok {
				values := value.([]interface{})
				texts := make([]string, len(values))
				for i, value := range values {
					texts[i] = value.(string)
				}
				attributesBuilder.PreferredUsername(texts...)
			}
			ldapBuilder.Attributes(attributesBuilder)
		}
		value, ok = ldapMap[bindDNKey]
		if ok {
			ldapBuilder.BindDN(value.(string))
		}
		value, ok = ldapMap[bindPasswordKey]
		if ok {
			ldapBuilder.BindPassword(value.(string))
		}
		value, ok = ldapMap[caKey]
		if ok {
			ldapBuilder.CA(value.(string))
		}
		value, ok = ldapMap[insecureKey]
		if ok {
			ldapBuilder.Insecure(value.(bool))
		}
		value, ok = ldapMap[urlKey]
		if ok {
			ldapBuilder.URL(value.(string))
		}
		builder.Type(cmv1.IdentityProviderType("LDAPIdentityProvider"))
		builder.LDAP(ldapBuilder)
	}

	// Build the object:
	ip, err := builder.Build()
	if err != nil {
		result = diag.FromErr(err)
	}
	return
}

// resourceIdentityProviderParse converts a SDK identity provider into the internal representation.
func resourceIdentityProviderParse(ip *cmv1.IdentityProvider,
	data *schema.ResourceData) (result diag.Diagnostics) {
	// Basic attributes:
	data.SetId(ip.ID())
	data.Set(nameKey, ip.Name())

	// Attributres for the `htpasswd` type:
	htpasswd, ok := ip.GetHtpasswd()
	if ok {
		htpasswdData := map[string]interface{}{}
		user, ok := htpasswd.GetUsername()
		if ok {
			htpasswdData[userKey] = user
		}
		password, ok := htpasswd.GetPassword()
		if ok {
			htpasswdData[passwordKey] = password
		}
		data.Set(htpasswdKey, htpasswdData)
	}

	// Attributes for the LDAP type:
	ldap, ok := ip.GetLDAP()
	if ok {
		ldapData := map[string]interface{}{}
		attributes, ok := ldap.GetAttributes()
		if ok {
			attributesData := map[string]interface{}{}
			email, ok := attributes.GetEmail()
			if ok {
				attributesData[emailKey] = email
			}
			id, ok := attributes.GetID()
			if ok {
				attributesData[idKey] = id
			}
			name, ok := attributes.GetName()
			if ok {
				attributesData[nameKey] = name
			}
			preferredUsername, ok := attributes.GetPreferredUsername()
			if ok {
				attributesData[preferredUsernameKey] = preferredUsername
			}
			ldapData[attributesKey] = attributesData
		}
		bindDN, ok := ldap.GetBindDN()
		if ok {
			ldapData[bindDNKey] = bindDN
		}
		bindPassword, ok := ldap.GetBindPassword()
		if ok {
			ldapData[bindPasswordKey] = bindPassword
		}
		ca, ok := ldap.GetCA()
		if ok {
			ldapData[caKey] = ca
		}
		insecure, ok := ldap.GetInsecure()
		if ok {
			ldapData[insecureKey] = insecure
		}
		url, ok := ldap.GetURL()
		if ok {
			ldapData[urlKey] = url
		}
		data.Set(ldapKey, ldapData)
	}

	return
}

// resourceIdentityProviderLookup tries to find an identity provider that matches the given data.
// Returns nil if no such identity provider exists.
func resourceIdentityProviderLookup(ctx context.Context, connection *sdk.Connection,
	data *schema.ResourceData) (ip *cmv1.IdentityProvider, result diag.Diagnostics) {
	// Get the resource that manages the collection of identity providers of the cluster:
	clusterID := data.Get(clusterIDKey).(string)
	ipsResource := connection.ClustersMgmt().V1().
		Clusters().
		Cluster(clusterID).
		IdentityProviders()

	// If the we know the identifier of the identity provider then use it:
	ipID := data.Id()
	if ipID != "" {
		ipResource := ipsResource.IdentityProvider(ipID)
		getResponse, err := ipResource.Get().SendContext(ctx)
		if getResponse != nil && getResponse.Status() == http.StatusNotFound {
			return
		}
		if err != nil {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"can't find identity provider with identifier '%s' for "+
						"cluster '%s'",
					ipID, clusterID,
				),
				Detail: err.Error(),
			})
			return
		}
		ip = getResponse.Body()
		return
	}

	// If we are here then we don't know the identifier of the identity provider, but we may
	// know the name. If we do then we need to fetch all the existing identity providers and
	// search locally because the identity providers collection doesn't support search.
	value, ok := data.GetOk(nameKey)
	if ok {
		ipName := value.(string)
		listResponse, err := ipsResource.List().SendContext(ctx)
		if err != nil {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"can't find identity provider with name '%s' for "+
						"cluster '%s'",
					ipName, ipID,
				),
				Detail: err.Error(),
			})
			return
		}
		listResponse.Items().Each(func(item *cmv1.IdentityProvider) bool {
			if item.Name() == ipName {
				ip = item
				return false
			}
			return true
		})
		return
	}

	return
}
