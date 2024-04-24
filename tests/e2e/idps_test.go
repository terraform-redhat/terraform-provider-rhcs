package e2e

import (
	"fmt"
	"net/http"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	l "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

var _ = Describe("Identity Providers", ci.Day2, ci.FeatureIDP, func() {
	defer GinkgoRecover()

	// all identity providers - declared for future cases
	type IDPServices struct {
		htpasswd,
		github,
		gitlab,
		google,
		ldap,
		multi_idp,
		openid exe.IDPService
	}

	var profile *ci.Profile

	var idpService IDPServices
	var importService exe.ImportService
	var htpasswdMap = []interface{}{map[string]string{}}

	var userName, password,
		googleIdpName,
		gitLabIdpName,
		googleIDPClientSecret, googleIDPClientId,
		gitlabIDPClientSecret, gitlabIDPClientId,
		githubIDPClientSecret, githubIDPClientId string

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()
	})

	Describe("provision and update", func() {
		Context("Htpasswd", func() {

			BeforeEach(func() {
				userName = "my-admin-user"
				password = h.GenerateRandomStringWithSymbols(15)
				htpasswdMap = []interface{}{map[string]string{"username": userName, "password": password}}
				idpService.htpasswd = *exe.NewIDPService(CON.HtpasswdDir) // init new htpasswd service
			})

			AfterEach(func() {
				err := idpService.htpasswd.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})
			It("will succeed - [id:63151]", ci.High, ci.Exclude, func() {
				By("Create htpasswd idp for an existing cluster")

				idpParam := &exe.IDPArgs{
					ClusterID:     clusterID,
					Name:          "OCP-63151-htpasswd-idp-test",
					HtpasswdUsers: htpasswdMap,
				}
				err := idpService.htpasswd.Apply(idpParam, false)
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
				if !profile.IsPrivateLink() {
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
							fmt.Sprintf("--kubeconfig %s", path.Join(CON.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
						},
						Timeout: 7,
					}
					_, err = openshift.OcLogin(*ocAtter)
					Expect(err).ToNot(HaveOccurred())
				} else {
					l.Logger.Infof("private_link is enabled, skipping login command check.")
				}
			})
			It("Update htpasswd idp - [id:73154]", ci.High, func() {
				htpIDPName := "tf-htpassed-idp"
				By("Create htpasswd idp for an existing cluster")
				idpParam := &exe.IDPArgs{
					ClusterID:     clusterID,
					Name:          htpIDPName,
					HtpasswdUsers: htpasswdMap,
				}
				err := idpService.htpasswd.Apply(idpParam, false)
				Expect(err).ToNot(HaveOccurred())
				idpID, _ := idpService.htpasswd.Output()

				By("List existing HtpasswdUsers and compare to the created one")
				htpasswdUsersList, _ := cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpID.ID)
				Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
				respUserName, _ := htpasswdUsersList.Items().Slice()[0].GetUsername()
				Expect(respUserName).To(Equal(userName))

				By("Update htpasswd idp password of 'my-admin-user'")
				newPassword := h.GenerateRandomStringWithSymbols(15)
				htpasswdMap = []interface{}{map[string]string{"username": userName, "password": newPassword}}
				idpParam = &exe.IDPArgs{
					ClusterID:     clusterID,
					Name:          htpIDPName,
					HtpasswdUsers: htpasswdMap,
				}
				err = idpService.htpasswd.Apply(idpParam, true)
				Expect(err).ToNot(HaveOccurred())

				By("Check resource state file is updated")
				resource, err := h.GetResource(CON.HtpasswdDir, "rhcs_identity_provider", "htpasswd_idp")
				Expect(err).ToNot(HaveOccurred())
				Expect(resource).To(MatchJQ(fmt.Sprintf(`.instances[0].attributes.htpasswd.users[] | select(.username == "%s") .password`, userName), newPassword))

				By("Update htpasswd idp by adding two new users")
				userName2 := "my-admin-user2"
				password2 := h.GenerateRandomStringWithSymbols(15)
				userName3 := "my-admin-user3"
				password3 := h.GenerateRandomStringWithSymbols(15)

				htpasswdMap = []interface{}{map[string]string{"username": userName, "password": password},
					map[string]string{"username": userName2, "password": password2},
					map[string]string{"username": userName3, "password": password3}}
				idpParam = &exe.IDPArgs{
					ClusterID:     clusterID,
					Name:          htpIDPName,
					HtpasswdUsers: htpasswdMap,
				}
				err = idpService.htpasswd.Apply(idpParam, true)
				Expect(err).ToNot(HaveOccurred())

				By("Update htpasswd idp on the second user password")
				newPassword2 := h.GenerateRandomStringWithSymbols(15)
				htpasswdMap = []interface{}{map[string]string{"username": userName, "password": password},
					map[string]string{"username": userName2, "password": newPassword2},
					map[string]string{"username": userName3, "password": password3}}
				idpParam = &exe.IDPArgs{
					ClusterID:     clusterID,
					Name:          htpIDPName,
					HtpasswdUsers: htpasswdMap,
				}
				err = idpService.htpasswd.Apply(idpParam, true)
				Expect(err).ToNot(HaveOccurred())

				By("Check resource state file is updated")
				resource, err = h.GetResource(CON.HtpasswdDir, "rhcs_identity_provider", "htpasswd_idp")
				Expect(err).ToNot(HaveOccurred())
				Expect(resource).To(MatchJQ(fmt.Sprintf(`.instances[0].attributes.htpasswd.users[] | select(.username == "%s") .password`, userName2), newPassword2))

				By("List existing HtpasswdUsers and compare to the created one")
				htpasswdUsersList, _ = cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpID.ID)
				Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
				Expect(htpasswdUsersList.Items().Len()).To(Equal(3))

				respUserSlice := htpasswdUsersList.Items().Slice()
				userNameToCheck := map[string]bool{
					userName:  true,
					userName2: true,
					userName3: true,
				}
				for _, user := range respUserSlice {
					_, existed := userNameToCheck[user.Username()]
					Expect(existed).To(BeTrue())
				}
			})
		})

		Context("LDAP", func() {
			BeforeEach(func() {
				userName = "newton"
				password = "password"
				idpService.ldap = *exe.NewIDPService(CON.LdapDir) // init new ldap service
			})

			AfterEach(func() {
				err := idpService.ldap.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeed - [id:63332]", ci.High,
				ci.Exclude,
				func() {
					By("Create LDAP idp for an existing cluster")

					idpParam := &exe.IDPArgs{
						ClusterID:  clusterID,
						Name:       "OCP-63332-ldap-idp-test",
						CA:         "",
						URL:        CON.LdapURL,
						Attributes: make(map[string]interface{}),
						Insecure:   true,
					}
					err := idpService.ldap.Apply(idpParam, false)
					Expect(err).ToNot(HaveOccurred())

					By("Login with created ldap idp")
					// this condition is for cases where the cluster profile
					// has private_link enabled, then regular login won't work
					if !profile.IsPrivateLink() {
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
								fmt.Sprintf("--kubeconfig %s", path.Join(CON.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
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
		Context("GitLab", func() {
			BeforeEach(func() {
				gitlabIDPClientId = h.GenerateRandomStringWithSymbols(20)
				gitlabIDPClientSecret = h.GenerateRandomStringWithSymbols(30)
				idpService.gitlab = *exe.NewIDPService(CON.GitlabDir) // init new gitlab service
			})

			AfterEach(func() {
				err := idpService.gitlab.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeed - [id:64028]", ci.High, func() {
				By("Create GitLab idp for an existing cluster")

				idpParam := &exe.IDPArgs{
					ClusterID:    clusterID,
					Name:         "OCP-64028-gitlab-idp-test",
					ClientID:     gitlabIDPClientId,
					ClientSecret: gitlabIDPClientSecret,
					URL:          CON.GitLabURL,
				}
				err := idpService.gitlab.Apply(idpParam, false)
				Expect(err).ToNot(HaveOccurred())

				By("Check gitlab idp created for the cluster")
				idpID, err := idpService.gitlab.Output()
				Expect(err).ToNot(HaveOccurred())

				resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Status()).To(Equal(http.StatusOK))
			})
		})
		Context("GitHub", func() {
			BeforeEach(func() {

				githubIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
				githubIDPClientId = h.GenerateRandomStringWithSymbols(30)
				idpService.github = *exe.NewIDPService(CON.GithubDir) // init new github service
			})

			AfterEach(func() {
				err := idpService.github.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeed - [id:64027]", ci.High, func() {
				By("Create GitHub idp for an existing cluster")

				idpParam := &exe.IDPArgs{
					ClusterID:     clusterID,
					Name:          "OCP-64027-github-idp-test",
					ClientID:      githubIDPClientId,
					ClientSecret:  githubIDPClientSecret,
					Organizations: CON.Organizations,
				}
				err := idpService.github.Apply(idpParam, false)
				Expect(err).ToNot(HaveOccurred())

				By("Check github idp created for the cluster")
				idpID, err := idpService.github.Output()
				Expect(err).ToNot(HaveOccurred())

				resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Status()).To(Equal(http.StatusOK))
			})
		})
		Context("Google", func() {
			BeforeEach(func() {

				googleIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
				googleIDPClientId = h.GenerateRandomStringWithSymbols(30)
				idpService.google = *exe.NewIDPService(CON.GoogleDir) // init new google service
			})

			AfterEach(func() {
				err := idpService.google.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeeed - [id:64029]", ci.High, func() {
				By("Create Google idp for an existing cluster")

				idpParam := &exe.IDPArgs{
					ClusterID:    clusterID,
					Name:         "OCP-64029-google-idp-test",
					ClientID:     googleIDPClientId,
					ClientSecret: googleIDPClientSecret,
					HostedDomain: CON.HostedDomain,
				}
				err := idpService.google.Apply(idpParam, false)
				Expect(err).ToNot(HaveOccurred())

				By("Check google idp created for the cluster")
				idpID, err := idpService.google.Output()
				Expect(err).ToNot(HaveOccurred())

				resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Status()).To(Equal(http.StatusOK))
			})
		})
		Context("Multi IDPs", func() {
			BeforeEach(func() {

				if profile.IsPrivateLink() {
					Skip("private_link is enabled, skipping test.")
				}

				userName = "newton"
				password = "password"
				googleIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
				googleIDPClientId = h.GenerateRandomStringWithSymbols(30)
				idpService.htpasswd = *exe.NewIDPService(CON.HtpasswdDir)  // init new htpasswd service
				idpService.multi_idp = *exe.NewIDPService(CON.MultiIDPDir) // init multi-idp service
			})

			It("will succeed - [id:64030]", ci.Medium, ci.Exclude, func() {

				By("Applying google & ldap idps users using terraform")

				idpParam := &exe.IDPArgs{
					ClusterID:    clusterID,
					Name:         "OCP-64030",
					ClientID:     googleIDPClientId,
					ClientSecret: googleIDPClientSecret,
					HostedDomain: CON.HostedDomain,
					CA:           "",
					URL:          CON.LdapURL,
					Attributes:   make(map[string]interface{}),
					Insecure:     true,
				}

				err := idpService.multi_idp.Apply(idpParam, false)
				Expect(err).ToNot(HaveOccurred())

				By("Login to the ldap user created with terraform")
				By("& cluster-admin user created on cluster deployment")

				resp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				server := resp.Body().API().URL()

				ocAtter := &openshift.OcAttributes{
					Server:    server,
					Username:  userName,
					Password:  password,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(CON.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
					},
					Timeout: 7,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())

				if !profile.AdminEnabled {
					Skip("The test configured only for cluster admin profile")
				}

				// login to the cluster using cluster-admin creds
				username := CON.ClusterAdminUser
				password := h.GetClusterAdminPassword()
				Expect(password).ToNot(BeEmpty())

				ocAtter = &openshift.OcAttributes{
					Server:    server,
					Username:  username,
					Password:  password,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(CON.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, username))),
					},
					Timeout: 10,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())

				defer func() {
					err := idpService.multi_idp.Destroy()
					Expect(err).ToNot(HaveOccurred())
				}()
			})
			It("with multiple users will reconcile a multiuser config - [id:66408]", ci.Medium, ci.Exclude, func() {

				userName = "first_user"
				password = h.GenerateRandomStringWithSymbols(15)
				By("Create 3 htpasswd users for existing cluster")
				htpasswdMap = []interface{}{
					map[string]string{"username": userName,
						"password": password},
					map[string]string{"username": "second_user",
						"password": h.GenerateRandomStringWithSymbols(15)},
					map[string]string{"username": "third_user",
						"password": h.GenerateRandomStringWithSymbols(15)}}

				idpParam := &exe.IDPArgs{
					ClusterID:     clusterID,
					Name:          "OCP-66408-htpasswd-multi-test",
					HtpasswdUsers: htpasswdMap,
				}
				err := idpService.htpasswd.Apply(idpParam, false)
				Expect(err).ToNot(HaveOccurred())

				By("Login to the cluster with one of the users created")
				resp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				server := resp.Body().API().URL()

				ocAtter := &openshift.OcAttributes{
					Server:    server,
					Username:  userName,
					Password:  password,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(CON.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
					},
					Timeout: 10,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())
				idpID, _ := idpService.htpasswd.Output()

				By("Delete one of the users using backend api")
				_, err = cms.DeleteIDP(ci.RHCSConnection, clusterID, idpID.ID)
				Expect(err).ToNot(HaveOccurred())

				// wait few minutes before trying to create the resource again
				time.Sleep(time.Minute * 5)

				By("Re-run terraform apply on the same resources")
				err = idpService.htpasswd.Apply(idpParam, false)
				Expect(err).ToNot(HaveOccurred())

				By("Re-login terraform apply on the same resources")

				// note - this step failes randmonly.
				// hence, the test is currently skipped for ci
				ocAtter = &openshift.OcAttributes{
					Server:    server,
					Username:  userName,
					Password:  password,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(CON.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
					},
					Timeout: 10,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())

				defer func() {
					err = idpService.htpasswd.Destroy()
					Expect(err).ToNot(HaveOccurred())
				}()
			})
		})
	})

	Describe("validate", func() {

		BeforeEach(func() {
			userName = "jacko"
			password = h.GenerateRandomStringWithSymbols(15)
			googleIdpName = "google-idp"
			gitLabIdpName = "gitlab-idp"
			gitlabIDPClientId = h.GenerateRandomStringWithSymbols(20)
			gitlabIDPClientSecret = h.GenerateRandomStringWithSymbols(30)
			githubIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
			githubIDPClientId = h.GenerateRandomStringWithSymbols(30)
			googleIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
			googleIDPClientId = h.GenerateRandomStringWithSymbols(30)

			idpService.htpasswd = *exe.NewIDPService(CON.HtpasswdDir) // init new htpasswd service
			idpService.ldap = *exe.NewIDPService(CON.LdapDir)         // init new ldap service
			idpService.github = *exe.NewIDPService(CON.GithubDir)     // init new github service
			idpService.gitlab = *exe.NewIDPService(CON.GitlabDir)     // init new gitlab service
			idpService.google = *exe.NewIDPService(CON.GoogleDir)     // init new google service
			importService = *exe.NewImportService(CON.ImportResourceDir)
		})

		It("the mandatory idp's attributes must be set - [id:68939]", ci.Medium, func() {

			By("Create htpasswd idp without/empty name field")
			idpParam := &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "",
				HtpasswdUsers: htpasswdMap,
			}
			err := idpService.htpasswd.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring(
					"The root module input variable \"name\" is not set, and has no default value"))

			By("Create htpasswd idp without/empty username field")
			userName = ""
			htpasswdMap = []interface{}{map[string]string{
				"username": userName, "password": password}}
			idpParam = &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "htpasswd-idp-test",
				HtpasswdUsers: htpasswdMap,
			}

			err = idpService.htpasswd.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("htpasswd.users[0].username username may not be empty/blank string"))

			By("Create htpasswd idp without/empty password field")
			userName = "jacko"
			htpasswdMap = []interface{}{map[string]string{
				"username": userName}}
			idpParam = &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "htpasswd-idp-test",
				HtpasswdUsers: htpasswdMap,
			}
			err = idpService.htpasswd.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("attribute \"password\" is required"))

			By("Create ldap idp without/empty name field")
			idpParam = &exe.IDPArgs{
				ClusterID:  clusterID,
				Name:       "",
				CA:         "",
				URL:        CON.LdapURL,
				Attributes: make(map[string]interface{}),
				Insecure:   true,
			}
			err = idpService.ldap.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring(
					"The root module input variable \"name\" is not set, " +
						"and has no default value"))

			By("Create ldap idp without url field")
			idpParam = &exe.IDPArgs{
				ClusterID:  clusterID,
				Name:       "ldap-idp-test",
				CA:         "",
				URL:        "",
				Attributes: make(map[string]interface{}),
				Insecure:   true,
			}
			err = idpService.ldap.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring(
				"Must set a configuration value for the ldap.url attribute"))

			By("Create ldap idp without attributes field")
			idpParam = &exe.IDPArgs{
				ClusterID: clusterID,
				Name:      "ldap-idp-test",
				CA:        "",
				Insecure:  true,
			}

			err = idpService.ldap.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring("provider has marked it as required"))

			By("Create github idp without/empty name field")
			idpParam = &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "",
				ClientID:      githubIDPClientId,
				ClientSecret:  githubIDPClientSecret,
				Organizations: CON.Organizations,
			}
			err = idpService.github.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring(
					"The root module input variable \"name\" is not set, and has no default value"))

			By("Create github idp without/empty client_id field")
			idpParam = &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "github-idp-test",
				ClientID:      "",
				ClientSecret:  githubIDPClientSecret,
				Organizations: CON.Organizations,
			}
			err = idpService.github.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring("No value for required variable"))

			By("Create github idp without/empty client_secret field")
			idpParam = &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "github-idp-test",
				ClientID:      githubIDPClientId,
				ClientSecret:  "",
				Organizations: CON.Organizations,
			}
			err = idpService.github.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring("No value for required variable"))

			By("Create gitlab idp without/empty name field")
			idpParam = &exe.IDPArgs{
				ClusterID:    clusterID,
				Name:         "",
				ClientID:     gitlabIDPClientId,
				ClientSecret: gitlabIDPClientSecret,
				URL:          CON.GitLabURL,
			}
			err = idpService.gitlab.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring(
					"The root module input variable \"name\" is not set, and has no default value"))

			By("Create gitlab idp without/empty client_id field")
			idpParam = &exe.IDPArgs{
				ClusterID:    clusterID,
				Name:         "gitlab-idp-test",
				ClientID:     "",
				ClientSecret: gitlabIDPClientSecret,
				URL:          CON.GitLabURL,
			}
			err = idpService.gitlab.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring("provider has marked it as required"))

			By("Create gitlab idp without/empty client_secret field")
			idpParam = &exe.IDPArgs{
				ClusterID:    clusterID,
				Name:         "gitlab-idp-test",
				ClientID:     gitlabIDPClientId,
				ClientSecret: "",
				URL:          CON.GitLabURL,
			}
			err = idpService.gitlab.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring("provider has marked it as required"))

			By("Create gitlab idp without url field")
			idpParam = &exe.IDPArgs{
				ClusterID:    clusterID,
				Name:         "gitlab-idp-test",
				URL:          "",
				ClientID:     gitlabIDPClientId,
				ClientSecret: gitlabIDPClientSecret,
			}
			err = idpService.gitlab.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring("Must set a configuration value for the gitlab.url"))

			By("Create google idp without/empty name field")
			idpParam = &exe.IDPArgs{
				ClusterID:    clusterID,
				Name:         "",
				ClientID:     googleIDPClientId,
				ClientSecret: googleIDPClientSecret,
				HostedDomain: CON.HostedDomain,
			}
			err = idpService.google.Apply(idpParam, false)
			Expect(err.Error()).Should(
				ContainSubstring(
					"The root module input variable \"name\" is not set, and has no default value"))

			By("Create google idp without/empty client_id field")
			idpParam = &exe.IDPArgs{
				ClusterID:    clusterID,
				Name:         "google-idp-test",
				ClientID:     "",
				ClientSecret: googleIDPClientSecret,
				HostedDomain: CON.HostedDomain,
			}
			err = idpService.google.Apply(idpParam, false)
			Expect(err.Error()).Should(
				ContainSubstring("provider has marked it as required"))

			By("Create google idp without/empty client_secret field")
			idpParam = &exe.IDPArgs{
				ClusterID:    clusterID,
				Name:         "google-idp-test",
				ClientID:     googleIDPClientId,
				ClientSecret: "",
				HostedDomain: CON.HostedDomain,
			}
			err = idpService.google.Apply(idpParam, false)
			Expect(err.Error()).Should(
				ContainSubstring("provider has marked it as required"))

		})

		It("htpasswd with empty user-password list will fail - [id:66409]", ci.Medium, func() {
			By("Validate idp can't be created with empty htpasswdMap")
			htpasswdMap = []interface{}{map[string]string{}}

			idpParam := &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "OCP-66409-htpasswd-idp-test",
				HtpasswdUsers: htpasswdMap,
			}
			err := idpService.htpasswd.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring(
					"attributes \"password\" and \"username\" are required"))
		})

		It("htpasswd password policy - [id:66410]", ci.Medium, func() {

			var usernameInvalid = "userWithInvalidPassword"
			var passwordInvalid string

			By("Validate idp can't be created with password less than 14")

			passwordInvalid = h.GenerateRandomStringWithSymbols(3)
			htpasswdMap = []interface{}{map[string]string{
				"username": userName, "password": password},
				map[string]string{"username": usernameInvalid,
					"password": passwordInvalid}}

			idpParam := &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "OCP-66410-htpasswd-idp-test",
				HtpasswdUsers: htpasswdMap,
			}
			err := idpService.htpasswd.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring(
					"password string length must be at least 14"))

			By("Validate idp can't be created without upercase letter in password")

			passwordInvalid = h.Subfix(3)
			htpasswdMap = []interface{}{map[string]string{
				"username": userName, "password": password},
				map[string]string{"username": usernameInvalid,
					"password": passwordInvalid}}
			idpParam = &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "OCP-66410-htpasswd-idp-test",
				HtpasswdUsers: htpasswdMap,
			}
			err = idpService.htpasswd.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring(
					"password must contain uppercase"))
		})

		It("htpasswd with duplicate usernames will fail - [id:66411]", ci.Medium, func() {
			By("Create 2 htpasswd idps with the same username")
			htpasswdMap = []interface{}{map[string]string{
				"username": userName, "password": password},
				map[string]string{"username": userName, "password": password}}

			idpParam := &exe.IDPArgs{
				ClusterID:     clusterID,
				Name:          "OCP-66411-htpasswd-idp-test",
				HtpasswdUsers: htpasswdMap,
			}
			err := idpService.htpasswd.Apply(idpParam, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).Should(
				ContainSubstring(
					"Usernames in HTPasswd user list must be unique"))
		})

		It("resource can be imported - [id:65981]",
			ci.Medium, ci.FeatureImport, func() {
				defer func() {
					_, importErr := importService.Destroy()
					Expect(importErr).ToNot(HaveOccurred())
				}()

				By("Create sample idps to test the import functionality")
				idpParam := &exe.IDPArgs{
					ClusterID:    clusterID,
					Name:         googleIdpName,
					ClientID:     googleIDPClientId,
					ClientSecret: googleIDPClientSecret,
					HostedDomain: CON.HostedDomain,
				}
				Expect(idpService.google.Apply(idpParam, false)).To(Succeed())

				idpParam = &exe.IDPArgs{
					ClusterID:    clusterID,
					Name:         gitLabIdpName,
					ClientID:     gitlabIDPClientId,
					ClientSecret: gitlabIDPClientSecret,
					URL:          CON.GitLabURL,
				}
				Expect(idpService.gitlab.Apply(idpParam, false)).To(Succeed())

				By("Run the command to import the idp")
				importParam := &exe.ImportArgs{
					ClusterID:    clusterID,
					ResourceKind: "rhcs_identity_provider",
					ResourceName: "idp_google_import",
					ObjectName:   googleIdpName,
				}
				Expect(importService.Import(importParam)).To(Succeed())

				By("Check resource state - import command succeeded")
				output, err := importService.ShowState(importParam)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(ContainSubstring(googleIDPClientId))
				Expect(output).To(ContainSubstring(CON.HostedDomain))

				By("Validate terraform import with no idp object name returns error")
				var unknownIdpName = "unknown_idp_name"
				importParam = &exe.ImportArgs{
					ClusterID:    clusterID,
					ResourceKind: "rhcs_identity_provider",
					ResourceName: "idp_google_import",
					ObjectName:   unknownIdpName,
				}

				err = importService.Import(importParam)
				Expect(err.Error()).To(ContainSubstring("identity provider '%s' not found", unknownIdpName))

				By("Validate terraform import with no clusterID returns error")

				var unknownClusterID = h.GenerateRandomStringWithSymbols(20)
				importParam = &exe.ImportArgs{
					ClusterID:    unknownClusterID,
					ResourceKind: "rhcs_identity_provider",
					ResourceName: "idp_gitlab_import",
					ObjectName:   gitLabIdpName,
				}

				err = importService.Import(importParam)
				Expect(err.Error()).To(ContainSubstring("Cluster %s not found", unknownClusterID))

			})
	})

	Describe("reconciliation", func() {
		var (
			gitLabIdpName         string
			gitlabIDPClientId     string
			gitlabIDPClientSecret string
			gitlabIdpOcmAPI       *cmv1.IdentityProviderBuilder
		)

		BeforeEach(func() {
			idpService.ldap = *exe.NewIDPService(CON.LdapDir)     // init new ldap service
			idpService.gitlab = *exe.NewIDPService(CON.GitlabDir) // init new gitlab service
			gitlabIdpOcmAPI = cmv1.NewIdentityProvider()
			gitlabIDPClientId = h.GenerateRandomStringWithSymbols(20)
			gitlabIDPClientSecret = h.GenerateRandomStringWithSymbols(30)
			userName = "newton"
			password = "password"
		})

		It("verify basic flow - [id:65808]",
			ci.Medium, ci.Exclude, func() {

				By("Create ldap idp user")
				idpParam := &exe.IDPArgs{
					ClusterID:  clusterID,
					Name:       "OCP-65808-ldap-reconcil",
					CA:         "",
					URL:        CON.LdapURL,
					Attributes: make(map[string]interface{}),
					Insecure:   true,
				}
				Expect(idpService.ldap.Apply(idpParam, false)).To(Succeed())
				idpID, _ := idpService.ldap.Output()

				By("Login to the ldap user")
				resp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				server := resp.Body().API().URL()

				ocAtter := &openshift.OcAttributes{
					Server:    server,
					Username:  userName,
					Password:  password,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(CON.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
					},
					Timeout: 7,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())

				By("Delete ldap idp by OCM API")
				cms.DeleteIDP(ci.RHCSConnection, clusterID, idpID.ID)
				_, err = cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(
					"Identity provider ID '%s' for cluster '%s' not found", idpID.ID, clusterID),
				)

				By("Re-apply the idp resource")
				Expect(idpService.ldap.Apply(idpParam, false)).To(Succeed())

				// Delete ldap idp from existing cluster after test ended
				defer func() {
					err := idpService.ldap.Destroy()
					Expect(err).ToNot(HaveOccurred())
				}()

				By("re-login with the ldpa idp username/password")
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())

			})

		It("try to restore/duplicate an existing IDP - [id:65816]",
			ci.Medium, func() {

				gitLabIdpName = "OCP-65816-gitlab-idp-reconcil"

				By("Create gitlab idp for existing cluster")
				idpParam := &exe.IDPArgs{
					ClusterID:    clusterID,
					Name:         gitLabIdpName,
					ClientID:     gitlabIDPClientId,
					ClientSecret: gitlabIDPClientSecret,
					URL:          CON.GitLabURL,
				}
				Expect(idpService.gitlab.Apply(idpParam, false)).To(Succeed())
				idpID, _ := idpService.gitlab.Output()

				By("Delete gitlab idp using OCM API")
				cms.DeleteIDP(ci.RHCSConnection, clusterID, idpID.ID)
				_, err = cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(
					"Identity provider ID '%s' for cluster '%s' not found", idpID.ID, clusterID),
				)

				By("Create gitlab idp using OCM api with the same parameters")
				requestBody, _ := gitlabIdpOcmAPI.Type("GitlabIdentityProvider").
					Name(gitLabIdpName).
					Gitlab(cmv1.NewGitlabIdentityProvider().
						ClientID(gitlabIDPClientId).
						ClientSecret(gitlabIDPClientSecret).
						URL(CON.GitLabURL)).
					MappingMethod("claim").
					Build()
				res, err := cms.CreateClusterIDP(ci.RHCSConnection, clusterID, requestBody)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Status()).To(Equal(http.StatusCreated))

				// Delete gitlab idp from existing cluster after test end
				defer cms.DeleteIDP(ci.RHCSConnection, clusterID, res.Body().ID())

				By("Re-apply gitlab idp using tf manifests with same ocm api args")
				err = idpService.gitlab.Apply(idpParam, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(
					"Identity Provider Name\n%s already exists", gitLabIdpName),
				)
			})

		It("try to restore to an updated IDP config - [id:65814]",
			ci.Medium, func() {

				gitLabIdpName = "OCP-65814-gitlab-idp-reconcil"

				By("Create gitlab idp for existing cluster")
				idpParam := &exe.IDPArgs{
					ClusterID:    clusterID,
					Name:         gitLabIdpName,
					ClientID:     gitlabIDPClientId,
					ClientSecret: gitlabIDPClientSecret,
					URL:          CON.GitLabURL,
				}
				Expect(idpService.gitlab.Apply(idpParam, false)).To(Succeed())
				idpID, _ := idpService.gitlab.Output()

				// Delete gitlab idp using OCM API after test end
				defer func() {
					cms.DeleteIDP(ci.RHCSConnection, clusterID, idpID.ID)
					_, err = cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpID.ID)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring(
						"Identity provider ID '%s' for cluster '%s' not found", idpID.ID, clusterID),
					)
				}()

				By("Edit gitlab idp using ocm api")
				newClientSecret := h.GenerateRandomStringWithSymbols(30)
				requestBody, _ := gitlabIdpOcmAPI.Type("GitlabIdentityProvider").
					Gitlab(cmv1.NewGitlabIdentityProvider().
						ClientID(gitlabIDPClientId).
						ClientSecret(newClientSecret).
						URL(CON.GitLabURL)).
					MappingMethod("claim").
					Build()

				resp, err := cms.PatchIDP(ci.RHCSConnection, clusterID, idpID.ID, requestBody)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Status()).To(Equal(http.StatusOK))

				// update is currently not supported for idp :: OCM-4622
				By("Update and apply idp using terraform")
				newClientSecret = h.GenerateRandomStringWithSymbols(30)
				idpParam = &exe.IDPArgs{
					ClusterID:    clusterID,
					Name:         gitLabIdpName,
					ClientID:     gitlabIDPClientId,
					ClientSecret: newClientSecret,
					URL:          CON.GitLabURL,
				}

				err = idpService.gitlab.Apply(idpParam, true)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(
					"This RHCS provider version does not support updating an existing IDP"),
				)
			})
	})
})
