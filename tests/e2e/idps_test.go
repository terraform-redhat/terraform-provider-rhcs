package e2e

import (
	"fmt"
	"net/http"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ci "github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	con "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	exe "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	H "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
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
			multi_idp,
			openid exe.IDPService
		}

		var profile *ci.Profile

		var idpService IDPServices
		var importService exe.ImportService
		var htpasswdMap = []interface{}{map[string]string{}}

		var userName, password,
			googleIDPClientSecret, googleIDPClientId,
			gitlabIDPClientSecret, gitlabIDPClientId,
			githubIDPClientSecret, githubIDPClientId string
		BeforeEach(func() {
			profile = ci.LoadProfileYamlFileByENV()
		})

		Describe("IDP Positive scenario test cases", func() {
			Context("Htpasswd IDP test cases", func() {

				BeforeEach(func() {
					userName = "jacko"
					password = h.GenerateRandomStringWithSymbols(15)
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
			Context("LDAP IDP test cases", func() {
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
								ClusterID:  clusterID,
								Name:       "OCP-63332-ldap-idp-test",
								CA:         "",
								URL:        con.LdapURL,
								Attributes: make(map[string]interface{}),
								Insecure:   true,
							}
							err := idpService.ldap.Apply(idpParam, false)
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
			Context("GitLab IDP test cases", func() {
				BeforeEach(func() {
					gitlabIDPClientId = h.GenerateRandomStringWithSymbols(20)
					gitlabIDPClientSecret = h.GenerateRandomStringWithSymbols(30)
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
							ClusterID:    clusterID,
							Name:         "OCP-64028-gitlab-idp-test",
							ClientID:     gitlabIDPClientId,
							ClientSecret: gitlabIDPClientSecret,
							URL:          con.GitLabURL,
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
			})
			Context("GitHub IDP test cases", func() {
				BeforeEach(func() {

					githubIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
					githubIDPClientId = h.GenerateRandomStringWithSymbols(30)
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
							ClusterID:     clusterID,
							Name:          "OCP-64027-github-idp-test",
							ClientID:      githubIDPClientId,
							ClientSecret:  githubIDPClientSecret,
							Organizations: con.Organizations,
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
			})
			Context("Google IDP test cases", func() {
				BeforeEach(func() {

					googleIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
					googleIDPClientId = h.GenerateRandomStringWithSymbols(30)
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
							ClusterID:    clusterID,
							Name:         "OCP-64029-google-idp-test",
							ClientID:     googleIDPClientId,
							ClientSecret: googleIDPClientSecret,
							HostedDomain: con.HostedDomain,
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
			})
			Context("Multi IDPs applying scenrios", func() {
				BeforeEach(func() {

					if profile.PrivateLink {
						Skip("private_link is enabled, skipping test.")
					}

					userName = "newton"
					password = "password"
					googleIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
					googleIDPClientId = h.GenerateRandomStringWithSymbols(30)
					idpService.htpasswd = *exe.NewIDPService(con.HtpasswdDir)  // init new htpasswd service
					idpService.multi_idp = *exe.NewIDPService(con.MultiIDPDir) // init multi-idp service
				})

				Context("Author:smiron-Medium-OCP-64030 @OCP-64030 @smiron", func() {
					It("OCP-64030 - Provision multiple IDPs against cluster using TF", ci.Day2, ci.Medium, ci.FeatureIDP, ci.Exclude, func() {

						By("Applying google & ldap idps users using terraform")

						idpParam := &exe.IDPArgs{
							ClusterID:    clusterID,
							Name:         "OCP-64030",
							ClientID:     googleIDPClientId,
							ClientSecret: googleIDPClientSecret,
							HostedDomain: con.HostedDomain,
							CA:           "",
							URL:          con.LdapURL,
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
								fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
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
						password := H.GetClusterAdminPassword()
						Expect(password).ToNot(BeEmpty())

						ocAtter = &openshift.OcAttributes{
							Server:    server,
							Username:  username,
							Password:  password,
							ClusterID: clusterID,
							AdditioanlFlags: []string{
								"--insecure-skip-tls-verify",
								fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, username))),
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
				})
				Context("Author:smiron-Medium-OCP-66408 @OCP-66408 @smiron", func() {
					It("OCP-66408 - htpasswd multiple users - reconcile a multiuser config", ci.Day2, ci.Medium, ci.FeatureIDP, ci.Exclude, func() {

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
								fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
							},
							Timeout: 10,
						}
						_, err = openshift.OcLogin(*ocAtter)
						Expect(err).ToNot(HaveOccurred())
						idpID, _ := idpService.htpasswd.Output()

						By("Delete one of the users using backend api")
						fmt.Println("blaaa:: ", idpID.ID)
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
								fmt.Sprintf("--kubeconfig %s", path.Join(con.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, userName))),
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
		})
		Describe("IDP Negative scenario test cases", func() {

			BeforeEach(func() {

				userName = "jacko"
				password = h.GenerateRandomStringWithSymbols(15)
				gitlabIDPClientId = h.GenerateRandomStringWithSymbols(20)
				gitlabIDPClientSecret = h.GenerateRandomStringWithSymbols(30)
				githubIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
				githubIDPClientId = h.GenerateRandomStringWithSymbols(30)
				googleIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
				googleIDPClientId = h.GenerateRandomStringWithSymbols(30)

				idpService.htpasswd = *exe.NewIDPService(con.HtpasswdDir) // init new htpasswd service
				idpService.ldap = *exe.NewIDPService(con.LdapDir)         // init new ldap service
				idpService.github = *exe.NewIDPService(con.GithubDir)     // init new github service
				idpService.gitlab = *exe.NewIDPService(con.GitlabDir)     // init new gitlab service
				idpService.google = *exe.NewIDPService(con.GoogleDir)     // init new google service
			})

			Context("Author:smiron-Medium-OCP-68939 @OCP-68939 @smiron", func() {
				It("OCP-68939 - Validate that the mandatory idp's attributes must be set", ci.Day2, ci.Medium, ci.FeatureIDP, func() {

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
					Expect(err.Error()).Should(ContainSubstring("Attribute 'username' is mandatory"))

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
						URL:        con.LdapURL,
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
						Organizations: con.Organizations,
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
						Organizations: con.Organizations,
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
						Organizations: con.Organizations,
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
						URL:          con.GitLabURL,
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
						URL:          con.GitLabURL,
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
						URL:          con.GitLabURL,
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
						HostedDomain: con.HostedDomain,
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
						HostedDomain: con.HostedDomain,
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
						HostedDomain: con.HostedDomain,
					}
					err = idpService.google.Apply(idpParam, false)
					Expect(err.Error()).Should(
						ContainSubstring("provider has marked it as required"))

				})
			})
			Context("Author:smiron-Medium-OCP-66409 @OCP-66409 @smiron", func() {
				It("OCP-66409 - htpasswd multiple users: empty user-password list", ci.Day2, ci.Medium, ci.FeatureIDP, func() {
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
			})
			Context("Author:smiron-Medium-OCP-66410 @OCP-66410 @smiron", func() {
				It("OCP-66410 - htpasswd multiple users - password policy violation", ci.Day2, ci.Medium, ci.FeatureIDP, func() {

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
					fmt.Println("blaaa::", passwordInvalid)
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
			})
			Context("Author:smiron-Medium-OCP-66411 @OCP-66411 @smiron", func() {
				It("OCP-66411 - htpasswd multiple users: duplicate user names", ci.Day2, ci.Medium, ci.FeatureIDP, func() {
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
			})
		})

		Describe("Validate terraform Import operations", func() {
			var (
				googleIdpName         = "google-idp"
				gitLabIdpName         = "gitlab-idp"
				googleIDPClientSecret string
				googleIDPClientId     string
				gitlabIDPClientId     string
				gitlabIDPClientSecret string
			)

			BeforeEach(func() {
				idpService.google = *exe.NewIDPService(con.GoogleDir)        // init new google service
				idpService.gitlab = *exe.NewIDPService(con.GitlabDir)        // init new gitlab service
				importService = *exe.NewImportService(con.ImportResourceDir) // init new import service
			})

			AfterEach(func() {
				err := idpService.google.Destroy()
				Expect(err).ToNot(HaveOccurred())

				err = idpService.gitlab.Destroy()
				Expect(err).ToNot(HaveOccurred())
			})

			Context("Author:smiron-Medium-OCP-65981 @OCP-65981 @smiron", func() {
				It("OCP-65981 - rhcs_identity_provider resource can be imported by the terraform import command",
					ci.Day2, ci.Medium, ci.FeatureIDP, ci.FeatureImport, func() {

						By("Create sample idps to test the import functionality")
						googleIDPClientSecret = h.GenerateRandomStringWithSymbols(20)
						googleIDPClientId = h.GenerateRandomStringWithSymbols(30)
						gitlabIDPClientId = h.GenerateRandomStringWithSymbols(20)
						gitlabIDPClientSecret = h.GenerateRandomStringWithSymbols(30)

						idpParam := &exe.IDPArgs{
							ClusterID:    clusterID,
							Name:         googleIdpName,
							ClientID:     googleIDPClientId,
							ClientSecret: googleIDPClientSecret,
							HostedDomain: con.HostedDomain,
						}
						Expect(idpService.google.Apply(idpParam, false)).To(Succeed())

						idpParam = &exe.IDPArgs{
							ClusterID:    clusterID,
							Name:         gitLabIdpName,
							ClientID:     gitlabIDPClientId,
							ClientSecret: gitlabIDPClientSecret,
							URL:          con.GitLabURL,
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
						Expect(output).To(ContainSubstring(con.HostedDomain))

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
						Expect(err.Error()).To(ContainSubstring("'%s' not found", unknownClusterID))

					})
			})
		})
	})
})
