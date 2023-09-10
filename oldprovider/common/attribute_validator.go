package common

***REMOVED***
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
***REMOVED***

type AttributeValidator struct {
	Desc      string
	MDDesc    string
	Validator func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED***
}

var _ tfsdk.AttributeValidator = &AttributeValidator{}

func (a *AttributeValidator***REMOVED*** Description(context.Context***REMOVED*** string {
	return a.Desc
}
func (a *AttributeValidator***REMOVED*** MarkdownDescription(context.Context***REMOVED*** string {
	return a.MDDesc
}
func (a *AttributeValidator***REMOVED*** Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse***REMOVED*** {
	a.Validator(ctx, req, resp***REMOVED***
}
