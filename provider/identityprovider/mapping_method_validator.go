package identityprovider

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
***REMOVED***

var validMappingMethods = []string{"claim", "add", "generate", "lookup"} // Default is @ index 0
var DefaultMappingMethod = validMappingMethods[0]

func MappingMethodValidators(***REMOVED*** []validator.String {
	desc := "validate mapping method"
	return []validator.String{
		attrvalidators.NewStringValidator(desc, func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
			mappingMethod := &types.String{}
			diag := req.Config.GetAttribute(ctx, req.Path, mappingMethod***REMOVED***
			if diag.HasError(***REMOVED*** || mappingMethod.IsUnknown(***REMOVED*** || mappingMethod.IsNull(***REMOVED*** {
				// No attribute to validate
				return
	***REMOVED***
			isValidMappingMethod := false
			for _, validMappingMethod := range validMappingMethods {
				if mappingMethod.ValueString(***REMOVED*** == validMappingMethod {
					isValidMappingMethod = true
		***REMOVED***
	***REMOVED***
			if !isValidMappingMethod {
				resp.Diagnostics.AddError("Invalid mapping_method.",
					fmt.Sprintf("Expected a valid mapping method. Options are %s. Got %s", validMappingMethods, mappingMethod.ValueString(***REMOVED******REMOVED***,
				***REMOVED***
	***REMOVED***
***REMOVED******REMOVED***,
	}
}
