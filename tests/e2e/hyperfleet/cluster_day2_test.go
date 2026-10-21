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
		routes := `[{"namespace":"openshift-console","name":"console","hostname":"console-day2.example.com","servingCertKeyPairSecret":{"name":"console-tls"}}]`
		provisionTime := "20m"
		maxNodes, maxGrace, priority := int64(12), int64(700), int64(-5)
		Expect(backend.Lifecycle.Update(exec.ClusterSpec{
			Autoscaler: &exec.ClusterAutoscalerSpec{
				MaxNodeProvisionTime: &provisionTime,
				MaxNodesTotal:        &maxNodes,
				MaxPodGracePeriod:    &maxGrace,
				PodPriorityThreshold: &priority,
			},
			ComponentRoutes: &routes,
		})).To(Succeed())
		state, err := backend.Lifecycle.Read(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Autoscaler).NotTo(BeNil())
		Expect(state.Autoscaler.MaxNodesTotal).To(Equal(int64(12)))
		Expect(state.Autoscaler.MaxPodGracePeriod).To(Equal(int64(700)))
		Expect(state.Autoscaler.PodPriorityThreshold).To(Equal(int64(-5)))
		Expect(state.Autoscaler.MaxNodeProvisionTime).To(Equal("20m"))
		Expect(state.ComponentRoutes).To(ContainSubstring("console-day2.example.com"))
	})
})
