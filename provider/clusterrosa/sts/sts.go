// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package sts

import (
	"context"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/openshift-online/ocm-common/pkg/aws/ststrust"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

const externalIdDescription = "External ID for trust policy condition in account roles"

type baseSts struct {
	OIDCEndpointURL       types.String `tfsdk:"oidc_endpoint_url"`
	OIDCConfigID          types.String `tfsdk:"oidc_config_id"`
	Thumbprint            types.String `tfsdk:"thumbprint"`
	RoleARN               types.String `tfsdk:"role_arn"`
	SupportRoleArn        types.String `tfsdk:"support_role_arn"`
	OperatorRolePrefix    types.String `tfsdk:"operator_role_prefix"`
	TrustPolicyExternalID types.String `tfsdk:"trust_policy_external_id"`
}

type ClassicSts struct {
	OIDCEndpointURL       types.String           `tfsdk:"oidc_endpoint_url"`
	OIDCConfigID          types.String           `tfsdk:"oidc_config_id"`
	Thumbprint            types.String           `tfsdk:"thumbprint"`
	RoleARN               types.String           `tfsdk:"role_arn"`
	SupportRoleArn        types.String           `tfsdk:"support_role_arn"`
	OperatorRolePrefix    types.String           `tfsdk:"operator_role_prefix"`
	TrustPolicyExternalID types.String           `tfsdk:"trust_policy_external_id"`
	InstanceIAMRoles      classicInstanceIAMRole `tfsdk:"instance_iam_roles"`
}

type HcpSts struct {
	OIDCEndpointURL       types.String       `tfsdk:"oidc_endpoint_url"`
	OIDCConfigID          types.String       `tfsdk:"oidc_config_id"`
	Thumbprint            types.String       `tfsdk:"thumbprint"`
	RoleARN               types.String       `tfsdk:"role_arn"`
	SupportRoleArn        types.String       `tfsdk:"support_role_arn"`
	OperatorRolePrefix    types.String       `tfsdk:"operator_role_prefix"`
	TrustPolicyExternalID types.String       `tfsdk:"trust_policy_external_id"`
	InstanceIAMRoles      hcpInstanceIAMRole `tfsdk:"instance_iam_roles"`
}

type classicInstanceIAMRole struct {
	WorkerRoleARN types.String `tfsdk:"worker_role_arn"`
	MasterRoleARN types.String `tfsdk:"master_role_arn"`
}

type hcpInstanceIAMRole struct {
	WorkerRoleARN types.String `tfsdk:"worker_role_arn"`
}

// ClassicStsResource returns the STS nested block schema for ROSA Classic cluster resources.
func ClassicStsResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"oidc_endpoint_url": schema.StringAttribute{
			Description: "OIDC Endpoint URL",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				// This passes the state through to the plan, so that the OIDC
				// provider URL will not be "unknown" at plan-time, resulting in
				// the oidc provider being needlessly replaced. This should be
				// OK, since the OIDC provider URL is not expected to change.
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"oidc_config_id": schema.StringAttribute{
			Description: "OIDC Configuration ID",
			Optional:    true,
		},
		"thumbprint": schema.StringAttribute{
			Description: "SHA1-hash value of the root CA of the issuer URL",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				// This passes the state through to the plan, so that the OIDC
				// thumbprint will not be "unknown" at plan-time, resulting in
				// the oidc provider being needlessly replaced. This should be
				// OK, since the thumbprint is not expected to change.
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"role_arn": schema.StringAttribute{
			Description: "Installer Role",
			Required:    true,
		},
		"support_role_arn": schema.StringAttribute{
			Description: "Support Role",
			Required:    true,
		},
		"instance_iam_roles": schema.SingleNestedAttribute{
			Description: "Instance IAM Roles",
			Attributes: map[string]schema.Attribute{
				"master_role_arn": schema.StringAttribute{
					Description: "Master/Control Plane Node Role ARN",
					Required:    true,
				},
				"worker_role_arn": schema.StringAttribute{
					Description: "Worker/Compute Node Role ARN",
					Required:    true,
				},
			},
			Required: true,
		},
		"operator_role_prefix": schema.StringAttribute{
			Description: "Operator IAM Role prefix",
			Required:    true,
		},
		"trust_policy_external_id": schema.StringAttribute{
			Description: externalIdDescription,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: trustPolicyExternalIDValidators(),
		},
	}
}

// ClassicStsDatasource returns the STS nested block schema for ROSA Classic cluster data sources.
func ClassicStsDatasource() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"oidc_endpoint_url": schema.StringAttribute{
			Description: "OIDC Endpoint URL",
			Computed:    true,
		},
		"oidc_config_id": schema.StringAttribute{
			Description: "OIDC Configuration ID",
			Computed:    true,
		},
		"thumbprint": schema.StringAttribute{
			Description: "SHA1-hash value of the root CA of the issuer URL",
			Computed:    true,
		},
		"role_arn": schema.StringAttribute{
			Description: "Installer Role",
			Computed:    true,
		},
		"support_role_arn": schema.StringAttribute{
			Description: "Support Role",
			Computed:    true,
		},
		"instance_iam_roles": schema.SingleNestedAttribute{
			Description: "Instance IAM Roles",
			Attributes: map[string]schema.Attribute{
				"master_role_arn": schema.StringAttribute{
					Description: "Master/Control Plane Node Role ARN",
					Computed:    true,
				},
				"worker_role_arn": schema.StringAttribute{
					Description: "Worker/Compute Node Role ARN",
					Computed:    true,
				},
			},
			Computed: true,
		},
		"operator_role_prefix": schema.StringAttribute{
			Description: "Operator IAM Role prefix",
			Computed:    true,
		},
		"trust_policy_external_id": dsschema.StringAttribute{
			Description: externalIdDescription,
			Computed:    true,
		},
	}
}

// HcpStsResource returns the STS nested block schema for ROSA HCP cluster resources.
func HcpStsResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"oidc_endpoint_url": schema.StringAttribute{
			Description: "OIDC Endpoint URL",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				// This passes the state through to the plan, so that the OIDC
				// provider URL will not be "unknown" at plan-time, resulting in
				// the oidc provider being needlessly replaced. This should be
				// OK, since the OIDC provider URL is not expected to change.
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"oidc_config_id": schema.StringAttribute{
			Description: "OIDC Configuration ID",
			Optional:    true,
		},
		"thumbprint": schema.StringAttribute{
			Description: "SHA1-hash value of the root CA of the issuer URL",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				// This passes the state through to the plan, so that the OIDC
				// thumbprint will not be "unknown" at plan-time, resulting in
				// the oidc provider being needlessly replaced. This should be
				// OK, since the thumbprint is not expected to change.
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"role_arn": schema.StringAttribute{
			Description: "Installer Role",
			Required:    true,
		},
		"support_role_arn": schema.StringAttribute{
			Description: "Support Role",
			Required:    true,
		},
		"instance_iam_roles": schema.SingleNestedAttribute{
			Description: "Instance IAM Roles",
			Attributes: map[string]schema.Attribute{
				"worker_role_arn": schema.StringAttribute{
					Description: "Worker/Compute Node Role ARN",
					Required:    true,
				},
			},
			Required: true,
		},
		"operator_role_prefix": schema.StringAttribute{
			Description: "Operator IAM Role prefix",
			Required:    true,
		},
		"trust_policy_external_id": schema.StringAttribute{
			Description: externalIdDescription,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
			Validators: trustPolicyExternalIDValidators(),
		},
	}
}

// HcpStsDatasource returns the STS nested block schema for ROSA HCP cluster data sources.
func HcpStsDatasource() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"oidc_endpoint_url": schema.StringAttribute{
			Description: "OIDC Endpoint URL",
			Computed:    true,
		},
		"oidc_config_id": schema.StringAttribute{
			Description: "OIDC Configuration ID",
			Computed:    true,
		},
		"thumbprint": schema.StringAttribute{
			Description: "SHA1-hash value of the root CA of the issuer URL",
			Computed:    true,
		},
		"role_arn": schema.StringAttribute{
			Description: "Installer Role",
			Computed:    true,
		},
		"support_role_arn": schema.StringAttribute{
			Description: "Support Role",
			Computed:    true,
		},
		"instance_iam_roles": schema.SingleNestedAttribute{
			Description: "Instance IAM Roles",
			Attributes: map[string]schema.Attribute{
				"worker_role_arn": schema.StringAttribute{
					Description: "Worker/Compute Node Role ARN",
					Computed:    true,
				},
			},
			Computed: true,
		},
		"operator_role_prefix": schema.StringAttribute{
			Description: "Operator IAM Role prefix",
			Computed:    true,
		},
		"trust_policy_external_id": dsschema.StringAttribute{
			Description: externalIdDescription,
			Computed:    true,
		},
	}
}

// trustPolicyExternalIDValidators returns plan-time validators for sts.trust_policy_external_id format.
func trustPolicyExternalIDValidators() []validator.String {
	return []validator.String{
		attrvalidators.NewStringValidator(
			"Must be a valid STS external ID for trust policy conditions.",
			func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
				if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
					return
				}
				if err := ststrust.ValidateSTSExternalIDFormat(req.ConfigValue.ValueString()); err != nil {
					resp.Diagnostics.AddAttributeError(req.Path, "Invalid trust_policy_external_id", err.Error())
				}
			},
		),
	}
}

// ValidateTrustPolicyExternalIDFromConfig validates trust_policy_external_id using Terraform config values.
// Both Classic and HCP cluster Create functions delegate to this helper.
func ValidateTrustPolicyExternalIDFromConfig(
	ctx context.Context,
	trustPolicyExternalID, roleARN, supportRoleARN types.String,
	region string,
) error {
	entered := ""
	if !trustPolicyExternalID.IsUnknown() && !trustPolicyExternalID.IsNull() {
		entered = trustPolicyExternalID.ValueString()
	}
	return ValidateTrustPolicyExternalID(ctx, entered, roleARN.ValueString(), supportRoleARN.ValueString(), region)
}

// stsExternalIDSource reads the STS external ID from an OCM cluster STS object.
type stsExternalIDSource interface {
	GetExternalID() (string, bool)
}

// PopulateTrustPolicyExternalIDFromSTS sets trust_policy_external_id from the OCM STS object.
func PopulateTrustPolicyExternalIDFromSTS(stsState stsExternalIDSource, target *types.String) {
	if stsState == nil {
		*target = types.StringNull()
		return
	}
	if externalID, ok := stsState.GetExternalID(); ok && externalID != "" {
		*target = types.StringValue(externalID)
		return
	}
	*target = types.StringNull()
}
