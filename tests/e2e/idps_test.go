package e2e

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("TF Test", func() {
	Describe("Identity Providers test cases", func() {

		// all identity providers - declared for future cases
		type IDPServices struct {
			htpasswd,
			github,
			gitlab,
			google,
			ldap,
			openid exe.IDPService
		}

		var idpService IDPServices
		var userName, password string

		Describe("Htpasswd IDP test cases", func() {
			var htpasswdMap = []interface{}{map[string]string{}}
			var userName, password string

			BeforeEach(func() {

				userName = "jacko"
				password = h.RandStringWithUpper(15)
				htpasswdMap = []interface{}{map[string]string{"username": userName, "password": password}}
				idpService.htpasswd = *exe.NewIDPService(con.HtpasswdDir) // init new htpasswd service
			})

			AfterEach(func() {
				err := idpService.htpasswd.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			Context("Author:smiron-High-OCP-63151 @OCP-63151 @smiron", func() {
				It("OCP-63151 - Provision HTPASSWD IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP, func() {
					By("Create htpasswd idp for an existing cluster")

					idpParam := &exe.IDPArgs{
						Token:         token,
						ClusterID:     clusterID,
						Name:          "htpasswd-idp-test",
						HtpasswdUsers: htpasswdMap,
					}
					err := idpService.htpasswd.Create(idpParam, "-auto-approve", "-no-color")
					Expect(err).ToNot(HaveOccurred())
					idpID, _ := idpService.htpasswd.Output()

					By("List existing HtpasswdUsers and compare to the created one")
					htpasswdUsersList, _ := cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpID.ID)
					Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
					respUserName, _ := htpasswdUsersList.Items().Slice()[0].GetUsername()
					Expect(respUserName).To(Equal(userName))

					By("Login with created htpasswd idp")
					getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred())
					server := getResp.Body().API().URL()

					ocAtter := &openshift.OcAttributes{
						Server:          server,
						Username:        userName,
						Password:        password,
						ClusterID:       clusterID,
						AdditioanlFlags: []string{"--insecure-skip-tls-verify"},
						Timeout:         7,
					}
					_, err = openshift.OcLogin(*ocAtter)
					Expect(err).ToNot(HaveOccurred())

				})
			})
		})
		Describe("LDAP IDP test cases", func() {

			BeforeEach(func() {

				userName = "newton"
				password = "password"
				idpService.ldap = *exe.NewIDPService(con.LdapDir) // init new ldap service
			})

			AfterEach(func() {
				err := idpService.ldap.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			Context("Author:smiron-High-OCP-63332 @OCP-63332 @smiron", func() {
				It("OCP-63332 - Provision LDAP IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP, func() {
					By("Create LDAP idp for an existing cluster")

					idpParam := &exe.IDPArgs{
						Token:     token,
						ClusterID: clusterID,
						Name:      "ldap-idp-test",
						CA:        "",
						URL:       con.LdapURL,
						Insecure:  true,
					}
					err := idpService.ldap.Create(idpParam, "-auto-approve", "-no-color")
					Expect(err).ToNot(HaveOccurred())

					By("Login with created ldap idp")
					getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred())
					server := getResp.Body().API().URL()

					ocAtter := &openshift.OcAttributes{
						Server:          server,
						Username:        userName,
						Password:        password,
						ClusterID:       clusterID,
						AdditioanlFlags: []string{"--insecure-skip-tls-verify"},
						Timeout:         7,
					}
					_, err = openshift.OcLogin(*ocAtter)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
