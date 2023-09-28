package e2e

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	conn "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
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

		Describe("Htpasswd IDP test cases", func() {
			var htpasswdMap = []interface{}{map[string]string{}}
			var htpasswdUsername, htpasswdPassword string

			BeforeEach(func() {

				htpasswdUsername = "jacko"
				htpasswdPassword = "1q2wFe4rpoe2318"
				htpasswdMap = []interface{}{map[string]string{"username": htpasswdUsername, "password": htpasswdPassword}}
				idpService.htpasswd = *exe.NewIDPService(con.HtpasswdDir) // init new htpasswd service
			})

			AfterEach(func() {
				err := idpService.htpasswd.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			Context("Author:smiron-High-OCP-63151 @OCP-63151 @smiron", func() {
				It("OCP-63151 - Provision HTPASSWD IDP against cluster using TF", func() {
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
					htpasswdUsersList, _ := cms.ListHtpasswdUsers(conn.RHCSConnection, clusterID, idpID.ID)
					Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
					respUserName, _ := htpasswdUsersList.Items().Slice()[0].GetUsername()
					Expect(respUserName).To(Equal(htpasswdUsername))

					By("Login with created htpasswd idp")
					getResp, err := cms.RetrieveClusterDetail(conn.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred())
					server := getResp.Body().API().URL()

					ocAtter := &exe.OcAttributes{
						Server:          server,
						Username:        htpasswdUsername,
						Password:        htpasswdPassword,
						ClusterID:       clusterID,
						AdditioanlFlags: nil,
						Timeout:         5,
					}
					err = exe.OcLogin(*ocAtter)
					Expect(err).ToNot(HaveOccurred())

				})
			})
		})
	})
})
