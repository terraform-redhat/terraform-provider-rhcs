// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/clusterworkflow"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Seaplane cluster Day 2 operations", ci.SeaplaneValidated, func() {
	It("updates autoscaler and component routes as a day-2 operation", func(ctx SpecContext) {
		profile, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if profile.Profile().GetClusterType() != constants.SEAPLANE {
			Skip("Day-2 test requires a Seaplane profile")
		}
		workspace := seaplaneTestWorkspace(profile.Profile().GetName(), "autoscaler-ingress-day2")
		backend, err := clusterworkflow.New(profile, "", workspace)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			destroyer, cleanupErr := clusterworkflow.NewDestroyer(profile, "", workspace)
			Expect(cleanupErr).NotTo(HaveOccurred())
			Expect(destroyer.Destroy(context.Background())).To(Succeed())
		})
		_, err = exec.CreateAndWaitReady(ctx, backend.Lifecycle)
		Expect(err).NotTo(HaveOccurred())
		updater, ok := backend.Lifecycle.(interface {
			Update(*exec.SeaplaneClusterArgs) error
		})
		Expect(ok).To(BeTrue())
		routes := `[{"namespace":"openshift-console","name":"console","hostname":"console-day2.example.com","servingCertKeyPairSecret":{"name":"console-tls"}}]`
		provisionTime := "20m"
		maxNodes, maxGrace, priority := int64(12), int64(700), int64(-5)
		Expect(updater.Update(&exec.SeaplaneClusterArgs{
			AutoscalerMaxNodeProvisionTime: &provisionTime,
			AutoscalerMaxNodesTotal:        &maxNodes,
			AutoscalerMaxPodGracePeriod:    &maxGrace,
			AutoscalerPodPriorityThreshold: &priority,
			ComponentRoutes:                &routes,
		})).To(Succeed())
		service, err := exec.NewSeaplaneClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())
		output, err := service.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(output.AutoscalerMaxNodesTotal).To(Equal(int64(12)))
		Expect(output.AutoscalerMaxPodGracePeriod).To(Equal(int64(700)))
		Expect(output.AutoscalerPodPriorityThreshold).To(Equal(int64(-5)))
		Expect(output.AutoscalerMaxNodeProvisionTime).To(Equal("20m"))
		Expect(output.ComponentRoutes).To(ContainSubstring("console-day2.example.com"))
	})

	It("updates additional allowed principals", ci.Day2, ci.High, func(ctx SpecContext) {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		Expect(profileHandler.Profile().GetClusterType()).To(Equal(constants.SEAPLANE))

		seaplaneURL := requireEnv("SEAPLANE_URL")
		awsRegion := config.GetRegion()
		if awsRegion == "" {
			awsRegion = awsRegionRE.FindString(seaplaneURL)
		}
		Expect(awsRegion).NotTo(BeEmpty(), "cannot derive AWS region from SEAPLANE_URL")

		workspace := profileHandler.Profile().GetName()
		clusterService, err := exec.NewSeaplaneClusterService(workspace)
		Expect(err).NotTo(HaveOccurred())
		clusterOutput, err := clusterService.Output()
		if err != nil || clusterOutput == nil || clusterOutput.ClusterID == "" {
			Skip("Test requires an existing Seaplane cluster in Terraform state; run CreateClusterByProfile first")
		}

		clusterArgs, err := clusterService.ReadTFVars()
		Expect(err).NotTo(HaveOccurred())
		clusterArgs.SeaplaneURL = &seaplaneURL
		clusterName := clusterOutput.ClusterName
		if clusterName == "" {
			clusterName = config.GetRHCSClusterName()
		}
		if clusterName == "" {
			clusterName = profileHandler.Profile().GetName()
		}

		originalPrincipals := clusterArgs.AdditionalAllowedPrincipals
		expectedRestoredPrincipals := []string(nil)
		if originalPrincipals != nil {
			expectedRestoredPrincipals = *originalPrincipals
		}

		iamService, err := exec.NewSeaplaneIAMService(workspace + "-iam")
		Expect(err).NotTo(HaveOccurred())
		iamOutput, err := iamService.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(iamOutput.WorkerRoleARN).NotTo(BeEmpty())

		clusters := mustBuildSeaplaneClient(seaplaneURL, awsRegion).SeaplaneV1alpha1().Clusters()
		DeferCleanup(func(cleanupCtx SpecContext) {
			clusterArgs.AdditionalAllowedPrincipals = originalPrincipals
			_, restoreErr := clusterService.Apply(clusterArgs)
			Expect(restoreErr).NotTo(HaveOccurred())

			restoredPrincipalsMatcher := Equal(expectedRestoredPrincipals)
			if expectedRestoredPrincipals == nil {
				restoredPrincipalsMatcher = BeEmpty()
			}
			Eventually(func() []string {
				cluster, getErr := clusters.Get(cleanupCtx, clusterName, spplatform.GetOptions{})
				if getErr != nil {
					return nil
				}
				return cluster.Spec.HostedCluster.Platform.AWS.AdditionalAllowedPrincipals
			}).WithContext(cleanupCtx).WithTimeout(5 * time.Minute).WithPolling(30 * time.Second).Should(
				restoredPrincipalsMatcher,
			)
		})

		updatedPrincipals := []string{iamOutput.WorkerRoleARN}
		clusterArgs.AdditionalAllowedPrincipals = &updatedPrincipals
		_, err = clusterService.Apply(clusterArgs)
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() []string {
			cluster, getErr := clusters.Get(ctx, clusterName, spplatform.GetOptions{})
			if getErr != nil {
				return nil
			}
			return cluster.Spec.HostedCluster.Platform.AWS.AdditionalAllowedPrincipals
		}).WithContext(ctx).WithTimeout(5 * time.Minute).WithPolling(30 * time.Second).Should(Equal(updatedPrincipals))
	})
})
