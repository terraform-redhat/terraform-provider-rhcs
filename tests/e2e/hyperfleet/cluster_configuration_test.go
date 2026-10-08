// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

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

var _ = Describe("HyperFleet cluster configuration", ci.HyperfleetValidated, func() {
	It("creates a cluster with a scheduler profile.”", func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(profileHandler.Profile().GetClusterType()).To(Equal(constants.HYPERFLEET))
		if profileHandler.Profile().GetSchedulerProfile() == "" {
			Skip("Test requires the HyperFleet scheduler profile")
		}

		workspace := profileHandler.Profile().GetName() + "-cluster-configuration"
		backend, err := clusterworkflow.New(profileHandler, "", workspace)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			destroyer, err := clusterworkflow.NewDestroyer(profileHandler, "", workspace)
			Expect(err).NotTo(HaveOccurred())
			Expect(destroyer.Destroy(context.Background())).To(Succeed())
		})

		clusterID, err := exec.CreateAndWaitReady(ctx, backend.Lifecycle)
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterID).NotTo(BeEmpty())

		clusterService, err := exec.NewHyperfleetClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())
		output, err := clusterService.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(output.SchedulerProfile).To(Equal(profileHandler.Profile().GetSchedulerProfile()))
	})
})
