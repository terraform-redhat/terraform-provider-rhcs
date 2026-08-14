// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// ImmutableBool returns a bool plan modifier that prevents changes to an
// attribute once the resource already exists in state.
func ImmutableBool() planmodifier.Bool {
	return immutableBoolModifier{
		"Prevents updates to this bool attribute after resource creation.",
		"Prevents updates to this bool attribute after resource creation.",
	}
}

type immutableBoolModifier struct {
	description         string
	markdownDescription string
}

func (m immutableBoolModifier) Description(_ context.Context) string {
	return m.description
}

func (m immutableBoolModifier) MarkdownDescription(_ context.Context) string {
	return m.markdownDescription
}

func (m immutableBoolModifier) PlanModifyBool(
	_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse,
) {
	// Do not validate immutability on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do not validate immutability on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	// Do not validate immutability if plan value is unknown (e.g., from interpolation).
	// Terraform will re-validate after the value becomes known.
	if req.PlanValue.IsUnknown() {
		return
	}

	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	resp.Diagnostics.AddError(
		common.AssertionErrorSummaryMessage,
		fmt.Sprintf(common.AssertionErrorDetailsMessage, req.Path.String(), req.StateValue, req.PlanValue),
	)
}
