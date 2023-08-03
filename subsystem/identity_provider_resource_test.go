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

const htpasswdValidPass = "123PasS8901234"
const htpasswdInValidPass = "my-pass"

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
			    	}`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
	    	      username = "my-user"
	    	      password = "` + htpasswdValidPass + `"
	    	    }
	    	  }
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.Should(BeNumerically("==", 1***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	Context("Identity Provider Success", func(***REMOVED*** {
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
			    	}`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "123",
			    	  "name": "my-cluster",
			    	  "state": "ready"
			    	}`***REMOVED***,
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
                        "password": "`+htpasswdValidPass+`",
			    	    "username": "my-user"
			    	  }
			    	}`***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "456",
			    	  "name": "my-ip",
                      "mapping_method": "claim",
			    	  "htpasswd": {
			    	    "user": "my-user"
			    	  }
			    	}`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
	    	      username = "my-user"
	    	      password = "` + htpasswdValidPass + `"
	    	    }
	    	  }
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Can't update an identity provider", func(***REMOVED*** {
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
                        "password": "`+htpasswdValidPass+`",
			    	    "username": "my-user"
			    	  }
			    	}`***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "456",
			    	  "name": "my-ip",
                      "mapping_method": "claim",
			    	  "htpasswd": {
			    	    "user": "my-user"
			    	  }
			    	}`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
	    	      username = "my-user"
	    	      password = "` + htpasswdValidPass + `"
	    	    }
	    	  }
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

			// update

			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(
						http.MethodGet,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers/456",
					***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "456",
			    	  "name": "my-ip",
                      "mapping_method": "claim",
			    	  "htpasswd": {
			    	    "user": "my-user"
			    	  }
			    	}`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			// Run the apply command for update:
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
	    	      username = "my-user-change"
	    	      password = "` + htpasswdValidPass + `"
	    	    }
	    	  }
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Reconcile an 'htpasswd' identity provider, when state exists but 404 from server", func(***REMOVED*** {
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
                        "password": "`+htpasswdValidPass+`",
			    	    "username": "my-user"
			    	  }
			    	}`***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "456",
			    	  "name": "my-ip",
                      "mapping_method": "claim",
			    	  "htpasswd": {
			    	    "user": "my-user"
			    	  }
			    	}`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
	    	      username = "my-user"
	    	      password = "` + htpasswdValidPass + `"
	    	    }
	    	  }
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

			// Prepare the server for upgrade
			server.AppendHandlers(
				// read from server (404***REMOVED***
				CombineHandlers(
					VerifyRequest(
						http.MethodGet,
						"/api/clusters_mgmt/v1/clusters/123/identity_providers/456",
					***REMOVED***,
					RespondWithJSON(http.StatusNotFound, "{}"***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "123",
			    	  "name": "my-cluster",
			    	  "state": "ready"
			    	}`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "123",
			    	  "name": "my-cluster",
			    	  "state": "ready"
			    	}`***REMOVED***,
				***REMOVED***,
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
                        "password": "`+htpasswdValidPass+`",
			    	    "username": "my-user"
			    	  }
			    	}`***REMOVED***,
					RespondWithJSON(http.StatusOK, `{
			    	  "id": "457",
			    	  "name": "my-ip",
                      "mapping_method": "claim",
			    	  "htpasswd": {
			    	    "user": "my-user"
			    	  }
			    	}`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
	    	  resource "rhcs_identity_provider" "my_ip" {
	    	    cluster = "123"
	    	    name    = "my-ip"
	    	    htpasswd = {
	    	      username = "my-user"
	    	      password = "` + htpasswdValidPass + `"
	    	    }
	    	  }
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			resource := terraform.Resource("rhcs_identity_provider", "my_ip"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "457"***REMOVED******REMOVED***
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
	    	`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		Context("Can create a 'github' identity provider", func(***REMOVED*** {
			Context("Invalid 'github' identity provider config", func(***REMOVED*** {
				It("Should fail with both 'teams' and 'organizations'", func(***REMOVED*** {
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
	    	        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should fail without 'teams' or 'organizations'", func(***REMOVED*** {
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
	    	        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should fail if teams contain an invalid format", func(***REMOVED*** {
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
	    	        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
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
	    	        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should fail with an invalid hostname", func(***REMOVED*** {
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
		        `***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***

		Context("Can create 'LDDAP' Identity provider", func(***REMOVED*** {
			Context("Invalid LDAP config", func(***REMOVED*** {
				It("Should fail with invalid email", func(***REMOVED*** {
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
        		    `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***
				It("Should fail if not both bind properties are set", func(***REMOVED*** {
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
        		    `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

	***REMOVED******REMOVED***
			It("Happy flow with default attributes", func(***REMOVED*** {
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
				            "ca": "my-ca",
				            "insecure": false,
				            "url": "ldap://my-server.com",
				            "attributes": {
				              "id": ["dn"],
				              "name": ["cn"],
				              "preferred_username": ["uid"]
				            }
				          }
				        }`***REMOVED***,
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
				        }`***REMOVED***,
					***REMOVED***,
				***REMOVED***

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
        		`***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
			It("Happy flow with bind values", func(***REMOVED*** {
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
				              "email": ["my@email.com"],
				              "name": ["my-name"],
				              "preferred_username": ["my-preferred-username"]
				            }
				          }
				        }`***REMOVED***,
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
				        }`***REMOVED***,
					***REMOVED***,
				***REMOVED***

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
        		`***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***

			It("Happy flow without bind values", func(***REMOVED*** {
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
				        }`***REMOVED***,
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
				        }`***REMOVED***,
					***REMOVED***,
				***REMOVED***

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
        		`***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***

		Context("Google identity provider", func(***REMOVED*** {
			Context("Invalid google config", func(***REMOVED*** {
				It("Should fail with invalid hosted_domain", func(***REMOVED*** {
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
    		        `***REMOVED***
					Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
		***REMOVED******REMOVED***

				It("Should fail when mapping_method is not lookup and no hosted_domain", func(***REMOVED*** {
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
    		          resource "rhcs_identity_provider" "my_ip" {
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
    		          resource "rhcs_identity_provider" "my_ip" {
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
    		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Should fail with invalid mapping_method", func(***REMOVED*** {
			// Run the apply command:
			terraform.Source(`
    		  resource "rhcs_identity_provider" "my_ip" {
    		    cluster = "123"
    		    name    = "my-ip"
                mapping_method = "invalid"
    		    htpasswd = {
    		      username = "my-user"
	    	      password = "` + htpasswdValidPass + `"
    		    }
    		  }
    		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("Should fail with invalid htpasswd password", func(***REMOVED*** {
			// Run the apply command:
			terraform.Source(`
    		  resource "rhcs_identity_provider" "my_ip" {
    		    cluster = "123"
    		    name    = "my-ip"
                mapping_method = "invalid"
    		    htpasswd = {
    		      username = "my-user"
	    	      password = "` + htpasswdInValidPass + `"
    		    }
    		  }
    		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("Identity provider import", func(***REMOVED*** {
	It("Can import an identity provider", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			// List IDPs to map name to ID:
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/identity_providers",
				***REMOVED***,
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
				***REMOVED***
				***REMOVED***
					]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Read the IDP to load the current state:
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/identity_providers/24vgs9hgnl5bukujvkcmgkvfgc01ss0r",
				***REMOVED***,
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
			***REMOVED***
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		terraform.Source(`
			resource "rhcs_identity_provider" "my-ip" {
				# (resource arguments***REMOVED***
	***REMOVED***
		`***REMOVED***

		Expect(terraform.Import("rhcs_identity_provider.my-ip", "123,my-ip"***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("rhcs_identity_provider", "my-ip"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.name", "my-ip"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.github.client_id", "99999"***REMOVED******REMOVED***
	}***REMOVED***

	It("Is an error if the identity provider isn't found", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			// List IDPs to map name to ID:
			CombineHandlers(
				VerifyRequest(
					http.MethodGet,
					"/api/clusters_mgmt/v1/clusters/123/identity_providers",
				***REMOVED***,
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
				***REMOVED***
				***REMOVED***
					]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		terraform.Source(`
			resource "rhcs_identity_provider" "my-ip" {
				# (resource arguments***REMOVED***
	***REMOVED***
		`***REMOVED***

		Expect(terraform.Import("rhcs_identity_provider.my-ip", "123,notfound"***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
