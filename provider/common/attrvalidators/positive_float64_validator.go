// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package attrvalidators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type positiveFloat64Validator struct{}

func (v positiveFloat64Validator) Description(_ context.Context) string {
	return "value must be strictly positive (> 0)"
}

func (v positiveFloat64Validator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v positiveFloat64Validator) ValidateFloat64(
	_ context.Context, req validator.Float64Request, resp *validator.Float64Response,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueFloat64() <= 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			fmt.Sprintf("Value must be strictly positive (> 0), got %g", req.ConfigValue.ValueFloat64()),
		)
	}
}

func PositiveFloat64() validator.Float64 {
	return positiveFloat64Validator{}
}
