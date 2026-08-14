// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// ImmutableFloat64 returns a float64 plan modifier that prevents changes to an
// attribute once the resource already exists in state.
func ImmutableFloat64() planmodifier.Float64 {
	return immutableFloat64Modifier{
		"Prevents updates to this float64 attribute after resource creation.",
		"Prevents updates to this float64 attribute after resource creation.",
	}
}

type immutableFloat64Modifier struct {
	description         string
	markdownDescription string
}

func (m immutableFloat64Modifier) Description(_ context.Context) string {
	return m.description
}

func (m immutableFloat64Modifier) MarkdownDescription(_ context.Context) string {
	return m.markdownDescription
}

func (m immutableFloat64Modifier) PlanModifyFloat64(
	_ context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response,
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
