// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Seaplane NodePool autoscaling", ci.SeaplaneValidated, func() {
	It("creates a nodepool with autoscaling and custom root volume size, then updates min/max replicas", func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(profileHandler.Profile().GetClusterType()).To(Equal(constants.SEAPLANE))
		if !profileHandler.Profile().IsAutoscaling() {
			Skip("Test requires a Seaplane autoscaling profile")
		}

		initialMin := profileHandler.Profile().GetMinReplicas()
		initialMax := profileHandler.Profile().GetMaxReplicas()
		Expect(initialMin).To(BeNumerically(">=", 0))
		Expect(initialMax).To(BeNumerically(">=", initialMin))

		workspace := profileHandler.Profile().GetName() + "-nodepool-autoscaling"
		backend, err := clusterworkflow.New(profileHandler, "", workspace)
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func(ctx SpecContext) {
			destroyer, destroyErr := clusterworkflow.NewDestroyer(profileHandler, "", workspace)
			Expect(destroyErr).NotTo(HaveOccurred())
			Expect(destroyer.Destroy(ctx)).To(Succeed())
		}, GracePeriod(teardownGracePeriod))

		clusterID, err := exec.CreateAndWaitReady(ctx, backend.Lifecycle)
		Expect(err).NotTo(HaveOccurred())

		seaplaneURL := os.Getenv("SEAPLANE_URL")
		region := config.GetRegion()
		if region == "" {
			region = awsRegionRE.FindString(seaplaneURL)
		}
		Expect(region).NotTo(BeEmpty())

		seaplaneClient := mustBuildSeaplaneClient(seaplaneURL, region)
		cluster, err := seaplaneClient.SeaplaneV1alpha1().Clusters().Get(ctx, clusterID, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(cluster.Spec.HostedCluster.Platform.AWS).NotTo(BeNil())
		Expect(cluster.Spec.HostedCluster.Platform.AWS.CloudProviderConfig).NotTo(BeNil())
		Expect(cluster.Spec.HostedCluster.Platform.AWS.CloudProviderConfig.Subnet).NotTo(BeNil())
		subnetID := cluster.Spec.HostedCluster.Platform.AWS.CloudProviderConfig.Subnet.ID
		Expect(subnetID).NotTo(BeNil())

		clusterName := cluster.Name
		nodePoolName := fmt.Sprintf("%s-autoscaling", clusterName)
		nodePoolService, err := exec.NewSeaplaneNodePoolService(workspace + "-nodepool")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func(_ SpecContext) {
			_, destroyErr := nodePoolService.Destroy()
			Expect(destroyErr).NotTo(HaveOccurred())
		}, GracePeriod(teardownGracePeriod))

		instanceType := os.Getenv("SEAPLANE_NODEPOOL_INSTANCE_TYPE")
		if instanceType == "" {
			instanceType = "m5.xlarge"
		}
		autoscalingEnabled := true
		diskSize := 500
		_, err = nodePoolService.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:  &seaplaneURL,
			AWSRegion:    &region,
			ClusterID:    &clusterID,
			ClusterName:  &clusterName,
			Name:         &nodePoolName,
			SubnetID:     subnetID,
			InstanceType: &instanceType,
			DiskSize:     &diskSize,
			Autoscaling: &exec.NodePoolAutoscaling{
				Enabled:     &autoscalingEnabled,
				MinReplicas: &initialMin,
				MaxReplicas: &initialMax,
			},
		})
		Expect(err).NotTo(HaveOccurred())
		nodePoolTFOut, err := nodePoolService.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(nodePoolTFOut.MinReplicas).NotTo(BeNil())
		Expect(*nodePoolTFOut.MinReplicas).To(Equal(initialMin))
		Expect(nodePoolTFOut.MaxReplicas).NotTo(BeNil())
		Expect(*nodePoolTFOut.MaxReplicas).To(Equal(initialMax))

		nodePools := seaplaneClient.SeaplaneV1alpha1().NodePools()
		assertAutoscaling := func(wantMin, wantMax int) func(*v1alpha1.NodePool) bool {
			return func(nodePool *v1alpha1.NodePool) bool {
				if nodePool == nil || nodePool.Spec.NodePool.Platform.AWS == nil ||
					nodePool.Spec.NodePool.Platform.AWS.RootVolume == nil ||
					nodePool.Spec.NodePool.Platform.AWS.RootVolume.Size != int64(diskSize) ||
					nodePool.Spec.NodePool.AutoScaling == nil ||
					nodePool.Spec.NodePool.AutoScaling.Min == nil {
					return false
				}
				return nodePool.Spec.NodePool.AutoScaling.Max == int32(wantMax) &&
					*nodePool.Spec.NodePool.AutoScaling.Min == int32(wantMin) &&
					nodePool.Status.Phase == v1alpha1.NodePoolPhaseReady
			}
		}

		Expect(
			nodePools.WaitUntil(
				ctx,
				nodePoolName,
				assertAutoscaling(initialMin, initialMax),
				pollInterval,
				nodepoolReadyTimeout,
			),
		).To(Succeed())

		updatedMin := initialMin + 1
		updatedMax := initialMax + 2
		_, err = nodePoolService.Apply(&exec.SeaplaneNodePoolArgs{
			SeaplaneURL:  &seaplaneURL,
			AWSRegion:    &region,
			ClusterID:    &clusterID,
			ClusterName:  &clusterName,
			Name:         &nodePoolName,
			SubnetID:     subnetID,
			InstanceType: &instanceType,
			DiskSize:     &diskSize,
			Autoscaling: &exec.NodePoolAutoscaling{
				Enabled:     &autoscalingEnabled,
				MinReplicas: &updatedMin,
				MaxReplicas: &updatedMax,
			},
		})
		Expect(err).NotTo(HaveOccurred())
		nodePoolTFOut, err = nodePoolService.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(nodePoolTFOut.MinReplicas).NotTo(BeNil())
		Expect(*nodePoolTFOut.MinReplicas).To(Equal(updatedMin))
		Expect(nodePoolTFOut.MaxReplicas).NotTo(BeNil())
		Expect(*nodePoolTFOut.MaxReplicas).To(Equal(updatedMax))
		Expect(
			nodePools.WaitUntil(
				ctx,
				nodePoolName,
				assertAutoscaling(updatedMin, updatedMax),
				pollInterval,
				nodepoolReadyTimeout,
			),
		).To(Succeed())
	})
})
