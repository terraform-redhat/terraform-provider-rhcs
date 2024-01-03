package e2e

import (

	// nolint
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("TF Test", func() {
	Describe("Create MachinePool test cases", func() {
		var mpService *exe.MachinePoolService
		var profile *ci.Profile

		BeforeEach(func() {
			profile = ci.LoadProfileYamlFileByENV()
			mpService = exe.NewMachinePoolService(con.MachinePoolDir)
		})
		AfterEach(func() {
			_, err := mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		})
		Context("Author:amalykhi-High-OCP-64757 @OCP-64757 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-64757 Create a second machine pool", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				By("Create a second machine pool")
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64757"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
				}

				_, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())
				_, err = mpService.Output()
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpOut, err := mpService.Output()
				Expect(err).ToNot(HaveOccurred())
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Replicas()).To(Equal(mpOut.Replicas))
				Expect(mpResponseBody.InstanceType()).To(Equal(mpOut.MachineType))
				Expect(mpResponseBody.ID()).To(Equal(mpOut.Name))
			})
		})
		// Will fail with known issue OCM-5285
		Context("Author:amalykhi-High-OCP-64905 @OCP-64905 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-64905 Edit/delete second machinepool labels", ci.Day2, ci.High, ci.FeatureMachinepool, ci.Exclude, func() {
				By("Create additional machinepool with labels")
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64905"
				creationLabels := map[string]string{"fo1": "bar1", "fo2": "baz2"}
				updatingLabels := map[string]string{"fo1": "bar3", "fo3": "baz3"}
				emptyLabels := map[string]string{}
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
					Labels:      &creationLabels,
				}

				_, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Labels()).To(Equal(creationLabels))

				By("Edit the labels of the machinepool")
				MachinePoolArgs.Labels = &updatingLabels
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Labels()).To(Equal(updatingLabels))

				By("Delete the labels of the machinepool")
				MachinePoolArgs.Labels = &emptyLabels
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Labels()).To(BeNil())

			})
		})
		Context("Author:amalykhi-Critical-OCP-68296 @OCP-68296 @amalykhi", func() {
			It("Author:amalykhi-Critical-OCP-68296 Enable/disable/update autoscaling for additional machinepool", ci.Day2, ci.Critical, ci.FeatureMachinepool, func() {
				By("Create additional machinepool with autoscaling")
				replicas := 9
				minReplicas := 3
				maxReplicas := 6
				machineType := "r5.xlarge"
				name := "ocp-68296"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:            clusterID,
					MinReplicas:        &minReplicas,
					MaxReplicas:        &maxReplicas,
					MachineType:        machineType,
					Name:               name,
					AutoscalingEnabled: true,
				}

				_, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Autoscaling().MinReplicas()).To(Equal(minReplicas))
				Expect(mpResponseBody.Autoscaling().MaxReplicas()).To(Equal(maxReplicas))

				By("Change the number of replicas of the machinepool")
				minReplicas = minReplicas * 2
				maxReplicas = maxReplicas * 2
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Autoscaling().MinReplicas()).To(Equal(minReplicas))
				Expect(mpResponseBody.Autoscaling().MaxReplicas()).To(Equal(maxReplicas))

				By("Disable autoscaling of the machinepool")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
				}

				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Autoscaling()).To(BeNil())

			})
		})
		Context("Author:amalykhi-High-OCP-64904 @ocp-64904 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-64904 Edit second machinepool taints", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				By("Create additional machinepool with labels")
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64904"
				taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": con.NoExecute}
				taint1 := map[string]string{"key": "k2", "value": "val2", "schedule_type": con.NoSchedule}
				taint2 := map[string]string{"key": "k3", "value": "val3", "schedule_type": con.PreferNoSchedule}
				taints := []map[string]string{taint0, taint1}
				emptyTaints := []map[string]string{}
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
					Taints:      taints,
				}

				_, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				respTaints := mpResponseBody.Taints()
				for index, taint := range respTaints {
					Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
					Expect(taint.Key()).To(Equal(taints[index]["key"]))
					Expect(taint.Value()).To(Equal(taints[index]["value"]))
				}
				By("Edit the existing taint of the machinepool")
				taint1["key"] = "k2updated"
				taint1["value"] = "val2updated"

				By("Append new one to the machinepool")
				taints = append(taints, taint2)

				By("Apply the changes to the machinepool")
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				respTaints = mpResponseBody.Taints()
				for index, taint := range respTaints {
					Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
					Expect(taint.Key()).To(Equal(taints[index]["key"]))
					Expect(taint.Value()).To(Equal(taints[index]["value"]))
				}

				By("Delete the taints of the machinepool")
				MachinePoolArgs.Taints = emptyTaints
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Taints()).To(BeNil())

			})
		})
		Context("Author:amalykhi-High-OCP-68283 @OCP-68283 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-68283 Check the validations for the machinepool creation rosa clusters", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				By("Check the validations for the machinepool creation rosa cluster")
				var (
					machinepoolName                                                                                                           = "ocp-68283"
					invalidMachinepoolName                                                                                                    = "%^#@"
					machineType, InvalidInstanceType                                                                                          = "r5.xlarge", "custom-4-16384"
					mpReplicas, minReplicas, maxReplicas, invalidMpReplicas, invalidMinReplicas4Mutilcluster, invalidMaxReplicas4Mutilcluster = 3, 3, 6, -3, 4, 7
				)
				By("Create machinepool with invalid name")
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &invalidMpReplicas,
					Name:        invalidMachinepoolName,
					MachineType: machineType,
				}
				_, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("Expected a valid value for 'name'"))

				By("Create machinepool with invalid replica value")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &invalidMpReplicas,
					Name:        machinepoolName,
					MachineType: machineType,
				}
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("Attribute 'replicas'\nmust be a non-negative integer"))

				By("Create machinepool with invalid instance type")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &mpReplicas,
					Name:        machinepoolName,
					MachineType: InvalidInstanceType,
				}
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("Machine type\n'%s' is not supported for cloud provider", InvalidInstanceType))

				By("Create machinepool with setting replicas and enable-autoscaling at the same time")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:            clusterID,
					Replicas:           &mpReplicas,
					Name:               machinepoolName,
					AutoscalingEnabled: true,
					MachineType:        machineType,
				}
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("when\nenabling autoscaling, should set value for maxReplicas"))

				By("Create machinepool with setting min-replicas large than max-replicas")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:            clusterID,
					MinReplicas:        &maxReplicas,
					MaxReplicas:        &minReplicas,
					Name:               machinepoolName,
					AutoscalingEnabled: true,
					MachineType:        machineType,
				}
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("'min_replicas' must be less than or equal to 'max_replicas'"))

				By("Create machinepool with setting min-replicas and max-replicas but without setting --enable-autoscaling")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:     clusterID,
					MinReplicas: &minReplicas,
					MaxReplicas: &maxReplicas,
					Name:        machinepoolName,
					MachineType: machineType,
				}
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("when\ndisabling autoscaling, cannot set min_replicas and/or max_replicas"))

				By("Create machinepool with setting min-replicas large than max-replicas")

				if profile.MultiAZ {
					By("Create machinepool with setting min-replicas and max-replicas not multiple 3 for multi-az")
					MachinePoolArgs = &exe.MachinePoolArgs{
						Cluster:            clusterID,
						MinReplicas:        &minReplicas,
						MaxReplicas:        &invalidMaxReplicas4Mutilcluster,
						Name:               machinepoolName,
						MachineType:        machineType,
						AutoscalingEnabled: true,
					}
					_, err = mpService.Apply(MachinePoolArgs, false)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("Multi AZ clusters require that the number of replicas be a\nmultiple of 3"))

					MachinePoolArgs = &exe.MachinePoolArgs{
						Cluster:            clusterID,
						MinReplicas:        &invalidMinReplicas4Mutilcluster,
						MaxReplicas:        &maxReplicas,
						Name:               machinepoolName,
						MachineType:        machineType,
						AutoscalingEnabled: true,
					}
					_, err = mpService.Apply(MachinePoolArgs, false)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("Multi AZ clusters require that the number of replicas be a\nmultiple of 3"))

				}

			})
		})
		Context("Author:amalykhi-High-OCP-65063 @OCP-65063 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-65063 Create single-az machinepool for multi-az cluster", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				if !profile.MultiAZ {
					Skip("The test is configured for MultiAZ cluster only")
				}
				getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				azs := getResp.Body().Nodes().AvailabilityZones()
				By("Create additional machinepool with availability zone specified")
				replicas := 1
				machineType := "r5.xlarge"
				name := "ocp-65063"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:          clusterID,
					Replicas:         &replicas,
					MachineType:      machineType,
					Name:             name,
					AvailabilityZone: azs[0],
				}

				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.AvailabilityZones()[0]).To(Equal(azs[0]))

				_, err = mpService.Destroy()
				Expect(err).ToNot(HaveOccurred())

				By("Create additional machinepool with subnet id specified")
				awsSubnetIds := getResp.Body().AWS().SubnetIDs()
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
					SubnetID:    awsSubnetIds[0],
				}

				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Subnets()[0]).To(Equal(awsSubnetIds[0]))

				_, err = mpService.Destroy()
				Expect(err).ToNot(HaveOccurred())

				By("Create additional machinepool with multi_availability_zone=false specified")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:          clusterID,
					Replicas:         &replicas,
					MachineType:      machineType,
					Name:             name,
					MultiAZ:          false,
					AvailabilityZone: azs[1],
				}

				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())
				By("Verify the parameters of the created machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(mpResponseBody.AvailabilityZones())).To(Equal(1))
			})
		})
		Context("Author:amalykhi-High-OCP-65071 @OCP-65071 @amalykhi", func() {
			It("Author:amalykhi-High-OCP-65071 subnet_id option is available for machinepool for BYO VPC single-az cluster", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				if profile.MultiAZ || !profile.BYOVPC {
					Skip("The test is configured for SingleAZ BYO VPC cluster only")
				}
				getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				var zones []string
				vpcOutput, err := ci.PrepareVPC(profile.Region, false, true, zones, getResp.Body().Name())

				By("Tag new subnet to be able to apply it to the machinepool")
				VpcTagService := exe.NewVPCTagService()
				tagKey := fmt.Sprintf("kubernetes.io/cluster/%s", getResp.Body().InfraID())
				tagValue := "shared"
				VPCTagArgs := &exe.VPCTagArgs{
					AWSRegion: getResp.Body().Region().DisplayName(),
					IDs:       vpcOutput.ClusterPrivateSubnets,
					TagKey:    tagKey,
					TagValue:  tagValue,
				}
				err = VpcTagService.Apply(VPCTagArgs, true)
				Expect(err).ToNot(HaveOccurred())
				By("Create additional machinepool with subnet id specified")
				replicas := 1
				machineType := "r5.xlarge"
				name := "ocp-65071"
				newZonePrivateSubnet := vpcOutput.ClusterPrivateSubnets[2]
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
					SubnetID:    newZonePrivateSubnet,
				}
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Subnets()[0]).To(Equal(newZonePrivateSubnet))

				replicas = 4
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the updated machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Replicas()).To(Equal(4))
			})
		})
		Context("Author:xueli-High-OCP-69144 @OCP-69144 @xueli", func() {
			It("Author:xueli-High-OCP-69144 Create machinepool with disk size will work via terraform provider", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				By("Create additional machinepool with disk size specified")
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-69144"
				diskSize := 249
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
					DiskSize:    diskSize,
				}

				_, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					_, err = mpService.Destroy()
					Expect(err).ToNot(HaveOccurred())
				}()

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.RootVolume().AWS().Size()).To(Equal(diskSize))
				Expect(mpResponseBody.InstanceType()).To(Equal(machineType))

				By("Update disksize is not allowed ")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
					DiskSize:    320,
				}

				output, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).To(HaveOccurred())
				Expect(output).Should(ContainSubstring("Attribute disk_size, cannot be changed from 249 to 320"))

				MachinePoolArgs.DiskSize = diskSize
				_, err = mpService.Destroy(MachinePoolArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Create another machinepool without disksize will create another machinepool with default value")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: "m5.2xlarge",
					Name:        name,
				}

				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.RootVolume().AWS().Size()).To(Equal(300))
				Expect(mpResponseBody.InstanceType()).To(Equal("m5.2xlarge"))

			})
		})
		Context("Author:xueli-High-OCP-69146 @OCP-69146 @xueli", func() {
			It("Author:xueli-High-OCP-69146 Create machinepool with additional security group set will work via terraform provider", ci.Day2, ci.High, ci.FeatureMachinepool, func() {
				By("")
				if !profile.BYOVPC {
					Skip("This case only works for BYOVPC cluster profile")
				}
				By("Prepare additional security groups")
				sgService := exe.NewSecurityGroupService()
				output, err := sgService.Output()
				Expect(err).ToNot(HaveOccurred())
				if output.SGIDs == nil {
					vpcService := exe.NewVPCService()
					vpcOutput, err := vpcService.Output()
					Expect(err).ToNot(HaveOccurred())
					sgArgs := &exe.SecurityGroupArgs{
						AWSRegion: profile.Region,
						VPCID:     vpcOutput.VPCID,
						SGNumber:  4,
					}
					err = sgService.Apply(sgArgs, true)
					Expect(err).ToNot(HaveOccurred())
					defer sgService.Destroy()
				}

				output, err = sgService.Output()
				Expect(err).ToNot(HaveOccurred())

				By("Create additional machinepool with security groups specified")
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-69146"

				sgIDs := output.SGIDs
				if len(sgIDs) >= 4 {
					sgIDs = sgIDs[0:4]
				}
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:                  clusterID,
					Replicas:                 &replicas,
					MachineType:              machineType,
					Name:                     name,
					AdditionalSecurityGroups: sgIDs,
				}

				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					_, err = mpService.Destroy()
					Expect(err).ToNot(HaveOccurred())
				}()

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(mpResponseBody.AWS().AdditionalSecurityGroupIds())).To(Equal(len(sgIDs)))
				for _, sg := range mpResponseBody.AWS().AdditionalSecurityGroupIds() {
					Expect(sg).To(BeElementOf(sgIDs))
				}

				By("Update security groups is not allowed to a machinepool")
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:                  clusterID,
					Replicas:                 &replicas,
					MachineType:              machineType,
					Name:                     name,
					AdditionalSecurityGroups: output.SGIDs[0:1],
				}

				applyOutput, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).To(HaveOccurred())
				Expect(applyOutput).Should(ContainSubstring("Attribute aws_additional_security_group_ids, cannot be changed"))

				By("Destroy the machinepool")
				mpService.CreationArgs.AdditionalSecurityGroups = sgIDs
				_, err = mpService.Destroy()
				Expect(err).ToNot(HaveOccurred())

				By("Create another machinepool without additional sg ")
				name = "add-69146"
				MachinePoolArgs = &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: "m5.2xlarge",
					Name:        name,
				}

				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.AWS().AdditionalSecurityGroupIds()).To(BeNil())
				Expect(mpResponseBody.InstanceType()).To(Equal("m5.2xlarge"))

			})
		})
		Context("Author:amalykhi-Medium-OCP-65645 @ocp-65645 @amalykhi", func() {
			It("Author:amalykhi-Medium-OCP-65645 MP reconciliation basic flow", ci.Day2, ci.Medium, ci.FeatureMachinepool, func() {
				By("Create additional machinepool with taints")
				replicas := 3
				machineType := "r5.xlarge"
				mpName := "ocp-65645"
				taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": con.NoExecute}
				taints := []map[string]string{taint0}
				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        mpName,
					Taints:      taints,
				}

				_, err := mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				_, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, mpName)
				Expect(err).ToNot(HaveOccurred())

				By("Delete machinepool by OCM API")
				cms.DeleteMachinePool(ci.RHCSConnection, clusterID, mpName)
				_, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, mpName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("Machine pool with id '%s' not found", mpName))

				By("ReApply the machinepool manifest")
				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, mpName)
				Expect(err).ToNot(HaveOccurred())
				respTaints := mpResponseBody.Taints()
				for index, taint := range respTaints {
					Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
					Expect(taint.Key()).To(Equal(taints[index]["key"]))
					Expect(taint.Value()).To(Equal(taints[index]["value"]))
				}

			})
		})
	})
})

var _ = Describe("TF Test, default machinepool day-2 testing", func() {
	Describe("Default MachinePool test cases", func() {

		var (
			dmpService             *exe.MachinePoolService
			defaultMachinePoolArgs exe.MachinePoolArgs
			mpService              *exe.MachinePoolService
			defaultMachinePoolNmae = "worker"
		)

		BeforeEach(func() {
			dmpService = exe.NewMachinePoolService(con.DefaultMachinePoolDir)
			mpService = exe.NewMachinePoolService(con.MachinePoolDir)

			_, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, defaultMachinePoolNmae)
			if err != nil && strings.Contains(err.Error(), fmt.Sprintf("Machine pool with id '%s' not found", defaultMachinePoolNmae)) {
				Skip("The default machinepool does not exist")
			}

			By("Make sure the default machinepool imported from cluster state")
			imported, _ := h.MakeSureDefaultMachinePoolImported()
			if !imported {
				By("Create default machinepool by importing from cluster state")
				resource, err := h.GetResource(con.ROSAClassic, "rhcs_cluster_rosa_classic", "rosa_sts_cluster")
				Expect(err).NotTo(HaveOccurred())
				Expect(resource).NotTo(BeNil())
				defaultMachinePoolArgs, err = exe.BuildDefaultMachinePoolArgsFromClusterState(resource)
				defaultMachinePoolArgs.Cluster = clusterID

				Expect(err).NotTo(HaveOccurred())
				_, err = dmpService.Apply(&defaultMachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())
			}
		})
		AfterEach(func() {
			// Check if current test is skipped, skip this AfterEach block too
			if CurrentSpecReport().Failure.Message == "The default machinepool does not exist" {
				return
			}

			By("Recover the default machinepool to the original state")
			_, err := dmpService.Apply(&defaultMachinePoolArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Destroy additonal mp")
			_, err = mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("Author:yuwan-Critical-OCP-69073 @OCP-69073 @yuwan", func() {
			It("Author:yuwan-High-OCP-69073 Check the validations and some negative scenarios of creating/editing/deleting default machinepool via terraform", ci.Day2, ci.Medium, ci.FeatureMachinepool, func() {
				defaultMachinePoolResource, err := h.GetResource(con.ROSAClassic, "rhcs_machine_pool", "dmp")
				Expect(err).NotTo(HaveOccurred())
				defaultMachinepoolArgFromMachinepool, err := exe.BuildDefaultMachinePoolArgsFromDefaultMachinePoolState(defaultMachinePoolResource)
				Expect(err).NotTo(HaveOccurred())
				defaultMachinepoolArgFromMachinepool.Cluster = clusterID

				dmpArgFromMachinepoolForTesting := defaultMachinepoolArgFromMachinepool

				By("Create machinepool with the default machinepool name 'worker' when it does exist")
				output, err := dmpService.Apply(&defaultMachinepoolArgFromMachinepool, false)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(ContainSubstring("No changes. Your infrastructure matches the configuration."))

				if h.DigBool(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "autoscaling_enabled") {
					By("Edit the deafult machinepool max and min replicas to 0")
					zeroReplicas := 0
					dmpArgFromMachinepoolForTesting.MaxReplicas = &zeroReplicas
					dmpArgFromMachinepoolForTesting.MinReplicas = &zeroReplicas
					Expect(err).To(HaveOccurred())
					Expect(output).To(ContainSubstring("Failed to update machine pool"))
					Expect(output).To(ContainSubstring("must be a integer greater than 0"))
				} else {
					By("Edit the deafult machinepool replicas to 0")
					zeroReplicas := 0
					dmpArgFromMachinepoolForTesting = defaultMachinepoolArgFromMachinepool
					dmpArgFromMachinepoolForTesting.Replicas = &zeroReplicas
					_, err = dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Failed to update machine pool"))
					Expect(err.Error()).To(ContainSubstring("least one machine pool able to run OCP workload is required"))
				}

				By("Check the machine type change will triger re-creation")
				dmpArgFromMachinepoolForTesting = defaultMachinepoolArgFromMachinepool
				dmpArgFromMachinepoolForTesting.MachineType = "r5.xlarge"
				out, err := dmpService.Plan(&dmpArgFromMachinepoolForTesting)
				Expect(err).ToNot(HaveOccurred())
				Expect(out).To(ContainSubstring("rhcs_machine_pool.mp must be replaced"))
				Expect(out).To(ContainSubstring("destroy and then create replacement"))

				By("Delete dmp without additional mp exists")
				resp, err := cms.ListMachinePool(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Total()).To(Equal(1), "multiple machinepools found")

				// Only check this when confirm no other machinepool existing
				output, err = dmpService.Destroy()
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(ContainSubstring("Warning: Cannot delete machine pool"))

			})
		})

		Context("Author:yuwan-Critical-OCP-69009 @OCP-69009 @yuwan", func() {
			It("Author:yuwan-High-OCP-69009 Check the default machinepool creation with the cluster and edit/delete it via terraform",
				ci.Day2, ci.Critical, ci.FeatureMachinepool, ci.Exclude,
				func() {
					taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": con.NoExecute}
					taint1 := map[string]string{"key": "k2", "value": "val2", "schedule_type": con.NoSchedule}
					taints := []map[string]string{taint0, taint1}

					defaultMPName := "worker"

					defaultMachinePoolResource, err := h.GetResource(con.ROSAClassic, "rhcs_machine_pool", "dmp")
					Expect(err).NotTo(HaveOccurred())
					defaultMachinepoolArgFromMachinepool, err := exe.BuildDefaultMachinePoolArgsFromDefaultMachinePoolState(defaultMachinePoolResource)
					Expect(err).NotTo(HaveOccurred())
					defaultMachinepoolArgFromMachinepool.Cluster = clusterID

					dmpArgFromMachinepoolForTesting := defaultMachinepoolArgFromMachinepool

					By("Edit the taints without additional machinepool")

					dmpArgFromMachinepoolForTesting.Taints = taints
					_, err = dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Failed to update machine pool"))
					Expect(err.Error()).To(ContainSubstring("least one machine pool able to run OCP workload is required. Pool should not"))
					dmpArgFromMachinepoolForTesting = defaultMachinepoolArgFromMachinepool

					if h.DigBool(h.DigArray(defaultMachinePoolResource, "instances")[0], "attributes", "autoscaling_enabled") {
						By("Edit default machinepool with autoscale configuration")
						minReplicas := 3
						maxReplicas := 6
						dmpArgFromMachinepoolForTesting.MinReplicas = &minReplicas
						dmpArgFromMachinepoolForTesting.MaxReplicas = &maxReplicas
						_, err := dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
						Expect(err).ToNot(HaveOccurred())

						By("Verify the parameters of the created machinepool")
						mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, defaultMPName)
						Expect(err).ToNot(HaveOccurred())
						Expect(mpResponseBody.Autoscaling().MinReplicas()).To(Equal(minReplicas))
						Expect(mpResponseBody.Autoscaling().MaxReplicas()).To(Equal(maxReplicas))
					} else {
						By("Edit default machinepool with replicas")
						replicas := 6
						dmpArgFromMachinepoolForTesting.Replicas = &replicas
						_, err := dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
						Expect(err).ToNot(HaveOccurred())

						By("Verify the parameters of the created machinepool")
						mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, defaultMPName)
						Expect(err).ToNot(HaveOccurred())
						Expect(mpResponseBody.Replicas()).To(Equal(replicas))
					}

					By("Edit default machinepool with labels")
					dmpArgFromMachinepoolForTesting = defaultMachinepoolArgFromMachinepool
					creationLabels := map[string]string{"fo1": "bar1", "fo2": "baz2"}
					dmpArgFromMachinepoolForTesting.Labels = &creationLabels
					_, err = dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
					Expect(err).ToNot(HaveOccurred())

					By("Verify the parameters of the created machinepool")
					mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, defaultMPName)
					Expect(err).ToNot(HaveOccurred())
					Expect(mpResponseBody.Labels()).To(Equal(creationLabels))

					By("Create an additional machinepool")
					replicas := 3
					machineType := "r5.xlarge"
					name := "amp-69009"

					MachinePoolArgs := &exe.MachinePoolArgs{
						Cluster:     clusterID,
						Replicas:    &replicas,
						MachineType: machineType,
						Name:        name,
						Taints:      taints,
					}

					_, err = mpService.Apply(MachinePoolArgs, false)
					Expect(err).ToNot(HaveOccurred())

					By("Edit the default machinepool with taints")
					dmpArgFromMachinepoolForTesting = defaultMachinepoolArgFromMachinepool
					_, err = dmpService.Apply(&defaultMachinepoolArgFromMachinepool, false)
					Expect(err).ToNot(HaveOccurred())

					By("Verify the parameters of the created machinepool")
					mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
					Expect(err).ToNot(HaveOccurred())
					respTaints := mpResponseBody.Taints()
					for index, taint := range respTaints {
						Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
						Expect(taint.Key()).To(Equal(taints[index]["key"]))
						Expect(taint.Value()).To(Equal(taints[index]["value"]))
					}
				})
		})

	})
})

var _ = Describe("TF Test, day-3 default machinepool testing", func() {
	Describe("Default MachinePool day-3 test cases", func() {

		var (
			dmpService             *exe.MachinePoolService
			defaultMachinePoolArgs exe.MachinePoolArgs
			mpService              *exe.MachinePoolService
		)

		BeforeEach(func() {
			dmpService = exe.NewMachinePoolService(con.DefaultMachinePoolDir)
			mpService = exe.NewMachinePoolService(con.MachinePoolDir)

			By("Make sure the default machinepool imported from cluster state")
			imported, _ := h.MakeSureDefaultMachinePoolImported()
			if !imported {
				By("Create default machinepool by importing from cluster state")
				resource, err := h.GetResource(con.ROSAClassic, "rhcs_cluster_rosa_classic", "rosa_sts_cluster")
				Expect(err).NotTo(HaveOccurred())
				Expect(resource).NotTo(BeNil())
				defaultMachinePoolArgs, err = exe.BuildDefaultMachinePoolArgsFromClusterState(resource)
				defaultMachinePoolArgs.Cluster = clusterID

				Expect(err).NotTo(HaveOccurred())
				_, err = dmpService.Apply(&defaultMachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())
			}
		})

		Context("Author:yuwan-Critical-OCP-69727 @OCP-69727 @yuwan", func() {
			It("Author:yuwan-High-OCP-69727 Check the default machinepool edit/delete operations with additional mp exists it via terraform	", ci.Day3, ci.Critical, ci.FeatureMachinepool, func() {
				defaultMachinePoolResource, err := h.GetResource(con.ROSAClassic, "rhcs_machine_pool", "dmp")
				Expect(err).NotTo(HaveOccurred())
				defaultMachinepoolArgFromMachinepool, err := exe.BuildDefaultMachinePoolArgsFromDefaultMachinePoolState(defaultMachinePoolResource)
				Expect(err).NotTo(HaveOccurred())
				defaultMachinepoolArgFromMachinepool.Cluster = clusterID

				By("Destroy default machinepool without additional machinepool existing")
				resp, err := cms.ListMachinePool(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				num, _ := resp.GetSize()
				Expect(num).To(Equal(1))

				output, err := dmpService.Destroy()
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(ContainSubstring("Warning: Cannot delete machine pool"))
				Expect(output).To(ContainSubstring("must have at least"))

				By("Create an additional machinepool")
				replicas := 3
				machineType := "r5.xlarge"
				name := "amp-69727"

				MachinePoolArgs := &exe.MachinePoolArgs{
					Cluster:     clusterID,
					Replicas:    &replicas,
					MachineType: machineType,
					Name:        name,
				}

				_, err = mpService.Apply(MachinePoolArgs, false)
				Expect(err).ToNot(HaveOccurred())

				By("Import the default machinepool state")
				_, err = dmpService.Apply(&defaultMachinepoolArgFromMachinepool, false)
				Expect(err).ToNot(HaveOccurred())

				By("Destroy default machinepool")
				output, err = dmpService.Destroy()
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(ContainSubstring("Destroy complete! Resources: 1 destroyed."))
			})
		})
	})
})
