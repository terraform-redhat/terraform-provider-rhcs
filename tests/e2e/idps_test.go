package e2e

import (
	"fmt"
	"net/http"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	CI "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	l "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
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

		var profile *CI.Profile
		profile = CI.LoadProfileYamlFileByENV()
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
				It("OCP-63151 - Provision HTPASSWD IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP,
					ci.Exclude,
					func() {
						By("Create htpasswd idp for an existing cluster")

						idpParam := &exe.IDPArgs{
							Token:         token,
							ClusterID:     clusterID,
							Name:          "OCP-63151-htpasswd-idp-test",
							HtpasswdUsers: htpasswdMap,
						}
						err := idpService.htpasswd.Create(idpParam)
						Expect(err).ToNot(HaveOccurred())
						idpID, _ := idpService.htpasswd.Output()

						By("List existing HtpasswdUsers and compare to the created one")
						htpasswdUsersList, _ := cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpID.ID)
						Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
						respUserName, _ := htpasswdUsersList.Items().Slice()[0].GetUsername()
						Expect(respUserName).To(Equal(userName))

						By("Login with created htpasswd idp")

						// this condition is for cases where the cluster profile
						// has private_link enabled, then regular login won't work
						if !profile.PrivateLink {

							getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
							Expect(err).ToNot(HaveOccurred())
							server := getResp.Body().API().URL()

							ocAtter := &openshift.OcAttributes{
								Server:    server,
								Username:  userName,
								Password:  password,
								ClusterID: clusterID,
								AdditioanlFlags: []string{
									"--insecure-skip-tls-verify",
									fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
								},
								Timeout: 7,
							}
							_, err = openshift.OcLogin(*ocAtter)
							Expect(err).ToNot(HaveOccurred())
						} else {
							l.Logger.Infof("private_link is enabled, skipping login command check.")
						}
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
				It("OCP-63332 - Provision LDAP IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP,
					ci.Exclude,
					func() {
						By("Create LDAP idp for an existing cluster")

						idpParam := &exe.IDPArgs{
							Token:     token,
							ClusterID: clusterID,
							Name:      "OCP-63332-ldap-idp-test",
							CA:        "",
							URL:       con.LdapURL,
							Insecure:  true,
						}
						err := idpService.ldap.Create(idpParam)
						Expect(err).ToNot(HaveOccurred())

						By("Login with created ldap idp")

						// this condition is for cases where the cluster profile
						// has private_link enabled, then regular login won't work
						if !profile.PrivateLink {
							getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
							Expect(err).ToNot(HaveOccurred())
							server := getResp.Body().API().URL()

							ocAtter := &openshift.OcAttributes{
								Server:    server,
								Username:  userName,
								Password:  password,
								ClusterID: clusterID,
								AdditioanlFlags: []string{
									"--insecure-skip-tls-verify",
									fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
								},
								Timeout: 7,
							}
							_, err = openshift.OcLogin(*ocAtter)
							Expect(err).ToNot(HaveOccurred())
						} else {
							l.Logger.Infof("private_link is enabled, skipping login command check.")
						}
					})
			})
		})
		Describe("GitLab IDP test cases", func() {

			var gitlabIDPClientSecret, gitlabIDPClientId string

			BeforeEach(func() {
				gitlabIDPClientId = h.RandStringWithUpper(20)
				gitlabIDPClientSecret = h.RandStringWithUpper(30)
				idpService.gitlab = *exe.NewIDPService(con.GitlabDir) // init new gitlab service
			})

			AfterEach(func() {
				err := idpService.gitlab.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			Context("Author:smiron-High-OCP-64028 @OCP-64028 @smiron", func() {
				It("OCP-64028 - Provision GitLab IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP, func() {
					By("Create GitLab idp for an existing cluster")

					idpParam := &exe.IDPArgs{
						Token:        token,
						ClusterID:    clusterID,
						Name:         "OCP-64028-gitlab-idp-test",
						ClientID:     gitlabIDPClientId,
						ClientSecret: gitlabIDPClientSecret,
						URL:          con.GitLabURL,
					}
					err := idpService.gitlab.Create(idpParam)
					Expect(err).ToNot(HaveOccurred())

					By("Check gitlab idp created for the cluster")
					idpID, err := idpService.gitlab.Output()
					Expect(err).ToNot(HaveOccurred())

					resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.Status()).To(Equal(http.StatusOK))
				})
			})
		})
		Describe("GitHub IDP test cases", func() {

			var githubIDPClientSecret, githubIDPClientId string

			BeforeEach(func() {
				githubIDPClientSecret = h.RandStringWithUpper(20)
				githubIDPClientId = h.RandStringWithUpper(30)
				idpService.github = *exe.NewIDPService(con.GithubDir) // init new github service
			})

			AfterEach(func() {
				err := idpService.github.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			Context("Author:smiron-High-OCP-64027 @OCP-64027 @smiron", func() {
				It("OCP-64027 - Provision GitHub IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP, func() {
					By("Create GitHub idp for an existing cluster")

					idpParam := &exe.IDPArgs{
						Token:         token,
						ClusterID:     clusterID,
						Name:          "OCP-64027-github-idp-test",
						ClientID:      githubIDPClientId,
						ClientSecret:  githubIDPClientSecret,
						Organizations: con.Organizations,
					}
					err := idpService.github.Create(idpParam)
					Expect(err).ToNot(HaveOccurred())

					By("Check github idp created for the cluster")
					idpID, err := idpService.github.Output()
					Expect(err).ToNot(HaveOccurred())

					resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.Status()).To(Equal(http.StatusOK))
				})
			})
		})
		Describe("Google IDP test cases", func() {

			var googleIDPClientSecret, googleIDPClientId string

			BeforeEach(func() {
				googleIDPClientSecret = h.RandStringWithUpper(20)
				googleIDPClientId = h.RandStringWithUpper(30)
				idpService.google = *exe.NewIDPService(con.GoogleDir) // init new google service
			})

			AfterEach(func() {
				err := idpService.google.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			Context("Author:smiron-High-OCP-64029 @OCP-64029 @smiron", func() {
				It("OCP-64029 - Provision Google IDP against cluster using TF", ci.Day2, ci.High, ci.FeatureIDP, func() {
					By("Create Google idp for an existing cluster")

					idpParam := &exe.IDPArgs{
						Token:        token,
						ClusterID:    clusterID,
						Name:         "OCP-64029-google-idp-test",
						ClientID:     googleIDPClientId,
						ClientSecret: googleIDPClientSecret,
						HostedDomain: con.HostedDomain,
					}
					err := idpService.google.Create(idpParam)
					Expect(err).ToNot(HaveOccurred())

					By("Check google idp created for the cluster")
					idpID, err := idpService.google.Output()
					Expect(err).ToNot(HaveOccurred())

					resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
					Expect(err).ToNot(HaveOccurred())
					Expect(resp.Status()).To(Equal(http.StatusOK))
				})
			})
		})
	})
})
