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

var _ = Describe("Immutable Bool Modifier", func() {
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.BoolAttribute{
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

	testPlan := func(value types.Bool) tfsdk.Plan {
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

	testState := func(value types.Bool) tfsdk.State {
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
		func(request planmodifier.BoolRequest, expectError bool) {
			resp := &planmodifier.BoolResponse{
				PlanValue: request.PlanValue,
			}

			ImmutableBool().PlanModifyBool(context.Background(), request, resp)

			Expect(resp.Diagnostics.HasError()).To(Equal(expectError))
			if expectError {
				Expect(resp.Diagnostics).To(HaveLen(1))
				Expect(resp.Diagnostics[0].Summary()).To(Equal(common.AssertionErrorSummaryMessage))
				Expect(resp.Diagnostics[0].Detail()).To(ContainSubstring("Attribute testattr, cannot be changed"))
			}
		},
		Entry("resource creation",
			planmodifier.BoolRequest{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.BoolValue(true)),
				PlanValue:  types.BoolValue(true),
				State:      nullState,
				StateValue: types.BoolNull(),
			},
			false,
		),
		Entry("resource destroy",
			planmodifier.BoolRequest{
				Path:       path.Root("testattr"),
				Plan:       nullPlan,
				PlanValue:  types.BoolNull(),
				State:      testState(types.BoolValue(true)),
				StateValue: types.BoolValue(true),
			},
			false,
		),
		Entry("unchanged value",
			planmodifier.BoolRequest{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.BoolValue(true)),
				PlanValue:  types.BoolValue(true),
				State:      testState(types.BoolValue(true)),
				StateValue: types.BoolValue(true),
			},
			false,
		),
		Entry("changed from true to false",
			planmodifier.BoolRequest{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.BoolValue(false)),
				PlanValue:  types.BoolValue(false),
				State:      testState(types.BoolValue(true)),
				StateValue: types.BoolValue(true),
			},
			true,
		),
		Entry("changed from false to true",
			planmodifier.BoolRequest{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.BoolValue(true)),
				PlanValue:  types.BoolValue(true),
				State:      testState(types.BoolValue(false)),
				StateValue: types.BoolValue(false),
			},
			true,
		),
		Entry("changed from null to value",
			planmodifier.BoolRequest{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.BoolValue(true)),
				PlanValue:  types.BoolValue(true),
				State:      testState(types.BoolNull()),
				StateValue: types.BoolNull(),
			},
			true,
		),
		Entry("changed from value to null",
			planmodifier.BoolRequest{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.BoolNull()),
				PlanValue:  types.BoolNull(),
				State:      testState(types.BoolValue(true)),
				StateValue: types.BoolValue(true),
			},
			true,
		),
		Entry("unknown plan value skips validation",
			planmodifier.BoolRequest{
				Path:       path.Root("testattr"),
				Plan:       testPlan(types.BoolUnknown()),
				PlanValue:  types.BoolUnknown(),
				State:      testState(types.BoolValue(true)),
				StateValue: types.BoolValue(true),
			},
			false,
		),
	)
})
