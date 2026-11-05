// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplaneoperatorroles

import (
	"context"
	"fmt"
	"regexp"

	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// newIAMClientFunc resolves the AWS IAM client. Tests replace this to avoid
// AWS calls.
var newIAMClientFunc func(ctx context.Context) (IAMRoleClient, error) = NewAWSIAMClient

// SeaplaneOperatorRolesResource manages the full set of IAM operator roles
// for a ROSA Seaplane V2 (HCP) operator-roles prefix.
type SeaplaneOperatorRolesResource struct{}

var _ resource.Resource = &SeaplaneOperatorRolesResource{}
var _ resource.ResourceWithConfigure = &SeaplaneOperatorRolesResource{}
var _ resource.ResourceWithImportState = &SeaplaneOperatorRolesResource{}

// New creates the rhcs_seaplane_operator_roles resource.
func New() resource.Resource {
	return &SeaplaneOperatorRolesResource{}
}

func (r *SeaplaneOperatorRolesResource) Metadata(ctx context.Context,
	req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_seaplane_operator_roles"
}

func (r *SeaplaneOperatorRolesResource) Schema(ctx context.Context,
	req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the full set of IAM operator roles (and the worker node instance " +
			"role) required by a ROSA Seaplane V2 (HCP) cluster, for a given operator roles " +
			"prefix and OIDC issuer. Role definitions and trust policies are resolved from the " +
			"shared seaplane `iam` definition set, so no hand-written IAM JSON is needed. " +
			"On refresh the roles are reconciled against the resolved definition set: missing " +
			"policies are attached, drifted trust policies and tags are repaired, and roles are " +
			"tagged with the applied `definition_version`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The operator roles prefix (also used as the import key).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"operator_roles_prefix": schema.StringAttribute{
				Description: "Prefix prepended to every operator role name " +
					"(roles are named `<prefix>-<suffix>`). Changing this replaces the role set.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"oidc_issuer_url": schema.StringAttribute{
				Description: "OIDC issuer URL of the (managed or unmanaged) OIDC config the " +
					"operator role trust policies are bound to, e.g. the `issuer_url` of a " +
					"`rhcs_seaplane_oidc_config` resource. Must use an `https://` scheme.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^https://`),
						"the OIDC issuer URL must use an https:// scheme"),
				},
			},
			"permissions_boundary": schema.StringAttribute{
				Description: "ARN of the permissions boundary attached to every operator role. " +
					"Optional; changing it updates the boundary in place.",
				Optional: true,
			},
			"account_id": schema.StringAttribute{
				Description: "AWS account ID hosting the operator roles.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"partition": schema.StringAttribute{
				Description: "AWS partition the roles live in (`aws`, `aws-cn`, or `aws-us-gov`).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"definition_version": schema.StringAttribute{
				Description: "Version of the seaplane `iam` definition set applied to the roles. " +
					"Roles are tagged with this value for drift detection.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"instance_profile_name": schema.StringAttribute{
				Description: "Name of the worker node EC2 instance profile.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"roles": schema.ListNestedAttribute{
				Description: "The resolved operator role set, in definition order.",
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Full IAM role name (`<prefix>-<suffix>`).",
							Computed:    true,
						},
						"suffix": schema.StringAttribute{
							Description: "Role name suffix from the definition set.",
							Computed:    true,
						},
						"arn": schema.StringAttribute{
							Description: "IAM role ARN.",
							Computed:    true,
						},
						"is_worker_role": schema.BoolAttribute{
							Description: "True for the worker node instance role (EC2 trust).",
							Computed:    true,
						},
						"service_accounts": schema.ListAttribute{
							Description: "OIDC trust subjects of the role. Empty for the worker role.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"managed_policy_arns": schema.ListAttribute{
							Description: "AWS-managed policy ARNs attached to the role.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"trust_policy": schema.StringAttribute{
							Description: "The role's assume-role (trust) policy document.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (r *SeaplaneOperatorRolesResource) Configure(ctx context.Context,
	req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// This resource talks only to AWS (like `rosa create operator-roles` in
	// seaplane mode); no OCM or Platform API connection is required, so there
	// is nothing to extract from the provider data.
}

func (r *SeaplaneOperatorRolesResource) Create(ctx context.Context,
	req resource.CreateRequest, resp *resource.CreateResponse) {
	state := &ResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := newClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefix := state.OperatorRolesPrefix.ValueString()
	state.ID = types.StringValue(prefix)

	accountID, partition, _, err := client.CallerIdentity(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve AWS identity", err.Error())
		return
	}

	err = r.apply(ctx, client, state, accountID, partition, true)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to create operator roles for prefix %q", prefix),
			err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SeaplaneOperatorRolesResource) Read(ctx context.Context,
	req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &ResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := newClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefix := state.ID.ValueString()
	state.OperatorRolesPrefix = types.StringValue(prefix)

	accountID, partition, _, err := client.CallerIdentity(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve AWS identity", err.Error())
		return
	}

	// Reconciliation on refresh: converge the observed roles with the
	// resolved definition set (attach missing policies, repair trust policies
	// and tags, re-tag with the applied definition version).
	err = r.apply(ctx, client, state, accountID, partition, true)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to reconcile operator roles for prefix %q", prefix),
			err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SeaplaneOperatorRolesResource) Update(ctx context.Context,
	req resource.UpdateRequest, resp *resource.UpdateResponse) {
	state := &ResourceModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := newClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID, partition, _, err := client.CallerIdentity(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve AWS identity", err.Error())
		return
	}

	// The prefix is RequiresReplace; reachable updates are the OIDC issuer
	// (trust-policy rewrite) and the permissions boundary.
	err = r.apply(ctx, client, state, accountID, partition, true)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to update operator roles for prefix %q", state.ID.ValueString()),
			err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *SeaplaneOperatorRolesResource) Delete(ctx context.Context,
	req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &ResourceModel{}
	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := newClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefix := state.ID.ValueString()
	accountID, partition, _, err := client.CallerIdentity(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve AWS identity", err.Error())
		return
	}

	desired, err := resolveDesired(prefix, accountID, partition, state.OIDCIssuerURL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to resolve operator role definitions for prefix %q", prefix),
			err.Error())
		return
	}

	for _, role := range desired {
		observed, err := client.GetRole(ctx, role.Name)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Failed to read IAM role %q", role.Name), err.Error())
			return
		}
		if observed == nil {
			continue
		}
		// Never delete a role that belongs to a different prefix; a name
		// collision means the role was not created by this resource.
		if tagged, ok := observed.Tags[rolePrefixTag]; ok && tagged != prefix {
			continue
		}
		if err := r.deleteRole(ctx, client, role, observed); err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Failed to delete IAM role %q", role.Name), err.Error())
			return
		}
	}
}

func (r *SeaplaneOperatorRolesResource) ImportState(ctx context.Context,
	req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// deleteRole removes one operator role: it deletes the worker instance
// profile, detaches only the policies owned by the definition set (never
// policies attached out of band), and deletes the role.
func (r *SeaplaneOperatorRolesResource) deleteRole(ctx context.Context, client IAMRoleClient,
	role desiredRole, observed *RoleSummary) error {
	if role.IsWorkerRole {
		if err := client.DeleteInstanceProfile(ctx, role.Name, role.Name); err != nil {
			return err
		}
	}
	attached := make(map[string]bool, len(observed.AttachedPolicies))
	for _, policy := range observed.AttachedPolicies {
		attached[policy.ARN] = true
	}
	for _, policyARN := range role.PolicyARNs {
		if attached[policyARN] {
			if err := client.DetachPolicy(ctx, role.Name, policyARN); err != nil {
				return err
			}
		}
	}
	return client.DeleteRole(ctx, role.Name)
}

// apply converges the AWS state with the resolved desired role set and
func (r *SeaplaneOperatorRolesResource) apply(ctx context.Context, client IAMRoleClient,
	state *ResourceModel, accountID, partition string, logDetails bool) error {
	prefix := state.OperatorRolesPrefix.ValueString()
	issuerURL := state.OIDCIssuerURL.ValueString()
	permissionsBoundary := state.PermissionsBoundary.ValueString()

	desired, err := resolveDesired(prefix, accountID, partition, issuerURL)
	if err != nil {
		return err
	}

	roles := make([]RoleModel, 0, len(desired))
	instanceProfileName := ""
	for _, role := range desired {
		observed, err := client.GetRole(ctx, role.Name)
		if err != nil {
			return err
		}

		actions := planReconcile(role, observed, prefix)
		if actions.ForeignRole {
			return fmt.Errorf(
				"IAM role %q already exists but is tagged with operator roles prefix %q; "+
					"refusing to manage a role that belongs to another prefix",
				role.Name, observed.Tags[rolePrefixTag])
		}

		arn := ""
		if observed != nil {
			arn = observed.ARN
		}
		if actions.CreateRole {
			tags := iamRoleTags(prefix)
			arn, err = client.CreateRole(ctx, role.Name, role.TrustPolicy, permissionsBoundary, tags)
			if err != nil {
				return err
			}
			if logDetails {
				tflog.Info(ctx, fmt.Sprintf("created IAM role %s", role.Name))
			}
		} else {
			if actions.UpdateTrustPolicy {
				if err := client.UpdateTrustPolicy(ctx, role.Name, role.TrustPolicy); err != nil {
					return err
				}
				if logDetails {
					tflog.Info(ctx, fmt.Sprintf("updated trust policy of IAM role %s", role.Name))
				}
			}
			// Keep the permissions boundary in sync with the configuration.
			if observed.PermissionsBound != permissionsBoundary {
				if err := client.UpdatePermissionsBoundary(ctx, role.Name, permissionsBoundary); err != nil {
					return err
				}
			}
			if len(actions.UpdateTags) > 0 {
				if err := client.TagRole(ctx, role.Name, actions.UpdateTags); err != nil {
					return err
				}
			}
		}
		for _, policyARN := range actions.AttachPolicies {
			if err := client.AttachPolicy(ctx, role.Name, policyARN); err != nil {
				return err
			}
			if logDetails {
				tflog.Info(ctx, fmt.Sprintf("attached policy %s to IAM role %s", policyARN, role.Name))
			}
		}
		if actions.EnsureInstanceProfile || actions.CreateRole && role.IsWorkerRole {
			if err := client.EnsureInstanceProfile(ctx, role.Name, role.Name, iamRoleTags(prefix)); err != nil {
				return err
			}
		}
		if role.IsWorkerRole {
			instanceProfileName = role.Name
		}

		roles = append(roles, RoleModel{
			Name:              types.StringValue(role.Name),
			Suffix:            types.StringValue(role.Spec.Name),
			ARN:               types.StringValue(arn),
			IsWorkerRole:      types.BoolValue(role.IsWorkerRole),
			ServiceAccounts:   stringList(role.Spec.ServiceAccounts),
			ManagedPolicyARNs: stringList(role.PolicyARNs),
			TrustPolicy:       types.StringValue(role.TrustPolicy),
		})
	}

	state.ID = types.StringValue(prefix)
	state.AccountID = types.StringValue(accountID)
	state.Partition = types.StringValue(partition)
	state.DefinitionVersion = types.StringValue(iamDefinitionVersion())
	state.InstanceProfileName = types.StringValue(instanceProfileName)
	state.Roles = roles
	return nil
}

// newClient resolves the AWS IAM client.
func newClient(ctx context.Context) (IAMRoleClient, diag.Diagnostics) {
	client, err := newIAMClientFunc(ctx)
	if err != nil {
		var diags diag.Diagnostics
		diags.AddError("Failed to create the AWS IAM client",
			fmt.Sprintf("%v - hint: verify the AWS credentials configured for the provider", err))
		return nil, diags
	}
	return client, nil
}
