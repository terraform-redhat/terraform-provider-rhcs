package common

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type AttributeValidator struct {
	Desc      string
	MDDesc    string
	Validator func(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse)
}

var _ tfsdk.AttributeValidator = &AttributeValidator{}

func (a *AttributeValidator) Description(context.Context) string {
	return a.Desc
}
func (a *AttributeValidator) MarkdownDescription(context.Context) string {
	return a.MDDesc
}
func (a *AttributeValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	a.Validator(ctx, req, resp)
}
