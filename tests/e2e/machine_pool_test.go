***REMOVED***

***REMOVED***

	// nolint

	"os"

	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"

***REMOVED***
***REMOVED***
***REMOVED***

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Create MachinePool test", func(***REMOVED*** {
		var cluster_id string
		BeforeEach(func(***REMOVED*** {
			cluster_id = os.Getenv("CLUSTER_ID"***REMOVED***
***REMOVED******REMOVED***
		AfterEach(func(***REMOVED*** {
			return
***REMOVED******REMOVED***
		It("MachinePoolExampleNegative", func(***REMOVED*** {

			mpParam := &EXE.MachinePoolArgs{
				Token: token,
				// OCMENV:      "staging",
				Cluster:     cluster_id,
				Replicas:    3,
				MachineType: "invalid",
				Name:        "testmp",
	***REMOVED***

			_, err := EXE.CreateMyTFMachinePool(mpParam, "-auto-approve", "-no-color"***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("is not supported for cloud provider 'aws'"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("MachinePoolExamplePositive", func(***REMOVED*** {
			mpParam := &EXE.MachinePoolArgs{
				Token: token,
				// OCMENV:      "staging",
				Cluster:     cluster_id,
				Replicas:    3,
				MachineType: "r5.xlarge",
				Name:        "testmp3",
	***REMOVED***

			_, err := EXE.CreateMyTFMachinePool(mpParam, "-no-color"***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

			err = EXE.DestroyMyTFMachinePool(mpParam, "-no-color"***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
