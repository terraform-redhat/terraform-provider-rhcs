package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("DNS Domain", func() {
	var (
		dnsService     exec.DnsDomainService
		profileHandler profilehandler.ProfileHandler
	)
	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		dnsService, err = profileHandler.Services().GetDnsDomainService()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dnsService.Destroy()
	})

	It("can create and destroy dnsdomain - [id:67570]",
		ci.Day2, ci.Medium, ci.FeatureIDP, func() {
			if profileHandler.Profile().IsHCP() {
				Skip("Test can run only on Classic cluster")
			}

			By("Retrieve DNS creation args")
			dnsArgs, err := dnsService.ReadTFVars()
			if err != nil {
				dnsArgs = &exec.DnsDomainArgs{}
			}

			By("Create/Apply dns-domain resource by terraform")
			_, err = dnsService.Apply(dnsArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Destroy dns-domain resource by terraform")
			_, err = dnsService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		})
})
