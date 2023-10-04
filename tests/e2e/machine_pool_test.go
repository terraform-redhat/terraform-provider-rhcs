***REMOVED***

***REMOVED***

	// nolint

***REMOVED***
***REMOVED***
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
***REMOVED***
***REMOVED***
***REMOVED***

var logger logging.Logger

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Create MachinePool test cases", func(***REMOVED*** {
		var mpService *exe.MachinePoolService
		BeforeEach(func(***REMOVED*** {
			mpService = exe.NewMachinePoolService(con.MachinePoolDir***REMOVED***
***REMOVED******REMOVED***
		AfterEach(func(***REMOVED*** {
			err := mpService.Destroy(***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
			return
***REMOVED******REMOVED***
		It("MachinePoolExampleNegative", func(***REMOVED*** {

			MachinePoolArgs := &exe.MachinePoolArgs{
				Token:       token,
				Cluster:     clusterID,
				Replicas:    3,
				MachineType: "invalid",
				Name:        "testmp",
	***REMOVED***

			err := mpService.Create(MachinePoolArgs***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("is not supported for cloud provider 'aws'"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:amalykhi-High-OCP-64757 @OCP-64757 @amalykhi", func(***REMOVED*** {
			It("Author:amalykhi-High-OCP-64757 Create a second machine pool", ci.Day2, ci.High, ci.FeatureMachinepool, func(***REMOVED*** {
				By("Create a second machine pool"***REMOVED***
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64757"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    replicas,
					MachineType: machineType,
					Name:        name,
		***REMOVED***

				err := mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				By("Verify the parameters of the created machinepool"***REMOVED***
				mpOut, err := mpService.Output(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Replicas(***REMOVED******REMOVED***.To(Equal(mpOut.Replicas***REMOVED******REMOVED***
				Expect(mpResponseBody.InstanceType(***REMOVED******REMOVED***.To(Equal(mpOut.MachineType***REMOVED******REMOVED***
				Expect(mpResponseBody.ID(***REMOVED******REMOVED***.To(Equal(mpOut.Name***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:amalykhi-High-OCP-64905 @OCP-64905 @amalykhi", func(***REMOVED*** {
			It("Author:amalykhi-High-OCP-64905 Edit/delete second machinepool labels", ci.Day2, ci.High, ci.FeatureMachinepool, func(***REMOVED*** {
				By("Create additional machinepool with labels"***REMOVED***
				replicas := 3
				machineType := "r5.xlarge"
				name := "ocp-64905"
				creationLabels := map[string]string{"fo1": "bar1", "fo2": "baz2"}
				updatingLabels := map[string]string{"fo1": "bar3", "fo3": "baz3"}
				emptyLabels := map[string]string{}
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     clusterID,
					Replicas:    replicas,
					MachineType: machineType,
					Name:        name,
					Labels:      creationLabels,
		***REMOVED***

				err := mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				By("Verify the parameters of the created machinepool"***REMOVED***
				mpResponseBody, err := cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Labels(***REMOVED******REMOVED***.To(Equal(creationLabels***REMOVED******REMOVED***
				By("Edit the labels of the machinepool"***REMOVED***
				MachinePoolArgs.Labels = updatingLabels
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Labels(***REMOVED******REMOVED***.To(Equal(updatingLabels***REMOVED******REMOVED***

				By("Delete the labels of the machinepool"***REMOVED***
				MachinePoolArgs.Labels = emptyLabels
				err = mpService.Create(MachinePoolArgs***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				mpResponseBody, err = cms.RetrieveClusterMachinePool(ci.RHCSConnection, clusterID, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mpResponseBody.Labels(***REMOVED******REMOVED***.To(BeNil(***REMOVED******REMOVED***

	***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
