***REMOVED***

***REMOVED***

	// nolint

***REMOVED***
	"os"

***REMOVED***
***REMOVED***
	"github.com/openshift-online/ocm-sdk-go/logging"
	conn "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	cms "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
***REMOVED***
***REMOVED***
***REMOVED***

var logger logging.Logger

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Create MachinePool test", func(***REMOVED*** {
		var cluster_id string
		var mpService *exe.MachinePoolService
		BeforeEach(func(***REMOVED*** {
			cluster_id = os.Getenv("CLUSTER_ID"***REMOVED***
			mpService = exe.NewMachinePoolService(con.MachinePoolDir***REMOVED***
***REMOVED******REMOVED***
		AfterEach(func(***REMOVED*** {
			err := mpService.Destroy(***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
			return
***REMOVED******REMOVED***
		It("MachinePoolExampleNegative", func(***REMOVED*** {

			mpParam := &exe.MachinePoolArgs{
				Token:       token,
				Cluster:     cluster_id,
				Replicas:    3,
				MachineType: "invalid",
				Name:        "testmp",
	***REMOVED***

			_, err := exe.CreateMyTFMachinePool(mpParam, "-auto-approve", "-no-color"***REMOVED***
			Expect(err***REMOVED***.To(HaveOccurred(***REMOVED******REMOVED***
			Expect(err.Error(***REMOVED******REMOVED***.Should(ContainSubstring("is not supported for cloud provider 'aws'"***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:amalykhi-High-OCP-64757 @OCP-64757 @amalykhi", func(***REMOVED*** {
			It("Create a second machine pool", func(***REMOVED*** {
				token := os.Getenv(con.TokenENVName***REMOVED***
				replicas := 3
				machine_type := "r5.xlarge"
				name := "testmp14"
				MachinePoolArgs := &exe.MachinePoolArgs{
					Token:       token,
					Cluster:     cluster_id,
					Replicas:    replicas,
					MachineType: machine_type,
					Name:        name,
		***REMOVED***

				err := mpService.Create(MachinePoolArgs***REMOVED***
				mp_out, err := mpService.Output(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				mp_response, err := cms.RetrieveClusterMachinePool(conn.RHCSConnection, cluster_id, name***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(mp_response.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***
				respBody := mp_response.Body(***REMOVED***
				Expect(respBody.Replicas(***REMOVED******REMOVED***.To(Equal(mp_out.Replicas***REMOVED******REMOVED***
				Expect(respBody.InstanceType(***REMOVED******REMOVED***.To(Equal(mp_out.MachineType***REMOVED******REMOVED***
				Expect(respBody.ID(***REMOVED******REMOVED***.To(Equal(mp_out.Name***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
