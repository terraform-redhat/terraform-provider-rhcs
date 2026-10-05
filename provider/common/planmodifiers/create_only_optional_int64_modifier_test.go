// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/ginkgo/v2/dsl/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateOnlyOptionalInt64", func() {
	DescribeTable("PlanModifyInt64",
		func(request planmodifier.Int64Request, expected types.Int64) {
			resp := &planmodifier.Int64Response{
				PlanValue: request.PlanValue,
			}
			CreateOnlyOptionalInt64().PlanModifyInt64(context.Background(), request, resp)
			Expect(resp.PlanValue).To(Equal(expected))
		},
		Entry("omitted config nulls plan even when state is known",
			planmodifier.Int64Request{
				ConfigValue: types.Int64Null(),
				StateValue:  types.Int64Value(15),
				PlanValue:   types.Int64Value(15),
			},
			types.Int64Null(),
		),
		Entry("config value passes through",
			planmodifier.Int64Request{
				ConfigValue: types.Int64Value(3),
				StateValue:  types.Int64Value(15),
				PlanValue:   types.Int64Unknown(),
			},
			types.Int64Value(3),
		),
		Entry("unknown config stays unknown",
			planmodifier.Int64Request{
				ConfigValue: types.Int64Unknown(),
				StateValue:  types.Int64Value(15),
				PlanValue:   types.Int64Unknown(),
			},
			types.Int64Unknown(),
		),
		Entry("omitted on create leaves plan null",
			planmodifier.Int64Request{
				ConfigValue: types.Int64Null(),
				StateValue:  types.Int64Null(),
				PlanValue:   types.Int64Unknown(),
			},
			types.Int64Null(),
		),
	)
})
