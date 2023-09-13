package attrvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type stringValidator struct {
	desc      string
	validator func(context.Context, validator.StringRequest, *validator.StringResponse)
}

func (v *stringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	v.validator(ctx, req, resp)
}
func (v *stringValidator) Description(ctx context.Context) string {
	return v.desc
}
func (v *stringValidator) MarkdownDescription(ctx context.Context) string {
	return v.desc
}

func NewStringValidator(desc string, validator func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse)) validator.String {
	return &stringValidator{
		desc:      desc,
		validator: validator,
	}

}
