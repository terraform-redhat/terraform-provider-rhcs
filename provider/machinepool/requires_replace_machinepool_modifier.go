/*
Copyright (c) 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package machinepool

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type machinepoolRequiresReplacer interface {
	planmodifier.Bool
	planmodifier.Float64
	planmodifier.Int64
	planmodifier.String
}

// requiresReplaceMachinepool returns a plan modifier that is identical to
// requiresReplace, except in the case of the default machine pool, where it
// returns an error instead of requiring replacement.
func requiresReplaceMachinepool() machinepoolRequiresReplacer {
	return requiresReplaceMachinepoolImpl{}
}

type requiresReplaceMachinepoolImpl struct{}

func (r requiresReplaceMachinepoolImpl) Description(_ context.Context) string {
	return "Return an error instead of trying to recreate the default machine pool"
}
func (r requiresReplaceMachinepoolImpl) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}

func (r requiresReplaceMachinepoolImpl) planModify(ctx context.Context, attrPath path.Path,
	config tfsdk.Config, configValue attr.Value,
	plan tfsdk.Plan, planValue attr.Value,
	state tfsdk.State, stateValue attr.Value) (bool, diag.Diagnostics) {
	tflog.Debug(ctx, "requiresReplaceMachinepool modifier", map[string]interface{}{
		"Attribute":   attrPath.String(),
		"Config":      configValue.String(),
		"Plan":        planValue.String(),
		"PlanIsNull":  plan.Raw.IsNull(),
		"State":       stateValue.String(),
		"StateIsNull": state.Raw.IsNull(),
	})

	// The resource is being created, so we allow it.
	if state.Raw.IsNull() {
		tflog.Debug(ctx, "requiresReplaceMachinepool modifier", map[string]interface{}{
			"Attribute": attrPath.String(),
			"Operation": "ResourceCreate",
			"Value":     configValue.String(),
		})
		return false, nil
	}

	// The resource is being destroyed, so we allow it.
	if plan.Raw.IsNull() {
		tflog.Debug(ctx, "requiresReplaceMachinepool modifier", map[string]interface{}{
			"Attribute": attrPath.String(),
			"Operation": "ResourceDestroy",
		})
		return false, nil
	}

	// The attribute is unchanged, so we allow it.
	if planValue.Equal(stateValue) {
		return false, nil
	}

	// At this point, we know that the attribute is changing, so we need to make sure it's not the default machine pool
	replace := true
	diags := diag.Diagnostics{}

	poolName := &types.String{}
	diags.Append(config.GetAttribute(ctx, path.Root("name"), poolName)...)
	if poolName.ValueString() == defaultMachinePoolName {
		diags.AddAttributeError(
			attrPath,
			"Attribute cannot be changed for the default machine pool",
			fmt.Sprintf("The attribute '%s' cannot be changed for the default machine pool because it"+
				" would require recreating the machine pool, which is not supported.", attrPath.String()),
		)
	}

	return replace, diags
}

func (r requiresReplaceMachinepoolImpl) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	replace, diags := r.planModify(ctx, req.Path, req.Config, req.ConfigValue, req.Plan, req.PlanValue, req.State, req.StateValue)
	resp.RequiresReplace = replace
	resp.Diagnostics.Append(diags...)
}

func (r requiresReplaceMachinepoolImpl) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	replace, diags := r.planModify(ctx, req.Path, req.Config, req.ConfigValue, req.Plan, req.PlanValue, req.State, req.StateValue)
	resp.RequiresReplace = replace
	resp.Diagnostics.Append(diags...)
}

func (r requiresReplaceMachinepoolImpl) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	replace, diags := r.planModify(ctx, req.Path, req.Config, req.ConfigValue, req.Plan, req.PlanValue, req.State, req.StateValue)
	resp.RequiresReplace = replace
	resp.Diagnostics.Append(diags...)
}

func (r requiresReplaceMachinepoolImpl) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	replace, diags := r.planModify(ctx, req.Path, req.Config, req.ConfigValue, req.Plan, req.PlanValue, req.State, req.StateValue)
	resp.RequiresReplace = replace
	resp.Diagnostics.Append(diags...)
}
