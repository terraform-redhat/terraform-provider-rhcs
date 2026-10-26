// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	. "github.com/onsi/gomega"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

var _ = Describe("NodePool node_drain_timeout", func() {
	It("creates and updates nodepool with node_drain_timeout", func() {
		seaplaneURL := requireEnv("SEAPLANE_URL")
		clusterName := requireEnv("SEAPLANE_CLUSTER_NAME")
		clusterID := requireEnv("SEAPLANE_CLUSTER_ID")
		subnetID := requireEnv("SEAPLANE_SUBNET_ID")
		instanceType := requireEnvWithDefault("SEAPLANE_NODEPOOL_INSTANCE_TYPE", "m5.xlarge")

		awsRegion := awsRegionRE.FindString(seaplaneURL)
		Expect(awsRegion).NotTo(BeEmpty(), fmt.Sprintf("cannot derive AWS region from SEAPLANE_URL %q", seaplaneURL))

		npName := clusterName + ".np-drain-timeout"
		workspace := "test-node-drain-timeout-" + fmt.Sprintf("%d", time.Now().Unix())

		seaplaneClient := mustBuildSeaplaneClient(seaplaneURL, awsRegion)
		nodepools := seaplaneClient.SeaplaneV1alpha1().NodePools()

		npSvc, err := exec.NewSeaplaneNodePoolService(workspace)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func(_ SpecContext) {
			By("Teardown: destroy nodepool with node_drain_timeout")
			_, err := npSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())
		}, GracePeriod(45*time.Minute))

		// ── Phase 1: Create nodepool with 5m drain timeout ────────────────
		By("Phase 1: create nodepool with node_drain_timeout='5m'")
		nodeDrainTimeout := "5m"
		replicas := 1
		_, err = npSvc.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:      &seaplaneURL,
			AWSRegion:        &awsRegion,
			ClusterID:        &clusterID,
			ClusterName:      &clusterName,
			Name:             &npName,
			SubnetID:         &subnetID,
			InstanceType:     &instanceType,
			Replicas:         &replicas,
			NodeDrainTimeout: &nodeDrainTimeout,
		})
		Expect(err).NotTo(HaveOccurred())

		npTFOut, err := npSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(npTFOut.NodePoolID).NotTo(BeEmpty())

		// ── Wait for nodepool to reach Ready ───────────────────────────────
		By("Waiting for nodepool to reach Ready phase")
		ctx := context.Background()
		err = nodepools.WaitUntil(
			ctx, npName,
			func(np *v1alpha1.NodePool) bool {
				if np == nil {
					Logger.Infof("[wait] nodepool %s: not found", npName)
					return false
				}
				Logger.Infof("[wait] nodepool %s phase: %s", npName, np.Status.Phase)
				return np.Status.Phase == v1alpha1.NodePoolPhaseReady
			},
			30*time.Second,
			45*time.Minute,
		)
		Expect(err).NotTo(HaveOccurred())

		// ── Verify node_drain_timeout was set via Platform API ─────────────
		By("Verifying node_drain_timeout='5m' was set via Platform API")
		npObj, err := nodepools.Get(ctx, npName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(npObj.Spec.NodePool.NodeDrainTimeout).To(Equal("5m"))
		Logger.Infof("[verify] nodepool %s has node_drain_timeout='%v'", npName, npObj.Spec.NodePool.NodeDrainTimeout)

		// ── Phase 2: Update drain timeout to 10m ───────────────────────────
		By("Phase 2: update node_drain_timeout to '10m'")
		updatedTimeout := "10m"
		_, err = npSvc.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:      &seaplaneURL,
			AWSRegion:        &awsRegion,
			ClusterID:        &clusterID,
			ClusterName:      &clusterName,
			Name:             &npName,
			SubnetID:         &subnetID,
			InstanceType:     &instanceType,
			Replicas:         &replicas,
			NodeDrainTimeout: &updatedTimeout,
		})
		Expect(err).NotTo(HaveOccurred())

		// ── Verify updated timeout ─────────────────────────────────────────
		By("Verifying node_drain_timeout was updated to '10m' via Platform API")
		err = nodepools.WaitUntil(
			ctx, npName,
			func(np *v1alpha1.NodePool) bool {
				if np == nil {
					return false
				}
				Logger.Infof("[wait] nodepool %s phase: %s, node_drain_timeout: %v",
					npName, np.Status.Phase, np.Spec.NodePool.NodeDrainTimeout)
				return np.Status.Phase == v1alpha1.NodePoolPhaseReady
			},
			30*time.Second,
			10*time.Minute,
		)
		Expect(err).NotTo(HaveOccurred())

		npUpdated, err := nodepools.Get(ctx, npName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(npUpdated.Spec.NodePool.NodeDrainTimeout).To(Equal("10m"))
		Logger.Infof("[verify] nodepool %s node_drain_timeout successfully updated to '%v'",
			npName, npUpdated.Spec.NodePool.NodeDrainTimeout)

		// ── Phase 3: Verify clean plan after update ────────────────────────
		By("Phase 3: verify terraform plan is clean after update")
		_, err = npSvc.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:      &seaplaneURL,
			AWSRegion:        &awsRegion,
			ClusterID:        &clusterID,
			ClusterName:      &clusterName,
			Name:             &npName,
			SubnetID:         &subnetID,
			InstanceType:     &instanceType,
			Replicas:         &replicas,
			NodeDrainTimeout: &updatedTimeout,
		})
		Expect(err).NotTo(HaveOccurred())

		npRefreshed, err := npSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(npRefreshed.Phase).To(Equal(string(v1alpha1.NodePoolPhaseReady)))
		Logger.Infof("[verify] nodepool %s remains in Ready phase after state refresh", npName)
	})
})
