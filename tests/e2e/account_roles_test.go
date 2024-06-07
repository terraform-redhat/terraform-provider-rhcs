package e2e

import (
	// nolint

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"strings"

	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("Edit Account roles", func() {
	defer GinkgoRecover()

	var err error
	var profile *CI.Profile
	var accService *EXE.AccountRoleService

	BeforeEach(func() {
		profile = CI.LoadProfileYamlFileByENV()
		Expect(err).ToNot(HaveOccurred())

		accService, err = EXE.NewAccountRoleService(CON.GetAddAccountRoleDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		accService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})

	It("can create account roles with default prefix - [id:65380]", CI.Day2, CI.Medium, func() {
		args := &EXE.AccountRolesArgs{
			AccountRolePrefix: "",
			OpenshiftVersion:  profile.MajorVersion,
			ChannelGroup:      profile.ChannelGroup,
		}
		accRoleOutput, err := accService.Apply(args, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(accRoleOutput.AccountRolePrefix).Should(ContainSubstring(CON.DefaultAccountRolesPrefix))

	})

	It("can delete account roles via account-role module - [id:63316]", CI.Day2, CI.Critical, func() {
		args := &EXE.AccountRolesArgs{
			AccountRolePrefix: "OCP-63316",
			OpenshiftVersion:  profile.MajorVersion,
			ChannelGroup:      profile.ChannelGroup,
		}
		_, err := accService.Apply(args, true)
		Expect(err).ToNot(HaveOccurred())
		accService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})
})

var _ = Describe("Create Account roles with shared vpc role", func() {
	defer GinkgoRecover()

	var err error
	var profile *CI.Profile
	var accService *EXE.AccountRoleService
	var oidcOpService *EXE.OIDCProviderOperatorRolesService
	var dnsService *EXE.DnsService
	var vpcService *EXE.VPCService
	var sharedVPCService *EXE.SharedVpcPolicyAndHostedZoneService

	BeforeEach(func() {
		profile = CI.LoadProfileYamlFileByENV()
		Expect(err).ToNot(HaveOccurred())
		accService, err = EXE.NewAccountRoleService(CON.GetAccountRoleDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())

		oidcOpService, err = EXE.NewOIDCProviderOperatorRolesService(CON.GetOIDCProviderOperatorRolesDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())

		dnsService, err = EXE.NewDnsDomainService()
		Expect(err).ToNot(HaveOccurred())

		vpcService = EXE.NewVPCService(CON.GetAWSVPCDefaultManifestDir(profile.GetClusterType()))

		sharedVPCService, err = EXE.NewSharedVpcPolicyAndHostedZoneService()
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		sharedVPCService.Destroy()
		dnsService.Destroy()
		vpcService.Destroy()
		oidcOpService.Destroy()
		accService.Destroy()
	})

	It("create and destroy account roles for shared vpc - [id:67574]", CI.Day2, CI.Medium, CI.NonHCPCluster, func() {
		By("Create account role without shared vpc role arn")
		accArgs := &EXE.AccountRolesArgs{
			AccountRolePrefix: "OCP-67574",
			OpenshiftVersion:  profile.MajorVersion,
			ChannelGroup:      profile.ChannelGroup,
		}
		accRoleOutput, err := accService.Apply(accArgs, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(accRoleOutput.AccountRolePrefix).Should(Equal(accArgs.AccountRolePrefix))

		By("Create operator role")
		oidcOpArgs := &EXE.OIDCProviderOperatorRolesArgs{
			AccountRolePrefix:  accRoleOutput.AccountRolePrefix,
			OperatorRolePrefix: accRoleOutput.AccountRolePrefix,
			OIDCConfig:         profile.OIDCConfig,
			AWSRegion:          profile.Region,
		}
		oidcOpOutput, err := oidcOpService.Apply(oidcOpArgs, true)
		Expect(err).ToNot(HaveOccurred())
		Expect(oidcOpOutput.OperatorRolePrefix).Should(Equal(oidcOpArgs.OperatorRolePrefix))

		By("verify policy permission")
		roleName := strings.Split(oidcOpOutput.IngressOperatorRoleArn, "/")[1]
		policies, err := helper.GetRoleAttachedPolicies(roleName)
		Expect(policies[0].Statement[0].Action).ToNot(ContainElement("sts:AssumeRole"))

		By("prepare route53")
		dnsDomainArgs := &EXE.DnsDomainArgs{}
		err = dnsService.Create(dnsDomainArgs)
		Expect(err).ToNot(HaveOccurred())
		dnsDomainOutput, err := dnsService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Create shared vpc role")
		clusterName := "OCP-67574"
		vpcArgs := &EXE.VPCArgs{
			Name:                      clusterName,
			AWSRegion:                 profile.Region,
			MultiAZ:                   profile.MultiAZ,
			VPCCIDR:                   CON.DefaultVPCCIDR,
			AWSSharedCredentialsFiles: []string{CON.SharedVpcAWSSharedCredentialsFileENV},
		}
		err = vpcService.Apply(vpcArgs, true)
		Expect(err).ToNot(HaveOccurred())
		vpcOutput, err := vpcService.Output()
		Expect(err).ToNot(HaveOccurred())
		sharedVPCArgs := &EXE.SharedVpcPolicyAndHostedZoneArgs{
			SharedVpcAWSSharedCredentialsFiles: []string{CON.SharedVpcAWSSharedCredentialsFileENV},
			Region:                             profile.Region,
			ClusterName:                        clusterName,
			DnsDomainId:                        dnsDomainOutput.DnsDomainId,
			IngressOperatorRoleArn:             oidcOpOutput.IngressOperatorRoleArn,
			InstallerRoleArn:                   accRoleOutput.InstallerRoleArn,
			ClusterAWSAccount:                  accRoleOutput.AWSAccountId,
			VpcId:                              vpcOutput.VPCID,
			Subnets:                            vpcOutput.ClusterPrivateSubnets,
		}
		err = sharedVPCService.Apply(sharedVPCArgs, true)
		Expect(err).ToNot(HaveOccurred())
		sharedVPCOutput, err := sharedVPCService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Add shared vpc role arn to account role")
		accArgs = &EXE.AccountRolesArgs{
			AccountRolePrefix: accRoleOutput.AccountRolePrefix,
			OpenshiftVersion:  profile.MajorVersion,
			ChannelGroup:      profile.ChannelGroup,
			SharedVpcRoleArn:  sharedVPCOutput.SharedRole,
		}
		accRoleOutput, err = accService.Apply(accArgs, true)
		Expect(err).ToNot(HaveOccurred())

		By("verify policy permission")
		policies, err = helper.GetRoleAttachedPolicies(roleName)
		Expect(policies[0].Statement[2].Action).To(ContainSubstring("sts:AssumeRole"))
	})
})
