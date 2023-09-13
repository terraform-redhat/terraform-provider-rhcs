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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type IdentityProviderResourceType struct {
}

type IdentityProviderResource struct {
	collection *cmv1.ClustersClient
}

func (t *IdentityProviderResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "Identity provider.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
			},
			"id": {
				Description: "Unique identifier of the identity provider.",
				Type:        types.StringType,
				Computed:    true,
			},
			"name": {
				Description: "Name of the identity provider.",
				Type:        types.StringType,
				Required:    true,
			},
			"mapping_method": {
				Description: "Specifies how new identities are mapped to users when they log in. Options are [add claim generate lookup] (default 'claim')",
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				Validators:  idps.MappingMethodValidators(),
			},
			"htpasswd": {
				Description: "Details of the 'htpasswd' identity provider.",
				Attributes:  idps.HtpasswdSchema(),
				Optional:    true,
				Validators:  idps.HTPasswdValidators(),
			},
			"gitlab": {
				Description: "Details of the Gitlab identity provider.",
				Attributes:  idps.GitlabSchema(),
				Optional:    true,
				Validators:  idps.GitlabValidators(),
			},
			"github": {
				Description: "Details of the Github identity provider.",
				Attributes:  idps.GithubSchema(),
				Optional:    true,
				Validators:  idps.GithubValidators(),
			},
			"google": {
				Description: "Details of the Google identity provider.",
				Attributes:  idps.GoogleSchema(),
				Optional:    true,
				Validators:  idps.GoogleValidators(),
			},
			"ldap": {
				Description: "Details of the LDAP identity provider.",
				Attributes:  idps.LDAPSchema(),
				Optional:    true,
				Validators:  idps.LDAPValidators(),
			},
			"openid": {
				Description: "Details of the OpenID identity provider.",
				Attributes:  idps.OpenidSchema(),
				Optional:    true,
			},
		},
	}
	return
}

func (t *IdentityProviderResourceType) NewResource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.Resource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation:
	// use it directly when needed.
	parent := p.(*Provider)

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt().V1().Clusters()

	// Create the resource:
	result = &IdentityProviderResource{
		collection: collection,
	}

	return
}

func (r *IdentityProviderResource) Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	// Get the plan:
	state := &IdentityProviderState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resource := r.collection.Cluster(state.Cluster.Value)
	// We expect the cluster to be already exist
	// Try to get it and if result with NotFound error, return error to user
	if resp, err := resource.Get().SendContext(ctx); err != nil && resp.Status() == http.StatusNotFound {
		message := fmt.Sprintf("Cluster %s not found, error: %v", state.Cluster.Value, err)
		tflog.Error(ctx, message)
		response.Diagnostics.AddError(
			"Can't poll cluster state",
			message,
		)
		return
	}

	// Wait till the cluster is ready:
	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()
	_, err := resource.Poll().
		Interval(30 * time.Second).
		Predicate(func(get *cmv1.ClusterGetResponse) bool {
			return get.Body().State() == cmv1.ClusterStateReady
		}).
		StartContext(pollCtx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't poll cluster state",
			fmt.Sprintf(
				"Can't poll state of cluster with identifier '%s': %v",
				state.Cluster.Value, err,
			),
		)
		return
	}

	// Create the identity provider:
	builder := cmv1.NewIdentityProvider()
	builder.Name(state.Name.Value)
	// handle mapping_method
	mappingMethod := idps.DefaultMappingMethod
	if !state.MappingMethod.Unknown && !state.MappingMethod.Null {
		mappingMethod = state.MappingMethod.Value
	}
	builder.MappingMethod(cmv1.IdentityProviderMappingMethod(mappingMethod))
	switch {
	case state.HTPasswd != nil:
		builder.Type(cmv1.IdentityProviderTypeHtpasswd)
		htpasswdBuilder := idps.CreateHTPasswdIDPBuilder(ctx, state.HTPasswd)
		builder.Htpasswd(htpasswdBuilder)
	case state.Gitlab != nil:
		builder.Type(cmv1.IdentityProviderTypeGitlab)
		gitlabBuilder, err := idps.CreateGitlabIDPBuilder(ctx, state.Gitlab)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.Gitlab(gitlabBuilder)
	case state.Github != nil:
		builder.Type(cmv1.IdentityProviderTypeGithub)
		githubBuilder, err := idps.CreateGithubIDPBuilder(ctx, state.Github)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.Github(githubBuilder)
	case state.Google != nil:
		builder.Type(cmv1.IdentityProviderTypeGoogle)
		googleBuilder, err := idps.CreateGoogleIDPBuilder(ctx, mappingMethod, state.Google)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.Google(googleBuilder)
	case state.LDAP != nil:
		builder.Type(cmv1.IdentityProviderTypeLDAP)
		ldapBuilder, err := idps.CreateLDAPIDPBuilder(ctx, state.LDAP)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.LDAP(ldapBuilder)
	case state.OpenID != nil:
		builder.Type(cmv1.IdentityProviderTypeOpenID)
		openidBuilder, err := idps.CreateOpenIDIDPBuilder(ctx, state.OpenID)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.OpenID(openidBuilder)
	}
	object, err := builder.Build()
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build identity provider",
			fmt.Sprintf(
				"Can't build identity provider with name '%s': %v",
				state.Name.Value, err,
			),
		)
		return
	}
	collection := resource.IdentityProviders()
	add, err := collection.Add().Body(object).SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't create identity provider",
			fmt.Sprintf(
				"Can't create identity provider with name '%s' for "+
					"cluster '%s': %v",
				state.Name.Value, state.Cluster.Value, err,
			),
		)
		return
	}
	object = add.Body()

	// Set the computed attributes:
	state.ID = types.String{
		Value: object.ID(),
	}
	state.MappingMethod = types.String{
		Value: string(object.MappingMethod()),
	}
	htpasswdObject := object.Htpasswd()
	gitlabObject := object.Gitlab()
	ldapObject := object.LDAP()
	openidObject := object.OpenID()
	switch {
	case htpasswdObject != nil:
		// Nothing, there are no computed attributes for `htpasswd` identity providers.
	case gitlabObject != nil:
		// Nothing, there are no computed attributes for `gitlab` identity providers.
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &idps.LDAPIdentityProvider{}
		}
		insecure, ok := ldapObject.GetInsecure()
		if ok {
			state.LDAP.Insecure = types.Bool{
				Value: insecure,
			}
		}
	case openidObject != nil:
	}

	// Save the state:
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *IdentityProviderResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse) {
	// Get the current state:
	state := &IdentityProviderState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the identity provider:
	resource := r.collection.Cluster(state.Cluster.Value).
		IdentityProviders().
		IdentityProvider(state.ID.Value)
	get, err := resource.Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("identity provider (%s) of cluster (%s) not found, removing from state",
			state.ID.Value, state.Cluster.Value,
		))
		response.State.RemoveResource(ctx)
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Can't find identity provider",
			fmt.Sprintf(
				"Can't find identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			),
		)
		return
	}

	object := get.Body()

	// Copy the identity provider data into the state:
	state.Name = types.String{
		Value: object.Name(),
	}

	htpasswdObject := object.Htpasswd()
	gitlabObject := object.Gitlab()
	ldapObject := object.LDAP()
	openidObject := object.OpenID()
	githubObject := object.Github()
	googleObject := object.Google()
	switch {
	case htpasswdObject != nil:
		if state.HTPasswd == nil {
			state.HTPasswd = &idps.HTPasswdIdentityProvider{}
		}
		if users, ok := htpasswdObject.GetUsers(); ok {
			users.Each(func(item *cmv1.HTPasswdUser) bool {
				state.HTPasswd.Users = append(state.HTPasswd.Users, idps.HTPasswdUser{
					Username: types.String{
						Value: item.Username(),
					},
					Password: types.String{
						Value: item.Password(),
					},
				})
				return true
			})
		}
	case gitlabObject != nil:
		if state.Gitlab == nil {
			state.Gitlab = &idps.GitlabIdentityProvider{}
		}
		ca, ok := gitlabObject.GetCA()
		if ok {
			state.Gitlab.CA = types.String{
				Value: ca,
			}
		}
		client_id, ok := gitlabObject.GetClientID()
		if ok {
			state.Gitlab.ClientID = types.String{
				Value: client_id,
			}
		}
		client_secret, ok := gitlabObject.GetClientSecret()
		if ok {
			state.Gitlab.ClientSecret = types.String{
				Value: client_secret,
			}
		}
		url, ok := gitlabObject.GetURL()
		if ok {
			state.Gitlab.URL = types.String{
				Value: url,
			}
		}
	case githubObject != nil:
		if state.Github == nil {
			state.Github = &idps.GithubIdentityProvider{}
		}
		ca, ok := githubObject.GetCA()
		if ok {
			state.Github.CA = types.String{
				Value: ca,
			}
		}
		client_id, ok := githubObject.GetClientID()
		if ok {
			state.Github.ClientID = types.String{
				Value: client_id,
			}
		}
		client_secret, ok := githubObject.GetClientSecret()
		if ok {
			state.Github.ClientSecret = types.String{
				Value: client_secret,
			}
		}
		hostname, ok := githubObject.GetHostname()
		if ok {
			state.Github.Hostname = types.String{
				Value: hostname,
			}
		}
		teams, ok := githubObject.GetTeams()
		if ok {
			state.Github.Teams = common.StringArrayToList(teams)
		}
		orgs, ok := githubObject.GetOrganizations()
		if ok {
			state.Github.Organizations = common.StringArrayToList(orgs)
		}
	case googleObject != nil:
		if state.Google == nil {
			state.Google = &idps.GoogleIdentityProvider{}
		}
		if client_id, ok := googleObject.GetClientID(); ok {
			state.Google.ClientID = types.String{
				Value: client_id,
			}
		}
		if client_secret, ok := googleObject.GetClientSecret(); ok {
			state.Google.ClientSecret = types.String{
				Value: client_secret,
			}
		}
		if hosted_domain, ok := googleObject.GetHostedDomain(); ok {
			state.Google.HostedDomain = types.String{
				Value: hosted_domain,
			}
		}
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &idps.LDAPIdentityProvider{}
		}
		bindDN, ok := ldapObject.GetBindDN()
		if ok {
			state.LDAP.BindDN = types.String{
				Value: bindDN,
			}
		}
		bindPassword, ok := ldapObject.GetBindPassword()
		if ok {
			state.LDAP.BindPassword = types.String{
				Value: bindPassword,
			}
		}
		ca, ok := ldapObject.GetCA()
		if ok {
			state.LDAP.CA = types.String{
				Value: ca,
			}
		}
		insecure, ok := ldapObject.GetInsecure()
		if ok {
			state.LDAP.Insecure = types.Bool{
				Value: insecure,
			}
		}
		url, ok := ldapObject.GetURL()
		if ok {
			state.LDAP.URL = types.String{
				Value: url,
			}
		}
		attributes, ok := ldapObject.GetAttributes()
		if ok {
			if state.LDAP.Attributes == nil {
				state.LDAP.Attributes = &idps.LDAPIdentityProviderAttributes{}
			}
			id, ok := attributes.GetID()
			if ok {
				state.LDAP.Attributes.ID = common.StringArrayToList(id)
			}
			email, ok := attributes.GetEmail()
			if ok {
				state.LDAP.Attributes.EMail = common.StringArrayToList(email)
			}
			name, ok := attributes.GetName()
			if ok {
				state.LDAP.Attributes.Name = common.StringArrayToList(name)
			}
			preferredUsername, ok := attributes.GetPreferredUsername()
			if ok {
				state.LDAP.Attributes.PreferredUsername = common.StringArrayToList(preferredUsername)
			}
		}
	case openidObject != nil:
		if state.OpenID == nil {
			state.OpenID = &idps.OpenIDIdentityProvider{}
		}
		ca, ok := openidObject.GetCA()
		if ok {
			state.OpenID.CA = types.String{
				Value: ca,
			}
		}
		client_id, ok := openidObject.GetClientID()
		if ok {
			state.OpenID.ClientID = types.String{
				Value: client_id,
			}
		}
		client_secret, ok := openidObject.GetClientSecret()
		if ok {
			state.OpenID.ClientSecret = types.String{
				Value: client_secret,
			}
		}
		claims, ok := openidObject.GetClaims()
		if ok {
			if state.OpenID.Claims == nil {
				state.OpenID.Claims = &idps.OpenIDIdentityProviderClaims{}
			}
			email, ok := claims.GetEmail()
			if ok {
				state.OpenID.Claims.EMail = common.StringArrayToList(email)
			}
			groups, ok := claims.GetGroups()
			if ok {
				state.OpenID.Claims.Groups = common.StringArrayToList(groups)
			}
			name, ok := claims.GetName()
			if ok {
				state.OpenID.Claims.Name = common.StringArrayToList(name)
			}
			preferredUsername, ok := claims.GetPreferredUsername()
			if ok {
				state.OpenID.Claims.PreferredUsername = common.StringArrayToList(preferredUsername)
			}
		}
		issuer, ok := openidObject.GetIssuer()
		if ok {
			state.OpenID.Issuer = types.String{
				Value: issuer,
			}
		}
	}
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Save the state:
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *IdentityProviderResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse) {
	response.Diagnostics.AddError("IDP Update not supported.", "This RHCS provider version does not support updating an existing IDP")
}

func (r *IdentityProviderResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse) {
	// Get the state:
	state := &IdentityProviderState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send the request to delete the identity provider:
	resource := r.collection.Cluster(state.Cluster.Value).
		IdentityProviders().
		IdentityProvider(state.ID.Value)
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete identity provider",
			fmt.Sprintf(
				"Can't delete identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			),
		)
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx)
}

func (r *IdentityProviderResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	// To import an identity provider, we need to know the cluster ID and the provider name.
	fields := strings.Split(request.ID, ",")
	if len(fields) != 2 || fields[0] == "" || fields[1] == "" {
		response.Diagnostics.AddError(
			"Invalid import identifier",
			"Identity provider to import should be specified as <cluster_id>,<provider_name>",
		)
		return
	}
	clusterID := fields[0]
	providerName := fields[1]

	client := r.collection.Cluster(clusterID)
	providerID, err := getIDPIDFromName(ctx, client, providerName)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't import identity provider",
			err.Error(),
		)
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("cluster"), clusterID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("id"), providerID)...)
}

// getIDPIDFromName returns the ID of the identity provider with the given name.
func getIDPIDFromName(ctx context.Context, client *cmv1.ClusterClient, name string) (string, error) {
	tflog.Debug(ctx, "Converting IDP name to ID", "name", name)
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

	// Find the identity provider with the given name
	for _, item := range identityProviders {
		if item.Name() == name {
			id := item.ID()
			tflog.Debug(ctx, "Found IDP", "name", name, "id", id)
			return id, nil
		}
	}

	return "", fmt.Errorf("identity provider '%s' not found", name)
}
