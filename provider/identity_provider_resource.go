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
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
)

type IdentityProviderResourceType struct {
}

type IdentityProviderResource struct {
	logger     logging.Logger
	collection *cmv1.ClustersClient
}

func (t *IdentityProviderResourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "Identity provider.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster_id": {
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
			"htpasswd": {
				Description: "Details of the 'htpasswd' identity provider.",
				Attributes:  t.htpasswdSchema(),
				Optional:    true,
			},
			"ldap": {
				Description: "Details of the LDAP identity provider.",
				Attributes:  t.ldapSchema(),
				Optional:    true,
			},
		},
	}
	return
}

func (t *IdentityProviderResourceType) htpasswdSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"username": {
			Description: "User name.",
			Type:        types.StringType,
			Required:    true,
		},
		"password": {
			Description: "User password.",
			Type:        types.StringType,
			Required:    true,
			Sensitive:   true,
		},
	})
}

func (t *IdentityProviderResourceType) ldapSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"bind_dn": {
			Type:     types.StringType,
			Required: true,
		},
		"bind_password": {
			Type:      types.StringType,
			Required:  true,
			Sensitive: true,
		},
		"ca": {
			Type:     types.StringType,
			Optional: true,
		},
		"insecure": {
			Type:     types.BoolType,
			Optional: true,
			Computed: true,
		},
		"url": {
			Type:     types.StringType,
			Required: true,
		},
		"attributes": {
			Attributes: t.ldapAttributesSchema(),
			Required:   true,
		},
	})
}

func (t *IdentityProviderResourceType) ldapAttributesSchema() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"email": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"id": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"name": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"preferred_username": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
	})
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
		logger:     parent.logger,
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

	// Wait till the cluster is ready:
	resource := r.collection.Cluster(state.ClusterID.Value)
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
				state.ClusterID.Value, err,
			),
		)
		return
	}

	// Create the identity provider:
	builder := cmv1.NewIdentityProvider()
	builder.Name(state.Name.Value)
	switch {
	case state.HTPasswd != nil:
		builder.Type(cmv1.IdentityProviderType("HTPasswdIdentityProvider"))
		htpasswdBuilder := cmv1.NewHTPasswdIdentityProvider()
		if !state.HTPasswd.Username.Null {
			htpasswdBuilder.Username(state.HTPasswd.Username.Value)
		}
		if !state.HTPasswd.Password.Null {
			htpasswdBuilder.Password(state.HTPasswd.Password.Value)
		}
		builder.Htpasswd(htpasswdBuilder)
	case state.LDAP != nil:
		builder.Type(cmv1.IdentityProviderType("LDAPIdentityProvider"))
		ldapBuilder := cmv1.NewLDAPIdentityProvider()
		if !state.LDAP.BindDN.Null {
			ldapBuilder.BindDN(state.LDAP.BindDN.Value)
		}
		if !state.LDAP.BindPassword.Null {
			ldapBuilder.BindPassword(state.LDAP.BindPassword.Value)
		}
		if !state.LDAP.CA.Null {
			ldapBuilder.CA(state.LDAP.CA.Value)
		}
		if !state.LDAP.Insecure.Null {
			ldapBuilder.Insecure(state.LDAP.Insecure.Value)
		}
		if !state.LDAP.URL.Null {
			ldapBuilder.URL(state.LDAP.URL.Value)
		}
		if state.LDAP.Attributes != nil {
			attributesBuilder := cmv1.NewLDAPAttributes()
			if state.LDAP.Attributes.ID != nil {
				attributesBuilder.ID(state.LDAP.Attributes.ID...)
			}
			if state.LDAP.Attributes.EMail != nil {
				attributesBuilder.Email(state.LDAP.Attributes.EMail...)
			}
			if state.LDAP.Attributes.Name != nil {
				attributesBuilder.Name(state.LDAP.Attributes.Name...)
			}
			if state.LDAP.Attributes.PreferredUsername != nil {
				attributesBuilder.PreferredUsername(
					state.LDAP.Attributes.PreferredUsername...,
				)
			}
			ldapBuilder.Attributes(attributesBuilder)
		}
		builder.LDAP(ldapBuilder)
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
				state.Name.Value, state.ClusterID.Value, err,
			),
		)
		return
	}
	object = add.Body()

	// Set the computed attributes:
	state.ID = types.String{
		Value: object.ID(),
	}
	htpasswdObject := object.Htpasswd()
	ldapObject := object.LDAP()
	switch {
	case htpasswdObject != nil:
		// Nothing, there are no computed attributes for `htpasswd` identity providers.
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &LDAPIdentityProvider{}
		}
		insecure, ok := ldapObject.GetInsecure()
		if ok {
			state.LDAP.Insecure = types.Bool{
				Value: insecure,
			}
		}
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
	resource := r.collection.Cluster(state.ClusterID.Value).
		IdentityProviders().
		IdentityProvider(state.ID.Value)
	get, err := resource.Get().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find identity provider",
			fmt.Sprintf(
				"Can't find identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.ClusterID.Value, err,
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
	ldapObject := object.LDAP()
	switch {
	case htpasswdObject != nil:
		if state.HTPasswd == nil {
			state.HTPasswd = &HTPasswdIdentityProvider{}
		}
		username, ok := htpasswdObject.GetUsername()
		if ok {
			state.HTPasswd.Username = types.String{
				Value: username,
			}
		}
		password, ok := htpasswdObject.GetPassword()
		if ok {
			state.HTPasswd.Password = types.String{
				Value: password,
			}
		}
	case ldapObject != nil:
		if state.LDAP == nil {
			state.LDAP = &LDAPIdentityProvider{}
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
				state.LDAP.Attributes = &LDAPIdentityProviderAttributes{}
			}
			id, ok := attributes.GetID()
			if ok {
				state.LDAP.Attributes.ID = id
			}
			email, ok := attributes.GetEmail()
			if ok {
				state.LDAP.Attributes.EMail = email
			}
			name, ok := attributes.GetName()
			if ok {
				state.LDAP.Attributes.Name = name
			}
			preferredUsername, ok := attributes.GetPreferredUsername()
			if ok {
				state.LDAP.Attributes.PreferredUsername = preferredUsername
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
	resource := r.collection.Cluster(state.ClusterID.Value).
		IdentityProviders().
		IdentityProvider(state.ID.Value)
	_, err := resource.Delete().SendContext(ctx)
	if err != nil {
		response.Diagnostics.AddError(
			"Can't delete identity provider",
			fmt.Sprintf(
				"Can't delete identity provider with identifier '%s' for "+
					"cluster '%s': %v",
				state.ID.Value, state.ClusterID.Value, err,
			),
		)
		return
	}

	// Remove the state:
	response.State.RemoveResource(ctx)
}

func (r *IdentityProviderResource) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest,
	response *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(
		ctx,
		tftypes.NewAttributePath().WithAttributeName("id"),
		request,
		response,
	)
}
