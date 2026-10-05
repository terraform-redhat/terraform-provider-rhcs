// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("Standard worker pool names", func() {
	It("matches HCP default names and rejects unrelated names", func() {
		Expect(StandardWorkerPoolNameRE.MatchString("worker")).To(BeTrue())
		Expect(StandardWorkerPoolNameRE.MatchString("workers")).To(BeTrue())
		Expect(StandardWorkerPoolNameRE.MatchString("workers-0")).To(BeTrue())
		Expect(StandardWorkerPoolNameRE.MatchString("workers-extra")).To(BeFalse())
	})
})

var _ = Describe("ReconcileCreateOnlyReplicas", func() {
	It("nulls state when config omits replicas", func() {
		got, warning := ReconcileCreateOnlyReplicas(
			false, types.Int64Value(15), types.Int64Null(), ClassicCreateOnlyReplicasWarning,
		)
		Expect(got.IsNull()).To(BeTrue())
		Expect(warning).To(BeEmpty())
	})

	It("syncs null state to config without warning", func() {
		got, warning := ReconcileCreateOnlyReplicas(
			true, types.Int64Null(), types.Int64Value(3), ClassicCreateOnlyReplicasWarning,
		)
		Expect(got).To(Equal(types.Int64Value(3)))
		Expect(warning).To(BeEmpty())
	})

	It("keeps equal state and config without warning", func() {
		got, warning := ReconcileCreateOnlyReplicas(
			true, types.Int64Value(3), types.Int64Value(3), ClassicCreateOnlyReplicasWarning,
		)
		Expect(got).To(Equal(types.Int64Value(3)))
		Expect(warning).To(BeEmpty())
	})

	It("syncs mismatched state to config with Classic create-only warning", func() {
		got, warning := ReconcileCreateOnlyReplicas(
			true, types.Int64Value(15), types.Int64Value(3), ClassicCreateOnlyReplicasWarning,
		)
		Expect(got).To(Equal(types.Int64Value(3)))
		Expect(warning).To(Equal(ClassicCreateOnlyReplicasWarning))
		Expect(warning).To(ContainSubstring("rhcs_machine_pool"))
	})

	It("syncs day-2 config increase with HCP create-only warning", func() {
		got, warning := ReconcileCreateOnlyReplicas(
			true, types.Int64Value(3), types.Int64Value(6), HCPCreateOnlyReplicasWarning,
		)
		Expect(got).To(Equal(types.Int64Value(6)))
		Expect(warning).To(Equal(HCPCreateOnlyReplicasWarning))
		Expect(warning).To(ContainSubstring("rhcs_hcp_machine_pool"))
	})
})
