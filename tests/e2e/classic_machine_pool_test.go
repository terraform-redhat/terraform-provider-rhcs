package e2e

import (

	// nolint
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

var _ = Describe("Create MachinePool", ci.Day2, ci.NonHCPCluster, ci.FeatureMachinepool, func() {
	defer GinkgoRecover()

	var mpService *exe.MachinePoolService
	var profile *ci.Profile

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()
		mpService = exe.NewMachinePoolService(con.ClassicMachinePoolDir)
	})

	AfterEach(func() {
		_, err := mpService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})

	It("can create a second machine pool - [id:64757]", ci.High, func() {
		By("Create a second machine pool")
		replicas := 3
		machineType := "r5.xlarge"
		name := "ocp-64757"
		MachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
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

	// Will fail with known issue OCM-5285
	It("can edit/delete second machinepool labels - [id:64905]", ci.High, ci.Exclude, func() {
		By("Create additional machinepool with labels")
		replicas := 3
		machineType := "r5.xlarge"
		name := "ocp-64905"
		creationLabels := map[string]string{"fo1": "bar1", "fo2": "baz2"}
		updatingLabels := map[string]string{"fo1": "bar3", "fo3": "baz3"}
		emptyLabels := map[string]string{}
		MachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
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

	It("can enable/disable/update autoscaling for additional machinepool - [id:68296]", ci.Critical, func() {
		By("Create additional machinepool with autoscaling")
		replicas := 9
		minReplicas := 3
		maxReplicas := 6
		machineType := "r5.xlarge"
		name := "ocp-68296"
		MachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:            &clusterID,
			MinReplicas:        &minReplicas,
			MaxReplicas:        &maxReplicas,
			MachineType:        &machineType,
			Name:               &name,
			AutoscalingEnabled: h.BoolPointer(true),
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
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
		}

		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the updated machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Autoscaling()).To(BeNil())
	})

	It("edit second machinepool taints - [id:64904]", ci.High, func() {
		By("Create additional machinepool with labels")
		replicas := 3
		machineType := "r5.xlarge"
		name := "ocp-64904"
		taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": con.NoExecute}
		taint1 := map[string]string{"key": "k2", "value": "val2", "schedule_type": con.NoSchedule}
		taint2 := map[string]string{"key": "k3", "value": "val3", "schedule_type": con.PreferNoSchedule}
		taints := []map[string]string{taint0, taint1}
		MachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
			Taints:      &taints,
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
		MachinePoolArgs.Taints = nil
		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the updated machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Taints()).To(BeNil())
	})

	It("can validate the machinepool creation - [id:68283]", ci.High, func() {
		By("Check the validations for the machinepool creation rosa cluster")
		var (
			machinepoolName                                                                                                           = "ocp-68283"
			invalidMachinepoolName                                                                                                    = "%^#@"
			machineType, InvalidInstanceType                                                                                          = "r5.xlarge", "custom-4-16384"
			mpReplicas, minReplicas, maxReplicas, invalidMpReplicas, invalidMinReplicas4Mutilcluster, invalidMaxReplicas4Mutilcluster = 3, 3, 6, -3, 4, 7
		)
		By("Create machinepool with invalid name")
		MachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &invalidMpReplicas,
			Name:        &invalidMachinepoolName,
			MachineType: &machineType,
		}
		_, err := mpService.Apply(MachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("Expected a valid value for 'name'"))

		By("Create machinepool with invalid replica value")
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &invalidMpReplicas,
			Name:        &machinepoolName,
			MachineType: &machineType,
		}
		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("Attribute 'replicas'\nmust be a non-negative integer"))

		By("Create machinepool with invalid instance type")
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &mpReplicas,
			Name:        &machinepoolName,
			MachineType: &InvalidInstanceType,
		}
		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("Machine type\n'%s' is not supported for cloud provider", InvalidInstanceType))

		By("Create machinepool with setting replicas and enable-autoscaling at the same time")
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:            &clusterID,
			Replicas:           &mpReplicas,
			Name:               &machinepoolName,
			AutoscalingEnabled: h.BoolPointer(true),
			MachineType:        &machineType,
		}
		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("when\nenabling autoscaling, should set value for maxReplicas"))

		By("Create machinepool with setting min-replicas large than max-replicas")
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:            &clusterID,
			MinReplicas:        &maxReplicas,
			MaxReplicas:        &minReplicas,
			Name:               &machinepoolName,
			AutoscalingEnabled: h.BoolPointer(true),
			MachineType:        &machineType,
		}
		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("'min_replicas' must be less than or equal to 'max_replicas'"))

		By("Create machinepool with setting min-replicas and max-replicas but without setting --enable-autoscaling")
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			MinReplicas: &minReplicas,
			MaxReplicas: &maxReplicas,
			Name:        &machinepoolName,
			MachineType: &machineType,
		}
		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("when\ndisabling autoscaling, cannot set min_replicas and/or max_replicas"))

		By("Create machinepool with setting min-replicas large than max-replicas")

		if profile.MultiAZ {
			By("Create machinepool with setting min-replicas and max-replicas not multiple 3 for multi-az")
			MachinePoolArgs = &exe.MachinePoolArgs{
				Cluster:            &clusterID,
				MinReplicas:        &minReplicas,
				MaxReplicas:        &invalidMaxReplicas4Mutilcluster,
				Name:               &machinepoolName,
				MachineType:        &machineType,
				AutoscalingEnabled: h.BoolPointer(true),
			}
			_, err = mpService.Apply(MachinePoolArgs, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("Multi AZ clusters require that the number of replicas be a\nmultiple of 3"))

			MachinePoolArgs = &exe.MachinePoolArgs{
				Cluster:            &clusterID,
				MinReplicas:        &invalidMinReplicas4Mutilcluster,
				MaxReplicas:        &maxReplicas,
				Name:               &machinepoolName,
				MachineType:        &machineType,
				AutoscalingEnabled: h.BoolPointer(true),
			}
			_, err = mpService.Apply(MachinePoolArgs, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("Multi AZ clusters require that the number of replicas be a\nmultiple of 3"))

		}
	})

	It("can create single-az machinepool for multi-az cluster - [id:65063]", ci.High, func() {
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
			Cluster:          &clusterID,
			Replicas:         &replicas,
			MachineType:      &machineType,
			Name:             &name,
			AvailabilityZone: &azs[0],
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
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
			SubnetID:    &awsSubnetIds[0],
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
			Cluster:          &clusterID,
			Replicas:         &replicas,
			MachineType:      &machineType,
			Name:             &name,
			MultiAZ:          h.BoolPointer(false),
			AvailabilityZone: &azs[1],
		}

		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).ToNot(HaveOccurred())
		By("Verify the parameters of the created machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mpResponseBody.AvailabilityZones())).To(Equal(1))
	})

	It("can create machinepool with subnet_id option for BYO VPC single-az cluster - [id:65071]", ci.High, func() {
		if profile.MultiAZ || !profile.BYOVPC {
			Skip("The test is configured for SingleAZ BYO VPC cluster only")
		}
		getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
		var zones []string
		vpcOutput, err := ci.PrepareVPC(profile.Region, true, zones, profile.GetClusterType(), getResp.Body().Name(), "")

		By("Tag new subnet to be able to apply it to the machinepool")
		VpcTagService := exe.NewVPCTagService()
		tagKey := fmt.Sprintf("kubernetes.io/cluster/%s", getResp.Body().InfraID())
		tagValue := "shared"
		VPCTagArgs := &exe.VPCTagArgs{
			AWSRegion: getResp.Body().Region().ID(),
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
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
			SubnetID:    &newZonePrivateSubnet,
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

	It("can create machinepool with disk size - [id:69144]", ci.High, func() {
		By("Create additional machinepool with disk size specified")
		replicas := 3
		machineType := "r5.xlarge"
		name := "ocp-69144"
		diskSize := 249
		MachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
			DiskSize:    &diskSize,
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
		Expect(mpResponseBody.RootVolume().AWS().Size()).To(Equal(diskSize))
		Expect(mpResponseBody.InstanceType()).To(Equal(machineType))

		By("Update disksize is not allowed ")
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
			DiskSize:    h.IntPointer(320),
		}

		output, err := mpService.Apply(MachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(output).Should(ContainSubstring("Attribute disk_size, cannot be changed from 249 to 320"))

		MachinePoolArgs.DiskSize = &diskSize
		_, err = mpService.Destroy(MachinePoolArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Create another machinepool without disksize will create another machinepool with default value")
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: h.StringPointer("m5.2xlarge"),
			Name:        &name,
		}

		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.RootVolume().AWS().Size()).To(Equal(300))
		Expect(mpResponseBody.InstanceType()).To(Equal("m5.2xlarge"))
	})

	It("can create machinepool with additional security group - [id:69146]", ci.High, func() {
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
			Cluster:                  &clusterID,
			Replicas:                 &replicas,
			MachineType:              &machineType,
			Name:                     &name,
			AdditionalSecurityGroups: &sgIDs,
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
		testAdditionalSecurityGroups := output.SGIDs[0:1]
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:                  &clusterID,
			Replicas:                 &replicas,
			MachineType:              &machineType,
			Name:                     &name,
			AdditionalSecurityGroups: &testAdditionalSecurityGroups,
		}

		applyOutput, err := mpService.Apply(MachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(applyOutput).Should(ContainSubstring("Attribute aws_additional_security_group_ids, cannot be changed"))

		By("Destroy the machinepool")
		mpService.CreationArgs.AdditionalSecurityGroups = &sgIDs
		_, err = mpService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Create another machinepool without additional sg ")
		name = "add-69146"
		MachinePoolArgs = &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: h.StringPointer("m5.2xlarge"),
			Name:        &name,
		}

		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.AWS().AdditionalSecurityGroupIds()).To(BeNil())
		Expect(mpResponseBody.InstanceType()).To(Equal("m5.2xlarge"))
	})

	It("can reconciliate with basic flow - [id:65645]", ci.Medium, func() {
		By("Create additional machinepool with taints")
		replicas := 3
		machineType := "r5.xlarge"
		mpName := "ocp-65645"
		taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": con.NoExecute}
		taints := []map[string]string{taint0}
		MachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &mpName,
			Taints:      &taints,
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

var _ = Describe("Import MachinePool", ci.Day2, ci.NonHCPCluster, ci.FeatureImport, func() {
	var mpService *exe.MachinePoolService
	var importService exe.ImportService

	BeforeEach(func() {
		mpService = exe.NewMachinePoolService(con.ClassicMachinePoolDir)
		importService = *exe.NewImportService(con.ImportResourceDir) // init new import service
	})
	AfterEach(func() {
		By("Destroy import service")
		_, importErr := importService.Destroy()
		Expect(importErr).ToNot(HaveOccurred())
	})

	It("can import resource - [id:66403]", ci.Medium, func() {
		By("Create additional machinepool for import")
		minReplicas := 3
		maxReplicas := 6
		creationLabels := map[string]string{"foo1": "bar1"}
		machineType := "r5.xlarge"
		name := "ocp-66403"
		MachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:            &clusterID,
			MinReplicas:        &minReplicas,
			MaxReplicas:        &maxReplicas,
			MachineType:        &machineType,
			Name:               &name,
			Labels:             &creationLabels,
			AutoscalingEnabled: h.BoolPointer(true),
		}

		_, err := mpService.Apply(MachinePoolArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Run the command to import the machinepool")
		importParam := &exe.ImportArgs{
			ClusterID:    clusterID,
			ResourceKind: "rhcs_machine_pool",
			ResourceName: "mp_import",
			ObjectName:   name,
		}
		Expect(importService.Import(importParam)).To(Succeed())

		By("Check resource state - import command succeeded")

		output, err := importService.ShowState(importParam)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring(machineType))
		Expect(output).To(ContainSubstring(name))
		Expect(output).To(MatchRegexp("foo1"))
		Expect(output).To(MatchRegexp("bar1"))
	})
})

var _ = Describe("Edit MachinePool", ci.Day2, ci.NonHCPCluster, ci.FeatureMachinepool, func() {

	var (
		dmpService                     *exe.MachinePoolService
		mpService                      *exe.MachinePoolService
		defaultMachinePoolNmae         = "worker"
		defaultMachinepoolResponse     *cmv1.MachinePool
		originalDefaultMachinepoolArgs exe.MachinePoolArgs
	)

	BeforeEach(func() {
		dmpService = exe.NewMachinePoolService(con.DefaultMachinePoolDir)
		mpService = exe.NewMachinePoolService(con.ClassicMachinePoolDir)

		defaultMachinepoolResponse, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, defaultMachinePoolNmae)
		if err != nil && strings.Contains(err.Error(), fmt.Sprintf("Machine pool with id '%s' not found", defaultMachinePoolNmae)) {
			Skip("The default machinepool does not exist")
		}
		originalDefaultMachinepoolArgs = exe.BuildMachinePoolArgsFromCSResponse(defaultMachinepoolResponse)
		originalDefaultMachinepoolArgs.Cluster = &clusterID

		By("Make sure the default machinepool imported from cluster state")
		imported, _ := h.CheckDefaultMachinePoolImported()
		if !imported {
			By("Create default machinepool by importing from CMS ")
			_, err = dmpService.Apply(&originalDefaultMachinepoolArgs, false)
			Expect(err).ToNot(HaveOccurred())
		}
	})
	AfterEach(func() {
		// Check if current test is skipped, skip this AfterEach block too
		if CurrentSpecReport().Failure.Message == "The default machinepool does not exist" {
			return
		}

		By("Recover the default machinepool to the original state")
		_, err := dmpService.Apply(&originalDefaultMachinepoolArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Destroy additonal mp")
		if mpService.CreationArgs != nil {
			_, err = mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("check the validations and some negative scenarios of creating/editing/deleting default machinepool - [id:69073]", ci.Medium, func() {
		dmpArgFromMachinepoolForTesting := originalDefaultMachinepoolArgs
		By("Create machinepool with the default machinepool name 'worker' when it does exist")
		output, err := dmpService.Apply(&originalDefaultMachinepoolArgs, false)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring("No changes. Your infrastructure matches the configuration."))
		if _, ok := defaultMachinepoolResponse.GetAutoscaling(); ok {
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
			dmpArgFromMachinepoolForTesting = originalDefaultMachinepoolArgs
			dmpArgFromMachinepoolForTesting.Replicas = &zeroReplicas
			_, err = dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to update machine pool"))
			Expect(err.Error()).To(ContainSubstring("least one machine pool able to run OCP workload is required"))
		}

		By("Check the machine type change will triger re-creation")
		dmpArgFromMachinepoolForTesting = originalDefaultMachinepoolArgs
		dmpArgFromMachinepoolForTesting.MachineType = h.StringPointer("r5.xlarge")
		out, err := dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("machine_type, cannot be changed"))

		By("Delete dmp without additional mp exists")
		resp, err := cms.ListMachinePool(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.Total()).To(Equal(1), "multiple machinepools found")

		// Only check this when confirm no other machinepool existing
		output, err = dmpService.Destroy()
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring("Warning: Cannot delete machine pool"))
	})

	It("check the default machinepool creation with the cluster and edit/delete it - [id:69009]",
		ci.Critical, ci.Exclude,
		func() {
			taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": con.NoExecute}
			taint1 := map[string]string{"key": "k2", "value": "val2", "schedule_type": con.NoSchedule}
			taints := []map[string]string{taint0, taint1}
			defaultMPName := "worker"
			dmpArgFromMachinepoolForTesting := originalDefaultMachinepoolArgs

			By("Edit the taints without additional machinepool")
			dmpArgFromMachinepoolForTesting.Taints = &taints
			_, err = dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to update machine pool"))
			Expect(err.Error()).To(ContainSubstring("least one machine pool able to run OCP workload is required. Pool should not"))
			dmpArgFromMachinepoolForTesting = originalDefaultMachinepoolArgs
			if _, ok := defaultMachinepoolResponse.GetAutoscaling(); ok {
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
			dmpArgFromMachinepoolForTesting = originalDefaultMachinepoolArgs
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
				Cluster:     &clusterID,
				Replicas:    &replicas,
				MachineType: &machineType,
				Name:        &name,
				// Taints:      &taints,
			}

			_, err = mpService.Apply(MachinePoolArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Edit the default machinepool with taints")
			dmpArgFromMachinepoolForTesting = originalDefaultMachinepoolArgs
			dmpArgFromMachinepoolForTesting.Taints = &taints
			_, err = dmpService.Apply(&dmpArgFromMachinepoolForTesting, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify the parameters of the default machinepool")
			mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, "worker")
			Expect(err).ToNot(HaveOccurred())
			respTaints := mpResponseBody.Taints()
			for index, taint := range respTaints {
				Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
				Expect(taint.Key()).To(Equal(taints[index]["key"]))
				Expect(taint.Value()).To(Equal(taints[index]["value"]))
			}
		})

})

var _ = Describe("Destroy MachinePool", ci.Day3, ci.NonHCPCluster, ci.FeatureMachinepool, func() {
	var (
		dmpService                 *exe.MachinePoolService
		defaultMachinePoolArgs     exe.MachinePoolArgs
		mpService                  *exe.MachinePoolService
		defaultMachinepoolResponse *cmv1.MachinePool
		defaultMachinePoolNmae     = "worker"
	)

	BeforeEach(func() {
		dmpService = exe.NewMachinePoolService(con.DefaultMachinePoolDir)
		mpService = exe.NewMachinePoolService(con.ClassicMachinePoolDir)

		defaultMachinepoolResponse, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, defaultMachinePoolNmae)
		Expect(err).ToNot(HaveOccurred())
		defaultMachinePoolArgs = exe.BuildMachinePoolArgsFromCSResponse(defaultMachinepoolResponse)
		defaultMachinePoolArgs.Cluster = &clusterID

		By("Make sure the default machinepool imported from cluster state")
		imported, _ := h.CheckDefaultMachinePoolImported()
		if !imported {
			By("Create default machinepool by importing from CMS ")
			_, err = dmpService.Apply(&defaultMachinePoolArgs, false)
			Expect(err).ToNot(HaveOccurred())
		}

	})

	It("check the default machinepool edit/delete operations with additional mp exists it - [id:69727]", ci.Critical, func() {
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
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &name,
		}

		_, err = mpService.Apply(MachinePoolArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Import the default machinepool state")
		_, err = dmpService.Apply(&defaultMachinePoolArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Destroy default machinepool")
		output, err = dmpService.Destroy()
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring("Destroy complete! Resources: 1 destroyed."))

		By("Create default machinepool after delete")
		DefaultMachinePoolArgs := &exe.MachinePoolArgs{
			Cluster:     &clusterID,
			Replicas:    &replicas,
			MachineType: &machineType,
			Name:        &defaultMachinePoolNmae,
		}
		output, err = dmpService.Apply(DefaultMachinePoolArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(output).To(ContainSubstring("machine pool 'worker' was deleted"))
		Expect(output).To(ContainSubstring("Please use a different name"))
	})
})
