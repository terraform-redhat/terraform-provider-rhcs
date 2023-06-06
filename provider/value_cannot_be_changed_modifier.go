package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type valueCannotBeChangedModifier struct {
}

func ValueCannotBeChangedModifier() tfsdk.AttributePlanModifier {
	return valueCannotBeChangedModifier{}
}
func (m valueCannotBeChangedModifier) Description(ctx context.Context) string {
	return "The value cannot be changed after the resource was created."
}

func (m valueCannotBeChangedModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m valueCannotBeChangedModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {

	if req.AttributeConfig == nil || req.AttributeState == nil || req.AttributePlan == nil {
		// shouldn't happen, but let's not panic if it does
		return
	}

	if req.State.Raw.IsNull() {
		// if we're creating the resource, no need to delete and
		// recreate it
		return
	}

	if req.Plan.Raw.IsNull() {
		// if we're deleting the resource, no need to delete and
		// recreate it
		return
	}

	attrSchema, err := req.State.Schema.AttributeAtPath(req.AttributePath)
	if err != nil && !errors.Is(err, errors.New("path leads to block, not an attribute")) {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error finding attribute schema",
			fmt.Sprintf("An unexpected error was encountered retrieving the schema for this attribute. This is always a bug in the provider.\n\nError: %s", err),
		)
		return
	}

	configRaw, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error converting config value",
			fmt.Sprintf("An unexpected error was encountered converting a %s to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", req.AttributeConfig.Type(ctx), err),
		)
		return
	}

	if configRaw == nil && attrSchema.Computed {
		// if the config is null and the attribute is computed, this
		// could be an out-of-band change, don't require blocking
		return
	}

	if req.AttributeState.Equal(req.AttributePlan) {
		tflog.Debug(ctx, "attribute state and attribute plan have the same value")
		return
	}

	// the attribute value was changes
	tflog.Debug(ctx, "attribute plan was changed")
	resp.Diagnostics.AddAttributeError(req.AttributePath, "Value cannot be changed", "This attribute is blocked for updating")
	return

}
