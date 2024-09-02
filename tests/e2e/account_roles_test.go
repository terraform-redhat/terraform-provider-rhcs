package e2e

import (
	// nolint

	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("Edit Account roles", func() {
	// To improve with OCM-8602
	defer GinkgoRecover()

	var err error
	var profile *ci.Profile
	var accService exec.AccountRoleService

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()
		Expect(err).ToNot(HaveOccurred())

		accService, err = exec.NewAccountRoleService(constants.GetAddAccountRoleDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		accService.Destroy()
	})

	It("can create account roles with default prefix - [id:65380]", ci.Day2, ci.Medium, func() {
		By("Create account roles with empty prefix")
		args := &exec.AccountRolesArgs{
			AccountRolePrefix: helper.EmptyStringPointer,
			OpenshiftVersion:  helper.StringPointer(profile.MajorVersion),
			ChannelGroup:      helper.StringPointer(profile.ChannelGroup),
		}
		_, err := accService.Apply(args)
		Expect(err).ToNot(HaveOccurred())
		accRoleOutput, err := accService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(accRoleOutput.AccountRolePrefix).ToNot(BeEmpty())

		By("Create account roles with no prefix defined")
		args = &exec.AccountRolesArgs{
			AccountRolePrefix: nil,
			OpenshiftVersion:  helper.StringPointer(profile.MajorVersion),
			ChannelGroup:      helper.StringPointer(profile.ChannelGroup),
		}
		_, err = accService.Apply(args)
		Expect(err).ToNot(HaveOccurred())
		accRoleOutput, err = accService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(accRoleOutput.AccountRolePrefix).ToNot(BeEmpty())

	})

	It("can delete account roles via account-role module - [id:63316]", ci.Day2, ci.Critical, func() {
		args := &exec.AccountRolesArgs{
			AccountRolePrefix: helper.StringPointer("OCP-63316"),
			OpenshiftVersion:  helper.StringPointer(profile.MajorVersion),
			ChannelGroup:      helper.StringPointer(profile.ChannelGroup),
		}
		_, err := accService.Apply(args)
		Expect(err).ToNot(HaveOccurred())
		accService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})
})

var _ = Describe("Create Account roles with shared vpc role", ci.Exclude, func() {
	// TODO Should be re-enabled with OCM-8602 and usage of different workspaces
	// Else this test does alter the account role and VPC configuration ...
	defer GinkgoRecover()

	var err error
	var profile *ci.Profile
	var accService exec.AccountRoleService
	var oidcOpService exec.OIDCProviderOperatorRolesService
	var dnsService exec.DnsDomainService
	var vpcService exec.VPCService
	var sharedVPCService exec.SharedVpcPolicyAndHostedZoneService

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()
		Expect(err).ToNot(HaveOccurred())
		accService, err = exec.NewAccountRoleService(constants.GetAccountRoleDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())

		oidcOpService, err = exec.NewOIDCProviderOperatorRolesService(constants.GetOIDCProviderOperatorRolesDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())

		dnsService, err = exec.NewDnsDomainService()
		Expect(err).ToNot(HaveOccurred())

		vpcService, err = exec.NewVPCService(constants.GetAWSVPCDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())

		sharedVPCService, err = exec.NewSharedVpcPolicyAndHostedZoneService()
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		sharedVPCService.Destroy()
		dnsService.Destroy()
		vpcService.Destroy()
		oidcOpService.Destroy()
		accService.Destroy()
	})

	It("create and destroy account roles for shared vpc - [id:67574]", ci.Day2, ci.Medium, func() {
		if profile.GetClusterType().HCP {
			Skip("Test can run only on Classic cluster")
		}
		By("Create account role without shared vpc role arn")
		accArgs := &exec.AccountRolesArgs{
			AccountRolePrefix: helper.StringPointer("OCP-67574"),
			OpenshiftVersion:  helper.StringPointer(profile.MajorVersion),
			ChannelGroup:      helper.StringPointer(profile.ChannelGroup),
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
			OIDCConfig:         helper.StringPointer(profile.OIDCConfig),
			AWSRegion:          helper.StringPointer(profile.Region),
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
			Name:                      helper.StringPointer(clusterName),
			AWSRegion:                 helper.StringPointer(profile.Region),
			MultiAZ:                   helper.BoolPointer(profile.MultiAZ),
			VPCCIDR:                   helper.StringPointer(constants.DefaultVPCCIDR),
			AWSSharedCredentialsFiles: helper.StringSlicePointer([]string{constants.SharedVpcAWSSharedCredentialsFileENV}),
		}
		_, err = vpcService.Apply(vpcArgs)
		Expect(err).ToNot(HaveOccurred())
		vpcOutput, err := vpcService.Output()
		Expect(err).ToNot(HaveOccurred())
		sharedVPCArgs := &exec.SharedVpcPolicyAndHostedZoneArgs{
			SharedVpcAWSSharedCredentialsFiles: helper.StringSlicePointer([]string{constants.SharedVpcAWSSharedCredentialsFileENV}),
			Region:                             helper.StringPointer(profile.Region),
			ClusterName:                        helper.StringPointer(clusterName),
			DnsDomainId:                        helper.StringPointer(dnsDomainOutput.DnsDomainId),
			IngressOperatorRoleArn:             helper.StringPointer(oidcOpOutput.IngressOperatorRoleArn),
			InstallerRoleArn:                   helper.StringPointer(accRoleOutput.InstallerRoleArn),
			ClusterAWSAccount:                  helper.StringPointer(accRoleOutput.AWSAccountId),
			VpcId:                              helper.StringPointer(vpcOutput.VPCID),
			Subnets:                            helper.StringSlicePointer(vpcOutput.ClusterPrivateSubnets),
		}
		_, err = sharedVPCService.Apply(sharedVPCArgs)
		Expect(err).ToNot(HaveOccurred())
		sharedVPCOutput, err := sharedVPCService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Add shared vpc role arn to account role")
		accArgs = &exec.AccountRolesArgs{
			AccountRolePrefix: helper.StringPointer(accRoleOutput.AccountRolePrefix),
			OpenshiftVersion:  helper.StringPointer(profile.MajorVersion),
			ChannelGroup:      helper.StringPointer(profile.ChannelGroup),
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
