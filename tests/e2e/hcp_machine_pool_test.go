package e2e

import (

	// nolint

	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("TF Test", func() {
	Describe("Create HCP Machine Pool test",
		ci.NonClassicCluster,
		ci.Day2,
		func() {

			var HCPMachinePoolService *exe.MachinePoolService
			var vpcOutput *exe.VPCOutput
			BeforeEach(func() {
				HCPMachinePoolService = exe.NewMachinePoolService(con.HCPMachinePoolDir)

				By("Get vpc output")
				vpcService := exe.NewVPCService()
				vpcOutput, err = vpcService.Output()
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				HCPMachinePoolService.Destroy()

			})
			It("Author:xueli-High-OCP-73068 @OCP-73068 @xueli Create nodepool with security groups via RHCS will work well", ci.High,
				func() {

					By("Prepare additional security groups")
					sgService := exe.NewSecurityGroupService()
					output, err := sgService.Output()
					Expect(err).ToNot(HaveOccurred())
					if output.SGIDs == nil {

						sgArgs := &exe.SecurityGroupArgs{
							AWSRegion: profile.Region,
							VPCID:     vpcOutput.VPCID,
							SGNumber:  4,
						}
						err = sgService.Apply(sgArgs, true)
						Expect(err).ToNot(HaveOccurred())
						// defer sgService.Destroy()
					}

					output, err = sgService.Output()
					Expect(err).ToNot(HaveOccurred())

					replicas := 0
					machineType := "r5.xlarge"
					name := "ocp-73068-2"
					sgIDs := output.SGIDs
					if len(sgIDs) >= 4 {
						sgIDs = sgIDs[0:4]
					}
					MachinePoolArgs := &exe.MachinePoolArgs{
						Cluster:                  clusterID,
						Replicas:                 &replicas,
						MachineType:              &machineType,
						Name:                     &name,
						AdditionalSecurityGroups: &output.SGIDs,
						AutoscalingEnabled:       helper.BoolPointer(false),
						AutoRepair:               helper.BoolPointer(false),
						SubnetID:                 &vpcOutput.ClusterPrivateSubnets[0],
					}
					_, err = HCPMachinePoolService.Apply(MachinePoolArgs, true)
					Expect(err).ToNot(HaveOccurred())
					defer func() {
						_, err = HCPMachinePoolService.Destroy()
						Expect(err).ToNot(HaveOccurred())
					}()

					By("Verify the parameters of the created machinepool")
					mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(mpResponseBody.AWSNodePool().AdditionalSecurityGroupIds())).To(Equal(len(sgIDs)))
					for _, sg := range mpResponseBody.AWSNodePool().AdditionalSecurityGroupIds() {
						Expect(sg).To(BeElementOf(sgIDs))
					}

					By("Update security groups is not allowed to a machinepool")
					testAdditionalSecurityGroups := output.SGIDs[0:1]
					MachinePoolArgs = &exe.MachinePoolArgs{
						AdditionalSecurityGroups: &testAdditionalSecurityGroups,
					}

					applyOutput, err := HCPMachinePoolService.Apply(MachinePoolArgs, false)
					Expect(err).To(HaveOccurred())
					Expect(applyOutput).Should(ContainSubstring("Attribute aws_additional_security_group_ids, cannot be changed"))

					By("Destroy the machinepool")
					_, err = HCPMachinePoolService.Destroy()
					Expect(err).ToNot(HaveOccurred())

					By("Create another machinepool without additional sg ")
					name = "add-73068"
					MachinePoolArgs = &exe.MachinePoolArgs{
						Cluster:     clusterID,
						Replicas:    &replicas,
						MachineType: helper.StringPointer("m5.2xlarge"),
						Name:        &name,
					}

					_, err = HCPMachinePoolService.Apply(MachinePoolArgs, false)
					Expect(err).ToNot(HaveOccurred())

					By("Verify the parameters of the created machinepool")
					mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
					Expect(err).ToNot(HaveOccurred())
					Expect(mpResponseBody.AWSNodePool().AdditionalSecurityGroupIds()).To(BeNil())
					Expect(mpResponseBody.AWSNodePool().InstanceType()).To(Equal("m5.2xlarge"))
				})

			It("Author:xueli-High-OCP-73069 @OCP-73069 @xueli Create nodepool with security groups via RHCS will work well", ci.Low,
				func() {

					By("Prepare additional security groups")
					replicas := 0
					machineType := "r5.xlarge"
					name := "ocp-73069"
					fakeSgIDs := []string{"sg-fake"}

					By("Run terraform apply cannot work with invalid sg IDs")
					MachinePoolArgs := &exe.MachinePoolArgs{
						Cluster:                  clusterID,
						Replicas:                 &replicas,
						MachineType:              &machineType,
						Name:                     &name,
						AdditionalSecurityGroups: &fakeSgIDs,
						AutoscalingEnabled:       helper.BoolPointer(false),
						AutoRepair:               helper.BoolPointer(false),
						SubnetID:                 &vpcOutput.ClusterPrivateSubnets[0],
					}
					output, err := HCPMachinePoolService.Apply(MachinePoolArgs, false)
					Expect(err).To(HaveOccurred())
					Expect(output).Should(ContainSubstring("is not attached to VPC"))

					By("Terraform plan with too many sg IDs cannot work")
					i := 0
					for i < 11 {
						fakeSgIDs = append(fakeSgIDs, fmt.Sprintf("sg-fakeid%d", i))
						i++
					}
					MachinePoolArgs = &exe.MachinePoolArgs{
						Cluster:                  clusterID,
						Replicas:                 &replicas,
						MachineType:              &machineType,
						Name:                     &name,
						AdditionalSecurityGroups: &fakeSgIDs,
						AutoscalingEnabled:       helper.BoolPointer(false),
						AutoRepair:               helper.BoolPointer(false),
						SubnetID:                 &vpcOutput.ClusterPrivateSubnets[0],
					}
					output, err = HCPMachinePoolService.Plan(MachinePoolArgs)
					Expect(err).To(HaveOccurred())
					defer func() {
						_, err = HCPMachinePoolService.Destroy()
						Expect(err).ToNot(HaveOccurred())
					}()

					Expect(output).Should(ContainSubstring("Security number shouldn't be more than 10"))

				})
		})

})
