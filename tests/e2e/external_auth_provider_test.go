// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("External Auth Provider", ci.Day2, ci.FeatureExternalAuth, func() {
	defer GinkgoRecover()

	var externalAuthProviderService exec.ExternalAuthProviderService

	BeforeEach(func() {
		profileHandler, err := profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if !profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Hosted cluster")
		}
		if !profileHandler.Profile().IsExternalAuthEnabled() {
			Skip("Test requires external auth enabled profile")
		}

		externalAuthProviderService, err = profileHandler.Services().GetExternalAuthProviderService()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if externalAuthProviderService != nil {
			_, err := externalAuthProviderService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("can create a simple external auth provider on a ready HCP cluster - [id:67579]",
		ci.Critical, func() {
			const (
				issuerURL      = "https://local.com"
				issuerAudience = "abc"
				groupsClaim    = "groups"
				usernameClaim  = "email"
				validationRule = "claim1:rule1"
			)
			name := helper.GenerateRandomName("provider", 2)
			args := &exec.ExternalAuthProviderArgs{
				Cluster:                   new(clusterID),
				Name:                      new(name),
				IssuerURL:                 new(issuerURL),
				IssuerAudiences:           &[]string{issuerAudience},
				ClaimMappingGroupsClaim:   new(groupsClaim),
				ClaimMappingUsernameClaim: new(usernameClaim),
				ClaimValidationRule:       &[]string{validationRule},
			}

			By("Create an external authentication provider")
			_, err := externalAuthProviderService.Apply(args)
			Expect(err).ToNot(HaveOccurred())

			By("Verify Terraform state")
			output, err := externalAuthProviderService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(output.ID).To(Equal(name))
			Expect(output.Name).To(Equal(name))
			Expect(output.Cluster).To(Equal(clusterID))
			Expect(output.IssuerURL).To(Equal(issuerURL))
			Expect(output.IssuerAudiences).To(ConsistOf(issuerAudience))
			Expect(output.ClaimMappingGroupsClaim).To(Equal(groupsClaim))
			Expect(output.ClaimMappingUsernameClaim).To(Equal(usernameClaim))
			Expect(output.ClaimValidationRule).To(ConsistOf(validationRule))

			By("Verify the provider through the OCM API")
			provider, err := cms.RetrieveClusterExternalAuthProvider(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(provider.ID()).To(Equal(name))
			Expect(provider.Issuer().URL()).To(Equal(issuerURL))
			Expect(provider.Issuer().Audiences()).To(ConsistOf(issuerAudience))
			Expect(provider.Claim().Mappings().Groups().Claim()).To(Equal(groupsClaim))
			Expect(provider.Claim().Mappings().UserName().Claim()).To(Equal(usernameClaim))
			Expect(provider.Claim().ValidationRules()).To(HaveLen(1))
			Expect(provider.Claim().ValidationRules()[0].Claim()).To(Equal("claim1"))
			Expect(provider.Claim().ValidationRules()[0].RequiredValue()).To(Equal("rule1"))
		})
})
