// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package attrvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Positive Float64 Validator", func() {
	DescribeTable("should validate correctly",
		func(value types.Float64, expectError bool) {
			req := validator.Float64Request{
				Path:        path.Root("test"),
				ConfigValue: value,
			}
			resp := &validator.Float64Response{}

			PositiveFloat64().ValidateFloat64(context.Background(), req, resp)

			Expect(resp.Diagnostics.HasError()).To(Equal(expectError))
		},
		Entry("positive value", types.Float64Value(1.5), false),
		Entry("small positive value", types.Float64Value(0.001), false),
		Entry("very small positive value", types.Float64Value(0.000001), false),
		Entry("zero is rejected", types.Float64Value(0), true),
		Entry("negative value is rejected", types.Float64Value(-1.0), true),
		Entry("null is accepted", types.Float64Null(), false),
		Entry("unknown is accepted", types.Float64Unknown(), false),
	)
})
