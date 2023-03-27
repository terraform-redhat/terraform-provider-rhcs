package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func oidcConfigResource() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"id": {
			Description: "ID",
			Type:        types.StringType,
			Optional:    true,
		},
		"issuer_url": {
			Description: "Issuer URL",
			Type:        types.StringType,
			Optional:    true,
		},
		"secret_arn": {
			Description: "Secret ARN for private key",
			Type:        types.StringType,
			Optional:    true,
		},
		"managed": {
			Description: "Managed indicator",
			Type:        types.BoolType,
			Optional:    true,
		},
		"reusable": {
			Description: "Reusable indicator",
			Type:        types.BoolType,
			Optional:    true,
		},
	})
}
