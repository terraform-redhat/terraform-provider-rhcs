/*
Copyright (c***REMOVED*** 2023 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
***REMOVED***

// The Immutable plan modifier supports the following attribute types:
type ImmutableModifier interface {
	planmodifier.Bool
	planmodifier.Float64
	planmodifier.Int64
	planmodifier.List
	planmodifier.Map
	planmodifier.Number
	planmodifier.Set
	planmodifier.String
}

type immutable struct{}

// Immutable returns a plan modifier that prevents an existing configuration
// value from being changed.
//
// - Immutable cannot be applied to any attribute that is computed because a
// computed attribute's value may change based on the state of the resource.
//
// - Immutable attributes cannot be "unknown" at plan time because we need to be
// able to check whether they match the state.
func Immutable(***REMOVED*** ImmutableModifier {
	return immutable{}
}

func (v immutable***REMOVED*** Description(_ context.Context***REMOVED*** string {
	return "once set, the value of this attribute may not be changed"
}

func (v immutable***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.Description(ctx***REMOVED***
}

// validateUnchanged checks to see if the value has been changed. The value is checked in a
// generic way so that we can use the same implementation regardless of the
// underlying attribute type.
func (v immutable***REMOVED*** validateUnchanged(ctx context.Context, attrPath path.Path,
	config tfsdk.Config, configValue attr.Value,
	plan tfsdk.Plan, planValue attr.Value,
	state tfsdk.State, stateValue attr.Value***REMOVED*** diag.Diagnostics {
	tflog.Debug(ctx, "Immutable modifier", map[string]interface{}{
		"Attribute":   attrPath.String(***REMOVED***,
		"Config":      configValue.String(***REMOVED***,
		"Plan":        planValue.String(***REMOVED***,
		"PlanIsNull":  plan.Raw.IsNull(***REMOVED***,
		"State":       stateValue.String(***REMOVED***,
		"StateIsNull": state.Raw.IsNull(***REMOVED***,
	}***REMOVED***

	// Immutable should not be applied to computed attributes because the value
	// of a computed attribute may change based on the state of the resource,
	// and that would cause a discrepancy between the configuration and the
	// state.
	//
	// Users will never see this error because it will trip unconditionally the
	// first time the code is run.
	attrSchema, diags := config.Schema.AttributeAtPath(ctx, attrPath***REMOVED***
	if attrSchema.IsComputed(***REMOVED*** {
		diags.AddAttributeError(attrPath,
			"The Immutable PlanModifier cannot be applied to computed attributes",
			"Immutable cannot be applied to \""+attrPath.String(***REMOVED***+"\", which is a computed attribute.",
		***REMOVED***
	}
	if diags.HasError(***REMOVED*** {
		return diags
	}

	// The resource is being created, so we allow it.
	if state.Raw.IsNull(***REMOVED*** {
		tflog.Debug(ctx, "Immutable modifier", map[string]interface{}{
			"Attribute": attrPath.String(***REMOVED***,
			"Operation": "ResourceCreate",
			"Value":     configValue.String(***REMOVED***,
***REMOVED******REMOVED***
		return nil
	}

	// The resource is being destroyed, so we allow it.
	if plan.Raw.IsNull(***REMOVED*** {
		tflog.Debug(ctx, "Immutable modifier", map[string]interface{}{
			"Attribute": attrPath.String(***REMOVED***,
			"Operation": "ResourceDestroy",
***REMOVED******REMOVED***
		return nil
	}

	// Configuration should have a known value and match the state
	if !configValue.IsUnknown(***REMOVED*** && configValue.Equal(stateValue***REMOVED*** {
		tflog.Debug(ctx, "Immutable modifier", map[string]interface{}{
			"Attribute": attrPath.String(***REMOVED***,
			"Operation": "ConfigMatchesState",
			"Value":     configValue.String(***REMOVED***,
***REMOVED******REMOVED***
		return nil
	}

	tflog.Debug(ctx, "Immutable modifier", map[string]interface{}{
		"Attribute": attrPath.String(***REMOVED***,
		"Operation": "ConfigChanged",
		"From":      stateValue.String(***REMOVED***,
		"To":        configValue.String(***REMOVED***,
	}***REMOVED***
	return diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(attrPath,
			"attribute \""+attrPath.String(***REMOVED***+"\" must have a known value and may not be changed.",
			"Attempted to change attribute \""+attrPath.String(***REMOVED***+"\" from "+stateValue.String(***REMOVED***+" to "+configValue.String(***REMOVED***+".",
		***REMOVED***,
	}
}

func (v immutable***REMOVED*** PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse***REMOVED*** {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (v immutable***REMOVED*** PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response***REMOVED*** {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (v immutable***REMOVED*** PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response***REMOVED*** {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (v immutable***REMOVED*** PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse***REMOVED*** {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (v immutable***REMOVED*** PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse***REMOVED*** {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (v immutable***REMOVED*** PlanModifyNumber(ctx context.Context, req planmodifier.NumberRequest, resp *planmodifier.NumberResponse***REMOVED*** {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (v immutable***REMOVED*** PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse***REMOVED*** {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

func (v immutable***REMOVED*** PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse***REMOVED*** {
	diags := v.validateUnchanged(ctx, req.Path,
		req.Config, req.ConfigValue,
		req.Plan, req.PlanValue,
		req.State, req.StateValue***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}
