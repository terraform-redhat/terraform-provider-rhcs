package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("External Authentication", ci.Day1, ci.FeatureExternalAuth, func() {
	defer GinkgoRecover()
	var (
		profileHandler profilehandler.ProfileHandler
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
})
