package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/thoas/go-funk"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

func EnumValueValidator(allowedList []string) []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate enum param",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {

				value := &types.String{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, value)
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

func FloatRangeValidators(description string, min float64, max float64) []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: description,
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				attribute := &types.String{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, attribute)

				if diag.HasError() {
					// No attribute to validate
					return
				}

				if attribute.Null || attribute.Unknown {
					// No need to validate
					return
				}

				number, err := strconv.ParseFloat(attribute.Value, 64)
				if err != nil {
					resp.Diagnostics.AddAttributeError(
						req.AttributePath,
						"Value cannot be parsed to a float value",
						fmt.Sprintf("Value '%s' cannot be parsed to a float value", attribute.Value),
					)
					return
				}

				if number < min || number > max {
					resp.Diagnostics.AddAttributeError(
						req.AttributePath,
						"Value out of range",
						fmt.Sprintf("Value '%f' is out of range %f - %f", number, min, max),
					)
				}
			},
		},
	}
}

func DurationStringValidators(description string) []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: description,
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
				attribute := &types.String{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, attribute)

				if diag.HasError() {
					// No attribute to validate
					return
				}

				if attribute.Null || attribute.Unknown {
					// No need to validate
					return
				}

				if _, err := time.ParseDuration(attribute.Value); err != nil {
					resp.Diagnostics.AddAttributeError(
						req.AttributePath,
						"Value cannot be parsed to a duration string",
						fmt.Sprintf("Value '%s' cannot be parsed to a duration string. A duration "+
							"string is a sequence of decimal numbers and a time unit suffix such as \"300m\", "+
							"\"1.5h\" or \"2h45m\"",
							attribute.Value),
					)
					return
				}
			},
		},
	}
}
