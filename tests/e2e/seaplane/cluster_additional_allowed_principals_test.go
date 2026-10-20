// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	spplatform "github.com/openshift-online/rosa-seaplane-api/clientset/platform"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Seaplane cluster additional allowed principals", ci.SeaplaneValidated, func() {
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
