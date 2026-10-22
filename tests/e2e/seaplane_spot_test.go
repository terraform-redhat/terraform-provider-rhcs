// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

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
	v1alpha1 "github.com/openshift-online/rosa-seaplane-api/api/v1alpha1/public"
	seaplane "github.com/openshift-online/rosa-seaplane-api/clientset"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"
	sprest "github.com/openshift-online/rosa-seaplane-api/clientset/rest"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

const (
	seaplaneSpotNodePoolReadyTimeout        = 45 * time.Minute
	seaplaneSpotNodePoolPollInterval        = 30 * time.Second
	seaplaneSpotNodePoolTeardownGracePeriod = seaplaneSpotNodePoolReadyTimeout + 30*time.Minute
)

var seaplaneSpotAWSRegionPattern = regexp.MustCompile(`[a-z]+-(?:[a-z]+-)+\d+`)

var _ = Describe("Seaplane Spot node pool lifecycle", ci.SeaplaneValidated, func() {
	It("creates, verifies, refreshes, and deletes a Spot node pool on an existing Seaplane cluster", func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if profileHandler.Profile().GetClusterType() != constants.SEAPLANE {
			Skip("Test requires a Seaplane profile")
		}
		Expect(clusterID).NotTo(BeEmpty(),
			"the shared E2E fixture must resolve an existing Seaplane cluster ID; set CLUSTER_ID or retain the selected profile's Terraform state")

		seaplaneURL := requireSeaplaneSpotEnv("SEAPLANE_URL")
		awsRegion := config.GetRegion()
		if awsRegion == "" {
			awsRegion = seaplaneSpotAWSRegionPattern.FindString(seaplaneURL)
		}
		Expect(awsRegion).NotTo(BeEmpty(), "REGION is required when it cannot be derived from SEAPLANE_URL")

		clusterName := config.GetRHCSClusterName()
		if clusterName == "" {
			clusterName = profileHandler.Profile().GetName()
		}
		seaplaneClient := mustBuildSeaplaneSpotClient(seaplaneURL, awsRegion)
		clusters := seaplaneClient.SeaplaneV1alpha1().Clusters()
		existingCluster, err := clusters.Get(ctx, clusterName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(existingCluster.UID)).To(Equal(clusterID),
			"the cluster returned by the Seaplane API must match the shared fixture cluster ID")
		Expect(existingCluster.Name).To(Equal(clusterName))
		Expect(existingCluster.Status.Phase).To(Equal(v1alpha1.ClusterPhaseReady),
			"the existing cluster must be Ready before creating a node pool")
		Expect(existingCluster.Spec.HostedCluster.Platform.AWS).NotTo(BeNil(),
			"the existing cluster must have an AWS platform configuration")
		awsPlatform := existingCluster.Spec.HostedCluster.Platform.AWS
		Expect(awsPlatform.CloudProviderConfig).NotTo(BeNil(),
			"the existing cluster must expose its AWS subnet configuration")
		Expect(awsPlatform.CloudProviderConfig.Subnet).NotTo(BeNil(),
			"the existing cluster must expose a subnet for the node pool")
		Expect(awsPlatform.CloudProviderConfig.Subnet.ID).NotTo(BeNil(),
			"the existing cluster must expose a subnet ID for the node pool")
		subnetID := *awsPlatform.CloudProviderConfig.Subnet.ID
		Expect(subnetID).NotTo(BeEmpty())
		Expect(awsPlatform.Region).To(Equal(awsRegion),
			"SEAPLANE_URL and the existing cluster must use the same AWS region")

		Expect(clusterName).NotTo(BeEmpty())
		spotNodePoolSuffix := fmt.Sprintf("spot-np-%d", time.Now().UnixNano())
		spotNodePoolName := clusterName + "." + spotNodePoolSuffix
		spotNodePoolWorkspace := clusterName + "-" + spotNodePoolSuffix
		spotNodePoolSvc, err := exec.NewSeaplaneNodePoolService(spotNodePoolWorkspace)
		Expect(err).NotTo(HaveOccurred())
		nodepools := seaplaneClient.SeaplaneV1alpha1().NodePools()

		var spotNodePoolApplyAttempted bool
		DeferCleanup(func(cleanupCtx SpecContext) {
			if !spotNodePoolApplyAttempted {
				return
			}
			By("Deleting the Spot node pool from the retained cluster")
			_, destroyErr := spotNodePoolSvc.Destroy()
			Expect(destroyErr).NotTo(HaveOccurred())

			waitErr := nodepools.WaitUntil(
				cleanupCtx,
				spotNodePoolName,
				func(nodePool *v1alpha1.NodePool) bool {
					if nodePool == nil {
						Logger.Infof("[wait] nodepool %s: deleted", spotNodePoolName)
						return true
					}
					Logger.Infof("[wait] nodepool %s phase: %s (waiting for deletion)", spotNodePoolName, nodePool.Status.Phase)
					return false
				},
				seaplaneSpotNodePoolPollInterval,
				seaplaneSpotNodePoolReadyTimeout,
			)
			Expect(waitErr).NotTo(HaveOccurred())
		}, GracePeriod(seaplaneSpotNodePoolTeardownGracePeriod))

		instanceType := requireSeaplaneSpotEnvWithDefault("SEAPLANE_NODEPOOL_INSTANCE_TYPE", "m5.xlarge")
		spotMarketType := "Spot"
		spotMaxPrice := "0.20"
		spotReplicas := 1
		spotNodePoolArgs := &exec.SeaplaneNodePoolArgs{
			SeaplaneURL:         &seaplaneURL,
			AWSRegion:           &awsRegion,
			ClusterID:           &clusterID,
			ClusterName:         &clusterName,
			Name:                &spotNodePoolName,
			SubnetID:            &subnetID,
			InstanceType:        &instanceType,
			Replicas:            &spotReplicas,
			PlacementMarketType: &spotMarketType,
			MaxPrice:            &spotMaxPrice,
		}

		By("Creating a Spot node pool on the existing cluster")
		spotNodePoolApplyAttempted = true
		_, err = spotNodePoolSvc.Apply(spotNodePoolArgs)
		Expect(err).NotTo(HaveOccurred())

		spotNodePoolOutput, err := spotNodePoolSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(spotNodePoolOutput.NodePoolID).NotTo(BeEmpty())
		Expect(spotNodePoolOutput.PlacementMarketType).To(Equal(spotMarketType))
		Expect(spotNodePoolOutput.MaxPrice).To(Equal(spotMaxPrice))

		By("Waiting for the Spot node pool to reach Ready with one replica")
		waitErr := nodepools.WaitUntil(
			ctx,
			spotNodePoolName,
			func(nodePool *v1alpha1.NodePool) bool {
				if nodePool == nil {
					Logger.Infof("[wait] nodepool %s: not found", spotNodePoolName)
					return false
				}
				observedReplicas := "<unset>"
				if nodePool.Status.Replicas != nil {
					observedReplicas = fmt.Sprint(*nodePool.Status.Replicas)
				}
				Logger.Infof("[wait] nodepool %s phase: %s replicas=%s", spotNodePoolName, nodePool.Status.Phase, observedReplicas)
				return nodePool.Status.Phase == v1alpha1.NodePoolPhaseReady &&
					nodePool.Status.Replicas != nil && int64(*nodePool.Status.Replicas) == int64(spotReplicas)
			},
			seaplaneSpotNodePoolPollInterval,
			seaplaneSpotNodePoolReadyTimeout,
		)
		Expect(waitErr).NotTo(HaveOccurred(),
			"Spot node pool did not become Ready with %d current replicas on existing cluster %q", spotReplicas, clusterID)

		By("Verifying the Spot node pool through the Platform API")
		spotNodePool, err := nodepools.Get(ctx, spotNodePoolName, spplatform.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(spotNodePool).NotTo(BeNil())
		Expect(spotNodePool.Spec.NodePool.Platform.AWS).NotTo(BeNil())
		spotAWSPlatform := spotNodePool.Spec.NodePool.Platform.AWS
		Expect(spotAWSPlatform.Placement).NotTo(BeNil())
		Expect(string(spotAWSPlatform.Placement.MarketType)).To(Equal(spotMarketType))
		Expect(spotAWSPlatform.Placement.Spot).NotTo(BeNil())
		Expect(spotAWSPlatform.Placement.Spot.MaxPrice).To(Equal(spotMaxPrice))
		Expect(spotNodePool.Status.Replicas).NotTo(BeNil())
		Expect(int64(*spotNodePool.Status.Replicas)).To(Equal(int64(spotReplicas)))

		By("Verifying the Spot node pool through Terraform state refresh")
		_, err = spotNodePoolSvc.Apply(spotNodePoolArgs)
		Expect(err).NotTo(HaveOccurred())
		refreshedNodePool, err := spotNodePoolSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(refreshedNodePool.NodePoolID).To(Equal(spotNodePoolOutput.NodePoolID))
		Expect(refreshedNodePool.Phase).To(Equal(string(v1alpha1.NodePoolPhaseReady)))
		Expect(refreshedNodePool.PlacementMarketType).To(Equal(spotMarketType))
		Expect(refreshedNodePool.MaxPrice).To(Equal(spotMaxPrice))
		Expect(refreshedNodePool.Replicas).NotTo(BeNil())
		Expect(*refreshedNodePool.Replicas).To(Equal(spotReplicas))
	})
})

func mustBuildSeaplaneSpotClient(seaplaneURL, region string) *seaplane.Clientset {
	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	Expect(err).NotTo(HaveOccurred())
	identity, err := sts.NewFromConfig(awsCfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	Expect(err).NotTo(HaveOccurred())
	client, err := seaplane.NewForConfig(&sprest.Config{
		Host:      seaplaneURL,
		Region:    region,
		AccountID: *identity.Account,
		CallerARN: *identity.Arn,
		AWSConfig: awsCfg,
	})
	Expect(err).NotTo(HaveOccurred())
	return client
}

func requireSeaplaneSpotEnv(key string) string {
	value := os.Getenv(key)
	Expect(value).NotTo(BeEmpty(), fmt.Sprintf("required env var %s is not set", key))
	return value
}

func requireSeaplaneSpotEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
