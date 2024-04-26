package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
)

var _ = Describe("Tuning Config", ci.NonClassicCluster, ci.FeatureTuningConfig, func() {
	var (
		tcService *exec.TuningConfigService
		tcArgs    *exec.TuningConfigArgs
	)

	verifyTuningConfigSpec := func(spec interface{}, specVmDirtyRatio, specPriority int) {
		tcSpec := spec.(map[string]interface{})
		tcProfileSpec := (tcSpec["profile"].([]interface{}))[0].(map[string]interface{})
		tcRecommendSpec := (tcSpec["recommend"].([]interface{}))[0].(map[string]interface{})
		Expect(tcProfileSpec["data"]).To(ContainSubstring(fmt.Sprintf("vm.dirty_ratio=\"%d\"", specVmDirtyRatio)))
		Expect(tcRecommendSpec["priority"]).To(BeEquivalentTo(specPriority))
	}

	BeforeEach(func() {
		tcService = exec.NewTuningConfigService(constants.TuningConfigDir)
	})

	AfterEach(func() {
		tcService.Destroy()
	})

	It("can create/edit/delete - [id:72521]",
		ci.Day2, ci.High, func() {
			namePrefix := "tc-72521"
			tcCount := 1
			specVMDirtyRatios := []int{65}
			specPriorities := []int{20}
			By("Create one tuning config")
			tcArgs = &exec.TuningConfigArgs{
				Cluster:           clusterID,
				NamePrefix:        namePrefix,
				Count:             &tcCount,
				SpecVMDirtyRatios: &specVMDirtyRatios,
				SpecPriorities:    &specPriorities,
			}
			_, err := tcService.Apply(tcArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning config")
			tcsResp, err := cms.ListTuningConfigs(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(tcsResp.Size()).To(Equal(tcCount))
			tc := tcsResp.Items().Get(0)
			Expect(tc.Name()).To(Equal(fmt.Sprintf("%s-%v", namePrefix, 0)))
			verifyTuningConfigSpec(tc.Spec(), specVMDirtyRatios[0], specPriorities[0])

			By("Add one more tuning config")
			tcCount = 2
			specVMDirtyRatios = append(specVMDirtyRatios, 45)
			specPriorities = append(specPriorities, 10)
			_, err = tcService.Apply(tcArgs, false)
			Expect(err).ToNot(HaveOccurred())

			By("Verify tuning configs")
			tcsResp, err = cms.ListTuningConfigs(ci.RHCSConnection, clusterID)
			Expect(err).ToNot(HaveOccurred())
			Expect(tcsResp.Size()).To(Equal(tcCount))
			for index, tc := range tcsResp.Items().Slice() {
				Expect(tc.Name()).To(Equal(fmt.Sprintf("%s-%v", namePrefix, index)))
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
				Expect(tc.Name()).To(Equal(fmt.Sprintf("%s-%v", namePrefix, index)))
				verifyTuningConfigSpec(tc.Spec(), specVMDirtyRatios[index], specPriorities[index])
			}

			By("Delete all")
			_, err = tcService.Destroy()
			Expect(err).ToNot(HaveOccurred())
		})
})
