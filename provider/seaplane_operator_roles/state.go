// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplaneoperatorroles

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RoleModel is the per-role state of the rhcs_seaplane_operator_roles
// resource and data source.
type RoleModel struct {
	// Name is the full IAM role name ("<prefix>-<suffix>").
	Name types.String `tfsdk:"name"`
	// Suffix is the role name suffix from the definition set (e.g. "ingress").
	Suffix types.String `tfsdk:"suffix"`
	// ARN is the IAM role ARN.
	ARN types.String `tfsdk:"arn"`
	// IsWorkerRole is true for the worker node instance role (EC2 trust).
	IsWorkerRole types.Bool `tfsdk:"is_worker_role"`
	// ServiceAccounts are the OIDC trust subjects. Empty for the worker role.
	ServiceAccounts []types.String `tfsdk:"service_accounts"`
	// ManagedPolicyARNs are the AWS-managed policy ARNs attached to the role.
	ManagedPolicyARNs []types.String `tfsdk:"managed_policy_arns"`
	// TrustPolicy is the role's assume-role policy document.
	TrustPolicy types.String `tfsdk:"trust_policy"`
}

// ResourceModel is the state of the rhcs_seaplane_operator_roles resource.
type ResourceModel struct {
	// ID is the operator roles prefix (also the import key).
	ID types.String `tfsdk:"id"`
	// OperatorRolesPrefix is the user-chosen role name prefix.
	OperatorRolesPrefix types.String `tfsdk:"operator_roles_prefix"`
	// OIDCIssuerURL is the OIDC issuer URL the operator role trust policies
	// are bound to.
	OIDCIssuerURL types.String `tfsdk:"oidc_issuer_url"`
	// PermissionsBoundary is the optional ARN of the permissions boundary
	// attached to every role.
	PermissionsBoundary types.String `tfsdk:"permissions_boundary"`
	// AccountID is the AWS account hosting the roles.
	AccountID types.String `tfsdk:"account_id"`
	// Partition is the AWS partition ("aws", "aws-cn", "aws-us-gov").
	Partition types.String `tfsdk:"partition"`
	// DefinitionVersion is the iam/ definition-set revision applied to the roles.
	DefinitionVersion types.String `tfsdk:"definition_version"`
	// InstanceProfileName is the worker node instance profile name.
	InstanceProfileName types.String `tfsdk:"instance_profile_name"`
	// Roles is the resolved role set, in definition order.
	Roles []RoleModel `tfsdk:"roles"`
}

// DataSourceModel is the state of the rhcs_seaplane_operator_roles data source.
type DataSourceModel struct {
	// ID is the operator roles prefix.
	ID types.String `tfsdk:"id"`
	// OperatorRolesPrefix is the prefix used to discover the roles.
	OperatorRolesPrefix types.String `tfsdk:"operator_roles_prefix"`
	// AccountID is the AWS account hosting the roles.
	AccountID types.String `tfsdk:"account_id"`
	// Partition is the AWS partition.
	Partition types.String `tfsdk:"partition"`
	// DefinitionVersion is the definition-version tag found on the roles
	// (empty when the roles predate version tagging).
	DefinitionVersion types.String `tfsdk:"definition_version"`
	// InstanceProfileName is the worker node instance profile name (empty when
	// no worker role is found).
	InstanceProfileName types.String `tfsdk:"instance_profile_name"`
	// Roles is the discovered role set.
	Roles []RoleModel `tfsdk:"roles"`
}
