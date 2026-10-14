// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package hyperfleet

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	. "github.com/onsi/gomega"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	"github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("Hyperfleet existing cluster node pool", func() {
	It("creates, updates, and clears labels, taints, and auto-repair", Label("HyperfleetValidated"), func() {
		hyperfleetURL := requireEnv("HYPERFLEET_URL")
		clusterID := requireEnv("HYPERFLEET_EXISTING_CLUSTER_ID")
		subnetID := requireEnv("HYPERFLEET_EXISTING_SUBNET_ID")
		awsRegion := awsRegionRE.FindString(hyperfleetURL)
		Expect(awsRegion).NotTo(BeEmpty(), fmt.Sprintf("cannot derive AWS region from HYPERFLEET_URL %q", hyperfleetURL))

		client := mustBuildHyperfleetClient(hyperfleetURL, awsRegion)
		ctx := context.Background()
		cluster, err := client.HyperfleetV1alpha1().Clusters().Get(ctx, clusterID, platform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(cluster.Status.Phase).To(Equal(v1alpha1.ClusterPhaseReady))
		clusterName := cluster.Name

		nodePoolName := fmt.Sprintf("labels-taints-%d", time.Now().UnixNano())
		workspace := fmt.Sprintf("hf-existing-labels-taints-%d", time.Now().UnixNano())
		nodePoolService, err := exec.NewHyperfleetNodePoolService(workspace)
		Expect(err).NotTo(HaveOccurred())
		nodePools := client.HyperfleetV1alpha1().NodePools(clusterID)

		DeferCleanup(func(cleanupCtx SpecContext) {
			By("Removing only the test node pool; keeping the existing cluster")
			_, destroyErr := nodePoolService.Destroy()
			Expect(destroyErr).NotTo(HaveOccurred())
			deleteErr := nodePools.WaitUntil(
				cleanupCtx, nodePoolName,
				func(nodePool *v1alpha1.NodePool) bool { return nodePool == nil },
				pollInterval, nodepoolReadyTimeout,
			)
			Expect(deleteErr).NotTo(HaveOccurred())
		}, GracePeriod(nodepoolReadyTimeout+15*time.Minute))

		replicas := 1
		autoRepair := false
		instanceType := requireEnvWithDefault("HYPERFLEET_NODEPOOL_INSTANCE_TYPE", "m5.xlarge")
		labels := map[string]string{
			"e2e-label-one": "value-one",
			"e2e-label-two": "value-two",
		}
		taints := []map[string]string{{
			"key":    "e2e-taint-one",
			"value":  "value-one",
			"effect": "NoSchedule",
		}}
		args := &exec.HyperfleetNodePoolArgs{
			HyperfleetURL: &hyperfleetURL,
			AWSRegion:     &awsRegion,
			ClusterID:     &clusterID,
			ClusterName:   &clusterName,
			Name:          &nodePoolName,
			SubnetID:      &subnetID,
			AutoRepair:    &autoRepair,
			Replicas:      &replicas,
			InstanceType:  &instanceType,
			Labels:        labels,
			Taints:        &taints,
		}

		By("Creating a test node pool on the existing Hyperfleet cluster")
		_, err = nodePoolService.Apply(args)
		Expect(err).NotTo(HaveOccurred())
		Expect(waitNodePoolReadyForUpdate(ctx, nodePools, nodePoolName)).To(Succeed())

		By("Verifying [72507] node pool labels, taints, and auto-repair on create")
		nodePool, err := nodePools.Get(ctx, nodePoolName, platform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(nodePool.Spec.Labels).To(Equal(labels))
		Expect(nodePool.Spec.NodePool.Taints).To(HaveLen(1))
		Expect(nodePool.Spec.NodePool.Taints[0].Key).To(Equal("e2e-taint-one"))
		Expect(nodePool.Spec.NodePool.Taints[0].Value).To(Equal("value-one"))
		Expect(string(nodePool.Spec.NodePool.Taints[0].Effect)).To(Equal("NoSchedule"))
		Expect(nodePool.Spec.AutoRepair).NotTo(BeNil())
		Expect(*nodePool.Spec.AutoRepair).To(BeFalse())

		By("Updating [72507] node pool auto-repair")
		autoRepair = true
		args.AutoRepair = &autoRepair
		_, err = nodePoolService.Apply(args)
		Expect(err).NotTo(HaveOccurred())
		nodePool, err = nodePools.Get(ctx, nodePoolName, platform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(nodePool.Spec.Labels).To(Equal(labels))
		Expect(nodePool.Spec.NodePool.Taints).To(HaveLen(1))
		Expect(nodePool.Spec.NodePool.Taints[0].Key).To(Equal("e2e-taint-one"))
		Expect(nodePool.Spec.AutoRepair).NotTo(BeNil())
		Expect(*nodePool.Spec.AutoRepair).To(BeTrue())

		By("Adding a taint to [72507] node pool")
		taints = append(taints, map[string]string{
			"key":    "e2e-taint-two",
			"value":  "",
			"effect": "NoExecute",
		})
		args.Taints = &taints
		_, err = nodePoolService.Apply(args)
		Expect(err).NotTo(HaveOccurred())
		nodePool, err = nodePools.Get(ctx, nodePoolName, platform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(nodePool.Spec.Labels).To(Equal(labels))
		Expect(nodePool.Spec.NodePool.Taints).To(HaveLen(2))
		Expect(nodePool.Spec.NodePool.Taints[0].Key).To(Equal("e2e-taint-one"))
		Expect(nodePool.Spec.NodePool.Taints[1].Key).To(Equal("e2e-taint-two"))
		Expect(string(nodePool.Spec.NodePool.Taints[1].Effect)).To(Equal("NoExecute"))
		Expect(nodePool.Spec.AutoRepair).NotTo(BeNil())
		Expect(*nodePool.Spec.AutoRepair).To(BeTrue())

		By("Updating [72507] node pool labels")
		labels = map[string]string{"e2e-label-three": "value-three"}
		args.Labels = labels
		_, err = nodePoolService.Apply(args)
		Expect(err).NotTo(HaveOccurred())
		nodePool, err = nodePools.Get(ctx, nodePoolName, platform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(nodePool.Spec.Labels).To(Equal(labels))
		Expect(nodePool.Spec.NodePool.Taints).To(HaveLen(2))
		Expect(nodePool.Spec.AutoRepair).NotTo(BeNil())
		Expect(*nodePool.Spec.AutoRepair).To(BeTrue())

		By("Clearing [72507] node pool labels and taints")
		args.Labels = nil
		args.Taints = nil
		_, err = nodePoolService.Apply(args)
		Expect(err).NotTo(HaveOccurred())
		nodePool, err = nodePools.Get(ctx, nodePoolName, platform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(nodePool.Spec.Labels).To(BeEmpty())
		Expect(nodePool.Spec.NodePool.Taints).To(BeEmpty())
		Expect(nodePool.Spec.AutoRepair).NotTo(BeNil())
		Expect(*nodePool.Spec.AutoRepair).To(BeTrue())
	})
})

func nodePoolReadyForUpdate(nodePool *v1alpha1.NodePool) bool {
	if nodePool == nil || nodePool.Status.Phase != v1alpha1.NodePoolPhaseReady {
		return false
	}
	for _, condition := range nodePool.Status.Conditions {
		if condition.Type == "Synced" && condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

func waitNodePoolReadyForUpdate(
	ctx context.Context,
	nodePools platform.NodePoolInterface,
	name string,
) error {
	deadline := time.Now().Add(nodepoolReadyTimeout)
	notFoundPolls := 0

	for {
		nodePool, err := nodePools.Get(ctx, name, platform.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				notFoundPolls++
				if notFoundPolls >= 2 {
					return fmt.Errorf(
						"node pool %q disappeared while waiting to become ready; verify the parent cluster still exists: %w",
						name,
						err,
					)
				}
			} else {
				return err
			}
		} else {
			notFoundPolls = 0
			if nodePoolReadyForUpdate(nodePool) {
				return nil
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for node pool %q to become ready", nodepoolReadyTimeout, name)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}
