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

var _ = Describe("Shared cluster autoscaler creation workflow", ci.Day1Prepare, ci.FeatureClusterAutoscaler, ci.SeaplaneValidated, func() {
	It("creates a cluster with the profile autoscaler settings through the selected backend", func(ctx SpecContext) {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		clusterType := profile.Profile().GetClusterType()
		if clusterType != constants.ROSA_HCP && clusterType != constants.SEAPLANE {
			Skip("shared cluster workflow requires a profile with cluster_type to be rosa-hcp or seaplane")
		}
		if profile.Profile().GetAutoscalerMaxNodeProvisionTime() == "" &&
			profile.Profile().GetAutoscalerMaxNodesTotal() == 0 &&
			profile.Profile().GetAutoscalerMaxPodGracePeriod() == 0 &&
			profile.Profile().GetAutoscalerPodPriorityThreshold() == 0 {
			Skip("test requires cluster autoscaler settings in the selected profile")
		}

		workspace := profile.Profile().GetName() + "-shared-cluster-autoscaler"
		backend, err := clusterworkflow.New(profile, token, workspace)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			destroyer, destroyErr := clusterworkflow.NewDestroyer(profile, token, workspace)
			Expect(destroyErr).NotTo(HaveOccurred())
			Expect(destroyer.Destroy(context.Background())).To(Succeed())
		})

		_, err = exec.CreateAndWaitReady(ctx, backend.Lifecycle)
		Expect(err).NotTo(HaveOccurred())
		state, err := backend.Lifecycle.Read(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Autoscaler).NotTo(BeNil())
		Expect(state.Autoscaler.MaxNodeProvisionTime).To(Equal(profile.Profile().GetAutoscalerMaxNodeProvisionTime()))
		Expect(state.Autoscaler.MaxNodesTotal).To(Equal(profile.Profile().GetAutoscalerMaxNodesTotal()))
		Expect(state.Autoscaler.MaxPodGracePeriod).To(Equal(profile.Profile().GetAutoscalerMaxPodGracePeriod()))
		Expect(state.Autoscaler.PodPriorityThreshold).To(Equal(profile.Profile().GetAutoscalerPodPriorityThreshold()))
	})
})
