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
				Attributes:  t.htpasswdSchema(***REMOVED***,
				Optional:    true,
	***REMOVED***,
			"ldap": {
				Description: "Details of the LDAP identity provider.",
				Attributes:  t.ldapSchema(***REMOVED***,
				Optional:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *IdentityProviderResourceType***REMOVED*** htpasswdSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"username": {
			Description: "User name.",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"password": {
			Description: "User password.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
***REMOVED***,
	}***REMOVED***
}

func (t *IdentityProviderResourceType***REMOVED*** ldapSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"bind_dn": {
			Type:     types.StringType,
			Required: true,
***REMOVED***,
		"bind_password": {
			Type:      types.StringType,
			Required:  true,
			Sensitive: true,
***REMOVED***,
		"ca": {
			Type:     types.StringType,
			Optional: true,
***REMOVED***,
		"insecure": {
			Type:     types.BoolType,
			Optional: true,
			Computed: true,
***REMOVED***,
		"url": {
			Type:     types.StringType,
			Required: true,
***REMOVED***,
		"attributes": {
			Attributes: t.ldapAttributesSchema(***REMOVED***,
			Required:   true,
***REMOVED***,
	}***REMOVED***
}

func (t *IdentityProviderResourceType***REMOVED*** ldapAttributesSchema(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"email": {
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"id": {
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"name": {
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
		"preferred_username": {
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Optional: true,
***REMOVED***,
	}***REMOVED***
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
		builder.Type(cmv1.IdentityProviderType("HTPasswdIdentityProvider"***REMOVED******REMOVED***
		htpasswdBuilder := cmv1.NewHTPasswdIdentityProvider(***REMOVED***
		if !state.HTPasswd.Username.Null {
			htpasswdBuilder.Username(state.HTPasswd.Username.Value***REMOVED***
***REMOVED***
		if !state.HTPasswd.Password.Null {
			htpasswdBuilder.Password(state.HTPasswd.Password.Value***REMOVED***
***REMOVED***
		builder.Htpasswd(htpasswdBuilder***REMOVED***
	case state.LDAP != nil:
		builder.Type(cmv1.IdentityProviderType("LDAPIdentityProvider"***REMOVED******REMOVED***
		ldapBuilder := cmv1.NewLDAPIdentityProvider(***REMOVED***
		if !state.LDAP.BindDN.Null {
			ldapBuilder.BindDN(state.LDAP.BindDN.Value***REMOVED***
***REMOVED***
		if !state.LDAP.BindPassword.Null {
			ldapBuilder.BindPassword(state.LDAP.BindPassword.Value***REMOVED***
***REMOVED***
		if !state.LDAP.CA.Null {
			ldapBuilder.CA(state.LDAP.CA.Value***REMOVED***
***REMOVED***
		if !state.LDAP.Insecure.Null {
			ldapBuilder.Insecure(state.LDAP.Insecure.Value***REMOVED***
***REMOVED***
		if !state.LDAP.URL.Null {
			ldapBuilder.URL(state.LDAP.URL.Value***REMOVED***
***REMOVED***
		if state.LDAP.Attributes != nil {
			attributesBuilder := cmv1.NewLDAPAttributes(***REMOVED***
			if state.LDAP.Attributes.ID != nil {
				attributesBuilder.ID(state.LDAP.Attributes.ID...***REMOVED***
	***REMOVED***
			if state.LDAP.Attributes.EMail != nil {
				attributesBuilder.Email(state.LDAP.Attributes.EMail...***REMOVED***
	***REMOVED***
			if state.LDAP.Attributes.Name != nil {
				attributesBuilder.Name(state.LDAP.Attributes.Name...***REMOVED***
	***REMOVED***
			if state.LDAP.Attributes.PreferredUsername != nil {
				attributesBuilder.PreferredUsername(
					state.LDAP.Attributes.PreferredUsername...,
				***REMOVED***
	***REMOVED***
			ldapBuilder.Attributes(attributesBuilder***REMOVED***
***REMOVED***
		builder.LDAP(ldapBuilder***REMOVED***
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
	ldapObject := object.LDAP(***REMOVED***
	switch {
	case htpasswdObject != nil:
		// Nothing, there are no computed attributes for `htpasswd` identity providers.
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &LDAPIdentityProvider{}
***REMOVED***
		insecure, ok := ldapObject.GetInsecure(***REMOVED***
		if ok {
			state.LDAP.Insecure = types.Bool{
				Value: insecure,
	***REMOVED***
***REMOVED***
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
	ldapObject := object.LDAP(***REMOVED***
	switch {
	case htpasswdObject != nil:
		if state.HTPasswd == nil {
			state.HTPasswd = &HTPasswdIdentityProvider{}
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
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &LDAPIdentityProvider{}
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
				state.LDAP.Attributes = &LDAPIdentityProviderAttributes{}
	***REMOVED***
			id, ok := attributes.GetID(***REMOVED***
			if ok {
				state.LDAP.Attributes.ID = id
	***REMOVED***
			email, ok := attributes.GetEmail(***REMOVED***
			if ok {
				state.LDAP.Attributes.EMail = email
	***REMOVED***
			name, ok := attributes.GetName(***REMOVED***
			if ok {
				state.LDAP.Attributes.Name = name
	***REMOVED***
			preferredUsername, ok := attributes.GetPreferredUsername(***REMOVED***
			if ok {
				state.LDAP.Attributes.PreferredUsername = preferredUsername
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
