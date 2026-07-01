/*
Copyright (c) 2024 Red Hat, Inc.

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
package upgrade

import (
	"context"
	"testing"

	semver "github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// TestUpgrade is the entry point for the Cluster Rosa Classic Upgrade Ginkgo suite.
func TestUpgrade(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Rosa Classic Upgrade Suite")
}

var _ = Describe("CheckAndCancelUpgrades", func() {
	ctx := context.Background()

	// newUpgrade builds a ClusterUpgrade with the given version and state.
	newUpgrade := func(version string, scheduleType cmv1.ScheduleType,
		state cmv1.UpgradePolicyStateValue) ClusterUpgrade {
		policy, err := cmv1.NewUpgradePolicy().
			ScheduleType(scheduleType).
			Version(version).
			Build()
		Expect(err).ToNot(HaveOccurred())

		policyState, err := cmv1.NewUpgradePolicyState().Value(state).Build()
		Expect(err).ToNot(HaveOccurred())

		return ClusterUpgrade{policy: policy, policyState: policyState}
	}

	It("skips recurring (automatic) policies that have no pinned version (issue #1186)", func() {
		// A recurring schedule reports an empty Version() because OCM selects the
		// target at runtime. Before the fix this crashed semver parsing and failed
		// the whole apply. The empty-version policy is skipped before any client
		// call, so a nil client is safe here.
		desired := semver.Must(semver.NewVersion("4.21.20"))
		upgrades := []ClusterUpgrade{
			newUpgrade("", cmv1.ScheduleTypeAutomatic, cmv1.UpgradePolicyStateValuePending),
		}

		correctUpgradePending, err := CheckAndCancelUpgrades(ctx, nil, upgrades, desired)

		Expect(err).ToNot(HaveOccurred())
		Expect(correctUpgradePending).To(BeFalse())
	})

	It("returns no pending upgrade when the policy list is empty", func() {
		desired := semver.Must(semver.NewVersion("4.21.20"))

		correctUpgradePending, err := CheckAndCancelUpgrades(ctx, nil, []ClusterUpgrade{}, desired)

		Expect(err).ToNot(HaveOccurred())
		Expect(correctUpgradePending).To(BeFalse())
	})

	It("reports a correct pending upgrade for a started policy matching the desired version", func() {
		desired := semver.Must(semver.NewVersion("4.21.20"))
		upgrades := []ClusterUpgrade{
			newUpgrade("4.21.20", cmv1.ScheduleTypeManual, cmv1.UpgradePolicyStateValueStarted),
		}

		correctUpgradePending, err := CheckAndCancelUpgrades(ctx, nil, upgrades, desired)

		Expect(err).ToNot(HaveOccurred())
		Expect(correctUpgradePending).To(BeTrue())
	})

	It("errors when a started upgrade is for a different version", func() {
		desired := semver.Must(semver.NewVersion("4.21.20"))
		upgrades := []ClusterUpgrade{
			newUpgrade("4.21.19", cmv1.ScheduleTypeManual, cmv1.UpgradePolicyStateValueStarted),
		}

		_, err := CheckAndCancelUpgrades(ctx, nil, upgrades, desired)

		Expect(err).To(HaveOccurred())
	})
})
