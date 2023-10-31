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
***REMOVED***

		Describe("Htpasswd IDP test cases", func(***REMOVED*** {
***REMOVED***
	***REMOVED***

			BeforeEach(func(***REMOVED*** {

***REMOVED***
				password = h.RandStringWithUpper(15***REMOVED***
***REMOVED***
				idpService.htpasswd = *exe.NewIDPService(con.HtpasswdDir***REMOVED*** // init new htpasswd service
	***REMOVED******REMOVED***

			AfterEach(func(***REMOVED*** {
				err := idpService.htpasswd.Destroy(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	***REMOVED******REMOVED***

			Context("Author:smiron-High-OCP-63151 @OCP-63151 @smiron", func(***REMOVED*** {
***REMOVED*** func(***REMOVED*** {
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
					htpasswdUsersList, _ := cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpID.ID***REMOVED***
					Expect(htpasswdUsersList.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***
					respUserName, _ := htpasswdUsersList.Items(***REMOVED***.Slice(***REMOVED***[0].GetUsername(***REMOVED***
					Expect(respUserName***REMOVED***.To(Equal(userName***REMOVED******REMOVED***

					By("Login with created htpasswd idp"***REMOVED***
					getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
					server := getResp.Body(***REMOVED***.API(***REMOVED***.URL(***REMOVED***

					ocAtter := &openshift.OcAttributes{
						Server:          server,
						Username:        userName,
						Password:        password,
						ClusterID:       clusterID,
						AdditioanlFlags: []string{"--insecure-skip-tls-verify"},
						Timeout:         7,
			***REMOVED***
					_, err = openshift.OcLogin(*ocAtter***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

		***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Describe("LDAP IDP test cases", func(***REMOVED*** {

			BeforeEach(func(***REMOVED*** {

***REMOVED***
***REMOVED***
				idpService.ldap = *exe.NewIDPService(con.LdapDir***REMOVED*** // init new ldap service
	***REMOVED******REMOVED***

			AfterEach(func(***REMOVED*** {
				err := idpService.ldap.Destroy(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	***REMOVED******REMOVED***

			Context("Author:smiron-High-OCP-63332 @OCP-63332 @smiron", func(***REMOVED*** {
***REMOVED*** func(***REMOVED*** {
					By("Create LDAP idp for an existing cluster"***REMOVED***

***REMOVED***
						Token:     token,
						ClusterID: clusterID,
						Name:      "ldap-idp-test",
						CA:        "",
						URL:       con.LdapURL,
						Insecure:  true,
			***REMOVED***
					err := idpService.ldap.Create(idpParam, "-auto-approve", "-no-color"***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

					By("Login with created ldap idp"***REMOVED***
					getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
					server := getResp.Body(***REMOVED***.API(***REMOVED***.URL(***REMOVED***

					ocAtter := &openshift.OcAttributes{
						Server:          server,
						Username:        userName,
						Password:        password,
						ClusterID:       clusterID,
						AdditioanlFlags: []string{"--insecure-skip-tls-verify"},
						Timeout:         7,
			***REMOVED***
					_, err = openshift.OcLogin(*ocAtter***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
