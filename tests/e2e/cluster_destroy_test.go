***REMOVED***

***REMOVED***
	"os"

***REMOVED***
***REMOVED***
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Create cluster test", func(***REMOVED*** {
		It("DestroyClusterByProfile", CI.Destroy,
			func(***REMOVED*** {
				// Destroy kubeconfig folder
				if _, err := os.Stat(CON.RHCS.KubeConfigDir***REMOVED***; err == nil {
					os.RemoveAll(CON.RHCS.KubeConfigDir***REMOVED***
		***REMOVED***

				// Generate/build cluster by profile selected
				profile := CI.LoadProfileYamlFileByENV(***REMOVED***
				err := CI.DestroyRHCSClusterByProfile(token, profile***REMOVED***

				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
