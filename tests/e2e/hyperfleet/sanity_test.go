// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	. "github.com/onsi/gomega"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hfrest "github.com/openshift-online/rosa-hyperfleet-api/clientset/rest"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/hyperfleet-operator/api/v1alpha1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

const (
	clusterReadyTimeout   = 90 * time.Minute
	clusterDeletedTimeout = 60 * time.Minute
	pollInterval          = 30 * time.Second
)

var awsRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

var _ = Describe("Hyperfleet sanity", func() {
	It("creates and destroys a hyperfleet cluster end-to-end", func() {
		hyperfleetURL := requireEnv("HYPERFLEET_URL")
		releaseImage := requireEnv("HYPERFLEET_RELEASE_IMAGE")
		clusterName := requireEnvWithDefault("HYPERFLEET_CLUSTER_NAME", "hf-e2e-sanity")
		operatorRolesPrefix := requireEnvWithDefault("HYPERFLEET_OPERATOR_ROLES_PREFIX", clusterName)
		vpcCIDR := requireEnvWithDefault("HYPERFLEET_VPC_CIDR", "10.0.0.0/16")

		awsRegion := awsRegionRE.FindString(hyperfleetURL)
		Expect(
			awsRegion,
		).NotTo(BeEmpty(), fmt.Sprintf("cannot derive AWS region from HYPERFLEET_URL %q", hyperfleetURL))
		availabilityZone := awsRegion + "a"

		workspace := clusterName

		hfClient := mustBuildHyperfleetClient(hyperfleetURL, awsRegion)
		clusters := hfClient.HyperfleetV1alpha1().Clusters("")

		// ── Create all three services upfront so DeferCleanups can be
		// registered in reverse teardown order before any apply runs.
		// LIFO execution order: cluster → IAM → VPC.
		vpcSvc, err := exec.NewHyperfleetVPCService(workspace)
		Expect(err).NotTo(HaveOccurred())

		iamSvc, err := exec.NewHyperfleetIAMService(workspace)
		Expect(err).NotTo(HaveOccurred())

		clusterSvc, err := exec.NewHyperfleetClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())

		// VPC destroy — registered first, runs last.
		var vpcOut *exec.HyperfleetVPCOutput
		DeferCleanup(func() {
			By("Teardown: destroy VPC")
			if vpcOut != nil {
				Expect(helper.DeleteClassicLoadBalancers(awsRegion, vpcOut.VPCID)).To(Succeed())
				Expect(helper.DeleteNonDefaultSecurityGroups(awsRegion, vpcOut.VPCID)).To(Succeed())
			}
			_, err := vpcSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())
		})

		// IAM destroy — registered second, runs second-to-last (after cluster is gone).
		DeferCleanup(func() {
			By("Teardown: destroy IAM roles")
			_, err := iamSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())
		})

		// Cluster destroy — registered last, runs first.
		var clusterID string
		DeferCleanup(func() {
			By("Teardown: destroy cluster and wait for deletion")
			if clusterID == "" {
				return
			}
			ctx := context.Background()
			_, destroyErr := clusterSvc.Destroy()
			Expect(destroyErr).NotTo(HaveOccurred())

			waitErr := clusters.WaitUntil(
				ctx, clusterID,
				func(c *v1alpha1.Cluster) bool {
					if c == nil {
						Logger.Infof("[wait] cluster %s: deleted", clusterID)
						return true
					}
					Logger.Infof("[wait] cluster %s phase: %s (waiting for deletion)", clusterID, c.Status.Phase)
					return false
				},
				pollInterval, clusterDeletedTimeout,
			)
			Expect(waitErr).NotTo(HaveOccurred())
			Logger.Infof("[teardown] cluster %s deletion confirmed — proceeding with IAM/VPC teardown", clusterID)
		}, GracePeriod(teardownGracePeriod))

		// np1 destroy — registered 5th, runs 2nd.
		// The API accepts the delete and the Terraform state is cleared immediately.
		// The underlying nodepool remains in Deleting phase (PDB prevents eviction)
		// until the cluster deletion cascades through it.
		DeferCleanup(func(ctx SpecContext) {
			By("Teardown: destroy node pool 1 (PDB-blocked; no SDK wait)")
			if clusterID == "" {
				return
			}
			_, err := np1Svc.Destroy()
			Expect(err).NotTo(HaveOccurred())

			// Log the observed phase for diagnostics without blocking on completion.
			nodepools := hfClient.HyperfleetV1alpha1().NodePools()
			_ = nodepools.WaitUntil(
				ctx, np1Name,
				func(np *v1alpha1.NodePool) bool {
					if np == nil {
						return true
					}
					Logger.Infof(
						"[verify] nodepool %s phase after delete request: %s (expected Deleting)",
						np1Name,
						np.Status.Phase,
					)
					return np.Status.Phase == v1alpha1.NodePoolPhaseDeleting
				},
				pollInterval, 2*pollInterval,
			)
		}, GracePeriod(teardownGracePeriod))

		// np2 destroy — registered 6th, runs 1st.
		// No PDB constraint; we confirm deletion via SDK before moving on.
		DeferCleanup(func(ctx SpecContext) {
			By("Teardown: destroy node pool 2 and wait for deletion")
			if clusterID == "" {
				return
			}
			_, err := np2Svc.Destroy()
			Expect(err).NotTo(HaveOccurred())
			waitErr := hfClient.HyperfleetV1alpha1().NodePools().WaitUntil(
				ctx, np2Name,
				func(np *v1alpha1.NodePool) bool {
					if np == nil {
						Logger.Infof("[wait] nodepool %s: deleted", np2Name)
						return true
					}
					Logger.Infof("[wait] nodepool %s phase: %s (waiting for deletion)", np2Name, np.Status.Phase)
					return false
				},
				pollInterval, nodepoolReadyTimeout,
			)
			Expect(waitErr).NotTo(HaveOccurred())
		}, GracePeriod(teardownGracePeriod))

		// ── Phase 1: VPC ───────────────────────────────────────────────────
		By("Phase 1: apply VPC")
		_, err = vpcSvc.Apply(&exec.HyperfleetVPCArgs{
			AWSRegion:        &awsRegion,
			NamePrefix:       &clusterName,
			VPCCIDR:          &vpcCIDR,
			AvailabilityZone: &availabilityZone,
		})
		Expect(err).NotTo(HaveOccurred())

		vpcOut, err = vpcSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(vpcOut.VPCID).NotTo(BeEmpty())
		Expect(vpcOut.PrivateSubnetID).NotTo(BeEmpty())

		// ── Phase 2: cluster ───────────────────────────────────────────────
		By("Phase 2: apply cluster")
		expiresAt := time.Now().UTC().Add(4 * time.Hour).Format(time.RFC3339)
		_, err = clusterSvc.Apply(&exec.HyperfleetClusterArgs{
			HyperfleetURL:       &hyperfleetURL,
			AWSRegion:           &awsRegion,
			ClusterName:         &clusterName,
			OperatorRolesPrefix: &operatorRolesPrefix,
			SubnetID:            &vpcOut.PrivateSubnetID,
			VPCID:               &vpcOut.VPCID,
			AvailabilityZone:    &availabilityZone,
			ReleaseImage:        &releaseImage,
			ExpirationTimestamp: &expiresAt,
		})
		Expect(err).NotTo(HaveOccurred())

		clusterTFOut, err := clusterSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterTFOut.ClusterID).NotTo(BeEmpty())
		Expect(clusterTFOut.OIDCIssuer).NotTo(BeEmpty())

		clusterID = clusterTFOut.ClusterID
		oidcIssuerURL := clusterTFOut.OIDCIssuer

		// ── Phase 3: IAM roles + OIDC provider ────────────────────────────
		By("Phase 3: apply IAM roles")
		_, err = iamSvc.Apply(&exec.HyperfleetIAMArgs{
			AWSRegion:           &awsRegion,
			OperatorRolesPrefix: &operatorRolesPrefix,
			OIDCIssuerURL:       &oidcIssuerURL,
		})
		Expect(err).NotTo(HaveOccurred())

		iamOut, err := iamSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(iamOut.OIDCProviderARN).NotTo(BeEmpty())
		Expect(iamOut.WorkerRoleARN).NotTo(BeEmpty())

		// ── Wait for cluster Ready ─────────────────────────────────────────
		By("Waiting for cluster to reach Ready phase")
		ctx := context.Background()
		err = clusters.WaitUntil(
			ctx, clusterID,
			func(c *v1alpha1.Cluster) bool {
				if c == nil {
					Logger.Infof("[wait] cluster %s: not found", clusterID)
					return false
				}
				Logger.Infof("[wait] cluster %s phase: %s", clusterID, c.Status.Phase)
				return c.Status.Phase == v1alpha1.ClusterPhaseReady
			},
			pollInterval, clusterReadyTimeout,
		)
		Expect(err).NotTo(HaveOccurred())

		// ── Verify via Terraform refresh ───────────────────────────────────
		By("Verifying cluster Ready phase via Terraform state refresh")
		_, err = clusterSvc.Apply(&exec.HyperfleetClusterArgs{
			HyperfleetURL:       &hyperfleetURL,
			AWSRegion:           &awsRegion,
			ClusterName:         &clusterName,
			OperatorRolesPrefix: &operatorRolesPrefix,
			SubnetID:            &vpcOut.PrivateSubnetID,
			VPCID:               &vpcOut.VPCID,
			AvailabilityZone:    &availabilityZone,
			ReleaseImage:        &releaseImage,
			ExpirationTimestamp: &expiresAt,
		})
		Expect(err).NotTo(HaveOccurred())

		refreshedOut, err := clusterSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(refreshedOut.Phase).To(Equal(string(v1alpha1.ClusterPhaseReady)))
		Expect(refreshedOut.APIURL).NotTo(BeEmpty())
	})
})

func mustBuildHyperfleetClient(hfURL, region string) *hyperfleet.Clientset {
	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	Expect(err).NotTo(HaveOccurred())

	identity, err := sts.NewFromConfig(awsCfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	Expect(err).NotTo(HaveOccurred())

	client, err := hyperfleet.NewForConfig(&hfrest.Config{
		Host:      hfURL,
		Region:    region,
		AccountID: *identity.Account,
		CallerARN: *identity.Arn,
		AWSConfig: awsCfg,
	})
	Expect(err).NotTo(HaveOccurred())
	return client
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	Expect(val).NotTo(BeEmpty(), fmt.Sprintf("required env var %s is not set", key))
	return val
}

func requireEnvWithDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
