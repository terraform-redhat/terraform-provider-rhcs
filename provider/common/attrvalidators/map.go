package attrvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type mapValidator struct {
	desc      string
	validator func(context.Context, validator.MapRequest, *validator.MapResponse)
}

func (v *mapValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	v.validator(ctx, req, resp)
}
func (v *mapValidator) Description(ctx context.Context) string {
	return v.desc
}
func (v *mapValidator) MarkdownDescription(ctx context.Context) string {
	return v.desc
}

func NewMapValidator(desc string, validator func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse)) validator.Map {
	return &mapValidator{
		desc:      desc,
		validator: validator,
	}
}
