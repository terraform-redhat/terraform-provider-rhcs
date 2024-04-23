package e2e

import (

	// nolint

	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("Cluster Autoscaler", ci.FeatureClusterAutoscaler, func() {
	defer GinkgoRecover()

	var caService *exec.ClusterAutoscalerService
	var clusterAutoScalerBodyForRecreate *cmv1.ClusterAutoscaler
	var clusterAutoscalerStatusBefore int

	BeforeEach(func() {
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
	It("can be added/destroyed to Classic cluster - [id:69137]", ci.Day2, ci.High, ci.NonHCPCluster, ci.FeatureClusterAutoscaler, func() {
		caService = exec.NewClusterAutoscalerService(constants.ClassicClusterAutoscalerDir)

		By("Delete clusterautoscaler when it exists in cluster")
		if clusterAutoscalerStatusBefore == http.StatusOK {
			caDeleteBody, err := cms.DeleteClusterAutoscaler(ci.RHCSConnection, clusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(caDeleteBody.Status()).To(Equal(http.StatusNoContent))
		}

		By("Create clusterautoscaler")
		max := 1
		min := 0
		resourceRange := &exec.ResourceRange{
			Max: max,
			Min: min,
		}
		maxNodesTotal := 10
		resourceLimits := &exec.ResourceLimits{
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
		scaleDown := &exec.ScaleDown{
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
		ClusterAutoscalerArgs := &exec.ClusterAutoscalerArgs{
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
		_, err = caService.Apply(ClusterAutoscalerArgs, false)
		Expect(err).ToNot(HaveOccurred())
		_, err = caService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the createdautoscaler")
		caOut, err := caService.Output()
		Expect(err).ToNot(HaveOccurred())
		caResponse, err := cms.RetrieveClusterAutoscaler(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(caResponse.Body().BalanceSimilarNodeGroups()).To(Equal(caOut.BalanceSimilarNodeGroups))
		Expect(caResponse.Body().SkipNodesWithLocalStorage()).To(Equal(caOut.SkipNodesWithLocalStorage))
		Expect(caResponse.Body().LogVerbosity()).To(Equal(caOut.LogVerbosity))
		Expect(caResponse.Body().MaxPodGracePeriod()).To(Equal(caOut.MaxPodGracePeriod))
		Expect(caResponse.Body().PodPriorityThreshold()).To(Equal(caOut.PodPriorityThreshold))
		Expect(caResponse.Body().IgnoreDaemonsetsUtilization()).To(Equal(caOut.IgnoreDaemonsetsUtilization))
		Expect(caResponse.Body().MaxNodeProvisionTime()).To(Equal(caOut.MaxNodeProvisionTime))
		Expect(caResponse.Body().BalancingIgnoredLabels()).To(Equal(caOut.BalancingIgnoredLabels))
		Expect(caResponse.Body().ResourceLimits().MaxNodesTotal()).To(Equal(caOut.MaxNodesTotal))
		Expect(caResponse.Body().ScaleDown().DelayAfterAdd()).To(Equal(caOut.DelayAfterAdd))
		Expect(caResponse.Body().ScaleDown().DelayAfterDelete()).To(Equal(caOut.DelayAfterDelete))
		Expect(caResponse.Body().ScaleDown().DelayAfterFailure()).To(Equal(caOut.DelayAfterFailure))
		Expect(caResponse.Body().ScaleDown().UnneededTime()).To(Equal(caOut.UnneededTime))
		Expect(caResponse.Body().ScaleDown().UtilizationThreshold()).To(Equal(caOut.UtilizationThreshold))
		Expect(caResponse.Body().ScaleDown().Enabled()).To(Equal(caOut.Enabled))
	})

	It("can be created/edited/deleted to HCP cluster - [id:72524][id:72525]",
		ci.Day2,
		ci.High,
		ci.NonClassicCluster,
		func() {
			caService = exec.NewClusterAutoscalerService(constants.HCPClusterAutoscalerDir)

			if clusterAutoscalerStatusBefore == http.StatusOK {
				By("Delete current cluster autoscaler")
				caDeleteBody, err := cms.DeleteClusterAutoscaler(ci.RHCSConnection, clusterID)
				Expect(err).NotTo(HaveOccurred())
				Expect(caDeleteBody.Status()).To(Equal(http.StatusNoContent))
			}

			By("Add autoscaler to cluster")
			maxNodesTotal := 10
			resourceLimits := &exec.ResourceLimits{
				MaxNodesTotal: maxNodesTotal,
			}
			maxPodGracePeriod := 10
			podPriorityThreshold := -10
			maxNodeProvisionTime := "1h"
			clusterAutoscalerArgs := &exec.ClusterAutoscalerArgs{
				Cluster:              clusterID,
				MaxPodGracePeriod:    maxPodGracePeriod,
				PodPriorityThreshold: podPriorityThreshold,
				MaxNodeProvisionTime: maxNodeProvisionTime,
				ResourceLimits:       resourceLimits,
			}
			_, err = caService.Apply(clusterAutoscalerArgs, false)
			Expect(err).ToNot(HaveOccurred())
			_, err = caService.Output()
			Expect(err).ToNot(HaveOccurred())

			By("Verify autoscaler attributes")
			caOut, err := caService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(caOut.MaxPodGracePeriod).To(Equal(maxPodGracePeriod))
			Expect(caOut.PodPriorityThreshold).To(Equal(podPriorityThreshold))
			Expect(caOut.MaxNodeProvisionTime).To(Equal(maxNodeProvisionTime))
			Expect(caOut.MaxNodesTotal).To(Equal(maxNodesTotal))

			By("Edit cluster autoscaler")
			maxNodesTotal = 20
			resourceLimits = &exec.ResourceLimits{
				MaxNodesTotal: maxNodesTotal,
			}
			maxPodGracePeriod = 5
			podPriorityThreshold = 3
			maxNodeProvisionTime = "60m"
			clusterAutoscalerArgs = &exec.ClusterAutoscalerArgs{
				Cluster:              clusterID,
				MaxPodGracePeriod:    maxPodGracePeriod,
				PodPriorityThreshold: podPriorityThreshold,
				MaxNodeProvisionTime: maxNodeProvisionTime,
				ResourceLimits:       resourceLimits,
			}
			_, err = caService.Apply(clusterAutoscalerArgs, false)
			Expect(err).ToNot(HaveOccurred())
			_, err = caService.Output()
			Expect(err).ToNot(HaveOccurred())

			By("Verify autoscaler attributes")
			caOut, err = caService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(caOut.MaxPodGracePeriod).To(Equal(maxPodGracePeriod))
			Expect(caOut.PodPriorityThreshold).To(Equal(podPriorityThreshold))
			Expect(caOut.MaxNodeProvisionTime).To(Equal(maxNodeProvisionTime))
			Expect(caOut.MaxNodesTotal).To(Equal(maxNodesTotal))

			By("Destroy autoscaler")
			_, err = caService.Destroy()
			Expect(err).ToNot(HaveOccurred())

			By("Check autoscaler is gone")
			caResponse, err := cms.RetrieveClusterAutoscaler(ci.RHCSConnection, clusterID)
			Expect(err).To(HaveOccurred())
			Expect(caResponse.Status()).To(Equal(http.StatusNotFound))
		})
})
