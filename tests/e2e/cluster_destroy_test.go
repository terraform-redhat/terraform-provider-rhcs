***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
***REMOVED***

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Create cluster test", func(***REMOVED*** {
		It("DestroyClusterByProfile", CI.Destroy,
			func(***REMOVED*** {

				// Generate/build cluster by profile selected
				profile := CI.LoadProfileYamlFileByENV(***REMOVED***
				err := CI.DestroyRHCSClusterByProfile(token, profile***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
