package e2e

import (

	// nolint

	"context"
	"encoding/json"
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
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/profilehandler"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("HCP MachinePool", ci.Day2, ci.FeatureMachinepool, func() {
	defer GinkgoRecover()
	var (
		mpService      exec.MachinePoolService
		vpcOutput      *exec.VPCOutput
		profileHandler profilehandler.ProfileHandler
	)

	BeforeEach(func() {
		var err error
		profileHandler, err = profilehandler.NewProfileHandlerFromYamlFile()
		Expect(err).ToNot(HaveOccurred())

		if !profileHandler.Profile().IsHCP() {
			Skip("Test can run only on Hosted cluster")
		}

		mpService, err = profileHandler.Services().GetMachinePoolsService()
		Expect(err).ToNot(HaveOccurred())

		By("Get vpc output")
		vpcService, err := profileHandler.Services().GetVPCService()
		Expect(err).ToNot(HaveOccurred())
		vpcOutput, err = vpcService.Output()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		mpService.Destroy()
	})

	getDefaultMPArgs := func(name string) *exec.MachinePoolArgs {
		replicas := 2
		machineType := "m5.2xlarge"
		subnetId := vpcOutput.PrivateSubnets[0]
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

	It("can be created with only required attributes - [id:72504]",
		ci.Critical, func() {
			By("Retrieve current cluster information")
			clusterRespBody, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())

			var initialMinReplicas int
			var initialMaxReplicas int
			var initialReplicas int
			if profileHandler.Profile().IsAutoscale() {
				initialMinReplicas = clusterRespBody.Body().Nodes().AutoscaleCompute().MinReplicas()
				initialMaxReplicas = clusterRespBody.Body().Nodes().AutoscaleCompute().MaxReplicas()
			} else {
				initialReplicas = clusterRespBody.Body().Nodes().Compute()
			}

			By("Create machinepool")
			replicas := 3
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72504", 2)
			subnetId := vpcOutput.PrivateSubnets[0]
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
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.ID()).To(Equal(name))
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))
			Expect(mpResponseBody.Subnet()).To(Equal(subnetId))
			Expect(mpResponseBody.AWSNodePool().InstanceType()).To(Equal(machineType))
			Expect(mpResponseBody.AutoRepair()).To(BeTrue())

			By("Wait for machinepool replicas available")
			err = wait.PollUntilContextTimeout(context.Background(), 30*time.Second, 20*time.Minute, false, func(context.Context) (bool, error) {
				clusterRespBody, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
				if err != nil {
					return false, err
				}
				if profileHandler.Profile().IsAutoscale() {
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
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).To(HaveOccurred())
		})

	It("can create/edit autoscaling - [id:72505]",
		ci.Critical, func() {
			minReplicas := 2
			maxReplicas := 4
			replicas := 3
			machineType := "m5.xlarge"
			name := helper.GenerateRandomName("np-72505", 2)
			subnetId := vpcOutput.PrivateSubnets[0]

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
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
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
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
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
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))

			By("Scale up")
			replicas = 4
			mpArgs.Replicas = helper.IntPointer(replicas)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).To(BeNil())
			Expect(mpResponseBody.Replicas()).To(Equal(replicas))

			By("Scale to zero")
			replicas = 0
			mpArgs.Replicas = helper.IntPointer(replicas)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			// Verify
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
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
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Autoscaling()).ToNot(BeNil())
			Expect(mpResponseBody.Autoscaling().MaxReplica()).To(Equal(maxReplicas))
			Expect(mpResponseBody.Autoscaling().MinReplica()).To(Equal(minReplicas))
		})

	It("can be created with security groups - [id:73068]", ci.High,
		func() {
			By("Prepare additional security groups")
			sgService, err := profileHandler.Services().GetSecurityGroupService()
			output, err := sgService.Output()
			Expect(err).ToNot(HaveOccurred())
			if output.SGIDs == nil {
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
				AdditionalSecurityGroups: helper.StringSlicePointer(sgIDs),
				AutoscalingEnabled:       helper.BoolPointer(false),
				AutoRepair:               helper.BoolPointer(false),
				SubnetID:                 helper.StringPointer(vpcOutput.PrivateSubnets[0]),
			}
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_, err = mpService.Destroy()
				Expect(err).ToNot(HaveOccurred())
			}()

			By("Verify the parameters of the created machinepool")
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
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
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
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
				SubnetID:                 helper.StringPointer(vpcOutput.PrivateSubnets[0]),
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
			subnetId := vpcOutput.PrivateSubnets[0]
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
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
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
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
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
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Taints()).To(BeEmpty())
			Expect(mpResponseBody.AutoRepair()).To(BeTrue())
			Expect(mpResponseBody.Labels()).To(BeEmpty())
		})

	It("can be created with specific version - [id:72509]",
		ci.High, func() {
			replicas := 3
			machineType := "m5.2xlarge"
			subnetId := vpcOutput.PrivateSubnets[0]
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
			clusterService, err := profileHandler.Services().GetClusterService()
			Expect(err).ToNot(HaveOccurred())
			clusterOutput, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterVersion := clusterOutput.ClusterVersion
			clusterSemVer, err := semver.NewVersion(clusterVersion)
			Expect(err).ToNot(HaveOccurred())

			By("Retrieve z-1 version")
			zLowerVersions := cms.SortVersions(cms.GetHcpLowerVersions(cms.RHCSConnection, clusterVersion, profileHandler.Profile().GetChannelGroup()))
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
					mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
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
			yLowerVersions := cms.SortVersions(cms.GetHcpLowerVersions(cms.RHCSConnection, throttleVersion, profileHandler.Profile().GetChannelGroup()))
			if len(yLowerVersions) > 0 {
				yVersion := yLowerVersions[len(yLowerVersions)-1]
				name := helper.GenerateRandomName("np-72509-z", 2)

				By("Create machinepool with y-1")
				mpArgs.Name = helper.StringPointer(name)
				mpArgs.OpenshiftVersion = helper.StringPointer(yVersion.RawID)
				_, err = mpService.Apply(mpArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify machinepool with y-1")
				mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(mpResponseBody.Version().ID()).To(Equal(yVersion.ID))

				By("Destroy machinepool with y-1")
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
			subnetId := vpcOutput.PrivateSubnets[0]
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
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.AWSNodePool().Tags()).To(HaveKeyWithValue("aaa", "bbb"))
			Expect(mpResponseBody.AWSNodePool().Tags()).To(HaveKeyWithValue("ccc", "ddd"))

			By("Remove machinepool")
			_, err = mpService.Destroy()
			Expect(err).ToNot(HaveOccurred())

			By("Create machinepool with empty tags")
			name = helper.GenerateRandomName("np-72510", 2)
			tags = map[string]string{}
			mpArgs.Name = helper.StringPointer(name)
			mpArgs.Tags = helper.StringMapPointer(tags)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tags are correctly set")
			output, err := mpService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(output.MachinePools[0].Tags)).To(Equal(0))
		})

	It("can be created with tuning configs - [id:72508]",
		ci.High, func() {
			var tuningconfigs []string
			var tcArgs *exec.TuningConfigArgs

			tcService, err := profileHandler.Services().GetTuningConfigService()
			Expect(err).ToNot(HaveOccurred())

			By("Create tuning configs")
			tcCount := 3
			tcName := "tc"
			var specs []exec.TuningConfigSpec
			for i := 0; i < tcCount; i++ {
				spec := helper.NewTuningConfigSpecRootStub(fmt.Sprintf("%s-%d", tcName, i), 65, 10)
				tc, err := json.Marshal(spec)
				Expect(err).ToNot(HaveOccurred())
				specs = append(specs, exec.NewTuningConfigSpecFromString(string(tc)))
			}
			tcArgs = &exec.TuningConfigArgs{
				Cluster: helper.StringPointer(clusterID),
				Name:    helper.StringPointer(tcName),
				Count:   helper.IntPointer(tcCount),
				Specs:   &specs,
			}
			_, err = tcService.Apply(tcArgs)
			Expect(err).ToNot(HaveOccurred())
			defer tcService.Destroy()
			tcOut, err := tcService.Output()
			Expect(err).ToNot(HaveOccurred())
			createdTuningConfigs := tcOut.Names
			Logger.Infof("Retrieved tuning configs: %v", createdTuningConfigs)

			By("Create machinepool")
			replicas := 3
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72508", 2)
			subnetId := vpcOutput.PrivateSubnets[0]
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
				Tags:               helper.StringMapPointer(profilehandler.Tags),
			}
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs are correctly set")
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.TuningConfigs()).To(Equal(tuningconfigs))

			By("Edit tuning configs")
			tuningconfigs = []string{createdTuningConfigs[0]}
			mpArgs.TuningConfigs = helper.StringSlicePointer(tuningconfigs)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs are correctly updated")
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.TuningConfigs()).To(Equal(tuningconfigs))

			By("Remove tuning configs")
			mpArgs.TuningConfigs = helper.EmptyStringSlicePointer
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs are correctly updated")
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.TuningConfigs()).To(BeEmpty())
		})

	It("can be created with kubeletconfig - [id:74520]", ci.High,
		func() {
			By("Prepare two kubeletconfigs to hosted cluster")
			kubeArgs := &exec.KubeletConfigArgs{
				KubeletConfigNumber: helper.IntPointer(2),
				NamePrefix:          helper.StringPointer("kube-74520"),
				PodPidsLimit:        helper.IntPointer(12345),
				Cluster:             helper.StringPointer(clusterID),
			}
			kubeService, err := profileHandler.Services().GetKubeletConfigService()
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
			subnetId := vpcOutput.PrivateSubnets[0]
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

			By("Check the kubeleteconfig is applied to the machinepool")
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.KubeletConfigs()[0]).To(Equal(kubeconfigs[0].Name))

			By("Update the kubeletconfig of the machinepool")
			mpArgs.KubeletConfigs = helper.StringPointer(kubeconfigs[1].Name)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Check the kubeleteconfig is applied to the machinepool")
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.KubeletConfigs()[0]).To(Equal(kubeconfigs[1].Name))

			By("Remove the kubeletconfig of the machinepool")
			mpArgs.KubeletConfigs = nil
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Check the kubeleteconfig is applied to the machinepool")
			mpResponseBody, err = cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.KubeletConfigs()).To(BeEmpty())
		})

	It("can be created with imdsv2 - [id:75373]", ci.Critical,
		func() {
			imdsv2Values := []string{constants.OptionalEc2MetadataHttpTokens,
				constants.RequiredEc2MetadataHttpTokens}

			replicas := 3
			machineType := "m5.xlarge"
			subnetId := vpcOutput.PrivateSubnets[0]

			for _, imdsv2Value := range imdsv2Values {
				By("Create a machinepool with --ec2-metadata-http-tokens = " + imdsv2Value)
				name := helper.GenerateRandomName("np-75373", 2)
				mpArgs := &exec.MachinePoolArgs{
					Cluster:               helper.StringPointer(clusterID),
					AutoscalingEnabled:    helper.BoolPointer(false),
					Replicas:              helper.IntPointer(replicas),
					Name:                  helper.StringPointer(name),
					SubnetID:              helper.StringPointer(subnetId),
					MachineType:           helper.StringPointer(machineType),
					AutoRepair:            helper.BoolPointer(true),
					Ec2MetadataHttpTokens: helper.StringPointer(imdsv2Value),
				}
				_, err := mpService.Apply(mpArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Verify imdsv2 value is correctly set")
				mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(mpResponseBody.AWSNodePool().Ec2MetadataHttpTokens())).
					To(Equal(imdsv2Value))

				mpService.Destroy()
			}
		})

	It("can create multiple instances - [id:72954]",
		ci.Low, func() {
			By("Create machinepool")
			mpCount := 2
			replicas := 3
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("np-72954", 2)
			subnetId := vpcOutput.PrivateSubnets[0]
			mpArgs := &exec.MachinePoolArgs{
				Count:              helper.IntPointer(mpCount),
				Cluster:            helper.StringPointer(clusterID),
				AutoscalingEnabled: helper.BoolPointer(false),
				Replicas:           helper.IntPointer(replicas),
				Name:               helper.StringPointer(name),
				SubnetID:           helper.StringPointer(subnetId),
				MachineType:        helper.StringPointer(machineType),
				AutoRepair:         helper.BoolPointer(true),
			}
			_, err := mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Verify mps have been created")
			mpsOut, err := mpService.Output()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(mpsOut.MachinePools)).To(Equal(2))
			var expectedNames []string
			for i := 0; i < mpCount; i++ {
				expectedNames = append(expectedNames, fmt.Sprintf("%s-%v", name, i))
			}
			for _, mp := range mpsOut.MachinePools {
				Expect(mp.Name).To(BeElementOf(expectedNames))
				Expect(mp.ClusterID).To(BeElementOf(clusterID))
				Expect(mp.Replicas).To(BeElementOf(replicas))
				Expect(mp.MachineType).To(BeElementOf(machineType))
			}

		})

	Context("can validate", func() {
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
			clusterResp, err := cms.RetrieveClusterDetail(cms.RHCSConnection, clusterID)
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
			versions := cms.GetHcpHigherVersions(cms.RHCSConnection, currentVersion, profileHandler.Profile().GetChannelGroup())
			if len(versions) > 0 {
				validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
					args.OpenshiftVersion = helper.StringPointer(versions[0].RawID)
				}, "must not be greater than Control Plane version")
			} else {
				Logger.Info("No version > CP version found to test against")
			}

			By("Try to create a nodepool with version < CP version-2")
			throttleVersion := fmt.Sprintf("%v.%v.0", currentSemVer.Major(), currentSemVer.Minor()-2)
			versions = cms.GetHcpLowerVersions(cms.RHCSConnection, throttleVersion, profileHandler.Profile().GetChannelGroup())
			if len(versions) > 0 {
				validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
					args.OpenshiftVersion = helper.StringPointer(versions[0].RawID)
				}, "must be no less than 2 minor versions behind the Control Plane version")
			} else {
				Logger.Info("No version < CP version - 2 found to test against")
			}

			By("Try to create a nodepool with not supported version")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.OpenshiftVersion = helper.StringPointer("4.10.67")
			}, "must be greater than the lowest supported version")

			By("Try to create a nodepool with wrong version")
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.OpenshiftVersion = helper.StringPointer("any_version")
			}, "'openshift-vany_version' not found")

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

		It("imdsv2 fields - [id:75393]", ci.Medium, func() {
			By("Try to create a nodepool with invalid imdsv2")
			mpName := helper.GenerateRandomName("np-75393", 2)
			validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
				args.Ec2MetadataHttpTokens = helper.StringPointer("invalid")
			}, "Expected a valid param. Options are [optional required]. Got invalid.")

			reqBody := []map[string]string{
				{
					"create_value": "required",
					"edit_value":   "optional",
				},
				{
					"create_value": "optional",
					"edit_value":   "required",
				},
			}

			for _, imdsv2Value := range reqBody {
				By("Create a machinepool with --ec2-metadata-http-tokens = " + imdsv2Value["create_value"])
				mpName := helper.GenerateRandomName("np-75393", 2)
				mpArgs := getDefaultMPArgs(mpName)
				mpArgs.Ec2MetadataHttpTokens = helper.StringPointer(imdsv2Value["create_value"])
				_, err := mpService.Apply(mpArgs)
				Expect(err).ToNot(HaveOccurred())

				By("Try to edit nodepool with ec2_metadata_http_tokens = " + imdsv2Value["edit_value"])
				errMsg := fmt.Sprintf(
					`Attribute aws_node_pool.ec2_metadata_http_tokens, cannot be changed from "%s" to "%s"`,
					imdsv2Value["create_value"], imdsv2Value["edit_value"])
				validateMPArgAgainstErrorSubstrings(mpName, func(args *exec.MachinePoolArgs) {
					args.Ec2MetadataHttpTokens = helper.StringPointer(imdsv2Value["edit_value"])
				}, errMsg)

				mpService.Destroy()
			}
		})
	})

	It("can import - [id:72960]", ci.Day2, ci.Medium, ci.FeatureImport,
		func() {
			importService, err := profileHandler.Services().GetImportService()
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				By("Destroy import service")
				importService.Destroy()
			}()

			By("Create additional machinepool for import")
			replicas := 2
			machineType := "m5.2xlarge"
			name := helper.GenerateRandomName("ocp-72960", 2)
			subnetId := vpcOutput.PrivateSubnets[0]
			tags := map[string]string{"foo1": "bar1"}
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

			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Run the command to import the machinepool")
			importParam := &exec.ImportArgs{
				ClusterID:  clusterID,
				Resource:   "rhcs_hcp_machine_pool.mp_import",
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

	Context("can upgrade", ci.Day2, ci.FeatureImport, func() {

		It("from z-1 - [id:72513]", ci.High, func() {
			name := helper.GenerateRandomName("np-72513", 2)
			mpArgs := getDefaultMPArgs(name)

			By("Retrieve cluster version")
			clusterService, err := profileHandler.Services().GetClusterService()
			Expect(err).ToNot(HaveOccurred())
			clusterOut, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterVersion := clusterOut.ClusterVersion
			Expect(err).ToNot(HaveOccurred())

			By("Get version z-1")
			zLowerVersions, err := cms.GetVersionsWithUpgradesToVersion(cms.RHCSConnection, clusterVersion, profileHandler.Profile().GetChannelGroup(), constants.Z, true, true, 1)
			Logger.Infof("Got versions %v", zLowerVersions)
			Expect(err).ToNot(HaveOccurred())
			if len(zLowerVersions) <= 0 {
				Skip("No Available version for upgrading on z-stream")
			}
			zversion := zLowerVersions[len(zLowerVersions)-1]

			By("Create machinepool with z-1")
			mpArgs.OpenshiftVersion = helper.StringPointer(zversion.RawID)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Check created machine pool")
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Version().ID()).To(Equal(zversion.ID))

			By("Upgrade the machinepool to cluster version")
			mpArgs.OpenshiftVersion = helper.StringPointer(clusterVersion)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Check upgrade policy")
			npUpPolicies, err := cms.ListNodePoolUpgradePolicies(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(npUpPolicies.Items().Slice())).To(Equal(1))
			Expect(npUpPolicies.Items().Get(0).Version()).To(Equal(clusterVersion))
		})

		It("from y-1 - [id:72512]", ci.High, func() {
			name := helper.GenerateRandomName("np-72512", 2)
			mpArgs := getDefaultMPArgs(name)

			By("Retrieve cluster version")
			clusterService, err := profileHandler.Services().GetClusterService()
			Expect(err).ToNot(HaveOccurred())
			clusterOut, err := clusterService.Output()
			Expect(err).ToNot(HaveOccurred())
			clusterVersion := clusterOut.ClusterVersion
			Expect(err).ToNot(HaveOccurred())

			By("Get version z-1")
			yLowerVersions, err := cms.GetVersionsWithUpgradesToVersion(cms.RHCSConnection, clusterVersion, profileHandler.Profile().GetChannelGroup(), constants.Y, true, true, 1)
			Expect(err).ToNot(HaveOccurred())
			if len(yLowerVersions) <= 0 {
				Skip("No Available version for upgrading on y-stream")
			}
			yVersion := yLowerVersions[len(yLowerVersions)-1]

			By("Create machinepool with y-1")
			mpArgs.OpenshiftVersion = helper.StringPointer(yVersion.RawID)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Check created machine pool")
			mpResponseBody, err := cms.RetrieveClusterNodePool(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(mpResponseBody.Version().ID()).To(Equal(yVersion.ID))

			By("Upgrade the machinepool to cluster version")
			mpArgs.OpenshiftVersion = helper.StringPointer(clusterVersion)
			_, err = mpService.Apply(mpArgs)
			Expect(err).ToNot(HaveOccurred())

			By("Check upgrade policy")
			npUpPolicies, err := cms.ListNodePoolUpgradePolicies(cms.RHCSConnection, clusterID, name)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(npUpPolicies.Items().Slice())).To(Equal(1))
			Expect(npUpPolicies.Items().Get(0).Version()).To(Equal(clusterVersion))
		})
	})
})
