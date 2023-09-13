package attrvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type objectValidator struct {
	desc      string
	validator func(context.Context, validator.ObjectRequest, *validator.ObjectResponse)
}

func (v *objectValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	v.validator(ctx, req, resp)
}
func (v *objectValidator) Description(ctx context.Context) string {
	return v.desc
}
func (v *objectValidator) MarkdownDescription(ctx context.Context) string {
	return v.desc
}

func NewObjectValidator(desc string, validator func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse)) validator.Object {
	return &objectValidator{
		desc:      desc,
		validator: validator,
	}

}
