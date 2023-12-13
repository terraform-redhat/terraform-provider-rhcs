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

package common

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ImmutableModifier is the Immutable plan modifier supports the following attribute types:
type ImmutableModifier interface {
	planmodifier.Bool
	planmodifier.Float64
	planmodifier.Int64
	planmodifier.List
	planmodifier.Map
	planmodifier.Number
	planmodifier.Set
	planmodifier.String
	planmodifier.Object
}

type immutable struct{}

// Immutable returns a plan modifier that prevents an existing configuration
// value from being changed.
//
//   - In case of computed attribute, if user doesn't set a value, the plan would be set to `unknown` and
//     this modifier is skipped
func Immutable() ImmutableModifier {
	return immutable{}
}

func (v immutable) Description(_ context.Context) string {
	return "once set, the value of this attribute may not be changed"
}

func (v immutable) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// validateUnchanged checks to see if the value has been changed. The value is checked in a
// generic way so that we can use the same implementation regardless of the
// underlying attribute type.
func (v immutable) validateUnchanged(ctx context.Context, attrPath path.Path,
	config tfsdk.Config, configValue attr.Value,
	plan tfsdk.Plan, planValue attr.Value,
	state tfsdk.State, stateValue attr.Value) diag.Diagnostics {
	tflog.Debug(ctx, "Immutable modifier", map[string]interface{}{
		"Attribute":   attrPath.String(),
		"Config":      configValue.String(),
		"Plan":        planValue.String(),
		"PlanIsNull":  plan.Raw.IsNull(),
		"State":       stateValue.String(),
		"StateIsNull": state.Raw.IsNull(),
	})

	// In case of computed attribute the
	// Do not replace on resource creation.
	if state.Raw.IsNull() {
		return nil
	}

	// Do not replace on resource destroy.
	if plan.Raw.IsNull() {
		return nil
	}

	// Do not replace if the plan and state values are equal.
	if planValue.Equal(stateValue) {
		return nil
	}

	attrSchema, diags := config.Schema.AttributeAtPath(ctx, attrPath)
	if diags.HasError() {
		return diags
	}

	if attrSchema.IsComputed() && planValue.IsUnknown() {
		return nil
	}

	tflog.Debug(ctx, "Immutable modifier", map[string]interface{}{
		"Attribute": attrPath.String(),
		"Operation": "ConfigChanged",
		"From":      stateValue.String(),
		"To":        configValue.String(),
	})
	return diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(attrPath,
			"attribute \""+attrPath.String()+"\" must have a known value and may not be changed.",
			"Attempted to change attribute \""+attrPath.String()+"\" from "+stateValue.String()+" to "+configValue.String()+".",
		),
	}
}

func (v immutable) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}

func (v immutable) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}

func (v immutable) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}

func (v immutable) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}

func (v immutable) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}

func (v immutable) PlanModifyNumber(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}

func (v immutable) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}

func (v immutable) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}

func (v immutable) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue)
	resp.Diagnostics.Append(diags...)
}
