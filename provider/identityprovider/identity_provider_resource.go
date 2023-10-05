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

package identityprovider

import (
	"context"
	"fmt"
	"net/http"
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
)

var _ resource.ResourceWithConfigure = &IdentityProviderResource{}
var _ resource.ResourceWithImportState = &IdentityProviderResource{}
var _ resource.ResourceWithValidateConfig = &IdentityProviderResource{}

var validMappingMethods = []string{"claim", "add", "generate", "lookup"} // Default is @ index 0
var defaultMappingMethod = validMappingMethods[0]

var listOfIDPTypesPathes = []path.Expression{
	path.MatchRoot("github"),
	path.MatchRoot("gitlab"),
	path.MatchRoot("google"),
	path.MatchRoot("htpasswd"),
	path.MatchRoot("ldap"),
	path.MatchRoot("openid"),
}

type IdentityProviderResource struct {
	collection *cmv1.ClustersClient
}

func New() resource.Resource {
	return &IdentityProviderResource{}
}
func (r *IdentityProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_provider"
}

func (r *IdentityProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Identity provider.",
		Attributes: map[string]schema.Attribute{
			"cluster": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Unique identifier of the identity provider.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the identity provider.",
				Required:    true,
			},
			"mapping_method": schema.StringAttribute{
				Description: "Specifies how new identities are mapped to users when they log in. Options are `add`, `claim`, `generate` and `lookup`. (default is `claim`)",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(validMappingMethods...),
				},
				Default: stringdefault.StaticString(defaultMappingMethod),
			},
			"htpasswd": schema.SingleNestedAttribute{
				Description: "Details of the 'htpasswd' identity provider.",
				Attributes:  htpasswdSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...),
				},
			},
			"gitlab": schema.SingleNestedAttribute{
				Description: "Details of the Gitlab identity provider.",
				Attributes:  gitlabSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...),
				},
			},
			"github": schema.SingleNestedAttribute{
				Description: "Details of the Github identity provider.",
				Attributes:  githubSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...),
				},
			},
			"google": schema.SingleNestedAttribute{
				Description: "Details of the Google identity provider.",
				Attributes:  googleSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...),
				},
			},
			"ldap": schema.SingleNestedAttribute{
				Description: "Details of the LDAP identity provider.",
				Attributes:  ldapSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...),
				},
			},
			"openid": schema.SingleNestedAttribute{
				Description: "Details of the OpenID identity provider.",
				Attributes:  openidSchema,
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(listOfIDPTypesPathes...),
				},
			},
		},
	}
	return
}

func (r *IdentityProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	collection, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.Connaction, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.collection = collection.ClustersMgmt().V1().Clusters()
}

func (r *IdentityProviderResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	idp := &IdentityProviderState{}
	diag := req.Config.Get(ctx, idp)

	if diag.HasError() {
		return
	}

	//TODO: add validations

}

func (r *IdentityProviderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	// Get the plan:
	state := &IdentityProviderState{}
	diags := request.Plan.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resource := r.collection.Cluster(state.Cluster.ValueString())
	// We expect the cluster to be already exist
	// Try to get it and if result with NotFound error, return error to user
	if resp, err := resource.Get().SendContext(ctx); err != nil && resp.Status() == http.StatusNotFound {
		message := fmt.Sprintf("Cluster %s not found, error: %v", state.Cluster.ValueString(), err)
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
				state.Cluster.ValueString(), err,
			),
		)
		return
	}

	// Create the identity provider:
	builder := cmv1.NewIdentityProvider()
	builder.Name(state.Name.ValueString())
	// handle mapping_method
	mappingMethod := defaultMappingMethod
	if common.HasValue(state.MappingMethod) {
		mappingMethod = state.MappingMethod.ValueString()
	}
	builder.MappingMethod(cmv1.IdentityProviderMappingMethod(mappingMethod))
	switch {
	case state.HTPasswd != nil:
		builder.Type(cmv1.IdentityProviderTypeHtpasswd)
		htpasswdBuilder := CreateHTPasswdIDPBuilder(ctx, state.HTPasswd)
		builder.Htpasswd(htpasswdBuilder)
	case state.Gitlab != nil:
		builder.Type(cmv1.IdentityProviderTypeGitlab)
		gitlabBuilder, err := CreateGitlabIDPBuilder(ctx, state.Gitlab)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.Gitlab(gitlabBuilder)
	case state.Github != nil:
		builder.Type(cmv1.IdentityProviderTypeGithub)
		githubBuilder, err := CreateGithubIDPBuilder(ctx, state.Github)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.Github(githubBuilder)
	case state.Google != nil:
		builder.Type(cmv1.IdentityProviderTypeGoogle)
		googleBuilder, err := CreateGoogleIDPBuilder(ctx, mappingMethod, state.Google)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.Google(googleBuilder)
	case state.LDAP != nil:
		builder.Type(cmv1.IdentityProviderTypeLDAP)
		ldapBuilder, err := CreateLDAPIDPBuilder(ctx, state.LDAP)
		if err != nil {
			response.Diagnostics.AddError(err.Error(), err.Error())
			return
		}
		builder.LDAP(ldapBuilder)
	case state.OpenID != nil:
		builder.Type(cmv1.IdentityProviderTypeOpenID)
		openidBuilder, err := CreateOpenIDIDPBuilder(ctx, state.OpenID)
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
				state.Name.ValueString(), err,
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
				state.Name.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return
	}
	object = add.Body()

	state.ID = types.StringValue(object.ID())

	state.MappingMethod = types.StringValue(string(object.MappingMethod()))
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
			state.LDAP = &LDAPIdentityProvider{}
		}
		insecure, ok := ldapObject.GetInsecure()
		if ok {
			state.LDAP.Insecure = types.BoolValue(insecure)
		}
	case openidObject != nil:
	}

	// Save the state:
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}

func (r *IdentityProviderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	// Get the current state:
	state := &IdentityProviderState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Find the identity provider:
	resource := r.collection.Cluster(state.Cluster.ValueString()).
		IdentityProviders().
		IdentityProvider(state.ID.ValueString())
	get, err := resource.Get().SendContext(ctx)
	if err != nil && get.Status() == http.StatusNotFound {
		tflog.Warn(ctx, fmt.Sprintf("identity provider (%s) of cluster (%s) not found, removing from state",
			state.ID.ValueString(), state.Cluster.ValueString(),
		))
		response.State.RemoveResource(ctx)
		return
	} else if err != nil {
		response.Diagnostics.AddError(
			"Can't find identity provider",
			fmt.Sprintf(
				"Can't find identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return
	}
	object := get.Body()

	// Copy the identity provider data into the state:
	state.Name = types.StringValue(object.Name())
	state.MappingMethod = types.StringValue(string(object.MappingMethod()))
	htpasswdObject := object.Htpasswd()
	gitlabObject := object.Gitlab()
	ldapObject := object.LDAP()
	openidObject := object.OpenID()
	githubObject := object.Github()
	googleObject := object.Google()
	switch {
	case htpasswdObject != nil:
		if state.HTPasswd == nil {
			state.HTPasswd = &HTPasswdIdentityProvider{}
		}
		if users, ok := htpasswdObject.GetUsers(); ok {
			users.Each(func(item *cmv1.HTPasswdUser) bool {
				state.HTPasswd.Users = append(state.HTPasswd.Users, HTPasswdUser{
					Username: types.StringValue(item.Username()),
					Password: types.StringValue(item.Password()),
				})
				return true
			})
		}
	case gitlabObject != nil:
		if state.Gitlab == nil {
			state.Gitlab = &GitlabIdentityProvider{}
		}
		ca, ok := gitlabObject.GetCA()
		if ok {
			state.Gitlab.CA = types.StringValue(ca)
		}
		client_id, ok := gitlabObject.GetClientID()
		if ok {
			state.Gitlab.ClientID = types.StringValue(client_id)
		}
		client_secret, ok := gitlabObject.GetClientSecret()
		if ok {
			state.Gitlab.ClientSecret = types.StringValue(client_secret)
		}
		url, ok := gitlabObject.GetURL()
		if ok {
			state.Gitlab.URL = types.StringValue(url)
		}
	case githubObject != nil:
		if state.Github == nil {
			state.Github = &GithubIdentityProvider{}
		}
		ca, ok := githubObject.GetCA()
		if ok {
			state.Github.CA = types.StringValue(ca)
		}
		client_id, ok := githubObject.GetClientID()
		if ok {
			state.Github.ClientID = types.StringValue(client_id)
		}
		client_secret, ok := githubObject.GetClientSecret()
		if ok {
			state.Github.ClientSecret = types.StringValue(client_secret)
		}
		hostname, ok := githubObject.GetHostname()
		if ok {
			state.Github.Hostname = types.StringValue(hostname)
		}
		teams, ok := githubObject.GetTeams()
		if ok {
			state.Github.Teams, err = common.StringArrayToList(teams)
			if err != nil {
				response.Diagnostics.AddError("failed to convert string slice to tf list", "GitHub Teams conversion failed")
			}
		} else {
			state.Github.Teams = types.ListNull(types.StringType)
		}
		orgs, ok := githubObject.GetOrganizations()
		if ok {
			state.Github.Organizations, err = common.StringArrayToList(orgs)
			if err != nil {
				response.Diagnostics.AddError("failed to convert string slice to tf list", "GitHub Organizations conversion failed")
			}
		} else {
			state.Github.Organizations = types.ListNull(types.StringType)
		}
	case googleObject != nil:
		if state.Google == nil {
			state.Google = &GoogleIdentityProvider{}
		}
		if client_id, ok := googleObject.GetClientID(); ok {
			state.Google.ClientID = types.StringValue(client_id)
		}
		if client_secret, ok := googleObject.GetClientSecret(); ok {
			state.Google.ClientSecret = types.StringValue(client_secret)
		}
		if hosted_domain, ok := googleObject.GetHostedDomain(); ok {
			state.Google.HostedDomain = types.StringValue(hosted_domain)
		}
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &LDAPIdentityProvider{}
		}
		bindDN, ok := ldapObject.GetBindDN()
		if ok {
			state.LDAP.BindDN = types.StringValue(bindDN)
		}
		bindPassword, ok := ldapObject.GetBindPassword()
		if ok {
			state.LDAP.BindPassword = types.StringValue(bindPassword)
		}
		ca, ok := ldapObject.GetCA()
		if ok {
			state.LDAP.CA = types.StringValue(ca)
		}
		insecure, ok := ldapObject.GetInsecure()
		if ok {
			state.LDAP.Insecure = types.BoolValue(insecure)
		}
		url, ok := ldapObject.GetURL()
		if ok {
			state.LDAP.URL = types.StringValue(url)
		}
		attributes, ok := ldapObject.GetAttributes()
		if ok {
			if state.LDAP.Attributes == nil {
				state.LDAP.Attributes = &LDAPIdentityProviderAttributes{}
			}
			id, ok := attributes.GetID()
			if ok {
				state.LDAP.Attributes.ID, err = common.StringArrayToList(id)
				if err != nil {
					response.Diagnostics.AddError("failed to convert LDAP attribute ID to tf list", err.Error())
				}
			}
			email, ok := attributes.GetEmail()
			if ok {
				state.LDAP.Attributes.EMail, err = common.StringArrayToList(email)
				if err != nil {
					response.Diagnostics.AddError("failed to convert LDAP attribute EMail to tf list", err.Error())
				}
			} else {
				state.LDAP.Attributes.EMail = types.ListNull(types.StringType)
			}
			name, ok := attributes.GetName()
			if ok {
				state.LDAP.Attributes.Name, err = common.StringArrayToList(name)
				if err != nil {
					response.Diagnostics.AddError("failed to convert LDAP attribute Name to tf list", err.Error())
				}
			}
			preferredUsername, ok := attributes.GetPreferredUsername()
			if ok {
				state.LDAP.Attributes.PreferredUsername, err = common.StringArrayToList(preferredUsername)
				if err != nil {
					response.Diagnostics.AddError("failed to convert LDAP attribute PreferredUsername to tf list", err.Error())
				}
			}
		}
	case openidObject != nil:
		if state.OpenID == nil {
			state.OpenID = &OpenIDIdentityProvider{}
		}
		ca, ok := openidObject.GetCA()
		if ok {
			state.OpenID.CA = types.StringValue(ca)
		}
		client_id, ok := openidObject.GetClientID()
		if ok {
			state.OpenID.ClientID = types.StringValue(client_id)
		}
		client_secret, ok := openidObject.GetClientSecret()
		if ok {
			state.OpenID.ClientSecret = types.StringValue(client_secret)
		}
		claims, ok := openidObject.GetClaims()
		if ok {
			if state.OpenID.Claims == nil {
				state.OpenID.Claims = &OpenIDIdentityProviderClaims{}
			}
			email, ok := claims.GetEmail()
			if ok {
				state.OpenID.Claims.EMail, err = common.StringArrayToList(email)
				if err != nil {
					response.Diagnostics.AddError("failed to convert OpenID claims EMail to tf list", err.Error())
				}
			}
			groups, ok := claims.GetGroups()
			if ok {
				state.OpenID.Claims.Groups, err = common.StringArrayToList(groups)
				if err != nil {
				}
			}
			name, ok := claims.GetName()
			if ok {
				state.OpenID.Claims.Name, err = common.StringArrayToList(name)
				if err != nil {
					response.Diagnostics.AddError("failed to convert OpenID claims Name to tf list", err.Error())
				}
			}
			preferredUsername, ok := claims.GetPreferredUsername()
			if ok {
				state.OpenID.Claims.PreferredUsername, err = common.StringArrayToList(preferredUsername)
				if err != nil {
					response.Diagnostics.AddError("failed to convert OpenID claims PreferredUsername to tf list", err.Error())
				}
			}
		}
		issuer, ok := openidObject.GetIssuer()
		if ok {
			state.OpenID.Issuer = types.StringValue(issuer)
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

func (r *IdentityProviderResource) Update(ctx context.Context, request resource.UpdateRequest,
	response *resource.UpdateResponse) {
	response.Diagnostics.AddError("IDP Update not supported.", "This RHCS provider version does not support updating an existing IDP")
}

func (r *IdentityProviderResource) Delete(ctx context.Context, request resource.DeleteRequest,
	response *resource.DeleteResponse) {
	// Get the state:
	state := &IdentityProviderState{}
	diags := request.State.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send the request to delete the identity provider:
	resource := r.collection.Cluster(state.Cluster.ValueString()).
		IdentityProviders().
		IdentityProvider(state.ID.ValueString())
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete identity provider",
			fmt.Sprintf(
				"Can't delete identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.ValueString(), state.Cluster.ValueString(), err,
			),
		)
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx)
}

func (r *IdentityProviderResource) ImportState(ctx context.Context, request resource.ImportStateRequest,
	response *resource.ImportStateResponse) {
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

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("cluster"), clusterID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), providerID)...)
}

// getIDPIDFromName returns the ID of the identity provider with the given name.
func getIDPIDFromName(ctx context.Context, client *cmv1.ClusterClient, name string) (string, error) {
	tflog.Debug(ctx, "Converting IDP name to ID", map[string]interface{}{"name": name})
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
			tflog.Debug(ctx, "Found IDP", map[string]interface{}{"name": name, "id": id})
			return id, nil
		}
	}

	return "", fmt.Errorf("identity provider '%s' not found", name)
}
