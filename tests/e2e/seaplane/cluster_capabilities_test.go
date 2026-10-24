// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"encoding/json"
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

// ROSAENG-65243: Image Registry Enable/Disable.
//
// The internal image registry is controlled through the HyperShift
// capabilities passthrough: spec.hostedCluster.capabilities.{enabled,disabled}
// is set at cluster creation only and is immutable afterwards. The rosa CLI
// equivalent is `--disable-cluster-capabilities ImageRegistry`
// (ROSAENG-68175); this test gives the Terraform flow the same coverage.
//
// The test validates the capability setting and cluster readiness — not
// registry pod health. In the ephemeral environment the registry is
// enabled-by-default but not backed by storage (image-registry CO Degraded
// with StorageExists=False), which is out of scope here.
var _ = Describe("Seaplane cluster capabilities id:ROSAENG-65243", ci.Day1, ci.SeaplaneValidated, ci.FeatureCapabilities, func() {
	// requiredProfile is pinned by this spec instead of being selected through
	// the CLUSTER_PROFILE environment variable, so the spec is self-sufficient:
	// it runs with its required configuration regardless of the ambient profile
	// of the suite run. The name must stay within the 18-character Seaplane
	// cluster-name limit because clusterworkflow defaults the cluster name to
	// the profile name.
	const requiredProfile = "sp-no-caps"

	It("creates a cluster with the ImageRegistry capability disabled", func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerByName(requiredProfile)
		Expect(err).NotTo(HaveOccurred())
		Expect(profileHandler.Profile().GetClusterType()).To(Equal(constants.SEAPLANE))
		disabled := profileHandler.Profile().GetDisableClusterCapabilities()
		// The profile is checked into the repo; a missing field is a config bug,
		// not a reason to skip.
		Expect(disabled).NotTo(BeEmpty(),
			"profile %s must define disable_cluster_capabilities", requiredProfile)

		workspace := seaplaneTestWorkspace(requiredProfile, "e2e")
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

		By("Verifying the disabled capabilities via Terraform output")
		clusterService, err := exec.NewSeaplaneClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())
		output, err := clusterService.Output()
		Expect(err).NotTo(HaveOccurred())
		var outputDisabled []string
		if output.DisabledCapabilities != "" {
			Expect(json.Unmarshal([]byte(output.DisabledCapabilities), &outputDisabled)).To(Succeed())
		}
		Expect(outputDisabled).To(ConsistOf(disabled))

		By("Verifying the disabled capabilities through the Platform API")
		seaplaneURL := os.Getenv("SEAPLANE_URL")
		Expect(seaplaneURL).NotTo(BeEmpty())
		// Prefer the region derived from the Platform API URL so the spec works
		// against any regional endpoint; the pinned profile region is the
		// fallback.
		region := awsRegionRE.FindString(seaplaneURL)
		if region == "" {
			region = profileHandler.Profile().GetRegion()
		}
		// The Platform API routes clusters by metadata.name; clusterID is the
		// Terraform identity (metadata.uid), so look up by the cluster name.
		cluster, err := mustBuildSeaplaneClient(seaplaneURL, region).
			SeaplaneV1alpha1().Clusters().Get(ctx, output.ClusterName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		capabilities := cluster.Spec.HostedCluster.Capabilities
		Expect(capabilities).NotTo(BeNil(), "capabilities must be reported on the hosted cluster spec")
		// capabilities.Disabled is a list of named string types; convert to
		// plain strings so gomega's set matching compares the same element
		// types as the profile list.
		var actualDisabled []string
		for _, capability := range capabilities.Disabled {
			actualDisabled = append(actualDisabled, string(capability))
		}
		Expect(actualDisabled).To(ConsistOf(disabled))
		if enabled := profileHandler.Profile().GetEnableClusterCapabilities(); len(enabled) == 0 {
			Expect(capabilities.Enabled).To(BeEmpty(),
				"no capabilities were enabled, so capabilities.enabled must be empty")
		}

		By("Verifying the cluster reaches Ready with the capability disabled")
		Expect(cluster.Status.Phase).To(Equal(v1alpha1.ClusterPhaseReady))
	})
})

// seaplaneTestWorkspace derives the Terraform workspace name for a spec run:
// the cluster name (RHCS_CLUSTER_NAME or the profile name) plus a suffix, so
// each spec's state stays isolated while sharing the cluster-name base.
func seaplaneTestWorkspace(profileName, suffix string) string {
	clusterName := os.Getenv("RHCS_CLUSTER_NAME")
	if clusterName == "" {
		clusterName = profileName
	}
	return clusterName + "-" + suffix
}
