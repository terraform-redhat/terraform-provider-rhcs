package e2e

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Tuning Config", ci.FeatureTuningConfig, ci.Day2, func() {
	var (
		tcService exec.TuningConfigService
		mpService exec.MachinePoolService
		profile   *ci.Profile
	)

	verifyTuningConfigSpec := func(spec interface{}, profileName string, specVmDirtyRatio, specPriority int) {
		Expect(spec).ToNot(BeEmpty())
		tcSpec := spec.(map[string]interface{})
		tcProfileSpec := (tcSpec["profile"].([]interface{}))[0].(map[string]interface{})
		tcRecommendSpec := (tcSpec["recommend"].([]interface{}))[0].(map[string]interface{})
		Expect(tcProfileSpec["name"]).To(ContainSubstring(profileName))
		Expect(tcProfileSpec["data"]).To(ContainSubstring(fmt.Sprintf("vm.dirty_ratio=\"%d\"", specVmDirtyRatio)))
		Expect(tcRecommendSpec["priority"]).To(BeEquivalentTo(specPriority))
	}

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
	})

	AfterEach(func() {
		mpService.Destroy()
		tcService.Destroy()
	})

	It("can create/edit/delete - [id:72521]", ci.High, func() {
		name := "tc-72521"
		tcCount := 1
		tc1Name := helper.GenerateRandomName("tuned01", 2)
		firstPriority := 10
		firstVMDirtyRatio := 25
		tc2Name := helper.GenerateRandomName("tuned02", 2)
		secondPriority := 20
		secondVMDirtyRatio := 65

		By("Create one tuning config")
		tc1Spec := helper.NewTuningConfigSpecRootStub(tc1Name, firstVMDirtyRatio, firstPriority)
		tc1JSON, err := json.Marshal(tc1Spec)
		Expect(err).ToNot(HaveOccurred())
		tcArgs := &exec.TuningConfigArgs{
			Cluster: helper.StringPointer(clusterID),
			Name:    helper.StringPointer(name),
			Count:   helper.IntPointer(tcCount),
			Specs: &[]exec.TuningConfigSpec{
				exec.NewTuningConfigSpecFromString(string(tc1JSON)),
			},
		}
		_, err = tcService.Apply(tcArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify tuning config")
		tcsResp, err := cms.ListTuningConfigs(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(tcsResp.Size()).To(Equal(tcCount))
		tc := tcsResp.Items().Get(0)
		Expect(tc.Name()).To(Equal(name))
		verifyTuningConfigSpec(tc.Spec(), tc1Name, firstVMDirtyRatio, firstPriority)

		By("Delete created tuning config")
		_, err = tcService.Destroy()
		Expect(err).ToNot(HaveOccurred())

		By("Create many tuning configs")
		tcCount = 2
		tc2Spec := helper.NewTuningConfigSpecRootStub(tc2Name, firstVMDirtyRatio, firstPriority)
		tc2YAML, err := yaml.Marshal(tc2Spec)
		Expect(err).ToNot(HaveOccurred())
		tcArgs.Count = helper.IntPointer(tcCount)
		tcArgs.Specs = &[]exec.TuningConfigSpec{
			exec.NewTuningConfigSpecFromString(string(tc1JSON)),
			exec.NewTuningConfigSpecFromString(string(tc2YAML)),
		}
		_, err = tcService.Apply(tcArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify tuning configs")
		tcsResp, err = cms.ListTuningConfigs(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(tcsResp.Size()).To(Equal(tcCount))
		var expectedNames []string
		for i := 0; i < tcCount; i++ {
			expectedNames = append(expectedNames, fmt.Sprintf("%s-%v", name, i))
		}
		expectedProfileNames := []string{tc1Name, tc2Name}
		expectedSpecVMDirtyRatios := []int{firstVMDirtyRatio, firstVMDirtyRatio}
		expectedSpecPriorities := []int{firstPriority, firstPriority}
		for _, tc := range tcsResp.Items().Slice() {
			Expect(tc.Name()).To(BeElementOf(expectedNames))
			index := 0
			for index < len(expectedNames) {
				if expectedNames[index] == tc.Name() {
					break
				}
				index++
			}
			verifyTuningConfigSpec(tc.Spec(), expectedProfileNames[index], expectedSpecVMDirtyRatios[index], expectedSpecPriorities[index])
		}

		By("Update second tuning config")
		tc2Spec.Profile[0].Data = helper.NewTuningConfigSpecProfileData(secondVMDirtyRatio)
		tc2Spec.Recommend[0].Priority = secondPriority
		tc2YAML, err = yaml.Marshal(tc2Spec)
		Expect(err).ToNot(HaveOccurred())
		specFile1, err := helper.CreateTempFileWithContent(string(tc1JSON))
		Expect(err).ToNot(HaveOccurred())
		specFile2, err := helper.CreateTempFileWithContent(string(tc2YAML))
		Expect(err).ToNot(HaveOccurred())
		tcArgs.Specs = &[]exec.TuningConfigSpec{
			exec.NewTuningConfigSpecFromFile(string(specFile1)),
			exec.NewTuningConfigSpecFromFile(string(specFile2)),
		}
		_, err = tcService.Apply(tcArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Verify tuning configs")
		tcsResp, err = cms.ListTuningConfigs(ci.RHCSConnection, clusterID)
		Expect(err).ToNot(HaveOccurred())
		Expect(tcsResp.Size()).To(Equal(tcCount))
		expectedNames = []string{}
		for i := 0; i < tcCount; i++ {
			expectedNames = append(expectedNames, fmt.Sprintf("%s-%v", name, i))
		}
		expectedProfileNames = []string{tc1Name, tc2Name}
		expectedSpecVMDirtyRatios = []int{firstVMDirtyRatio, secondVMDirtyRatio}
		expectedSpecPriorities = []int{firstPriority, secondPriority}
		for _, tc := range tcsResp.Items().Slice() {
			Expect(tc.Name()).To(BeElementOf(expectedNames))
			index := 0
			for index < len(expectedNames) {
				if expectedNames[index] == tc.Name() {
					break
				}
				index++
			}
			verifyTuningConfigSpec(tc.Spec(), expectedProfileNames[index], expectedSpecVMDirtyRatios[index], expectedSpecPriorities[index])
		}

		By("Delete all")
		_, err = tcService.Destroy()
		Expect(err).ToNot(HaveOccurred())
	})

	It("can validate - [id:72522]", ci.Medium, func() {
		tcName := helper.GenerateRandomName("tc-72522", 3)
		getDefaultTCArgs := func() *exec.TuningConfigArgs {
			tc1Spec := helper.NewTuningConfigSpecRootStub("default-profile", 65, 10)
			tc1JSON, err := json.Marshal(tc1Spec)
			Expect(err).ToNot(HaveOccurred())
			return &exec.TuningConfigArgs{
				Cluster: helper.StringPointer(clusterID),
				Name:    helper.StringPointer(tcName),
				Specs: &[]exec.TuningConfigSpec{
					exec.NewTuningConfigSpecFromString(string(tc1JSON)),
				},
			}
		}

		validateTCArgAgainstErrorSubstrings := func(updateFields func(args *exec.TuningConfigArgs), errSubStrings ...string) {
			tcArgs := getDefaultTCArgs()
			updateFields(tcArgs)
			_, err := tcService.Apply(tcArgs)
			Expect(err).To(HaveOccurred())
			for _, errStr := range errSubStrings {
				helper.ExpectTFErrorContains(err, errStr)
			}
		}

		By("Try to create tuning config with empty cluster ID")
		validateTCArgAgainstErrorSubstrings(func(args *exec.TuningConfigArgs) {
			args.Cluster = helper.EmptyStringPointer
		}, "Attribute cluster cluster ID may not be empty/blank string, got:")

		By("Try to create tuning config with empty name")
		validateTCArgAgainstErrorSubstrings(func(args *exec.TuningConfigArgs) {
			args.Name = helper.EmptyStringPointer
		}, "The name must be a lowercase RFC 1123 subdomain")

		By("Try to create tuning config with empty spec")
		validateTCArgAgainstErrorSubstrings(func(args *exec.TuningConfigArgs) {
			args.Specs = &[]exec.TuningConfigSpec{
				exec.NewTuningConfigSpecFromString(""),
			}
		}, "Attribute 'spec' must be set")

		By("Create tuning config for edit")
		_, err := tcService.Apply(getDefaultTCArgs())
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
			validateTCArgAgainstErrorSubstrings(func(args *exec.TuningConfigArgs) {
				args.Cluster = helper.StringPointer(otherClusterID)
			}, "Attribute cluster, cannot be changed from")
		} else {
			Logger.Info("No other cluster accessible for testing this change")
		}

		By("Try to edit cluster field with wrong value")
		validateTCArgAgainstErrorSubstrings(func(args *exec.TuningConfigArgs) {
			args.Cluster = helper.StringPointer("wrong")
		}, "Attribute cluster, cannot be changed from")

		By("Try to edit name field")
		validateTCArgAgainstErrorSubstrings(func(args *exec.TuningConfigArgs) {
			args.Name = helper.StringPointer("new_name")
		}, "Attribute name, cannot be changed from")

		By("Try to edit spec field with non json value")
		validateTCArgAgainstErrorSubstrings(func(args *exec.TuningConfigArgs) {
			args.Specs = &[]exec.TuningConfigSpec{
				exec.NewTuningConfigSpecFromString("wrong"),
			}
		}, "cannot unmarshal string")

		By("Try to edit spec field with non yaml value")
		validateTCArgAgainstErrorSubstrings(func(args *exec.TuningConfigArgs) {
			args.Specs = &[]exec.TuningConfigSpec{
				exec.NewTuningConfigSpecFromString("wrong"),
			}
		}, "cannot unmarshal string")

		By("Get vpc output")
		vpcService, err := exec.NewVPCService(constants.GetAWSVPCDefaultManifestDir(profile.GetClusterType()))
		Expect(err).ToNot(HaveOccurred())
		vpcOutput, err := vpcService.Output()
		Expect(err).ToNot(HaveOccurred())

		By("Create machinepool with tuning config")
		replicas := 3
		machineType := "m5.2xlarge"
		name := helper.GenerateRandomName("np-72522", 2)
		subnetId := vpcOutput.PrivateSubnets[0]
		tuningConfigs := []string{tcName}
		mpArgs := &exec.MachinePoolArgs{
			Cluster:            helper.StringPointer(clusterID),
			AutoscalingEnabled: helper.BoolPointer(false),
			Replicas:           helper.IntPointer(replicas),
			Name:               helper.StringPointer(name),
			SubnetID:           helper.StringPointer(subnetId),
			MachineType:        helper.StringPointer(machineType),
			AutoRepair:         helper.BoolPointer(true),
			TuningConfigs:      helper.StringSlicePointer(tuningConfigs),
		}
		_, err = mpService.Apply(mpArgs)
		Expect(err).ToNot(HaveOccurred())

		By("Delete Tuning config used by machinepool")
		_, err = tcService.Destroy()
		Expect(err).To(HaveOccurred())
	})
})
