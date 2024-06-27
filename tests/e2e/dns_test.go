package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("DNS Domain", func() {
	var dnsService exec.DnsDomainService
	BeforeEach(func() {
		var err error
		dnsService, err = exec.NewDnsDomainService(constants.DNSDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dnsService.Destroy()
	})

	It("can create and destroy dnsdomain - [id:67570]",
		ci.Day2, ci.Medium, ci.FeatureIDP, ci.NonHCPCluster, func() {

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
