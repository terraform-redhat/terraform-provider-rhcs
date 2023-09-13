package attrvalidators

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
***REMOVED***

type objectValidator struct {
	desc      string
	validator func(context.Context, validator.ObjectRequest, *validator.ObjectResponse***REMOVED***
}

func (v *objectValidator***REMOVED*** ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED*** {
	v.validator(ctx, req, resp***REMOVED***
}
func (v *objectValidator***REMOVED*** Description(ctx context.Context***REMOVED*** string {
	return v.desc
}
func (v *objectValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.desc
}

func NewObjectValidator(desc string, validator func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED******REMOVED*** validator.Object {
	return &objectValidator{
		desc:      desc,
		validator: validator,
	}

}
