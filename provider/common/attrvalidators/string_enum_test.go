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

var _ = Describe("Enum value validator", func() {
	allowed := []string{"digest", "wildcard"}

	DescribeTable("should validate correctly",
		func(request validator.StringRequest, expectedErr bool) {
			response := validator.StringResponse{}
			EnumValueValidator(allowed).ValidateString(context.Background(), request, &response)
			Expect(response.Diagnostics.HasError()).To(Equal(expectedErr))
		},
		Entry("allowed value -> ok",
			validator.StringRequest{
				Path:           path.Root("type"),
				PathExpression: path.MatchRoot("type"),
				ConfigValue:    types.StringValue("digest"),
			},
			false,
		),
		Entry("disallowed value -> error",
			validator.StringRequest{
				Path:           path.Root("type"),
				PathExpression: path.MatchRoot("type"),
				ConfigValue:    types.StringValue("invalid"),
			},
			true,
		),
		Entry("null value -> ok",
			validator.StringRequest{
				Path:           path.Root("type"),
				PathExpression: path.MatchRoot("type"),
				ConfigValue:    types.StringNull(),
			},
			false,
		),
		Entry("unknown value -> ok",
			validator.StringRequest{
				Path:           path.Root("type"),
				PathExpression: path.MatchRoot("type"),
				ConfigValue:    types.StringUnknown(),
			},
			false,
		),
	)
})
