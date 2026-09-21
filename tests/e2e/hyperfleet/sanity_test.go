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
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	hfrest "github.com/openshift-online/rosa-hyperfleet-api/clientset/rest"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

const (
	clusterReadyTimeout    = 90 * time.Minute
	clusterDeletedTimeout  = 60 * time.Minute
	nodepoolReadyTimeout   = 45 * time.Minute
	oidcConfigReadyTimeout = 15 * time.Minute
	pollInterval           = 30 * time.Second
	vpcDestroyTimeout      = 30 * time.Minute

	// teardownGracePeriod bounds how long Ginkgo lets a cleanup node keep running
	// after an interrupt (Ctrl-C) before force-aborting it. Without an override,
	// Ginkgo aborts each cleanup node at the default 30s grace period — far too
	// short for cluster/nodepool/VPC deletion, so teardown is killed mid-wait
	// with [TIMEDOUT] and AWS resources are orphaned. This is sized to the
	// longest teardown wait (cluster deletion) plus the terraform destroy
	// latency that precedes it. A grace period this large is harmless: a node
	// still returns as soon as its own work completes — the value only caps how
	// long Ginkgo waits before force-aborting. A second interrupt still
	// force-skips all remaining cleanup.
	teardownGracePeriod = clusterDeletedTimeout + 30*time.Minute
)

var awsRegionRE = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

var _ = Describe("Hyperfleet sanity", func() {
	It("creates and destroys a hyperfleet cluster end-to-end", func() {
		hyperfleetURL := requireEnv("HYPERFLEET_URL")
		clusterName := requireEnvWithDefault("HYPERFLEET_CLUSTER_NAME", fmt.Sprintf("hf-e2e-%d", time.Now().Unix()))
		operatorRolesPrefix := requireEnvWithDefault("HYPERFLEET_OPERATOR_ROLES_PREFIX", clusterName)
		vpcCIDR := requireEnvWithDefault("HYPERFLEET_VPC_CIDR", "10.0.0.0/16")
		instanceType := requireEnvWithDefault("HYPERFLEET_NODEPOOL_INSTANCE_TYPE", "m5.xlarge")

		awsRegion := awsRegionRE.FindString(hyperfleetURL)
		Expect(awsRegion).NotTo(BeEmpty(), fmt.Sprintf("cannot derive AWS region from HYPERFLEET_URL %q", hyperfleetURL))
		availabilityZone := awsRegion + "a"

		// np1: blocked in Deleting phase by a PDB on the cluster. The hyperfleet
		// API accepts the delete request (Terraform state is cleared immediately),
		// but the underlying nodepool stays in Deleting until the cluster itself
		// is deleted — at which point the operator cascades the cleanup.
		// np2: no PDB constraint; deletion completes and is confirmed via SDK.
		np1Name := clusterName + "-np1"
		np2Name := clusterName + "-np2"

		workspace := clusterName

		hfClient := mustBuildHyperfleetClient(hyperfleetURL, awsRegion)
		clusters := hfClient.HyperfleetV1alpha1().Clusters()

		// ── Create all services upfront so DeferCleanups can be registered in
		// reverse teardown order before any apply runs.
		// LIFO execution: np2 → np1 → cluster → IAM → OidcConfig → VPC.
		vpcSvc, err := exec.NewHyperfleetVPCService(workspace)
		Expect(err).NotTo(HaveOccurred())

		oidcSvc, err := exec.NewHyperfleetOidcConfigService(workspace)
		Expect(err).NotTo(HaveOccurred())

		iamSvc, err := exec.NewHyperfleetIAMService(workspace)
		Expect(err).NotTo(HaveOccurred())

		clusterSvc, err := exec.NewHyperfleetClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())

		// Each nodepool uses its own workspace so their Terraform states are isolated.
		np1Svc, err := exec.NewHyperfleetNodePoolService(workspace + "-np1")
		Expect(err).NotTo(HaveOccurred())

		np2Svc, err := exec.NewHyperfleetNodePoolService(workspace + "-np2")
		Expect(err).NotTo(HaveOccurred())

		// VPC destroy — registered 1st, runs last.
		var vpcOut *exec.HyperfleetVPCOutput
		DeferCleanup(func(ctx SpecContext) {
			By("Teardown: destroy VPC")
			// The cluster operator releases VPC resources asynchronously after the
			// cluster object is deleted, and it leaks resources Terraform does not
			// track (classic ELBs, VPC endpoints, security groups, hosted-zone
			// records) that block terraform destroy. Re-prune them on every retry —
			// not just once — so transient states (e.g. a vpce-private-router
			// security group still pinned by an endpoint ENI that has not finished
			// releasing) resolve within the retry window instead of failing the
			// whole teardown on the first attempt.
			Eventually(func() error {
				if vpcOut != nil {
					if err := helper.DeleteClassicLoadBalancers(awsRegion, vpcOut.VPCID); err != nil {
						Logger.Infof("[teardown] classic ELB cleanup failed (will retry): %v", err)
						return err
					}
					// Delete VPC endpoints before security groups: their managed ENIs
					// pin the <infra-id>-vpce-private-router security group, so the SG
					// cannot be removed until the endpoints are gone and their ENIs
					// have drained.
					if err := helper.DeleteVPCEndpoints(awsRegion, vpcOut.VPCID); err != nil {
						Logger.Infof("[teardown] VPC endpoint cleanup failed (will retry): %v", err)
						return err
					}
					if err := helper.DeleteNonDefaultSecurityGroups(awsRegion, vpcOut.VPCID); err != nil {
						Logger.Infof("[teardown] security group cleanup failed (will retry): %v", err)
						return err
					}
					// The cluster operator writes CNAME records (api.*, *.apps.*) into
					// the private <name>.hypershift.local hosted zone. terraform destroy
					// of aws_route53_zone.hyperfleet fails with HostedZoneNotEmpty while
					// those records remain, so purge them before destroying the VPC.
					if err := helper.PurgeHostedZoneRecords(awsRegion, vpcOut.HostedZoneID); err != nil {
						Logger.Infof("[teardown] hosted zone purge failed (will retry): %v", err)
						return err
					}
				}
				_, err := vpcSvc.Destroy()
				if err != nil {
					Logger.Infof("[teardown] VPC destroy attempt failed (will retry): %v", err)
				}
				return err
			}).WithContext(ctx).WithTimeout(vpcDestroyTimeout).WithPolling(2 * time.Minute).Should(Succeed())
		}, GracePeriod(teardownGracePeriod))

		// OidcConfig destroy — registered 2nd, runs 4th.
		// OidcConfig is independent of cluster and can be deleted at any time.
		DeferCleanup(func(_ SpecContext) {
			By("Teardown: destroy OIDC config")
			_, err := oidcSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())
		}, GracePeriod(teardownGracePeriod))

		// IAM destroy — registered 3rd, runs 5th.
		// Accepts SpecContext (unused) because GracePeriod requires a
		// context-accepting callback, even though this teardown does not wait.
		DeferCleanup(func(_ SpecContext) {
			By("Teardown: destroy IAM roles")
			_, err := iamSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())
		}, GracePeriod(teardownGracePeriod))

		// Cluster destroy — registered 4th, runs 3rd.
		// The operator cascades deletion to any nodepool still in Deleting phase
		// (e.g. np1), so no extra handling is needed here.
		var clusterID string
		DeferCleanup(func(ctx SpecContext) {
			By("Teardown: destroy cluster and wait for deletion")
			if clusterID == "" {
				return
			}
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
			nodepools := hfClient.HyperfleetV1alpha1().NodePools(clusterID)
			_ = nodepools.WaitUntil(
				ctx, np1Name,
				func(np *v1alpha1.NodePool) bool {
					if np == nil {
						return true
					}
					Logger.Infof("[verify] nodepool %s phase after delete request: %s (expected Deleting)", np1Name, np.Status.Phase)
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

			waitErr := hfClient.HyperfleetV1alpha1().NodePools(clusterID).WaitUntil(
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

		// ── Phase 2: OidcConfig ─────────────────────────────────────────────
		// OidcConfig (managed mode) computes its own issuer URL.
		By("Phase 2: apply OidcConfig (managed mode)")
		oidcConfigName := clusterName + "-oidc"
		_, err = oidcSvc.Apply(&exec.HyperfleetOidcConfigArgs{
			HyperfleetURL: &hyperfleetURL,
			Name:          &oidcConfigName,
		})
		Expect(err).NotTo(HaveOccurred())

		oidcOut, err := oidcSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(oidcOut.OidcConfigID).NotTo(BeEmpty(), "OidcConfig should have an ID after creation")
		Expect(oidcOut.IssuerURL).NotTo(BeEmpty(), "OidcConfig should compute issuer URL in managed mode")
		// Phase will be Pending until cluster is created and associated
		Logger.Infof("[verify] OidcConfig %s created: issuer_url=%s, phase=%s", oidcOut.OidcConfigID, oidcOut.IssuerURL, oidcOut.Phase)

		// ── Phase 3: IAM roles + OIDC provider ────────────────────────────
		// Use the OidcConfig issuer URL to create IAM resources
		// Thumbprint will be available once OidcConfig reaches Ready after cluster association
		By("Phase 3: apply IAM roles")
		_, err = iamSvc.Apply(&exec.HyperfleetIAMArgs{
			AWSRegion:           &awsRegion,
			OperatorRolesPrefix: &operatorRolesPrefix,
			OIDCIssuerURL:       &oidcOut.IssuerURL, // Use OidcConfig computed issuer URL
		})
		Expect(err).NotTo(HaveOccurred())

		iamOut, err := iamSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(iamOut.OIDCProviderARN).NotTo(BeEmpty())
		Expect(iamOut.WorkerRoleARN).NotTo(BeEmpty())

		// ── Phase 4: cluster ───────────────────────────────────────────────
		// Cluster references the OidcConfig by ID
		By("Phase 4: apply cluster")
		expiresAt := time.Now().UTC().Add(4 * time.Hour).Format(time.RFC3339)
		_, err = clusterSvc.Apply(&exec.HyperfleetClusterArgs{
			HyperfleetURL:       &hyperfleetURL,
			AWSRegion:           &awsRegion,
			ClusterName:         &clusterName,
			OperatorRolesPrefix: &operatorRolesPrefix,
			SubnetID:            &vpcOut.PrivateSubnetID,
			VPCID:               &vpcOut.VPCID,
			AvailabilityZone:    &availabilityZone,
			ExpirationTimestamp: &expiresAt,
			OIDCConfigID:        &oidcOut.OidcConfigID, // Reference the OidcConfig
		})
		Expect(err).NotTo(HaveOccurred())

		clusterTFOut, err := clusterSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterTFOut.ClusterID).NotTo(BeEmpty())

		clusterID = clusterTFOut.ClusterID

		// ── Wait for OidcConfig Ready (after cluster association) ─────────
		// OidcConfig only reaches Ready once associated with the cluster
		By("Waiting for OidcConfig to reach Ready phase after cluster association")
		oidcConfigs := hfClient.HyperfleetV1alpha1().OidcConfigs()
		ctx := context.Background()
		err = oidcConfigs.WaitUntil(
			ctx, oidcOut.OidcConfigID,
			func(oc *v1alpha1.OidcConfig) bool {
				if oc == nil {
					Logger.Infof("[wait] oidcconfig %s: not found", oidcOut.OidcConfigID)
					return false
				}
				Logger.Infof("[wait] oidcconfig %s phase: %s", oidcOut.OidcConfigID, oc.Status.Phase)
				return oc.Status.Phase == v1alpha1.OidcConfigPhaseReady
			},
			pollInterval, oidcConfigReadyTimeout,
		)
		Expect(err).NotTo(HaveOccurred())

		// Refresh OidcConfig state to get computed thumbprint
		_, err = oidcSvc.Apply(&exec.HyperfleetOidcConfigArgs{
			HyperfleetURL: &hyperfleetURL,
			Name:          &oidcConfigName,
		})
		Expect(err).NotTo(HaveOccurred())

		oidcRefreshed, err := oidcSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(oidcRefreshed.Phase).To(Equal(string(v1alpha1.OidcConfigPhaseReady)), "OidcConfig phase should be Ready after cluster association")
		Expect(oidcRefreshed.Thumbprint).NotTo(BeEmpty(), "OidcConfig thumbprint should be computed after phase is Ready")
		Logger.Infof("[verify] OidcConfig %s Ready: thumbprint=%s", oidcRefreshed.OidcConfigID, oidcRefreshed.Thumbprint)

		// ── Wait for cluster Ready ─────────────────────────────────────────
		By("Waiting for cluster to reach Ready phase")
		clustersCli := hfClient.HyperfleetV1alpha1().Clusters()
		err = clustersCli.WaitUntil(
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

		// ── Verify cluster via Terraform refresh ───────────────────────────
		By("Verifying cluster Ready phase via Terraform state refresh")
		_, err = clusterSvc.Apply(&exec.HyperfleetClusterArgs{
			HyperfleetURL:       &hyperfleetURL,
			AWSRegion:           &awsRegion,
			ClusterName:         &clusterName,
			OperatorRolesPrefix: &operatorRolesPrefix,
			SubnetID:            &vpcOut.PrivateSubnetID,
			VPCID:               &vpcOut.VPCID,
			AvailabilityZone:    &availabilityZone,
			ExpirationTimestamp: &expiresAt,
			OIDCConfigID:        &oidcRefreshed.OidcConfigID, // Reference the OidcConfig
		})
		Expect(err).NotTo(HaveOccurred())

		refreshedOut, err := clusterSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(refreshedOut.Phase).To(Equal(string(v1alpha1.ClusterPhaseReady)))
		Expect(refreshedOut.APIURL).NotTo(BeEmpty())

		// ── Phase 5: node pools ────────────────────────────────────────────
		By("Phase 5a: apply node pool 1")
		np1Replicas := 2
		np2Replicas := 1
		_, err = np1Svc.Apply(&exec.HyperfleetNodePoolArgs{
			HyperfleetURL: &hyperfleetURL,
			AWSRegion:     &awsRegion,
			ClusterID:     &clusterID,
			ClusterName:   &clusterName,
			Name:          &np1Name,
			SubnetID:      &vpcOut.PrivateSubnetID,
			InstanceType:  &instanceType,
			Replicas:      &np1Replicas,
		})
		Expect(err).NotTo(HaveOccurred())

		np1TFOut, err := np1Svc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(np1TFOut.NodePoolID).NotTo(BeEmpty())

		By("Phase 5b: apply node pool 2")
		_, err = np2Svc.Apply(&exec.HyperfleetNodePoolArgs{
			HyperfleetURL: &hyperfleetURL,
			AWSRegion:     &awsRegion,
			ClusterID:     &clusterID,
			ClusterName:   &clusterName,
			Name:          &np2Name,
			SubnetID:      &vpcOut.PrivateSubnetID,
			InstanceType:  &instanceType,
			Replicas:      &np2Replicas,
		})
		Expect(err).NotTo(HaveOccurred())

		np2TFOut, err := np2Svc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(np2TFOut.NodePoolID).NotTo(BeEmpty())

		// ── Wait for both node pools to reach Ready ────────────────────────
		By("Waiting for node pool 1 to reach Ready phase")
		nodepools := hfClient.HyperfleetV1alpha1().NodePools(clusterID)
		err = nodepools.WaitUntil(
			ctx, np1Name,
			func(np *v1alpha1.NodePool) bool {
				if np == nil {
					Logger.Infof("[wait] nodepool %s: not found", np1Name)
					return false
				}
				Logger.Infof("[wait] nodepool %s phase: %s", np1Name, np.Status.Phase)
				return np.Status.Phase == v1alpha1.NodePoolPhaseReady
			},
			pollInterval, nodepoolReadyTimeout,
		)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for node pool 2 to reach Ready phase")
		err = nodepools.WaitUntil(
			ctx, np2Name,
			func(np *v1alpha1.NodePool) bool {
				if np == nil {
					Logger.Infof("[wait] nodepool %s: not found", np2Name)
					return false
				}
				Logger.Infof("[wait] nodepool %s phase: %s", np2Name, np.Status.Phase)
				return np.Status.Phase == v1alpha1.NodePoolPhaseReady
			},
			pollInterval, nodepoolReadyTimeout,
		)
		Expect(err).NotTo(HaveOccurred())

		// ── Verify node pools via Terraform refresh ────────────────────────
		By("Verifying node pool 1 Ready phase via Terraform state refresh")
		_, err = np1Svc.Apply(&exec.HyperfleetNodePoolArgs{
			HyperfleetURL: &hyperfleetURL,
			AWSRegion:     &awsRegion,
			ClusterID:     &clusterID,
			ClusterName:   &clusterName,
			Name:          &np1Name,
			SubnetID:      &vpcOut.PrivateSubnetID,
			InstanceType:  &instanceType,
			Replicas:      &np1Replicas,
		})
		Expect(err).NotTo(HaveOccurred())

		np1Refreshed, err := np1Svc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(np1Refreshed.Phase).To(Equal(string(v1alpha1.NodePoolPhaseReady)))

		By("Verifying node pool 2 Ready phase via Terraform state refresh")
		_, err = np2Svc.Apply(&exec.HyperfleetNodePoolArgs{
			HyperfleetURL: &hyperfleetURL,
			AWSRegion:     &awsRegion,
			ClusterID:     &clusterID,
			ClusterName:   &clusterName,
			Name:          &np2Name,
			SubnetID:      &vpcOut.PrivateSubnetID,
			InstanceType:  &instanceType,
			Replicas:      &np2Replicas,
		})
		Expect(err).NotTo(HaveOccurred())

		np2Refreshed, err := np2Svc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(np2Refreshed.Phase).To(Equal(string(v1alpha1.NodePoolPhaseReady)))

		// ── Phase 5: scale np2 replicas ────────────────────────────────────
		By("Phase 5: scale node pool 2 from 1 to 2 replicas")
		scaledReplicas := 2
		_, err = np2Svc.Apply(&exec.HyperfleetNodePoolArgs{
			HyperfleetURL: &hyperfleetURL,
			AWSRegion:     &awsRegion,
			ClusterID:     &clusterID,
			ClusterName:   &clusterName,
			Name:          &np2Name,
			SubnetID:      &vpcOut.PrivateSubnetID,
			InstanceType:  &instanceType,
			Replicas:      &scaledReplicas,
		})
		Expect(err).NotTo(HaveOccurred())

		np2Scaled, err := np2Svc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(np2Scaled.Replicas).NotTo(BeNil())
		Expect(*np2Scaled.Replicas).To(Equal(scaledReplicas))

		By("Verifying node pool 2 scaled replicas through the Platform API")
		err = nodepools.WaitUntil(
			ctx, np2Name,
			func(np *v1alpha1.NodePool) bool {
				if np == nil || np.Spec.NodePool.Replicas == nil {
					return false
				}
				Logger.Infof("[wait] nodepool %s phase: %s, replicas: %d", np2Name, np.Status.Phase, *np.Spec.NodePool.Replicas)
				return np.Status.Phase == v1alpha1.NodePoolPhaseReady &&
					*np.Spec.NodePool.Replicas == int32(scaledReplicas)
			},
			pollInterval, nodepoolReadyTimeout,
		)
		Expect(err).NotTo(HaveOccurred())
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
