package e2e

import (
	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Trusted IPs", func() {
	defer GinkgoRecover()

	var (
		err               error
		trustedIPsService exec.TrustedIPsService
		profileHandler    profilehandler.ProfileHandler
	)

	BeforeEach(func() {
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		trustedIPsService, err = profileHandler.Services().GetTrustedIPsService()
		Expect(err).ToNot(HaveOccurred())
	})

	It("can list trusted IPs - [id:75233]", ci.Day1Supplemental, ci.Medium, func() {
		By("Get list of trusted IPs")
		args := exec.TrustedIPsArgs{}
		_, err = trustedIPsService.Apply(&args)
		Expect(err).ToNot(HaveOccurred())

		By("Make sure the output isn't empty")
		output, err := trustedIPsService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(output.TrustedIPs.Items).ToNot(BeEmpty())

		By("Make sure that the IDs aren't empty")
		for _, item := range output.TrustedIPs.Items {
			Expect(item.Id).ToNot(BeEmpty())
		}

		By("Make sure at least one IP is enabled")
		var anyEnabled bool = false
		for _, item := range output.TrustedIPs.Items {
			if item.Enabled {
				anyEnabled = true
				break
			}
		}
		Expect(anyEnabled).To(BeTrue())
	})
})
