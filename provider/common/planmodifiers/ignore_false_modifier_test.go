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
)

var _ = Describe("Ignore False Modifier", func() {
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"testattr": schema.BoolAttribute{
				Optional: true,
			},
		},
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

	nullState := tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			testSchema.Type().TerraformType(context.Background()),
			nil,
		),
	}

	DescribeTable("should normalize false to null",
		func(request planmodifier.BoolRequest, expectedValue types.Bool) {
			resp := &planmodifier.BoolResponse{
				PlanValue: request.PlanValue,
			}

			IgnoreFalse().PlanModifyBool(context.Background(), request, resp)

			Expect(resp.Diagnostics.HasError()).To(BeFalse())
			Expect(resp.PlanValue).To(Equal(expectedValue))
		},
		Entry("false is normalized to null",
			planmodifier.BoolRequest{
				Path:      path.Root("testattr"),
				Plan:      testPlan(types.BoolValue(false)),
				PlanValue: types.BoolValue(false),
				State:     nullState,
			},
			types.BoolNull(),
		),
		Entry("true is preserved",
			planmodifier.BoolRequest{
				Path:      path.Root("testattr"),
				Plan:      testPlan(types.BoolValue(true)),
				PlanValue: types.BoolValue(true),
				State:     nullState,
			},
			types.BoolValue(true),
		),
		Entry("null is preserved",
			planmodifier.BoolRequest{
				Path:      path.Root("testattr"),
				Plan:      testPlan(types.BoolNull()),
				PlanValue: types.BoolNull(),
				State:     nullState,
			},
			types.BoolNull(),
		),
		Entry("unknown is preserved",
			planmodifier.BoolRequest{
				Path:      path.Root("testattr"),
				Plan:      testPlan(types.BoolUnknown()),
				PlanValue: types.BoolUnknown(),
				State:     nullState,
			},
			types.BoolUnknown(),
		),
	)
})
