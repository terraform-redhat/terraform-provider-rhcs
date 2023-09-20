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

package identityprovider

***REMOVED***
	"context"
***REMOVED***
***REMOVED***
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

var _ resource.ResourceWithConfigure = &IdentityProviderResource{}
var _ resource.ResourceWithImportState = &IdentityProviderResource{}
var _ resource.ResourceWithValidateConfig = &IdentityProviderResource{}

var validMappingMethods = []string{"claim", "add", "generate", "lookup"} // Default is @ index 0
var defaultMappingMethod = validMappingMethods[0]

var listOfIDPTypesPathes = []path.Expression{
	path.MatchRoot("github"***REMOVED***,
	path.MatchRoot("gitlab"***REMOVED***,
	path.MatchRoot("google"***REMOVED***,
	path.MatchRoot("htpasswd"***REMOVED***,
	path.MatchRoot("ldap"***REMOVED***,
	path.MatchRoot("openid"***REMOVED***,
}

type IdentityProviderResource struct {
	collection *cmv1.ClustersClient
}

func New(***REMOVED*** resource.Resource {
	return &IdentityProviderResource{}
}
func (r *IdentityProviderResource***REMOVED*** Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_identity_provider"
}

func (r *IdentityProviderResource***REMOVED*** Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "Identity provider.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
	***REMOVED***,
			"id": schema.StringAttribute{
				Description: "Unique identifier of the identity provider.",
				Computed:    true,
	***REMOVED***,
			"name": schema.StringAttribute{
				Description: "Name of the identity provider.",
				Required:    true,
	***REMOVED***,
			"mapping_method": schema.StringAttribute{
				Description: "Specifies how new identities are mapped to users when they log in. Options are [add claim generate lookup] (default 'claim'***REMOVED***",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(validMappingMethods...***REMOVED***,
		***REMOVED***,
				Default: stringdefault.StaticString(defaultMappingMethod***REMOVED***,
	***REMOVED***,
			"htpasswd": schema.SingleNestedAttribute{
				Description: "Details of the 'htpasswd' identity provider.",
				Attributes:  htpasswdSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"gitlab": schema.SingleNestedAttribute{
				Description: "Details of the Gitlab identity provider.",
				Attributes:  gitlabSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"github": schema.SingleNestedAttribute{
				Description: "Details of the Github identity provider.",
				Attributes:  githubSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"google": schema.SingleNestedAttribute{
				Description: "Details of the Google identity provider.",
				Attributes:  googleSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"ldap": schema.SingleNestedAttribute{
				Description: "Details of the LDAP identity provider.",
				Attributes:  ldapSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...***REMOVED***,
		***REMOVED***,
	***REMOVED***,
			"openid": schema.SingleNestedAttribute{
				Description: "Details of the OpenID identity provider.",
				Attributes:  openidSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...***REMOVED***,
		***REMOVED***,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (r *IdentityProviderResource***REMOVED*** Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse***REMOVED*** {
	if req.ProviderData == nil {
		return
	}

	collection, ok := req.ProviderData.(*sdk.Connection***REMOVED***
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData***REMOVED***,
		***REMOVED***
		return
	}

	r.collection = collection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
}

func (r *IdentityProviderResource***REMOVED*** ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse***REMOVED*** {
	idp := &IdentityProviderState{}
	diag := req.Config.Get(ctx, idp***REMOVED***

	if diag.HasError(***REMOVED*** {
		return
	}

	//TODO: add validations

}

func (r *IdentityProviderResource***REMOVED*** Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse***REMOVED*** {
	// Get the plan:
	state := &IdentityProviderState{}
	diags := request.Plan.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	resource := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***
	// We expect the cluster to be already exist
	// Try to get it and if result with NotFound error, return error to user
	if resp, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***; err != nil && resp.Status(***REMOVED*** == http.StatusNotFound {
		message := fmt.Sprintf("Cluster %s not found, error: %v", state.Cluster.ValueString(***REMOVED***, err***REMOVED***
		tflog.Error(ctx, message***REMOVED***
		response.Diagnostics.AddError(
			"Can't poll cluster state",
			message,
		***REMOVED***
		return
	}

	// Wait till the cluster is ready:
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
				state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Create the identity provider:
	builder := cmv1.NewIdentityProvider(***REMOVED***
	builder.Name(state.Name.ValueString(***REMOVED******REMOVED***
	// handle mapping_method
	mappingMethod := defaultMappingMethod
	if common.HasValue(state.MappingMethod***REMOVED*** {
		mappingMethod = state.MappingMethod.ValueString(***REMOVED***
	}
	builder.MappingMethod(cmv1.IdentityProviderMappingMethod(mappingMethod***REMOVED******REMOVED***
	switch {
	case state.HTPasswd != nil:
		builder.Type(cmv1.IdentityProviderTypeHtpasswd***REMOVED***
		htpasswdBuilder := CreateHTPasswdIDPBuilder(ctx, state.HTPasswd***REMOVED***
		builder.Htpasswd(htpasswdBuilder***REMOVED***
	case state.Gitlab != nil:
		builder.Type(cmv1.IdentityProviderTypeGitlab***REMOVED***
		gitlabBuilder, err := CreateGitlabIDPBuilder(ctx, state.Gitlab***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(err.Error(***REMOVED***, err.Error(***REMOVED******REMOVED***
			return
***REMOVED***
		builder.Gitlab(gitlabBuilder***REMOVED***
	case state.Github != nil:
		builder.Type(cmv1.IdentityProviderTypeGithub***REMOVED***
		githubBuilder, err := CreateGithubIDPBuilder(ctx, state.Github***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(err.Error(***REMOVED***, err.Error(***REMOVED******REMOVED***
			return
***REMOVED***
		builder.Github(githubBuilder***REMOVED***
	case state.Google != nil:
		builder.Type(cmv1.IdentityProviderTypeGoogle***REMOVED***
		googleBuilder, err := CreateGoogleIDPBuilder(ctx, mappingMethod, state.Google***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(err.Error(***REMOVED***, err.Error(***REMOVED******REMOVED***
			return
***REMOVED***
		builder.Google(googleBuilder***REMOVED***
	case state.LDAP != nil:
		builder.Type(cmv1.IdentityProviderTypeLDAP***REMOVED***
		ldapBuilder, err := CreateLDAPIDPBuilder(ctx, state.LDAP***REMOVED***
		if err != nil {
			response.Diagnostics.AddError(err.Error(***REMOVED***, err.Error(***REMOVED******REMOVED***
			return
***REMOVED***
		builder.LDAP(ldapBuilder***REMOVED***
	case state.OpenID != nil:
		builder.Type(cmv1.IdentityProviderTypeOpenID***REMOVED***
		openidBuilder, err := CreateOpenIDIDPBuilder(ctx, state.OpenID***REMOVED***
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
				state.Name.ValueString(***REMOVED***, err,
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
				state.Name.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object = add.Body(***REMOVED***

	state.ID = types.StringValue(object.ID(***REMOVED******REMOVED***

	state.MappingMethod = types.StringValue(string(object.MappingMethod(***REMOVED******REMOVED******REMOVED***
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
			state.LDAP = &LDAPIdentityProvider{}
***REMOVED***
		insecure, ok := ldapObject.GetInsecure(***REMOVED***
		if ok {
			state.LDAP.Insecure = types.BoolValue(insecure***REMOVED***
***REMOVED***
	case openidObject != nil:
	}

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

func (r *IdentityProviderResource***REMOVED*** Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse***REMOVED*** {
	// Get the current state:
	state := &IdentityProviderState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Find the identity provider:
	resource := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.
		IdentityProviders(***REMOVED***.
		IdentityProvider(state.ID.ValueString(***REMOVED******REMOVED***
	get, err := resource.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil && get.Status(***REMOVED*** == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("identity provider (%s***REMOVED*** of cluster (%s***REMOVED*** not found, removing from state",
			state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***,
		***REMOVED******REMOVED***
		response.State.RemoveResource(ctx***REMOVED***
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Can't find identity provider",
			fmt.Sprintf(
				"Can't find identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***

	// Copy the identity provider data into the state:
	state.Name = types.StringValue(object.Name(***REMOVED******REMOVED***
	state.MappingMethod = types.StringValue(string(object.MappingMethod(***REMOVED******REMOVED******REMOVED***
	htpasswdObject := object.Htpasswd(***REMOVED***
	gitlabObject := object.Gitlab(***REMOVED***
	ldapObject := object.LDAP(***REMOVED***
	openidObject := object.OpenID(***REMOVED***
	githubObject := object.Github(***REMOVED***
	googleObject := object.Google(***REMOVED***
	switch {
	case htpasswdObject != nil:
		if state.HTPasswd == nil {
			state.HTPasswd = &HTPasswdIdentityProvider{}
***REMOVED***
		if users, ok := htpasswdObject.GetUsers(***REMOVED***; ok {
			users.Each(func(item *cmv1.HTPasswdUser***REMOVED*** bool {
				state.HTPasswd.Users = append(state.HTPasswd.Users, HTPasswdUser{
					Username: types.StringValue(item.Username(***REMOVED******REMOVED***,
					Password: types.StringValue(item.Password(***REMOVED******REMOVED***,
		***REMOVED******REMOVED***
				return true
	***REMOVED******REMOVED***
***REMOVED***
	case gitlabObject != nil:
		if state.Gitlab == nil {
			state.Gitlab = &GitlabIdentityProvider{}
***REMOVED***
		ca, ok := gitlabObject.GetCA(***REMOVED***
		if ok {
			state.Gitlab.CA = types.StringValue(ca***REMOVED***
***REMOVED***
		client_id, ok := gitlabObject.GetClientID(***REMOVED***
		if ok {
			state.Gitlab.ClientID = types.StringValue(client_id***REMOVED***
***REMOVED***
		client_secret, ok := gitlabObject.GetClientSecret(***REMOVED***
		if ok {
			state.Gitlab.ClientSecret = types.StringValue(client_secret***REMOVED***
***REMOVED***
		url, ok := gitlabObject.GetURL(***REMOVED***
		if ok {
			state.Gitlab.URL = types.StringValue(url***REMOVED***
***REMOVED***
	case githubObject != nil:
		if state.Github == nil {
			state.Github = &GithubIdentityProvider{}
***REMOVED***
		ca, ok := githubObject.GetCA(***REMOVED***
		if ok {
			state.Github.CA = types.StringValue(ca***REMOVED***
***REMOVED***
		client_id, ok := githubObject.GetClientID(***REMOVED***
		if ok {
			state.Github.ClientID = types.StringValue(client_id***REMOVED***
***REMOVED***
		client_secret, ok := githubObject.GetClientSecret(***REMOVED***
		if ok {
			state.Github.ClientSecret = types.StringValue(client_secret***REMOVED***
***REMOVED***
		hostname, ok := githubObject.GetHostname(***REMOVED***
		if ok {
			state.Github.Hostname = types.StringValue(hostname***REMOVED***
***REMOVED***
		teams, ok := githubObject.GetTeams(***REMOVED***
		if ok {
			state.Github.Teams, err = common.StringArrayToList(teams***REMOVED***
			if err != nil {
				response.Diagnostics.AddError("failed to convert string slice to tf list", "GitHub Teams conversion failed"***REMOVED***
	***REMOVED***
***REMOVED*** else {
			state.Github.Teams = types.ListNull(types.StringType***REMOVED***
***REMOVED***
		orgs, ok := githubObject.GetOrganizations(***REMOVED***
		if ok {
			state.Github.Organizations, err = common.StringArrayToList(orgs***REMOVED***
			if err != nil {
				response.Diagnostics.AddError("failed to convert string slice to tf list", "GitHub Organizations conversion failed"***REMOVED***
	***REMOVED***
***REMOVED*** else {
			state.Github.Organizations = types.ListNull(types.StringType***REMOVED***
***REMOVED***
	case googleObject != nil:
		if state.Google == nil {
			state.Google = &GoogleIdentityProvider{}
***REMOVED***
		if client_id, ok := googleObject.GetClientID(***REMOVED***; ok {
			state.Google.ClientID = types.StringValue(client_id***REMOVED***
***REMOVED***
		if client_secret, ok := googleObject.GetClientSecret(***REMOVED***; ok {
			state.Google.ClientSecret = types.StringValue(client_secret***REMOVED***
***REMOVED***
		if hosted_domain, ok := googleObject.GetHostedDomain(***REMOVED***; ok {
			state.Google.HostedDomain = types.StringValue(hosted_domain***REMOVED***
***REMOVED***
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &LDAPIdentityProvider{}
***REMOVED***
		bindDN, ok := ldapObject.GetBindDN(***REMOVED***
		if ok {
			state.LDAP.BindDN = types.StringValue(bindDN***REMOVED***
***REMOVED***
		bindPassword, ok := ldapObject.GetBindPassword(***REMOVED***
		if ok {
			state.LDAP.BindPassword = types.StringValue(bindPassword***REMOVED***
***REMOVED***
		ca, ok := ldapObject.GetCA(***REMOVED***
		if ok {
			state.LDAP.CA = types.StringValue(ca***REMOVED***
***REMOVED***
		insecure, ok := ldapObject.GetInsecure(***REMOVED***
		if ok {
			state.LDAP.Insecure = types.BoolValue(insecure***REMOVED***
***REMOVED***
		url, ok := ldapObject.GetURL(***REMOVED***
		if ok {
			state.LDAP.URL = types.StringValue(url***REMOVED***
***REMOVED***
		attributes, ok := ldapObject.GetAttributes(***REMOVED***
		if ok {
			if state.LDAP.Attributes == nil {
				state.LDAP.Attributes = &LDAPIdentityProviderAttributes{}
	***REMOVED***
			id, ok := attributes.GetID(***REMOVED***
			if ok {
				state.LDAP.Attributes.ID, err = common.StringArrayToList(id***REMOVED***
				if err != nil {
					response.Diagnostics.AddError("failed to convert LDAP attribute ID to tf list", err.Error(***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED***
			email, ok := attributes.GetEmail(***REMOVED***
			if ok {
				state.LDAP.Attributes.EMail, err = common.StringArrayToList(email***REMOVED***
				if err != nil {
					response.Diagnostics.AddError("failed to convert LDAP attribute EMail to tf list", err.Error(***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED*** else {
				state.LDAP.Attributes.EMail = types.ListNull(types.StringType***REMOVED***
	***REMOVED***
			name, ok := attributes.GetName(***REMOVED***
			if ok {
				state.LDAP.Attributes.Name, err = common.StringArrayToList(name***REMOVED***
				if err != nil {
					response.Diagnostics.AddError("failed to convert LDAP attribute Name to tf list", err.Error(***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED***
			preferredUsername, ok := attributes.GetPreferredUsername(***REMOVED***
			if ok {
				state.LDAP.Attributes.PreferredUsername, err = common.StringArrayToList(preferredUsername***REMOVED***
				if err != nil {
					response.Diagnostics.AddError("failed to convert LDAP attribute PreferredUsername to tf list", err.Error(***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED***
***REMOVED***
	case openidObject != nil:
		if state.OpenID == nil {
			state.OpenID = &OpenIDIdentityProvider{}
***REMOVED***
		ca, ok := openidObject.GetCA(***REMOVED***
		if ok {
			state.OpenID.CA = types.StringValue(ca***REMOVED***
***REMOVED***
		client_id, ok := openidObject.GetClientID(***REMOVED***
		if ok {
			state.OpenID.ClientID = types.StringValue(client_id***REMOVED***
***REMOVED***
		client_secret, ok := openidObject.GetClientSecret(***REMOVED***
		if ok {
			state.OpenID.ClientSecret = types.StringValue(client_secret***REMOVED***
***REMOVED***
		claims, ok := openidObject.GetClaims(***REMOVED***
		if ok {
			if state.OpenID.Claims == nil {
				state.OpenID.Claims = &OpenIDIdentityProviderClaims{}
	***REMOVED***
			email, ok := claims.GetEmail(***REMOVED***
			if ok {
				state.OpenID.Claims.EMail, err = common.StringArrayToList(email***REMOVED***
				if err != nil {
					response.Diagnostics.AddError("failed to convert OpenID claims EMail to tf list", err.Error(***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED***
			groups, ok := claims.GetGroups(***REMOVED***
			if ok {
				state.OpenID.Claims.Groups, err = common.StringArrayToList(groups***REMOVED***
				if err != nil {
		***REMOVED***
	***REMOVED***
			name, ok := claims.GetName(***REMOVED***
			if ok {
				state.OpenID.Claims.Name, err = common.StringArrayToList(name***REMOVED***
				if err != nil {
					response.Diagnostics.AddError("failed to convert OpenID claims Name to tf list", err.Error(***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED***
			preferredUsername, ok := claims.GetPreferredUsername(***REMOVED***
			if ok {
				state.OpenID.Claims.PreferredUsername, err = common.StringArrayToList(preferredUsername***REMOVED***
				if err != nil {
					response.Diagnostics.AddError("failed to convert OpenID claims PreferredUsername to tf list", err.Error(***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED***
***REMOVED***
		issuer, ok := openidObject.GetIssuer(***REMOVED***
		if ok {
			state.OpenID.Issuer = types.StringValue(issuer***REMOVED***
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

func (r *IdentityProviderResource***REMOVED*** Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse***REMOVED*** {
	response.Diagnostics.AddError("IDP Update not supported.", "This RHCS provider version does not support updating an existing IDP"***REMOVED***
}

func (r *IdentityProviderResource***REMOVED*** Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse***REMOVED*** {
	// Get the state:
	state := &IdentityProviderState{}
	diags := request.State.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	// Send the request to delete the identity provider:
	resource := r.collection.Cluster(state.Cluster.ValueString(***REMOVED******REMOVED***.
		IdentityProviders(***REMOVED***.
		IdentityProvider(state.ID.ValueString(***REMOVED******REMOVED***
	_, err := resource.Delete(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete identity provider",
			fmt.Sprintf(
				"Can't delete identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.ValueString(***REMOVED***, state.Cluster.ValueString(***REMOVED***, err,
			***REMOVED***,
		***REMOVED***
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx***REMOVED***
}

func (r *IdentityProviderResource***REMOVED*** ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse***REMOVED*** {
	// To import an identity provider, we need to know the cluster ID and the provider name.
	fields := strings.Split(request.ID, ","***REMOVED***
	if len(fields***REMOVED*** != 2 || fields[0] == "" || fields[1] == "" {
		response.Diagnostics.AddError(
			"Invalid import identifier",
			"Identity provider to import should be specified as <cluster_id>,<provider_name>",
		***REMOVED***
		return
	}
	clusterID := fields[0]
	providerName := fields[1]
	client := r.collection.Cluster(clusterID***REMOVED***
	providerID, err := getIDPIDFromName(ctx, client, providerName***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't import identity provider",
			err.Error(***REMOVED***,
		***REMOVED***
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("cluster"***REMOVED***, clusterID***REMOVED***...***REMOVED***
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"***REMOVED***, providerID***REMOVED***...***REMOVED***
}

// getIDPIDFromName returns the ID of the identity provider with the given name.
func getIDPIDFromName(ctx context.Context, client *cmv1.ClusterClient, name string***REMOVED*** (string, error***REMOVED*** {
	tflog.Debug(ctx, "Converting IDP name to ID", map[string]interface{}{"name": name}***REMOVED***
	// Get the list of identity providers for the cluster:
	pClient := client.IdentityProviders(***REMOVED***
	identityProviders := []*cmv1.IdentityProvider{}
	page := 1
	size := 100
	for {
		resp, err := pClient.List(***REMOVED***.
			Page(page***REMOVED***.
			Size(size***REMOVED***.
			SendContext(ctx***REMOVED***
		if err != nil {
			return "", fmt.Errorf("failed to list identity providers: %v", err***REMOVED***
***REMOVED***
		identityProviders = append(identityProviders, resp.Items(***REMOVED***.Slice(***REMOVED***...***REMOVED***
		if resp.Size(***REMOVED*** < size {
			break
***REMOVED***
		page++
	}

	// Find the identity provider with the given name
	for _, item := range identityProviders {
		if item.Name(***REMOVED*** == name {
			id := item.ID(***REMOVED***
			tflog.Debug(ctx, "Found IDP", map[string]interface{}{"name": name, "id": id}***REMOVED***
			return id, nil
***REMOVED***
	}

	return "", fmt.Errorf("identity provider '%s' not found", name***REMOVED***
}
