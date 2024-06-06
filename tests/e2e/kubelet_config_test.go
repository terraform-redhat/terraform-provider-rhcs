package e2e

import (

	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("Kubelet config", ci.NonHCPCluster, func() {
	defer GinkgoRecover()

	var kcService exe.KubeletConfigService

	BeforeEach(func() {
		var err error
		kcService, err = exe.NewKubeletConfigService(CON.KubeletConfigDir)
		Expect(err).ToNot(HaveOccurred())
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

		By("Verify the created kubeletconfig")
		kubeletConfig, err := cms.RetrieveKubeletConfig(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeletConfig.PodPidsLimit()).To(Equal(podPidsLimit))

		By("Update kubeletConfig")
		podPidsLimit = 12346
		kcArgs.PodPidsLimit = helper.IntPointer(podPidsLimit)
		_, err = kcService.Apply(kcArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the created kubeletconfig")
		kubeletConfig, err = cms.RetrieveKubeletConfig(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(kubeletConfig.PodPidsLimit()).To(Equal(podPidsLimit))

		By("Destroy the kubeletconfig")
		_, err = kcService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Verify the created kubeletconfig")
		_, err = cms.RetrieveKubeletConfig(ci.RHCSConnection, clusterID)
		Expect(err).To(HaveOccurred())

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
	})
})
