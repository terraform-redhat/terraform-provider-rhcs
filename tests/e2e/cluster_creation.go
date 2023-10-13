***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
***REMOVED***

var _ = Describe("RHCS Provider Test", func(***REMOVED*** {
	Describe("Create cluster test", func(***REMOVED*** {
		It("CreateClusterByProfile", CI.Day1Prepare,
			func(***REMOVED*** {

				// Generate/build cluster by profile selected
				profile := CI.LoadProfileYamlFileByENV(***REMOVED***
				clusterID, err := CI.CreateRHCSClusterByProfile(token, profile***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(clusterID***REMOVED***.ToNot(BeEmpty(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
