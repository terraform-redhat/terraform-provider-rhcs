package e2e

import (

	// nolint

	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmsv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"

	"github.com/openshift-online/ocm-common/pkg/aws/aws_client"
	"github.com/openshift-online/ocm-common/pkg/test/vpc_client"
)

var _ = Describe("Create MachinePool", ci.Day2, ci.FeatureMachinepool, func() {
	defer GinkgoRecover()

	var (
		mpService      exec.MachinePoolService
		profileHandler profilehandler.ProfileHandler
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Classic cluster")
		}

		mpService, err = profileHandler.Services().GetMachinePoolsService()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		mpService.Destroy()
	})

	It("can create a second machine pool - [id:64757]", ci.High, func() {
		By("Create a second machine pool")
		replicas := 3
		machineType := "r5.xlarge"
		name := "ocp-64757"
		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
		}

		_, err := mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())
		_, err = mpService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpsOut, err := mpService.Output()
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mpsOut.MachinePools)).To(Equal(1))
		mpOut := mpsOut.MachinePools[0]

		mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
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
		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
			Labels:      helper.StringMapPointer(creationLabels),
		}

		_, err := mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Labels()).To(Equal(creationLabels))

		By("Edit the labels of the machinepool")
		mpArgs.Labels = helper.StringMapPointer(updatingLabels)
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Labels()).To(Equal(updatingLabels))

		By("Delete the labels of the machinepool")
		mpArgs.Labels = helper.StringMapPointer(emptyLabels)
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the updated machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
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
		mpArgs := &exec.MachinePoolArgs{
			Cluster:            helper.StringPointer(clusterID),
			MinReplicas:        helper.IntPointer(minReplicas),
			MaxReplicas:        helper.IntPointer(maxReplicas),
			MachineType:        helper.StringPointer(machineType),
			Name:               helper.StringPointer(name),
			AutoscalingEnabled: helper.BoolPointer(true),
		}

		_, err := mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Autoscaling().MinReplicas()).To(Equal(minReplicas))
		Expect(mpResponseBody.Autoscaling().MaxReplicas()).To(Equal(maxReplicas))

		By("Change the number of replicas of the machinepool")
		minReplicas = minReplicas * 2
		maxReplicas = maxReplicas * 2
		mpArgs.MinReplicas = helper.IntPointer(minReplicas)
		mpArgs.MaxReplicas = helper.IntPointer(maxReplicas)
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the updated machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Autoscaling().MinReplicas()).To(Equal(minReplicas))
		Expect(mpResponseBody.Autoscaling().MaxReplicas()).To(Equal(maxReplicas))

		By("Disable autoscaling of the machinepool")
		mpArgs = &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
		}

		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the updated machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Autoscaling()).To(BeNil())
	})

	It("edit second machinepool taints - [id:64904]", ci.High, func() {
		By("Create additional machinepool with labels")
		replicas := 3
		machineType := "r5.xlarge"
		name := "ocp-64904"
		taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": constants.NoExecute}
		taint1 := map[string]string{"key": "k2", "value": "val2", "schedule_type": constants.NoSchedule}
		taint2 := map[string]string{"key": "k3", "value": "val3", "schedule_type": constants.PreferNoSchedule}
		taints := []map[string]string{taint0, taint1}
		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
			Taints:      &taints,
		}

		_, err := mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
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
		mpArgs.Taints = &taints
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the updated machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		respTaints = mpResponseBody.Taints()
		for index, taint := range respTaints {
			Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
			Expect(taint.Key()).To(Equal(taints[index]["key"]))
			Expect(taint.Value()).To(Equal(taints[index]["value"]))
		}

		By("Delete the taints of the machinepool")
		mpArgs.Taints = nil
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the updated machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Taints()).To(BeNil())
	})

	It("can validate the machinepool creation - [id:68283]", ci.High, func() {
		By("Check the validations for the machinepool creation rosa cluster")
		var (
			machinepoolName                  = "ocp-68283"
			invalidMachinepoolName           = "%^#@"
			machineType, InvalidInstanceType = "r5.xlarge", "custom-4-16384"
			mpReplicas                       = 3
			minReplicas                      = 3
			maxReplicas                      = 6
			invalidMpReplicas                = -3
			invalidMinReplicas4Mutilcluster  = 4
			invalidMaxReplicas4Mutilcluster  = 7
		)
		By("Create machinepool with invalid name")
		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(invalidMpReplicas),
			Name:        helper.StringPointer(invalidMachinepoolName),
			MachineType: helper.StringPointer(machineType),
		}
		_, err := mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("Expected a valid value for 'name'"))

		By("Create machinepool with invalid replica value")
		mpArgs = &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(invalidMpReplicas),
			Name:        helper.StringPointer(machinepoolName),
			MachineType: helper.StringPointer(machineType),
		}
		_, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		helper.ExpectTFErrorContains(err, "Attribute 'replicas' must be a non-negative integer")

		By("Create machinepool with invalid instance type")
		mpArgs = &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(mpReplicas),
			Name:        helper.StringPointer(machinepoolName),
			MachineType: helper.StringPointer(InvalidInstanceType),
		}
		_, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		helper.ExpectTFErrorContains(err, "Machine type '"+InvalidInstanceType+"' is not supported for cloud provider")

		By("Create machinepool with setting replicas and enable-autoscaling at the same time")
		mpArgs = &exec.MachinePoolArgs{
			Cluster:            helper.StringPointer(clusterID),
			Replicas:           helper.IntPointer(mpReplicas),
			Name:               helper.StringPointer(machinepoolName),
			AutoscalingEnabled: helper.BoolPointer(true),
			MachineType:        helper.StringPointer(machineType),
		}
		_, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		helper.ExpectTFErrorContains(err, "when enabling autoscaling, should set value for maxReplicas")

		By("Create machinepool with setting min-replicas large than max-replicas")
		mpArgs = &exec.MachinePoolArgs{
			Cluster:            helper.StringPointer(clusterID),
			MinReplicas:        helper.IntPointer(maxReplicas),
			MaxReplicas:        helper.IntPointer(minReplicas),
			Name:               helper.StringPointer(machinepoolName),
			AutoscalingEnabled: helper.BoolPointer(true),
			MachineType:        helper.StringPointer(machineType),
		}
		_, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("'min_replicas' must be less than or equal to 'max_replicas'"))

		By("Create machinepool with setting min-replicas and max-replicas but without setting --enable-autoscaling")
		mpArgs = &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			MinReplicas: helper.IntPointer(minReplicas),
			MaxReplicas: helper.IntPointer(maxReplicas),
			Name:        helper.StringPointer(machinepoolName),
			MachineType: helper.StringPointer(machineType),
		}
		_, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		helper.ExpectTFErrorContains(err, "when disabling autoscaling, cannot set min_replicas and/or max_replicas")

		By("Create machinepool with setting min-replicas large than max-replicas")
		if profileHandler.Profile().IsMultiAZ() {
			By("Create machinepool with setting min-replicas and max-replicas not multiple 3 for multi-az")
			mpArgs = &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				MinReplicas:        helper.IntPointer(minReplicas),
				MaxReplicas:        helper.IntPointer(invalidMaxReplicas4Mutilcluster),
				Name:               helper.StringPointer(machinepoolName),
				MachineType:        helper.StringPointer(machineType),
				AutoscalingEnabled: helper.BoolPointer(true),
			}
			_, err = mpService.Apply(mpArgs)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Multi AZ clusters require that the number of replicas be a multiple of 3")

			mpArgs = &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				MinReplicas:        helper.IntPointer(invalidMinReplicas4Mutilcluster),
				MaxReplicas:        helper.IntPointer(maxReplicas),
				Name:               helper.StringPointer(machinepoolName),
				MachineType:        helper.StringPointer(machineType),
				AutoscalingEnabled: helper.BoolPointer(true),
			}
			_, err = mpService.Apply(mpArgs)
			Expect(err).To(HaveOccurred())
			helper.ExpectTFErrorContains(err, "Multi AZ clusters require that the number of replicas be a multiple of 3")

		}
	})

	It("can create single-az machinepool for multi-az cluster - [id:65063]", ci.High, func() {
		if !profileHandler.Profile().IsMultiAZ() {
			Skip("The test is configured for MultiAZ cluster only")
		}
		getResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		azs := getResp.Body().Nodes().AvailabilityZones()
		By("Create additional machinepool with availability zone specified")
		replicas := 1
		machineType := "r5.xlarge"
		name := "ocp-65063"
		MachinePoolArgs := &exec.MachinePoolArgs{
			Cluster:          helper.StringPointer(clusterID),
			Replicas:         helper.IntPointer(replicas),
			MachineType:      helper.StringPointer(machineType),
			Name:             helper.StringPointer(name),
			AvailabilityZone: helper.StringPointer(azs[0]),
		}

		_, err = mpService.Apply(MachinePoolArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.AvailabilityZones()[0]).To(Equal(azs[0]))

		_, err = mpService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Create additional machinepool with subnet id specified")
		awsSubnetIds := getResp.Body().AWS().SubnetIDs()
		MachinePoolArgs = &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
			SubnetID:    helper.StringPointer(awsSubnetIds[0]),
		}

		_, err = mpService.Apply(MachinePoolArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Subnets()[0]).To(Equal(awsSubnetIds[0]))

		_, err = mpService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Create additional machinepool with multi_availability_zone=false specified")
		MachinePoolArgs = &exec.MachinePoolArgs{
			Cluster:          helper.StringPointer(clusterID),
			Replicas:         helper.IntPointer(replicas),
			MachineType:      helper.StringPointer(machineType),
			Name:             helper.StringPointer(name),
			MultiAZ:          helper.BoolPointer(false),
			AvailabilityZone: helper.StringPointer(azs[1]),
		}

		_, err = mpService.Apply(MachinePoolArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mpResponseBody.AvailabilityZones())).To(Equal(1))
	})

	It("can create machinepool with subnet_id option for BYO VPC single-az cluster - [id:65071]", ci.High, func() {
		if profileHandler.Profile().IsMultiAZ() || !profileHandler.Profile().IsBYOVPC() {
			Skip("The test is configured for SingleAZ BYO VPC cluster only")
		}

		By("Retrieve current availability Zone")
		vpcService, err := profileHandler.Services().GetVPCService()
		Expect(err).ToNot(HaveOccurred())
		vpcOutput, err := vpcService.Output()
		Expect(err).ToNot(HaveOccurred())
		az := vpcOutput.AvailabilityZones[0]

		By("Add 2 subnets to VPC")
		awsClient, err := aws_client.CreateAWSClient("", "")
		Expect(err).ToNot(HaveOccurred())
		vpcClient, err := vpc_client.GenerateVPCBySubnet(vpcOutput.PrivateSubnets[0], profileHandler.Profile().GetRegion())
		Expect(err).ToNot(HaveOccurred())
		subnet, err := vpcClient.CreateSubnet(az)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			By("Remove added subnet from VPC")
			// subnet delettion should be all dependency of the subnet is removed
			// In this cases, the machinepool is the subnet dependency, it needs some time to be deleted from aws
			var deleteErr error
			Eventually(func() error {
				_, err := awsClient.DeleteSubnet(subnet.ID)
				if err != nil {
					if strings.Contains(err.Error(), "DependencyViolation") {
						return err
					}
					deleteErr = err
					return nil
				}
				return nil
			}).WithTimeout(10 * time.Minute).WithPolling(100 * time.Second).Should(Succeed())
			Expect(deleteErr).ToNot(HaveOccurred())
		}()

		By("Tag new subnet to be able to apply it to the machinepool")
		clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		vpcTagService, err := profileHandler.Services().GetVPCTagService()
		Expect(err).ToNot(HaveOccurred())
		tagKey := fmt.Sprintf("kubernetes.io/cluster/%s", clusterResp.Body().InfraID())
		tagValue := "shared"
		VPCTagArgs := &exec.VPCTagArgs{
			AWSRegion: helper.StringPointer(clusterResp.Body().Region().ID()),
			IDs:       helper.StringSlicePointer(append(vpcOutput.PrivateSubnets, subnet.ID)),
			TagKey:    helper.StringPointer(tagKey),
			TagValue:  helper.StringPointer(tagValue),
		}
		_, err = vpcTagService.Apply(VPCTagArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Create additional machinepool with subnet id specified")
		replicas := 1
		machineType := "r5.xlarge"
		name := "ocp-65071"

		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
			SubnetID:    helper.StringPointer(subnet.ID),
		}
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())
		defer mpService.Destroy()

		mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Subnets()[0]).To(Equal(subnet.ID))

		replicas = 4
		mpArgs.Replicas = helper.IntPointer(replicas)
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the updated machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.Replicas()).To(Equal(4))
	})

	It("can create machinepool with additional security group - [id:69146]", ci.High, func() {
		if !profileHandler.Profile().IsBYOVPC() {
			Skip("This case only works for BYOVPC cluster profile")
		}

		By("Prepare additional security groups")
		sgService, err := profileHandler.Services().GetSecurityGroupService()
		Expect(err).ToNot(HaveOccurred())
		output, err := sgService.Output()
		Expect(err).ToNot(HaveOccurred())
		if output.SGIDs == nil {
			vpcService, err := profileHandler.Services().GetVPCService()
			Expect(err).ToNot(HaveOccurred())
			vpcOutput, err := vpcService.Output()
			Expect(err).ToNot(HaveOccurred())
			sgArgs := &exec.SecurityGroupArgs{
				AWSRegion: helper.StringPointer(profileHandler.Profile().GetRegion()),
				VPCID:     helper.StringPointer(vpcOutput.VPCID),
				SGNumber:  helper.IntPointer(4),
			}
			_, err = sgService.Apply(sgArgs)
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
		mpArgs := &exec.MachinePoolArgs{
			Cluster:                  helper.StringPointer(clusterID),
			Replicas:                 helper.IntPointer(replicas),
			MachineType:              helper.StringPointer(machineType),
			Name:                     helper.StringPointer(name),
			AdditionalSecurityGroups: helper.StringSlicePointer(sgIDs),
		}

		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			_, err = mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		}()

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mpResponseBody.AWS().AdditionalSecurityGroupIds())).To(Equal(len(sgIDs)))
		for _, sg := range mpResponseBody.AWS().AdditionalSecurityGroupIds() {
			Expect(sg).To(BeElementOf(sgIDs))
		}

		By("Update security groups is not allowed to a machinepool")
		testAdditionalSecurityGroups := output.SGIDs[0:1]
		mpArgs = &exec.MachinePoolArgs{
			Cluster:                  helper.StringPointer(clusterID),
			Replicas:                 helper.IntPointer(replicas),
			MachineType:              helper.StringPointer(machineType),
			Name:                     helper.StringPointer(name),
			AdditionalSecurityGroups: helper.StringSlicePointer(testAdditionalSecurityGroups),
		}

		applyOutput, err := mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		Expect(applyOutput).Should(ContainSubstring("Attribute aws_additional_security_group_ids, cannot be changed"))

		By("Destroy the machinepool")
		_, err = mpService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Create another machinepool without additional sg ")
		name = "add-69146"
		mpArgs = &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer("m5.2xlarge"),
			Name:        helper.StringPointer(name),
		}

		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(mpResponseBody.AWS().AdditionalSecurityGroupIds()).To(BeNil())
		Expect(mpResponseBody.InstanceType()).To(Equal("m5.2xlarge"))
	})

	It("can create machinepool with customized tags - [id:73942]", ci.High, func() {

		By("Create a machinepool with variable aws_tags")
		name := "mp-73942"
		replicas := 0
		machineType := "r5.xlarge"

		validTags := map[string]string{
			"tagsKey": "tagValue",
		}
		mpArgs := &exe.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
			Tags:        helper.StringMapPointer(validTags),
		}
		_, err := mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Check the machinepool detail state")
		resp, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, name)
		Expect(err).ToNot(HaveOccurred())
		for tagKey, tagValue := range validTags {
			Expect(tagValue).To(BeElementOf(resp.AWS().Tags()[tagKey]))
		}

		By("Update the machinepool tags is not allowed")
		validTags["tagKey2"] = "tagValue2"
		_, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("Attribute aws_tags, cannot be changed from"))
	})

	It("can reconciliate with basic flow - [id:65645]", ci.Medium, func() {
		By("Create additional machinepool with taints")
		replicas := 3
		machineType := "r5.xlarge"
		mpName := "ocp-65645"
		taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": constants.NoExecute}
		taints := []map[string]string{taint0}
		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(mpName),
			Taints:      &taints,
		}

		_, err := mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the created machinepool")
		_, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, mpName)
		Expect(err).ToNot(HaveOccurred())

		By("Delete machinepool by OCM API")
		cms.DeleteMachinePool(cms.RHCSConnection, clusterID, mpName)
		_, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, mpName)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).Should(ContainSubstring("Machine pool with id '%s' not found", mpName))

		By("Sleep 1 min to wait for the MP deletion processed")
		time.Sleep(time.Minute)

		By("Re-apply the machinepool manifest")
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify the parameters of the machinepool")
		mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, mpName)
		Expect(err).ToNot(HaveOccurred())
		respTaints := mpResponseBody.Taints()
		for index, taint := range respTaints {
			Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
			Expect(taint.Key()).To(Equal(taints[index]["key"]))
			Expect(taint.Value()).To(Equal(taints[index]["value"]))
		}
	})

	It("will validate well - [id:63139]", ci.Medium, func() {
		By("Will validate the subnet")
		mpArgs := &exe.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(0),
			MachineType: helper.StringPointer("m5.xlarge"),
			Name:        helper.StringPointer("invalidsub"),
			SubnetID:    helper.StringPointer("subnet-invalidsubnetid"),
		}
		output, err := mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).Should(ContainSubstring("Failed to find subnet"))

		By("Will validate the instance type")
		mpArgs = &exe.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(0),
			MachineType: helper.StringPointer("invalid"),
			Name:        helper.StringPointer("invalidinstype"),
		}
		output, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).Should(ContainSubstring("Machine type 'invalid' is not supported"))

		By("Create machinepool creation plan with invalid tags")
		invalidTags := map[string]string{
			"aws:tags": "awsvalue",
		}
		replicas := 0
		machineType := "r5.xlarge"
		name := "mp-73942"

		mpArgs = &exe.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
			Tags:        helper.StringMapPointer(invalidTags),
		}
		output, err = mpService.Apply(mpArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).Should(ContainSubstring("Tags that begin with 'aws:' are reserved"))

	})
})

var _ = Describe("Import MachinePool", ci.Day2, ci.FeatureImport, func() {
	var profileHandler profilehandler.ProfileHandler
	var mpService exec.MachinePoolService
	var importService exec.ImportService

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Classic cluster")
		}

		mpService, err = profileHandler.Services().GetMachinePoolsService()
		Expect(err).ToNot(HaveOccurred())

		importService, err = profileHandler.Services().GetImportService()
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		By("Destroy machinepool")
		mpService.Destroy()

		By("Destroy import service")
		importService.Destroy()
	})

	It("can import resource - [id:66403]", ci.Medium, func() {
		By("Create additional machinepool for import")
		minReplicas := 3
		maxReplicas := 6
		creationLabels := map[string]string{"foo1": "bar1"}
		machineType := "r5.xlarge"
		name := "ocp-66403"
		mpArgs := &exec.MachinePoolArgs{
			Cluster:            helper.StringPointer(clusterID),
			MinReplicas:        helper.IntPointer(minReplicas),
			MaxReplicas:        helper.IntPointer(maxReplicas),
			MachineType:        helper.StringPointer(machineType),
			Name:               helper.StringPointer(name),
			Labels:             helper.StringMapPointer(creationLabels),
			AutoscalingEnabled: helper.BoolPointer(true),
		}

		_, err := mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Run the command to import the machinepool")
		importParam := &exec.ImportArgs{
			ClusterID:  clusterID,
			Resource:   "rhcs_machine_pool.mp_import",
			ObjectName: name,
		}
		_, err = importService.Import(importParam)
		Expect(err).To(Succeed())

		By("Check resource state - import command succeeded")
		output, err := importService.ShowState(importParam.Resource)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring(machineType))
		Expect(output).To(ContainSubstring(name))
		Expect(output).To(MatchRegexp("foo1"))
		Expect(output).To(MatchRegexp("bar1"))
	})
})

var _ = Describe("Edit MachinePool", ci.Day2, ci.FeatureMachinepool, func() {

	var (
		profileHandler                 profilehandler.ProfileHandler
		dmpService                     exec.MachinePoolService
		mpService                      exec.MachinePoolService
		defaultMachinePoolName         = "worker"
		defaultMachinepoolResponse     *cmsv1.MachinePool
		originalDefaultMachinepoolArgs *exec.MachinePoolArgs
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Classic cluster")
		}

		dmpTFWorkspace := helper.GenerateRandomName("dft-"+profileHandler.Profile().GetName(), 2)
		dmpService, err = exec.NewMachinePoolService(dmpTFWorkspace, profileHandler.Profile().GetClusterType())
		Expect(err).ToNot(HaveOccurred())
		mpService, err = profileHandler.Services().GetMachinePoolsService()
		Expect(err).ToNot(HaveOccurred())

		defaultMachinepoolResponse, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, defaultMachinePoolName)
		if err != nil && strings.Contains(err.Error(), fmt.Sprintf("Machine pool with id '%s' not found", defaultMachinePoolName)) {
			Skip("The default machinepool does not exist")
		}

		By("Create default machinepool by importing from CMS ")
		originalDefaultMachinepoolArgs = exec.BuildMachinePoolArgsFromCSResponse(clusterID, defaultMachinepoolResponse)
		_, err = dmpService.Apply(originalDefaultMachinepoolArgs)
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		// Check if current test is skipped, skip this AfterEach block too
		if CurrentSpecReport().Failure.Message == "The default machinepool does not exist" {
			return
		}

		By("Recover the default machinepool to the original state")
		dmpService.Apply(originalDefaultMachinepoolArgs)

		By("Destroy additonal mp")
		mpService.Destroy()
	})

	It("check the validations and some negative scenarios of creating/editing/deleting default machinepool - [id:69073]", ci.Medium, func() {
		By("Create machinepool with the default machinepool name 'worker' when it does exist")
		output, err := dmpService.Apply(originalDefaultMachinepoolArgs)
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring("No changes. Your infrastructure matches the configuration."))

		dmpArgFromMachinepoolForTesting, err := dmpService.ReadTFVars()
		Expect(err).ToNot(HaveOccurred())
		zeroReplicas := 0
		if _, ok := defaultMachinepoolResponse.GetAutoscaling(); ok {
			By("Edit the deafult machinepool max and min replicas to 0")
			dmpArgFromMachinepoolForTesting.MaxReplicas = helper.IntPointer(zeroReplicas)
			dmpArgFromMachinepoolForTesting.MinReplicas = helper.IntPointer(zeroReplicas)
			_, err = dmpService.Apply(dmpArgFromMachinepoolForTesting)
			Expect(err).To(HaveOccurred())
			Expect(output).To(ContainSubstring("Failed to update machine pool"))
			Expect(output).To(ContainSubstring("must be a integer greater than 0"))
		} else {
			By("Edit the deafult machinepool replicas to 0")
			dmpArgFromMachinepoolForTesting.Replicas = helper.IntPointer(zeroReplicas)
			_, err = dmpService.Apply(dmpArgFromMachinepoolForTesting)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to update machine pool"))
			Expect(err.Error()).To(ContainSubstring("least one machine pool able to run OCP workload is required"))
		}

		By("Check the machine type change will triger re-creation")
		dmpArgFromMachinepoolForTesting, err = dmpService.ReadTFVars()
		Expect(err).ToNot(HaveOccurred())
		dmpArgFromMachinepoolForTesting.MachineType = helper.StringPointer("r5.xlarge")
		out, err := dmpService.Apply(dmpArgFromMachinepoolForTesting)
		Expect(err).To(HaveOccurred())
		Expect(out).To(ContainSubstring("machine_type, cannot be changed"))

		By("Delete dmp without additional mp exists")
		resp, err := cms.ListMachinePool(cms.RHCSConnection, clusterID)
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
			taint0 := map[string]string{"key": "k1", "value": "val", "schedule_type": constants.NoExecute}
			taint1 := map[string]string{"key": "k2", "value": "val2", "schedule_type": constants.NoSchedule}
			taints := []map[string]string{taint0, taint1}
			defaultMPName := "worker"

			dmpArgFromMachinepoolForTesting, err := dmpService.ReadTFVars()
			Expect(err).ToNot(HaveOccurred())

			By("Edit the taints without additional machinepool")
			dmpArgFromMachinepoolForTesting.Taints = &taints
			_, err = dmpService.Apply(dmpArgFromMachinepoolForTesting)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to update machine pool"))
			Expect(err.Error()).To(ContainSubstring("least one machine pool able to run OCP workload is required. Pool should not"))

			dmpArgFromMachinepoolForTesting, err = dmpService.ReadTFVars()
			Expect(err).ToNot(HaveOccurred())
			if _, ok := defaultMachinepoolResponse.GetAutoscaling(); ok {
				By("Edit default machinepool with autoscale configuration")
				minReplicas := 3
				maxReplicas := 6
				dmpArgFromMachinepoolForTesting.MinReplicas = helper.IntPointer(minReplicas)
				dmpArgFromMachinepoolForTesting.MaxReplicas = helper.IntPointer(maxReplicas)
				_, err := dmpService.Apply(dmpArgFromMachinepoolForTesting)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, defaultMPName)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Autoscaling().MinReplicas()).To(Equal(minReplicas))
				Expect(mpResponseBody.Autoscaling().MaxReplicas()).To(Equal(maxReplicas))
			} else {
				By("Edit default machinepool with replicas")
				replicas := 6
				dmpArgFromMachinepoolForTesting.Replicas = helper.IntPointer(replicas)
				_, err := dmpService.Apply(dmpArgFromMachinepoolForTesting)
				Expect(err).ToNot(HaveOccurred())

				By("Verify the parameters of the created machinepool")
				mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, defaultMPName)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Replicas()).To(Equal(replicas))
			}

			By("Edit default machinepool with labels")
			creationLabels := map[string]string{"fo1": "bar1", "fo2": "baz2"}
			dmpArgFromMachinepoolForTesting.Labels = helper.StringMapPointer(creationLabels)
			_, err = dmpService.Apply(dmpArgFromMachinepoolForTesting)
			Expect(err).ToNot(HaveOccurred())

			By("Verify the parameters of the created machinepool")
			mpResponseBody, err := cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, defaultMPName)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Labels()).To(Equal(creationLabels))

			By("Create an additional machinepool")
			replicas := 3
			machineType := "r5.xlarge"
			name := "amp-69009"

			mpArgs := &exec.MachinePoolArgs{
				Cluster:     helper.StringPointer(clusterID),
				Replicas:    helper.IntPointer(replicas),
				MachineType: helper.StringPointer(machineType),
				Name:        helper.StringPointer(name),
			}

			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Edit the default machinepool with taints")
			dmpArgFromMachinepoolForTesting.Taints = &taints
			_, err = dmpService.Apply(dmpArgFromMachinepoolForTesting)
			Expect(err).ToNot(HaveOccurred())

			By("Verify the parameters of the default machinepool")
			mpResponseBody, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, "worker")
			Expect(err).ToNot(HaveOccurred())
			respTaints := mpResponseBody.Taints()
			for index, taint := range respTaints {
				Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
				Expect(taint.Key()).To(Equal(taints[index]["key"]))
				Expect(taint.Value()).To(Equal(taints[index]["value"]))
			}
		})

})

var _ = Describe("Destroy MachinePool", ci.Day3, ci.FeatureMachinepool, func() {
	var (
		profileHandler             profilehandler.ProfileHandler
		dmpService                 exec.MachinePoolService
		defaultMachinePoolArgs     *exec.MachinePoolArgs
		mpService                  exec.MachinePoolService
		defaultMachinepoolResponse *cmsv1.MachinePool
		defaultMachinePoolName     = "worker"
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Classic cluster")
		}

		dmpTFWorkspace := helper.GenerateRandomName("dft-"+profileHandler.Profile().GetName(), 2)
		dmpService, err = exec.NewMachinePoolService(dmpTFWorkspace, profileHandler.Profile().GetClusterType())
		Expect(err).ToNot(HaveOccurred())
		mpService, err = profileHandler.Services().GetMachinePoolsService()
		Expect(err).ToNot(HaveOccurred())

		defaultMachinepoolResponse, err = cms.RetrieveClusterMachinePool(cms.RHCSConnection, clusterID, defaultMachinePoolName)
		Expect(err).ToNot(HaveOccurred())
		defaultMachinePoolArgs = exec.BuildMachinePoolArgsFromCSResponse(clusterID, defaultMachinepoolResponse)
		defaultMachinePoolName = *defaultMachinePoolArgs.Name

		By("Make sure the default machinepool imported from cluster state")
		_, err = dmpService.ShowState("rhcs_machine_pool.mp")
		if err != nil {
			By("Create default machinepool by importing from CMS ")
			_, err = dmpService.Apply(defaultMachinePoolArgs)
			Expect(err).ToNot(HaveOccurred())
		}

	})

	It("check the default machinepool edit/delete operations with additional mp exists it - [id:69727]", ci.Critical, func() {
		By("Destroy default machinepool without additional machinepool existing")
		resp, err := cms.ListMachinePool(cms.RHCSConnection, clusterID)
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

		mpArgs := &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(name),
		}

		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Import the default machinepool state")
		_, err = dmpService.Apply(defaultMachinePoolArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Destroy default machinepool")
		output, err = dmpService.Destroy()
		Expect(err).ToNot(HaveOccurred())
		Expect(output).To(ContainSubstring("Destroy complete! Resources: 1 destroyed."))

		By("Create default machinepool after delete")
		defaultMachinePoolArgs = &exec.MachinePoolArgs{
			Cluster:     helper.StringPointer(clusterID),
			Replicas:    helper.IntPointer(replicas),
			MachineType: helper.StringPointer(machineType),
			Name:        helper.StringPointer(defaultMachinePoolName),
		}
		output, err = dmpService.Apply(defaultMachinePoolArgs)
		Expect(err).To(HaveOccurred())
		Expect(output).To(ContainSubstring("machine pool 'worker' was deleted"))
		Expect(output).To(ContainSubstring("Please use a different name"))
	})
})
