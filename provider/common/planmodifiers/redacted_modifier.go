package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Redacted() planmodifier.Map {
	return redactedModifier{
		"Changes the plan to reflect the redacted response from the backend and improve change detection.",
		"Changes the plan to reflect the redacted response from the backend and improve change detection.",
	}
}

type redactedModifier struct {
	description         string
	markdownDescription string
}

// Description returns a human-readable description of the plan modifier.
func (m redactedModifier) Description(_ context.Context) string {
	return m.description
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m redactedModifier) MarkdownDescription(_ context.Context) string {
	return m.markdownDescription
}

// PlanModifyMap implements the plan modification logic.
func (m redactedModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest,
	resp *planmodifier.MapResponse) {
	// Do not replace on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do not replace on resource destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	// Do not replace if the plan and state values are equal.
	if req.PlanValue.Equal(req.StateValue) {
		return
	}

	// If plan contains only keys that in the state has a value of REDACTED -> change the plan value to REDACTED
	// If plan contains at least one key that has not the value of REDACTED -> leave plan unchanged
	// In this way, we won't show diffs if there are no changes
	// If there are changes, we will also show diff if value is unchanged, but we don't patch partially maps, so we
	// need to send the full values to the backend or the existing ones will be replaced wrongly with REDACTED
	newPlanMapWithRedacted := map[string]attr.Value{}
	atLeastOneNonRedacted := false

	isLengthChanged := len(req.PlanValue.Elements()) != len(req.StateValue.Elements())
	if isLengthChanged {
		// send back the unchanged plan
		resp.PlanValue = req.PlanValue
		return
	}

	// If length is the same, we need to compare maps keys for equality
	for k, v := range req.PlanValue.Elements() {
		if _, ok := req.StateValue.Elements()[k]; ok {
			newPlanMapWithRedacted[k] = types.StringValue("REDACTED")
		} else {
			atLeastOneNonRedacted = true
			newPlanMapWithRedacted[k] = v
		}

	}
	if atLeastOneNonRedacted {
		// send back the unchanged plan
		resp.PlanValue = req.PlanValue
	}

	// send back the modified plan
	newRedactedPlan, diags := types.MapValue(types.StringType, newPlanMapWithRedacted)
	resp.Diagnostics.Append(diags...)
	resp.PlanValue = newRedactedPlan
}
