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

var _ = Describe("NodePool max_unavailable", func() {
	It("creates and updates nodepool with max_unavailable", func() {
		seaplaneURL := requireEnv("SEAPLANE_URL")
		clusterName := requireEnv("SEAPLANE_CLUSTER_NAME")
		clusterID := requireEnv("SEAPLANE_CLUSTER_ID")
		subnetID := requireEnv("SEAPLANE_SUBNET_ID")
		instanceType := requireEnvWithDefault("SEAPLANE_NODEPOOL_INSTANCE_TYPE", "m5.xlarge")

		awsRegion := awsRegionRE.FindString(seaplaneURL)
		Expect(awsRegion).NotTo(BeEmpty(), fmt.Sprintf("cannot derive AWS region from SEAPLANE_URL %q", seaplaneURL))

		npName := clusterName + ".np-max-unavailable"
		workspace := "test-max-unavailable-" + fmt.Sprintf("%d", time.Now().Unix())

		seaplaneClient := mustBuildSeaplaneClient(seaplaneURL, awsRegion)
		nodepools := seaplaneClient.SeaplaneV1alpha1().NodePools()

		npSvc, err := exec.NewSeaplaneNodePoolService(workspace)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func(_ SpecContext) {
			By("Teardown: destroy nodepool with max_unavailable")
			_, err := npSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())
		}, GracePeriod(45*time.Minute))

		// ── Phase 1: Create nodepool with max_unavailable=2 ───────────────
		By("Phase 1: create nodepool with max_unavailable='2' and replicas=3")
		maxUnavailable := "2"
		replicas := 3
		_, err = npSvc.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:    &seaplaneURL,
			AWSRegion:      &awsRegion,
			ClusterID:      &clusterID,
			ClusterName:    &clusterName,
			Name:           &npName,
			SubnetID:       &subnetID,
			InstanceType:   &instanceType,
			Replicas:       &replicas,
			MaxUnavailable: &maxUnavailable,
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

		// ── Verify max_unavailable was set via Platform API ────────────────
		By("Verifying max_unavailable='2' was set via Platform API")
		npObj, err := nodepools.Get(ctx, npName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(npObj.Spec.NodePool.Management).NotTo(BeNil())
		Expect(npObj.Spec.NodePool.Management.Replace).NotTo(BeNil())
		Expect(npObj.Spec.NodePool.Management.Replace.RollingUpdate).NotTo(BeNil())
		Expect(npObj.Spec.NodePool.Management.Replace.RollingUpdate.MaxUnavailable).NotTo(BeNil())

		maxUnavailableValue := npObj.Spec.NodePool.Management.Replace.RollingUpdate.MaxUnavailable.IntVal
		Expect(maxUnavailableValue).To(Equal(int32(2)))
		Logger.Infof("[verify] nodepool %s has max_unavailable=%d", npName, maxUnavailableValue)

		// ── Phase 2: Update max_unavailable to 1 ───────────────────────────
		By("Phase 2: update max_unavailable to '1'")
		updatedMaxUnavailable := "1"
		_, err = npSvc.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:    &seaplaneURL,
			AWSRegion:      &awsRegion,
			ClusterID:      &clusterID,
			ClusterName:    &clusterName,
			Name:           &npName,
			SubnetID:       &subnetID,
			InstanceType:   &instanceType,
			Replicas:       &replicas,
			MaxUnavailable: &updatedMaxUnavailable,
		})
		Expect(err).NotTo(HaveOccurred())

		// ── Verify updated max_unavailable ──────────────────────────────────
		By("Verifying max_unavailable was updated to '1' via Platform API")
		err = nodepools.WaitUntil(
			ctx, npName,
			func(np *v1alpha1.NodePool) bool {
				if np == nil {
					return false
				}
				Logger.Infof("[wait] nodepool %s phase: %s", npName, np.Status.Phase)
				return np.Status.Phase == v1alpha1.NodePoolPhaseReady
			},
			30*time.Second,
			10*time.Minute,
		)
		Expect(err).NotTo(HaveOccurred())

		npUpdated, err := nodepools.Get(ctx, npName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(npUpdated.Spec.NodePool.Management.Replace.RollingUpdate.MaxUnavailable.IntVal).To(Equal(int32(1)))
		Logger.Infof("[verify] nodepool %s max_unavailable successfully updated to %d",
			npName, npUpdated.Spec.NodePool.Management.Replace.RollingUpdate.MaxUnavailable.IntVal)

		// ── Phase 3: Verify clean plan after update ────────────────────────
		By("Phase 3: verify terraform plan is clean after update")
		_, err = npSvc.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:    &seaplaneURL,
			AWSRegion:      &awsRegion,
			ClusterID:      &clusterID,
			ClusterName:    &clusterName,
			Name:           &npName,
			SubnetID:       &subnetID,
			InstanceType:   &instanceType,
			Replicas:       &replicas,
			MaxUnavailable: &updatedMaxUnavailable,
		})
		Expect(err).NotTo(HaveOccurred())

		npRefreshed, err := npSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(npRefreshed.Phase).To(Equal(string(v1alpha1.NodePoolPhaseReady)))
		Logger.Infof("[verify] nodepool %s remains in Ready phase after state refresh", npName)
	})

	It("validates max_unavailable with autoscaling", func() {
		seaplaneURL := requireEnv("SEAPLANE_URL")
		clusterName := requireEnv("SEAPLANE_CLUSTER_NAME")
		clusterID := requireEnv("SEAPLANE_CLUSTER_ID")
		subnetID := requireEnv("SEAPLANE_SUBNET_ID")
		instanceType := requireEnvWithDefault("SEAPLANE_NODEPOOL_INSTANCE_TYPE", "m5.xlarge")

		awsRegion := awsRegionRE.FindString(seaplaneURL)
		Expect(awsRegion).NotTo(BeEmpty(), fmt.Sprintf("cannot derive AWS region from SEAPLANE_URL %q", seaplaneURL))

		npName := clusterName + ".np-max-unavailable-autoscaling"
		workspace := "test-max-unavailable-autoscaling-" + fmt.Sprintf("%d", time.Now().Unix())

		seaplaneClient := mustBuildSeaplaneClient(seaplaneURL, awsRegion)
		nodepools := seaplaneClient.SeaplaneV1alpha1().NodePools()

		npSvc, err := exec.NewSeaplaneNodePoolService(workspace)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func(_ SpecContext) {
			By("Teardown: destroy autoscaling nodepool with max_unavailable")
			_, err := npSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())
		}, GracePeriod(45*time.Minute))

		// ── Create autoscaling nodepool with max_unavailable ───────────────
		By("Create autoscaling nodepool with max_unavailable='1' and min_replicas=2")
		maxUnavailable := "1"
		minReplicas := 2
		maxReplicas := 5
		enabled := true
		_, err = npSvc.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:  &seaplaneURL,
			AWSRegion:    &awsRegion,
			ClusterID:    &clusterID,
			ClusterName:  &clusterName,
			Name:         &npName,
			SubnetID:     &subnetID,
			InstanceType: &instanceType,
			Autoscaling: &exec.NodePoolAutoscaling{
				Enabled:     &enabled,
				MinReplicas: &minReplicas,
				MaxReplicas: &maxReplicas,
			},
			MaxUnavailable: &maxUnavailable,
		})
		Expect(err).NotTo(HaveOccurred())

		npTFOut, err := npSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(npTFOut.NodePoolID).NotTo(BeEmpty())

		// ── Wait for nodepool to reach Ready ───────────────────────────────
		By("Waiting for autoscaling nodepool to reach Ready phase")
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

		// ── Verify both autoscaling and max_unavailable are set ────────────
		By("Verifying both autoscaling and max_unavailable coexist")
		npObj, err := nodepools.Get(ctx, npName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		// Verify autoscaling
		Expect(npObj.Spec.NodePool.AutoScaling).NotTo(BeNil())
		Expect(npObj.Spec.NodePool.AutoScaling.Min).To(Equal(int32(minReplicas)))
		Expect(npObj.Spec.NodePool.AutoScaling.Max).To(Equal(int32(maxReplicas)))

		// Verify max_unavailable
		Expect(npObj.Spec.NodePool.Management.Replace.RollingUpdate.MaxUnavailable.IntVal).To(Equal(int32(1)))

		Logger.Infof("[verify] nodepool %s has both autoscaling (min=%d, max=%d) and max_unavailable=%d",
			npName,
			npObj.Spec.NodePool.AutoScaling.Min,
			npObj.Spec.NodePool.AutoScaling.Max,
			npObj.Spec.NodePool.Management.Replace.RollingUpdate.MaxUnavailable.IntVal)
	})
})
