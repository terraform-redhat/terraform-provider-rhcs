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

const managedOidcConfig = `{
  "href": "/api/clusters_mgmt/v1/oidc_configs/23f6gk51qi5ng15mm095c90hhajbf7c5",
  "id": "23f6gk51qi5ng15mm095c90hhajbf7c5",
  "issuer_url": "https://d3gt1gce2zmg3d.cloudfront.net/23f6gk51qi5ng15mm095c90hhajbf7c5",
  "managed": true,
  "reusable": true
}`

const unManagedOidcConfig = `{
  "href": "/api/clusters_mgmt/v1/oidc_configs/23f6gk51qi5ng15mm095c90hhajbf7c5",
  "id": "23f6gk51qi5ng15mm095c90hhajbf7c5",
  "issuer_url": "https://oidc-f3y4.s3.us-east-1.amazonaws.com",
  "secret_arn": "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c",
  "managed": false,
  "reusable": true
}`

const clusterListIsEmpty = `{
  "kind": "ClusterList",
  "page": 0,
  "size": 0,
  "total": 0,
  "items": [
  ]
}`
const clusterListIsNotEmpty = `{
  "kind": "ClusterList",
  "page": 1,
  "size": 1,
  "total": 1,
  "items": [
		{
			"name": "cluster-name",
***REMOVED***
  ]
}`

const getOidcConfigURL = "/api/clusters_mgmt/v1/oidc_configs/23f6gk51qi5ng15mm095c90hhajbf7c5"
const installerRoleARN = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
const unManagedIssuerURL = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
const managedIssuerURL = "https://d3gt1gce2zmg3d.cloudfront.net/23f6gk51qi5ng15mm095c90hhajbf7c5"
const managedOidcEndpointURL = "d3gt1gce2zmg3d.cloudfront.net/23f6gk51qi5ng15mm095c90hhajbf7c5"
const unManagedOidcEndpointURL = "oidc-f3y4.s3.us-east-1.amazonaws.com"
const secretARN = "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
const ID = "23f6gk51qi5ng15mm095c90hhajbf7c5"
const thumbrprint = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"

var _ = Describe("OIDC config creation", func(***REMOVED*** {
	It("Can create managed OIDC config", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/oidc_configs"***REMOVED***,
				VerifyJQ(`.managed`, true***REMOVED***,
				RespondWithJSON(http.StatusOK, managedOidcConfig***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, getOidcConfigURL***REMOVED***,
				RespondWithJSON(http.StatusOK, managedOidcConfig***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, getOidcConfigURL***REMOVED***,
				RespondWithJSON(http.StatusOK, managedOidcConfig***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				RespondWithJSON(http.StatusOK, clusterListIsEmpty***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodDelete, getOidcConfigURL***REMOVED***,
				RespondWithJSON(http.StatusNoContent, managedOidcConfig***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = true
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("rhcs_rosa_oidc_config", "oidc_config"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", ID***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.issuer_url", managedIssuerURL***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.managed", true***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.thumbprint", thumbrprint***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.oidc_endpoint_url", managedOidcEndpointURL***REMOVED******REMOVED***
		Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	Context("Create unmanaged OIDC config", func(***REMOVED*** {
		BeforeEach(func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/oidc_configs"***REMOVED***,
					VerifyJQ(`.managed`, false***REMOVED***,
					VerifyJQ(`.installer_role_arn`, installerRoleARN***REMOVED***,
					VerifyJQ(`.issuer_url`, unManagedIssuerURL***REMOVED***,
					VerifyJQ(`.secret_arn`, secretARN***REMOVED***,
					RespondWithJSON(http.StatusOK, unManagedOidcConfig***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, getOidcConfigURL***REMOVED***,
					RespondWithJSON(http.StatusOK, unManagedOidcConfig***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, getOidcConfigURL***REMOVED***,
					RespondWithJSON(http.StatusOK, unManagedOidcConfig***REMOVED***,
				***REMOVED***,
			***REMOVED***
***REMOVED******REMOVED***
		It("Succeed to destroy it", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
					RespondWithJSON(http.StatusOK, clusterListIsEmpty***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodDelete, getOidcConfigURL***REMOVED***,
					RespondWithJSON(http.StatusNoContent, unManagedOidcConfig***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = false
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
***REMOVED***
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			validateTerraformResourceState(***REMOVED***
			Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Fail on destroy due to a cluster that using it", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
					RespondWithJSON(http.StatusOK, clusterListIsNotEmpty***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodDelete, getOidcConfigURL***REMOVED***,
					RespondWithJSON(http.StatusNoContent, unManagedOidcConfig***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = false
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
***REMOVED***
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			validateTerraformResourceState(***REMOVED***

			// fail on destroy
			Expect(terraform.Destroy(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Fail on destroy because fail to get if there is a cluster that using it", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, clusterListIsNotEmpty***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodDelete, getOidcConfigURL***REMOVED***,
					RespondWithJSON(http.StatusNoContent, unManagedOidcConfig***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = false
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
***REMOVED***
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			validateTerraformResourceState(***REMOVED***

			// fail on destroy
			Expect(terraform.Destroy(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Fail on destroy because fail to remove the oidc config resource from OCM", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
					RespondWithJSON(http.StatusOK, clusterListIsEmpty***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodDelete, getOidcConfigURL***REMOVED***,
					RespondWithJSON(http.StatusInternalServerError, unManagedOidcConfig***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = false
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
***REMOVED***
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			validateTerraformResourceState(***REMOVED***

			// fail on destroy
			Expect(terraform.Destroy(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	It("Try to create managed OIDC config with unsupported attributes and fail", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/oidc_configs"***REMOVED***,
				VerifyJQ(`.managed`, true***REMOVED***,
				VerifyJQ(`.installer_role_arn`, installerRoleARN***REMOVED***,
				VerifyJQ(`.issuer_url`, unManagedIssuerURL***REMOVED***,
				VerifyJQ(`.secret_arn`, secretARN***REMOVED***,
				RespondWithJSON(http.StatusOK, unManagedOidcConfig***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, getOidcConfigURL***REMOVED***,
				RespondWithJSON(http.StatusOK, unManagedOidcConfig***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, getOidcConfigURL***REMOVED***,
				RespondWithJSON(http.StatusOK, unManagedOidcConfig***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				RespondWithJSON(http.StatusOK, clusterListIsEmpty***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodDelete, getOidcConfigURL***REMOVED***,
				RespondWithJSON(http.StatusNoContent, unManagedOidcConfig***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		resource "rhcs_rosa_oidc_config" "oidc_config" {
			  managed = true
			  secret_arn =  "arn:aws:secretsmanager:us-east-1:765374464689:secret:rosa-private-key-oidc-f3y4-fEqj4c"
			  issuer_url = "https://oidc-f3y4.s3.us-east-1.amazonaws.com"
			  installer_role_arn = "arn:aws:iam::765374464689:role/terr-account2-Installer-Role"
***REMOVED***
		`***REMOVED***
		// expect to fail
		Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

}***REMOVED***

func validateTerraformResourceState(***REMOVED*** {
	resource := terraform.Resource("rhcs_rosa_oidc_config", "oidc_config"***REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", ID***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(".attributes.installer_role_arn", installerRoleARN***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(".attributes.managed", false***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(".attributes.issuer_url", unManagedIssuerURL***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(".attributes.secret_arn", secretARN***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(".attributes.thumbprint", thumbrprint***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(".attributes.oidc_endpoint_url", unManagedOidcEndpointURL***REMOVED******REMOVED***

}
