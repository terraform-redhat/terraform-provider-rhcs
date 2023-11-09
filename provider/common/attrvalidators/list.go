package attrvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type listValidator struct {
	desc      string
	validator func(context.Context, validator.ListRequest, *validator.ListResponse)
}

func (v *listValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	v.validator(ctx, req, resp)
}
func (v *listValidator) Description(ctx context.Context) string {
	return v.desc
}
func (v *listValidator) MarkdownDescription(ctx context.Context) string {
	return v.desc
}

func NewListValidator(desc string, validator func(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse)) validator.List {
	return &listValidator{
		desc:      desc,
		validator: validator,
	}
}
