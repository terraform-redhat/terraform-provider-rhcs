package attrvalidators

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
***REMOVED***

type mapValidator struct {
	desc      string
	validator func(context.Context, validator.MapRequest, *validator.MapResponse***REMOVED***
}

func (v *mapValidator***REMOVED*** ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse***REMOVED*** {
	v.validator(ctx, req, resp***REMOVED***
}
func (v *mapValidator***REMOVED*** Description(ctx context.Context***REMOVED*** string {
	return v.desc
}
func (v *mapValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.desc
}

func NewMapValidator(desc string, validator func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse***REMOVED******REMOVED*** validator.Map {
	return &mapValidator{
		desc:      desc,
		validator: validator,
	}
}
