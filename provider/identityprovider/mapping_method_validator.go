package identityprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

var validMappingMethods = []string{"claim", "add", "generate", "lookup"} // Default is @ index 0
var DefaultMappingMethod = validMappingMethods[0]

func MappingMethodValidators() []validator.String {
	desc := "validate mapping method"
	return []validator.String{
		attrvalidators.NewStringValidator(desc, func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
			mappingMethod := &types.String{}
			diag := req.Config.GetAttribute(ctx, req.Path, mappingMethod)
			if diag.HasError() || mappingMethod.IsUnknown() || mappingMethod.IsNull() {
				// No attribute to validate
				return
			}
			isValidMappingMethod := false
			for _, validMappingMethod := range validMappingMethods {
				if mappingMethod.ValueString() == validMappingMethod {
					isValidMappingMethod = true
				}
			}
			if !isValidMappingMethod {
				resp.Diagnostics.AddError("Invalid mapping_method.",
					fmt.Sprintf("Expected a valid mapping method. Options are %s. Got %s", validMappingMethods, mappingMethod.ValueString()),
				)
			}
		}),
	}
}
