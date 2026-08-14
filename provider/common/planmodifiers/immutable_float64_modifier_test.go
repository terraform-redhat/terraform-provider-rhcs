// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

var _ = Describe("Immutable Float64 Modifier", func() {
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.Float64Attribute{
				Optional: true,
			},
		},
	}

	nullPlan := tfsdk.Plan{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			testSchema.Type().TerraformType(context.Background()),
			nil,
		),
	}

	nullState := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			testSchema.Type().TerraformType(context.Background()),
			nil,
		),
	}

	testPlan := func(value types.Float64) tfsdk.Plan {
		tfValue, err := value.ToTerraformValue(context.Background())
		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}
		return tfsdk.Plan{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(context.Background()),
				map[string]tftypes.Value{
					"testattr": tfValue,
				},
			),
		}
	}

	testState := func(value types.Float64) tfsdk.State {
		tfValue, err := value.ToTerraformValue(context.Background())
		if err != nil {
			panic("ToTerraformValue error: " + err.Error())
		}
		return tfsdk.State{
			Schema: testSchema,
			Raw: tftypes.NewValue(
				testSchema.Type().TerraformType(context.Background()),
				map[string]tftypes.Value{
					"testattr": tfValue,
				},
			),
		}
	}

	DescribeTable("should produce expected diagnostics",
		func(request planmodifier.Float64Request, expectError bool) {
			resp := &planmodifier.Float64Response{
				PlanValue: request.PlanValue,
			}

			ImmutableFloat64().PlanModifyFloat64(context.Background(), request, resp)

			Expect(resp.Diagnostics.HasError()).To(Equal(expectError))
			if expectError {
				Expect(resp.Diagnostics).To(HaveLen(1))
				Expect(resp.Diagnostics[0].Summary()).To(Equal(common.AssertionErrorSummaryMessage))
				Expect(resp.Diagnostics[0].Detail()).To(ContainSubstring("Attribute testattr, cannot be changed"))
			}
		},
		Entry("resource creation",
			planmodifier.Float64Request{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.Float64Value(1.5)),
				PlanValue:  types.Float64Value(1.5),
				State:      nullState,
				StateValue: types.Float64Null(),
			},
			false,
		),
		Entry("resource destroy",
			planmodifier.Float64Request{
				Path:       path.Root("testattr"),
				Plan:       nullPlan,
				PlanValue:  types.Float64Null(),
				State:      testState(types.Float64Value(1.5)),
				StateValue: types.Float64Value(1.5),
			},
			false,
		),
		Entry("unchanged value",
			planmodifier.Float64Request{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.Float64Value(1.5)),
				PlanValue:  types.Float64Value(1.5),
				State:      testState(types.Float64Value(1.5)),
				StateValue: types.Float64Value(1.5),
			},
			false,
		),
		Entry("changed value",
			planmodifier.Float64Request{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.Float64Value(2.5)),
				PlanValue:  types.Float64Value(2.5),
				State:      testState(types.Float64Value(1.5)),
				StateValue: types.Float64Value(1.5),
			},
			true,
		),
		Entry("changed from null to value",
			planmodifier.Float64Request{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.Float64Value(1.5)),
				PlanValue:  types.Float64Value(1.5),
				State:      testState(types.Float64Null()),
				StateValue: types.Float64Null(),
			},
			true,
		),
		Entry("changed from value to null",
			planmodifier.Float64Request{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.Float64Null()),
				PlanValue:  types.Float64Null(),
				State:      testState(types.Float64Value(1.5)),
				StateValue: types.Float64Value(1.5),
			},
			true,
		),
		Entry("unknown plan value skips validation",
			planmodifier.Float64Request{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.Float64Unknown()),
				PlanValue:  types.Float64Unknown(),
				State:      testState(types.Float64Value(1.5)),
				StateValue: types.Float64Value(1.5),
			},
			false,
		),
	)
})
