/*
Copyright (c) 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package classic

import (
	"net/http"

	"github.com/terraform-redhat/terraform-provider-rhcs/build"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/terraform-redhat/terraform-provider-rhcs/subsystem/framework"
)

const htpasswdValidPass = "123PasS8901234"
const htpasswdValidPass2 = "123PasS89012342"
const hashedPass = "hash(123PasS8901234)"
const htpasswdInValidPass = "my-pass"

var _ = Describe("Identity provider creation", func() {

	const users1 = `"items": [{
					"username": "my-user",
					"password": "` + htpasswdValidPass + `"
				}]`

	const users2 = `"items": [{
					"username": "my-user",
					"password": "` + htpasswdValidPass2 + `"
				}]`

	const users3 = `
				"items": [{
						"username": "my-user",
						"password": "` + htpasswdValidPass2 + `"
					},
					{
						"username": "my-user2",
						"password": "` + htpasswdValidPass2 + `"
				}]`

	const templatePt1 = `
		{
			"kind": "IdentityProvider",
			"id": "456",
			"mapping_method": "claim",
			"htpasswd": {
				"users": {`

	const templatePt2 = `
				}
			},
			"name": "my-ip"
		}`

	const template = templatePt1 + users1 + templatePt2
	const template2 = templatePt1 + users2 + templatePt2
	const template3 = templatePt1 + users3 + templatePt2

	Context("Identity Provider Failure", func() {
		It("fails if name is empty", func() {
			Terraform.Source(`
			resource "rhcs_identity_provider" "my_idp" {
					cluster = ""
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`The argument "name" is required, but no definition was found`)
		})
		It("fails if cluster is empty", func() {
			Terraform.Source(`
			resource "rhcs_identity_provider" "my_idp" {
					name = "my-idp"
					cluster = ""
				}
			`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`Attribute cluster cluster ID may not be empty/blank string`)
		})
		It("cluster_id not found", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusNotFound, `{
			    	  "id": "123",
			    	  "name": "my-cluster",
			    	  "state": "ready"
			    	}`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_idp" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
                  users = [{
                    username = "my-user"
	    	        password = "` + htpasswdValidPass + `"
                  }]
	    	    }
	    	  }
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Cluster 123 not found")
		})
		Context("Cluster exists, but invalid config", func() {
			BeforeEach(func() {
				// The first thing that the provider will do for any operation on identity providers
				// is check that the cluster is ready, so we always need to prepare the server to
				// respond to that:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusOK, `{
			        	  "id": "123",
			        	  "name": "my-cluster",
			        	  "state": "ready"
			        	}`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
						RespondWithJSON(http.StatusOK, `{
			        	  "id": "123",
			        	  "name": "my-cluster",
			        	  "state": "ready"
			        	}`),
					),
				)
			})

			It("Can't create a 'htpasswd' identity provider. No users provided", func() {
				// Run the apply command:
				Terraform.Source(`
	    	      resource "rhcs_identity_provider" "my_idp" {
	    	        cluster = "123"
	    	        name    = "my-ip"
	    	        htpasswd = {
                      users = []
	    	        }
	    	      }
	    	    `)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("Attribute htpasswd.users list must contain at least 1 elements")
			})
			It("Can't create a 'htpasswd' identity provider. duplication of username", func() {
				// Run the apply command:
				Terraform.Source(`
	    	      resource "rhcs_identity_provider" "my_idp" {
	    	        cluster = "123"
	    	        name    = "my-ip"
	    	        htpasswd = {
                      users = [
                        {
                            username = "foo"
                            password = "` + htpasswdValidPass + `"
                        },
                        {
                            username = "foo"
                            password = "` + htpasswdValidPass + `"
                        }
                      ]
	    	        }
	    	      }
	    	    `)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("Usernames in HTPasswd user list must be unique")
			})
			It("Can't create a 'htpasswd' identity provider. invalid username", func() {
				// Run the apply command:
				Terraform.Source(`
	    	      resource "rhcs_identity_provider" "my_idp" {
	    	        cluster = "123"
	    	        name    = "my-ip"
	    	        htpasswd = {
                      users = [{
	    	            username = "my%user"
	    	            password = "` + htpasswdValidPass + `"
                      }]
	    	        }
	    	      }
	    	    `)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("Attribute htpasswd.users[0].username username may not contain the characters")
			})
			It("Can't create a 'htpasswd' identity provider. invalid password", func() {
				// Run the apply command:
				Terraform.Source(`
	    	      resource "rhcs_identity_provider" "my_idp" {
	    	        cluster = "123"
	    	        name    = "my-ip"
	    	        htpasswd = {
                      users = [{
	    	            username = "my-user"
	    	            password = "` + htpasswdInValidPass + `"
                      }]
	    	        }
	    	      }
	    	    `)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).ToNot(BeZero())
				runOutput.VerifyErrorContainsSubstring("Attribute htpasswd.users[0].password password must contain uppercase characters")
			})
		})
	})

	Context("Identity Provider Success", func() {
		BeforeEach(func() {
			// The first thing that the provider will do for any operation on identity providers
			// is check that the cluster is ready, so we always need to prepare the server to
			// respond to that:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "123",
			    	  "name": "my-cluster",
			    	  "state": "ready"
			    	}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "123",
			    	  "name": "my-cluster",
			    	  "state": "ready"
			    	}`),
				),
			)
		})

		It("Can create a 'htpasswd' identity provider", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					),
					VerifyJSON(`{
			    	  "kind": "IdentityProvider",
			    	  "type": "HTPasswdIdentityProvider",
                      "mapping_method": "claim",
			    	  "name": "my-ip",
			    	  "htpasswd": {
                        "users": {"items":[{"username": "my-user", "hashed_password": "`+hashedPass+`"}]}
			    	  }
			    	}`),
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "456",
			    	  "name": "my-ip",
                      "mapping_method": "claim",
			    	  "htpasswd": {
                        "users": {"items":[{"username": "my-user", "password": "`+htpasswdValidPass+`"}]}
			    	  }
			    	}`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_idp" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
                  users = [{
	    	        username = "my-user"
	    	        password = "` + htpasswdValidPass + `"
                  }]
	    	    }
	    	  }
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		It("Reconcile an 'htpasswd' identity provider, when state exists but 404 from server", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					),
					VerifyJSON(`{
			    	  "kind": "IdentityProvider",
			    	  "type": "HTPasswdIdentityProvider",
                      "mapping_method": "claim",
			    	  "name": "my-ip",
			    	  "htpasswd": {
                        "users": {"items":[{"username": "my-user", "hashed_password": "`+hashedPass+`"}]}
			    	  }
			    	}`),
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "456",
			    	  "name": "my-ip",
                      "mapping_method": "claim",
			    	  "htpasswd": {
                        "users": {"items":[{"username": "my-user", "password": "`+htpasswdValidPass+`"}]}
			    	  }
			    	}`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_idp" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
                  users = [{
	    	        username = "my-user"
	    	        password = "` + htpasswdValidPass + `"
                  }]
	    	    }
	    	  }
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())

			// Prepare the server for upgrade
			TestServer.AppendHandlers(
				// read from server (404)
				CombineHandlers(
					VerifyRequest(
						http.MethodGet,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers/456",
					),
					RespondWithJSON(http.StatusNotFound, "{}"),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "123",
			    	  "name": "my-cluster",
			    	  "state": "ready"
			    	}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "123",
			    	  "name": "my-cluster",
			    	  "state": "ready"
			    	}`),
				),
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					),
					VerifyJSON(`{
			    	  "kind": "IdentityProvider",
			    	  "type": "HTPasswdIdentityProvider",
                      "mapping_method": "claim",
			    	  "name": "my-ip",
			    	  "htpasswd": {
                        "users": {"items":[{"username": "my-user", "hashed_password": "`+hashedPass+`"}]}
			    	  }
			    	}`),
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "457",
			    	  "name": "my-ip",
                      "mapping_method": "claim",
			    	  "htpasswd": {
                        "users": {"items":[{"username": "my-user", "password": "`+htpasswdValidPass+`"}]}
			    	  }
			    	}`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_idp" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
                  users = [{
	    	        username = "my-user"
	    	        password = "` + htpasswdValidPass + `"
                  }]
	    	    }
	    	  }
	    	`)
			runOutput = Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_identity_provider", "my_idp")
			Expect(resource).To(MatchJQ(".attributes.id", "457"))
		})

		It("Can create a 'gitlab' identity provider", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					),
					VerifyJSON(`{
	    			  "kind": "IdentityProvider",
	    			  "type": "GitlabIdentityProvider",
                      "mapping_method": "claim",
	    			  "name": "my-ip",
	    			  "gitlab": {
	    			    "ca": "test-ca",
	    			    "url": "https://test.gitlab.com",
	    			    "client_id": "test-client",
	    			    "client_secret": "test-secret"
	    			  }
	    			}`),
					RespondWithJSON(http.StatusOK, `{
	    			  "id": "456",
	    			  "name": "my-ip",
                      "mapping_method": "claim",
	    			  "gitlab": {
	    			    "ca": "test-ca",
	    			    "url": "https://test.gitlab.com",
	    			    "client_id": "test-client",
	    			    "client_secret": "test-secret"
	    			  }
	    			}`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_idp" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    gitlab = {
	    	      ca = "test-ca"
	    	      url = "https://test.gitlab.com"
	    		  client_id = "test-client"
	    		  client_secret = "test-secret"
	    	    }
	    	  }
	    	`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
		})

		Context("Can create a 'github' identity provider", func() {
			Context("Invalid 'github' identity provider config", func() {
				It("Should fail with both 'teams' and 'organizations'", func() {
					Terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_idp" {
	    	            cluster = "123"
	    	            name    = "my-ip"
	    	            github = {
	    	              ca = "test-ca"
	    	        	  client_id = "test-client"
	    	        	  client_secret = "test-secret"
                          organizations = ["my-org"]
                          teams = ["valid/team"]
	    	            }
	    	          }
	    	        `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).ToNot(BeZero())
					runOutput.VerifyErrorContainsSubstring("2 attributes specified when one (and only one) of [github.teams.<.teams,github.teams.<.organizations] is required")
				})

				It("Should fail without 'teams' or 'organizations'", func() {
					Terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_idp" {
	    	            cluster = "123"
	    	            name    = "my-ip"
	    	            github = {
	    	              ca = "test-ca"
	    	        	  client_id = "test-client"
	    	        	  client_secret = "test-secret"
	    	            }
	    	          }
	    	        `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).ToNot(BeZero())
					runOutput.VerifyErrorContainsSubstring("No attribute specified when one (and only one) of [github.teams.<.teams,github.teams.<.organizations] is required")
				})

				It("Should fail if teams contain an invalid format", func() {
					Terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_idp" {
	    	            cluster = "123"
	    	            name    = "my-ip"
	    	            github = {
	    	              ca = "test-ca"
	    	        	  client_id = "test-client"
	    	        	  client_secret = "test-secret"
                          teams = ["invalidteam"]
	    	            }
	    	          }
	    	        `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).ToNot(BeZero())
					runOutput.VerifyErrorContainsSubstring("Expected a GitHub team to follow the form '<org>/<team>', Got invalidteam")
					Terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_idp" {
	    	            cluster = "123"
	    	            name    = "my-ip"
	    	            github = {
	    	              ca = "test-ca"
	    	        	  client_id = "test-client"
	    	        	  client_secret = "test-secret"
                          teams = ["valid/team", "invalidteam"]
	    	            }
	    	          }
	    	        `)
					runOutput = Terraform.Apply()
					Expect(runOutput.ExitCode).ToNot(BeZero())
					runOutput.VerifyErrorContainsSubstring("Expected a GitHub team to follow the form '<org>/<team>', Got invalidteam")
				})

				It("Should fail with an invalid hostname", func() {
					Terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_idp" {
	    	            cluster = "123"
	    	            name    = "my-ip"
	    	            github = {
	    	              ca = "test-ca"
	    	        	  client_id = "test-client"
	    	        	  client_secret = "test-secret"
                          organizations = ["org"]
                          hostname = "invalidhostname"
	    	            }
	    	          }
	    	        `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).ToNot(BeZero())
					runOutput.VerifyErrorContainsSubstring("Expected a valid GitHub hostname. Got invalidhostname")
				})
			})
			It("Happy flow with org restriction", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(
							http.MethodPost,
							"/api/clusters_mgmt/v1/clusters/123/identity_providers",
						),
						VerifyJSON(`{
    				      "kind": "IdentityProvider",
    				      "type": "GithubIdentityProvider",
                          "mapping_method": "claim",
    				      "name": "my-ip",
    				      "github": {
    				        "ca": "test-ca",
    				        "client_id": "test-client",
    				        "client_secret": "test-secret",
                            "organizations": ["my-org"]
    				      }
    				    }`),
						RespondWithJSON(http.StatusOK, `{
    				      "id": "456",
    				      "name": "my-ip",
                          "mapping_method": "claim",
    				      "github": {
    				        "ca": "test-ca",
    				        "url": "https://test.gitlab.com",
    				        "client_id": "test-client",
    				        "client_secret": "test-secret",
                            "organizations": ["my-org"]
    				      }
    				    }`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
    		      resource "rhcs_identity_provider" "my_idp" {
    		        cluster = "123"
    		        name    = "my-ip"
    		        github = {
    		          ca = "test-ca"
    		    	  client_id = "test-client"
    		    	  client_secret = "test-secret"
                      organizations = ["my-org"]
    		        }
    		      }
    		    `)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
			})

			It("Happy flow with team restriction", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(
							http.MethodPost,
							"/api/clusters_mgmt/v1/clusters/123/identity_providers",
						),
						VerifyJSON(`{
    				      "kind": "IdentityProvider",
    				      "type": "GithubIdentityProvider",
                          "mapping_method": "claim",
    				      "name": "my-ip",
    				      "github": {
    				        "ca": "test-ca",
    				        "client_id": "test-client",
    				        "client_secret": "test-secret",
                            "teams": ["valid/team"]
    				      }
    				    }`),
						RespondWithJSON(http.StatusOK, `{
    				      "id": "456",
    				      "name": "my-ip",
                          "mapping_method": "claim",
    				      "github": {
    				        "ca": "test-ca",
    				        "url": "https://test.gitlab.com",
    				        "client_id": "test-client",
    				        "client_secret": "test-secret",
                            "teams": ["valid/team"]
    				      }
    				    }`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
		          resource "rhcs_identity_provider" "my_idp" {
		            cluster = "123"
		            name    = "my-ip"
		            github = {
		              ca = "test-ca"
		        	  client_id = "test-client"
		        	  client_secret = "test-secret"
                      teams = ["valid/team"]
		            }
		          }
		        `)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
			})
		})

		Context("Can create 'LDAP' Identity provider", func() {
			Context("Invalid LDAP config", func() {
				It("Should fail if not both bind properties are set", func() {
					// Run the apply command:
					Terraform.Source(`
        		      resource "rhcs_identity_provider" "my_idp" {
        		        cluster    = "123"
        		        name       = "my-ip"
        		        ldap = {
        		          bind_dn       = "my-bind-dn"
        		          insecure      = false
        		          ca            = "my-ca"
        		          url           = "ldap://my-server.com"
        		          attributes    = {
        		            id                 = ["dn"]
        		            email              = ["mail"]
        		            name               = ["cn"]
        		            preferred_username = ["uid"]
        		          }
        		        }
        		      }
        		    `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).ToNot(BeZero())
					runOutput.VerifyErrorContainsSubstring(`Attribute "ldap.bind_password" must be specified`)
				})

			})
			It("Happy flow with default attributes", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(
							http.MethodPost,
							"/api/clusters_mgmt/v1/clusters/123/identity_providers",
						),
						VerifyJSON(`{
				          "kind": "IdentityProvider",
				          "type": "LDAPIdentityProvider",
                          "mapping_method": "claim",
				          "name": "my-ip",
				          "ldap": {
				            "ca": "my-ca",
				            "insecure": false,
				            "url": "ldap://my-server.com",
				            "attributes": {
				              "id": ["dn"],
				              "email": ["mail"],
				              "name": ["cn"],
				              "preferred_username": ["uid"]
				            }
				          }
				        }`),
						RespondWithJSON(http.StatusOK, `{
				          "id": "456",
				          "name": "my-ip",
                          "mapping_method": "claim",
				          "ldap": {
				            "ca": "my-ca",
				            "insecure": false,
				            "url": "ldap://my-server.com",
				            "attributes": {
				              "id": ["dn"],
				              "email": ["mail"],
				              "name": ["cn"],
				              "preferred_username": ["uid"]
				            }
				          }
				        }`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
        		  resource "rhcs_identity_provider" "my_idp" {
        		    cluster    = "123"
        		    name       = "my-ip"
        		    ldap = {
        		      insecure      = false
        		      ca            = "my-ca"
        		      url           = "ldap://my-server.com"
									attributes    = {}
        		    }
        		  }
        		`)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
			})
			It("Happy flow with bind values", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(
							http.MethodPost,
							"/api/clusters_mgmt/v1/clusters/123/identity_providers",
						),
						VerifyJSON(`{
				          "kind": "IdentityProvider",
				          "type": "LDAPIdentityProvider",
                          "mapping_method": "claim",
				          "name": "my-ip",
				          "ldap": {
				            "bind_dn": "my-bind-dn",
				            "bind_password": "my-bind-password",
				            "ca": "my-ca",
				            "insecure": false,
				            "url": "ldap://my-server.com",
				            "attributes": {
				              "id": ["dn"],
				              "email": ["mail"],
				              "name": ["cn"],
				              "preferred_username": ["uid"]
				            }
				          }
				        }`),
						RespondWithJSON(http.StatusOK, `{
				          "id": "456",
				          "name": "my-ip",
                          "mapping_method": "claim",
				          "ldap": {
				            "bind_dn": "my-bind-dn",
				            "bind_password": "my-bind-password",
				            "ca": "my-ca",
				            "insecure": false,
				            "url": "ldap://my-server.com",
				            "attributes": {
				              "id": ["dn"],
				              "email": ["mail"],
				              "name": ["cn"],
				              "preferred_username": ["uid"]
				            }
				          }
				        }`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
        		  resource "rhcs_identity_provider" "my_idp" {
        		    cluster    = "123"
        		    name       = "my-ip"
        		    ldap = {
        		      bind_dn       = "my-bind-dn"
        		      bind_password = "my-bind-password"
        		      insecure      = false
        		      ca            = "my-ca"
        		      url           = "ldap://my-server.com"
        		      attributes    = {
										id                 = ["dn"]
										email              = ["mail"]
										name               = ["cn"]
										preferred_username = ["uid"]
        		      }
        		    }
        		  }
        		`)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
			})

			It("Happy flow without bind values", func() {
				// Prepare the server:
				TestServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(
							http.MethodPost,
							"/api/clusters_mgmt/v1/clusters/123/identity_providers",
						),
						VerifyJSON(`{
				          "kind": "IdentityProvider",
				          "type": "LDAPIdentityProvider",
                          "mapping_method": "claim",
				          "name": "my-ip",
				          "ldap": {
				            "ca": "my-ca",
				            "insecure": false,
				            "url": "ldap://my-server.com",
				            "attributes": {
				              "id": ["dn"],
				              "email": ["mail"],
				              "name": ["cn"],
				              "preferred_username": ["uid"]
				            }
				          }
				        }`),
						RespondWithJSON(http.StatusOK, `{
				          "id": "456",
				          "name": "my-ip",
                          "mapping_method": "claim",
				          "ldap": {
				            "ca": "my-ca",
				            "insecure": false,
				            "url": "ldap://my-server.com",
				            "attributes": {
				              "id": ["dn"],
				              "email": ["mail"],
				              "name": ["cn"],
				              "preferred_username": ["uid"]
				            }
				          }
				        }`),
					),
				)

				// Run the apply command:
				Terraform.Source(`
        		  resource "rhcs_identity_provider" "my_idp" {
        		    cluster    = "123"
        		    name       = "my-ip"
        		    ldap = {
        		      insecure      = false
        		      ca            = "my-ca"
        		      url           = "ldap://my-server.com"
        		      attributes    = {
										id                 = ["dn"]
										email              = ["mail"]
										name               = ["cn"]
										preferred_username = ["uid"]
        		      }
        		    }
        		  }
        		`)
				runOutput := Terraform.Apply()
				Expect(runOutput.ExitCode).To(BeZero())
			})
		})

		Context("Google identity provider", func() {
			Context("Invalid google config", func() {
				It("Should fail with invalid hosted_domain", func() {
					// Run the apply command:
					Terraform.Source(`
    		          resource "rhcs_identity_provider" "my_idp" {
    		            cluster = "123"
    		            name    = "my-ip"
    		            google = {
    		        	  client_id = "test-client"
    		        	  client_secret = "test-secret"
                          hosted_domain = "examplecom"
    		            }
    		          }
    		        `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).ToNot(BeZero())
					runOutput.VerifyErrorContainsSubstring("Expected a valid Google hosted_domain. Got examplecom")
				})

				It("Should fail when mapping_method is not lookup and no hosted_domain", func() {
					// Run the apply command:
					Terraform.Source(`
    		          resource "rhcs_identity_provider" "my_idp" {
    		            cluster = "123"
    		            name    = "my-ip"
    		            google = {
    		        	  client_id = "test-client"
    		        	  client_secret = "test-secret"
    		            }
    		          }
    		        `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).ToNot(BeZero())
					runOutput.VerifyErrorContainsSubstring("Expected a valid hosted_domain since mapping_method is set to claim")
				})
			})

			Context("Happy flow", func() {
				It("Should create provider", func() {
					// Prepare the server:
					TestServer.AppendHandlers(
						CombineHandlers(
							VerifyRequest(
								http.MethodPost,
								"/api/clusters_mgmt/v1/clusters/123/identity_providers",
							),
							VerifyJSON(`{
    			    	      "kind": "IdentityProvider",
    			    	      "type": "GoogleIdentityProvider",
                              "mapping_method": "claim",
    			    	      "name": "my-ip",
    			    	      "google": {
    			    	        "client_id": "test-client",
    			    	        "client_secret": "test-secret",
                                "hosted_domain": "example.com"
    			    	      }
    			    	    }`),
							RespondWithJSON(http.StatusOK, `{
    			    	      "id": "456",
    			    	      "name": "my-ip",
                              "mapping_method": "claim",
    			    	      "google": {
    			    	        "client_id": "test-client",
    			    	        "client_secret": "test-secret",
                                "hosted_domain": "example.com"
    			    	      }
    			    	    }`),
						),
					)

					// Run the apply command:
					Terraform.Source(`
    		          resource "rhcs_identity_provider" "my_idp" {
    		            cluster = "123"
    		            name    = "my-ip"
    		            google = {
    		        	  client_id = "test-client"
    		        	  client_secret = "test-secret"
                          hosted_domain = "example.com"
    		            }
    		          }
    		        `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).To(BeZero())
				})

				Describe("Htpasswd IDP tests", func() {

					It("Should create htpasswd IDP", func() {
						// Prepare the server:
						TestServer.AppendHandlers(
							CombineHandlers(
								VerifyRequest(
									http.MethodPost,
									"/api/clusters_mgmt/v1/clusters/123/identity_providers",
								),
								VerifyJSON(`{
									"kind": "IdentityProvider",
									"htpasswd": {
										"users": {
											"items": [{
												"hashed_password": "hash(123PasS8901234)",
												"username": "my-user"
											},
											{
												"hashed_password": "hash(123PasS89012342)",
												"username": "my-user2"
											}]
										}
									},
									"mapping_method": "claim",
									"name": "my-ip",
									"type": "HTPasswdIdentityProvider"
								}`),
								RespondWithJSON(http.StatusOK, `{
									"id": "456",
									"name": "my-ip",
									"mapping_method": "claim",
									"htpasswd": {
										"client_id": "test-client",
										"client_secret": "test-secret",
										"hosted_domain": "example.com"
									}
								}`),
							),
						)

						// Run the apply command:
						Terraform.Source(`
	    	  				resource "rhcs_identity_provider" "my_idp" {
	    	    				cluster = "123"
					    	    name    = "my-ip"
	    					    htpasswd = {
    	              				users = [{
										username = "my-user",
										password = "` + htpasswdValidPass + `"
									},
									{
										username = "my-user2",
										password = "` + htpasswdValidPass2 + `"
									}]
								}
	    		  			}
	    				`)
						runOutput := Terraform.Apply()
						Expect(runOutput.ExitCode).To(BeZero())
					})

					It("Should delete htpasswd provider user (update)", func() {
						// Prepare the server:
						TestServer.AppendHandlers(
							CombineHandlers(
								VerifyRequest(
									http.MethodPost,
									"/api/clusters_mgmt/v1/clusters/123/identity_providers",
								),
								RespondWithPatchedJSON(http.StatusOK, template, `[{
									"kind": "IdentityProvider",
								    "op": "replace",
				    				"path": "/htpasswd/users",
									"mapping_method": "claim",
									"value": {
										"username": "my-user",
										"password": "`+htpasswdValidPass+`"
									}
								}]`),
							),
						)

						// Run the apply command:
						Terraform.Source(`
	    	  				resource "rhcs_identity_provider" "my_idp" {
	    	    				cluster = "123"
					    	    name    = "my-ip"
	    					    htpasswd = {
    	              				users = [{
										username = "my-user"
										password = "` + htpasswdValidPass + `"
									}]
								}
	    		  			}
	    				`)
						runOutput := Terraform.Apply()
						Expect(runOutput.ExitCode).To(BeZero())
					})

					It("Should edit htpasswd provider user's password (update)", func() {
						// Prepare the server:
						TestServer.AppendHandlers(
							CombineHandlers(
								VerifyRequest(
									http.MethodPost,
									"/api/clusters_mgmt/v1/clusters/123/identity_providers",
								),
								RespondWithPatchedJSON(http.StatusOK, template2, `[{
									"kind": "IdentityProvider",
								    "op": "replace",
				    				"path": "/htpasswd/users",
									"mapping_method": "claim",
									"value": {
										"username": "my-user",
										"password": "`+htpasswdValidPass2+`"
									}
								}]`),
							),
						)

						// Run the apply command:
						Terraform.Source(`
	    	  				resource "rhcs_identity_provider" "my_idp" {
	    	    				cluster = "123"
					    	    name    = "my-ip"
	    					    htpasswd = {
    	              				users = [{
										username = "my-user"
										password = "` + htpasswdValidPass2 + `"
									}]
								}
	    		  			}
	    				`)
						runOutput := Terraform.Apply()
						Expect(runOutput.ExitCode).To(BeZero())
					})

					It("Should add htpasswd provider user (update)", func() {
						// Prepare the server:
						TestServer.AppendHandlers(
							CombineHandlers(
								VerifyRequest(
									http.MethodPost,
									"/api/clusters_mgmt/v1/clusters/123/identity_providers",
								),
								RespondWithPatchedJSON(http.StatusOK, template3, `[{
									"kind": "IdentityProvider",
								    "op": "replace",
				    				"path": "/htpasswd/users",
									"mapping_method": "claim",
									"value": {
										"items": [{
											"username": "my-user",
											"password": "`+htpasswdValidPass2+`"
										},
										{
											"username": "my-user2",
											"password": "`+htpasswdValidPass2+`"
										}]
									}
								}]`),
							),
						)

						// Run the apply command:
						Terraform.Source(`
	    	  				resource "rhcs_identity_provider" "my_idp" {
	    	    				cluster = "123"
					    	    name    = "my-ip"
	    					    htpasswd = {
    	              				users = [{
											username = "my-user"
											password = "` + htpasswdValidPass2 + `"
										},
										{
											username = "my-user2",
											password = "` + htpasswdValidPass2 + `"
										}
									]
								}
	    		  			}
	    				`)
						runOutput := Terraform.Apply()
						Expect(runOutput.ExitCode).To(BeZero())
					})
				})

				It("Should create provider without hosted_domain when mapping_method is set to 'lookup'", func() {
					// Prepare the server:
					TestServer.AppendHandlers(
						CombineHandlers(
							VerifyRequest(
								http.MethodPost,
								"/api/clusters_mgmt/v1/clusters/123/identity_providers",
							),
							VerifyJSON(`{
    			    	      "kind": "IdentityProvider",
    			    	      "type": "GoogleIdentityProvider",
    			    	      "name": "my-ip",
                              "mapping_method": "lookup",
    			    	      "google": {
    			    	        "client_id": "test-client",
    			    	        "client_secret": "test-secret"
    			    	      }
    			    	    }`),
							RespondWithJSON(http.StatusOK, `{
    			    	      "id": "456",
    			    	      "name": "my-ip",
                              "mapping_method": "lookup",
    			    	      "google": {
    			    	        "client_id": "test-client",
    			    	        "client_secret": "test-secret"
    			    	      }
    			    	    }`),
						),
					)

					// Run the apply command:
					Terraform.Source(`
    		          resource "rhcs_identity_provider" "my_idp" {
    		            cluster = "123"
    		            name    = "my-ip"
                        mapping_method = "lookup"
    		            google = {
    		        	  client_id = "test-client"
    		        	  client_secret = "test-secret"
    		            }
    		          }
    		        `)
					runOutput := Terraform.Apply()
					Expect(runOutput.ExitCode).To(BeZero())
				})
			})
		})

		It("Can create an OpenID identity provider", func() {
			// Prepare the server:
			TestServer.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					),
					VerifyJSON(`{
    				  "kind": "IdentityProvider",
    				  "type": "OpenIDIdentityProvider",
                      "mapping_method": "claim",
    				  "name": "my-ip",
    				  "open_id": {
    					"ca": "test_ca",
    					"claims": {
    						"email": [
    							"email"
    						],
    						"groups": [
    							"admins"
    						],
    						"name": [
    							"name",
    							"email"
    						],
    						"preferred_username": [
    							"preferred_username",
    							"email"
    						]
    					},
    					"client_id": "test_client",
    					"client_secret": "test_secret",
    					"extra_authorize_parameters": {
    					  "test_key": "test_value"
    					},
    					"extra_scopes": [
    					  "email",
    					  "profile"
    					],
    					"issuer": "https://test.okta.com"
    					}
    				}`),
					RespondWithJSON(http.StatusOK, `{
    					"kind": "IdentityProvider",
    					"type": "OpenIDIdentityProvider",
    					"href": "/api/clusters_mgmt/v1/clusters/123/identity_providers/456",
    					"id": "456",
    					"name": "my-ip",
                        "mapping_method": "claim",
    					"open_id": {
    						"claims": {
    							"email": [
    								"email"
    							],
    							"groups": [
    								"admins"
    							],
    							"name": [
    								"name",
    								"email"
    							],
    							"preferred_username": [
    								"preferred_username",
    								"email"
    							]
    						},
    						"client_id": "test_client",
    						"extra_authorize_parameters": {
    							"test_key": "test_value"
    						},
    						"extra_scopes": [
    							"email",
    							"profile"
    						],
    						"issuer": "https://test.okta.com"
    					}
    				}`),
				),
			)

			// Run the apply command:
			Terraform.Source(`
    		  resource "rhcs_identity_provider" "my_idp" {
    		    cluster    				= "123"
    		    name       				= "my-ip"
    		    openid = {
    				ca            						= "test_ca"
    				issuer								= "https://test.okta.com"
    				client_id 							= "test_client"
    				client_secret						= "test_secret"
    				extra_scopes 						= ["email","profile"]
    				extra_authorize_parameters 			= {
    					test_key              = "test_value"
    		      	}
    				claims = {
    					email              = ["email"]
    					groups			   = ["admins"]
    					name               = ["name","email"]
    					preferred_username = ["preferred_username","email"]
    		      	}
    		    }
    		  }
    		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).To(BeZero())
			resource := Terraform.Resource("rhcs_identity_provider", "my_idp")
			Expect(resource).To(MatchJQ(`.attributes.openid.extra_authorize_parameters.test_key`, "test_value"))
		})

		It("Should fail with invalid mapping_method", func() {
			// Run the apply command:
			Terraform.Source(`
    		  resource "rhcs_identity_provider" "my_idp" {
    		    cluster = "123"
    		    name    = "my-ip"
                mapping_method = "invalid"
    		    htpasswd = {
                  users = [{
                    username = "my-user"
                    password = "` + htpasswdValidPass + `"
                  }]
    		    }
    		  }
    		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring(`Attribute mapping_method value must be one of:`)
		})
		It("Should fail with invalid htpasswd password", func() {
			// Run the apply command:
			Terraform.Source(`
    		  resource "rhcs_identity_provider" "my_idp" {
    		    cluster = "123"
    		    name    = "my-ip"
                mapping_method = "claim"
    		    htpasswd = {
                  users = [{
                    username = "my-user"
                    password = "` + htpasswdInValidPass + `"
                  }]
    		    }
    		  }
    		`)
			runOutput := Terraform.Apply()
			Expect(runOutput.ExitCode).ToNot(BeZero())
			runOutput.VerifyErrorContainsSubstring("Attribute htpasswd.users[0].password password must contain uppercase")
			runOutput.VerifyErrorContainsSubstring("Attribute htpasswd.users[0].password string length must be at least 14")
		})
	})
})

var _ = Describe("Identity provider import", func() {
	template := `{
	  "id": "123",
	  "external_id": "123",
	  "infra_id": "my-cluster-123",
	  "name": "my-cluster",
	  "domain_prefix": "my-cluster",
	  "state": "ready",
	  "region": {
	    "id": "us-west-1"
	  },
	  "aws": {
	    "ec2_metadata_http_tokens": "optional"
	  },
	  "multi_az": true,
	  "api": {
	    "url": "https://my-api.example.com"
	  },
	  "console": {
	    "url": "https://my-console.example.com"
	  },
      "properties": {
         "rosa_tf_version": "` + build.Version + `",
         "rosa_tf_commit": "` + build.Commit + `"
      },
	  "nodes": {
	    "compute": 3,
        "availability_zones": ["us-west-1a"],
	    "compute_machine_type": {
	      "id": "r5.xlarge"
	    }
	  },
	  "network": {
	    "machine_cidr": "10.0.0.0/16",
	    "service_cidr": "172.30.0.0/16",
	    "pod_cidr": "10.128.0.0/14",
	    "host_prefix": 23
	  },
	  
	  "version": {
		  "id": "openshift-4.8.0"
	  },
      "dns" : {
          "base_domain": "mycluster-api.example.com"
      }
	}`

	It("Can import an identity provider", func() {
		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, template),
			),
			// List IDPs to map name to ID:
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/identity_providers",
				),
				RespondWithJSON(http.StatusOK, `{
					"kind": "IdentityProviderList",
					"href": "/api/clusters_mgmt/v1/clusters/24vg6o424djht8h6lpoli2urg69t7vnt/identity_providers",
					"page": 1,
					"size": 1,
					"total": 1,
					"items": [
						{
						"kind": "IdentityProvider",
						"type": "GithubIdentityProvider",
						"href": "/api/clusters_mgmt/v1/clusters/24vg6o424djht8h6lpoli2urg69t7vnt/identity_providers/24vgs9hgnl5bukujvkcmgkvfgc01ss0r",
						"id": "24vgs9hgnl5bukujvkcmgkvfgc01ss0r",
						"name": "my-ip",
						"mapping_method": "claim",
						"github": {
							"client_id": "99999",
							"organizations": [
								"myorg"
							]
						}
						}
					]
				}`),
			),
			// Read the IDP to load the current state:
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/identity_providers/24vgs9hgnl5bukujvkcmgkvfgc01ss0r",
				),
				RespondWithJSON(http.StatusOK, `{
					"kind": "IdentityProvider",
					"type": "GithubIdentityProvider",
					"href": "/api/clusters_mgmt/v1/clusters/24vg6o424djht8h6lpoli2urg69t7vnt/identity_providers/24vgs9hgnl5bukujvkcmgkvfgc01ss0r",
					"id": "24vgs9hgnl5bukujvkcmgkvfgc01ss0r",
					"name": "my-ip",
					"mapping_method": "claim",
					"github": {
						"client_id": "99999",
						"organizations": [
							"myorg"
						]
					}
				}`),
			),
		)

		Terraform.Source(`
			resource "rhcs_identity_provider" "my-ip" {
				# (resource arguments)
			}
		`)
		runOutput := Terraform.Import("rhcs_identity_provider.my-ip", "123,my-ip")
		Expect(runOutput.ExitCode).To(BeZero())
		resource := Terraform.Resource("rhcs_identity_provider", "my-ip")
		Expect(resource).To(MatchJQ(".attributes.name", "my-ip"))
		Expect(resource).To(MatchJQ(".attributes.github.client_id", "99999"))
	})

	It("Is an error if the identity provider isn't found", func() {
		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, template),
			),
			// List IDPs to map name to ID:
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/identity_providers",
				),
				RespondWithJSON(http.StatusOK, `{
					"kind": "IdentityProviderList",
					"href": "/api/clusters_mgmt/v1/clusters/24vg6o424djht8h6lpoli2urg69t7vnt/identity_providers",
					"page": 1,
					"size": 1,
					"total": 1,
					"items": [
						{
						"kind": "IdentityProvider",
						"type": "GithubIdentityProvider",
						"href": "/api/clusters_mgmt/v1/clusters/24vg6o424djht8h6lpoli2urg69t7vnt/identity_providers/24vgs9hgnl5bukujvkcmgkvfgc01ss0r",
						"id": "24vgs9hgnl5bukujvkcmgkvfgc01ss0r",
						"name": "my-ip",
						"mapping_method": "claim",
						"github": {
							"client_id": "99999",
							"organizations": [
								"myorg"
							]
						}
						}
					]
				}`),
			),
		)

		Terraform.Source(`
			resource "rhcs_identity_provider" "my-ip" {
				# (resource arguments)
			}
		`)
		runOutput := Terraform.Import("rhcs_identity_provider.my-ip", "123,notfound")
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("identity provider 'notfound' not found")
	})

	It("import for non exist cluster - should fail", func() {
		// Prepare the server:
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusNotFound, template),
			),
		)

		Terraform.Source(`
			resource "rhcs_identity_provider" "my-ip" {
				# (resource arguments)
			}
		`)
		runOutput := Terraform.Import("rhcs_identity_provider.my-ip", "123,notfound")
		Expect(runOutput.ExitCode).NotTo(BeZero())
		runOutput.VerifyErrorContainsSubstring("Cluster 123 not found, error: status is 404 and identifier is '123'")
	})
})
