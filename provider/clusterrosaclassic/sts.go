package clusterrosaclassic

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func stsResource() map[string]schema.Attribute {
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

func stsDatasource() map[string]dsschema.Attribute {
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
