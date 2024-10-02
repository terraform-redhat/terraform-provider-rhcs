package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Redacted Modifier", func() {
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.MapAttribute{
				ElementType: types.StringType,
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

	testPlan := func(value types.Map) tfsdk.Plan {
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

	testState := func(value types.Map) tfsdk.State {
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

	DescribeTable("should produce the right plan",
		func(request planmodifier.MapRequest,
			expected *planmodifier.MapResponse) {
			resp := &planmodifier.MapResponse{
				PlanValue: request.PlanValue,
			}

			Redacted().PlanModifyMap(context.Background(), request, resp)
			Expect(resp).To(BeEquivalentTo(expected))
		},
		Entry("resource creation",
			planmodifier.MapRequest{
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				Plan:        testPlan(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")})),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				State:       nullState,
				StateValue:  types.MapNull(types.StringType),
			},
			&planmodifier.MapResponse{
				PlanValue:       types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				RequiresReplace: false,
			},
		),
		Entry("resource destroy",
			planmodifier.MapRequest{
				ConfigValue: types.MapNull(types.StringType),
				Plan:        nullPlan,
				PlanValue:   types.MapNull(types.StringType),
				State:       testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED")})),
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED")}),
			},
			&planmodifier.MapResponse{
				PlanValue:       types.MapNull(types.StringType),
				RequiresReplace: false,
			},
		),
		Entry("planvalue statevalue equal, no modification",
			planmodifier.MapRequest{
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				Plan:        testPlan(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")})),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				State:       testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")})),
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
			},
			&planmodifier.MapResponse{
				PlanValue:       types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				RequiresReplace: false,
			},
		),
		Entry("one REDACTED in state, plan modified to reflect that",
			planmodifier.MapRequest{
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				Plan:        testPlan(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("PEM")})),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("PEM")}),
				State:       testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED")})),
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED")}),
			},
			&planmodifier.MapResponse{
				PlanValue:       types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED")}),
				RequiresReplace: false,
			},
		),
		Entry("one REDACTED in state but different one in plan, plan not modified",
			planmodifier.MapRequest{
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				Plan:        testPlan(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey2": types.StringValue("PEM")})),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"testkey2": types.StringValue("PEM")}),
				State:       testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED")})),
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED")}),
			},
			&planmodifier.MapResponse{
				PlanValue:       types.MapValueMust(types.StringType, map[string]attr.Value{"testkey2": types.StringValue("PEM")}),
				RequiresReplace: false,
			},
		),
		Entry("two REDACTED in state but different one in plan, plan not modified",
			planmodifier.MapRequest{
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				Plan:        testPlan(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey2": types.StringValue("PEM")})),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"testkey2": types.StringValue("PEM")}),
				State:       testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED"), "testkey8": types.StringValue("REDACTED")})),
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED"), "testkey8": types.StringValue("REDACTED")}),
			},
			&planmodifier.MapResponse{
				PlanValue:       types.MapValueMust(types.StringType, map[string]attr.Value{"testkey2": types.StringValue("PEM")}),
				RequiresReplace: false,
			},
		),
		Entry("two REDACTED in state and only one of them in plan, plan not modified",
			planmodifier.MapRequest{
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("test")}),
				Plan:        testPlan(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey8": types.StringValue("PEM")})),
				PlanValue:   types.MapValueMust(types.StringType, map[string]attr.Value{"testkey8": types.StringValue("PEM")}),
				State:       testState(types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED"), "testkey8": types.StringValue("REDACTED")})),
				StateValue:  types.MapValueMust(types.StringType, map[string]attr.Value{"testkey": types.StringValue("REDACTED"), "testkey8": types.StringValue("REDACTED")}),
			},
			&planmodifier.MapResponse{
				PlanValue:       types.MapValueMust(types.StringType, map[string]attr.Value{"testkey8": types.StringValue("PEM")}),
				RequiresReplace: false,
			},
		),
	)
})
