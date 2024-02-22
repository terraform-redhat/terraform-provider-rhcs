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

package autoscaler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

func StringFloatRangeValidator(desc string, min float64, max float64) validator.String {
	return attrvalidators.NewStringValidator(desc, func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
		attribute := &types.String{}
		diag := req.Config.GetAttribute(ctx, req.Path, attribute)

		if diag.HasError() {
			// No attribute to validate
			return
		}

		if attribute.IsNull() || attribute.IsUnknown() {
			// No need to validate
			return
		}

		number, err := strconv.ParseFloat(attribute.ValueString(), 64)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Value cannot be parsed to a float value",
				fmt.Sprintf("Value '%s' cannot be parsed to a float value", attribute.ValueString()),
			)
			return
		}

		if number < min || number > max {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Value out of range",
				fmt.Sprintf("Value '%f' is out of range %f - %f", number, min, max),
			)
		}
	})
}

func RangeValidator(desc string) validator.Object {
	return attrvalidators.NewObjectValidator("min and max must be not negative values and max must be greater or equal to min",
		func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
			resourceRange := &ResourceRange{}
			diag := req.Config.GetAttribute(ctx, req.Path, resourceRange)
			if diag.HasError() {
				// No attribute to validate
				return
			}

			steps := []string{}
			for _, step := range req.Path.Steps() {
				steps = append(steps, fmt.Sprintf("%s", step))
			}
			if resourceRange.Min.ValueInt64() < 0 {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid resource range",
					fmt.Sprintf("Attribute '%s.min' value must be at least 0, got: %d",
						strings.Join(steps, "."), resourceRange.Min.ValueInt64()),
				)
			}
			if resourceRange.Min.ValueInt64() > resourceRange.Max.ValueInt64() {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid resource range",
					fmt.Sprintf("In '%s' attribute, max value must be greater or equal to min value", strings.Join(steps, ".")),
				)
			}
		})
}

func DurationStringValidator(desc string) validator.String {
	return attrvalidators.NewStringValidator(desc, func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
		attribute := &types.String{}
		diag := req.Config.GetAttribute(ctx, req.Path, attribute)

		if diag.HasError() {
			// No attribute to validate
			return
		}

		if attribute.IsNull() || attribute.IsUnknown() {
			// No need to validate
			return
		}

		if _, err := time.ParseDuration(attribute.ValueString()); err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Value cannot be parsed to a duration string",
				fmt.Sprintf("Value '%s' cannot be parsed to a duration string. A duration "+
					"string is a sequence of decimal numbers and a time unit suffix such as \"300m\", "+
					"\"1.5h\" or \"2h45m\"",
					attribute.ValueString()),
			)
			return
		}
	})
}
