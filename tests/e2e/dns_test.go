package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("DNS Domain", func() {
	var dnsService *exec.DnsService
	BeforeEach(func() {
		var err error
		dnsService, err = exec.NewDnsDomainService(constants.DNSDir)
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		err := dnsService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})
	It("can create and destroy dnsdomain - [id:67570]",
		ci.Day2, ci.Medium, ci.FeatureIDP, ci.NonHCPCluster, func() {

			By("Create/Apply dns-domain resource by terraform")
			dnsArgs := &exec.DnsDomainArgs{}
			err := dnsService.Create(dnsArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Destroy dns-domain resource by terraform")
			err = dnsService.Destroy(dnsArgs)
			Expect(err).ToNot(HaveOccurred())
		})
})
