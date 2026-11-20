// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Seaplane cluster IPv4 networking", ci.SeaplaneValidated, func() {
	It("host prefix and CIDRs are set correctly", ci.Day1Post, ci.Medium, ci.FeatureClusterNetwork, func(ctx SpecContext) {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(profile.Profile().GetClusterType()).To(Equal(constants.SEAPLANE))
		machineCIDR := profile.Profile().GetMachineCIDR()
		serviceCIDR := profile.Profile().GetServiceCIDR()
		podCIDR := profile.Profile().GetPodCIDR()
		hostPrefix := profile.Profile().GetHostPrefix()
		if machineCIDR == "" || serviceCIDR == "" || podCIDR == "" || hostPrefix <= 0 {
			Skip("Test requires a profile containing CIDRs and host prefix")
		}

		seaplaneURL := requireEnv("SEAPLANE_URL")
		region := awsRegionRE.FindString(seaplaneURL)
		if region == "" {
			region = profile.Profile().GetRegion()
		}

		service, err := exec.NewSeaplaneClusterService(profile.Profile().GetName())
		Expect(err).NotTo(HaveOccurred())
		output, err := service.Output()
		Expect(err).NotTo(HaveOccurred())

		clusterName := ""
		clusterID := ""
		if output != nil && output.ClusterID != "" && output.ClusterName != "" {
			clusterName = output.ClusterName
			clusterID = output.ClusterID
		} else {
			By("No existing cluster in profile workspace, creating ephemeral cluster to verify networking")
			Expect(os.Getenv("SEAPLANE_URL")).NotTo(BeEmpty())
			Expect(os.Getenv("SEAPLANE_RELEASE_IMAGE")).NotTo(BeEmpty(), "SEAPLANE_RELEASE_IMAGE is required to create cluster")
			workspace := seaplaneTestWorkspace(profile.Profile().GetName(), "ipv4-net")
			backend, err := clusterworkflow.New(profile, "", workspace)
			Expect(err).NotTo(HaveOccurred())

			DeferCleanup(func() {
				destroyer, destroyErr := clusterworkflow.NewDestroyer(profile, "", workspace)
				Expect(destroyErr).NotTo(HaveOccurred())
				Expect(destroyer.Destroy(context.Background())).To(Succeed())
			})

			clusterID, err = exec.CreateAndWaitReady(ctx, backend.Lifecycle)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterID).NotTo(BeEmpty())

			clusterService, err := exec.NewSeaplaneClusterService(workspace)
			Expect(err).NotTo(HaveOccurred())
			ephemeralOut, err := clusterService.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(ephemeralOut.ClusterID).To(Equal(clusterID))
			clusterName = ephemeralOut.ClusterName
			Expect(clusterName).NotTo(BeEmpty())
		}

		By("Retrieve cluster detail")
		client := mustBuildSeaplaneClient(seaplaneURL, region)
		cluster, err := client.SeaplaneV1alpha1().Clusters().Get(ctx, clusterName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(cluster.UID)).To(Equal(clusterID))

		By("Check cluster networking against the profile")
		network := cluster.Spec.HostedCluster.Networking
		Expect(network.MachineNetwork).To(HaveLen(1))
		Expect(network.ServiceNetwork).To(HaveLen(1))
		Expect(network.ClusterNetwork).To(HaveLen(1))
		Expect(network.MachineNetwork[0].CIDR.String()).To(Equal(machineCIDR))
		Expect(network.ServiceNetwork[0].CIDR.String()).To(Equal(serviceCIDR))
		Expect(network.ClusterNetwork[0].CIDR.String()).To(Equal(podCIDR))
		Expect(int(network.ClusterNetwork[0].HostPrefix)).To(Equal(hostPrefix))
	})
})
