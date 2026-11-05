// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplaneoperatorroles

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/openshift-online/rosa-seaplane-api/iam"
)

// SeaplaneOperatorRolesDataSource discovers the operator roles that exist in
// the AWS account for an operator-roles prefix. It can be used to import an
// existing role set into a rhcs_seaplane_operator_roles resource, or to
// inspect roles created by the rosa CLI.
type SeaplaneOperatorRolesDataSource struct{}

var _ datasource.DataSource = &SeaplaneOperatorRolesDataSource{}
var _ datasource.DataSourceWithConfigure = &SeaplaneOperatorRolesDataSource{}

// NewDataSource creates the rhcs_seaplane_operator_roles data source.
func NewDataSource() datasource.DataSource {
	return &SeaplaneOperatorRolesDataSource{}
}

func (s *SeaplaneOperatorRolesDataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_seaplane_operator_roles"
}

func (s *SeaplaneOperatorRolesDataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Discovers the IAM operator roles that exist in the AWS account for a " +
			"ROSA Seaplane V2 (HCP) operator-roles prefix, with their ARNs, attached " +
			"AWS-managed policies, tags, and applied definition version. Useful for " +
			"importing an existing role set (e.g. one created by the rosa CLI) into a " +
			"`rhcs_seaplane_operator_roles` resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The operator roles prefix.",
				Computed:    true,
			},
			"operator_roles_prefix": schema.StringAttribute{
				Description: "Prefix used to discover the operator roles " +
					"(roles are named `<prefix>-<suffix>`).",
				Required: true,
			},
			"account_id": schema.StringAttribute{
				Description: "AWS account ID hosting the operator roles.",
				Computed:    true,
			},
			"partition": schema.StringAttribute{
				Description: "AWS partition the roles live in (`aws`, `aws-cn`, or `aws-us-gov`).",
				Computed:    true,
			},
			"definition_version": schema.StringAttribute{
				Description: "Definition-version tag found on the discovered roles " +
					"(empty when the roles predate version tagging).",
				Computed: true,
			},
			"instance_profile_name": schema.StringAttribute{
				Description: "Name of the worker node EC2 instance profile " +
					"(empty when no worker role is found).",
				Computed: true,
			},
			"roles": schema.ListNestedAttribute{
				Description: "The discovered operator roles.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Full IAM role name (`<prefix>-<suffix>`).",
							Computed:    true,
						},
						"suffix": schema.StringAttribute{
							Description: "Role name suffix (the part after the prefix).",
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
							Description: "OIDC trust subjects declared by the definition set for this role.",
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

func (s *SeaplaneOperatorRolesDataSource) Configure(ctx context.Context,
	req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// This data source talks only to AWS; no provider connection is needed.
}

func (s *SeaplaneOperatorRolesDataSource) Read(ctx context.Context,
	req datasource.ReadRequest, resp *datasource.ReadResponse) {
	state := &DataSourceModel{}
	resp.Diagnostics.Append(req.Config.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := newClient(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefix := state.OperatorRolesPrefix.ValueString()

	accountID, partition, _, err := client.CallerIdentity(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve AWS identity", err.Error())
		return
	}

	all, err := client.ListRolesByPrefix(ctx, prefix)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to list IAM roles for prefix %q", prefix), err.Error())
		return
	}

	// Only keep roles that either carry the rosa_role_prefix tag for this
	// prefix or were created by this definition set's naming convention.
	definitions := definitionSuffixes()
	roles := make([]RoleModel, 0, len(definitions))
	instanceProfileName := ""
	definitionVersion := ""
	for _, observed := range all {
		tagged, hasTag := observed.Tags[rolePrefixTag]
		suffix, knownSuffix := suffixAfterPrefix(observed.Name, prefix)
		if hasTag && tagged != prefix {
			continue
		}
		if !hasTag && !knownSuffix {
			continue
		}
		if knownSuffix {
			spec := definitions[suffix]
			policies := make([]string, 0, len(observed.AttachedPolicies))
			for _, policy := range observed.AttachedPolicies {
				policies = append(policies, policy.ARN)
			}
			if spec.IsWorkerRole {
				instanceProfileName = observed.Name
			}
			if v, ok := observed.Tags[iam.TagDefinitionVersion]; ok {
				definitionVersion = v
			}
			roles = append(roles, RoleModel{
				Name:              types.StringValue(observed.Name),
				Suffix:            types.StringValue(suffix),
				ARN:               types.StringValue(observed.ARN),
				IsWorkerRole:      types.BoolValue(spec.IsWorkerRole),
				ServiceAccounts:   stringList(spec.ServiceAccounts),
				ManagedPolicyARNs: stringList(policies),
				TrustPolicy:       types.StringValue(observed.TrustPolicy),
			})
		}
	}

	if len(roles) == 0 {
		description := fmt.Sprintf("No operator roles found for prefix %q", prefix)
		tflog.Error(ctx, description)
		resp.Diagnostics.AddError(
			description,
			"hint: verify the operator roles prefix and the AWS account, "+
				"or run 'rosa list operator-roles' to discover existing prefixes",
		)
		return
	}

	state.ID = types.StringValue(prefix)
	state.AccountID = types.StringValue(accountID)
	state.Partition = types.StringValue(partition)
	state.DefinitionVersion = types.StringValue(definitionVersion)
	state.InstanceProfileName = types.StringValue(instanceProfileName)
	state.Roles = roles

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// definitionSuffixes maps role name suffixes to their definitions, for
// classifying discovered roles.
func definitionSuffixes() map[string]roleDefinition {
	set := defaultDefinitionSet()
	result := make(map[string]roleDefinition, len(set.Roles))
	for _, spec := range set.Roles {
		result[spec.Name] = spec
	}
	return result
}

// suffixAfterPrefix splits a role name into the prefix and suffix. It returns
// ok=false when the name does not start with "<prefix>-".
func suffixAfterPrefix(name, prefix string) (suffix string, ok bool) {
	rest, found := strings.CutPrefix(name, prefix+"-")
	if !found {
		return "", false
	}
	return rest, true
}
