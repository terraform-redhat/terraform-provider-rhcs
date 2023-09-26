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

package provider

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

const htpasswdValidPass = "123PasS8901234"
const htpasswdInValidPass = "my-pass"

var _ = Describe("Identity provider creation", func() {

	Context("Idebtity Provider Failure", func() {
		It("cluster_id not found", func() {
			// Prepare the server:
			server.AppendHandlers(
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
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
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
			Expect(terraform.Apply()).ToNot(BeZero())
		})
		Context("Cluster exists, but invalid config", func() {
			BeforeEach(func() {
				// The first thing that the provider will do for any operation on identity providers
				// is check that the cluster is ready, so we always need to prepare the server to
				// respond to that:
				server.AppendHandlers(
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
				terraform.Source(`
	    	      resource "rhcs_identity_provider" "my_ip" {
	    	        cluster = "123"
	    	        name    = "my-ip"
	    	        htpasswd = {
                      users = []
	    	        }
	    	      }
	    	    `)
				Expect(terraform.Apply()).ToNot(BeZero())
			})
			It("Can't create a 'htpasswd' identity provider. duplication of username", func() {
				// Run the apply command:
				terraform.Source(`
	    	      resource "rhcs_identity_provider" "my_ip" {
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
				Expect(terraform.Apply()).ToNot(BeZero())
			})
			It("Can't create a 'htpasswd' identity provider. invalid username", func() {
				// Run the apply command:
				terraform.Source(`
	    	      resource "rhcs_identity_provider" "my_ip" {
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
				Expect(terraform.Apply()).ToNot(BeZero())
			})
			It("Can't create a 'htpasswd' identity provider. invalid password", func() {
				// Run the apply command:
				terraform.Source(`
	    	      resource "rhcs_identity_provider" "my_ip" {
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
				Expect(terraform.Apply()).ToNot(BeZero())
			})
		})
	})

	Context("Identity Provider Success", func() {
		BeforeEach(func() {
			// The first thing that the provider will do for any operation on identity providers
			// is check that the cluster is ready, so we always need to prepare the server to
			// respond to that:
			server.AppendHandlers(
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
			server.AppendHandlers(
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
                        "users": {"items":[{"username": "my-user", "password": "`+htpasswdValidPass+`"}]}
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
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
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
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Can't update an identity provider", func() {
			// Prepare the server:
			server.AppendHandlers(
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
                        "users": {"items":[{"username": "my-user", "password": "`+htpasswdValidPass+`"}]}
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
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
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
			Expect(terraform.Apply()).To(BeZero())

			// update

			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodGet,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers/456",
					),
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
			// Run the apply command for update:
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
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
			Expect(terraform.Apply()).ToNot(BeZero())
		})
		It("Reconcile an 'htpasswd' identity provider, when state exists but 404 from server", func() {
			// Prepare the server:
			server.AppendHandlers(
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
                        "users": {"items":[{"username": "my-user", "password": "`+htpasswdValidPass+`"}]}
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
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
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
			Expect(terraform.Apply()).To(BeZero())

			// Prepare the server for upgrade
			server.AppendHandlers(
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
                        "users": {"items":[{"username": "my-user", "password": "`+htpasswdValidPass+`"}]}
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
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
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
			Expect(terraform.Apply()).To(BeZero())
			resource := terraform.Resource("rhcs_identity_provider", "my_ip")
			Expect(resource).To(MatchJQ(".attributes.id", "457"))
		})

		It("Can create a 'gitlab' identity provider", func() {
			// Prepare the server:
			server.AppendHandlers(
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
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
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
			Expect(terraform.Apply()).To(BeZero())
		})

		Context("Can create a 'github' identity provider", func() {
			Context("Invalid 'github' identity provider config", func() {
				It("Should fail with both 'teams' and 'organizations'", func() {
					terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_ip" {
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
					Expect(terraform.Apply()).ToNot(BeZero())
				})

				It("Should fail without 'teams' or 'organizations'", func() {
					terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_ip" {
	    	            cluster = "123"
	    	            name    = "my-ip"
	    	            github = {
	    	              ca = "test-ca"
	    	        	  client_id = "test-client"
	    	        	  client_secret = "test-secret"
	    	            }
	    	          }
	    	        `)
					Expect(terraform.Apply()).ToNot(BeZero())
				})

				It("Should fail if teams contain an invalid format", func() {
					terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_ip" {
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
					Expect(terraform.Apply()).ToNot(BeZero())
					terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_ip" {
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
					Expect(terraform.Apply()).ToNot(BeZero())
				})

				It("Should fail with an invalid hostname", func() {
					terraform.Source(`
	    	          resource "rhcs_identity_provider" "my_ip" {
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
					Expect(terraform.Apply()).ToNot(BeZero())
				})
			})
			It("Happy flow with org restriction", func() {
				// Prepare the server:
				server.AppendHandlers(
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
				terraform.Source(`
    		      resource "rhcs_identity_provider" "my_ip" {
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
				Expect(terraform.Apply()).To(BeZero())
			})

			It("Happy flow with team restriction", func() {
				// Prepare the server:
				server.AppendHandlers(
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
				terraform.Source(`
		          resource "rhcs_identity_provider" "my_ip" {
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
				Expect(terraform.Apply()).To(BeZero())
			})
		})

		Context("Can create 'LDDAP' Identity provider", func() {
			Context("Invalid LDAP config", func() {
				It("Should fail with invalid email", func() {
					// Run the apply command:
					terraform.Source(`
        		      resource "rhcs_identity_provider" "my_ip" {
        		        cluster    = "123"
        		        name       = "my-ip"
        		        ldap = {
        		          insecure      = false
        		          ca            = "my-ca"
        		          url           = "ldap://my-server.com"
        		          attributes    = {
        		            id                 = ["my-id"]
        		            email              = ["my-email"]
        		            name               = ["my-name"]
        		            preferred_username = ["my-preferred-username"]
        		          }
        		        }
        		      }
        		    `)
					Expect(terraform.Apply()).ToNot(BeZero())
				})
				It("Should fail if not both bind properties are set", func() {
					// Run the apply command:
					terraform.Source(`
        		      resource "rhcs_identity_provider" "my_ip" {
        		        cluster    = "123"
        		        name       = "my-ip"
        		        ldap = {
        		          bind_dn       = "my-bind-dn"
        		          insecure      = false
        		          ca            = "my-ca"
        		          url           = "ldap://my-server.com"
        		          attributes    = {
        		            id                 = ["my-id"]
        		            email              = ["my@email.com"]
        		            name               = ["my-name"]
        		            preferred_username = ["my-preferred-username"]
        		          }
        		        }
        		      }
        		    `)
					Expect(terraform.Apply()).ToNot(BeZero())
				})

			})
			It("Happy flow with default attributes", func() {
				// Prepare the server:
				server.AppendHandlers(
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
				              "name": ["cn"],
				              "preferred_username": ["uid"]
				            }
				          }
				        }`),
					),
				)

				// Run the apply command:
				terraform.Source(`
        		  resource "rhcs_identity_provider" "my_ip" {
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
				Expect(terraform.Apply()).To(BeZero())
			})
			It("Happy flow with bind values", func() {
				// Prepare the server:
				server.AppendHandlers(
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
				              "id": ["my-id"],
				              "email": ["my@email.com"],
				              "name": ["my-name"],
				              "preferred_username": ["my-preferred-username"]
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
				              "id": ["my-id"],
				              "email": ["my@email.com"],
				              "name": ["my-name"],
				              "preferred_username": ["my-preferred-username"]
				            }
				          }
				        }`),
					),
				)

				// Run the apply command:
				terraform.Source(`
        		  resource "rhcs_identity_provider" "my_ip" {
        		    cluster    = "123"
        		    name       = "my-ip"
        		    ldap = {
        		      bind_dn       = "my-bind-dn"
        		      bind_password = "my-bind-password"
        		      insecure      = false
        		      ca            = "my-ca"
        		      url           = "ldap://my-server.com"
        		      attributes    = {
        		        id                 = ["my-id"]
        		        email              = ["my@email.com"]
        		        name               = ["my-name"]
        		        preferred_username = ["my-preferred-username"]
        		      }
        		    }
        		  }
        		`)
				Expect(terraform.Apply()).To(BeZero())
			})

			It("Happy flow without bind values", func() {
				// Prepare the server:
				server.AppendHandlers(
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
				              "id": ["my-id"],
				              "email": ["my@email.com"],
				              "name": ["my-name"],
				              "preferred_username": ["my-preferred-username"]
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
				              "id": ["my-id"],
				              "email": ["my@email.com"],
				              "name": ["my-name"],
				              "preferred_username": ["my-preferred-username"]
				            }
				          }
				        }`),
					),
				)

				// Run the apply command:
				terraform.Source(`
        		  resource "rhcs_identity_provider" "my_ip" {
        		    cluster    = "123"
        		    name       = "my-ip"
        		    ldap = {
        		      insecure      = false
        		      ca            = "my-ca"
        		      url           = "ldap://my-server.com"
        		      attributes    = {
        		        id                 = ["my-id"]
        		        email              = ["my@email.com"]
        		        name               = ["my-name"]
        		        preferred_username = ["my-preferred-username"]
        		      }
        		    }
        		  }
        		`)
				Expect(terraform.Apply()).To(BeZero())
			})
		})

		Context("Google identity provider", func() {
			Context("Invalid google config", func() {
				It("Should fail with invalid hosted_domain", func() {
					// Run the apply command:
					terraform.Source(`
    		          resource "rhcs_identity_provider" "my_ip" {
    		            cluster = "123"
    		            name    = "my-ip"
    		            google = {
    		        	  client_id = "test-client"
    		        	  client_secret = "test-secret"
                          hosted_domain = "examplecom"
    		            }
    		          }
    		        `)
					Expect(terraform.Apply()).ToNot(BeZero())
				})

				It("Should fail when mapping_method is not lookup and no hosted_domain", func() {
					// Run the apply command:
					terraform.Source(`
    		          resource "rhcs_identity_provider" "my_ip" {
    		            cluster = "123"
    		            name    = "my-ip"
    		            google = {
    		        	  client_id = "test-client"
    		        	  client_secret = "test-secret"
    		            }
    		          }
    		        `)
					Expect(terraform.Apply()).ToNot(BeZero())
				})

			})

			Context("Happy flow", func() {
				It("Should create provider", func() {
					// Prepare the server:
					server.AppendHandlers(
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
					terraform.Source(`
    		          resource "rhcs_identity_provider" "my_ip" {
    		            cluster = "123"
    		            name    = "my-ip"
    		            google = {
    		        	  client_id = "test-client"
    		        	  client_secret = "test-secret"
                          hosted_domain = "example.com"
    		            }
    		          }
    		        `)
					Expect(terraform.Apply()).To(BeZero())
				})

				It("Should create provider without hosted_domain when mapping_method is set to 'lookup'", func() {
					// Prepare the server:
					server.AppendHandlers(
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
					terraform.Source(`
    		          resource "rhcs_identity_provider" "my_ip" {
    		            cluster = "123"
    		            name    = "my-ip"
                        mapping_method = "lookup"
    		            google = {
    		        	  client_id = "test-client"
    		        	  client_secret = "test-secret"
    		            }
    		          }
    		        `)
					Expect(terraform.Apply()).To(BeZero())
				})
			})
		})

		It("Can create an OpenID identity provider", func() {
			// Prepare the server:
			server.AppendHandlers(
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
			terraform.Source(`
    		  resource "rhcs_identity_provider" "my_ip" {
    		    cluster    				= "123"
    		    name       				= "my-ip"
    		    openid = {
    				ca            			= "test_ca"
    				issuer					= "https://test.okta.com"
    				client_id 				= "test_client"
    				client_secret			= "test_secret"
    				extra_scopes 			= ["email","profile"]
    				claims = {
    					email              = ["email"]
    					groups			   = ["admins"]
    					name               = ["name","email"]
    					preferred_username = ["preferred_username","email"]
    		      	}
    		    }
    		  }
    		`)
			Expect(terraform.Apply()).To(BeZero())
		})

		It("Should fail with invalid mapping_method", func() {
			// Run the apply command:
			terraform.Source(`
    		  resource "rhcs_identity_provider" "my_ip" {
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
			Expect(terraform.Apply()).ToNot(BeZero())
		})
		It("Should fail with invalid htpasswd password", func() {
			// Run the apply command:
			terraform.Source(`
    		  resource "rhcs_identity_provider" "my_ip" {
    		    cluster = "123"
    		    name    = "my-ip"
                mapping_method = "invalid"
    		    htpasswd = {
                  users = [{
                    username = "my-user"
                    password = "` + htpasswdInValidPass + `"
                  }]
    		    }
    		  }
    		`)
			Expect(terraform.Apply()).ToNot(BeZero())
		})
	})
})

var _ = Describe("Identity provider import", func() {
	It("Can import an identity provider", func() {
		// Prepare the server:
		server.AppendHandlers(
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

		terraform.Source(`
			resource "rhcs_identity_provider" "my-ip" {
				# (resource arguments)
			}
		`)

		Expect(terraform.Import("rhcs_identity_provider.my-ip", "123,my-ip")).To(BeZero())
		resource := terraform.Resource("rhcs_identity_provider", "my-ip")
		Expect(resource).To(MatchJQ(".attributes.name", "my-ip"))
		Expect(resource).To(MatchJQ(".attributes.github.client_id", "99999"))
	})

	It("Is an error if the identity provider isn't found", func() {
		// Prepare the server:
		server.AppendHandlers(
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

		terraform.Source(`
			resource "rhcs_identity_provider" "my-ip" {
				# (resource arguments)
			}
		`)

		Expect(terraform.Import("rhcs_identity_provider.my-ip", "123,notfound")).NotTo(BeZero())
	})
})
