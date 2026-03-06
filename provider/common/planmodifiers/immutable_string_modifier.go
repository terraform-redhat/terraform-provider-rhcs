package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// ImmutableString returns a string plan modifier that prevents changes to an
// attribute once the resource already exists in state.
func ImmutableString() planmodifier.String {
	return immutableStringModifier{
		"Prevents updates to this string attribute after resource creation.",
		"Prevents updates to this string attribute after resource creation.",
	}
}

type immutableStringModifier struct {
	description         string
	markdownDescription string
}

func (m immutableStringModifier) Description(_ context.Context) string {
	return m.description
}

func (m immutableStringModifier) MarkdownDescription(_ context.Context) string {
	return m.markdownDescription
}

func (m immutableStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do not validate immutability on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do not validate immutability on resource destroy.
	if req.Plan.Raw.IsNull() {
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
