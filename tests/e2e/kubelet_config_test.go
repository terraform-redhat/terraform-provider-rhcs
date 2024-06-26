package e2e

import (

	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("Kubelet config", func() {
	defer GinkgoRecover()

	var kcService exe.KubeletConfigService

	var cluster *cmv1.Cluster

	BeforeEach(func() {

		var err error
		kcService, err = exe.NewKubeletConfigService(CON.KubeletConfigDir)
		Expect(err).ToNot(HaveOccurred())

		resp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		cluster = resp.Body()

	})

	AfterEach(func() {
		kcService.Destroy()
	})

	It("can be created - [id:70128]", ci.Day2, ci.High, func() {
		By("Create kubeletconfig")
		podPidsLimit := 12345
		kcArgs := &exe.KubeletConfigArgs{
			PodPidsLimit: helper.IntPointer(podPidsLimit),
			Cluster:      helper.StringPointer(clusterID),
		}

		_, err := kcService.Apply(kcArgs)
		Expect(err).ToNot(HaveOccurred())
		kubeletconfigs, err := kcService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeletconfigs).ToNot(BeEmpty())

		By("Verify the created kubeletconfig")
		if cluster.Hypershift().Enabled() {
			for _, kubeConfig := range kubeletconfigs {
				kubeletConfig, err := cms.RetrieveHCPKubeletConfig(ci.RHCSConnection, clusterID, kubeConfig.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(kubeletConfig.PodPidsLimit()).To(Equal(podPidsLimit))
			}
		} else {
			kubeletConfig, err := cms.RetrieveKubeletConfig(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeletConfig.PodPidsLimit()).To(Equal(podPidsLimit))
		}

		By("Update kubeletConfig")
		podPidsLimit = 12346
		kcArgs.PodPidsLimit = helper.IntPointer(podPidsLimit)
		_, err = kcService.Apply(kcArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the updated kubeletconfig")
		if cluster.Hypershift().Enabled() {
			for _, kubeConfig := range kubeletconfigs {
				kubeletConfig, err := cms.RetrieveHCPKubeletConfig(ci.RHCSConnection, clusterID, kubeConfig.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(kubeletConfig.PodPidsLimit()).To(Equal(podPidsLimit))
			}
		} else {
			kubeletConfig, err := cms.RetrieveKubeletConfig(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(kubeletConfig.PodPidsLimit()).To(Equal(podPidsLimit))
		}

		By("Destroy the kubeletconfig")
		_, err = kcService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Verify the created kubeletconfig")
		if cluster.Hypershift().Enabled() {
			for _, kubeConfig := range kubeletconfigs {
				_, err := cms.RetrieveHCPKubeletConfig(ci.RHCSConnection, clusterID, kubeConfig.ID)
				Expect(err).To(HaveOccurred())
			}
		} else {
			_, err := cms.RetrieveKubeletConfig(ci.RHCSConnection, clusterID)
			Expect(err).To(HaveOccurred())
		}

		By("Create multiple kubeletconfigs to the cluster")
		kcArgs = &exe.KubeletConfigArgs{
			PodPidsLimit:        helper.IntPointer(podPidsLimit),
			Cluster:             helper.StringPointer(clusterID),
			KubeLetConfigNumber: helper.IntPointer(2),
			NamePrefix:          helper.StringPointer("kube-70128"),
		}
		_, err = kcService.Apply(kcArgs)
		if cluster.Hypershift().Enabled() {
			Expect(err).ToNot(HaveOccurred())
			kubeletconfigs, err = kcService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(kubeletconfigs)).To(Equal(*kcArgs.KubeLetConfigNumber))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("classic cluster can only have 1 kubeletconfig"))
		}
	})

	It("will validate well - [id:70129]", ci.Day2, ci.Medium, func() {
		By("Create kubeletconfig")
		podPidsLimit := 1
		kcArgs := &exe.KubeletConfigArgs{
			PodPidsLimit: helper.IntPointer(podPidsLimit),
			Cluster:      helper.StringPointer(clusterID),
		}

		output, err := kcService.Plan(kcArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).Should(ContainSubstring("The requested podPidsLimit of '%d' is below the minimum allowable value of", *kcArgs.PodPidsLimit))

		kcArgs.PodPidsLimit = helper.IntPointer(1234567890)
		output, err = kcService.Plan(kcArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).Should(ContainSubstring("The requested podPidsLimit of '%d' is above the default maximum value", *kcArgs.PodPidsLimit))

		kcArgs.PodPidsLimit = helper.IntPointer(1234567)
		output, err = kcService.Plan(kcArgs)
		Expect(err).ToNot(HaveOccurred())

		Expect(output).Should(ContainSubstring("The requested podPidsLimit of '%d' is above the default maximum of", *kcArgs.PodPidsLimit))

		if cluster.Hypershift().Enabled() {
			By("Create more than 100 kubeletconfig is not allowed")
			kcArgs.PodPidsLimit = helper.IntPointer(4096)
			kcArgs.KubeLetConfigNumber = helper.IntPointer(300)
			kcArgs.NamePrefix = helper.StringPointer("kc-70129")
			_, err = kcService.Apply(kcArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("Maximum allowed is '100'"))
		}

	})
})
