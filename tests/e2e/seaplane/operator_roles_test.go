// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:staticcheck
	. "github.com/onsi/gomega"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
)

// Acceptance coverage for the rhcs_seaplane_operator_roles resource and data
// source (ROSAENG-75451): create, read (refresh), discovery, import, and
// delete, without requiring a cluster. The operator role definitions and
// trust policies come from the shared seaplane iam/ module, so the test only
// needs an OIDC config to supply a real issuer URL.
var _ = Describe("Seaplane operator roles", func() {
	var profileHandler profilehandler.ProfileHandler

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).NotTo(HaveOccurred())
		if profileHandler.Profile().GetClusterType() != constants.SEAPLANE {
			Skip("The operator roles acceptance test requires a Seaplane profile")
		}
	})

	It("creates, discovers, imports, and destroys the operator role set", func() {
		seaplaneURL := requireEnv("SEAPLANE_URL")
		awsRegion := awsRegionRE.FindString(seaplaneURL)
		Expect(awsRegion).NotTo(BeEmpty(),
			fmt.Sprintf("cannot derive AWS region from SEAPLANE_URL %q", seaplaneURL))

		clusterName := fmt.Sprintf("sp-oproles-%d", time.Now().Unix())
		oidcConfigName := clusterName + ".oidc"
		operatorRolesPrefix := clusterName

		oidcSvc, err := exec.NewSeaplaneOidcConfigService(clusterName + "-oidc")
		Expect(err).NotTo(HaveOccurred())
		rolesSvc, err := exec.NewSeaplaneOperatorRolesService(clusterName + "-roles")
		Expect(err).NotTo(HaveOccurred())
		discoverySvc, err := exec.NewSeaplaneOperatorRolesDiscoveryService(clusterName + "-discovery")
		Expect(err).NotTo(HaveOccurred())
		importSvc, err := exec.NewSeaplaneOperatorRolesService(clusterName + "-import")
		Expect(err).NotTo(HaveOccurred())

		DeferCleanup(func() {
			By("Teardown: destroy operator roles")
			_, err := rolesSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())

			By("Teardown: destroy OIDC config")
			_, err = oidcSvc.Destroy()
			Expect(err).NotTo(HaveOccurred())
		})

		By("creating a managed OIDC config to supply the issuer URL")
		_, err = oidcSvc.Apply(&exec.SeaplaneOidcConfigArgs{
			SeaplaneURL: &seaplaneURL,
			Name:        &oidcConfigName,
		})
		Expect(err).NotTo(HaveOccurred())

		oidcOut, err := oidcSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(oidcOut.IssuerURL).NotTo(BeEmpty(), "managed OidcConfig should compute an issuer URL")

		rolesArgs := &exec.SeaplaneOperatorRolesArgs{
			SeaplaneURL:         &seaplaneURL,
			OperatorRolesPrefix: &operatorRolesPrefix,
			OIDCIssuerURL:       &oidcOut.IssuerURL,
		}

		By("creating the operator role set")
		_, err = rolesSvc.Apply(rolesArgs)
		Expect(err).NotTo(HaveOccurred())

		rolesOut, err := rolesSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(rolesOut.RoleCount).To(Equal(8), "expected 7 operator roles plus the worker role")
		Expect(rolesOut.WorkerRoleARN).NotTo(BeEmpty())
		Expect(rolesOut.InstanceProfileName).To(Equal(operatorRolesPrefix + "-ROSA-Worker-Role"))
		Expect(rolesOut.DefinitionVersion).NotTo(BeEmpty())

		By("refreshing the role set (reconciliation must be a no-op)")
		planOut, err := rolesSvc.Plan(rolesArgs)
		Expect(err).NotTo(HaveOccurred())
		Logger.Infof("[verify] refresh plan output: %s", planOut)

		By("discovering the role set through the data source")
		_, err = discoverySvc.Apply(&exec.SeaplaneOperatorRolesDiscoveryArgs{
			OperatorRolesPrefix: &operatorRolesPrefix,
		})
		Expect(err).NotTo(HaveOccurred())

		discoveryOut, err := discoverySvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(discoveryOut.RoleCount).To(Equal(8))
		Expect(discoveryOut.DefinitionVersion).To(Equal(rolesOut.DefinitionVersion))

		By("importing the existing role set by prefix")
		_, err = importSvc.Import(rolesArgs, operatorRolesPrefix)
		Expect(err).NotTo(HaveOccurred())

		importOut, err := importSvc.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(importOut.RoleCount).To(Equal(8))

		By("planning after import (no changes expected)")
		importPlan, err := importSvc.Plan(rolesArgs)
		Expect(err).NotTo(HaveOccurred())
		Logger.Infof("[verify] post-import plan output: %s", importPlan)

		By("destroying the imported state copy (roles must remain)")
		_, err = importSvc.Destroy()
		Expect(err).NotTo(HaveOccurred())

		By("destroying the operator role set")
		_, err = rolesSvc.Destroy()
		Expect(err).NotTo(HaveOccurred())

		By("confirming the data source no longer finds the roles")
		_, err = discoverySvc.Apply(&exec.SeaplaneOperatorRolesDiscoveryArgs{
			OperatorRolesPrefix: &operatorRolesPrefix,
		})
		Expect(err).To(HaveOccurred(), "discovery should fail once the roles are deleted")
	})
})
