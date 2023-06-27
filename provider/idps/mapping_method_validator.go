package idps

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-red-hat-cloud-services/provider/common"
)

var validMappingMethods = []string{"claim", "add", "generate", "lookup"} // Default is @ index 0
var DefaultMappingMethod = validMappingMethods[0]

func MappingMethodValidators() []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate mapping_method",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {

				mappingMethod := &types.String{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, mappingMethod)
				if diag.HasError() || mappingMethod.Unknown || mappingMethod.Null {
					// No attribute to validate
					return
				}
				isValidMappingMethod := false
				for _, validMappingMethod := range validMappingMethods {
					if mappingMethod.Value == validMappingMethod {
						isValidMappingMethod = true
					}
				}
				if !isValidMappingMethod {
					resp.Diagnostics.AddError("Invalid mapping_method.",
						fmt.Sprintf("Expected a valid mapping method. Options are %s. Got %s", validMappingMethods, mappingMethod.Value),
					)
				}
			},
		},
	}
}
