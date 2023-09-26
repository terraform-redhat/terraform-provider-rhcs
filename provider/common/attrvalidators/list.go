package attrvalidators

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
***REMOVED***

type listValidator struct {
	desc      string
	validator func(context.Context, validator.ListRequest, *validator.ListResponse***REMOVED***
}

func (v *listValidator***REMOVED*** ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse***REMOVED*** {
	v.validator(ctx, req, resp***REMOVED***
}
func (v *listValidator***REMOVED*** Description(ctx context.Context***REMOVED*** string {
	return v.desc
}
func (v *listValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.desc
}

func NewListValidator(desc string, validator func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse***REMOVED******REMOVED*** validator.List {
	return &listValidator{
		desc:      desc,
		validator: validator,
	}
}
