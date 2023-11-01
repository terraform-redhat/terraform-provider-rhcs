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
***REMOVED***
***REMOVED***
					func(***REMOVED*** {
						By("Create htpasswd idp for an existing cluster"***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
							Name:          "htpasswd-idp-test",
***REMOVED***
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

***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
								fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName***REMOVED******REMOVED******REMOVED***,
					***REMOVED***,
***REMOVED***
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
***REMOVED***
***REMOVED***
					func(***REMOVED*** {
						By("Create LDAP idp for an existing cluster"***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
							Name:      "ldap-idp-test",
***REMOVED***
***REMOVED***
***REMOVED***
				***REMOVED***
						err := idpService.ldap.Create(idpParam, "-auto-approve", "-no-color"***REMOVED***
						Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

						By("Login with created ldap idp"***REMOVED***
						getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID***REMOVED***
						Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
						server := getResp.Body(***REMOVED***.API(***REMOVED***.URL(***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
								fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName***REMOVED******REMOVED******REMOVED***,
					***REMOVED***,
***REMOVED***
				***REMOVED***
						_, err = openshift.OcLogin(*ocAtter***REMOVED***
						Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
			***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Describe("GitLab IDP test cases", func(***REMOVED*** {

***REMOVED***

			BeforeEach(func(***REMOVED*** {
				gitlabIDPClientId = h.RandStringWithUpper(20***REMOVED***
				gitlabIDPClientSecret = h.RandStringWithUpper(30***REMOVED***
				idpService.gitlab = *exe.NewIDPService(con.GitlabDir***REMOVED*** // init new gitlab service
	***REMOVED******REMOVED***

			AfterEach(func(***REMOVED*** {
				err := idpService.gitlab.Destroy(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	***REMOVED******REMOVED***

			Context("Author:smiron-High-OCP-64028 @OCP-64028 @smiron", func(***REMOVED*** {
				It("OCP-64028 - Provision GitLab IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP, func(***REMOVED*** {
					By("Create GitLab idp for an existing cluster"***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
						Name:         "gitlab-idp-test",
***REMOVED***
***REMOVED***
***REMOVED***
			***REMOVED***
					err := idpService.gitlab.Create(idpParam, "-auto-approve", "-no-color"***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

					By("Check gitlab idp created for the cluster"***REMOVED***
					idpID, err := idpService.gitlab.Output(***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

					resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
					Expect(resp.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***
		***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
		Describe("GitHub IDP test cases", func(***REMOVED*** {

***REMOVED***

			BeforeEach(func(***REMOVED*** {
				githubIDPClientSecret = h.RandStringWithUpper(20***REMOVED***
				githubIDPClientId = h.RandStringWithUpper(30***REMOVED***
				idpService.github = *exe.NewIDPService(con.GithubDir***REMOVED*** // init new github service
	***REMOVED******REMOVED***

			AfterEach(func(***REMOVED*** {
				err := idpService.github.Destroy(***REMOVED***
				Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	***REMOVED******REMOVED***

			Context("Author:smiron-High-OCP-64027 @OCP-64027 @smiron", func(***REMOVED*** {
				It("OCP-64027 - Provision GitHub IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP, func(***REMOVED*** {
					By("Create GitHub idp for an existing cluster"***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
						Name:          "github-idp-test",
***REMOVED***
***REMOVED***
***REMOVED***
			***REMOVED***
					err := idpService.github.Create(idpParam, "-auto-approve", "-no-color"***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

					By("Check github idp created for the cluster"***REMOVED***
					idpID, err := idpService.github.Output(***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***

					resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID***REMOVED***
					Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
					Expect(resp.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***
		***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
