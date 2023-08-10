package e2e

import (

	// nolint

	"os"

	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TF Test", func() {
	Describe("Create MachinePool test", func() {
		var cluster_id string
		BeforeEach(func() {
			cluster_id = os.Getenv("CLUSTER_ID")
		})
		AfterEach(func() {
			return
		})
		It("MachinePoolExampleNegative", func() {

			mpParam := &EXE.MachinePoolArgs{
				Token: token,
				// OCMENV:      "staging",
				Cluster:     cluster_id,
				Replicas:    3,
				MachineType: "invalid",
				Name:        "testmp",
			}

			_, err := EXE.CreateMyTFMachinePool(mpParam, "-auto-approve", "-no-color")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("is not supported for cloud provider 'aws'"))
		})
		It("MachinePoolExamplePositive", func() {
			mpParam := &EXE.MachinePoolArgs{
				Token: token,
				// OCMENV:      "staging",
				Cluster:     cluster_id,
				Replicas:    3,
				MachineType: "r5.xlarge",
				Name:        "testmp3",
			}

			_, err := EXE.CreateMyTFMachinePool(mpParam, "-no-color")
			Expect(err).ToNot(HaveOccurred())

			err = EXE.DestroyMyTFMachinePool(mpParam, "-no-color")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
