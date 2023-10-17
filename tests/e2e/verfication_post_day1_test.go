***REMOVED***

***REMOVED***
	// nolint

	"strings"

***REMOVED***
***REMOVED***

	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	CMS "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	H "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
***REMOVED***

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Verfication/Post day 1 tests", func(***REMOVED*** {
		var err error
		var profile *CI.Profile

		BeforeEach(func(***REMOVED*** {
			profile = CI.LoadProfileYamlFileByENV(***REMOVED***
			Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		AfterEach(func(***REMOVED*** {
***REMOVED******REMOVED***

		Context("Author:smiron-High-OCP-63140 @OCP-63140 @smiron", func(***REMOVED*** {
			It("Verify fips is enabled/disabled post cluster creation", CI.Day1Post, CI.High, func(***REMOVED*** {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(getResp.Body(***REMOVED***.FIPS(***REMOVED******REMOVED***.To(Equal(profile.FIPS***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:smiron-High-OCP-63133 @OCP-63133 @smiron", func(***REMOVED*** {
			It("Verify private_link is enabled/disabled post cluster creation", CI.Day1Post, CI.High, func(***REMOVED*** {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(getResp.Body(***REMOVED***.AWS(***REMOVED***.PrivateLink(***REMOVED******REMOVED***.To(Equal(profile.PrivateLink***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:smiron-High-OCP-63143 @OCP-63143 @smiron", func(***REMOVED*** {
			It("Verify etcd-encryption is enabled/disabled post cluster creation", CI.Day1Post, CI.High, func(***REMOVED*** {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(getResp.Body(***REMOVED***.EtcdEncryption(***REMOVED******REMOVED***.To(Equal(profile.Etcd***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:smiron-Medium-OCP-64023 @OCP-64023 @smiron", func(***REMOVED*** {
			It("Verify compute_machine_type value is set post cluster creation", CI.Day1Post, CI.Medium, func(***REMOVED*** {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(getResp.Body(***REMOVED***.Nodes(***REMOVED***.ComputeMachineType(***REMOVED***.ID(***REMOVED******REMOVED***.To(Equal(profile.ComputeMachineType***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:smiron-Medium-OCP-63141 @OCP-63141 @smiron", func(***REMOVED*** {
			It("Verify availability zones and multi-az is set post cluster creation", CI.Day1Post, CI.Medium, func(***REMOVED*** {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID***REMOVED***
				zonesArray := strings.Split(profile.Zones, ","***REMOVED***
				clusterAvailZones := getResp.Body(***REMOVED***.Nodes(***REMOVED***.AvailabilityZones(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				Expect(getResp.Body(***REMOVED***.MultiAZ(***REMOVED******REMOVED***.To(Equal(profile.MultiAZ***REMOVED******REMOVED***
				if clusterAvailZones != nil {
					Expect(clusterAvailZones***REMOVED***.
						To(Equal(H.JoinStringWithArray(profile.Region, zonesArray***REMOVED******REMOVED******REMOVED***
		***REMOVED*** else {
					Expect(clusterAvailZones***REMOVED***.To(Equal(nil***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Context("Author:smiron-High-OCP-68423 @OCP-68423 @smiron", func(***REMOVED*** {
			It("Verify compute_labels are set post cluster creation", CI.Day1Post, CI.High, func(***REMOVED*** {
				getResp, err := CMS.RetrieveClusterDetail(CI.RHCSConnection, clusterID***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
				if profile.Labeling {
					Expect(getResp.Body(***REMOVED***.Nodes(***REMOVED***.ComputeLabels(***REMOVED******REMOVED***.To(Equal(CON.DefaultMPLabels***REMOVED******REMOVED***
		***REMOVED*** else {
					Expect(getResp.Body(***REMOVED***.Nodes(***REMOVED***.ComputeLabels(***REMOVED******REMOVED***.To(Equal(""***REMOVED******REMOVED***
		***REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
