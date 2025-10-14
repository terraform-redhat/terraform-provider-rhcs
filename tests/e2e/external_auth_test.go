package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var (
	consoleClientID         = "abc"
	consoleClientSecret     = "efgh"
	issuerURL               = "https://local.com"
	issuerAudience          = "abc"
	ca                      = "----BEGIN CERTIFICATE-----MIIDNTCCAh2gAwIBAgIUAegBu2L2aoOizuGxf/" +
		"fxBCU10oswDQYJKoZIhvcNAQELS3nCXMvI8q0E-----END CERTIFICATE-----"
	groupClaim              = "groups"
	userNameClaim           = "email"
	claimValidationRuleClaim = "claim1:rule1"
)

var _ = Describe("External Authentication", ci.Day1, ci.FeatureExternalAuth, func() {
	defer GinkgoRecover()
	var (
		profileHandler               profilehandler.ProfileHandler
		externalAuthProviderService  exec.ExternalAuthProviderService
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if !profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Hosted cluster")
		}

		if !profileHandler.Profile().IsExternalAuthEnabled() {
			Skip("Test requires external auth enabled profile")
		}

		// Initialize external auth provider service
		externalAuthProviderService, err = exec.NewExternalAuthProviderService(
			profileHandler.Profile().GetClusterManifestsDir(), 
			constants.HCPClusterType,
		)
		Expect(err).ToNot(HaveOccurred())
	})

	It("Verify external authentication is enabled on cluster creation - [id:external-auth-01]", ci.Critical, func() {
		By("Retrieve cluster information")
		clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		By("Verify external authentication config is enabled")
		cluster := clusterResp.Body()
		externalAuthConfig, exists := cluster.GetExternalAuthConfig()
		Expect(exists).To(BeTrue(), "External auth config should exist")
		Expect(externalAuthConfig.Enabled()).To(BeTrue(), "External auth should be enabled")

		Logger.Infof("Successfully verified external authentication is enabled for cluster %s", clusterID)
	})

	It("Verify cluster output contains external auth information - [id:external-auth-02]", ci.Medium, func() {
		By("Get cluster service output")
		clusterService, err := profileHandler.Services().GetClusterService()
		Expect(err).ToNot(HaveOccurred())

		clusterOutput, err := clusterService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Verify external auth providers enabled field is present and true")
		Expect(clusterOutput.ExternalAuthProvidersEnabled).ToNot(BeNil(), "ExternalAuthProvidersEnabled should not be nil")
		Expect(*clusterOutput.ExternalAuthProvidersEnabled).To(BeTrue(), "ExternalAuthProvidersEnabled should be true")

		Logger.Infof("Successfully verified external auth in cluster output")
	})

	It("Create and manage external auth provider - [id:external-auth-03]", ci.Medium, func() {
		By("Creating external auth provider with basic configuration")
		externalAuthProviderArgs := &exec.ExternalAuthProviderArgs{
			Cluster:         helper.StringPointer(clusterID),
			ID:              helper.StringPointer("test-provider"),
			IssuerURL:       helper.StringPointer(issuerURL),
			IssuerAudiences: &[]string{issuerAudience},
		}

		_, err := externalAuthProviderService.Apply(externalAuthProviderArgs)
		Expect(err).ToNot(HaveOccurred())
		Logger.Infof("Successfully created external auth provider")

		By("Verifying external auth provider was created")
		output, err := externalAuthProviderService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(output.ID).To(Equal("test-provider"))
		Expect(output.Cluster).To(Equal(clusterID))

		By("Verifying external auth provider exists via OCM API")
		cluster, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())

		externalAuths, err := cms.RHCSConnection.ClustersMgmt().V1().
			Clusters().Cluster(clusterID).
			ExternalAuthConfig().ExternalAuths().List().Send()
		Expect(err).ToNot(HaveOccurred())
		
		found := false
		externalAuths.Items().Each(func(provider *cmv1.ExternalAuth) bool {
			if provider.ID() == "test-provider" {
				found = true
				Expect(provider.Issuer().URL()).To(Equal(issuerURL))
				Expect(provider.Issuer().Audiences()).To(ContainElement(issuerAudience))
				return false
			}
			return true
		})
		Expect(found).To(BeTrue(), "External auth provider should exist in cluster")

		By("Cleaning up external auth provider")
		_, err = externalAuthProviderService.Destroy()
		Expect(err).ToNot(HaveOccurred())
		Logger.Infof("Successfully destroyed external auth provider")
	})

	It("Create external auth provider with full configuration - [id:external-auth-04]", ci.Medium, func() {
		By("Creating external auth provider with complete configuration")
		claimRules := []exec.ClaimValidationRule{
			exec.NewClaimValidationRule("aud", issuerAudience),
			exec.NewClaimValidationRule("iss", issuerURL),
		}

		externalAuthProviderArgs := &exec.ExternalAuthProviderArgs{
			Cluster:                 helper.StringPointer(clusterID),
			ID:                      helper.StringPointer("full-config-provider"),
			IssuerURL:               helper.StringPointer(issuerURL),
			IssuerAudiences:         &[]string{issuerAudience, "api-gateway"},
			IssuerCA:                helper.StringPointer(ca),
			ConsoleClientID:         helper.StringPointer(consoleClientID),
			ConsoleClientSecret:     helper.StringPointer(consoleClientSecret),
			ClaimMappingUsernameKey: helper.StringPointer(userNameClaim),
			ClaimMappingGroupsKey:   helper.StringPointer(groupClaim),
			ClaimValidationRules:    &claimRules,
		}

		_, err := externalAuthProviderService.Apply(externalAuthProviderArgs)
		Expect(err).ToNot(HaveOccurred())
		Logger.Infof("Successfully created external auth provider with full configuration")

		By("Verifying external auth provider configuration via OCM API")
		externalAuths, err := cms.RHCSConnection.ClustersMgmt().V1().
			Clusters().Cluster(clusterID).
			ExternalAuthConfig().ExternalAuths().List().Send()
		Expect(err).ToNot(HaveOccurred())

		found := false
		externalAuths.Items().Each(func(provider *cmv1.ExternalAuth) bool {
			if provider.ID() == "full-config-provider" {
				found = true
				Expect(provider.Issuer().URL()).To(Equal(issuerURL))
				Expect(provider.Issuer().Audiences()).To(ContainElements(issuerAudience, "api-gateway"))
				Expect(provider.Issuer().CA()).To(Equal(ca))
				
				if clients := provider.Clients(); len(clients) > 0 {
					Expect(clients[0].ID()).To(Equal(consoleClientID))
					Expect(clients[0].Secret()).To(Equal(consoleClientSecret))
				}
				
				if claim := provider.Claim(); claim != nil {
					if mappings := claim.Mappings(); mappings != nil {
						if username := mappings.UserName(); username != nil {
							Expect(username.Claim()).To(Equal(userNameClaim))
						}
						if groups := mappings.Groups(); groups != nil {
							Expect(groups.Claim()).To(Equal(groupClaim))
						}
					}
					
					if rules := claim.ValidationRules(); len(rules) > 0 {
						Expect(len(rules)).To(Equal(2))
					}
				}
				return false
			}
			return true
		})
		Expect(found).To(BeTrue(), "Full configuration external auth provider should exist")

		By("Cleaning up external auth provider")
		_, err = externalAuthProviderService.Destroy()
		Expect(err).ToNot(HaveOccurred())
		Logger.Infof("Successfully destroyed external auth provider with full configuration")
	})
})