// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package attrvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Not empty map validator", func() {
	DescribeTable("should validate correctly",
		func(request validator.MapRequest, expectedErr bool) {
			response := validator.MapResponse{}
			NotEmptyMapValidator().ValidateMap(context.Background(), request, &response)
			Expect(response.Diagnostics.HasError()).To(Equal(expectedErr))
		},
		Entry("non-empty map -> ok",
			validator.MapRequest{
				Path:           path.Root("labels"),
				PathExpression: path.MatchRoot("labels"),
				ConfigValue: types.MapValueMust(types.StringType, map[string]attr.Value{
					"env": types.StringValue("prod"),
				}),
			},
			false,
		),
		Entry("empty map -> error",
			validator.MapRequest{
				Path:           path.Root("labels"),
				PathExpression: path.MatchRoot("labels"),
				ConfigValue:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
			},
			true,
		),
		Entry("null map -> ok",
			validator.MapRequest{
				Path:           path.Root("labels"),
				PathExpression: path.MatchRoot("labels"),
				ConfigValue:    types.MapNull(types.StringType),
			},
			false,
		),
		Entry("unknown map -> ok",
			validator.MapRequest{
				Path:           path.Root("labels"),
				PathExpression: path.MatchRoot("labels"),
				ConfigValue:    types.MapUnknown(types.StringType),
			},
			false,
		),
	)
})
