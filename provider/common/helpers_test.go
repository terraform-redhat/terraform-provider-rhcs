// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2" // nolint
	. "github.com/onsi/gomega"    // nolint
	ocmerrors "github.com/openshift-online/ocm-sdk-go/errors"
	"github.com/pkg/errors"
)

var _ = Describe("Helper function tests", func() {
	Context("ShouldPatchMap", func() {
		map0 := types.Map{}
		map1, _ := ConvertStringMapToMapType(map[string]string{
			"key1": "val1",
		})
		map2, _ := ConvertStringMapToMapType(map[string]string{
			"key": "val1",
		})
		map3, _ := ConvertStringMapToMapType(map[string]string{
			"key1": "val",
		})
		map4, _ := ConvertStringMapToMapType(map[string]string{
			"key1": "val",
			"key2": "val2",
		})

		It("returns true when maps are different", func() {
			r, ok := ShouldPatchMap(map0, map1)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map1))

			r, ok = ShouldPatchMap(map1, map0)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map0))

			r, ok = ShouldPatchMap(map1, map2)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map2))

			r, ok = ShouldPatchMap(map1, map3)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map3))

			r, ok = ShouldPatchMap(map1, map4)
			Expect(ok).To(BeTrue())
			Expect(r).To(Equal(map4))
		})

		It("returns false when maps are identical", func() {
			_, ok := ShouldPatchMap(map1, map1)
			Expect(ok).To(BeFalse())
		})
	})

	DescribeTable("ShouldPatchInt",
		func(state, plan types.Int64, expectedValue int64, expectedPatch bool) {
			value, ok := ShouldPatchInt(state, plan)
			Expect(ok).To(Equal(expectedPatch))
			if expectedPatch {
				Expect(value).To(Equal(expectedValue))
			}
		},
		Entry("plan null -> no patch", types.Int64Null(), types.Int64Null(), int64(0), false),
		Entry("plan unknown -> no patch", types.Int64Unknown(), types.Int64Unknown(), int64(0), false),
		Entry("state null, plan set -> patch", types.Int64Null(), types.Int64Value(3), int64(3), true),
		Entry("state unknown, plan set -> patch", types.Int64Unknown(), types.Int64Value(3), int64(3), true),
		Entry("equal values -> no patch", types.Int64Value(2), types.Int64Value(2), int64(0), false),
		Entry("changed value -> patch", types.Int64Value(2), types.Int64Value(5), int64(5), true),
	)

	DescribeTable("ShouldPatchString",
		func(state, plan types.String, expectedValue string, expectedPatch bool) {
			value, ok := ShouldPatchString(state, plan)
			Expect(ok).To(Equal(expectedPatch))
			if expectedPatch {
				Expect(value).To(Equal(expectedValue))
			}
		},
		Entry("plan null -> no patch", types.StringNull(), types.StringNull(), "", false),
		Entry("plan unknown -> no patch", types.StringUnknown(), types.StringUnknown(), "", false),
		Entry("state null, plan set -> patch", types.StringNull(), types.StringValue("new"), "new", true),
		Entry("state unknown, plan set -> patch", types.StringUnknown(), types.StringValue("new"), "new", true),
		Entry("equal values -> no patch", types.StringValue("same"), types.StringValue("same"), "", false),
		Entry("changed value -> patch", types.StringValue("old"), types.StringValue("new"), "new", true),
	)

	DescribeTable("ShouldPatchBool",
		func(state, plan types.Bool, expectedValue bool, expectedPatch bool) {
			value, ok := ShouldPatchBool(state, plan)
			Expect(ok).To(Equal(expectedPatch))
			if expectedPatch {
				Expect(value).To(Equal(expectedValue))
			}
		},
		Entry("plan null -> no patch", types.BoolNull(), types.BoolNull(), false, false),
		Entry("plan unknown -> no patch", types.BoolUnknown(), types.BoolUnknown(), false, false),
		Entry("state null, plan true -> patch", types.BoolNull(), types.BoolValue(true), true, true),
		Entry("state unknown, plan false -> patch", types.BoolUnknown(), types.BoolValue(false), false, true),
		Entry("equal values -> no patch", types.BoolValue(true), types.BoolValue(true), false, false),
		Entry("changed value -> patch", types.BoolValue(true), types.BoolValue(false), false, true),
	)

	DescribeTable("ShouldPatchList",
		func(state, plan types.List, expectedPatch bool) {
			_, ok := ShouldPatchList(state, plan)
			Expect(ok).To(Equal(expectedPatch))
		},
		Entry("identical lists -> no patch",
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			false,
		),
		Entry("different lists -> patch",
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("b")}),
			true,
		),
	)

	DescribeTable("IsStringAttributeUnknownOrEmpty",
		func(param types.String, expected bool) {
			Expect(IsStringAttributeUnknownOrEmpty(param)).To(Equal(expected))
		},
		Entry("unknown", types.StringUnknown(), true),
		Entry("null", types.StringNull(), true),
		Entry("empty string", types.StringValue(""), true),
		Entry("non-empty", types.StringValue("value"), false),
	)

	DescribeTable("IsStringAttributeKnownAndEmpty",
		func(param types.String, expected bool) {
			Expect(IsStringAttributeKnownAndEmpty(param)).To(Equal(expected))
		},
		Entry("unknown", types.StringUnknown(), false),
		Entry("null", types.StringNull(), true),
		Entry("empty string", types.StringValue(""), true),
		Entry("non-empty", types.StringValue("value"), false),
	)

	DescribeTable("IsGreaterThanOrEqual",
		func(version1, version2 string, expected bool, expectErr bool) {
			result, err := IsGreaterThanOrEqual(version1, version2)
			if expectErr {
				Expect(err).To(HaveOccurred())
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry("equal with openshift-v prefix", "openshift-v4.14.0", "4.14.0", true, false),
		Entry("greater version", "4.15.0", "4.14.0", true, false),
		Entry("less version", "4.13.0", "4.14.0", false, false),
		Entry("invalid version", "not-a-version", "4.14.0", false, true),
	)

	DescribeTable("IsValidDomain",
		func(candidate string, expected bool) {
			Expect(IsValidDomain(candidate)).To(Equal(expected))
		},
		Entry("valid fqdn", "apps.example.com", true),
		Entry("valid with trailing dot", "apps.example.com.", true),
		Entry("single label", "localhost", false),
		Entry("invalid characters", "bad_domain.com", false),
	)

	DescribeTable("EmptiableStringToStringType",
		func(input string, expectNull bool, expectedValue string) {
			result := EmptiableStringToStringType(input)
			if expectNull {
				Expect(result.IsNull()).To(BeTrue())
				return
			}
			Expect(result.ValueString()).To(Equal(expectedValue))
		},
		Entry("empty string -> null", "", true, ""),
		Entry("non-empty -> value", "cluster.example.com", false, "cluster.example.com"),
	)

	DescribeTable("ValidateTimeout",
		func(timeOut *int64, defaultTimeout int64, expected *int64, expectErr bool) {
			result, err := ValidateTimeout(timeOut, defaultTimeout)
			if expectErr {
				Expect(err).To(MatchError(ContainSubstring("timeout must be greater than 0 minutes")))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal(*expected))
		},
		Entry("nil uses default", nil, int64(45), new(int64(45)), false),
		Entry("positive unchanged", new(int64(30)), int64(45), new(int64(30)), false),
		Entry("zero errors", new(int64(0)), int64(45), nil, true),
		Entry("negative errors", new(int64(-1)), int64(45), nil, true),
	)

	Describe("HandleErr", func() {
		It("uses OCM reason when set", func() {
			ocmErr, err := ocmerrors.UnmarshalErrorStatus(`{
				"kind": "Error",
				"id": "400",
				"code": "CLUSTERS-MGMT-400",
				"reason": "cluster is not ready"
			}`, 400)
			Expect(err).NotTo(HaveOccurred())

			result := HandleErr(ocmErr, errors.New("fallback message"))
			Expect(result.Error()).To(ContainSubstring("cluster is not ready"))
		})

		It("falls back to err when reason is empty", func() {
			ocmErr, err := ocmerrors.UnmarshalErrorStatus(`{
				"kind": "Error",
				"id": "500",
				"code": "CLUSTERS-MGMT-500"
			}`, 500)
			Expect(err).NotTo(HaveOccurred())

			result := HandleErr(ocmErr, errors.New("underlying transport error"))
			Expect(result.Error()).To(ContainSubstring("underlying transport error"))
		})
	})

	DescribeTable("ValidateStateAndPlanEquals",
		func(state, plan types.String, expectError bool) {
			diags := diag.Diagnostics{}
			ValidateStateAndPlanEquals(state, plan, "test_attr", &diags)
			Expect(diags.HasError()).To(Equal(expectError))
			if expectError {
				Expect(diags.Errors()[0].Summary()).To(Equal(AssertionErrorSummaryMessage))
			}
		},
		Entry("equal attributes -> ok", types.StringValue("same"), types.StringValue("same"), false),
		Entry("different attributes -> error", types.StringValue("old"), types.StringValue("new"), true),
	)
})

//go:fix inline
func ptrInt64(v int64) *int64 {
	return new(v)
}
