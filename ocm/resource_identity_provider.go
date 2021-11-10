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

package ocm

***REMOVED***
	"context"
***REMOVED***
***REMOVED***
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

func resourceIdentityProvider(***REMOVED*** *schema.Resource {
	return &schema.Resource{
		Description: "Creates an identity provider.",
		Schema: map[string]*schema.Schema{
			clusterIDKey: {
				Description: "Identifier of the cluster.",
				Type:        schema.TypeString,
				Required:    true,
	***REMOVED***,
			nameKey: {
				Description: "Name of the identity provider.",
				Type:        schema.TypeString,
				Required:    true,
	***REMOVED***,
			htpasswdKey: {
				Description: "Details for the 'htpassw' identity provider.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ConflictsWith: []string{
					ldapKey,
		***REMOVED***,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						userKey: {
							Description: "User name.",
							Type:        schema.TypeString,
							Required:    true,
				***REMOVED***,
						passwordKey: {
							Description: "User password.",
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			ldapKey: {
				Description: "Details for the LDAP identity provider.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ConflictsWith: []string{
					htpasswdKey,
		***REMOVED***,
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
								***REMOVED***,
							***REMOVED***,
									idKey: {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
								***REMOVED***,
							***REMOVED***,
									nameKey: {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
								***REMOVED***,
							***REMOVED***,
									preferredUsernameKey: {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
								***REMOVED***,
							***REMOVED***,
						***REMOVED***,
					***REMOVED***,
				***REMOVED***,
						bindDNKey: {
							Type:     schema.TypeString,
							Required: true,
				***REMOVED***,
						bindPasswordKey: {
							Type:     schema.TypeString,
							Required: true,
				***REMOVED***,
						caKey: {
							Type:     schema.TypeString,
							Optional: true,
				***REMOVED***,
						insecureKey: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
				***REMOVED***,
						urlKey: {
							Type:     schema.TypeString,
							Required: true,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
		CreateContext: resourceIdentityProviderCreate,
		ReadContext:   resourceIdentityProviderRead,
		UpdateContext: resourceIdentityProviderUpdate,
		DeleteContext: resourceIdentityProviderDelete,
	}
}

func resourceIdentityProviderCreate(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	// Get the connection:
	connection := config.(*sdk.Connection***REMOVED***

	// Check if the identity provider already exists. If it does exist then we don't need to do
	// anything else.
	var ip *cmv1.IdentityProvider
	ip, result = resourceIdentityProviderLookup(ctx, connection, data***REMOVED***
	if result.HasError(***REMOVED*** {
		return
	}

	// Currently we need to wait till the cluster is ready because the server explicitly rejects
	// requests to create identity providers when the cluster isn't ready yet:
	clusterID := data.Get(clusterIDKey***REMOVED***.(string***REMOVED***
	clusterResource := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***
	clusterResource.Poll(***REMOVED***.
		Interval(1 * time.Minute***REMOVED***.
		Predicate(func(getResponse *cmv1.ClusterGetResponse***REMOVED*** bool {
			cluster := getResponse.Body(***REMOVED***
			return cluster.State(***REMOVED*** == cmv1.ClusterStateReady
***REMOVED******REMOVED***.
		StartContext(ctx***REMOVED***

	// If the identity provider doesn't exist yet then try to create it:
	if ip == nil {
		ip, result = resourceIdentityProviderRender(data***REMOVED***
		if result.HasError(***REMOVED*** {
			return
***REMOVED***
		ipsResource := clusterResource.IdentityProviders(***REMOVED***
		addResponse, err := ipsResource.Add(***REMOVED***.
			Body(ip***REMOVED***.
			SendContext(ctx***REMOVED***
		if err != nil {
			result = diag.FromErr(err***REMOVED***
			return
***REMOVED***
		ip = addResponse.Body(***REMOVED***
	}

	// Copy the identity provider data:
	result = resourceIdentityProviderParse(ip, data***REMOVED***
	return
}

func resourceIdentityProviderRead(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	// Get the connection:
	connection := config.(*sdk.Connection***REMOVED***

	// Try to find the identity provider:
	var ip *cmv1.IdentityProvider
	ip, result = resourceIdentityProviderLookup(ctx, connection, data***REMOVED***
	if result.HasError(***REMOVED*** {
		return
	}

	// If there is no matching identity provider the mark it for creation:
	if ip == nil {
		data.SetId(""***REMOVED***
		return
	}

	// Parse the identity provider data:
	result = resourceIdentityProviderParse(ip, data***REMOVED***
	if result.HasError(***REMOVED*** {
		return
	}

	return
}

func resourceIdentityProviderUpdate(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	return
}

func resourceIdentityProviderDelete(ctx context.Context, data *schema.ResourceData,
	config interface{}***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	// Get the connection:
	connection := config.(*sdk.Connection***REMOVED***

	// Send the request to delete the identity provider:
	clusterID := data.Get(clusterIDKey***REMOVED***.(string***REMOVED***
	ipID := data.Id(***REMOVED***
	clusterResource := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***.Cluster(clusterID***REMOVED***
	ipResource := clusterResource.IdentityProviders(***REMOVED***.IdentityProvider(ipID***REMOVED***
	deleteResponse, err := ipResource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if deleteResponse != nil && deleteResponse.Status(***REMOVED*** == http.StatusNotFound {
		return
	}
	if err != nil {
		result = diag.FromErr(err***REMOVED***
		return
	}

	return
}

// resourceIdentityProviderRender converts the internal representation of an identity provider into
// corresponding SDK identity provider object.
func resourceIdentityProviderRender(data *schema.ResourceData***REMOVED*** (ip *cmv1.IdentityProvider,
	result diag.Diagnostics***REMOVED*** {
	// Prepare the builder and set the basic attributes:
	builder := cmv1.NewIdentityProvider(***REMOVED***
	var value interface{}
	var ok bool
	value, ok = data.GetOk(nameKey***REMOVED***
	if ok {
		builder.Name(value.(string***REMOVED******REMOVED***
	}

	// Attributes for the `htpasswd` type:
	value, ok = data.GetOk(htpasswdKey***REMOVED***
	if ok {
		htpasswdList := value.([]interface{}***REMOVED***
		htpasswdMap := htpasswdList[0].(map[string]interface{}***REMOVED***
		htpasswdBuilder := cmv1.NewHTPasswdIdentityProvider(***REMOVED***
		value, ok = htpasswdMap[userKey]
		if ok {
			htpasswdBuilder.Username(value.(string***REMOVED******REMOVED***
***REMOVED***
		value, ok = htpasswdMap[passwordKey]
		if ok {
			htpasswdBuilder.Password(value.(string***REMOVED******REMOVED***
***REMOVED***
		builder.Type(cmv1.IdentityProviderType("HTPasswdIdentityProvider"***REMOVED******REMOVED***
		builder.Htpasswd(htpasswdBuilder***REMOVED***
	}

	// Attributes for the `ldap` type:
	value, ok = data.GetOk(ldapKey***REMOVED***
	if ok {
		ldapList := value.([]interface{}***REMOVED***
		ldapMap := ldapList[0].(map[string]interface{}***REMOVED***
		ldapBuilder := cmv1.NewLDAPIdentityProvider(***REMOVED***
		value, ok = ldapMap[attributesKey]
		if ok {
			attributesList := value.([]interface{}***REMOVED***
			attributesMap := attributesList[0].(map[string]interface{}***REMOVED***
			attributesBuilder := cmv1.NewLDAPAttributes(***REMOVED***
			value, ok = attributesMap[emailKey]
			if ok {
				values := value.([]interface{}***REMOVED***
				texts := make([]string, len(values***REMOVED******REMOVED***
				for i, value := range values {
					texts[i] = value.(string***REMOVED***
		***REMOVED***
				attributesBuilder.Email(texts...***REMOVED***
	***REMOVED***
			value, ok = attributesMap[idKey]
			if ok {
				values := value.([]interface{}***REMOVED***
				texts := make([]string, len(values***REMOVED******REMOVED***
				for i, value := range values {
					texts[i] = value.(string***REMOVED***
		***REMOVED***
				attributesBuilder.ID(texts...***REMOVED***
	***REMOVED***
			value, ok = attributesMap[nameKey]
			if ok {
				values := value.([]interface{}***REMOVED***
				texts := make([]string, len(values***REMOVED******REMOVED***
				for i, value := range values {
					texts[i] = value.(string***REMOVED***
		***REMOVED***
				attributesBuilder.Name(texts...***REMOVED***
	***REMOVED***
			value, ok = attributesMap[preferredUsernameKey]
			if ok {
				values := value.([]interface{}***REMOVED***
				texts := make([]string, len(values***REMOVED******REMOVED***
				for i, value := range values {
					texts[i] = value.(string***REMOVED***
		***REMOVED***
				attributesBuilder.PreferredUsername(texts...***REMOVED***
	***REMOVED***
			ldapBuilder.Attributes(attributesBuilder***REMOVED***
***REMOVED***
		value, ok = ldapMap[bindDNKey]
		if ok {
			ldapBuilder.BindDN(value.(string***REMOVED******REMOVED***
***REMOVED***
		value, ok = ldapMap[bindPasswordKey]
		if ok {
			ldapBuilder.BindPassword(value.(string***REMOVED******REMOVED***
***REMOVED***
		value, ok = ldapMap[caKey]
		if ok {
			ldapBuilder.CA(value.(string***REMOVED******REMOVED***
***REMOVED***
		value, ok = ldapMap[insecureKey]
		if ok {
			ldapBuilder.Insecure(value.(bool***REMOVED******REMOVED***
***REMOVED***
		value, ok = ldapMap[urlKey]
		if ok {
			ldapBuilder.URL(value.(string***REMOVED******REMOVED***
***REMOVED***
		builder.Type(cmv1.IdentityProviderType("LDAPIdentityProvider"***REMOVED******REMOVED***
		builder.LDAP(ldapBuilder***REMOVED***
	}

	// Build the object:
	ip, err := builder.Build(***REMOVED***
	if err != nil {
		result = diag.FromErr(err***REMOVED***
	}
	return
}

// resourceIdentityProviderParse converts a SDK identity provider into the internal representation.
func resourceIdentityProviderParse(ip *cmv1.IdentityProvider,
	data *schema.ResourceData***REMOVED*** (result diag.Diagnostics***REMOVED*** {
	// Basic attributes:
	data.SetId(ip.ID(***REMOVED******REMOVED***
	data.Set(nameKey, ip.Name(***REMOVED******REMOVED***

	// Attributres for the `htpasswd` type:
	htpasswd, ok := ip.GetHtpasswd(***REMOVED***
	if ok {
		htpasswdData := map[string]interface{}{}
		user, ok := htpasswd.GetUsername(***REMOVED***
		if ok {
			htpasswdData[userKey] = user
***REMOVED***
		password, ok := htpasswd.GetPassword(***REMOVED***
		if ok {
			htpasswdData[passwordKey] = password
***REMOVED***
		data.Set(htpasswdKey, htpasswdData***REMOVED***
	}

	// Attributes for the LDAP type:
	ldap, ok := ip.GetLDAP(***REMOVED***
	if ok {
		ldapData := map[string]interface{}{}
		attributes, ok := ldap.GetAttributes(***REMOVED***
		if ok {
			attributesData := map[string]interface{}{}
			email, ok := attributes.GetEmail(***REMOVED***
			if ok {
				attributesData[emailKey] = email
	***REMOVED***
			id, ok := attributes.GetID(***REMOVED***
			if ok {
				attributesData[idKey] = id
	***REMOVED***
			name, ok := attributes.GetName(***REMOVED***
			if ok {
				attributesData[nameKey] = name
	***REMOVED***
			preferredUsername, ok := attributes.GetPreferredUsername(***REMOVED***
			if ok {
				attributesData[preferredUsernameKey] = preferredUsername
	***REMOVED***
			ldapData[attributesKey] = attributesData
***REMOVED***
		bindDN, ok := ldap.GetBindDN(***REMOVED***
		if ok {
			ldapData[bindDNKey] = bindDN
***REMOVED***
		bindPassword, ok := ldap.GetBindPassword(***REMOVED***
		if ok {
			ldapData[bindPasswordKey] = bindPassword
***REMOVED***
		ca, ok := ldap.GetCA(***REMOVED***
		if ok {
			ldapData[caKey] = ca
***REMOVED***
		insecure, ok := ldap.GetInsecure(***REMOVED***
		if ok {
			ldapData[insecureKey] = insecure
***REMOVED***
		url, ok := ldap.GetURL(***REMOVED***
		if ok {
			ldapData[urlKey] = url
***REMOVED***
		data.Set(ldapKey, ldapData***REMOVED***
	}

	return
}

// resourceIdentityProviderLookup tries to find an identity provider that matches the given data.
// Returns nil if no such identity provider exists.
func resourceIdentityProviderLookup(ctx context.Context, connection *sdk.Connection,
	data *schema.ResourceData***REMOVED*** (ip *cmv1.IdentityProvider, result diag.Diagnostics***REMOVED*** {
	// Get the resource that manages the collection of identity providers of the cluster:
	clusterID := data.Get(clusterIDKey***REMOVED***.(string***REMOVED***
	ipsResource := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.
		Clusters(***REMOVED***.
		Cluster(clusterID***REMOVED***.
		IdentityProviders(***REMOVED***

	// If the we know the identifier of the identity provider then use it:
	ipID := data.Id(***REMOVED***
	if ipID != "" {
		ipResource := ipsResource.IdentityProvider(ipID***REMOVED***
		getResponse, err := ipResource.Get(***REMOVED***.SendContext(ctx***REMOVED***
		if getResponse != nil && getResponse.Status(***REMOVED*** == http.StatusNotFound {
			return
***REMOVED***
		if err != nil {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"can't find identity provider with identifier '%s' for "+
						"cluster '%s'",
					ipID, clusterID,
				***REMOVED***,
				Detail: err.Error(***REMOVED***,
	***REMOVED******REMOVED***
			return
***REMOVED***
		ip = getResponse.Body(***REMOVED***
		return
	}

	// If we are here then we don't know the identifier of the identity provider, but we may
	// know the name. If we do then we need to fetch all the existing identity providers and
	// search locally because the identity providers collection doesn't support search.
	value, ok := data.GetOk(nameKey***REMOVED***
	if ok {
		ipName := value.(string***REMOVED***
		listResponse, err := ipsResource.List(***REMOVED***.SendContext(ctx***REMOVED***
		if err != nil {
			result = append(result, diag.Diagnostic{
				Severity: diag.Error,
				Summary: fmt.Sprintf(
					"can't find identity provider with name '%s' for "+
						"cluster '%s'",
					ipName, ipID,
				***REMOVED***,
				Detail: err.Error(***REMOVED***,
	***REMOVED******REMOVED***
			return
***REMOVED***
		listResponse.Items(***REMOVED***.Each(func(item *cmv1.IdentityProvider***REMOVED*** bool {
			if item.Name(***REMOVED*** == ipName {
				ip = item
				return false
	***REMOVED***
			return true
***REMOVED******REMOVED***
		return
	}

	return
}
