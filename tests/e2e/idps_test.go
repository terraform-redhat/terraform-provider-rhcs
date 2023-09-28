***REMOVED***

***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***
	conn "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***

var _ = Describe("TF Test", func(***REMOVED*** {
	Describe("Identity Providers test cases", func(***REMOVED*** {

***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***

***REMOVED***

		Describe("Htpasswd IDP test cases", func(***REMOVED*** {
***REMOVED***
			var htpasswdUsername, htpasswdPassword string

			BeforeEach(func(***REMOVED*** {

				htpasswdUsername = "jacko"
				htpasswdPassword = "1q2wFe4rpoe2318"
				htpasswdMap = []interface{}{map[string]string{"username": htpasswdUsername, "password": htpasswdPassword}}
				idpService.htpasswd = *exe.NewIDPService(con.HtpasswdDir***REMOVED*** // init new htpasswd service
	***REMOVED******REMOVED***

			AfterEach(func(***REMOVED*** {
				err := idpService.htpasswd.Destroy(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	***REMOVED******REMOVED***

			Context("Author:smiron-High-OCP-63151 @OCP-63151 @smiron", func(***REMOVED*** {
				It("OCP-63151 - Provision HTPASSWD IDP against cluster using TF", func(***REMOVED*** {
					By("Create htpasswd idp for an existing cluster"***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
						Name:          "htpasswd-idp-test",
						HtpasswdUsers: htpasswdMap,
			***REMOVED***
					err := idpService.htpasswd.Create(idpParam, "-auto-approve", "-no-color"***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
					idpID, _ := idpService.htpasswd.Output(***REMOVED***

					By("List existing HtpasswdUsers and compare to the created one"***REMOVED***
					htpasswdUsersList, _ := cms.ListHtpasswdUsers(conn.RHCSConnection, clusterID, idpID.ID***REMOVED***
					Expect(htpasswdUsersList.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***
					respUserName, _ := htpasswdUsersList.Items(***REMOVED***.Slice(***REMOVED***[0].GetUsername(***REMOVED***
					Expect(respUserName***REMOVED***.To(Equal(htpasswdUsername***REMOVED******REMOVED***

					By("Login with created htpasswd idp"***REMOVED***
					getResp, err := cms.RetrieveClusterDetail(conn.RHCSConnection, clusterID***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
					server := getResp.Body(***REMOVED***.API(***REMOVED***.URL(***REMOVED***

					ocAtter := &exe.OcAttributes{
						Server:          server,
						Username:        htpasswdUsername,
						Password:        htpasswdPassword,
						ClusterID:       clusterID,
						AdditioanlFlags: nil,
						Timeout:         5,
			***REMOVED***
					err = exe.OcLogin(*ocAtter***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

		***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
