package e2e

import (
	"encoding/base32"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/segmentio/ksuid"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

var _ = Describe("Break Glass Credential", ci.Day2, ci.FeatureBreakGlassCredential, func() {
	defer GinkgoRecover()
	var (
		profileHandler profilehandler.ProfileHandler
		bgcService     exe.BreakGlassCredentialService
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

		bgcService, err = profileHandler.Services().GetBreakGlassCredentialService()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if bgcService != nil {
			bgcService.Destroy()
		}
	})

	It("Verify break glass credential can be created/imported - [id:break-glass-credential-01]", ci.Day2, ci.Critical, func() {
		By("Create break glass credential")
		bgcArgs := &exe.BreakGlassCredentialArgs{
			Cluster: helper.StringPointer(clusterID),
		}
		_, err := bgcService.Apply(bgcArgs)
		Expect(err).ToNot(HaveOccurred())
		bgcOutput, err := bgcService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(bgcOutput.Kubeconfig).ToNot(BeEmpty())
		Expect(bgcOutput.Status).To(Equal("issued"))

		By("Verify generated kubeconfig can be used to access cluster")
		f, err := os.CreateTemp("", "kubeconfig")
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		defer os.Remove(f.Name())
		_, err = f.WriteString(bgcOutput.Kubeconfig)
		Expect(err).ToNot(HaveOccurred())

		stdout, stderr, err := helper.RunCMD(fmt.Sprintf("oc get pods -A --kubeconfig %s", f.Name()))
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).ToNot(BeEmpty())
		Expect(stderr).To(BeEmpty())

		By("Destroy break glass credential")
		_, err = bgcService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Create break glass credential with custom arguments")
		// the break glass credential can only be revoked but not removed
		// so we use a random username to avoid conflicts
		username := base32.StdEncoding.EncodeToString(ksuid.New().Bytes())
		bgcArgs = &exe.BreakGlassCredentialArgs{
			Cluster:            helper.StringPointer(clusterID),
			Username:           helper.StringPointer(username),
			ExpirationDuration: helper.StringPointer("1h"),
		}
		_, err = bgcService.Apply(bgcArgs)
		Expect(err).ToNot(HaveOccurred())
		bgcOutput, err = bgcService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(bgcOutput.Status).To(Equal("issued"))
		Expect(bgcOutput.Username).To(Equal(username))

		Logger.Infof("Successfully verified break glass credential can be created for cluster %s", clusterID)
	})

	It("Verify break glass credential creation is validated correctly - [id:break-glass-credential-02]", ci.Day2, ci.Medium, func() {
		By("Create break glass credential with invalid username")
		bgcArgs := &exe.BreakGlassCredentialArgs{
			Cluster:  helper.StringPointer(clusterID),
			Username: helper.StringPointer("test user"),
		}
		output, err := bgcService.Plan(bgcArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).To(ContainSubstring("The username 'test user' must respect the regexp '^[a-zA-Z0-9-.]*$'"))

		By("Create break glass credential with invalid expiration duration")
		bgcArgs = &exe.BreakGlassCredentialArgs{
			Cluster:            helper.StringPointer(clusterID),
			ExpirationDuration: helper.StringPointer("25h"),
		}
		output, err = bgcService.Plan(bgcArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).To(ContainSubstring("The expiration duration needs to be at maximum 24 hours"))
	})
})
