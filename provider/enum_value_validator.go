package provider

import (
	"context"
	"fmt"
	"github.com/thoas/go-funk"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
)

func EnumValueValidator(allowedList []string) []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate enum param",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {

				value := &types.String{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, value)
				fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAA", common.IsStringAttributeEmpty(*value), *value)
				if diag.HasError() || common.IsStringAttributeEmpty(*value) {
					// No attribute to validate
					return
				}

				if funk.Contains(allowedList, value.Value) {
					return
				}

				resp.Diagnostics.AddError(fmt.Sprintf("Invalid %s.", req.AttributePath.LastStep()),
					fmt.Sprintf("Expected a valid %s param. Options are %s. Got %s.",
						req.AttributePath.LastStep(), allowedList, value.Value),
				)
			},
		},
	}
}
