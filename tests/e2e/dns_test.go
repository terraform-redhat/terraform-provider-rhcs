// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
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

		// Use a dedicated temp workspace so shared-vpc day1 PrepareRoute53 DNS
		// (profile workspace) is not destroyed while the cluster still exists.
		// Same isolation pattern as id:67574 in account_roles_test.go.
		tempWorkspace := helper.GenerateRandomName("ocp-67570"+profileHandler.Profile().GetName(), 2)
		Logger.Infof("Using temp workspace '%s' for creating resources", tempWorkspace)

		dnsService, err = exec.NewDnsDomainService(tempWorkspace, profileHandler.Profile().GetClusterType())
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

			By("Create/Apply dns-domain resource by terraform")
			_, err := dnsService.Apply(&exec.DnsDomainArgs{})
			Expect(err).ToNot(HaveOccurred())

			By("Destroy dns-domain resource by terraform")
			_, err = dnsService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		})
})
