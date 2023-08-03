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
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	common2 "github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"net/http"
	"strings"
)

func ResourceIdentityProvider() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIdentityProviderCreate,
		ReadContext:   resourceIdentityProviderRead,
		UpdateContext: resourceIdentityProviderUpdate,
		DeleteContext: resourceIdentityProviderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceIDPImport,
		},
		Schema: IdentityProviderFields(),
	}
}

func resourceIdentityProviderCreate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin Creating")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	err := validateAttributes(resourceData)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Failed in validation",
				Detail:   err.Error(),
			}}
	}

	identityProviderState := identityProviderFromResourceData(resourceData)
	// Wait till the cluster is ready:
	if err := common2.WaitTillClusterIsReadyOrFail(ctx, clusterCollection, identityProviderState.Cluster); err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't poll cluster state",
				Detail: fmt.Sprintf(
					"Can't poll state of cluster with identifier '%s': %v",
					identityProviderState.Cluster, err),
			}}
	}

	// Create the identity clusterservice:
	builder := cmv1.NewIdentityProvider()
	builder.Name(identityProviderState.Name)
	// handle mapping_method
	mappingMethod := DefaultMappingMethod
	if identityProviderState.MappingMethod != nil && *identityProviderState.MappingMethod != "" {
		mappingMethod = *identityProviderState.MappingMethod
	}
	builder.MappingMethod(cmv1.IdentityProviderMappingMethod(mappingMethod))
	switch {
	case identityProviderState.HTPasswd != nil:
		builder.Type(cmv1.IdentityProviderTypeHtpasswd)
		htpasswdBuilder := CreateHTPasswdIDPBuilder(identityProviderState.HTPasswd)
		builder.Htpasswd(htpasswdBuilder)
	case identityProviderState.Gitlab != nil:
		builder.Type(cmv1.IdentityProviderTypeGitlab)
		gitlabBuilder := CreateGitlabIDPBuilder(identityProviderState.Gitlab)
		builder.Gitlab(gitlabBuilder)
	case identityProviderState.Github != nil:
		builder.Type(cmv1.IdentityProviderTypeGithub)
		githubBuilder := CreateGithubIDPBuilder(identityProviderState.Github)
		builder.Github(githubBuilder)
	case identityProviderState.Google != nil:
		builder.Type(cmv1.IdentityProviderTypeGoogle)
		googleBuilder, err := CreateGoogleIDPBuilder(mappingMethod, identityProviderState.Google)
		if err != nil {
			return []diag.Diagnostic{
				{
					Severity: diag.Error,
					Summary:  "Can't create identity provider",
					Detail:   err.Error(),
				}}
		}
		builder.Google(googleBuilder)
	case identityProviderState.LDAP != nil:
		builder.Type(cmv1.IdentityProviderTypeLDAP)
		ldapBuilder := CreateLDAPIDPBuilder(identityProviderState.LDAP)
		builder.LDAP(ldapBuilder)
	case identityProviderState.OpenID != nil:
		builder.Type(cmv1.IdentityProviderTypeOpenID)
		openidBuilder := CreateOpenIDIDPBuilder(identityProviderState.OpenID)
		builder.OpenID(openidBuilder)
	}

	object, err := builder.Build()
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't build identity clusterservice",
				Detail: fmt.Sprintf(
					"Can't build identity clusterservice with name '%s': %v",
					identityProviderState.Name, err),
			}}
	}
	collection := clusterCollection.Cluster(identityProviderState.Cluster).IdentityProviders()
	add, err := collection.Add().Body(object).SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't create identity clusterservice",
				Detail: fmt.Sprintf(
					"Can't create identity clusterservice with name '%s' for "+
						"cluster '%s': %v",
					identityProviderState.Name, identityProviderState.Cluster, err,
				),
			}}
	}

	object = add.Body()
	identityProviderToResourceData(object, resourceData)
	return
}

func identityProviderToResourceData(object *cmv1.IdentityProvider,
	resourceData *schema.ResourceData) {
	// Set the computed attributes:
	resourceData.SetId(object.ID())
	resourceData.Set("name", object.Name())

	resourceData.Set("mapping_method", string(object.MappingMethod()))
	resourceData.Set("htpasswd", FlatHtpasswd(object))
	resourceData.Set("gitlab", FlatGitlab(object))
	resourceData.Set("github", FlatGithub(object))
	resourceData.Set("google", FlatGoogle(object))
	resourceData.Set("ldap", FlatLDAP(object))
	resourceData.Set("openid", FlatOpenID(object))
}

func validateAttributes(resourceData *schema.ResourceData) error {
	if github, ok := resourceData.GetOk("github"); ok {
		err := common2.ValidAllDiag(common2.ListOfMapValidator, GithubValidators)(github)
		if err != nil {
			return err
		}
	}
	if htpasswd, ok := resourceData.GetOk("htpasswd"); ok {
		err := common2.ValidAllDiag(common2.ListOfMapValidator, HTPasswdValidators)(htpasswd)
		if err != nil {
			return err
		}
	}
	if gitlab, ok := resourceData.GetOk("gitlab"); ok {
		err := common2.ValidAllDiag(common2.ListOfMapValidator, GitlabValidators)(gitlab)
		if err != nil {
			return err
		}
	}
	if google, ok := resourceData.GetOk("google"); ok {
		err := common2.ValidAllDiag(common2.ListOfMapValidator, GoogleValidators)(google)
		if err != nil {
			return err
		}
	}
	if ldap, ok := resourceData.GetOk("ldap"); ok {
		err := common2.ValidAllDiag(common2.ListOfMapValidator, LDAPValidators)(ldap)
		if err != nil {
			return err
		}
		// validate `attributes` which it's a required attribute
		l := ldap.([]interface{})
		ldapMap := l[0].(map[string]interface{})
		err = common2.ValidAllDiag(common2.ListOfMapValidator, ldapAttrsValidator)(ldapMap["attributes"])
		if err != nil {
			return err
		}
	}
	return nil
}

func identityProviderFromResourceData(resourceData *schema.ResourceData) *IdentityProviderState {
	result := &IdentityProviderState{
		Cluster: resourceData.Get("cluster").(string),
		Name:    resourceData.Get("name").(string),
		ID:      resourceData.Id(),
	}

	result.MappingMethod = common2.GetOptionalString(resourceData, "mapping_method")
	result.HTPasswd = ExpandHTPASSFromResourceData(resourceData)
	result.Gitlab = ExpandGitlabFromResourceData(resourceData)
	result.Github = ExpandGithubFromResourceData(resourceData)
	result.Google = ExpandGoogleFromResourceData(resourceData)
	result.LDAP = ExpandLDAPFromResourceData(resourceData)
	result.OpenID = ExpandOpenIDFromResourceData(resourceData)
	return result
}

func resourceIdentityProviderRead(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin Reading")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	identityProviderState := identityProviderFromResourceData(resourceData)

	// Find the identity clusterservice:
	resource := clusterCollection.Cluster(identityProviderState.Cluster).
		IdentityProviders().
		IdentityProvider(identityProviderState.ID)
	get, err := resource.Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		summary := fmt.Sprintf("identity clusterservice (%s) of cluster (%s) not found, removing from state",
			identityProviderState.ID, identityProviderState.Cluster,
		)
		tflog.Warn(ctx, summary)
		resourceData.SetId("")
		return []diag.Diagnostic{
			{
				Severity: diag.Warning,
				Summary:  summary,
				Detail: fmt.Sprintf(
					"cluster (%s) not found, removing from state",
					identityProviderState.ID),
			}}
	} else if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't find identity clusterservice",
				Detail: fmt.Sprintf(
					"Can't find identity clusterservice with identifier '%s' for "+
						"cluster '%s': %v",
					identityProviderState.ID, identityProviderState.Cluster, err),
			}}
	}

	object := get.Body()
	identityProviderToResourceData(object, resourceData)
	return
}

func resourceIdentityProviderUpdate(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin updating")
	return []diag.Diagnostic{
		{
			Severity: diag.Error,
			Summary:  "IDP Update not supported.",
			Detail:   "This RHCS provider version does not support updating an existing IDP",
		}}
}
func resourceIdentityProviderDelete(ctx context.Context, resourceData *schema.ResourceData, meta any) (diags diag.Diagnostics) {
	tflog.Debug(ctx, "begin deleting")
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters()
	identityProviderState := identityProviderFromResourceData(resourceData)

	// Send the request to delete the identity clusterservice:
	resource := clusterCollection.Cluster(identityProviderState.Cluster).
		IdentityProviders().
		IdentityProvider(identityProviderState.ID)
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Can't delete identity clusterservice",
				Detail: fmt.Sprintf(
					"Can't delete identity clusterservice with identifier '%s' for "+
						"cluster '%s': %v",
					identityProviderState.ID, identityProviderState.Cluster, err,
				),
			}}
	}

	// Remove the state:
	resourceData.SetId("")
	return
}

func resourceIDPImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// To import an identity clusterservice, we need to know the cluster ID and the clusterservice name.
	fields := strings.Split(d.Id(), ",")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		return nil, fmt.Errorf("Invalid import identifier (%s), expected <cluster_id>,<provider_name>", d.Id())
	}

	clusterID := fields[0]
	providerName := fields[1]
	clusterCollection := meta.(*sdk.Connection).ClustersMgmt().V1().Clusters().Cluster(clusterID)

	providerID, err := getIDPIDFromName(ctx, clusterCollection, providerName)
	if err != nil {
		return nil, fmt.Errorf("Can't import identity clusterservice %v", err.Error())
	}

	d.SetId(providerID)
	d.Set("cluster", clusterID)
	d.Set("name", providerName)

	return []*schema.ResourceData{d}, nil
}

// getIDPIDFromName returns the ID of the identity clusterservice with the given name.
func getIDPIDFromName(ctx context.Context, client *cmv1.ClusterClient, name string) (string, error) {
	tflog.Debug(ctx, fmt.Sprintf("Converting IDP name to ID, name %s", name))
	// Get the list of identity providers for the cluster:
	pClient := client.IdentityProviders()
	identityProviders := []*cmv1.IdentityProvider{}
	page := 1
	size := 100
	for {
		resp, err := pClient.List().
			Page(page).
			Size(size).
			SendContext(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list identity providers: %v", err)
		}
		identityProviders = append(identityProviders, resp.Items().Slice()...)
		if resp.Size() < size {
			break
		}
		page++
	}

	// Find the identity clusterservice with the given name
	for _, item := range identityProviders {

		if item.Name() == name {
			id := item.ID()
			tflog.Debug(ctx, fmt.Sprintf("Found IDP name %s id %s", name, id))
			return id, nil
		}
	}

	return "", fmt.Errorf("identity clusterservice '%s' not found", name)
}
