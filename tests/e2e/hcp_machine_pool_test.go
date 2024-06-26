package e2e

import (

	// nolint

	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("HCP MachinePool", ci.Day2, ci.FeatureMachinepool, func() {
	defer GinkgoRecover()
	var (
		mpService exec.MachinePoolService
		tcService exec.TuningConfigService
		vpcOutput *exec.VPCOutput
		profile   *ci.Profile
	)

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()

		if !profile.GetClusterType().HCP {
			Skip("Test can run only on Hosted cluster")
		}

		var err error
		mpService, err = exec.NewMachinePoolService(constants.HCPMachinePoolDir)
		Expect(err).ToNot(HaveOccurred())
		tcService, err = exec.NewTuningConfigService(constants.TuningConfigDir)
		Expect(err).ToNot(HaveOccurred())

		By("Get vpc output")
		vpcService, err := exec.NewVPCService(constants.GetAWSVPCDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())
		vpcOutput, err = vpcService.Output()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		mpService.Destroy()
	})

	It("can be created with only required attributes - [id:72504]",
		ci.Critical, func() {
			By("Retrieve current cluster information")
			clusterRespBody, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			var initialMinReplicas int
			var initialMaxReplicas int
			var initialReplicas int
			if profile.Autoscale {
				initialMinReplicas = clusterRespBody.Body().Nodes().AutoscaleCompute().MinReplicas()
				initialMaxReplicas = clusterRespBody.Body().Nodes().AutoscaleCompute().MaxReplicas()
			} else {
				initialReplicas = clusterRespBody.Body().Nodes().Compute()
			}

			By("Create machinepool")
			replicas := 3
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72504", 2)
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			mpArgs := &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           helper.IntPointer(replicas),
				Name:               helper.StringPointer(name),
				SubnetID:           helper.StringPointer(subnetId),
				MachineType:        helper.StringPointer(machineType),
				AutoRepair:         helper.BoolPointer(true),
			}
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify attributes are correctly set")
			mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.ID()).To(Equal(name))
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))
			Expect(mpResponseBody.Subnet()).To(Equal(subnetId))
			Expect(mpResponseBody.AWSNodePool().InstanceType()).To(Equal(machineType))
			Expect(mpResponseBody.AutoRepair()).To(BeTrue())

			By("Wait for machinepool replicas available")
			err = wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 20*time.Minute, false, func(context.Context) (bool, error) {
				clusterRespBody, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				if err != nil {
					return false, err
				}
				if profile.Autoscale {
					return clusterRespBody.Body().Nodes().AutoscaleCompute().MaxReplicas() == (initialMaxReplicas+replicas) &&
						clusterRespBody.Body().Nodes().AutoscaleCompute().MinReplicas() == (initialMinReplicas+replicas), nil
				} else {
					return clusterRespBody.Body().Nodes().Compute() == (initialReplicas + replicas), nil
				}
			})
			helper.AssertWaitPollNoErr(err, "Replicas are not ready after 600")

			By("Delete machinepool")
			_, err = mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).To(HaveOccurred())
		})

	It("can create/edit autoscaling - [id:72505]",
		ci.Critical, func() {
			minReplicas := 2
			maxReplicas := 4
			replicas := 3
			machineType := "m5.xlarge"
			name := helper.GenerateRandomName("np-72505", 2)
			subnetId := vpcOutput.ClusterPrivateSubnets[0]

			By("Create machinepool")
			mpArgs := &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				AutoscalingEnabled: helper.BoolPointer(true),
				MinReplicas:        helper.IntPointer(minReplicas),
				MaxReplicas:        helper.IntPointer(maxReplicas),
				Name:               helper.StringPointer(name),
				SubnetID:           helper.StringPointer(subnetId),
				MachineType:        helper.StringPointer(machineType),
				AutoRepair:         helper.BoolPointer(true),
			}
			_, err := mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).ToNot(BeNil())
			Expect(mpResponseBody.Autoscaling().MaxReplica()).To(Equal(maxReplicas))
			Expect(mpResponseBody.Autoscaling().MinReplica()).To(Equal(minReplicas))

			By("Update autoscaling")
			minReplicas = 1
			maxReplicas = 3
			mpArgs.MinReplicas = helper.IntPointer(minReplicas)
			mpArgs.MaxReplicas = helper.IntPointer(maxReplicas)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).ToNot(BeNil())
			Expect(mpResponseBody.Autoscaling().MaxReplica()).To(Equal(maxReplicas))
			Expect(mpResponseBody.Autoscaling().MinReplica()).To(Equal(minReplicas))

			By("Disable autoscaling")
			mpArgs.AutoscalingEnabled = helper.BoolPointer(false)
			mpArgs.MinReplicas = nil
			mpArgs.MaxReplicas = nil
			mpArgs.Replicas = helper.IntPointer(replicas)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))

			By("Scale up")
			replicas = 4
			mpArgs.Replicas = helper.IntPointer(replicas)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))

			By("Scale to zero")
			replicas = 0
			mpArgs.Replicas = helper.IntPointer(replicas)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))

			By("Enable back autoscaling")
			minReplicas = 1
			maxReplicas = 2
			mpArgs.Replicas = nil
			mpArgs.AutoscalingEnabled = helper.BoolPointer(true)
			mpArgs.MinReplicas = helper.IntPointer(minReplicas)
			mpArgs.MaxReplicas = helper.IntPointer(maxReplicas)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).ToNot(BeNil())
			Expect(mpResponseBody.Autoscaling().MaxReplica()).To(Equal(maxReplicas))
			Expect(mpResponseBody.Autoscaling().MinReplica()).To(Equal(minReplicas))
		})

	It("can be created with security groups - [id:73068]", ci.High,
		func() {
			By("Prepare additional security groups")
			sgService, err := exec.NewSecurityGroupService()
			output, err := sgService.Output()
			Expect(err).ToNot(HaveOccurred())
			if output.SGIDs == nil {
				sgArgs := &exec.SecurityGroupArgs{
					AWSRegion: helper.StringPointer(profile.Region),
					VPCID:     helper.StringPointer(vpcOutput.VPCID),
					SGNumber:  helper.IntPointer(4),
				}
				_, err = sgService.Apply(sgArgs)
				Expect(err).ToNot(HaveOccurred())
				defer sgService.Destroy()
			}

			output, err = sgService.Output()
			Expect(err).ToNot(HaveOccurred())

			replicas := 0
			machineType := "r5.xlarge"
			name := "ocp-73068"
			sgIDs := output.SGIDs
			if len(sgIDs) >= 4 {
				sgIDs = sgIDs[0:4]
			}

			// workaround
			By("Create machinepool")
			mpArgs := &exec.MachinePoolArgs{
				Cluster:                  helper.StringPointer(clusterID),
				Replicas:                 helper.IntPointer(replicas),
				MachineType:              helper.StringPointer(machineType),
				Name:                     helper.StringPointer(name),
				AdditionalSecurityGroups: helper.StringSlicePointer(output.SGIDs),
				AutoscalingEnabled:       helper.BoolPointer(false),
				AutoRepair:               helper.BoolPointer(false),
				SubnetID:                 helper.StringPointer(vpcOutput.ClusterPrivateSubnets[0]),
			}
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_, err = mpService.Destroy()
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
			mpArgs.AdditionalSecurityGroups = helper.StringSlicePointer(output.SGIDs[0:1])
			applyOutput, err := mpService.Apply(mpArgs)
			Expect(err).To(HaveOccurred())
			Expect(applyOutput).Should(ContainSubstring("aws_node_pool.additional_security_group_ids, cannot be changed"))

			By("Destroy the machinepool")
			_, err = mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())

			By("Create another machinepool without additional sg")
			name = "add-73068"
			machineType = "m5.2xlarge"
			mpArgs.Name = helper.StringPointer(name)
			mpArgs.MachineType = helper.StringPointer(machineType)
			mpArgs.AdditionalSecurityGroups = nil
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify the parameters of the created machinepool")
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.AWSNodePool().AdditionalSecurityGroupIds()).To(BeNil())
			Expect(mpResponseBody.AWSNodePool().InstanceType()).To(Equal("m5.2xlarge"))
		})

	It("can validate nodepool creation with security groups - [id:73069]", ci.Low,
		func() {

			By("Prepare additional security groups")
			replicas := 0
			machineType := "r5.xlarge"
			name := "ocp-73069"
			fakeSgIDs := []string{"sg-fake"}

			By("Run terraform apply cannot work with invalid sg IDs")
			mpArgs := &exec.MachinePoolArgs{
				Cluster:                  helper.StringPointer(clusterID),
				Replicas:                 helper.IntPointer(replicas),
				MachineType:              helper.StringPointer(machineType),
				Name:                     helper.StringPointer(name),
				AdditionalSecurityGroups: helper.StringSlicePointer(fakeSgIDs),
				AutoscalingEnabled:       helper.BoolPointer(false),
				AutoRepair:               helper.BoolPointer(false),
				SubnetID:                 helper.StringPointer(vpcOutput.ClusterPrivateSubnets[0]),
			}
			output, err := mpService.Apply(mpArgs)
			Expect(err).To(HaveOccurred())
			Expect(output).Should(ContainSubstring("is not attached to VPC"))

			By("Terraform plan with too many sg IDs cannot work")
			i := 0
			for i < 11 {
				fakeSgIDs = append(fakeSgIDs, fmt.Sprintf("sg-fakeid%d", i))
				i++
			}
			mpArgs.AdditionalSecurityGroups = helper.StringSlicePointer(fakeSgIDs)
			output, err = mpService.Plan(mpArgs)
			Expect(err).To(HaveOccurred())
			Expect(output).Should(
				MatchRegexp(`Attribute aws_node_pool.additional_security_group_ids list must contain at[\s\S]?most 10 elements, got: %d`, len(fakeSgIDs)))

		})

	It("can create/edit/delete labels/taints/autorepair - [id:72507]",
		ci.High, func() {
			By("Create machinepool")
			replicas := 3
			machineType := "m5.xlarge"
			name := helper.GenerateRandomName("np-72507", 2)
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			labels := map[string]string{
				"l1": "v1",
				"l2": "v2",
			}
			taint1 := map[string]string{"key": "t1", "value": "v1", "schedule_type": constants.NoSchedule}
			taints := []map[string]string{taint1}
			mpArgs := &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           helper.IntPointer(replicas),
				Name:               helper.StringPointer(name),
				SubnetID:           helper.StringPointer(subnetId),
				MachineType:        helper.StringPointer(machineType),
				AutoRepair:         helper.BoolPointer(false),
				Labels:             helper.StringMapPointer(labels),
				Taints:             &taints,
			}
			_, err := mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			respTaints := mpResponseBody.Taints()
			for index, taint := range respTaints {
				Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
				Expect(taint.Key()).To(Equal(taints[index]["key"]))
				Expect(taint.Value()).To(Equal(taints[index]["value"]))
			}
			Expect(mpResponseBody.AutoRepair()).To(BeFalse())
			Expect(mpResponseBody.Labels()).To(Equal(labels))

			By("Update labels/taints/autorepair")
			taints = append(taints, map[string]string{"key": "t2", "value": "", "schedule_type": constants.NoExecute})
			labels = map[string]string{
				"l3": "v3",
			}
			mpArgs.AutoRepair = helper.BoolPointer(true)
			mpArgs.Taints = &taints
			mpArgs.Labels = helper.StringMapPointer(labels)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			respTaints = mpResponseBody.Taints()
			for index, taint := range respTaints {
				Expect(taint.Effect()).To(Equal(taints[index]["schedule_type"]))
				Expect(taint.Key()).To(Equal(taints[index]["key"]))
				Expect(taint.Value()).To(Equal(taints[index]["value"]))
			}
			Expect(mpResponseBody.AutoRepair()).To(BeTrue())
			Expect(mpResponseBody.Labels()).To(Equal(labels))

			By("Remove labels/taints")
			mpArgs.Labels = nil
			mpArgs.Taints = nil
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Taints()).To(BeEmpty())
			Expect(mpResponseBody.AutoRepair()).To(BeTrue())
			Expect(mpResponseBody.Labels()).To(BeEmpty())
		})

	It("can be created with specific version - [id:72509]",
		ci.High, func() {
			replicas := 3
			machineType := "m5.2xlarge"
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			mpArgs := &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           helper.IntPointer(replicas),
				Name:               helper.StringPointer(helper.GenerateRandomName("np-72509", 2)),
				SubnetID:           helper.StringPointer(subnetId),
				MachineType:        helper.StringPointer(machineType),
				AutoRepair:         helper.BoolPointer(true),
			}

			By("Retrieve cluster version")
			clusterService, err := exec.NewClusterService(constants.GetClusterManifestsDir(profile.GetClusterType()))
			Expect(err).ToNot(HaveOccurred())
			clusterOutput, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterVersion := clusterOutput.ClusterVersion
			clusterSemVer, err := semver.NewVersion(clusterVersion)
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve z-1 version")
			zLowerVersions := cms.SortVersions(cms.GetHcpLowerVersions(ci.RHCSConnection, clusterVersion, profile.ChannelGroup))
			if len(zLowerVersions) > 0 {
				zversion := zLowerVersions[len(zLowerVersions)-1]
				zSemVer, err := semver.NewVersion(zversion.RawID)
				Expect(err).ToNot(HaveOccurred())

				if zSemVer.Major() == clusterSemVer.Major() && zSemVer.Minor() == clusterSemVer.Minor() {
					name := helper.GenerateRandomName("np-72509-z", 2)

					By("Create machinepool with z-1")
					mpArgs.Name = helper.StringPointer(name)
					mpArgs.OpenshiftVersion = helper.StringPointer(zversion.RawID)
					_, err = mpService.Apply(mpArgs)
					Expect(err).ToNot(HaveOccurred())

					By("Verify machinepool with z-1")
					mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
					Expect(err).ToNot(HaveOccurred())
					Expect(mpResponseBody.Version().ID()).To(Equal(zversion.ID))

					By("Destroy machinepool with z-1")
					_, err = mpService.Destroy()
					Expect(err).ToNot(HaveOccurred())
				} else {
					Logger.Infof("Cannot test `z-1` creation as the greatest lower `z-1` is: %s", zversion)
				}
			} else {
				Logger.Infof("Cannot test `z-1` creation as there is no version available")
			}

			By("Retrieve y-1 version")
			throttleVersion := fmt.Sprintf("%v.%v.0", clusterSemVer.Major(), clusterSemVer.Minor())
			yLowerVersions := cms.SortVersions(cms.GetHcpLowerVersions(ci.RHCSConnection, throttleVersion, profile.ChannelGroup))
			if len(yLowerVersions) > 0 {
				yVersion := yLowerVersions[len(yLowerVersions)-1]
				name := helper.GenerateRandomName("np-72509-z", 2)

				By("Create machinepool with y-1")
				mpArgs.Name = helper.StringPointer(name)
				mpArgs.OpenshiftVersion = helper.StringPointer(yVersion.RawID)
				_, err = mpService.Apply(mpArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify machinepool with z-1")
				mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Version().ID()).To(Equal(yVersion.ID))

				By("Destroy machinepool with z-1")
				_, err = mpService.Destroy()
				Expect(err).ToNot(HaveOccurred())
			}
		})

	It("can be created with tags - [id:72510]",
		ci.Critical, func() {
			By("Create machinepool")
			replicas := 3
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72510", 2)
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			tags := map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
			}
			mpArgs := &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           helper.IntPointer(replicas),
				Name:               helper.StringPointer(name),
				SubnetID:           helper.StringPointer(subnetId),
				MachineType:        helper.StringPointer(machineType),
				AutoRepair:         helper.BoolPointer(true),
				Tags:               helper.StringMapPointer(tags),
			}
			_, err := mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tags are correctly set")
			mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.AWSNodePool().Tags()).To(HaveKeyWithValue("aaa", "bbb"))
			Expect(mpResponseBody.AWSNodePool().Tags()).To(HaveKeyWithValue("ccc", "ddd"))
		})

	It("can be created with tuning configs - [id:72508]",
		ci.High, func() {
			var tuningconfigs []string
			var tcArgs *exec.TuningConfigArgs

			By("Create tuning configs")
			tcCount := 3
			tcName := "tc"
			tcArgs = &exec.TuningConfigArgs{
				Cluster:           helper.StringPointer(clusterID),
				Name:              helper.StringPointer(tcName),
				Count:             helper.IntPointer(tcCount),
				SpecVMDirtyRatios: helper.IntSlicePointer([]int{65, 65, 65}),
				SpecPriorities:    helper.IntSlicePointer([]int{10, 10, 10}),
			}
			_, err := tcService.Apply(tcArgs)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_, err = tcService.Destroy()
				Expect(err).ToNot(HaveOccurred())
			}()
			tcOut, err := tcService.Output()
			Expect(err).ToNot(HaveOccurred())
			createdTuningConfigs := tcOut.Names
			Logger.Infof("Retrieved tuning configs: %v", createdTuningConfigs)

			By("Create machinepool")
			replicas := 3
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72508", 2)
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			tuningconfigs = append(tuningconfigs, createdTuningConfigs...)
			mpArgs := &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           helper.IntPointer(replicas),
				Name:               helper.StringPointer(name),
				SubnetID:           helper.StringPointer(subnetId),
				MachineType:        helper.StringPointer(machineType),
				AutoRepair:         helper.BoolPointer(true),
				TuningConfigs:      helper.StringSlicePointer(tuningconfigs),
				Tags:               helper.StringMapPointer(constants.Tags),
			}
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs are correctly set")
			mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.TuningConfigs()).To(Equal(tuningconfigs))

			By("Edit tuning configs")
			tuningconfigs = []string{createdTuningConfigs[0]}
			mpArgs.TuningConfigs = helper.StringSlicePointer(tuningconfigs)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs are correctly updated")
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.TuningConfigs()).To(Equal(tuningconfigs))

			By("Remove tuning configs")
			mpArgs.TuningConfigs = helper.EmptyStringSlicePointer
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs are correctly updated")
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.TuningConfigs()).To(BeEmpty())
		})

	It("can be created with kubeletconfig - [id:74520]", ci.High,
		func() {
			By("Prepare two kubeletconfigs to hosted cluster")
			kubeArgs := &exec.KubeletConfigArgs{
				KubeLetConfigNumber: helper.IntPointer(2),
				NamePrefix:          helper.StringPointer("kube-74520"),
				PodPidsLimit:        helper.IntPointer(12345),
				Cluster:             helper.StringPointer(clusterID),
			}
			kubeService, err := exec.NewKubeletConfigService()
			Expect(err).ToNot(HaveOccurred())
			_, err = kubeService.Apply(kubeArgs)
			Expect(err).ToNot(HaveOccurred())
			kubeconfigs, err := kubeService.Output()
			Expect(err).ToNot(HaveOccurred())
			defer kubeService.Destroy()
			Expect(len(kubeconfigs)).To(Equal(2))

			By("Create machinepool with kubeletconfig")
			replicas := 1
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72504", 2)
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			mpArgs := &exec.MachinePoolArgs{
				Cluster:            &clusterID,
				AutoscalingEnabled: helper.BoolPointer(false),
				AutoRepair:         helper.BoolPointer(false),
				KubeletConfigs:     helper.StringPointer(kubeconfigs[0].Name),
				Replicas:           &replicas,
				Name:               &name,
				SubnetID:           &subnetId,
				MachineType:        &machineType,
			}
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			defer mpService.Destroy()

			By("Check the kueleteconfig is applied to the machinepool")
			mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.KubeletConfigs()[0]).To(Equal(kubeconfigs[0].Name))

			By("Update the kubeletconfig of the machinepool")
			mpArgs.KubeletConfigs = helper.StringPointer(kubeconfigs[1].Name)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Check the kubeleteconfig is applied to the machinepool")
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.KubeletConfigs()[0]).To(Equal(kubeconfigs[1].Name))

			By("Remove the kubeletconfig of the machinepool")
			mpArgs.KubeletConfigs = nil
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Check the kueleteconfig is applied to the machinepool")
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.KubeletConfigs()).To(BeEmpty())
		})

	Context("can validate", func() {
		getDefaultMPArgs := func(name string) *exec.MachinePoolArgs {
			replicas := 2
			machineType := "m5.2xlarge"
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			return &exec.MachinePoolArgs{
				Cluster:            helper.StringPointer(clusterID),
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           helper.IntPointer(replicas),
				Name:               helper.StringPointer(name),
				SubnetID:           helper.StringPointer(subnetId),
				MachineType:        helper.StringPointer(machineType),
				AutoRepair:         helper.BoolPointer(true),
			}
		}
		validateMPArgAgainstErrorSubstrings := func(mpName string, updateFields func(args *exec.MachinePoolArgs), errSubStrings ...string) {
			mpArgs := getDefaultMPArgs(mpName)
			updateFields(mpArgs)
			_, err := mpService.Apply(mpArgs)
			Expect(err).To(HaveOccurred())
			for _, errStr := range errSubStrings {
				helper.ExpectTFErrorContains(err, errStr)
			}
		}

		It("creation fields - [id:72514]", ci.Medium, func() {
			mpName := helper.GenerateRandomName("np-72514", 2)

			By("Retrieve current cluster information")
			clusterResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			By("Try to create a nodepool with empty cluster")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Cluster = helper.EmptyStringPointer
			}, "Attribute cluster cluster ID may not be empty/blank string, got: ")

			By("Try to create a nodepool with empty name")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Name = helper.EmptyStringPointer
			}, "Attribute name name may not be empty/blank string")

			By("Try to create a nodepool with wrong name")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Name = helper.StringPointer("any_wrong_$name")
			},
				"Expected a valid value",
				"'name' matching",
			)

			By("Try to create a nodepool with empty subnet_id")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.SubnetID = helper.EmptyStringPointer
			}, "Attribute subnet_id subnet ID may not be empty/blank string")

			By("Try to create a nodepool with wrong subnet")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.SubnetID = helper.StringPointer("subnet-0123456789")
			}, "The subnet ID 'subnet-0123456789' does not exist")

			By("Try to create a nodepool with autoscaling disabled and without replicas")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.AutoscalingEnabled = helper.BoolPointer(false)
				args.Replicas = nil
			}, "please provide a value for 'replicas' when 'autoscaling.enabled' is set to")

			By("Try to create a nodepool with replicas = -2")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Replicas = helper.IntPointer(-2)
			}, "must be a non-negative integer.")

			By("Try to create a nodepool with empty instance_type")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.MachineType = helper.EmptyStringPointer
			}, "'aws_node_pool.instance_type' cannot be empty.")

			By("Try to create a nodepool with version > CP version")
			currentVersion := clusterResp.Body().Version().RawID()
			currentSemVer, _ := semver.NewVersion(currentVersion)
			versions := cms.GetHcpHigherVersions(ci.RHCSConnection, currentVersion, profile.ChannelGroup)
			if len(versions) > 0 {
				validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
					args.OpenshiftVersion = helper.StringPointer(versions[0].RawID)
				}, "must not be greater than Control Plane version")
			} else {
				Logger.Info("No version > CP version found to test against")
			}

			By("Try to create a nodepool with version < CP version-2")
			throttleVersion := fmt.Sprintf("%v.%v.0", currentSemVer.Major(), currentSemVer.Minor()-2)
			versions = cms.GetHcpLowerVersions(ci.RHCSConnection, throttleVersion, profile.ChannelGroup)
			if len(versions) > 0 {
				validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
					args.OpenshiftVersion = helper.StringPointer(versions[0].RawID)
				}, "must be greater than the lowest supported version")
			} else {
				Logger.Info("No version < CP version - 2 found to test against")
			}

			By("Try to create a nodepool with not supported version")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.OpenshiftVersion = helper.StringPointer("4.8.0")
			}, "must be greater than the lowest supported version")

			By("Try to create a nodepool with wrong version")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.OpenshiftVersion = helper.StringPointer("any_version")
			}, "'openshift-vany_version-candidate' not found")

			By("Try to create a nodepool with autoscaling enabled and without min replicas")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.AutoscalingEnabled = helper.BoolPointer(true)
				args.Replicas = nil
				args.MaxReplicas = helper.IntPointer(3)
			},
				"These attributes must be configured together:",
				"[autoscaling.min_replicas,autoscaling.max_replicas]",
			)

			By("Try to create a nodepool with autoscaling enabled and without max replicas")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.AutoscalingEnabled = helper.BoolPointer(true)
				args.Replicas = nil
				args.MinReplicas = helper.IntPointer(1)
			},
				"These attributes must be configured together:",
				"[autoscaling.min_replicas,autoscaling.max_replicas]",
			)

			By("Try to create a nodepool with autoscaling enabled and without any replicas")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.AutoscalingEnabled = helper.BoolPointer(true)
				args.Replicas = nil
			}, "enabling autoscaling, should set value for maxReplicas")

			By("Try to create a nodepool with autoscaling enabled and min_replicas=0")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.AutoscalingEnabled = helper.BoolPointer(true)
				args.Replicas = nil
				args.MinReplicas = helper.IntPointer(0)
				args.MaxReplicas = helper.IntPointer(3)
			}, "'autoscaling.min_replica' must be greater than zero")

			By("Try to create a nodepool with both replicas and autoscaling enabled")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.AutoscalingEnabled = helper.BoolPointer(true)
				args.MinReplicas = helper.IntPointer(1)
				args.MaxReplicas = helper.IntPointer(3)
			},
				"These attributes cannot be configured together:",
				"[replicas,autoscaling.min_replicas]",
			)

			By("Try to create a nodepool with taint with no key, eg `=v1:NoSchedule`")
			taint1 := map[string]string{"key": "", "value": "v1", "schedule_type": constants.NoSchedule}
			taints := []map[string]string{taint1}
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Taints = &taints
			}, "taint key is mandatory")

			By("Try to create a nodepool with taint with wring scheduletype, eg `k1=v1:Wrong`")
			taint1 = map[string]string{"key": "k1", "value": "v1", "schedule_type": "Wrong"}
			taints = []map[string]string{taint1}
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Taints = &taints
			}, "Attribute taints[0].schedule_type value must be one of")

			By("Try to create a nodepool with system tags")
			newMapValue := map[string]string{
				"api.openshift.com/id": "any id",
			}

			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Tags = helper.StringMapPointer(newMapValue)
			}, "'aws_node_pool.tags' can not contain system tag 'api.openshift.com/id'")

			By("Try to create the nodepool wih not existing kubeletconfig")
			mpArgs := getDefaultMPArgs(mpName)
			mpArgs.KubeletConfigs = helper.StringPointer("notexisting")
			_, err = mpService.Apply(mpArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(`KubeletConfig with name[\\n\s]*'notexisting'[\\n\n\t\s]*does not exist for cluster`))

		})

		It("edit fields - [id:73431]", ci.Medium, func() {
			mpName := helper.GenerateRandomName("np-73431", 2)

			By("Create machinepool")
			mpArgs := getDefaultMPArgs(mpName)
			_, err := mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Try to edit cluster")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Cluster = helper.StringPointer("2a7il826aa41csgpiab2s1un856498ut")
			}, "Attribute cluster, cannot be changed from")

			By("Try to edit name")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Name = helper.StringPointer("anyName")
			}, "Attribute name, cannot be changed from")

			By("Try to edit subnet")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.SubnetID = helper.StringPointer("subnet-0a3fbd578b6af3e12")
			}, "Attribute aws_node_pool.subnet_id, cannot be changed from")

			By("Try to edit tags")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Tags = helper.StringMapPointer(map[string]string{
					"tag1": "value1",
				})
			}, "Attribute aws_node_pool.tags, cannot be changed from")

			By("Try to edit compute machine type")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.MachineType = helper.StringPointer("m5.xlarge")
			}, "Attribute aws_node_pool.instance_type, cannot be changed from")

			By("Try to update taint with no key, eg `=v1:NoSchedule`")
			taint1 := map[string]string{"key": "", "value": "v1", "schedule_type": constants.NoSchedule}
			taints := []map[string]string{taint1}
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Taints = &taints
			}, "key is mandatory")

			By("Try to update taint with wrong scheduletype, eg `k1=v1:Wrong`")
			taint1 = map[string]string{"key": "k1", "value": "v1", "schedule_type": "Wrong"}
			taints = []map[string]string{taint1}

			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Taints = &taints
			}, "Attribute taints[0].schedule_type value must be one of")

			By("Try to update the nodepool wih not existing kubeletconfig")
			mpArgs = getDefaultMPArgs(mpName)
			mpArgs.KubeletConfigs = helper.StringPointer("notexisting")
			_, err = mpService.Apply(mpArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(`KubeletConfig with name 'notexisting'[\n\t\s]*does not exist for cluster`))
		})
	})
})
