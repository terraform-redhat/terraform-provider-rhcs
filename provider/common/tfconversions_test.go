// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
)

var _ = Describe("Terraform conversion helpers", func() {
	DescribeTable("BoolWithTrueDefault",
		func(tfVal types.Bool, expected bool) {
			Expect(BoolWithTrueDefault(tfVal)).To(Equal(expected))
		},
		Entry("null defaults true", types.BoolNull(), true),
		Entry("unknown defaults true", types.BoolUnknown(), true),
		Entry("known true", types.BoolValue(true), true),
		Entry("known false", types.BoolValue(false), false),
	)

	DescribeTable("BoolWithFalseDefault",
		func(tfVal types.Bool, expected bool) {
			Expect(BoolWithFalseDefault(tfVal)).To(Equal(expected))
		},
		Entry("null defaults false", types.BoolNull(), false),
		Entry("unknown defaults false", types.BoolUnknown(), false),
		Entry("known true", types.BoolValue(true), true),
		Entry("known false", types.BoolValue(false), false),
	)

	DescribeTable("OptionalInt64",
		func(tfVal types.Int64, expectNil bool, expected int64) {
			result := OptionalInt64(tfVal)
			if expectNil {
				Expect(result).To(BeNil())
				return
			}
			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal(expected))
		},
		Entry("null -> nil", types.Int64Null(), true, int64(0)),
		Entry("unknown -> nil", types.Int64Unknown(), true, int64(0)),
		Entry("known value", types.Int64Value(42), false, int64(42)),
	)

	DescribeTable("OptionalString",
		func(tfVal types.String, expectNil bool, expected string) {
			result := OptionalString(tfVal)
			if expectNil {
				Expect(result).To(BeNil())
				return
			}
			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal(expected))
		},
		Entry("null -> nil", types.StringNull(), true, ""),
		Entry("unknown -> nil", types.StringUnknown(), true, ""),
		Entry("known value", types.StringValue("value"), false, "value"),
	)

	DescribeTable("OptionalList",
		func(tfVal types.List, expectNil bool, expected []string) {
			result := OptionalList(tfVal)
			if expectNil {
				Expect(result).To(BeNil())
				return
			}
			Expect(result).To(Equal(expected))
		},
		Entry("null -> nil", types.ListNull(types.StringType), true, nil),
		Entry("unknown -> nil", types.ListUnknown(types.StringType), true, nil),
		Entry("known list",
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
			false,
			[]string{"a", "b"},
		),
	)

	It("StringListToArray round-trips string list values", func() {
		list := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("x"), types.StringValue("y")})
		result, err := StringListToArray(context.Background(), list)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]string{"x", "y"}))
	})

	It("StringListToArray returns nil for null list", func() {
		result, err := StringListToArray(context.Background(), types.ListNull(types.StringType))
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})

	It("StringListToArray returns nil for unknown list", func() {
		result, err := StringListToArray(context.Background(), types.ListUnknown(types.StringType))
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})

	It("StringArrayToList round-trips string slice", func() {
		list, err := StringArrayToList([]string{"one", "two"})
		Expect(err).NotTo(HaveOccurred())

		result, err := StringListToArray(context.Background(), list)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal([]string{"one", "two"}))
	})

	It("ConvertStringMapToMapType round-trips map values", func() {
		input := map[string]string{"env": "prod", "team": "platform"}
		tfMap, err := ConvertStringMapToMapType(input)
		Expect(err).NotTo(HaveOccurred())

		result, err := OptionalMap(context.Background(), tfMap)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(input))
	})

	It("OptionalMap returns nil for null map", func() {
		result, err := OptionalMap(context.Background(), types.MapNull(types.StringType))
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})

	It("OptionalMap returns nil for unknown map", func() {
		result, err := OptionalMap(context.Background(), types.MapUnknown(types.StringType))
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})
})
