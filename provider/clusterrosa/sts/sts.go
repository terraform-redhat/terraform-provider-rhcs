package sts

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type baseSts struct {
	OIDCEndpointURL    types.String `tfsdk:"oidc_endpoint_url"`
	OIDCConfigID       types.String `tfsdk:"oidc_config_id"`
	Thumbprint         types.String `tfsdk:"thumbprint"`
	RoleARN            types.String `tfsdk:"role_arn"`
	SupportRoleArn     types.String `tfsdk:"support_role_arn"`
	OperatorRolePrefix types.String `tfsdk:"operator_role_prefix"`
}

type ClassicSts struct {
	OIDCEndpointURL    types.String           `tfsdk:"oidc_endpoint_url"`
	OIDCConfigID       types.String           `tfsdk:"oidc_config_id"`
	Thumbprint         types.String           `tfsdk:"thumbprint"`
	RoleARN            types.String           `tfsdk:"role_arn"`
	SupportRoleArn     types.String           `tfsdk:"support_role_arn"`
	OperatorRolePrefix types.String           `tfsdk:"operator_role_prefix"`
	InstanceIAMRoles   classicInstanceIAMRole `tfsdk:"instance_iam_roles"`
}

type HcpSts struct {
	OIDCEndpointURL    types.String       `tfsdk:"oidc_endpoint_url"`
	OIDCConfigID       types.String       `tfsdk:"oidc_config_id"`
	Thumbprint         types.String       `tfsdk:"thumbprint"`
	RoleARN            types.String       `tfsdk:"role_arn"`
	SupportRoleArn     types.String       `tfsdk:"support_role_arn"`
	OperatorRolePrefix types.String       `tfsdk:"operator_role_prefix"`
	InstanceIAMRoles   hcpInstanceIAMRole `tfsdk:"instance_iam_roles"`
}

type baseInstanceIAMRole struct {
	WorkerRoleARN types.String `tfsdk:"worker_role_arn"`
}

type classicInstanceIAMRole struct {
	WorkerRoleARN types.String `tfsdk:"worker_role_arn"`
	MasterRoleARN types.String `tfsdk:"master_role_arn"`
}

type hcpInstanceIAMRole struct {
	WorkerRoleARN types.String `tfsdk:"worker_role_arn"`
}

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
	}
}

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
	}
}

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
	}
}

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
	}
}
