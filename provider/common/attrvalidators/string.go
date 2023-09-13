package attrvalidators

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
***REMOVED***

type stringValidator struct {
	desc      string
	validator func(context.Context, validator.StringRequest, *validator.StringResponse***REMOVED***
}

func (v *stringValidator***REMOVED*** ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
	v.validator(ctx, req, resp***REMOVED***
}
func (v *stringValidator***REMOVED*** Description(ctx context.Context***REMOVED*** string {
	return v.desc
}
func (v *stringValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.desc
}

func NewStringValidator(desc string, validator func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED******REMOVED*** validator.String {
	return &stringValidator{
		desc:      desc,
		validator: validator,
	}

}
