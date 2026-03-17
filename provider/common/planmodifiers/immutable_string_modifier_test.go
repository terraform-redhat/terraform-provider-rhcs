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

var _ = Describe("Immutable String Modifier", func() {
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.StringAttribute{
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

	testPlan := func(value types.String) tfsdk.Plan {
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

	testState := func(value types.String) tfsdk.State {
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
		func(request planmodifier.StringRequest, expectError bool) {
			resp := &planmodifier.StringResponse{
				PlanValue: request.PlanValue,
			}

			ImmutableString().PlanModifyString(context.Background(), request, resp)

			Expect(resp.Diagnostics.HasError()).To(Equal(expectError))
			if expectError {
				Expect(resp.Diagnostics).To(HaveLen(1))
				Expect(resp.Diagnostics[0].Summary()).To(Equal(common.AssertionErrorSummaryMessage))
				Expect(resp.Diagnostics[0].Detail()).To(ContainSubstring("Attribute testattr, cannot be changed"))
			}
		},
		Entry("resource creation",
			planmodifier.StringRequest{
				Path:        path.Root("testattr"),
				ConfigValue: types.StringValue("new"),
				Plan:        testPlan(types.StringValue("new")),
				PlanValue:   types.StringValue("new"),
				State:       nullState,
				StateValue:  types.StringNull(),
			},
			false,
		),
		Entry("resource destroy",
			planmodifier.StringRequest{
				Path:        path.Root("testattr"),
				ConfigValue: types.StringNull(),
				Plan:        nullPlan,
				PlanValue:   types.StringNull(),
				State:       testState(types.StringValue("old")),
				StateValue:  types.StringValue("old"),
			},
			false,
		),
		Entry("unchanged value",
			planmodifier.StringRequest{
				Path:        path.Root("testattr"),
				ConfigValue: types.StringValue("same"),
				Plan:        testPlan(types.StringValue("same")),
				PlanValue:   types.StringValue("same"),
				State:       testState(types.StringValue("same")),
				StateValue:  types.StringValue("same"),
			},
			false,
		),
		Entry("changed from null to value",
			planmodifier.StringRequest{
				Path:        path.Root("testattr"),
				ConfigValue: types.StringValue("new"),
				Plan:        testPlan(types.StringValue("new")),
				PlanValue:   types.StringValue("new"),
				State:       testState(types.StringNull()),
				StateValue:  types.StringNull(),
			},
			true,
		),
		Entry("changed from value to null",
			planmodifier.StringRequest{
				Path:        path.Root("testattr"),
				ConfigValue: types.StringNull(),
				Plan:        testPlan(types.StringNull()),
				PlanValue:   types.StringNull(),
				State:       testState(types.StringValue("old")),
				StateValue:  types.StringValue("old"),
			},
			true,
		),
		Entry("changed from known value to unknown",
			planmodifier.StringRequest{
				Path:        path.Root("testattr"),
				ConfigValue: types.StringUnknown(),
				Plan:        testPlan(types.StringUnknown()),
				PlanValue:   types.StringUnknown(),
				State:       testState(types.StringValue("old")),
				StateValue:  types.StringValue("old"),
			},
			true,
		),
	)
})
