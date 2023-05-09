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
***REMOVED***
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/idps"
***REMOVED***

type IdentityProviderResourceType struct {
}

type IdentityProviderResource struct {
	logger     logging.Logger
	collection *cmv1.ClustersClient
}

func (t *IdentityProviderResourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "Identity provider.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster": {
				Description: "Identifier of the cluster.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"id": {
				Description: "Unique identifier of the identity provider.",
				Type:        types.StringType,
				Computed:    true,
	***REMOVED***,
			"name": {
				Description: "Name of the identity provider.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"htpasswd": {
				Description: "Details of the 'htpasswd' identity provider.",
				Attributes:  idps.HtpasswdSchema(***REMOVED***,
				Optional:    true,
	***REMOVED***,
			"gitlab": {
				Description: "Details of the Gitlab identity provider.",
				Attributes:  idps.GitlabSchema(***REMOVED***,
				Optional:    true,
				Validators:  idps.GitlabValidators(***REMOVED***,
	***REMOVED***,
			"github": {
				Description: "Details of the Github identity provider.",
				Attributes:  idps.GithubSchema(***REMOVED***,
				Optional:    true,
				Validators:  idps.GithubValidators(***REMOVED***,
	***REMOVED***,
			"ldap": {
				Description: "Details of the LDAP identity provider.",
				Attributes:  idps.LdapSchema(***REMOVED***,
				Optional:    true,
	***REMOVED***,
			"openid": {
				Description: "Details of the OpenID identity provider.",
				Attributes:  idps.OpenidSchema(***REMOVED***,
				Optional:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *IdentityProviderResourceType***REMOVED*** NewResource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.Resource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	// use it directly when needed.
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	collection := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***

	// Create the resource:
	result = &IdentityProviderResource{
		logger:     parent.logger,
		collection: collection,
	}

	return
}

func (r *IdentityProviderResource***REMOVED*** Create(ctx context.Context,
	request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse***REMOVED*** {
	// Get the plan:
	state := &IdentityProviderState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Wait till the cluster is ready:
	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***
	pollCtx, cancel := context.WithTimeout(ctx, 1*time.Hour***REMOVED***
	defer cancel(***REMOVED***
	_, err := resource.Poll(***REMOVED***.
		Interval(30 * time.Second***REMOVED***.
		Predicate(func(get *cmv1.ClusterGetResponse***REMOVED*** bool {
			return get.Body(***REMOVED***.State(***REMOVED*** == cmv1.ClusterStateReady
***REMOVED******REMOVED***.
		StartContext(pollCtx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't poll cluster state",
			fmt.Sprintf(
				"Can't poll state of cluster with identifier '%s': %v",
				state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Create the identity provider:
	builder := cmv1.NewIdentityProvider(***REMOVED***
	builder.Name(state.Name.Value***REMOVED***
	switch {
	case state.HTPasswd != nil:
		builder.Type(cmv1.IdentityProviderTypeHtpasswd***REMOVED***
		htpasswdBuilder := idps.CreateHTPasswdIDPBuilder(ctx, state.HTPasswd***REMOVED***
		builder.Htpasswd(htpasswdBuilder***REMOVED***
	case state.Gitlab != nil:
		builder.Type(cmv1.IdentityProviderTypeGitlab***REMOVED***
		gitlabBuilder, err := idps.CreateGitlabIDPBuilder(ctx, state.Gitlab***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(err.Error(***REMOVED***, err.Error(***REMOVED******REMOVED***
			return
***REMOVED***
		builder.Gitlab(gitlabBuilder***REMOVED***
	case state.Github != nil:
		builder.Type(cmv1.IdentityProviderTypeGithub***REMOVED***
		githubBuilder, err := idps.CreateGithubIDPBuilder(ctx, state.Github***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(err.Error(***REMOVED***, err.Error(***REMOVED******REMOVED***
			return
***REMOVED***
		builder.Github(githubBuilder***REMOVED***
	case state.LDAP != nil:
		builder.Type(cmv1.IdentityProviderTypeLDAP***REMOVED***
		ldapBuilder, err := idps.CreateLdapIDPBuilder(ctx, state.LDAP***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(err.Error(***REMOVED***, err.Error(***REMOVED******REMOVED***
			return
***REMOVED***
		builder.LDAP(ldapBuilder***REMOVED***
	case state.OpenID != nil:
		builder.Type(cmv1.IdentityProviderTypeOpenID***REMOVED***
		openidBuilder, err := idps.CreateOpenIDIDPBuilder(ctx, state.OpenID***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(err.Error(***REMOVED***, err.Error(***REMOVED******REMOVED***
			return
***REMOVED***
		builder.OpenID(openidBuilder***REMOVED***
	}
	object, err := builder.Build(***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't build identity provider",
			fmt.Sprintf(
				"Can't build identity provider with name '%s': %v",
				state.Name.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	collection := resource.IdentityProviders(***REMOVED***
	add, err := collection.Add(***REMOVED***.Body(object***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't create identity provider",
			fmt.Sprintf(
				"Can't create identity provider with name '%s' for "+
					"cluster '%s': %v",
				state.Name.Value, state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	// Set the computed attributes:
	state.ID = types.String{
		Value: object.ID(***REMOVED***,
	}
	htpasswdObject := object.Htpasswd(***REMOVED***
	gitlabObject := object.Gitlab(***REMOVED***
	ldapObject := object.LDAP(***REMOVED***
	openidObject := object.OpenID(***REMOVED***
	switch {
	case htpasswdObject != nil:
		// Nothing, there are no computed attributes for `htpasswd` identity providers.
	case gitlabObject != nil:
		// Nothing, there are no computed attributes for `gitlab` identity providers.
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &idps.LDAPIdentityProvider{}
***REMOVED***
		insecure, ok := ldapObject.GetInsecure(***REMOVED***
		if ok {
			state.LDAP.Insecure = types.Bool{
				Value: insecure,
	***REMOVED***
***REMOVED***
	case openidObject != nil:
	}

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *IdentityProviderResource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadResourceRequest,
	response *tfsdk.ReadResourceResponse***REMOVED*** {
	// Get the current state:
	state := &IdentityProviderState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the identity provider:
	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***.
		IdentityProviders(***REMOVED***.
		IdentityProvider(state.ID.Value***REMOVED***
	get, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find identity provider",
			fmt.Sprintf(
				"Can't find identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Copy the identity provider data into the state:
	state.Name = types.String{
		Value: object.Name(***REMOVED***,
	}
	htpasswdObject := object.Htpasswd(***REMOVED***
	gitlabObject := object.Gitlab(***REMOVED***
	ldapObject := object.LDAP(***REMOVED***
	openidObject := object.OpenID(***REMOVED***
	githubObject := object.Github(***REMOVED***
	switch {
	case htpasswdObject != nil:
		if state.HTPasswd == nil {
			state.HTPasswd = &idps.HTPasswdIdentityProvider{}
***REMOVED***
		username, ok := htpasswdObject.GetUsername(***REMOVED***
		if ok {
			state.HTPasswd.Username = types.String{
				Value: username,
	***REMOVED***
***REMOVED***
		password, ok := htpasswdObject.GetPassword(***REMOVED***
		if ok {
			state.HTPasswd.Password = types.String{
				Value: password,
	***REMOVED***
***REMOVED***
	case gitlabObject != nil:
		if state.Gitlab == nil {
			state.Gitlab = &idps.GitlabIdentityProvider{}
***REMOVED***
		ca, ok := gitlabObject.GetCA(***REMOVED***
		if ok {
			state.Gitlab.CA = types.String{
				Value: ca,
	***REMOVED***
***REMOVED***
		client_id, ok := gitlabObject.GetClientID(***REMOVED***
		if ok {
			state.Gitlab.ClientID = types.String{
				Value: client_id,
	***REMOVED***
***REMOVED***
		client_secret, ok := gitlabObject.GetClientSecret(***REMOVED***
		if ok {
			state.Gitlab.ClientSecret = types.String{
				Value: client_secret,
	***REMOVED***
***REMOVED***
		url, ok := gitlabObject.GetURL(***REMOVED***
		if ok {
			state.Gitlab.URL = types.String{
				Value: url,
	***REMOVED***
***REMOVED***
	case githubObject != nil:
		if state.Github == nil {
			state.Github = &idps.GithubIdentityProvider{}
***REMOVED***
		ca, ok := githubObject.GetCA(***REMOVED***
		if ok {
			state.Github.CA = types.String{
				Value: ca,
	***REMOVED***
***REMOVED***
		client_id, ok := githubObject.GetClientID(***REMOVED***
		if ok {
			state.Github.ClientID = types.String{
				Value: client_id,
	***REMOVED***
***REMOVED***
		client_secret, ok := githubObject.GetClientSecret(***REMOVED***
		if ok {
			state.Github.ClientSecret = types.String{
				Value: client_secret,
	***REMOVED***
***REMOVED***
		hostname, ok := githubObject.GetHostname(***REMOVED***
		if ok {
			state.Github.Hostname = types.String{
				Value: hostname,
	***REMOVED***
***REMOVED***
		teams, ok := githubObject.GetTeams(***REMOVED***
		if ok {
			state.Github.Teams = common.StringArrayToList(teams***REMOVED***
***REMOVED***
		orgs, ok := githubObject.GetOrganizations(***REMOVED***
		if ok {
			state.Github.Organizations = common.StringArrayToList(orgs***REMOVED***
***REMOVED***
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &idps.LDAPIdentityProvider{}
***REMOVED***
		bindDN, ok := ldapObject.GetBindDN(***REMOVED***
		if ok {
			state.LDAP.BindDN = types.String{
				Value: bindDN,
	***REMOVED***
***REMOVED***
		bindPassword, ok := ldapObject.GetBindPassword(***REMOVED***
		if ok {
			state.LDAP.BindPassword = types.String{
				Value: bindPassword,
	***REMOVED***
***REMOVED***
		ca, ok := ldapObject.GetCA(***REMOVED***
		if ok {
			state.LDAP.CA = types.String{
				Value: ca,
	***REMOVED***
***REMOVED***
		insecure, ok := ldapObject.GetInsecure(***REMOVED***
		if ok {
			state.LDAP.Insecure = types.Bool{
				Value: insecure,
	***REMOVED***
***REMOVED***
		url, ok := ldapObject.GetURL(***REMOVED***
		if ok {
			state.LDAP.URL = types.String{
				Value: url,
	***REMOVED***
***REMOVED***
		attributes, ok := ldapObject.GetAttributes(***REMOVED***
		if ok {
			if state.LDAP.Attributes == nil {
				state.LDAP.Attributes = &idps.LDAPIdentityProviderAttributes{}
	***REMOVED***
			id, ok := attributes.GetID(***REMOVED***
			if ok {
				state.LDAP.Attributes.ID = common.StringArrayToList(id***REMOVED***
	***REMOVED***
			email, ok := attributes.GetEmail(***REMOVED***
			if ok {
				state.LDAP.Attributes.EMail = common.StringArrayToList(email***REMOVED***
	***REMOVED***
			name, ok := attributes.GetName(***REMOVED***
			if ok {
				state.LDAP.Attributes.Name = common.StringArrayToList(name***REMOVED***
	***REMOVED***
			preferredUsername, ok := attributes.GetPreferredUsername(***REMOVED***
			if ok {
				state.LDAP.Attributes.PreferredUsername = common.StringArrayToList(preferredUsername***REMOVED***
	***REMOVED***
***REMOVED***
	case openidObject != nil:
		if state.OpenID == nil {
			state.OpenID = &idps.OpenIDIdentityProvider{}
***REMOVED***
		ca, ok := openidObject.GetCA(***REMOVED***
		if ok {
			state.OpenID.CA = types.String{
				Value: ca,
	***REMOVED***
***REMOVED***
		client_id, ok := openidObject.GetClientID(***REMOVED***
		if ok {
			state.OpenID.ClientID = types.String{
				Value: client_id,
	***REMOVED***
***REMOVED***
		client_secret, ok := openidObject.GetClientSecret(***REMOVED***
		if ok {
			state.OpenID.ClientSecret = types.String{
				Value: client_secret,
	***REMOVED***
***REMOVED***
		claims, ok := openidObject.GetClaims(***REMOVED***
		if ok {
			if state.OpenID.Claims == nil {
				state.OpenID.Claims = &idps.OpenIDIdentityProviderClaims{}
	***REMOVED***
			email, ok := claims.GetEmail(***REMOVED***
			if ok {
				state.OpenID.Claims.EMail = common.StringArrayToList(email***REMOVED***
	***REMOVED***
			groups, ok := claims.GetGroups(***REMOVED***
			if ok {
				state.OpenID.Claims.Groups = common.StringArrayToList(groups***REMOVED***
	***REMOVED***
			name, ok := claims.GetName(***REMOVED***
			if ok {
				state.OpenID.Claims.Name = common.StringArrayToList(name***REMOVED***
	***REMOVED***
			preferredUsername, ok := claims.GetPreferredUsername(***REMOVED***
			if ok {
				state.OpenID.Claims.PreferredUsername = common.StringArrayToList(preferredUsername***REMOVED***
	***REMOVED***
***REMOVED***
		issuer, ok := openidObject.GetIssuer(***REMOVED***
		if ok {
			state.OpenID.Issuer = types.String{
				Value: issuer,
	***REMOVED***
***REMOVED***
	}
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *IdentityProviderResource***REMOVED*** Update(ctx context.Context, request tfsdk.UpdateResourceRequest,
	response *tfsdk.UpdateResourceResponse***REMOVED*** {
}

func (r *IdentityProviderResource***REMOVED*** Delete(ctx context.Context, request tfsdk.DeleteResourceRequest,
	response *tfsdk.DeleteResourceResponse***REMOVED*** {
	// Get the state:
	state := &IdentityProviderState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete the identity provider:
	resource := r.collection.Cluster(state.Cluster.Value***REMOVED***.
		IdentityProviders(***REMOVED***.
		IdentityProvider(state.ID.Value***REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete identity provider",
			fmt.Sprintf(
				"Can't delete identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.Cluster.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *IdentityProviderResource***REMOVED*** ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse***REMOVED*** {
	tfsdk.ResourceImportStatePassthroughID(
		ctx,
		tftypes.NewAttributePath(***REMOVED***.WithAttributeName("id"***REMOVED***,
		request,
		response,
	***REMOVED***
}
