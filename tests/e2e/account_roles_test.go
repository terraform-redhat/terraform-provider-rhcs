package e2e

import (
	// nolint

	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

var _ = Describe("Edit Account roles", func() {
	defer GinkgoRecover()

	var (
		accService     exec.AccountRoleService
		profileHandler profilehandler.ProfileHandler
		majorVersion   string
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		accTFWorkspace := helper.GenerateRandomName("add-"+profileHandler.Profile().GetName(), 2)
		accService, err = exec.NewAccountRoleService(accTFWorkspace, profileHandler.Profile().GetClusterType())
		Expect(err).ToNot(HaveOccurred())

		By("Retrieve cluster version")
		clusterService, err := profileHandler.Services().GetClusterService()
		Expect(err).ToNot(HaveOccurred())
		cOut, err := clusterService.Output()
		Expect(err).ToNot(HaveOccurred())
		majorVersion = helper.GetMajorVersion(cOut.ClusterVersion)

	})
	AfterEach(func() {
		accService.Destroy()
	})

	It("can create account roles with default prefix - [id:65380]", ci.Day2, ci.Medium, func() {
		By("Create account roles with empty prefix")
		args := &exec.AccountRolesArgs{
			AccountRolePrefix: helper.EmptyStringPointer,
			OpenshiftVersion:  helper.StringPointer(majorVersion),
		}
		_, err := accService.Apply(args)
		Expect(err).ToNot(HaveOccurred())
		accRoleOutput, err := accService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(accRoleOutput.AccountRolePrefix).ToNot(BeEmpty())

		By("Create account roles with no prefix defined")
		args = &exec.AccountRolesArgs{
			AccountRolePrefix: nil,
			OpenshiftVersion:  helper.StringPointer(majorVersion),
		}
		_, err = accService.Apply(args)
		Expect(err).ToNot(HaveOccurred())
		accRoleOutput, err = accService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(accRoleOutput.AccountRolePrefix).ToNot(BeEmpty())

	})

	It("can delete account roles via account-role module - [id:63316]", ci.Day2, ci.Critical, func() {
		args := &exec.AccountRolesArgs{
			AccountRolePrefix: helper.StringPointer(helper.GenerateRandomName("OCP-63316", 10)),
			OpenshiftVersion:  helper.StringPointer(majorVersion),
		}
		_, err := accService.Apply(args)
		Expect(err).ToNot(HaveOccurred())
		accService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})
})

var _ = Describe("Create Account roles with shared vpc role", ci.Exclude, func() {
	defer GinkgoRecover()

	var (
		profileHandler   profilehandler.ProfileHandler
		accService       exec.AccountRoleService
		oidcOpService    exec.OIDCProviderOperatorRolesService
		dnsService       exec.DnsDomainService
		vpcService       exec.VPCService
		sharedVPCService exec.SharedVpcPolicyAndHostedZoneService
		majorVersion     string
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())
		clusterType := profileHandler.Profile().GetClusterType()

		tempWorkspace := helper.GenerateRandomName("ocp-67574"+profileHandler.Profile().GetName(), 2)
		Logger.Infof("Using temp workspace '%s' for creating resources", tempWorkspace)

		accService, err = exec.NewAccountRoleService(tempWorkspace, clusterType)
		Expect(err).ToNot(HaveOccurred())

		oidcOpService, err = exec.NewOIDCProviderOperatorRolesService(tempWorkspace, clusterType)
		Expect(err).ToNot(HaveOccurred())

		dnsService, err = exec.NewDnsDomainService(tempWorkspace, clusterType)
		Expect(err).ToNot(HaveOccurred())

		vpcService, err = exec.NewVPCService(tempWorkspace, clusterType)
		Expect(err).ToNot(HaveOccurred())

		sharedVPCService, err = exec.NewSharedVpcPolicyAndHostedZoneService(tempWorkspace, clusterType)
		Expect(err).ToNot(HaveOccurred())

		By("Retrieve cluster version")
		clusterService, err := profileHandler.Services().GetClusterService()
		Expect(err).ToNot(HaveOccurred())
		cOut, err := clusterService.Output()
		Expect(err).ToNot(HaveOccurred())
		majorVersion = helper.GetMajorVersion(cOut.ClusterVersion)
	})
	AfterEach(func() {
		sharedVPCService.Destroy()
		dnsService.Destroy()
		vpcService.Destroy()
		oidcOpService.Destroy()
		accService.Destroy()
	})

	It("create and destroy account roles for shared vpc - [id:67574]", ci.Day2, ci.Medium, func() {
		if profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Classic cluster")
		}
		By("Create account role without shared vpc role arn")
		accArgs := &exec.AccountRolesArgs{
			AccountRolePrefix: helper.StringPointer(helper.GenerateRandomName("OCP-67574", 2)),
			OpenshiftVersion:  helper.StringPointer(majorVersion),
		}
		_, err := accService.Apply(accArgs)
		Expect(err).ToNot(HaveOccurred())
		accRoleOutput, err := accService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(accRoleOutput.AccountRolePrefix).Should(Equal(accArgs.AccountRolePrefix))

		By("Create operator role")
		oidcOpArgs := &exec.OIDCProviderOperatorRolesArgs{
			AccountRolePrefix:  helper.StringPointer(accRoleOutput.AccountRolePrefix),
			OperatorRolePrefix: helper.StringPointer(accRoleOutput.AccountRolePrefix),
			OIDCConfig:         helper.StringPointer(profileHandler.Profile().GetOIDCConfig()),
		}
		_, err = oidcOpService.Apply(oidcOpArgs)
		Expect(err).ToNot(HaveOccurred())
		oidcOpOutput, err := oidcOpService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(oidcOpOutput.OperatorRolePrefix).Should(Equal(oidcOpArgs.OperatorRolePrefix))

		By("verify policy permission")
		roleName := strings.Split(oidcOpOutput.IngressOperatorRoleArn, "/")[1]
		policies, err := helper.GetRoleAttachedPolicies(roleName)
		Expect(policies[0].Statement[0].Action).ToNot(ContainElement("sts:AssumeRole"))

		By("prepare route53")
		dnsDomainArgs := &exec.DnsDomainArgs{}
		_, err = dnsService.Apply(dnsDomainArgs)
		Expect(err).ToNot(HaveOccurred())
		dnsDomainOutput, err := dnsService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Create shared vpc role")
		clusterName := "OCP-67574"
		vpcArgs := &exec.VPCArgs{
			NamePrefix:                helper.StringPointer(clusterName),
			AWSRegion:                 helper.StringPointer(profileHandler.Profile().GetRegion()),
			VPCCIDR:                   helper.StringPointer(profilehandler.DefaultVPCCIDR),
			AWSSharedCredentialsFiles: helper.StringSlicePointer([]string{config.GetSharedVpcAWSSharedCredentialsFile()}),
		}
		_, err = vpcService.Apply(vpcArgs)
		Expect(err).ToNot(HaveOccurred())
		vpcOutput, err := vpcService.Output()
		Expect(err).ToNot(HaveOccurred())
		sharedVPCArgs := &exec.SharedVpcPolicyAndHostedZoneArgs{
			SharedVpcAWSSharedCredentialsFiles: helper.StringSlicePointer([]string{config.GetSharedVpcAWSSharedCredentialsFile()}),
			Region:                             helper.StringPointer(profileHandler.Profile().GetRegion()),
			ClusterName:                        helper.StringPointer(clusterName),
			DnsDomainId:                        helper.StringPointer(dnsDomainOutput.DnsDomainId),
			IngressOperatorRoleArn:             helper.StringPointer(oidcOpOutput.IngressOperatorRoleArn),
			InstallerRoleArn:                   helper.StringPointer(accRoleOutput.InstallerRoleArn),
			ClusterAWSAccount:                  helper.StringPointer(accRoleOutput.AWSAccountId),
			VpcId:                              helper.StringPointer(vpcOutput.VPCID),
			Subnets:                            helper.StringSlicePointer(vpcOutput.PrivateSubnets),
		}
		_, err = sharedVPCService.Apply(sharedVPCArgs)
		Expect(err).ToNot(HaveOccurred())
		sharedVPCOutput, err := sharedVPCService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Add shared vpc role arn to account role")
		accArgs = &exec.AccountRolesArgs{
			AccountRolePrefix: helper.StringPointer(accRoleOutput.AccountRolePrefix),
			OpenshiftVersion:  helper.StringPointer(majorVersion),
			SharedVpcRoleArn:  helper.StringPointer(sharedVPCOutput.SharedRole),
		}
		_, err = accService.Apply(accArgs)
		Expect(err).ToNot(HaveOccurred())
		accRoleOutput, err = accService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("verify policy permission")
		policies, err = helper.GetRoleAttachedPolicies(roleName)
		Expect(policies[0].Statement[2].Action).To(ContainSubstring("sts:AssumeRole"))
	})
})
