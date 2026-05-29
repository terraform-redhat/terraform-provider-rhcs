/*
Copyright (c) 2025 Red Hat, Inc.

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

package wifconfig

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// WifConfigResource implements the rhcs_wif_config resource.
type WifConfigResource struct {
	wifConfigs *cmv1.WifConfigsClient
}

var _ resource.Resource = &WifConfigResource{}
var _ resource.ResourceWithConfigure = &WifConfigResource{}
var _ resource.ResourceWithImportState = &WifConfigResource{}
var _ resource.ResourceWithModifyPlan = &WifConfigResource{}

// New creates a new WIF config resource.
func New() resource.Resource {
	return &WifConfigResource{}
}

func (r *WifConfigResource) Metadata(
	ctx context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_wif_config"
}

func (r *WifConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Workload Identity Federation (WIF) configuration for OSD clusters on GCP. " +
			"Best practice: use one WIF config per cluster. " +
			"WIF configs are version-specific (openshift_version); " +
			"one-per-cluster simplifies lifecycle (create/destroy together) " +
			"and avoids upgrade conflicts when clusters differ in version.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the WIF config.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable display name for the WIF config. Immutable after create.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization": schema.StringAttribute{
				Description: "OCM organization ID owning this config.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"openshift_version": schema.StringAttribute{
				Description: "OpenShift version (x.y.z) to scope WIF IAM resources. Patch (.z) is stripped for OCM (roles use " +
					"x.y). When set, OCM returns a version-specific blueprint. Omitted when unset.",
				Optional: true,
			},
			"gcp": schema.SingleNestedAttribute{
				Description: "GCP-specific WIF configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"project_id": schema.StringAttribute{
						Description: "GCP project ID where WIF resources are configured.",
						Required:    true,
					},
					"project_number": schema.StringAttribute{
						Description: "GCP project number for WIF resources.",
						Required:    true,
					},
					"role_prefix": schema.StringAttribute{
						Description: "Prefix for GCP custom role names.",
						Required:    true,
					},
					"federated_project_id": schema.StringAttribute{
						Description: "GCP project ID where WorkloadIdentityPool resources are configured (computed by OCM, defaults " +
							"to project_id).",
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"federated_project_number": schema.StringAttribute{
						Description: "GCP project number for WorkloadIdentityPool (computed by OCM).",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"impersonator_email": schema.StringAttribute{
						Description: "Service account email used by OCM to access other service accounts (computed by OCM).",
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"workload_identity_pool": schema.SingleNestedAttribute{
						Description: "Workload identity pool blueprint from OCM (computed). Use with the example GCP WIF wiring under " +
							"examples/.",
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"pool_id": schema.StringAttribute{
								Description: "GCP workload identity pool ID.",
								Computed:    true,
							},
							"identity_provider": schema.SingleNestedAttribute{
								Description: "OIDC identity provider configuration for the pool.",
								Computed:    true,
								Attributes: map[string]schema.Attribute{
									"identity_provider_id": schema.StringAttribute{
										Description: "Identity provider ID (e.g., oidc).",
										Computed:    true,
									},
									"issuer_url": schema.StringAttribute{
										Description: "OIDC issuer URL.",
										Computed:    true,
									},
									"jwks": schema.StringAttribute{
										Description: "JWKS JSON string for token validation.",
										Computed:    true,
									},
									"allowed_audiences": schema.ListAttribute{
										Description: "Allowed OIDC audiences.",
										Computed:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
					"service_accounts": schema.ListNestedAttribute{
						Description: "Service accounts blueprint from OCM (computed). Each entry defines a GCP SA with roles and " +
							"access method.",
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"service_account_id": schema.StringAttribute{
									Description: "GCP service account ID.",
									Computed:    true,
								},
								"access_method": schema.StringAttribute{
									Description: "Access method: impersonate, wif, or vm.",
									Computed:    true,
								},
								"osd_role": schema.StringAttribute{
									Description: "OSD role name.",
									Computed:    true,
								},
								"roles": schema.ListNestedAttribute{
									Description: "IAM roles for this service account.",
									Computed:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"role_id": schema.StringAttribute{
												Description: "Role ID (predefined or custom).",
												Computed:    true,
											},
											"predefined": schema.BoolAttribute{
												Description: "True if this is a predefined GCP role.",
												Computed:    true,
											},
											"permissions": schema.ListAttribute{
												Description: "Permissions for custom roles.",
												Computed:    true,
												ElementType: types.StringType,
											},
											"resource_bindings": schema.ListNestedAttribute{
												Description: "Resource-level bindings (target SA for iam.serviceAccountUser etc.).",
												Computed:    true,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"type": schema.StringAttribute{
															Description: "Resource type (e.g., iam.serviceAccounts).",
															Computed:    true,
														},
														"name": schema.StringAttribute{
															Description: "Target resource name.",
															Computed:    true,
														},
													},
												},
											},
										},
									},
								},
								"credential_request": schema.SingleNestedAttribute{
									Description: "OpenShift credential request (namespace and SA names for WIF principals).",
									Computed:    true,
									Attributes: map[string]schema.Attribute{
										"namespace": schema.StringAttribute{
											Description: "OpenShift namespace for the credential secret.",
											Computed:    true,
										},
										"service_account_names": schema.ListAttribute{
											Description: "OpenShift service account names that can assume this GCP SA.",
											Computed:    true,
											ElementType: types.StringType,
										},
									},
								},
							},
						},
					},
					"support": schema.SingleNestedAttribute{
						Description: "Support access configuration from OCM (computed).",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"principal": schema.StringAttribute{
								Description: "Support group principal (e.g., sd-sre-platform-gcp-access@redhat.com).",
								Computed:    true,
							},
							"roles": schema.ListNestedAttribute{
								Description: "Roles bound to the support group.",
								Computed:    true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"role_id": schema.StringAttribute{Description: "Role ID.", Computed: true},
										"predefined": schema.BoolAttribute{
											Description: "True if predefined.",
											Computed:    true,
										},
										"permissions": schema.ListAttribute{
											Description: "Custom role permissions.",
											Computed:    true,
											ElementType: types.StringType,
										},
										"resource_bindings": schema.ListNestedAttribute{
											Description: "Resource bindings.",
											Computed:    true,
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"type": schema.StringAttribute{Computed: true},
													"name": schema.StringAttribute{Computed: true},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *WifConfigResource) Configure(
	ctx context.Context,
	req resource.ConfigureRequest,
	resp *resource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	r.wifConfigs = conn.ClustersMgmt().V1().GCP().WifConfigs()
}

// ModifyPlan marks OCM-computed fields as unknown during create so Terraform
// accepts the provider-returned values after apply. These nested Computed
// attributes inside a Required block are planned as null without this modifier.
func (r *WifConfigResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if req.State.Raw.IsNull() {
		var plan WifConfigState
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() || plan.GCP == nil {
			return
		}
		plan.GCP.FederatedProjectID = types.StringUnknown()
		plan.GCP.FederatedProjectNumber = types.StringUnknown()
		plan.GCP.ImpersonatorEmail = types.StringUnknown()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

func (r *WifConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WifConfigState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wifObj, err := r.buildWifConfigObject(&plan, false)
	if err != nil {
		resp.Diagnostics.AddError("failed to build WIF config", err.Error())
		return
	}

	addResp, err := r.wifConfigs.Add().Body(wifObj).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create WIF config", err.Error())
		return
	}
	obj := addResp.Body()

	r.populateState(obj, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WifConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WifConfigState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.wifConfigs.WifConfig(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		if getResp != nil && getResp.Status() == http.StatusNotFound {
			tflog.Warn(ctx, "WIF config not found, removing from state", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to get WIF config", err.Error())
		return
	}
	obj := getResp.Body()

	r.populateState(obj, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WifConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WifConfigState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wifObj, err := r.buildWifConfigObject(&plan, true)
	if err != nil {
		resp.Diagnostics.AddError("failed to build WIF config", err.Error())
		return
	}

	_, err = r.wifConfigs.WifConfig(plan.ID.ValueString()).Update().Body(wifObj).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to update WIF config", err.Error())
		return
	}

	getResp, err := r.wifConfigs.WifConfig(plan.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to get WIF config after update", err.Error())
		return
	}
	r.populateState(getResp.Body(), &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WifConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WifConfigState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// OCM's cluster->wif reference cleanup is eventually consistent: a
	// cluster Get can return 404 while the WIF still appears in-use. Retry
	// Delete with backoff so the destroy graph (cluster -> wif) doesn't
	// race that window. Also treat a 404 as success (idempotent).
	const maxAttempts = 12
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		delResp, err := r.wifConfigs.WifConfig(state.ID.ValueString()).Delete().SendContext(ctx)
		if err == nil {
			resp.State.RemoveResource(ctx)
			return
		}
		if delResp != nil && delResp.Status() == http.StatusNotFound {
			tflog.Warn(ctx, "WIF config already gone, treating delete as successful", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		if attempt == maxAttempts {
			resp.Diagnostics.AddError("failed to delete WIF config", err.Error())
			return
		}
		// Backoff: 15s, 30s, 45s ... up to a few minutes per attempt.
		// Covers the eventual-consistency window after cluster uninstall.
		wait := time.Duration(attempt*15) * time.Second
		tflog.Info(ctx, "WIF delete retry pending", map[string]interface{}{
			"id":      state.ID.ValueString(),
			"attempt": attempt,
			"wait":    wait.String(),
			"err":     err.Error(),
		})
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("failed to delete WIF config", ctx.Err().Error())
			return
		case <-time.After(wait):
		}
	}
}

func (r *WifConfigResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *WifConfigResource) buildWifConfigObject(s *WifConfigState, forUpdate bool) (*cmv1.WifConfig, error) {
	gcpBuilder := cmv1.NewWifGcp().
		ProjectId(s.GCP.ProjectID.ValueString()).
		ProjectNumber(s.GCP.ProjectNumber.ValueString()).
		RolePrefix(s.GCP.RolePrefix.ValueString())

	if common.HasValue(s.GCP.FederatedProjectID) {
		gcpBuilder.FederatedProjectId(s.GCP.FederatedProjectID.ValueString())
	}
	if common.HasValue(s.GCP.FederatedProjectNumber) {
		gcpBuilder.FederatedProjectNumber(s.GCP.FederatedProjectNumber.ValueString())
	}
	if common.HasValue(s.GCP.ImpersonatorEmail) {
		gcpBuilder.ImpersonatorEmail(s.GCP.ImpersonatorEmail.ValueString())
	}

	builder := cmv1.NewWifConfig()
	if !forUpdate {
		builder = builder.DisplayName(s.DisplayName.ValueString())
	}
	builder = builder.Gcp(gcpBuilder)

	if common.HasValue(s.OpenshiftVersion) {
		templateID := versionToTemplateID(s.OpenshiftVersion.ValueString())
		builder.WifTemplates(templateID)
	}

	return builder.Build()
}

// versionToTemplateID converts an OpenShift version (e.g. "4.21.3") to the OCM WIF template ID (e.g. "v4.21").
// WIF roles are x.y-scoped; the patch (.z) is stripped.
func versionToTemplateID(version string) string {
	re := regexp.MustCompile(`^(\d+\.\d+)(?:\.\d+)?$`)
	if m := re.FindStringSubmatch(version); len(m) > 1 {
		return "v" + m[1]
	}
	return version
}

func (r *WifConfigResource) populateState(obj *cmv1.WifConfig, state *WifConfigState) {
	state.ID = types.StringValue(obj.ID())
	state.DisplayName = types.StringValue(obj.DisplayName())
	if obj.Organization() != nil {
		state.Organization = types.StringValue(obj.Organization().ID())
	} else {
		state.Organization = types.StringValue("")
	}
	if obj.Gcp() != nil {
		gcp := obj.Gcp()
		state.GCP = &WifGcpState{
			ProjectID:     types.StringValue(gcp.ProjectId()),
			ProjectNumber: types.StringValue(gcp.ProjectNumber()),
			RolePrefix:    types.StringValue(gcp.RolePrefix()),
		}
		if gcp.FederatedProjectId() != "" {
			state.GCP.FederatedProjectID = types.StringValue(gcp.FederatedProjectId())
		}
		if gcp.FederatedProjectNumber() != "" {
			state.GCP.FederatedProjectNumber = types.StringValue(gcp.FederatedProjectNumber())
		}
		if gcp.ImpersonatorEmail() != "" {
			state.GCP.ImpersonatorEmail = types.StringValue(gcp.ImpersonatorEmail())
		}
		state.GCP.WorkloadIdentityPool = r.populatePoolState(gcp.WorkloadIdentityPool())
		state.GCP.ServiceAccounts = r.populateServiceAccountsState(gcp.ServiceAccounts())
		state.GCP.Support = r.populateSupportState(gcp.Support())
	}
}

func (r *WifConfigResource) populatePoolState(pool *cmv1.WifPool) types.Object {
	if pool == nil || pool.Empty() {
		return types.ObjectNull(poolAttrTypes)
	}
	attrs := map[string]attr.Value{"pool_id": types.StringValue(pool.PoolId())}
	if idp := pool.IdentityProvider(); idp != nil && !idp.Empty() {
		audiences, _ := types.ListValueFrom(context.Background(), types.StringType, idp.AllowedAudiences())
		attrs["identity_provider"] = types.ObjectValueMust(identityProviderAttrTypes, map[string]attr.Value{
			"identity_provider_id": types.StringValue(idp.IdentityProviderId()),
			"issuer_url":           types.StringValue(idp.IssuerUrl()),
			"jwks":                 types.StringValue(idp.Jwks()),
			"allowed_audiences":    audiences,
		})
	} else {
		attrs["identity_provider"] = types.ObjectNull(identityProviderAttrTypes)
	}
	obj, _ := types.ObjectValue(poolAttrTypes, attrs)
	return obj
}

func (r *WifConfigResource) populateServiceAccountsState(sas []*cmv1.WifServiceAccount) types.List {
	if len(sas) == 0 {
		return types.ListNull(serviceAccountObjectType)
	}
	// Copy and sort by service_account_id for deterministic state; OCM may return different order on each read.
	sorted := make([]*cmv1.WifServiceAccount, len(sas))
	copy(sorted, sas)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ServiceAccountId() < sorted[j].ServiceAccountId()
	})
	elems := make([]attr.Value, 0, len(sorted))
	for _, sa := range sorted {
		if sa == nil || sa.Empty() {
			continue
		}
		attrs := map[string]attr.Value{
			"service_account_id": types.StringValue(sa.ServiceAccountId()),
			"access_method":      types.StringValue(string(sa.AccessMethod())),
			"osd_role":           types.StringValue(sa.OsdRole()),
			"roles":              r.populateRolesState(sa.Roles()),
		}
		if cr := sa.CredentialRequest(); cr != nil && !cr.Empty() {
			var namespace string
			if sr := cr.SecretRef(); sr != nil && !sr.Empty() {
				namespace = sr.Namespace()
			}
			saNames := r.sortedStringList(cr.ServiceAccountNames())
			attrs["credential_request"] = types.ObjectValueMust(
				map[string]attr.Type{
					"namespace":             types.StringType,
					"service_account_names": types.ListType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"namespace":             types.StringValue(namespace),
					"service_account_names": saNames,
				},
			)
		} else {
			attrs["credential_request"] = types.ObjectNull(map[string]attr.Type{
				"namespace":             types.StringType,
				"service_account_names": types.ListType{ElemType: types.StringType},
			})
		}
		elems = append(elems, types.ObjectValueMust(serviceAccountAttrTypes, attrs))
	}
	listVal, err := types.ListValue(serviceAccountObjectType, elems)
	if err != nil {
		return types.ListNull(serviceAccountObjectType)
	}
	return listVal
}

// sortedStringList returns a sorted list of strings for deterministic state.
func (r *WifConfigResource) sortedStringList(sl []string) types.List {
	if len(sl) == 0 {
		return types.ListNull(types.StringType)
	}
	slCopy := make([]string, len(sl))
	_ = copy(slCopy, sl)
	sort.Strings(slCopy)
	list, _ := types.ListValueFrom(context.Background(), types.StringType, slCopy)
	return list
}

// sortedPermissionsList returns a sorted list of permissions for deterministic state.
func (r *WifConfigResource) sortedPermissionsList(perms []string) types.List {
	return r.sortedStringList(perms)
}

func (r *WifConfigResource) populateRolesState(roles []*cmv1.WifRole) types.List {
	if len(roles) == 0 {
		return types.ListNull(roleObjectType)
	}
	// Sort by role_id for deterministic state; OCM may return different order on each read.
	sorted := make([]*cmv1.WifRole, len(roles))
	copy(sorted, roles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RoleId() < sorted[j].RoleId()
	})
	elems := make([]attr.Value, 0, len(sorted))
	for _, role := range sorted {
		if role == nil || role.Empty() {
			continue
		}
		perms := r.sortedPermissionsList(role.Permissions())
		resourceBindings := r.populateResourceBindingsState(role.ResourceBindings())
		elems = append(elems, types.ObjectValueMust(roleAttrTypes, map[string]attr.Value{
			"role_id":           types.StringValue(role.RoleId()),
			"predefined":        types.BoolValue(role.Predefined()),
			"permissions":       perms,
			"resource_bindings": resourceBindings,
		}))
	}
	listVal, err := types.ListValue(roleObjectType, elems)
	if err != nil {
		return types.ListNull(roleObjectType)
	}
	return listVal
}

func (r *WifConfigResource) populateResourceBindingsState(bindings []*cmv1.WifResourceBinding) types.List {
	if len(bindings) == 0 {
		return types.ListNull(resourceBindingObjectType)
	}
	// Sort by type+name for deterministic state.
	sorted := make([]*cmv1.WifResourceBinding, len(bindings))
	copy(sorted, bindings)
	sort.Slice(sorted, func(i, j int) bool {
		ki, kj := sorted[i].Type()+"/"+sorted[i].Name(), sorted[j].Type()+"/"+sorted[j].Name()
		return ki < kj
	})
	elems := make([]attr.Value, 0, len(sorted))
	for _, b := range sorted {
		if b == nil || b.Empty() {
			continue
		}
		elems = append(elems, types.ObjectValueMust(resourceBindingAttrTypes, map[string]attr.Value{
			"type": types.StringValue(b.Type()),
			"name": types.StringValue(b.Name()),
		}))
	}
	listVal, err := types.ListValue(resourceBindingObjectType, elems)
	if err != nil {
		return types.ListNull(resourceBindingObjectType)
	}
	return listVal
}

func (r *WifConfigResource) populateSupportState(support *cmv1.WifSupport) types.Object {
	if support == nil || support.Empty() {
		return types.ObjectNull(supportAttrTypes)
	}
	roles := r.populateRolesState(support.Roles())
	obj, _ := types.ObjectValue(supportAttrTypes, map[string]attr.Value{
		"principal": types.StringValue(support.Principal()),
		"roles":     roles,
	})
	return obj
}

var (
	identityProviderAttrTypes = map[string]attr.Type{
		"identity_provider_id": types.StringType,
		"issuer_url":           types.StringType,
		"jwks":                 types.StringType,
		"allowed_audiences":    types.ListType{ElemType: types.StringType},
	}
	identityProviderObjectType = types.ObjectType{
		AttrTypes: identityProviderAttrTypes,
	}
	poolAttrTypes = map[string]attr.Type{
		"pool_id":           types.StringType,
		"identity_provider": identityProviderObjectType,
	}
	serviceAccountObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"service_account_id": types.StringType,
			"access_method":      types.StringType,
			"osd_role":           types.StringType,
			"roles":              types.ListType{ElemType: roleObjectType},
			"credential_request": credentialRequestObjectType,
		},
	}
	serviceAccountAttrTypes = map[string]attr.Type{
		"service_account_id": types.StringType,
		"access_method":      types.StringType,
		"osd_role":           types.StringType,
		"roles":              types.ListType{ElemType: roleObjectType},
		"credential_request": credentialRequestObjectType,
	}
	roleObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"role_id":           types.StringType,
			"predefined":        types.BoolType,
			"permissions":       types.ListType{ElemType: types.StringType},
			"resource_bindings": types.ListType{ElemType: resourceBindingObjectType},
		},
	}
	roleAttrTypes = map[string]attr.Type{
		"role_id":           types.StringType,
		"predefined":        types.BoolType,
		"permissions":       types.ListType{ElemType: types.StringType},
		"resource_bindings": types.ListType{ElemType: resourceBindingObjectType},
	}
	resourceBindingObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type": types.StringType,
			"name": types.StringType,
		},
	}
	resourceBindingAttrTypes = map[string]attr.Type{
		"type": types.StringType,
		"name": types.StringType,
	}
	credentialRequestObjectType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"namespace":             types.StringType,
			"service_account_names": types.ListType{ElemType: types.StringType},
		},
	}
	supportAttrTypes = map[string]attr.Type{
		"principal": types.StringType,
		"roles":     types.ListType{ElemType: roleObjectType},
	}
)
