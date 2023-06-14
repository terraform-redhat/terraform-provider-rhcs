/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
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

***REMOVED***
***REMOVED***

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

var _ = Describe("Identity provider creation", func(***REMOVED*** {

	Context("Idebtity Provider Failure", func(***REMOVED*** {
		It("cluster_id not found", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, `{
				  "id": "123",
				  "name": "my-cluster",
				  "state": "ready"
		***REMOVED***`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "ocm_identity_provider" "my_ip" {
		    cluster = "123"
		    name    = "my-ip"
		    htpasswd = {
		      username = "my-user"
		      password = "my-password"
		    }
		  }
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.Should(BeNumerically("==", 1***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	Context("Idebtity Provider Success", func(***REMOVED*** {
		BeforeEach(func(***REMOVED*** {
			// The first thing that the provider will do for any operation on identity providers
			// is check that the cluster is ready, so we always need to prepare the server to
			// respond to that:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
				  "id": "123",
				  "name": "my-cluster",
				  "state": "ready"
		***REMOVED***`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
				  "id": "123",
				  "name": "my-cluster",
				  "state": "ready"
		***REMOVED***`***REMOVED***,
				***REMOVED***,
			***REMOVED***
***REMOVED******REMOVED***

		It("Can create a 'htpasswd' identity provider", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					***REMOVED***,
					VerifyJSON(`{
				  "kind": "IdentityProvider",
				  "type": "HTPasswdIdentityProvider",
                  "mapping_method": "claim",
				  "name": "my-ip",
				  "htpasswd": {
				    "password": "my-password",
				    "username": "my-user"
				  }
		***REMOVED***`***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
				  "id": "456",
				  "name": "my-ip",
                  "mapping_method": "claim",
				  "htpasswd": {
				    "user": "my-user"
				  }
		***REMOVED***`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "ocm_identity_provider" "my_ip" {
		    cluster = "123"
		    name    = "my-ip"
		    htpasswd = {
		      username = "my-user"
		      password = "my-password"
		    }
		  }
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Can create a 'gitlab' identity provider", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					***REMOVED***,
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
		***REMOVED***`***REMOVED***,
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
		***REMOVED***`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "ocm_identity_provider" "my_ip" {
		    cluster = "123"
		    name    = "my-ip"
		    gitlab = {
		      ca = "test-ca"
		      url = "https://test.gitlab.com"
			  client_id = "test-client"
			  client_secret = "test-secret"
		    }
		  }
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		Context("Can create a 'github' identity provider", func(***REMOVED*** {
			Context("Invalid 'github' identity provider config", func(***REMOVED*** {
				It("Should fail with both 'teams' and 'organizations'", func(***REMOVED*** {
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
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
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should fail without 'teams' or 'organizations'", func(***REMOVED*** {
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
		            cluster = "123"
		            name    = "my-ip"
		            github = {
		              ca = "test-ca"
		        	  client_id = "test-client"
		        	  client_secret = "test-secret"
		            }
		          }
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should fail if teams contain an invalid format", func(***REMOVED*** {
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
		            cluster = "123"
		            name    = "my-ip"
		            github = {
		              ca = "test-ca"
		        	  client_id = "test-client"
		        	  client_secret = "test-secret"
                      teams = ["invalidteam"]
		            }
		          }
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
		            cluster = "123"
		            name    = "my-ip"
		            github = {
		              ca = "test-ca"
		        	  client_id = "test-client"
		        	  client_secret = "test-secret"
                      teams = ["valid/team", "invalidteam"]
		            }
		          }
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should fail with an invalid hostname", func(***REMOVED*** {
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
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
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***
	***REMOVED******REMOVED***
			It("Happy flow with org restriction", func(***REMOVED*** {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(
							http.MethodPost,
							"/api/clusters_mgmt/v1/clusters/123/identity_providers",
						***REMOVED***,
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
				    }`***REMOVED***,
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
				    }`***REMOVED***,
					***REMOVED***,
				***REMOVED***

				// Run the apply command:
				terraform.Source(`
		      resource "ocm_identity_provider" "my_ip" {
		        cluster = "123"
		        name    = "my-ip"
		        github = {
		          ca = "test-ca"
		    	  client_id = "test-client"
		    	  client_secret = "test-secret"
                  organizations = ["my-org"]
		        }
		      }
		    `***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***

			It("Happy flow with team restriction", func(***REMOVED*** {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(
							http.MethodPost,
							"/api/clusters_mgmt/v1/clusters/123/identity_providers",
						***REMOVED***,
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
				    }`***REMOVED***,
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
				    }`***REMOVED***,
					***REMOVED***,
				***REMOVED***

				// Run the apply command:
				terraform.Source(`
		      resource "ocm_identity_provider" "my_ip" {
		        cluster = "123"
		        name    = "my-ip"
		        github = {
		          ca = "test-ca"
		    	  client_id = "test-client"
		    	  client_secret = "test-secret"
                  teams = ["valid/team"]
		        }
		      }
		    `***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Can create an LDAP identity provider", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					***REMOVED***,
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
				      "email": ["my-email"],
				      "name": ["my-name"],
				      "preferred_username": ["my-preferred-username"]
				    }
				  }
		***REMOVED***`***REMOVED***,
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
				      "email": ["my-email"],
				      "name": ["my-name"],
				      "preferred_username": ["my-preferred-username"]
				    }
				  }
		***REMOVED***`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "ocm_identity_provider" "my_ip" {
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
		        email              = ["my-email"]
		        name               = ["my-name"]
		        preferred_username = ["my-preferred-username"]
		      }
		    }
		  }
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		Context("Google identity provider", func(***REMOVED*** {
			Context("Invalid google config", func(***REMOVED*** {
				It("Should fail with invalid hosted_domain", func(***REMOVED*** {
					// Run the apply command:
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
		            cluster = "123"
		            name    = "my-ip"
		            google = {
		        	  client_id = "test-client"
		        	  client_secret = "test-secret"
                      hosted_domain = "examplecom"
		            }
		          }
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should fail when mapping_method is not lookup and no hosted_domain", func(***REMOVED*** {
					// Run the apply command:
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
		            cluster = "123"
		            name    = "my-ip"
		            google = {
		        	  client_id = "test-client"
		        	  client_secret = "test-secret"
		            }
		          }
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

	***REMOVED******REMOVED***

			Context("Happy flow", func(***REMOVED*** {
				It("Should create provider", func(***REMOVED*** {
					// Prepare the server:
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(
								http.MethodPost,
								"/api/clusters_mgmt/v1/clusters/123/identity_providers",
							***REMOVED***,
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
			    	    }`***REMOVED***,
							RespondWithJSON(http.StatusOK, `{
			    	      "id": "456",
			    	      "name": "my-ip",
                          "mapping_method": "claim",
			    	      "google": {
			    	        "client_id": "test-client",
			    	        "client_secret": "test-secret",
                            "hosted_domain": "example.com"
			    	      }
			    	    }`***REMOVED***,
						***REMOVED***,
					***REMOVED***

					// Run the apply command:
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
		            cluster = "123"
		            name    = "my-ip"
		            google = {
		        	  client_id = "test-client"
		        	  client_secret = "test-secret"
                      hosted_domain = "example.com"
		            }
		          }
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should create provider without hosted_domain when mapping_method is set to 'lookup'", func(***REMOVED*** {
					// Prepare the server:
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(
								http.MethodPost,
								"/api/clusters_mgmt/v1/clusters/123/identity_providers",
							***REMOVED***,
							VerifyJSON(`{
			    	      "kind": "IdentityProvider",
			    	      "type": "GoogleIdentityProvider",
			    	      "name": "my-ip",
                          "mapping_method": "lookup",
			    	      "google": {
			    	        "client_id": "test-client",
			    	        "client_secret": "test-secret"
			    	      }
			    	    }`***REMOVED***,
							RespondWithJSON(http.StatusOK, `{
			    	      "id": "456",
			    	      "name": "my-ip",
                          "mapping_method": "lookup",
			    	      "google": {
			    	        "client_id": "test-client",
			    	        "client_secret": "test-secret"
			    	      }
			    	    }`***REMOVED***,
						***REMOVED***,
					***REMOVED***

					// Run the apply command:
					terraform.Source(`
		          resource "ocm_identity_provider" "my_ip" {
		            cluster = "123"
		            name    = "my-ip"
                    mapping_method = "lookup"
		            google = {
		        	  client_id = "test-client"
		        	  client_secret = "test-secret"
		            }
		          }
		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Can create an OpenID identity provider", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodPost,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers",
					***REMOVED***,
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
			***REMOVED***,
					"client_id": "test_client",
					"client_secret": "test_secret",
					"extra_scopes": [
					  "email",
					  "profile"
					],
					"issuer": "https://test.okta.com"
			***REMOVED***
		***REMOVED***`***REMOVED***,
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
				***REMOVED***,
						"client_id": "test_client",
						"extra_scopes": [
							"email",
							"profile"
						],
						"issuer": "https://test.okta.com"
			***REMOVED***
		***REMOVED***`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "ocm_identity_provider" "my_ip" {
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
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Should fail with invalid mapping_method", func(***REMOVED*** {
			// Run the apply command:
			terraform.Source(`
		  resource "ocm_identity_provider" "my_ip" {
		    cluster = "123"
		    name    = "my-ip"
            mapping_method = "invalid"
		    htpasswd = {
		      username = "my-user"
		      password = "my-password"
		    }
		  }
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
