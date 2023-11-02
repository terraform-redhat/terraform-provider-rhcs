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

package clusterautoscaler

***REMOVED***
	"context"
***REMOVED***
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
***REMOVED***

func stringFloatRangeValidator(desc string, min float64, max float64***REMOVED*** validator.String {
	return attrvalidators.NewStringValidator(desc, func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
		attribute := &types.String{}
		diag := req.Config.GetAttribute(ctx, req.Path, attribute***REMOVED***

		if diag.HasError(***REMOVED*** {
			// No attribute to validate
			return
***REMOVED***

		if attribute.IsNull(***REMOVED*** || attribute.IsUnknown(***REMOVED*** {
			// No need to validate
			return
***REMOVED***

		number, err := strconv.ParseFloat(attribute.ValueString(***REMOVED***, 64***REMOVED***
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Value cannot be parsed to a float value",
				fmt.Sprintf("Value '%s' cannot be parsed to a float value", attribute.ValueString(***REMOVED******REMOVED***,
			***REMOVED***
			return
***REMOVED***

		if number < min || number > max {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Value out of range",
				fmt.Sprintf("Value '%f' is out of range %f - %f", number, min, max***REMOVED***,
			***REMOVED***
***REMOVED***
	}***REMOVED***
}

func rangeValidator(desc string***REMOVED*** validator.Object {
	return attrvalidators.NewObjectValidator("max must be greater or equal to min",
		func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED*** {
			resourceRange := &AutoscalerResourceRange{}
			diag := req.Config.GetAttribute(ctx, req.Path, resourceRange***REMOVED***
			if diag.HasError(***REMOVED*** {
				// No attribute to validate
				return
	***REMOVED***

			steps := []string{}
			for _, step := range req.Path.Steps(***REMOVED*** {
				steps = append(steps, fmt.Sprintf("%s", step***REMOVED******REMOVED***
	***REMOVED***
			if resourceRange.Min.ValueInt64(***REMOVED*** > resourceRange.Max.ValueInt64(***REMOVED*** {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid resource range",
					fmt.Sprintf("In '%s' attribute, max value must be greater or equal to min value", strings.Join(steps, "."***REMOVED******REMOVED***,
				***REMOVED***
	***REMOVED***
***REMOVED******REMOVED***
}

func durationStringValidator(desc string***REMOVED*** validator.String {
	return attrvalidators.NewStringValidator(desc, func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
		attribute := &types.String{}
		diag := req.Config.GetAttribute(ctx, req.Path, attribute***REMOVED***

		if diag.HasError(***REMOVED*** {
			// No attribute to validate
			return
***REMOVED***

		if attribute.IsNull(***REMOVED*** || attribute.IsUnknown(***REMOVED*** {
			// No need to validate
			return
***REMOVED***

		if _, err := time.ParseDuration(attribute.ValueString(***REMOVED******REMOVED***; err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Value cannot be parsed to a duration string",
				fmt.Sprintf("Value '%s' cannot be parsed to a duration string. A duration "+
					"string is a sequence of decimal numbers and a time unit suffix such as \"300m\", "+
					"\"1.5h\" or \"2h45m\"",
					attribute.ValueString(***REMOVED******REMOVED***,
			***REMOVED***
			return
***REMOVED***
	}***REMOVED***
}
