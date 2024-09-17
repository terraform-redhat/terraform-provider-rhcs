package e2e

import (

	// nolint

	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cmsv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

var _ = Describe("Cluster Autoscaler", ci.Day2, ci.FeatureClusterAutoscaler, func() {
	defer GinkgoRecover()

	var caService exec.ClusterAutoscalerService
	var clusterAutoScalerBodyForRecreate *cmsv1.ClusterAutoscaler
	var clusterAutoscalerStatusBefore int
	var profile *ci.Profile

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()
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
	It("can be added/destroyed to Classic cluster - [id:69137]", ci.High, func() {
		if profile.GetClusterType().HCP {
			Skip("Test can run only on Classic cluster")
		}

		var err error
		caService, err = exec.NewClusterAutoscalerService(constants.ClassicClusterAutoscalerDir)
		Expect(err).NotTo(HaveOccurred())

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
			Max: helper.IntPointer(max),
			Min: helper.IntPointer(min),
		}
		maxNodesTotal := 10
		resourceLimits := &exec.ResourceLimits{
			Cores:         resourceRange,
			MaxNodesTotal: helper.IntPointer(maxNodesTotal),
			Memory:        resourceRange,
		}
		delayAfterAdd := "3h"
		delayAfterDelete := "3h"
		delayAfterFailure := "3h"
		unneededTime := "1h"
		utilizationThreshold := "0.5"
		enabled := true
		scaleDown := &exec.ScaleDown{
			DelayAfterAdd:        helper.StringPointer(delayAfterAdd),
			DelayAfterDelete:     helper.StringPointer(delayAfterDelete),
			DelayAfterFailure:    helper.StringPointer(delayAfterFailure),
			UnneededTime:         helper.StringPointer(unneededTime),
			UtilizationThreshold: helper.StringPointer(utilizationThreshold),
			Enabled:              helper.BoolPointer(enabled),
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
			Cluster:                     &clusterID,
			BalanceSimilarNodeGroups:    helper.BoolPointer(balanceSimilarNodeGroups),
			SkipNodesWithLocalStorage:   helper.BoolPointer(skipNodesWithLocalStorage),
			LogVerbosity:                helper.IntPointer(logVerbosity),
			MaxPodGracePeriod:           helper.IntPointer(maxPodGracePeriod),
			PodPriorityThreshold:        helper.IntPointer(podPriorityThreshold),
			IgnoreDaemonsetsUtilization: helper.BoolPointer(ignoreDaemonsetsUtilization),
			MaxNodeProvisionTime:        helper.StringPointer(maxNodeProvisionTime),
			BalancingIgnoredLabels:      helper.StringSlicePointer(balancingIgnoredLabels),
			ResourceLimits:              resourceLimits,
			ScaleDown:                   scaleDown,
		}
		_, err = caService.Apply(ClusterAutoscalerArgs)
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
		ci.High,
		func() {
			if !profile.GetClusterType().HCP {
				Skip("Test can run only on Hosted cluster")
			}

			var err error
			caService, err = exec.NewClusterAutoscalerService(constants.HCPClusterAutoscalerDir)
			Expect(err).NotTo(HaveOccurred())

			if clusterAutoscalerStatusBefore == http.StatusOK {
				By("Delete current cluster autoscaler")
				caDeleteBody, err := cms.DeleteClusterAutoscaler(ci.RHCSConnection, clusterID)
				Expect(err).NotTo(HaveOccurred())
				Expect(caDeleteBody.Status()).To(Equal(http.StatusNoContent))
			}

			By("Add autoscaler to cluster")
			maxNodesTotal := 10
			resourceLimits := &exec.ResourceLimits{
				MaxNodesTotal: helper.IntPointer(maxNodesTotal),
			}
			maxPodGracePeriod := 10
			podPriorityThreshold := -10
			maxNodeProvisionTime := "1h"
			clusterAutoscalerArgs := &exec.ClusterAutoscalerArgs{
				Cluster:              helper.StringPointer(clusterID),
				MaxPodGracePeriod:    helper.IntPointer(maxPodGracePeriod),
				PodPriorityThreshold: helper.IntPointer(podPriorityThreshold),
				MaxNodeProvisionTime: helper.StringPointer(maxNodeProvisionTime),
				ResourceLimits:       resourceLimits,
			}
			_, err = caService.Apply(clusterAutoscalerArgs)
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
				MaxNodesTotal: helper.IntPointer(maxNodesTotal),
			}
			maxPodGracePeriod = 5
			podPriorityThreshold = 3
			maxNodeProvisionTime = "60m"

			clusterAutoscalerArgs.MaxPodGracePeriod = helper.IntPointer(maxPodGracePeriod)
			clusterAutoscalerArgs.PodPriorityThreshold = helper.IntPointer(podPriorityThreshold)
			clusterAutoscalerArgs.MaxNodeProvisionTime = helper.StringPointer(maxNodeProvisionTime)
			clusterAutoscalerArgs.ResourceLimits = resourceLimits
			_, err = caService.Apply(clusterAutoscalerArgs)
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

	It("can be validated against HCP cluster - [id:72526]", ci.Medium, func() {
		if !profile.GetClusterType().HCP {
			Skip("Test can run only on Hosted cluster")
		}

		var err error
		caService, err = exec.NewClusterAutoscalerService(constants.HCPClusterAutoscalerDir)
		Expect(err).NotTo(HaveOccurred())

		if clusterAutoscalerStatusBefore == http.StatusOK {
			By("Delete current cluster autoscaler")
			caDeleteBody, err := cms.DeleteClusterAutoscaler(ci.RHCSConnection, clusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(caDeleteBody.Status()).To(Equal(http.StatusNoContent))
		}

		By("Try to create tuning config with empty cluster ID")
		args := &exec.ClusterAutoscalerArgs{
			Cluster:        helper.EmptyStringPointer,
			ResourceLimits: &exec.ResourceLimits{},
		}
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Attribute cluster cluster ID may not be empty/blank string, got:"))

		By("Try to create tuning config with wrong cluster ID")
		value := "wrong"
		args = &exec.ClusterAutoscalerArgs{
			Cluster:        helper.StringPointer(value),
			ResourceLimits: &exec.ResourceLimits{},
		}
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Cluster 'wrong' not found"))

		By("Create Autoscaler")
		args = &exec.ClusterAutoscalerArgs{
			Cluster:        helper.StringPointer(clusterID),
			ResourceLimits: &exec.ResourceLimits{},
		}
		_, err = caService.Apply(args)
		Expect(err).ToNot(HaveOccurred())

		By("Try to edit cluster with other cluster ID")
		clustersResp, err := cms.ListClusters(ci.RHCSConnection)
		Expect(err).ToNot(HaveOccurred())
		var otherClusterID string
		for _, cluster := range clustersResp.Items().Slice() {
			if cluster.ID() != clusterID {
				otherClusterID = cluster.ID()
				break
			}
		}
		if otherClusterID != "" {
			args = &exec.ClusterAutoscalerArgs{
				Cluster:        helper.StringPointer(otherClusterID),
				ResourceLimits: &exec.ResourceLimits{},
			}
			_, err = caService.Apply(args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute cluster, cannot be changed from"))
		} else {
			Logger.Info("No other cluster accessible for testing this change")
		}

		By("Try to edit autoscaler `max_node_provision_time` with negative value")
		args = &exec.ClusterAutoscalerArgs{
			Cluster:              helper.StringPointer(clusterID),
			MaxNodeProvisionTime: helper.StringPointer("-1h"),
			ResourceLimits:       &exec.ResourceLimits{},
		}
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Only positive durations are allowed"))
	})

	It("can be validated against Classic cluster - [id:76199]", ci.Medium, func() {
		if profile.GetClusterType().HCP {
			Skip("Test can run only on Classic cluster")
		}

		defaultClusterAutoscalerArgs := func() *exec.ClusterAutoscalerArgs {
			max := 1
			min := 0
			resourceRange := &exec.ResourceRange{
				Max: helper.IntPointer(max),
				Min: helper.IntPointer(min),
			}
			return &exec.ClusterAutoscalerArgs{
				Cluster: helper.StringPointer(clusterID),
				ResourceLimits: &exec.ResourceLimits{
					Cores:  resourceRange,
					Memory: resourceRange,
				},
				ScaleDown: &exec.ScaleDown{},
			}
		}

		var err error
		caService, err = exec.NewClusterAutoscalerService(constants.ClassicClusterAutoscalerDir)
		Expect(err).NotTo(HaveOccurred())

		if clusterAutoscalerStatusBefore == http.StatusOK {
			By("Delete current cluster autoscaler")
			caDeleteBody, err := cms.DeleteClusterAutoscaler(ci.RHCSConnection, clusterID)
			Expect(err).NotTo(HaveOccurred())
			Expect(caDeleteBody.Status()).To(Equal(http.StatusNoContent))
		}

		By("Try to create autoscaler with empty cluster ID")
		args := defaultClusterAutoscalerArgs()
		args.Cluster = helper.EmptyStringPointer
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Attribute cluster cluster ID may not be empty/blank string, got:"))

		By("Try to create autoscaler with wrong cluster ID")
		args = defaultClusterAutoscalerArgs()
		args.Cluster = helper.StringPointer("wrong")
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Cluster 'wrong' not found"))

		By("Create Autoscaler")
		args = defaultClusterAutoscalerArgs()
		_, err = caService.Apply(args)
		Expect(err).ToNot(HaveOccurred())

		By("Try to edit autoscaler with other cluster ID")
		clustersResp, err := cms.ListClusters(ci.RHCSConnection)
		Expect(err).ToNot(HaveOccurred())
		var otherClusterID string
		for _, cluster := range clustersResp.Items().Slice() {
			if cluster.ID() != clusterID {
				otherClusterID = cluster.ID()
				break
			}
		}
		if otherClusterID != "" {
			args = defaultClusterAutoscalerArgs()
			args.Cluster = helper.StringPointer(otherClusterID)
			_, err = caService.Apply(args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute cluster, cannot be changed from"))
		} else {
			Logger.Info("No other cluster accessible for testing this change")
		}

		By("Try to edit autoscaler `max_node_provision_time` with negative value")
		args = defaultClusterAutoscalerArgs()
		args.MaxNodeProvisionTime = helper.StringPointer("-1h")
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Only positive durations are allowed"))

		By("Try to edit autoscaler `unneeded_time` with negative value")
		args = defaultClusterAutoscalerArgs()
		args.ScaleDown.UnneededTime = helper.StringPointer("-1h")
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Only positive durations are allowed"))

		By("Try to edit autoscaler `delay_after_add` with negative value")
		args = defaultClusterAutoscalerArgs()
		args.ScaleDown.DelayAfterAdd = helper.StringPointer("-3h")
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Only positive durations are allowed"))

		By("Try to edit autoscaler `delay_after_delete` with negative value")
		args = defaultClusterAutoscalerArgs()
		args.ScaleDown.DelayAfterDelete = helper.StringPointer("-3h")
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Only positive durations are allowed"))

		By("Try to edit autoscaler `delay_after_failure` with negative value")
		args = defaultClusterAutoscalerArgs()
		args.ScaleDown.DelayAfterFailure = helper.StringPointer("-3h")
		_, err = caService.Apply(args)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Only positive durations are allowed"))
	})
})
