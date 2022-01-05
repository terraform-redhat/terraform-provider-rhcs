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

***REMOVED***                      // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

var _ = Describe("Identity provider creation", func(***REMOVED*** {
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
		***REMOVED***
	}***REMOVED***

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
				  "name": "my-ip",
				  "htpasswd": {
				    "password": "my-password",
				    "username": "my-user"
				  }
		***REMOVED***`***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
				  "id": "456",
				  "name": "my-ip",
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
	}***REMOVED***

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
	}***REMOVED***
}***REMOVED***
