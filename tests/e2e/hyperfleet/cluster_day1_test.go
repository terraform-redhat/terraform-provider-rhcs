// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("HyperFleet cluster configuration", ci.HyperfleetValidated, func() {
	It("creates a cluster with a scheduler profile", func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(profileHandler.Profile().GetClusterType()).To(Equal(constants.HYPERFLEET))
		if profileHandler.Profile().GetSchedulerProfile() == "" {
			Skip("Test requires the HyperFleet scheduler profile")
		}

		workspace := hyperfleetTestWorkspace(profileHandler.Profile().GetName(), "cluster-configuration")
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

		state, err := backend.Lifecycle.Read(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.SchedulerProfile).To(Equal(profileHandler.Profile().GetSchedulerProfile()))
	})

	It("configures autoscaler during cluster creation", func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(profileHandler.Profile().GetClusterType()).To(Equal(constants.HYPERFLEET))
		if profileHandler.Profile().GetAutoscalerMaxNodesTotal() == 0 {
			Skip("Test requires the HyperFleet optional configuration profile")
		}

		workspace := hyperfleetTestWorkspace(profileHandler.Profile().GetName(), "cluster-configuration-autoscaler")
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

		state, err := backend.Lifecycle.Read(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Autoscaler).NotTo(BeNil())
		Expect(state.Autoscaler.MaxNodeProvisionTime).To(Equal(profileHandler.Profile().GetAutoscalerMaxNodeProvisionTime()))
		Expect(state.Autoscaler.MaxNodesTotal).To(Equal(profileHandler.Profile().GetAutoscalerMaxNodesTotal()))
		Expect(state.Autoscaler.MaxPodGracePeriod).To(Equal(profileHandler.Profile().GetAutoscalerMaxPodGracePeriod()))
		Expect(state.Autoscaler.PodPriorityThreshold).To(Equal(profileHandler.Profile().GetAutoscalerPodPriorityThreshold()))
	})
})

func hyperfleetTestWorkspace(profileName, suffix string) string {
	clusterName := os.Getenv("RHCS_CLUSTER_NAME")
	if clusterName == "" {
		clusterName = profileName
	}
	return clusterName + "-" + suffix
}
