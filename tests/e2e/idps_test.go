package e2e

import (
	"fmt"
	"net/http"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cmsv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	. "github.com/openshift-online/ocm-sdk-go/testing"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/ci"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/cms"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/openshift"
)

// all identity providers - declared for future cases
type IDPServices struct {
	htpasswd,
	github,
	gitlab,
	google,
	ldap,
	multi_idp,
	openid exec.IDPService
}

var (
	defaultHTPUsername = "my-admin-user"
	defaultHTPPassword = helper.GenerateRandomStringWithSymbols(15)

	defaultLDAPUsername = "newton"
	defaultLDAPPassword = "password"

	defaultGoogleIDPClientId     = helper.GenerateRandomStringWithSymbols(30)
	defaultGoogleIDPClientSecret = helper.GenerateRandomStringWithSymbols(20)

	defaultGitlabIDPClientId     = helper.GenerateRandomStringWithSymbols(20)
	defaultGitlabIDPClientSecret = helper.GenerateRandomStringWithSymbols(30)

	defaultGithubIDPClientId     = helper.GenerateRandomStringWithSymbols(30)
	defaultGithubIDPClientSecret = helper.GenerateRandomStringWithSymbols(20)
)

func getDefaultHTPasswordArgs(idpName string) *exec.IDPArgs {
	return &exec.IDPArgs{
		ClusterID: helper.StringPointer(clusterID),
		Name:      helper.StringPointer(idpName),
		HtpasswdUsers: &[]exec.HTPasswordUser{
			{
				Username: helper.StringPointer(defaultHTPUsername),
				Password: helper.StringPointer(defaultHTPPassword),
			},
		},
	}
}
func getDefaultLDAPArgs(idpName string) *exec.IDPArgs {
	return &exec.IDPArgs{
		ClusterID:      helper.StringPointer(clusterID),
		Name:           helper.StringPointer(idpName),
		CA:             helper.EmptyStringPointer,
		URL:            helper.StringPointer(constants.LdapURL),
		LDAPAttributes: &exec.LDAPAttributes{},
		Insecure:       helper.BoolPointer(true),
	}
}
func getDefaultGitHubArgs(idpName string) *exec.IDPArgs {
	return &exec.IDPArgs{
		ClusterID:     helper.StringPointer(clusterID),
		Name:          helper.StringPointer(idpName),
		ClientID:      helper.StringPointer(defaultGithubIDPClientId),
		ClientSecret:  helper.StringPointer(defaultGithubIDPClientSecret),
		Organizations: helper.StringSlicePointer(constants.Organizations),
	}
}
func getDefaultGitlabArgs(idpName string) *exec.IDPArgs {
	return &exec.IDPArgs{
		ClusterID:    helper.StringPointer(clusterID),
		Name:         helper.StringPointer(idpName),
		ClientID:     helper.StringPointer(defaultGitlabIDPClientId),
		ClientSecret: helper.StringPointer(defaultGitlabIDPClientSecret),
		URL:          helper.StringPointer(constants.GitLabURL),
	}
}
func getDefaultGoogleArgs(idpName string) *exec.IDPArgs {
	return &exec.IDPArgs{
		ClusterID:    helper.StringPointer(clusterID),
		Name:         helper.StringPointer(idpName),
		ClientID:     helper.StringPointer(defaultGoogleIDPClientId),
		ClientSecret: helper.StringPointer(defaultGoogleIDPClientSecret),
		HostedDomain: helper.StringPointer(constants.HostedDomain),
	}
}

var _ = Describe("Identity Providers", ci.Day2, ci.FeatureIDP, func() {
	defer GinkgoRecover()

	var profile *ci.Profile
	var idpServices = IDPServices{}

	BeforeEach(func() {
		profile = ci.LoadProfileYamlFileByENV()
		idpServices = IDPServices{}
	})

	AfterEach(func() {
		if idpServices.htpasswd != nil {
			idpServices.htpasswd.Destroy()
			idpServices.htpasswd = nil
		}
		if idpServices.github != nil {
			idpServices.github.Destroy()
			idpServices.github = nil
		}
		if idpServices.gitlab != nil {
			idpServices.gitlab.Destroy()
			idpServices.gitlab = nil
		}
		if idpServices.google != nil {
			idpServices.google.Destroy()
			idpServices.google = nil
		}
		if idpServices.ldap != nil {
			idpServices.ldap.Destroy()
			idpServices.ldap = nil
		}
		if idpServices.multi_idp != nil {
			idpServices.multi_idp.Destroy()
			idpServices.multi_idp = nil
		}
		if idpServices.openid != nil {
			idpServices.openid.Destroy()
			idpServices.openid = nil
		}
	})

	Describe("provision and update", func() {
		Context("Htpasswd", func() {
			BeforeEach(func() {
				var err error
				idpServices.htpasswd, err = exec.NewIDPService(constants.HtpasswdDir) // init new htpasswd service
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeed - [id:63151]", ci.High, ci.Exclude, func() {
				By("Create htpasswd idp for an existing cluster")
				idpParam := getDefaultHTPasswordArgs("OCP-63151-htpasswd-idp-test")
				_, err := idpServices.htpasswd.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())
				idpOutput, err := idpServices.htpasswd.Output()
				Expect(err).ToNot(HaveOccurred())

				By("List existing HtpasswdUsers and compare to the created one")
				htpasswdUsersList, _ := cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
				respUserName, _ := htpasswdUsersList.Items().Slice()[0].GetUsername()
				Expect(respUserName).To(Equal(defaultHTPUsername))

				By("Login with created htpasswd idp")
				// this condition is for cases where the cluster profile
				// has private_link enabled, then regular login won't work
				if !profile.IsPrivateLink() {
					getResp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
					Expect(err).ToNot(HaveOccurred())
					server := getResp.Body().API().URL()

					ocAtter := &openshift.OcAttributes{
						Server:    server,
						Username:  defaultHTPUsername,
						Password:  defaultHTPPassword,
						ClusterID: clusterID,
						AdditioanlFlags: []string{
							"--insecure-skip-tls-verify",
							fmt.Sprintf("--kubeconfig %s", path.Join(constants.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, defaultHTPUsername))),
						},
						Timeout: 7,
					}
					_, err = openshift.OcLogin(*ocAtter)
					Expect(err).ToNot(HaveOccurred())
				} else {
					Logger.Infof("private_link is enabled, skipping login command check.")
				}
			})
			It("Update htpasswd idp - [id:73154]", ci.High, func() {
				htpIDPName := "tf-htpassed-idp"

				By("Create htpasswd idp for an existing cluster")
				idpParam := getDefaultHTPasswordArgs(htpIDPName)
				_, err := idpServices.htpasswd.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())
				idpOutput, err := idpServices.htpasswd.Output()
				Expect(err).ToNot(HaveOccurred())

				By("List existing HtpasswdUsers and compare to the created one")
				htpasswdUsersList, _ := cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
				respUserName, _ := htpasswdUsersList.Items().Slice()[0].GetUsername()
				Expect(respUserName).To(Equal(defaultHTPUsername))

				By("Update htpasswd idp password of 'my-admin-user'")
				newPassword := helper.GenerateRandomStringWithSymbols(15)
				(*idpParam.HtpasswdUsers)[0].Password = helper.StringPointer(newPassword)
				_, err = idpServices.htpasswd.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Check resource state file is updated")
				resource, err := helper.GetResource(constants.HtpasswdDir, "rhcs_identity_provider", "htpasswd_idp")
				Expect(err).ToNot(HaveOccurred())
				Expect(resource).To(MatchJQ(fmt.Sprintf(`.instances[0].attributes.htpasswd.users[] | select(.username == "%s") .password`, defaultHTPUsername), newPassword))

				By("Update htpasswd idp by adding two new users")
				userName2 := "my-admin-user2"
				password2 := helper.GenerateRandomStringWithSymbols(15)
				userName3 := "my-admin-user3"
				password3 := helper.GenerateRandomStringWithSymbols(15)

				htpUsers := (*idpParam.HtpasswdUsers)
				htpUsers = append(htpUsers,
					exec.HTPasswordUser{
						Username: helper.StringPointer(userName2),
						Password: helper.StringPointer(password2),
					},
					exec.HTPasswordUser{
						Username: helper.StringPointer(userName3),
						Password: helper.StringPointer(password3),
					})
				idpParam.HtpasswdUsers = &htpUsers
				_, err = idpServices.htpasswd.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Update htpasswd idp on the second user password")
				newPassword2 := helper.GenerateRandomStringWithSymbols(15)
				(*idpParam.HtpasswdUsers)[1].Password = helper.StringPointer(newPassword2)
				_, err = idpServices.htpasswd.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Check resource state file is updated")
				resource, err = helper.GetResource(constants.HtpasswdDir, "rhcs_identity_provider", "htpasswd_idp")
				Expect(err).ToNot(HaveOccurred())
				Expect(resource).To(MatchJQ(fmt.Sprintf(`.instances[0].attributes.htpasswd.users[] | select(.username == "%s") .password`, userName2), newPassword2))

				By("List existing HtpasswdUsers and compare to the created one")
				htpasswdUsersList, _ = cms.ListHtpasswdUsers(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(htpasswdUsersList.Status()).To(Equal(http.StatusOK))
				Expect(htpasswdUsersList.Items().Len()).To(Equal(3))

				respUserSlice := htpasswdUsersList.Items().Slice()
				userNameToCheck := map[string]bool{
					defaultHTPUsername: true,
					userName2:          true,
					userName3:          true,
				}
				for _, user := range respUserSlice {
					_, existed := userNameToCheck[user.Username()]
					Expect(existed).To(BeTrue())
				}
			})
		})

		Context("LDAP", func() {
			BeforeEach(func() {
				defaultLDAPUsername = "newton"
				defaultLDAPPassword = "password"

				var err error
				idpServices.ldap, err = exec.NewIDPService(constants.LdapDir) // init new ldap service
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeed - [id:63332]", ci.High,
				ci.Exclude,
				func() {
					By("Create LDAP idp for an existing cluster")

					idpParam := &exec.IDPArgs{
						ClusterID:      helper.StringPointer(clusterID),
						Name:           helper.StringPointer("OCP-63332-ldap-idp-test"),
						CA:             helper.EmptyStringPointer,
						URL:            helper.StringPointer(constants.LdapURL),
						LDAPAttributes: &exec.LDAPAttributes{},
						Insecure:       helper.BoolPointer(true),
					}
					_, err := idpServices.ldap.Apply(idpParam)
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
							Username:  defaultLDAPUsername,
							Password:  defaultLDAPPassword,
							ClusterID: clusterID,
							AdditioanlFlags: []string{
								"--insecure-skip-tls-verify",
								fmt.Sprintf("--kubeconfig %s", path.Join(constants.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, defaultLDAPUsername))),
							},
							Timeout: 7,
						}
						_, err = openshift.OcLogin(*ocAtter)
						Expect(err).ToNot(HaveOccurred())
					} else {
						Logger.Infof("private_link is enabled, skipping login command check.")
					}
				})
		})
		Context("GitLab", func() {
			BeforeEach(func() {
				var err error
				idpServices.gitlab, err = exec.NewIDPService(constants.GitlabDir) // init new gitlab service
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeed - [id:64028]", ci.High, func() {
				By("Create GitLab idp for an existing cluster")

				idpParam := &exec.IDPArgs{
					ClusterID:    helper.StringPointer(clusterID),
					Name:         helper.StringPointer("OCP-64028-gitlab-idp-test"),
					ClientID:     helper.StringPointer(defaultGitlabIDPClientId),
					ClientSecret: helper.StringPointer(defaultGitlabIDPClientSecret),
					URL:          helper.StringPointer(constants.GitLabURL),
				}
				_, err := idpServices.gitlab.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Check gitlab idp created for the cluster")
				idpOutput, err := idpServices.gitlab.Output()
				Expect(err).ToNot(HaveOccurred())

				resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Status()).To(Equal(http.StatusOK))
			})
		})
		Context("GitHub", func() {
			BeforeEach(func() {
				var err error
				idpServices.github, err = exec.NewIDPService(constants.GithubDir) // init new github service
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeed - [id:64027]", ci.High, func() {
				By("Create GitHub idp for an existing cluster")

				idpParam := &exec.IDPArgs{
					ClusterID:     helper.StringPointer(clusterID),
					Name:          helper.StringPointer("OCP-64027-github-idp-test"),
					ClientID:      helper.StringPointer(defaultGithubIDPClientId),
					ClientSecret:  helper.StringPointer(defaultGithubIDPClientSecret),
					Organizations: helper.StringSlicePointer(constants.Organizations),
				}
				_, err := idpServices.github.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Check github idp created for the cluster")
				idpOutput, err := idpServices.github.Output()
				Expect(err).ToNot(HaveOccurred())

				resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Status()).To(Equal(http.StatusOK))
			})
		})
		Context("Google", func() {
			BeforeEach(func() {
				var err error
				idpServices.google, err = exec.NewIDPService(constants.GoogleDir) // init new google service
				Expect(err).ToNot(HaveOccurred())
			})

			It("will succeeed - [id:64029]", ci.High, func() {
				By("Create Google idp for an existing cluster")

				idpParam := &exec.IDPArgs{
					ClusterID:    helper.StringPointer(clusterID),
					Name:         helper.StringPointer("OCP-64029-google-idp-test"),
					ClientID:     helper.StringPointer(defaultGoogleIDPClientId),
					ClientSecret: helper.StringPointer(defaultGoogleIDPClientSecret),
					HostedDomain: helper.StringPointer(constants.HostedDomain),
				}
				_, err := idpServices.google.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Check google idp created for the cluster")
				idpOutput, err := idpServices.google.Output()
				Expect(err).ToNot(HaveOccurred())

				resp, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Status()).To(Equal(http.StatusOK))
			})
		})
		Context("Multi IDPs", func() {
			BeforeEach(func() {
				if profile.IsPrivateLink() {
					Skip("private_link is enabled, skipping test.")
				}

				defaultLDAPUsername = "newton"
				defaultLDAPPassword = "password"
			})

			It("will succeed - [id:64030]", ci.Medium, ci.Exclude, func() {
				var err error
				idpServices.multi_idp, err = exec.NewIDPService(constants.MultiIDPDir) // init multi-idp service
				Expect(err).ToNot(HaveOccurred())

				By("Applying google & ldap idps users using terraform")
				idpParam := &exec.IDPArgs{
					ClusterID:      helper.StringPointer(clusterID),
					Name:           helper.StringPointer("OCP-64030"),
					ClientID:       helper.StringPointer(defaultGoogleIDPClientId),
					ClientSecret:   helper.StringPointer(defaultGoogleIDPClientSecret),
					HostedDomain:   helper.StringPointer(constants.HostedDomain),
					CA:             helper.EmptyStringPointer,
					URL:            helper.StringPointer(constants.LdapURL),
					LDAPAttributes: &exec.LDAPAttributes{},
					Insecure:       helper.BoolPointer(true),
				}

				_, err = idpServices.multi_idp.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Login to the ldap user created with terraform")
				By("& cluster-admin user created on cluster deployment")

				resp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				server := resp.Body().API().URL()

				ocAtter := &openshift.OcAttributes{
					Server:    server,
					Username:  defaultLDAPUsername,
					Password:  defaultLDAPPassword,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(constants.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, defaultLDAPUsername))),
					},
					Timeout: 7,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())

				if !profile.AdminEnabled {
					Skip("The test configured only for cluster admin profile")
				}

				// login to the cluster using cluster-admin creds
				username := constants.ClusterAdminUser
				password := helper.GetClusterAdminPassword()
				Expect(password).ToNot(BeEmpty())

				ocAtter = &openshift.OcAttributes{
					Server:    server,
					Username:  username,
					Password:  password,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(constants.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, username))),
					},
					Timeout: 10,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())
			})
			It("with multiple users will reconcile a multiuser config - [id:66408]", ci.Medium, ci.Exclude, func() {
				var err error
				idpServices.htpasswd, err = exec.NewIDPService(constants.HtpasswdDir) // init new htpasswd service
				Expect(err).ToNot(HaveOccurred())

				By("Create 3 htpasswd users for existing cluster")
				idpParam := getDefaultHTPasswordArgs("OCP-66408-htpasswd-multi-test")

				htpUsers := (*idpParam.HtpasswdUsers)
				htpUsers = append(htpUsers,
					exec.HTPasswordUser{
						Username: helper.StringPointer("second_user"),
						Password: helper.StringPointer(helper.GenerateRandomStringWithSymbols(15)),
					},
					exec.HTPasswordUser{
						Username: helper.StringPointer("third_user"),
						Password: helper.StringPointer(helper.GenerateRandomStringWithSymbols(15)),
					},
				)
				_, err = idpServices.htpasswd.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Login to the cluster with one of the users created")
				resp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				server := resp.Body().API().URL()

				ocAtter := &openshift.OcAttributes{
					Server:    server,
					Username:  defaultHTPUsername,
					Password:  defaultHTPPassword,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(constants.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, defaultHTPUsername))),
					},
					Timeout: 10,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())
				idpOutput, err := idpServices.htpasswd.Output()
				Expect(err).ToNot(HaveOccurred())

				By("Delete one of the users using backend api")
				_, err = cms.DeleteIDP(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(err).ToNot(HaveOccurred())

				// wait few minutes before trying to create the resource again
				time.Sleep(time.Minute * 5)

				By("Re-run terraform apply on the same resources")
				_, err = idpServices.htpasswd.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Re-login terraform apply on the same resources")

				// note - this step failes randmonly.
				// hence, the test is currently skipped for ci
				ocAtter = &openshift.OcAttributes{
					Server:    server,
					Username:  defaultHTPUsername,
					Password:  defaultHTPPassword,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(constants.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, defaultHTPUsername))),
					},
					Timeout: 10,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("validate", func() {
		validateIDPArgAgainstErrorSubstrings := func(svc exec.IDPService, idpArgs *exec.IDPArgs, errSubStrings ...string) {
			_, err := svc.Apply(idpArgs)
			Expect(err).To(HaveOccurred())
			for _, errStr := range errSubStrings {
				helper.ExpectTFErrorContains(err, errStr)
			}
		}

		It("the mandatory idp's attributes must be set - [id:68939]", ci.Medium, func() {
			var err error
			idpServices.htpasswd, err = exec.NewIDPService(constants.HtpasswdDir) // init new htpasswd service
			Expect(err).ToNot(HaveOccurred())
			idpServices.ldap, err = exec.NewIDPService(constants.LdapDir) // init new ldap service
			Expect(err).ToNot(HaveOccurred())
			idpServices.github, err = exec.NewIDPService(constants.GithubDir) // init new github service
			Expect(err).ToNot(HaveOccurred())
			idpServices.gitlab, err = exec.NewIDPService(constants.GitlabDir) // init new gitlab service
			Expect(err).ToNot(HaveOccurred())
			idpServices.google, err = exec.NewIDPService(constants.GoogleDir) // init new google service
			Expect(err).ToNot(HaveOccurred())
			idpName := "ocp-68939-htp"

			By("Create htpasswd idp with empty name field")
			args := getDefaultHTPasswordArgs(idpName)
			args.Name = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.htpasswd, args, "Attribute 'name' is mandatory")

			By("Create htpasswd idp with empty username field")
			args = getDefaultHTPasswordArgs(idpName)
			(*args.HtpasswdUsers)[0].Username = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.htpasswd, args, "htpasswd.users[0].username username may not be empty/blank string")

			By("Create htpasswd idp with empty password field")
			args = getDefaultHTPasswordArgs(idpName)
			(*args.HtpasswdUsers)[0].Password = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.htpasswd, args, "htpasswd.users[0].password password must contain numbers or symbols")

			By("Create ldap idp with empty name field")
			args = getDefaultLDAPArgs(idpName)
			args.Name = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.ldap, args, "Attribute 'name' is mandatory")

			By("Create ldap idp with empty url field")
			args = getDefaultLDAPArgs(idpName)
			args.URL = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.ldap, args, "'url' field is mandatory")

			By("Create ldap idp without attributes field")
			args = getDefaultLDAPArgs(idpName)
			args.LDAPAttributes = nil
			validateIDPArgAgainstErrorSubstrings(idpServices.ldap, args, "provider has marked it as required")

			By("Create github idp with empty name field")
			args = getDefaultGitHubArgs(idpName)
			args.Name = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.github, args, "Attribute 'name' is mandatory")

			By("Create github idp without client_id field")
			args = getDefaultGitHubArgs(idpName)
			args.ClientID = nil
			validateIDPArgAgainstErrorSubstrings(idpServices.github, args, "No value for required variable")

			By("Create github idp with empty client_id field")
			args = getDefaultGitHubArgs(idpName)
			args.ClientID = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.github, args, "Attribute 'client_id' is mandatory")

			By("Create github idp without client_secret field")
			args = getDefaultGitHubArgs(idpName)
			args.ClientSecret = nil
			validateIDPArgAgainstErrorSubstrings(idpServices.github, args, "No value for required variable")

			By("Create github idp with empty client_secret field")
			args = getDefaultGitHubArgs(idpName)
			args.ClientSecret = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.github, args, "Attribute 'client_secret' is mandatory")

			By("Create gitlab idp with empty name field")
			args = getDefaultGitlabArgs(idpName)
			args.Name = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.gitlab, args, "Attribute 'name' is mandatory")

			By("Create gitlab idp without client_id field")
			args = getDefaultGitlabArgs(idpName)
			args.ClientID = nil
			validateIDPArgAgainstErrorSubstrings(idpServices.gitlab, args, "provider has marked it as required")

			By("Create gitlab idp with empty client_id field")
			args = getDefaultGitlabArgs(idpName)
			args.ClientID = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.gitlab, args, "Attribute 'client_id' is mandatory")

			By("Create gitlab idp with empty client_secret field")
			args = getDefaultGitlabArgs(idpName)
			args.ClientSecret = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.gitlab, args, "Attribute 'client_secret' is mandatory")

			By("Create google idp with empty name field")
			args = getDefaultGoogleArgs(idpName)
			args.Name = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.google, args, "Attribute 'name' is mandatory")

			By("Create google idp without client_id field")
			args = getDefaultGoogleArgs(idpName)
			args.ClientID = nil
			validateIDPArgAgainstErrorSubstrings(idpServices.google, args, "provider has marked it as required")

			By("Create google idp with empty client_id field")
			args = getDefaultGoogleArgs(idpName)
			args.ClientID = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.google, args, "Attribute 'client_id' is mandatory")

			By("Create google idp without client_secret field")
			args = getDefaultGoogleArgs(idpName)
			args.ClientSecret = nil
			validateIDPArgAgainstErrorSubstrings(idpServices.google, args, "provider has marked it as required")

			By("Create google idp with empty client_secret field")
			args = getDefaultGoogleArgs(idpName)
			args.ClientSecret = helper.EmptyStringPointer
			validateIDPArgAgainstErrorSubstrings(idpServices.google, args, "Attribute 'client_secret' is mandatory")
		})

		It("htpasswd with empty user-password list will fail - [id:66409]", ci.Medium, func() {
			var err error
			idpServices.htpasswd, err = exec.NewIDPService(constants.HtpasswdDir) // init new htpasswd service
			Expect(err).ToNot(HaveOccurred())
			idpName := "ocp-66409"

			By("Validate idp can't be created with empty htpasswdMap")
			args := getDefaultHTPasswordArgs(idpName)
			args.HtpasswdUsers = &[]exec.HTPasswordUser{}
			validateIDPArgAgainstErrorSubstrings(idpServices.htpasswd, args, "Attribute htpasswd.users list must contain at least 1 elements")
		})

		It("htpasswd password policy - [id:66410]", ci.Medium, func() {
			var err error
			idpServices.htpasswd, err = exec.NewIDPService(constants.HtpasswdDir) // init new htpasswd service
			Expect(err).ToNot(HaveOccurred())
			idpName := "ocp-66410"

			var usernameInvalid = "userWithInvalidPassword"

			By("Validate idp can't be created with password less than 14")
			args := getDefaultHTPasswordArgs(idpName)
			newHTPwd := append(*args.HtpasswdUsers, exec.HTPasswordUser{
				Username: helper.StringPointer(usernameInvalid),
				Password: helper.StringPointer(helper.GenerateRandomStringWithSymbols(3)),
			})
			args.HtpasswdUsers = &newHTPwd
			validateIDPArgAgainstErrorSubstrings(idpServices.htpasswd, args, "password string length must be at least 14")

			By("Validate idp can't be created without upercase letter in password")
			args = getDefaultHTPasswordArgs(idpName)
			newHTPwd = append(*args.HtpasswdUsers, exec.HTPasswordUser{
				Username: helper.StringPointer(usernameInvalid),
				Password: helper.StringPointer(helper.Subfix(3)),
			})
			args.HtpasswdUsers = &newHTPwd
			validateIDPArgAgainstErrorSubstrings(idpServices.htpasswd, args, "password must contain uppercase")
		})

		It("htpasswd with duplicate usernames will fail - [id:66411]", ci.Medium, func() {
			var err error
			idpServices.htpasswd, err = exec.NewIDPService(constants.HtpasswdDir) // init new htpasswd service
			Expect(err).ToNot(HaveOccurred())
			idpName := "ocp-66411"

			By("Create 2 htpasswd idps with the same username")
			args := getDefaultHTPasswordArgs(idpName)
			newHTPwd := append(*args.HtpasswdUsers, exec.HTPasswordUser{
				Username: helper.StringPointer(defaultHTPUsername),
				Password: helper.StringPointer(defaultHTPPassword),
			})
			args.HtpasswdUsers = &newHTPwd
			validateIDPArgAgainstErrorSubstrings(idpServices.htpasswd, args, "Usernames in HTPasswd user list must be unique")
		})

		It("resource can be imported - [id:65981]",
			ci.Medium, ci.FeatureImport, func() {
				var err error

				idpServices.gitlab, err = exec.NewIDPService(constants.GitlabDir) // init new gitlab service
				Expect(err).ToNot(HaveOccurred())
				idpServices.google, err = exec.NewIDPService(constants.GoogleDir) // init new google service
				Expect(err).ToNot(HaveOccurred())
				importService, err := exec.NewImportService(constants.ImportResourceDir)
				Expect(err).ToNot(HaveOccurred())

				By("Create sample idps to test the import functionality")
				idpGoogleName := "ocp-65981-google"
				idpParam := getDefaultGoogleArgs(idpGoogleName)
				_, err = idpServices.google.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				idpGitlabName := "ocp-65981-gitlab"
				idpParam = getDefaultGitlabArgs(idpGitlabName)
				_, err = idpServices.gitlab.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("Run the command to import the idp")
				defer func() {
					importService.Destroy()
				}()
				successImportParams := &exec.ImportArgs{
					ClusterID:  clusterID,
					Resource:   "rhcs_identity_provider.idp_google_import",
					ObjectName: idpGoogleName,
				}
				importParam := successImportParams
				_, err = importService.Import(importParam)
				Expect(err).ToNot(HaveOccurred())

				By("Check resource state - import command succeeded")
				output, err := importService.ShowState(importParam.Resource)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(ContainSubstring(defaultGoogleIDPClientId))
				Expect(output).To(ContainSubstring(constants.HostedDomain))

				By("Validate terraform import with no idp object name returns error")
				var unknownIdpName = "unknown_idp_name"
				importParam = &exec.ImportArgs{
					ClusterID:  clusterID,
					Resource:   "rhcs_identity_provider.idp_google_import",
					ObjectName: unknownIdpName,
				}
				_, err = importService.Import(importParam)
				Expect(err.Error()).To(ContainSubstring("identity provider '%s' not found", unknownIdpName))

				By("Validate terraform import with no clusterID returns error")
				var unknownClusterID = helper.GenerateRandomStringWithSymbols(20)
				importParam = &exec.ImportArgs{
					ClusterID:  unknownClusterID,
					Resource:   "rhcs_identity_provider.idp_gitlab_import",
					ObjectName: idpGitlabName,
				}
				_, err = importService.Import(importParam)
				Expect(err.Error()).To(ContainSubstring("Cluster %s not found", unknownClusterID))

			})
	})

	Describe("reconciliation", func() {
		var (
			gitlabIdpOcmAPI *cmsv1.IdentityProviderBuilder
		)

		BeforeEach(func() {
			gitlabIdpOcmAPI = cmsv1.NewIdentityProvider()

		})

		It("verify basic flow - [id:65808]",
			ci.Medium, ci.Exclude, func() {
				var err error
				idpServices.ldap, err = exec.NewIDPService(constants.LdapDir) // init new ldap service
				Expect(err).ToNot(HaveOccurred())
				idpName := "ocp-65808"

				By("Create ldap idp user")
				idpParam := getDefaultLDAPArgs(idpName)
				_, err = idpServices.ldap.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())
				idpOutput, err := idpServices.ldap.Output()
				Expect(err).ToNot(HaveOccurred())

				By("Login to the ldap user")
				resp, err := cms.RetrieveClusterDetail(ci.RHCSConnection, clusterID)
				Expect(err).ToNot(HaveOccurred())
				server := resp.Body().API().URL()

				ocAtter := &openshift.OcAttributes{
					Server:    server,
					Username:  defaultLDAPUsername,
					Password:  defaultLDAPPassword,
					ClusterID: clusterID,
					AdditioanlFlags: []string{
						"--insecure-skip-tls-verify",
						fmt.Sprintf("--kubeconfig %s", path.Join(constants.RHCS.KubeConfigDir, fmt.Sprintf("%s.%s", clusterID, defaultLDAPUsername))),
					},
					Timeout: 7,
				}
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())

				By("Delete ldap idp by OCM API")
				cms.DeleteIDP(ci.RHCSConnection, clusterID, idpOutput.ID)
				_, err = cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(
					"Identity provider ID '%s' for cluster '%s' not found", idpOutput.ID, clusterID),
				)

				By("Re-apply the idp resource")
				_, err = idpServices.ldap.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())

				By("re-login with the ldpa idp username/password")
				_, err = openshift.OcLogin(*ocAtter)
				Expect(err).ToNot(HaveOccurred())

			})

		It("try to restore/duplicate an existing IDP - [id:65816]",
			ci.Medium, func() {
				var err error
				idpServices.gitlab, err = exec.NewIDPService(constants.GitlabDir) // init new gitlab service
				Expect(err).ToNot(HaveOccurred())
				gitLabIDPName := "OCP-65816-gitlab-idp-reconcil"

				By("Create gitlab idp for existing cluster")
				idpParam := getDefaultGitlabArgs(gitLabIDPName)
				_, err = idpServices.gitlab.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())
				idpOutput, err := idpServices.gitlab.Output()
				Expect(err).ToNot(HaveOccurred())

				By("Delete gitlab idp using OCM API")
				cms.DeleteIDP(ci.RHCSConnection, clusterID, idpOutput.ID)
				_, err = cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpOutput.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(
					"Identity provider ID '%s' for cluster '%s' not found", idpOutput.ID, clusterID),
				)

				By("Create gitlab idp using OCM api with the same parameters")
				requestBody, _ := gitlabIdpOcmAPI.Type("GitlabIdentityProvider").
					Name(gitLabIDPName).
					Gitlab(cmsv1.NewGitlabIdentityProvider().
						ClientID(defaultGitlabIDPClientId).
						ClientSecret(defaultGitlabIDPClientSecret).
						URL(constants.GitLabURL)).
					MappingMethod("claim").
					Build()
				res, err := cms.CreateClusterIDP(ci.RHCSConnection, clusterID, requestBody)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Status()).To(Equal(http.StatusCreated))

				// Delete gitlab idp from existing cluster after test end
				defer cms.DeleteIDP(ci.RHCSConnection, clusterID, res.Body().ID())

				By("Re-apply gitlab idp using tf manifests with same ocm api args")
				_, err = idpServices.gitlab.Apply(idpParam)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(
					"Identity Provider Name\n%s already exists", gitLabIDPName),
				)
			})

		It("try to restore to an updated IDP config - [id:65814]",
			ci.Medium, func() {
				var err error
				idpServices.gitlab, err = exec.NewIDPService(constants.GitlabDir) // init new gitlab service
				Expect(err).ToNot(HaveOccurred())
				gitLabIDPName := "OCP-65814-gitlab-idp-reconcil"

				By("Create gitlab idp for existing cluster")
				idpParam := getDefaultGitlabArgs(gitLabIDPName)
				_, err = idpServices.gitlab.Apply(idpParam)
				Expect(err).ToNot(HaveOccurred())
				idpOutput, err := idpServices.gitlab.Output()
				Expect(err).ToNot(HaveOccurred())

				// Delete gitlab idp using OCM API after test end
				defer func() {
					cms.DeleteIDP(ci.RHCSConnection, clusterID, idpOutput.ID)
					_, err := cms.RetrieveClusterIDPDetail(ci.RHCSConnection, clusterID, idpOutput.ID)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring(
						"Identity provider ID '%s' for cluster '%s' not found", idpOutput.ID, clusterID),
					)
				}()

				By("Edit gitlab idp using ocm api")
				newClientSecret := helper.GenerateRandomStringWithSymbols(30)
				requestBody, _ := gitlabIdpOcmAPI.Type("GitlabIdentityProvider").
					Gitlab(cmsv1.NewGitlabIdentityProvider().
						ClientID(defaultGitlabIDPClientSecret).
						ClientSecret(newClientSecret).
						URL(constants.GitLabURL)).
					MappingMethod("claim").
					Build()

				resp, err := cms.PatchIDP(ci.RHCSConnection, clusterID, idpOutput.ID, requestBody)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.Status()).To(Equal(http.StatusOK))

				// update is currently not supported for idp :: OCM-4622
				By("Update and apply idp using terraform")
				newClientSecret = helper.GenerateRandomStringWithSymbols(30)
				idpParam.ClientSecret = helper.StringPointer(newClientSecret)
				_, err = idpServices.gitlab.Apply(idpParam)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring(
					"This RHCS provider version does not support updating an existing IDP"),
				)
			})
	})
})
