package provider

***REMOVED***
	"context"
***REMOVED***
	"github.com/thoas/go-funk"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-ocm/provider/common"
***REMOVED***

func EnumValueValidator(allowedList []string***REMOVED*** []tfsdk.AttributeValidator {
	return []tfsdk.AttributeValidator{
		&common.AttributeValidator{
			Desc: "Validate enum param",
			Validator: func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {

				value := &types.String{}
				diag := req.Config.GetAttribute(ctx, req.AttributePath, value***REMOVED***
				if diag.HasError(***REMOVED*** || common.IsStringAttributeEmpty(*value***REMOVED*** {
					// No attribute to validate
					return
		***REMOVED***

				if funk.Contains(allowedList, value.Value***REMOVED*** {
					return
		***REMOVED***

				resp.Diagnostics.AddError(fmt.Sprintf("Invalid %s.", req.AttributePath.LastStep(***REMOVED******REMOVED***,
					fmt.Sprintf("Expected a valid %s param. Options are %s. Got %s.",
						req.AttributePath.LastStep(***REMOVED***, allowedList, value.Value***REMOVED***,
				***REMOVED***
	***REMOVED***,
***REMOVED***,
	}
}
