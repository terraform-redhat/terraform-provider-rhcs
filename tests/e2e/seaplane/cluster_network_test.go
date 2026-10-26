// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
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

		service, err := exec.NewSeaplaneClusterService(profile.Profile().GetName())
		Expect(err).NotTo(HaveOccurred())
		output, err := service.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(output.ClusterID).NotTo(BeEmpty(), "an existing profile-managed cluster is required")
		Expect(output.ClusterName).NotTo(BeEmpty())

		By("Retrieve cluster detail")
		client := mustBuildSeaplaneClient(requireEnv("SEAPLANE_URL"), profile.Profile().GetRegion())
		cluster, err := client.SeaplaneV1alpha1().Clusters().Get(ctx, output.ClusterName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(cluster.UID)).To(Equal(output.ClusterID))

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
