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

var _ = Describe("HCP MachinePool", ci.Day2, ci.NonClassicCluster, ci.FeatureMachinepool, func() {
	defer GinkgoRecover()
	var (
		mpService *exec.MachinePoolService
		tcService *exec.TuningConfigService
		mpArgs    *exec.MachinePoolArgs
		vpcOutput *exec.VPCOutput
	)

	BeforeEach(func() {
		mpService = exec.NewMachinePoolService(constants.HCPMachinePoolDir)
		tcService = exec.NewTuningConfigService(constants.TuningConfigDir)

		By("Get vpc output")
		vpcService := exec.NewVPCService()
		vpcOutput, err = vpcService.Output()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		mpService.Destroy()
		tcService.Destroy()
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
			mpArgs = &exec.MachinePoolArgs{
				Cluster:            clusterID,
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           &replicas,
				Name:               &name,
				SubnetID:           &subnetId,
				MachineType:        &machineType,
				AutoRepair:         helper.BoolPointer(true),
			}
			_, err = mpService.Apply(mpArgs, false)
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
			mpArgs = &exec.MachinePoolArgs{
				Cluster:            clusterID,
				AutoscalingEnabled: helper.BoolPointer(true),
				MinReplicas:        &minReplicas,
				MaxReplicas:        &maxReplicas,
				Name:               &name,
				SubnetID:           &subnetId,
				MachineType:        &machineType,
				AutoRepair:         helper.BoolPointer(true),
			}
			_, err := mpService.Apply(mpArgs, false)
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
			_, err = mpService.Apply(mpArgs, false)
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
			mpArgs.Replicas = &replicas
			_, err = mpService.Apply(mpArgs, false)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))

			By("Scale up")
			replicas = 4
			_, err = mpService.Apply(mpArgs, false)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))

			By("Scale to zero")
			replicas = 0
			_, err = mpService.Apply(mpArgs, false)
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
			mpArgs.MinReplicas = &minReplicas
			mpArgs.MaxReplicas = &maxReplicas
			_, err = mpService.Apply(mpArgs, false)
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
			sgService := exec.NewSecurityGroupService()
			output, err := sgService.Output()
			Expect(err).ToNot(HaveOccurred())
			if output.SGIDs == nil {

				sgArgs := &exec.SecurityGroupArgs{
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

			replicas := 0
			machineType := "r5.xlarge"
			name := "ocp-73068"
			sgIDs := output.SGIDs
			if len(sgIDs) >= 4 {
				sgIDs = sgIDs[0:4]
			}

			// workaround
			By("Check cluster tags and set it to the machinepool")
			resp, err := cms.ListNodePools(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			npDetail, err := cms.RetrieveNodePool(ci.RHCSConnection, clusterID, resp[0].ID(), map[string]interface{}{
				"fetchUserTagsOnly": true,
			})
			Expect(err).ToNot(HaveOccurred())
			workAroundTags := npDetail.AWSNodePool().Tags()

			mpArgs = &exec.MachinePoolArgs{
				Cluster:                  clusterID,
				Replicas:                 &replicas,
				MachineType:              &machineType,
				Name:                     &name,
				AdditionalSecurityGroups: &output.SGIDs,
				AutoscalingEnabled:       helper.BoolPointer(false),
				AutoRepair:               helper.BoolPointer(false),
				SubnetID:                 &vpcOutput.ClusterPrivateSubnets[0],
				Tags:                     &workAroundTags,
			}
			_, err = mpService.Apply(mpArgs, true)
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
			testAdditionalSecurityGroups := output.SGIDs[0:1]
			mpArgs = &exec.MachinePoolArgs{
				AdditionalSecurityGroups: &testAdditionalSecurityGroups,
			}

			applyOutput, err := mpService.Apply(mpArgs, false)
			Expect(err).To(HaveOccurred())
			Expect(applyOutput).Should(ContainSubstring("aws_node_pool.additional_security_group_ids, cannot be changed"))

			By("Destroy the machinepool")
			_, err = mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())

			By("Create another machinepool without additional sg ")
			name = "add-73068"
			mpArgs = &exec.MachinePoolArgs{
				Cluster:            clusterID,
				Replicas:           &replicas,
				MachineType:        helper.StringPointer("m5.2xlarge"),
				Name:               &name,
				Tags:               &workAroundTags,
				AutoscalingEnabled: helper.BoolPointer(false),
				AutoRepair:         helper.BoolPointer(false),
				SubnetID:           &vpcOutput.ClusterPrivateSubnets[0],
			}

			_, err = mpService.Apply(mpArgs, false)
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
			mpArgs = &exec.MachinePoolArgs{
				Cluster:                  clusterID,
				Replicas:                 &replicas,
				MachineType:              &machineType,
				Name:                     &name,
				AdditionalSecurityGroups: &fakeSgIDs,
				AutoscalingEnabled:       helper.BoolPointer(false),
				AutoRepair:               helper.BoolPointer(false),
				SubnetID:                 &vpcOutput.ClusterPrivateSubnets[0],
			}
			output, err := mpService.Apply(mpArgs, false)
			Expect(err).To(HaveOccurred())
			Expect(output).Should(ContainSubstring("is not attached to VPC"))

			By("Terraform plan with too many sg IDs cannot work")
			i := 0
			for i < 11 {
				fakeSgIDs = append(fakeSgIDs, fmt.Sprintf("sg-fakeid%d", i))
				i++
			}
			mpArgs = &exec.MachinePoolArgs{
				Cluster:                  clusterID,
				Replicas:                 &replicas,
				MachineType:              &machineType,
				Name:                     &name,
				AdditionalSecurityGroups: &fakeSgIDs,
				AutoscalingEnabled:       helper.BoolPointer(false),
				AutoRepair:               helper.BoolPointer(false),
				SubnetID:                 &vpcOutput.ClusterPrivateSubnets[0],
			}
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
			mpArgs = &exec.MachinePoolArgs{
				Cluster:            clusterID,
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           &replicas,
				Name:               &name,
				SubnetID:           &subnetId,
				MachineType:        &machineType,
				AutoRepair:         helper.BoolPointer(false),
				Labels:             &labels,
				Taints:             &taints,
			}
			_, err := mpService.Apply(mpArgs, false)
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
			mpArgs.AutoRepair = helper.BoolPointer(true)
			taint2 := map[string]string{"key": "t2", "value": "", "schedule_type": constants.NoExecute}
			taints = append(taints, taint2)
			labels = map[string]string{
				"l3": "v3",
			}
			mpArgs.Labels = &labels
			_, err = mpService.Apply(mpArgs, false)
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
			_, err = mpService.Apply(mpArgs, false)
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
			if profile.ChannelGroup != "" && profile.ChannelGroup != "stable" {
				Skip("Skip test due channel group not `stable` and https://issues.redhat.com/browse/OCM-7083")
			}

			replicas := 3
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72509", 2)
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			mpArgs = &exec.MachinePoolArgs{
				Cluster:            clusterID,
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           &replicas,
				Name:               &name,
				SubnetID:           &subnetId,
				MachineType:        &machineType,
				AutoRepair:         helper.BoolPointer(true),
			}

			By("Retrieve cluster version")
			clusterService, err = exec.NewClusterService(constants.GetClusterManifestsDir(profile.GetClusterType()))
			Expect(err).ToNot(HaveOccurred())
			clusterOutput, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterVersion := clusterOutput.ClusterVersion
			clusterSemVer, err := semver.NewVersion(clusterVersion)
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve z-1 version")
			zLowerVersions := cms.SortRawVersions(cms.GetHcpLowerVersions(ci.RHCSConnection, clusterVersion, exec.StableChannel)) // Should be changed once https://issues.redhat.com/browse/OCM-7083 is solved
			if len(zLowerVersions) > 0 {
				zversion := zLowerVersions[len(zLowerVersions)-1]
				zSemVer, err := semver.NewVersion(zversion)
				Expect(err).ToNot(HaveOccurred())

				if zSemVer.Major() == clusterSemVer.Major() && zSemVer.Minor() == clusterSemVer.Minor() {
					By("Create machinepool with z-1")
					mpArgs.OpenshiftVersion = &zversion
					_, err = mpService.Apply(mpArgs, true)
					Expect(err).ToNot(HaveOccurred())

					By("Verify machinepool with z-1")
					mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
					Expect(err).ToNot(HaveOccurred())
					Expect(mpResponseBody.Version()).To(Equal(zversion))

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
			yLowerVersions := cms.SortRawVersions(cms.GetHcpLowerVersions(ci.RHCSConnection, throttleVersion, exec.StableChannel)) // Channel should be changed once https://issues.redhat.com/browse/OCM-7083 is solved
			if len(yLowerVersions) > 0 {
				yVersion := yLowerVersions[len(zLowerVersions)-1]

				By("Create machinepool with y-1")
				mpArgs.OpenshiftVersion = &yVersion
				_, err = mpService.Apply(mpArgs, true)
				Expect(err).ToNot(HaveOccurred())

				By("Verify machinepool with z-1")
				mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Version()).To(Equal(yVersion))

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
			mpArgs = &exec.MachinePoolArgs{
				Cluster:            clusterID,
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           &replicas,
				Name:               &name,
				SubnetID:           &subnetId,
				MachineType:        &machineType,
				AutoRepair:         helper.BoolPointer(true),
				Tags:               &tags,
			}
			_, err = mpService.Apply(mpArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tags are correctly set")
			mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.AWSNodePool().Tags()).To(HaveKeyWithValue("aaa", "bbb"))
			Expect(mpResponseBody.AWSNodePool().Tags()).To(HaveKeyWithValue("ccc", "ddd"))

			// Workaround to make sure we have the correct tags
			tags = mpResponseBody.AWSNodePool().Tags()

			By("Edit machinepool tags")
			delete(tags, "aaa")
			tags["ccc"] = "fff"
			_, err = mpService.Apply(mpArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tags are correctly updated")
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.AWSNodePool().Tags()).ToNot(HaveKey("aaa"))
			Expect(mpResponseBody.AWSNodePool().Tags()).To(HaveKeyWithValue("ccc", "fff"))
		})

	It("can be created with tuning configs - [id:72508]",
		ci.High, func() {
			var tuningconfigs []string
			var tcArgs *exec.TuningConfigArgs

			By("Create tuning configs")
			tcCount := 3
			tcArgs = &exec.TuningConfigArgs{
				Cluster:           clusterID,
				NamePrefix:        "tc",
				Count:             &tcCount,
				SpecVMDirtyRatios: &[]int{65, 65, 65},
				SpecPriorities:    &[]int{10, 10, 10},
			}
			_, err := tcService.Apply(tcArgs, false)
			Expect(err).ToNot(HaveOccurred())
			tcOut, err := tcService.Output()
			Expect(err).ToNot(HaveOccurred())
			createdTuningConfigs := tcOut.Names
			Logger.Infof("Retrieved tuning configs: %v", createdTuningConfigs)

			By("Create machinepool")
			replicas := 3
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72504", 2)
			subnetId := vpcOutput.ClusterPrivateSubnets[0]
			tuningconfigs = append(tuningconfigs, createdTuningConfigs...)
			mpArgs = &exec.MachinePoolArgs{
				Cluster:            clusterID,
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           &replicas,
				Name:               &name,
				SubnetID:           &subnetId,
				MachineType:        &machineType,
				AutoRepair:         helper.BoolPointer(true),
				TuningConfigs:      &tuningconfigs,
				Tags:               &constants.Tags,
			}
			_, err = mpService.Apply(mpArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs are correctly set")
			mpResponseBody, err := cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.TuningConfigs()).To(Equal(tuningconfigs))

			By("Edit tuning configs")
			tuningconfigs = []string{createdTuningConfigs[0]}
			mpArgs.TuningConfigs = &tuningconfigs
			_, err = mpService.Apply(mpArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs are correctly updated")
			mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.TuningConfigs()).To(Equal(tuningconfigs))

			// Disabled as blocked by https://issues.redhat.com/browse/OCM-7347
			// By("Remove tuning configs")
			// mpArgs.TuningConfigs = &[]string{}
			// _, err = mpService.Apply(mpArgs, false)
			// Expect(err).ToNot(HaveOccurred())

			// By("Verify tuning configs are correctly updated")
			// mpResponseBody, err = cms.RetrieveClusterNodePool(ci.RHCSConnection, clusterID, name)
			// Expect(err).ToNot(HaveOccurred())
			// Expect(mpResponseBody.TuningConfigs()).To(BeEmpty())
		})
})
