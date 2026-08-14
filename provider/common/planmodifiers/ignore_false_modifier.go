// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// IgnoreFalse returns a plan modifier that normalizes an explicit false to null,
// making false equivalent to omitting the attribute entirely.
func IgnoreFalse() planmodifier.Bool {
	return ignoreFalseModifier{}
}

type ignoreFalseModifier struct{}

func (m ignoreFalseModifier) Description(_ context.Context) string {
	return "Normalizes false to null so the attribute behaves as if it was not set."
}

func (m ignoreFalseModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes `false` to null so the attribute behaves as if it was not set."
}

func (m ignoreFalseModifier) PlanModifyBool(
	_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse,
) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	if !req.PlanValue.ValueBool() {
		resp.PlanValue = types.BoolNull()
	}
}
