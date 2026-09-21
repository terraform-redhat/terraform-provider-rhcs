/*
Copyright (c) 2026 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package classic

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core"  // nolint
	. "github.com/onsi/ginkgo/v2/dsl/table" // nolint
	. "github.com/onsi/gomega"              // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func TestMachinePool(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Classic Machine Pool Resource Suite")
}

var _ = Describe("Classic Machine Pool state", func() {
	emptyList := types.ListValueMust(types.StringType, []attr.Value{})

	DescribeTable("preserves an explicitly configured empty security group list when OCM omits it",
		func(rawMachinePool string) {
			machinePool, err := cmv1.UnmarshalMachinePool([]byte(rawMachinePool))
			Expect(err).ToNot(HaveOccurred())

			state := &MachinePoolState{
				AdditionalSecurityGroupIds: emptyList,
				AwsTags:                    types.MapNull(types.StringType),
			}
			Expect(populateState(context.Background(), machinePool, state, nil)).To(Succeed())
			Expect(state.AdditionalSecurityGroupIds).To(Equal(emptyList))
		},
		Entry("AWS object without the list", `{"id":"test-pool","aws":{}}`),
		Entry("no AWS object", `{"id":"test-pool"}`),
	)

	DescribeTable("keeps an unset security group list null when OCM omits it",
		func(rawMachinePool string) {
			machinePool, err := cmv1.UnmarshalMachinePool([]byte(rawMachinePool))
			Expect(err).ToNot(HaveOccurred())

			state := &MachinePoolState{
				AdditionalSecurityGroupIds: types.ListNull(types.StringType),
				AwsTags:                    types.MapNull(types.StringType),
			}
			Expect(populateState(context.Background(), machinePool, state, nil)).To(Succeed())
			Expect(state.AdditionalSecurityGroupIds.IsNull()).To(BeTrue())
		},
		Entry("AWS object without the list", `{"id":"test-pool","aws":{}}`),
		Entry("no AWS object", `{"id":"test-pool"}`),
	)

	It("seeds an empty security group list from the plan", func() {
		state := &MachinePoolState{}
		plan := &MachinePoolState{AdditionalSecurityGroupIds: emptyList}

		adjustInitialStateToPlan(state, plan)

		Expect(state.AdditionalSecurityGroupIds).To(Equal(emptyList))
	})

	DescribeTable("validates immutable security group lists",
		func(stateValue, planValue types.List, expectError bool) {
			diags := diag.Diagnostics{}
			validateImmutableList(stateValue, planValue, "aws_additional_security_group_ids", &diags)

			Expect(diags.HasError()).To(Equal(expectError))
		},
		Entry("allows imported null to become configured empty", types.ListNull(types.StringType), emptyList, false),
		Entry("allows configured empty to become omitted", emptyList, types.ListNull(types.StringType), false),
		Entry("rejects a different security group", emptyList,
			types.ListValueMust(types.StringType, []attr.Value{types.StringValue("sg-1")}), true),
	)
})
