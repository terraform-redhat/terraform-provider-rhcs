package idps

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
***REMOVED***

var validMappingMethods = []string{"claim", "add", "generate", "lookup"} // Default is @ index 0
var DefaultMappingMethod = validMappingMethods[0]

func MappingMethodValidators(***REMOVED*** []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate mapping_method",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {

				mappingMethod := &types.String{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, mappingMethod***REMOVED***
				if diag.HasError(***REMOVED*** || mappingMethod.Unknown || mappingMethod.Null {
					// No attribute to validate
					return
		***REMOVED***
				isValidMappingMethod := false
				for _, validMappingMethod := range validMappingMethods {
					if mappingMethod.Value == validMappingMethod {
						isValidMappingMethod = true
			***REMOVED***
		***REMOVED***
				if !isValidMappingMethod {
					resp.Diagnostics.AddError("Invalid mapping_method.",
						fmt.Sprintf("Expected a valid mapping method. Options are %s. Got %s", validMappingMethods, mappingMethod.Value***REMOVED***,
					***REMOVED***
		***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}
