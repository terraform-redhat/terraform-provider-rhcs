package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

var _ = Describe("Tuning Config", ci.NonClassicCluster, ci.FeatureTuningConfig, ci.Day2, func() {
	var (
		tcService *exec.TuningConfigService
		tcArgs    *exec.TuningConfigArgs

		mpService *exec.MachinePoolService
	)

	verifyTuningConfigSpec := func(spec interface{}, specVmDirtyRatio, specPriority int) {
		tcSpec := spec.(map[string]interface{})
		tcProfileSpec := (tcSpec["profile"].([]interface{}))[0].(map[string]interface{})
		tcRecommendSpec := (tcSpec["recommend"].([]interface{}))[0].(map[string]interface{})
		Expect(tcProfileSpec["data"]).To(ContainSubstring(fmt.Sprintf("vm.dirty_ratio=\"%d\"", specVmDirtyRatio)))
		Expect(tcRecommendSpec["priority"]).To(BeEquivalentTo(specPriority))
	}

	BeforeEach(func() {
		mpService = exec.NewMachinePoolService(constants.HCPMachinePoolDir)
		tcService = exec.NewTuningConfigService(constants.TuningConfigDir)
	})

	AfterEach(func() {
		mpService.Destroy()
		tcService.Destroy()
	})

	It("can create/edit/delete - [id:72521]", ci.High, func() {
		name := "tc-72521"
		tcCount := 1
		By("Create one tuning config")
		tcArgs = &exec.TuningConfigArgs{
			Cluster: &clusterID,
			Name:    &name,
			Count:   &tcCount,
		}
		_, err := tcService.Apply(tcArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Verify tuning config")
		tcsResp, err := cms.ListTuningConfigs(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(tcsResp.Size()).To(Equal(tcCount))
		tc := tcsResp.Items().Get(0)
		Expect(tc.Name()).To(Equal(name))
		Expect(tc.Spec()).ToNot(BeEmpty())

		By("Delete created tuning config")
		_, err = tcService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Create many tuning configs")
		tcCount = 2
		specVMDirtyRatios := []int{65, 45}
		specPriorities := []int{20, 10}
		tcArgs = &exec.TuningConfigArgs{
			Cluster:           &clusterID,
			Name:              &name,
			Count:             &tcCount,
			SpecVMDirtyRatios: &specVMDirtyRatios,
			SpecPriorities:    &specPriorities,
		}
		_, err = tcService.Apply(tcArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Verify tuning configs")
		tcsResp, err = cms.ListTuningConfigs(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(tcsResp.Size()).To(Equal(tcCount))
		for index, tc := range tcsResp.Items().Slice() {
			Expect(tc.Name()).To(Equal(fmt.Sprintf("%s-%v", name, index)))
			verifyTuningConfigSpec(tc.Spec(), specVMDirtyRatios[index], specPriorities[index])
		}

		By("Update first tuning config")
		specVMDirtyRatios[0] = 55
		specPriorities[0] = 1
		_, err = tcService.Apply(tcArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Verify tuning configs")
		tcsResp, err = cms.ListTuningConfigs(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(tcsResp.Size()).To(Equal(tcCount))
		for index, tc := range tcsResp.Items().Slice() {
			Expect(tc.Name()).To(Equal(fmt.Sprintf("%s-%v", name, index)))
			verifyTuningConfigSpec(tc.Spec(), specVMDirtyRatios[index], specPriorities[index])
		}

		By("Delete all")
		_, err = tcService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})

	It("can validate - [id:72522]", ci.Medium, func() {
		getDefaultTCArgs := func() *exec.TuningConfigArgs {
			name := "tc-72522"
			return &exec.TuningConfigArgs{
				Cluster: &clusterID,
				Name:    &name,
			}
		}

		By("Try to create tuning config with empty cluster ID")
		tcArgs = getDefaultTCArgs()
		tcArgs.Cluster = &constants.EmptyStringValue
		_, err := tcService.Apply(tcArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Attribute cluster cluster ID may not be empty/blank string, got:"))

		By("Try to create tuning config with empty name")
		tcArgs = getDefaultTCArgs()
		tcArgs.Name = &constants.EmptyStringValue
		_, err = tcService.Apply(tcArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("The name must be a lowercase RFC 1123 subdomain"))

		By("Try to create tuning config with empty spec")
		tcArgs = getDefaultTCArgs()
		tcArgs.Spec = &constants.EmptyStringValue
		_, err = tcService.Apply(tcArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(" Attribute 'spec' must\nbe set"))

		By("Create tuning config for edit")
		tcArgs = getDefaultTCArgs()
		_, err = tcService.Apply(tcArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Try to edit cluster with other cluster ID")
		clustersResp, err := cms.ListClusters(ci.RHCSConnection)
		Expect(err).ToNot(HaveOccurred())
		var otherClusterID string
		for _, cluster := range clustersResp.Items().Slice() {
			if cluster.ID() != clusterID {
				otherClusterID = cluster.ID()
				break
			}
		}
		if otherClusterID != "" {
			tcArgs = getDefaultTCArgs()
			tcArgs.Cluster = &otherClusterID
			_, err = tcService.Apply(tcArgs, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Attribute cluster, cannot be changed from"))
		} else {
			Logger.Info("No other cluster accessible for testing this change")
		}

		By("Try to edit cluster field with wrong value")
		value := "wrong"
		tcArgs = getDefaultTCArgs()
		tcArgs.Cluster = &value
		_, err = tcService.Apply(tcArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Attribute cluster, cannot be changed from"))

		By("Try to edit name field")
		value = "new_name"
		tcArgs = getDefaultTCArgs()
		tcArgs.Name = &value
		_, err = tcService.Apply(tcArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Attribute name, cannot be changed from"))

		By("Try to edit spec field with non json value")
		value = "wrong"
		tcArgs = getDefaultTCArgs()
		tcArgs.Spec = &value
		_, err = tcService.Apply(tcArgs, false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cannot unmarshal string"))

		By("Get vpc output")
		vpcService := exec.NewVPCService()
		vpcOutput, err := vpcService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Create machinepool with tuning config")
		replicas := 3
		machineType := "m5.2xlarge"
		name := helper.GenerateRandomName("np-72522", 2)
		subnetId := vpcOutput.ClusterPrivateSubnets[0]
		tuningConfigs := []string{*tcArgs.Name}
		mpArgs := &exec.MachinePoolArgs{
			Cluster:            &clusterID,
			AutoscalingEnabled: helper.BoolPointer(false),
			Replicas:           &replicas,
			Name:               &name,
			SubnetID:           &subnetId,
			MachineType:        &machineType,
			AutoRepair:         helper.BoolPointer(true),
			TuningConfigs:      &tuningConfigs,
		}
		_, err = mpService.Apply(mpArgs, false)
		Expect(err).ToNot(HaveOccurred())

		By("Delete Tuning config used by machinepool")
		_, err = tcService.Destroy()
		Expect(err).To(HaveOccurred())
	})
})
