// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

// Scheduler profiles are HyperFleet-only. Keeping this as a separate test
// file makes the profile-specific behavior explicit without changing the
// existing cluster creation tests.
var _ = Describe("HyperFleet scheduler profile", ci.Day1Prepare, func() {
	It("creates a cluster with the configured scheduler profile", func(ctx SpecContext) {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if profile.Profile().GetClusterType() != constants.HYPERFLEET {
			Skip("scheduler profile test requires a HyperFleet profile")
		}
		if profile.Profile().GetSchedulerProfile() == "" {
			Skip("test requires a configured HyperFleet scheduler profile")
		}

		workspace := profile.Profile().GetName() + "-scheduler-profile"
		backend, err := clusterworkflow.New(profile, "", workspace)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			destroyer, destroyErr := clusterworkflow.NewDestroyer(profile, "", workspace)
			Expect(destroyErr).NotTo(HaveOccurred())
			Expect(destroyer.Destroy(context.Background())).To(Succeed())
		})

		clusterID, err := exec.CreateAndWaitReady(ctx, backend.Lifecycle)
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterID).NotTo(BeEmpty())
		state, err := backend.Lifecycle.Read(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.SchedulerProfile).To(Equal(profile.Profile().GetSchedulerProfile()))
	})
})
