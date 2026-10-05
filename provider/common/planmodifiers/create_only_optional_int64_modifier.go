// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CreateOnlyOptionalInt64 returns a plan modifier for Optional+Computed create-only
// int64 attributes (for example cluster replicas). When config omits the attribute,
// the plan is null so Update can clear prior state. Unlike UseStateForUnknown, it
// never copies prior state into the plan for an omitted attribute.
func CreateOnlyOptionalInt64() planmodifier.Int64 {
	return createOnlyOptionalInt64Modifier{}
}

type createOnlyOptionalInt64Modifier struct{}

func (m createOnlyOptionalInt64Modifier) Description(_ context.Context) string {
	return "When the attribute is omitted from config, the plan is null instead of copying prior state."
}

func (m createOnlyOptionalInt64Modifier) MarkdownDescription(_ context.Context) string {
	return "When the attribute is omitted from config, the plan is null instead of copying prior state."
}

// PlanModifyInt64 forces omitted config to plan null so create-only reconcile can
// clear state. Config values pass through; unknown config stays unknown.
func (m createOnlyOptionalInt64Modifier) PlanModifyInt64(
	_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response,
) {
	if req.ConfigValue.IsUnknown() {
		resp.PlanValue = types.Int64Unknown()
		return
	}
	if req.ConfigValue.IsNull() {
		resp.PlanValue = types.Int64Null()
		return
	}
	resp.PlanValue = req.ConfigValue
}
