// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

// ROSAENG-68098 / ROSAENG-65120: custom network CIDRs.
//
// Machine, service, and pod network CIDRs plus host prefix are set at cluster
// creation only (immutable afterwards). The rosa CLI equivalent is
// --machine-cidr/--pod-cidr/--service-cidr/--host-prefix; V1 TF analogs are
// the negative day-one overlap cases and the cluster_edit immutability
// assertions. Constraints: with the byovpc flow machine_cidr must contain the
// VPC CIDR (SEAPLANE_VPC_CIDR, default 10.0.0.0/16), and the pod/service
// CIDRs must not overlap it or each other.
var _ = Describe("Seaplane cluster network CIDRs", ci.Day1, ci.SeaplaneValidated, func() {
	It("creates a cluster with custom machine, pod, and service CIDRs", func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(profileHandler.Profile().GetClusterType()).To(Equal(constants.SEAPLANE))
		machineCIDR := profileHandler.Profile().GetMachineCIDR()
		podCIDR := profileHandler.Profile().GetPodCIDR()
		serviceCIDR := profileHandler.Profile().GetServiceCIDR()
		if machineCIDR == "" && podCIDR == "" && serviceCIDR == "" {
			Skip("Test requires a Seaplane profile with custom network CIDRs")
		}
		hostPrefix := profileHandler.Profile().GetHostPrefix()

		workspace := profileHandler.Profile().GetName() + "-custom-cidrs"
		backend, err := clusterworkflow.New(profileHandler, "", workspace)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			destroyer, destroyErr := clusterworkflow.NewDestroyer(profileHandler, "", workspace)
			Expect(destroyErr).NotTo(HaveOccurred())
			Expect(destroyer.Destroy(context.Background())).To(Succeed())
		})

		clusterID, err := exec.CreateAndWaitReady(ctx, backend.Lifecycle)
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterID).NotTo(BeEmpty())

		By("Verifying the CIDRs via Terraform output")
		clusterService, err := exec.NewSeaplaneClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())
		output, err := clusterService.Output()
		Expect(err).NotTo(HaveOccurred())
		if machineCIDR != "" {
			Expect(output.MachineCIDR).To(Equal(machineCIDR))
		}
		if podCIDR != "" {
			Expect(output.PodCIDR).To(Equal(podCIDR))
		}
		if serviceCIDR != "" {
			Expect(output.ServiceCIDR).To(Equal(serviceCIDR))
		}

		By("Verifying the networking through the Platform API")
		seaplaneURL := os.Getenv("SEAPLANE_URL")
		Expect(seaplaneURL).NotTo(BeEmpty())
		// Prefer the region derived from the Platform API URL so the spec works
		// against any regional endpoint; the profile region is the fallback.
		region := awsRegionRE.FindString(seaplaneURL)
		if region == "" {
			region = profileHandler.Profile().GetRegion()
		}
		// The Platform API routes clusters by metadata.name; clusterID is the
		// Terraform identity (metadata.uid), so look up by the cluster name.
		cluster, err := mustBuildSeaplaneClient(seaplaneURL, region).
			SeaplaneV1alpha1().Clusters().Get(ctx, output.ClusterName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		networking := cluster.Spec.HostedCluster.Networking
		if machineCIDR != "" {
			Expect(networking.MachineNetwork).To(HaveLen(1))
			Expect(networking.MachineNetwork[0].CIDR.String()).To(Equal(machineCIDR))
		}
		if podCIDR != "" {
			Expect(networking.ClusterNetwork).To(HaveLen(1))
			Expect(networking.ClusterNetwork[0].CIDR.String()).To(Equal(podCIDR))
			if hostPrefix != 0 {
				Expect(int(networking.ClusterNetwork[0].HostPrefix)).To(Equal(hostPrefix))
			}
		}
		if serviceCIDR != "" {
			Expect(networking.ServiceNetwork).To(HaveLen(1))
			Expect(networking.ServiceNetwork[0].CIDR.String()).To(Equal(serviceCIDR))
		}

		By("Verifying the cluster reaches Ready with the custom networking")
		// CreateAndWaitReady already blocked until Ready; a misconfigured
		// network would have failed provisioning.
		Expect(cluster.Status.Phase).To(Equal(v1alpha1.ClusterPhaseReady))
	})
})
