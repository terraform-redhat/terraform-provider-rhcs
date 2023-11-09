package e2e

import (
	// nolint

	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CMS "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	H "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("TF Test", func() {
	Describe("Verfication/Post day 1 tests", func() {
		var err error
		var profile *CI.Profile

		BeforeEach(func() {
			profile = CI.LoadProfileYamlFileByENV()
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
		})

		Context("Author:smiron-High-OCP-63140 @OCP-63140 @smiron", func() {
			It("Verify fips is enabled/disabled post cluster creation", CI.Day1Post, CI.High, func() {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(getResp.Body().FIPS()).To(Equal(profile.FIPS))
			})
		})
		Context("Author:smiron-High-OCP-63133 @OCP-63133 @smiron", func() {
			It("Verify private_link is enabled/disabled post cluster creation", CI.Day1Post, CI.High, func() {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(getResp.Body().AWS().PrivateLink()).To(Equal(profile.PrivateLink))
			})
		})
		Context("Author:smiron-High-OCP-63143 @OCP-63143 @smiron", func() {
			It("Verify etcd-encryption is enabled/disabled post cluster creation", CI.Day1Post, CI.High, func() {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(getResp.Body().EtcdEncryption()).To(Equal(profile.Etcd))
			})
		})
		Context("Author:smiron-Medium-OCP-64023 @OCP-64023 @smiron", func() {
			It("Verify compute_machine_type value is set post cluster creation", CI.Day1Post, CI.Medium, func() {
				if profile.ComputeMachineType != "" {
					getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred())
					Expect(getResp.Body().Nodes().ComputeMachineType().ID()).To(Equal(profile.ComputeMachineType))
				}
			})
		})
		Context("Author:smiron-Medium-OCP-63141 @OCP-63141 @smiron", func() {
			It("Verify availability zones and multi-az is set post cluster creation", CI.Day1Post, CI.Medium, func() {
				vpcService := EXE.NewVPCService()
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
				zonesArray := strings.Split(profile.Zones, ",")
				clusterAvailZones := getResp.Body().Nodes().AvailabilityZones()
				Expect(err).ToNot(HaveOccurred())
				Expect(getResp.Body().MultiAZ()).To(Equal(profile.MultiAZ))
				if profile.Zones != "" {
					Expect(clusterAvailZones).
						To(Equal(H.JoinStringWithArray(profile.Region, zonesArray)))
				} else {
					vpcOut, _ := vpcService.Output()
					Expect(clusterAvailZones).To(Equal(vpcOut.AZs))
				}
			})
		})
		Context("Author:smiron-High-OCP-68423 @OCP-68423 @smiron", func() {
			It("Verify compute_labels are set post cluster creation", CI.Day1Post, CI.High, func() {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				if profile.Labeling {
					Expect(getResp.Body().Nodes().ComputeLabels()).To(Equal(CON.DefaultMPLabels))
				} else {
					Expect(getResp.Body().Nodes().ComputeLabels()).To(Equal(CON.NilMap))
				}
			})
		})
	})
})
