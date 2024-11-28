package sharedvpc

import (
	"context"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

type SharedVpc struct {
	IngressPrivateHostedZoneId               types.String `tfsdk:"ingress_private_hosted_zone_id"`
	InternalCommunicationPrivateHostedZoneId types.String `tfsdk:"internal_communication_private_hosted_zone_id"`
	Route53RoleArn                           types.String `tfsdk:"route53_role_arn"`
	VpceRoleArn                              types.String `tfsdk:"vpce_role_arn"`
}

func SharedVpcResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"ingress_private_hosted_zone_id": schema.StringAttribute{
			//nolint:lll
			Description: "ID assigned by AWS to private Route 53 hosted zone associated with intended shared VPC, e.g. 'Z05646003S02O1ENCDCSN'.",
			Required:    true,
		},
		"internal_communication_private_hosted_zone_id": schema.StringAttribute{
			//noling:lll
			Description: "ID assigned by AWS to private Route 53 hosted zone associated with intended shared VPC, e.g. 'Z05646003S02O1ENCDCSN'.",
			Optional:    true,
		},
		"route53_role_arn": schema.StringAttribute{
			//nolint:lll
			Description: "AWS IAM role ARN with a policy attached, granting permissions necessary to create and manage Route 53 DNS records in private Route 53 hosted zone associated with intended shared VPC.",
			Required:    true,
		},
		"vpce_role_arn": schema.StringAttribute{
			//nolint:lll
			Description: "AWS IAM role ARN with a policy attached, granting permissions necessary to create and manage VPC Endpoints associated with intended shared VPC.",
			Required:    true,
		},
	}
}

func HcpStsDatasource() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"ingress_private_hosted_zone_id": schema.StringAttribute{
			//nolint:lll
			Description: "ID assigned by AWS to private Route 53 hosted zone associated with intended shared VPC, e.g. 'Z05646003S02O1ENCDCSN'.",
			Computed:    true,
		},
		"internal_communication_private_hosted_zone_id": schema.StringAttribute{
			//noling:lll
			Description: "ID assigned by AWS to private Route 53 hosted zone associated with intended shared VPC, e.g. 'Z05646003S02O1ENCDCSN'.",
			Computed:    true,
		},
		"route53_role_arn": schema.StringAttribute{
			//nolint:lll
			Description: "AWS IAM role ARN with a policy attached, granting permissions necessary to create and manage Route 53 DNS records in private Route 53 hosted zone associated with intended shared VPC.",
			Computed:    true,
		},
		"vpce_role_arn": schema.StringAttribute{
			//nolint:lll
			Description: "AWS IAM role ARN with a policy attached, granting permissions necessary to create and manage VPC Endpoints associated with intended shared VPC.",
			Computed:    true,
		},
	}
}

var HcpSharedVpcValidator = attrvalidators.NewObjectValidator("Shared VPC attribute must include all attributes",
	func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		sharedVpc := SharedVpc{}
		d := req.ConfigValue.As(ctx, &sharedVpc, basetypes.ObjectAsOptions{})
		if d.HasError() {
			// No attribute to validate
			return
		}
		errSum := "Invalid shared_vpc attribute assignment"

		// validate ID and ARN are not empty
		valuesToCheck := []basetypes.StringValue{
			sharedVpc.IngressPrivateHostedZoneId,
			sharedVpc.InternalCommunicationPrivateHostedZoneId,
			sharedVpc.Route53RoleArn,
			sharedVpc.VpceRoleArn,
		}
		for _, value := range valuesToCheck {
			if common.IsStringAttributeKnownAndEmpty(value) {
				resp.Diagnostics.AddError(errSum, "Invalid configuration, all attributes are required")
				return
			}
		}
	})
