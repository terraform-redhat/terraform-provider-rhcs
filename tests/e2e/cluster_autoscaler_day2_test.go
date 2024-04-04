package e2e

import (

	// nolint

	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("TF Test", func() {
	Describe("Create cluster autoscaler test cases", func() {
		var caService *exe.ClusterAutoscalerService
		var clusterAutoScalerBodyForRecreate *cmv1.ClusterAutoscaler
		var clusterAutoscalerStatusBefore int

		BeforeEach(func() {
			tfExecHelper, err := ci.GetTerraformExecHelperForProfile(profile)
			Expect(err).ToNot(HaveOccurred())
			caService, err = tfExecHelper.GetClusterAutoscalerService()
			Expect(err).ToNot(HaveOccurred())

			caRetrieveBody, _ := cms.RetrieveClusterAutoscaler(ci.RHCSConnection, clusterID)
			clusterAutoscalerStatusBefore = caRetrieveBody.Status()
			if clusterAutoscalerStatusBefore == http.StatusOK {
				clusterAutoScalerBodyForRecreate = caRetrieveBody.Body()
			}
		})
		AfterEach(func() {
			By("Recover clusterautoscaler")
			clusterAutoscalerAfter, _ := cms.RetrieveClusterAutoscaler(ci.RHCSConnection, clusterID)
			if (clusterAutoscalerAfter.Status() == clusterAutoscalerStatusBefore) && clusterAutoscalerStatusBefore != http.StatusNotFound {
				recreateAutoscaler, err := cms.PatchClusterAutoscaler(ci.RHCSConnection, clusterID, clusterAutoScalerBodyForRecreate)
				Expect(err).NotTo(HaveOccurred())
				Expect(recreateAutoscaler.Status()).To(Equal(http.StatusOK))
			} else if clusterAutoscalerAfter.Status() == http.StatusOK && clusterAutoscalerStatusBefore == http.StatusNotFound {
				deleteAutoscaler, err := cms.DeleteClusterAutoscaler(ci.RHCSConnection, clusterID)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleteAutoscaler.Status()).To(Equal(http.StatusNoContent))
			} else if clusterAutoscalerAfter.Status() == http.StatusNotFound && clusterAutoscalerStatusBefore == http.StatusOK {
				recreateAutoscaler, err := cms.CreateClusterAutoscaler(ci.RHCSConnection, clusterID, clusterAutoScalerBodyForRecreate)
				Expect(err).NotTo(HaveOccurred())
				Expect(recreateAutoscaler.Status()).To(Equal(http.StatusCreated))
			}
		})
		Context("Author:zhsun-High-OCP-69137 @zhsun", func() {
			It("Author:zhsun-High-OCP-69137 - Create cluster autoscaler with ocm terraform provider", ci.Day2, ci.High, ci.NonHCPCluster, ci.FeatureClusterautoscaler, func() {
				By("Delete clusterautoscaler when it exists in cluster")
				if clusterAutoscalerStatusBefore == http.StatusOK {
					caDeleteBody, err := cms.DeleteClusterAutoscaler(ci.RHCSConnection, clusterID)
					Expect(err).NotTo(HaveOccurred())
					Expect(caDeleteBody.Status()).To(Equal(http.StatusNoContent))
				}

				By("Create clusterautoscaler")
				max := 1
				min := 0
				resourceRange := &exe.ResourceRange{
					Max: max,
					Min: min,
				}
				maxNodesTotal := 10
				resourceLimits := &exe.ResourceLimits{
					Cores:         resourceRange,
					MaxNodesTotal: maxNodesTotal,
					Memory:        resourceRange,
				}
				delayAfterAdd := "3h"
				delayAfterDelete := "3h"
				delayAfterFailure := "3h"
				unneededTime := "1h"
				utilizationThreshold := "0.5"
				enabled := true
				scaleDown := &exe.ScaleDown{
					DelayAfterAdd:        delayAfterAdd,
					DelayAfterDelete:     delayAfterDelete,
					DelayAfterFailure:    delayAfterFailure,
					UnneededTime:         unneededTime,
					UtilizationThreshold: utilizationThreshold,
					Enabled:              enabled,
				}
				balanceSimilarNodeGroups := true
				skipNodesWithLocalStorage := true
				logVerbosity := 1
				maxPodGracePeriod := 10
				podPriorityThreshold := -10
				ignoreDaemonsetsUtilization := true
				maxNodeProvisionTime := "1h"
				balancingIgnoredLabels := []string{"l1", "l2"}
				clusterAutoscalerArgs := &exe.ClusterAutoscalerArgs{
					Cluster:                     clusterID,
					BalanceSimilarNodeGroups:    balanceSimilarNodeGroups,
					SkipNodesWithLocalStorage:   skipNodesWithLocalStorage,
					LogVerbosity:                logVerbosity,
					MaxPodGracePeriod:           maxPodGracePeriod,
					PodPriorityThreshold:        podPriorityThreshold,
					IgnoreDaemonsetsUtilization: ignoreDaemonsetsUtilization,
					MaxNodeProvisionTime:        maxNodeProvisionTime,
					BalancingIgnoredLabels:      balancingIgnoredLabels,
					ResourceLimits:              resourceLimits,
					ScaleDown:                   scaleDown,
				}
				_, err := caService.Apply(clusterAutoscalerArgs)
				Expect(err).ToNot(HaveOccurred())
				_, err = caService.Output()
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the createdautoscaler")
				caOut, err := caService.Output()
				Expect(err).ToNot(HaveOccurred())
				caResponseBody, err := cms.RetrieveClusterAutoscaler(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(caResponseBody.Body().BalanceSimilarNodeGroups()).To(Equal(caOut.BalanceSimilarNodeGroups))
				Expect(caResponseBody.Body().SkipNodesWithLocalStorage()).To(Equal(caOut.SkipNodesWithLocalStorage))
				Expect(caResponseBody.Body().LogVerbosity()).To(Equal(caOut.LogVerbosity))
				Expect(caResponseBody.Body().MaxPodGracePeriod()).To(Equal(caOut.MaxPodGracePeriod))
				Expect(caResponseBody.Body().PodPriorityThreshold()).To(Equal(caOut.PodPriorityThreshold))
				Expect(caResponseBody.Body().IgnoreDaemonsetsUtilization()).To(Equal(caOut.IgnoreDaemonsetsUtilization))
				Expect(caResponseBody.Body().MaxNodeProvisionTime()).To(Equal(caOut.MaxNodeProvisionTime))
				Expect(caResponseBody.Body().BalancingIgnoredLabels()).To(Equal(caOut.BalancingIgnoredLabels))
				Expect(caResponseBody.Body().ResourceLimits().MaxNodesTotal()).To(Equal(caOut.MaxNodesTotal))
				Expect(caResponseBody.Body().ScaleDown().DelayAfterAdd()).To(Equal(caOut.DelayAfterAdd))
				Expect(caResponseBody.Body().ScaleDown().DelayAfterDelete()).To(Equal(caOut.DelayAfterDelete))
				Expect(caResponseBody.Body().ScaleDown().DelayAfterFailure()).To(Equal(caOut.DelayAfterFailure))
				Expect(caResponseBody.Body().ScaleDown().UnneededTime()).To(Equal(caOut.UnneededTime))
				Expect(caResponseBody.Body().ScaleDown().UtilizationThreshold()).To(Equal(caOut.UtilizationThreshold))
				Expect(caResponseBody.Body().ScaleDown().Enabled()).To(Equal(caOut.Enabled))
			})
		})
	})
})
