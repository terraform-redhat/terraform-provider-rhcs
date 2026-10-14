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

var _ = Describe("HyperFleet cluster Day 2 operations", ci.HyperfleetValidated, func() {
	It("updates autoscaler and component routes as a day-2 operation", func(ctx SpecContext) {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if profile.Profile().GetClusterType() != constants.HYPERFLEET {
			Skip("Day-2 test requires a HyperFleet profile")
		}
		workspace := hyperfleetTestWorkspace(profile.Profile().GetName(), "autoscaler-ingress-day2")
		backend, err := clusterworkflow.New(profile, "", workspace)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			destroyer, cleanupErr := clusterworkflow.NewDestroyer(profile, "", workspace)
			Expect(cleanupErr).NotTo(HaveOccurred())
			Expect(destroyer.Destroy(context.Background())).To(Succeed())
		})
		_, err = exec.CreateAndWaitReady(ctx, backend.Lifecycle)
		Expect(err).NotTo(HaveOccurred())
		updater, ok := backend.Lifecycle.(interface {
			Update(*exec.HyperfleetClusterArgs) error
		})
		Expect(ok).To(BeTrue())
		routes := `[{"namespace":"openshift-console","name":"console","hostname":"console-day2.example.com","servingCertKeyPairSecret":{"name":"console-tls"}}]`
		provisionTime := "20m"
		maxNodes, maxGrace, priority := int64(12), int64(700), int64(-5)
		Expect(updater.Update(&exec.HyperfleetClusterArgs{
			AutoscalerMaxNodeProvisionTime: &provisionTime,
			AutoscalerMaxNodesTotal:        &maxNodes,
			AutoscalerMaxPodGracePeriod:    &maxGrace,
			AutoscalerPodPriorityThreshold: &priority,
			ComponentRoutes:                &routes,
		})).To(Succeed())
		service, err := exec.NewHyperfleetClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())
		output, err := service.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(output.AutoscalerMaxNodesTotal).To(Equal(int64(12)))
		Expect(output.AutoscalerMaxPodGracePeriod).To(Equal(int64(700)))
		Expect(output.AutoscalerPodPriorityThreshold).To(Equal(int64(-5)))
		Expect(output.AutoscalerMaxNodeProvisionTime).To(Equal("20m"))
		Expect(output.ComponentRoutes).To(ContainSubstring("console-day2.example.com"))
	})
})
