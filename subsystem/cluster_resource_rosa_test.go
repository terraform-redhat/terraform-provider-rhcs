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
	"encoding/json"
***REMOVED***

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	"github.com/terraform-redhat/terraform-provider-ocm/build"
***REMOVED***

const versionListPage1 = `{
	"kind": "VersionList",
	"page": 1,
	"size": 2,
	"total": 2,
	"items": [{
			"kind": "Version",
			"id": "openshift-v4.10.1",
			"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.1",
			"raw_id": "4.10.1"
***REMOVED***,
		{
			"kind": "Version",
			"id": "openshift-v4.10.1",
			"href": "/api/clusters_mgmt/v1/versions/openshift-v4.11.1",
			"raw_id": "4.11.1"
***REMOVED***
	]
}`

var _ = Describe("ocm_cluster_rosa_classic - create", func(***REMOVED*** {
	// This is the cluster that will be returned by the server when asked to create or retrieve
	// a cluster.
	template := `{
	  "id": "123",
	  "name": "my-cluster",
	  "region": {
	    "id": "us-west-1"
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
	  "network": {
	    "machine_cidr": "10.0.0.0/16",
	    "service_cidr": "172.30.0.0/16",
	    "pod_cidr": "10.128.0.0/14",
	    "host_prefix": 23
	  },
	  "version": {
		  "id": "openshift-4.8.0"
	  }
	}`

	const templateReadyState = `{
	  "id": "123",
	  "name": "my-cluster",
	  "state": "ready",
	  "region": {
	    "id": "us-west-1"
	  },
	  "multi_az": true,
	  "api": {
	    "url": "https://my-api.example.com"
	  },
	  "console": {
	    "url": "https://my-console.example.com"
	  },
	  "network": {
	    "machine_cidr": "10.0.0.0/16",
	    "service_cidr": "172.30.0.0/16",
	    "pod_cidr": "10.128.0.0/14",
	    "host_prefix": 23
	  },
	  "version": {
		  "id": "openshift-4.8.0"
	  }
	}`

	Context("Test channel groups", func(***REMOVED*** {
		It("doesn't append the channel group when on the default channel", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
					RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
					VerifyJQ(`.version.id`, "openshift-v4.11.1"***REMOVED***,
					RespondWithPatchedJSON(http.StatusCreated, template, `[
						{
						  "op": "add",
						  "path": "/aws",
						  "value": {
							  "sts" : {
								  "oidc_endpoint_url": "https://oidc_endpoint_url",
								  "thumbprint": "111111",
								  "role_arn": "",
								  "support_role_arn": "",
								  "instance_iam_roles" : {
									"master_role_arn" : "",
									"worker_role_arn" : ""
								  },
								  "operator_role_prefix" : "test"
							  }
						  }
				***REMOVED***,
						{
						  "op": "add",
						  "path": "/nodes",
						  "value": {
							"compute": 3,
							"compute_machine_type": {
								"id": "r5.xlarge"
					***REMOVED***
						  }
				***REMOVED***,
						{
							"op": "replace",
						***REMOVED***: "/version/id",
							"value": "openshift-v4.11.1"
				***REMOVED***
						]`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			terraform.Source(`
			resource "ocm_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  master_role_arn = "",
					  worker_role_arn = "",
				  }
			  }
			  version = "openshift-v4.11.1"
	***REMOVED***
		  `***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("appends the channel group when on a non-default channel", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, versionListPage1, `[
						{
							"op": "add",
						***REMOVED***: "/items/-",
							"value": {
								"kind": "Version",
								"id": "openshift-v4.50.0-fast",
								"href": "/api/clusters_mgmt/v1/versions/openshift-v4.50.0-fast",
								"raw_id": "4.50.0",
								"channel_group": "fast"
					***REMOVED***
				***REMOVED***
					]`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
					VerifyJQ(`.version.id`, "openshift-v4.50.0-fast"***REMOVED***,
					VerifyJQ(`.version.channel_group`, "fast"***REMOVED***,
					RespondWithPatchedJSON(http.StatusCreated, template, `[
						{
						  "op": "add",
						  "path": "/aws",
						  "value": {
							  "sts" : {
								  "oidc_endpoint_url": "https://oidc_endpoint_url",
								  "thumbprint": "111111",
								  "role_arn": "",
								  "support_role_arn": "",
								  "instance_iam_roles" : {
									"master_role_arn" : "",
									"worker_role_arn" : ""
								  },
								  "operator_role_prefix" : "test"
							  }
						  }
				***REMOVED***,
						{
						  "op": "add",
						  "path": "/nodes",
						  "value": {
							"compute": 3,
							"compute_machine_type": {
								"id": "r5.xlarge"
					***REMOVED***
						  }
				***REMOVED***,
						{
							"op": "replace",
						***REMOVED***: "/version/id",
							"value": "openshift-v4.50.0-fast"
				***REMOVED***,
						{
							"op": "add",
						***REMOVED***: "/version/channel_group",
							"value": "fast"
				***REMOVED***
						]`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			terraform.Source(`
			resource "ocm_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  master_role_arn = "",
					  worker_role_arn = "",
				  }
			  }
			  channel_group = "fast"
			  version = "openshift-v4.50.0"
	***REMOVED***
		  `***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
		It("returns an error when the version is not found in the channel group", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, versionListPage1, `[
						{
							"op": "add",
						***REMOVED***: "/items/-",
							"value": {
								"kind": "Version",
								"id": "openshift-v4.50.0-fast",
								"href": "/api/clusters_mgmt/v1/versions/openshift-v4.50.0-fast",
								"raw_id": "4.50.0",
								"channel_group": "fast"
					***REMOVED***
				***REMOVED***
					]`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			terraform.Source(`
			resource "ocm_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
			  aws_account_id = "123"
			  sts = {
				  operator_role_prefix = "test"
				  role_arn = "",
				  support_role_arn = "",
				  instance_iam_roles = {
					  master_role_arn = "",
					  worker_role_arn = "",
				  }
			  }
			  channel_group = "fast"
			  version = "openshift-v4.99.99"
	***REMOVED***
		  `***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates basic cluster", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.properties.rosa_tf_version`, build.Version***REMOVED***,
				VerifyJQ(`.properties.rosa_tf_commit`, build.Commit***REMOVED***,
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("ocm_cluster_rosa_classic", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "openshift-4.8.0"***REMOVED******REMOVED***
	}***REMOVED***
	It("Creates basic cluster with properties", func(***REMOVED*** {
		prop_key := "my_prop_key"
		prop_val := "my_prop_val"
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.properties.rosa_tf_version`, build.Version***REMOVED***,
				VerifyJQ(`.properties.rosa_tf_commit`, build.Commit***REMOVED***,
				VerifyJQ(`.properties.`+prop_key, prop_val***REMOVED***,
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***,
                    {
                      "op": "add",
                      "path": "/properties",
                      "value": {
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            properties = { ` +
			prop_key + ` = "` + prop_val + `"` +
			`}
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
	It("Should fail cluster creation when trying to override reserved properties", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.properties.rosa_tf_version`, build.Version***REMOVED***,
				VerifyJQ(`.properties.rosa_tf_commit`, build.Commit***REMOVED***,
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			properties = {
   				rosa_tf_version = "bob"
	***REMOVED***
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	Context("Test destroy cluster", func(***REMOVED*** {
		BeforeEach(func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
					RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
					VerifyJQ(`.name`, "my-cluster"***REMOVED***,
					VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
					VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
					VerifyJQ(`.product.id`, "rosa"***REMOVED***,
					VerifyJQ(`.aws.sts.instance_iam_roles.master_role_arn`, ""***REMOVED***,
					VerifyJQ(`.aws.sts.instance_iam_roles.worker_role_arn`, ""***REMOVED***,
					VerifyJQ(`.aws.sts.operator_role_prefix`, "test"***REMOVED***,
					VerifyJQ(`.aws.sts.role_arn`, ""***REMOVED***,
					VerifyJQ(`.aws.sts.support_role_arn`, ""***REMOVED***,
					VerifyJQ(`.aws.account_id`, "123"***REMOVED***,
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, templateReadyState***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, templateReadyState***REMOVED***,
				***REMOVED***,
			***REMOVED***
***REMOVED******REMOVED***

		It("Disable waiting in destroy resource", func(***REMOVED*** {
			terraform.Source(`
				  resource "ocm_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					disable_waiting_in_destroy = true
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
				***REMOVED***
			***REMOVED***
				  }
			`***REMOVED***

			// it should return a warning so exit code will be "0":
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

***REMOVED******REMOVED***

		It("Wait in destroy resource but use the default timeout", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, template***REMOVED***,
				***REMOVED***,
			***REMOVED***
			terraform.Source(`
				  resource "ocm_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
				***REMOVED***
			***REMOVED***
				  }
			`***REMOVED***

			// it should return a warning so exit code will be "0":
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Wait in destroy resource and set timeout to a negative value", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, template***REMOVED***,
				***REMOVED***,
			***REMOVED***
			terraform.Source(`
				  resource "ocm_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					destroy_timeout = -1
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
				***REMOVED***
			***REMOVED***
				  }
			`***REMOVED***

			// it should return a warning so exit code will be "0":
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Wait in destroy resource and set timeout to a positive value", func(***REMOVED*** {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, template***REMOVED***,
				***REMOVED***,
			***REMOVED***
			terraform.Source(`
				  resource "ocm_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					destroy_timeout = 10
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
				***REMOVED***
			***REMOVED***
				  }
			`***REMOVED***

			// it should return a warning so exit code will be "0":
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			Expect(terraform.Destroy(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***

	It("Disable workload monitor and update it", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.disable_user_workload_monitoring`, true***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "disable_user_workload_monitoring" : true
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		terraform.Source(`
				  resource "ocm_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					disable_workload_monitoring = true
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
				***REMOVED***
			***REMOVED***
				  }
			`***REMOVED***

		// it should return a warning so exit code will be "0":
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// apply for update the workload monitor to be enabled
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "disable_user_workload_monitoring" : "true"
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "disable_user_workload_monitoring" : "false"
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		terraform.Source(`
				  resource "ocm_cluster_rosa_classic" "my_cluster" {
					name           = "my-cluster"
					cloud_region   = "us-west-1"
					aws_account_id = "123"
					disable_workload_monitoring = false
					sts = {
						operator_role_prefix = "test"
						role_arn = "",
						support_role_arn = "",
						instance_iam_roles = {
							master_role_arn = "",
							worker_role_arn = "",
				***REMOVED***
			***REMOVED***
				  }
			`***REMOVED***

		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("ocm_cluster_rosa_classic", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.disable_workload_monitoring`, false***REMOVED******REMOVED***

	}***REMOVED***

	It("Creates cluster with http proxy and update it", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.proxy.http_proxy`, "http://proxy.com"***REMOVED***,
				VerifyJQ(`.proxy.https_proxy`, "https://proxy.com"***REMOVED***,
				VerifyJQ(`.additional_trust_bundle`, "123"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/proxy",
					  "value": {
						  "http_proxy" : "http://proxy.com",
						  "https_proxy" : "https://proxy.com"
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "additional_trust_bundle" : "123"
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			proxy = {
				http_proxy = "http://proxy.com",
				https_proxy = "https://proxy.com",
				additional_trust_bundle = "123",
	***REMOVED***
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***

		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// apply for update the proxy's attributes
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/proxy",
					  "value": {
						  "http_proxy" : "http://proxy.com",
						  "https_proxy" : "https://proxy.com"
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "additional_trust_bundle" : "123"
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				VerifyJQ(`.proxy.https_proxy`, "https://proxy2.com"***REMOVED***,
				VerifyJQ(`.proxy.no_proxy`, "test"***REMOVED***,
				VerifyJQ(`.additional_trust_bundle`, "123"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/proxy",
					  "value": {
						  "https_proxy" : "https://proxy2.com",
						  "no_proxy" : "test"
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/",
					  "value": {
						  "additional_trust_bundle" : "123"
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// update the attribute "proxy"
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			proxy = {
				https_proxy = "https://proxy2.com",
				no_proxy = "test"
				additional_trust_bundle = "123",
	***REMOVED***
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
		resource := terraform.Resource("ocm_cluster_rosa_classic", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.proxy.https_proxy`, "https://proxy2.com"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.proxy.no_proxy`, "test"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.proxy.additional_trust_bundle`, "123"***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates cluster with default_mp_labels and update them", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.nodes.compute_labels.label_key1`, "label_value1"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
                        "compute_labels": {
                            "label_key1": "label_value1"
                        }
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            default_mp_labels = {
                label_key1 = "label_value1"
            }
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***

		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// apply for update the default_mp_labels
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
                        "compute_labels": {
                            "label_key1": "label_value1"
                        }
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
			// Update handler and response
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				VerifyJQ(`.nodes.compute_labels.changed_label`, "changed"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
                        "compute_labels": {
                            "changed_label": "changed"
                        }
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// update the attribute "proxy"
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            default_mp_labels = {
                changed_label = "changed"
            }
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED***, "Failed to update cluster with changed default_mp_labels"***REMOVED***
		resource := terraform.Resource("ocm_cluster_rosa_classic", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.default_mp_labels.changed_label`, "changed"***REMOVED******REMOVED***
	}***REMOVED***

	It("Except to fail on proxy validators", func(***REMOVED*** {
		// Expected at least one of the following: http-proxy, https-proxy
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			proxy = {
				no_proxy = "test1, test2"
				additional_trust_bundle = "123",
	***REMOVED***
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***

		// Expected at least one of the following: http-proxy, https-proxy, additional-trust-bundle
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			proxy = {
	***REMOVED***
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***

	}***REMOVED***
	It("Creates cluster with aws subnet ids & private link", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.aws.subnet_ids.[0]`, "id1"***REMOVED***,
				VerifyJQ(`.aws.private_link`, true***REMOVED***,
				VerifyJQ(`.nodes.availability_zones.[0]`, "az1"***REMOVED***,
				VerifyJQ(`.api.listening`, "internal"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": true,
						  "subnet_ids": ["id1", "id2", "id3"],
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
						"op": "add",
					***REMOVED***: "/availability_zones",
						"value": ["az1", "az2", "az3"]
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			availability_zones = ["az1","az2","az3"]
			aws_private_link = true
			aws_subnet_ids = [
				"id1", "id2", "id3"
			]
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates cluster when private link is false", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.aws.private_link`, false***REMOVED***,
				VerifyJQ(`.api.listening`, nil***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": false,
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			aws_private_link = false
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates rosa sts cluster with autoscaling and update the default machine pool", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.aws.sts.role_arn`, ""***REMOVED***,
				VerifyJQ(`.aws.sts.support_role_arn`, ""***REMOVED***,
				VerifyJQ(`.aws.sts.instance_iam_roles.master_role_arn`, ""***REMOVED***,
				VerifyJQ(`.aws.sts.instance_iam_roles.worker_role_arn`, ""***REMOVED***,
				VerifyJQ(`.aws.sts.operator_role_prefix`, "terraform-operator"***REMOVED***,
				VerifyJQ(`.nodes.autoscale_compute.kind`, "MachinePoolAutoscaling"***REMOVED***,
				VerifyJQ(`.nodes.autoscale_compute.max_replicas`, float64(4***REMOVED******REMOVED***,
				VerifyJQ(`.nodes.autoscale_compute.min_replicas`, float64(2***REMOVED******REMOVED***,
				VerifyJQ(`.nodes.compute_labels.label_key1`, "label_value1"***REMOVED***,
				VerifyJQ(`.nodes.compute_labels.label_key2`, "label_value2"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"autoscale_compute": {
							"min_replicas": 2,
							"max_replicas": 4
				***REMOVED***,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***,
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
				***REMOVED***
					  }
			***REMOVED***
				  ]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		resource "ocm_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			autoscaling_enabled = "true"
			min_replicas = "2"
			max_replicas = "4"
			default_mp_labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
	***REMOVED***
			sts = {
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
				  master_role_arn = "",
				  worker_role_arn = ""
		***REMOVED***,
				"operator_role_prefix" : "terraform-operator"
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// apply for update the min_replica from 2 to 3
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
							  },
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"autoscale_compute": {
							"min_replicas": 2,
							"max_replicas": 4
				***REMOVED***,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***,
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
				***REMOVED***
					  }
			***REMOVED***
				  ]`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				VerifyJQ(`.nodes.autoscale_compute.kind`, "MachinePoolAutoscaling"***REMOVED***,
				VerifyJQ(`.nodes.autoscale_compute.max_replicas`, float64(4***REMOVED******REMOVED***,
				VerifyJQ(`.nodes.autoscale_compute.min_replicas`, float64(3***REMOVED******REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
							  },
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"autoscale_compute": {
							"min_replicas": 3,
							"max_replicas": 4
				***REMOVED***,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***,
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
				***REMOVED***
					  }
			***REMOVED***
				  ]`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		// Run the apply command:
		terraform.Source(`
		resource "ocm_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			autoscaling_enabled = "true"
			min_replicas = "3"
			max_replicas = "4"
			default_mp_labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
	***REMOVED***
			sts = {
				role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
				support_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
				instance_iam_roles = {
				  master_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
				  worker_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
		***REMOVED***,
				"operator_role_prefix" : "terraform-operator"
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// apply for update the autoscaling group to compute nodes
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
							  },
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"autoscale_compute": {
							"min_replicas": 3,
							"max_replicas": 4
				***REMOVED***,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***,
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
				***REMOVED***
					  }
			***REMOVED***
				  ]`***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				VerifyJQ(`.nodes.compute`, float64(4***REMOVED******REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
							  },
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 4,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***,
						"compute_labels": {
							"label_key1": "label_value1",
				    		"label_key2": "label_value2"
				***REMOVED***
					  }
			***REMOVED***
				  ]`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		// Run the apply command:
		terraform.Source(`
		resource "ocm_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			replicas = 4
			default_mp_labels = {
				"label_key1" = "label_value1",
				"label_key2" = "label_value2"
	***REMOVED***
			sts = {
				role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
				support_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
				instance_iam_roles = {
				  master_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
				  worker_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
		***REMOVED***,
				"operator_role_prefix" : "terraform-operator"
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates rosa sts cluster with OIDC Configuration ID", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.aws.sts.role_arn`, ""***REMOVED***,
				VerifyJQ(`.aws.sts.support_role_arn`, ""***REMOVED***,
				VerifyJQ(`.aws.sts.instance_iam_roles.master_role_arn`, ""***REMOVED***,
				VerifyJQ(`.aws.sts.instance_iam_roles.worker_role_arn`, ""***REMOVED***,
				VerifyJQ(`.aws.sts.operator_role_prefix`, "terraform-operator"***REMOVED***,
				VerifyJQ(`.aws.sts.oidc_config.id`, "aaa"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "oidc_config": {
								"id": "aaa",
								"secret_arn": "aaa",
								"issuer_url": "https://oidc_endpoint_url",
								"reusable": true,
								"managed": false
							  },
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "terraform-operator"
						  }
					  }
			***REMOVED***,
					{
						"op": "add",
					***REMOVED***: "/nodes",
						"value": {
						  "compute": 3,
						  "compute_machine_type": {
							  "id": "r5.xlarge"
						  }
				***REMOVED***
					  }
				  ]`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		// Run the apply command:
		terraform.Source(`
		resource "ocm_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
				  master_role_arn = "",
				  worker_role_arn = ""
		***REMOVED***,
				"operator_role_prefix" : "terraform-operator",
				"oidc_config_id" = "aaa"
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Fails to create cluster with incompatible account role's version and fail", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "arn:aws:iam::765374464689:role/terr-account-Installer-Role",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			version = "openshift-v4.12"
			sts = {
				operator_role_prefix = "test"
				role_arn = "arn:aws:iam::765374464689:role/terr-account-Installer-Role",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		// expect to get an error
		Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Create cluster with http token", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.aws.ec2_metadata_http_tokens`, "required"***REMOVED***,
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens" : "required",
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			ec2_metadata_http_tokens = "required"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Fails to create cluster with http tokens and not supported version", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.aws.ec2_metadata_http_tokens`, "required"***REMOVED***,
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens" : "required",
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			ec2_metadata_http_tokens = "required"
			version = "openshift-v4.10"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		// expect to get an error
		Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Fails to create cluster with http tokens with not supported value", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.aws.http_tokens_state`, "bad_string"***REMOVED***,
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens" : "bad_string",
						  "sts" : {
							  "oidc_endpoint_url": "https://oidc_endpoint_url",
							  "thumbprint": "111111",
							  "role_arn": "",
							  "support_role_arn": "",
							  "instance_iam_roles" : {
								"master_role_arn" : "",
								"worker_role_arn" : ""
							  },
							  "operator_role_prefix" : "test"
						  }
					  }
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			ec2_metadata_http_tokens = "bad_string"
			version = "openshift-v4.12"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
		  }
		`***REMOVED***
		// expect to get an error
		Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("ocm_cluster_rosa_classic - upgrade", func(***REMOVED*** {
	const template = `{
		"id": "123",
		"name": "my-cluster",
		"region": {
		  "id": "us-west-1"
***REMOVED***,
		"aws": {
			"sts": {
				"oidc_endpoint_url": "https://127.0.0.2",
				"thumbprint": "111111",
				"role_arn": "",
				"support_role_arn": "",
				"instance_iam_roles" : {
					"master_role_arn" : "",
					"worker_role_arn" : ""
		***REMOVED***,
				"operator_role_prefix" : "test"
	***REMOVED***
***REMOVED***,
		"multi_az": true,
		"api": {
		  "url": "https://my-api.example.com"
***REMOVED***,
		"console": {
		  "url": "https://my-console.example.com"
***REMOVED***,
		"network": {
		  "machine_cidr": "10.0.0.0/16",
		  "service_cidr": "172.30.0.0/16",
		  "pod_cidr": "10.128.0.0/14",
		  "host_prefix": 23
***REMOVED***,
		"nodes": {
			"compute": 3,
			"compute_machine_type": {
				"id": "r5.xlarge"
	***REMOVED***
***REMOVED***,
		"version": {
			"id": "openshift-v4.8.0"
***REMOVED***
	}`
	const v4_8_0Info = `{
		"kind": "Version",
		"id": "openshift-v4.8.0",
		"href": "/api/clusters_mgmt/v1/versions/openshift-v4.8.0",
		"raw_id": "4.8.0",
		"enabled": true,
		"default": false,
		"channel_group": "stable",
		"available_upgrades": [
			"4.10.1"
		],
		"rosa_enabled": true
	}`
	const v4_10_1Info = `{
		"kind": "Version",
		"id": "openshift-v4.10.1",
		"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.1",
		"raw_id": "4.10.1",
		"enabled": true,
		"default": false,
		"channel_group": "stable",
		"available_upgrades": [],
		"rosa_enabled": true
	}`
	const operIAMList = `{
		"kind": "OperatorIAMRoleList",
		"href": "/api/clusters_mgmt/v1/123/sts_operator_roles",
		"page": 1,
		"size": 6,
		"total": 6,
		"items": [
		  {
			"id": "",
			"name": "ebs-cloud-credentials",
			"role_arn": ""
		  },
		  {
			"id": "",
			"role_arn": ""
		  },
		  {
			"id": "",
			"name": "aws-cloud-credentials",
			"role_arn": ""
		  },
		  {
			"id": "",
			"name": "cloud-credential-operator-iam-ro-creds",
			"role_arn": ""
		  },
		  {
			"id": "",
			"name": "installer-cloud-credentials",
			"role_arn": ""
		  },
		  {
			"id": "",
			"name": "cloud-credentials",
			"role_arn": ""
		  }
		]
	}`
	const upgradePoliciesEmpty = `{
		"kind": "UpgradePolicyList",
		"page": 1,
		"size": 0,
		"total": 0,
		"items": []
	}`
	BeforeEach(func(***REMOVED*** {
		Expect(json.Valid([]byte(template***REMOVED******REMOVED******REMOVED***.To(BeTrue(***REMOVED******REMOVED***
		Expect(json.Valid([]byte(v4_8_0Info***REMOVED******REMOVED******REMOVED***.To(BeTrue(***REMOVED******REMOVED***

		// Create a cluster for us to upgrade:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				RespondWithJSON(http.StatusCreated, template***REMOVED***,
			***REMOVED***,
		***REMOVED***
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***,
				"operator_iam_roles": [
					{
					  "id": "",
					  "name": "ebs-cloud-credentials",
					  "namespace": "openshift-cluster-csi-drivers",
					  "role_arn": "",
					  "service_account": ""
			***REMOVED***,
					{
					  "id": "",
					  "name": "cloud-credentials",
					  "namespace": "openshift-cloud-network-config-controller",
					  "role_arn": "",
					  "service_account": ""
			***REMOVED***,
					{
					  "id": "",
					  "name": "aws-cloud-credentials",
					  "namespace": "openshift-machine-api",
					  "role_arn": "",
					  "service_account": ""
			***REMOVED***,
					{
					  "id": "",
					  "name": "cloud-credential-operator-iam-ro-creds",
					  "namespace": "openshift-cloud-credential-operator",
					  "role_arn": "",
					  "service_account": ""
			***REMOVED***,
					{
					  "id": "",
					  "name": "installer-cloud-credentials",
					  "namespace": "openshift-image-registry",
					  "role_arn": "",
					  "service_account": ""
			***REMOVED***,
					{
					  "id": "",
					  "name": "cloud-credentials",
					  "namespace": "openshift-ingress-operator",
					  "role_arn": "",
					  "service_account": ""
			***REMOVED***
				]
	***REMOVED***
***REMOVED***
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Verify initial cluster version
		resource := terraform.Resource("ocm_cluster_rosa_classic", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "openshift-v4.8.0"***REMOVED******REMOVED***
	}***REMOVED***

	It("Upgrades cluster", func(***REMOVED*** {
		server.AppendHandlers(
			// Refresh cluster state
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, template***REMOVED***,
			***REMOVED***,
			// Validate upgrade versions
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.8.0"***REMOVED***,
				RespondWithJSON(http.StatusOK, v4_8_0Info***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"***REMOVED***,
				RespondWithJSON(http.StatusOK, v4_10_1Info***REMOVED***,
			***REMOVED***,
			// Validate roles
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/sts_operator_roles"***REMOVED***,
				RespondWithJSON(http.StatusOK, operIAMList***REMOVED***,
			***REMOVED***,
			// Look for existing upgrade policies
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"***REMOVED***,
				RespondWithJSON(http.StatusOK, upgradePoliciesEmpty***REMOVED***,
			***REMOVED***,
			// Create an upgrade policy
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"***REMOVED***,
				VerifyJQ(".version", "4.10.1"***REMOVED***,
				RespondWithJSON(http.StatusCreated, `
				{
					"kind": "UpgradePolicy",
					"id": "123",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
					"schedule_type": "manual",
					"upgrade_type": "OSD",
					"version": "4.10.1",
					"next_run": "2023-06-09T20:59:00Z",
					"cluster_id": "123",
					"enable_minor_version_upgrades": true
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Patch the cluster (w/ no changes***REMOVED***
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, template***REMOVED***,
			***REMOVED***,
		***REMOVED***
		// Perform upgrade
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
			version = "openshift-v4.10.1"
***REMOVED***`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Does nothing if upgrade is in progress", func(***REMOVED*** {
		server.AppendHandlers(
			// Refresh cluster state
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, template***REMOVED***,
			***REMOVED***,
			// Validate upgrade versions
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.8.0"***REMOVED***,
				RespondWithJSON(http.StatusOK, v4_8_0Info***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"***REMOVED***,
				RespondWithJSON(http.StatusOK, v4_10_1Info***REMOVED***,
			***REMOVED***,
			// Validate roles
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/sts_operator_roles"***REMOVED***,
				RespondWithJSON(http.StatusOK, operIAMList***REMOVED***,
			***REMOVED***,
			// Look for existing upgrade policies
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"page": 1,
					"size": 0,
					"total": 0,
					"items": [
						{
							"kind": "UpgradePolicy",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
							"schedule_type": "manual",
							"upgrade_type": "OSD",
							"version": "4.10.0",
							"next_run": "2023-06-09T20:59:00Z",
							"cluster_id": "123",
							"enable_minor_version_upgrades": true
				***REMOVED***
					]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Check it's state
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"id": "456",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state",
					"description": "Upgrade in progress",
					"value": "started"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		// Perform try the upgrade
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
			version = "openshift-v4.10.1"
***REMOVED***`***REMOVED***
		// Will fail due to upgrade in progress
		Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Cancels and upgrade for the wrong version & schedules new", func(***REMOVED*** {
		server.AppendHandlers(
			// Refresh cluster state
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, template***REMOVED***,
			***REMOVED***,
			// Validate upgrade versions
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.8.0"***REMOVED***,
				RespondWithJSON(http.StatusOK, v4_8_0Info***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"***REMOVED***,
				RespondWithJSON(http.StatusOK, v4_10_1Info***REMOVED***,
			***REMOVED***,
			// Validate roles
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/sts_operator_roles"***REMOVED***,
				RespondWithJSON(http.StatusOK, operIAMList***REMOVED***,
			***REMOVED***,
			// Look for existing upgrade policies
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"page": 1,
					"size": 0,
					"total": 0,
					"items": [
						{
							"kind": "UpgradePolicy",
							"id": "456",
							"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
							"schedule_type": "manual",
							"upgrade_type": "OSD",
							"version": "4.10.0",
							"next_run": "2023-06-09T20:59:00Z",
							"cluster_id": "123",
							"enable_minor_version_upgrades": true
				***REMOVED***
					]
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Check it's state
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state"***REMOVED***,
				RespondWithJSON(http.StatusOK, `{
					"kind": "UpgradePolicyState",
					"id": "456",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456/state",
					"description": "",
					"value": "scheduled"
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Delete existing upgrade policy
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/456"***REMOVED***,
				RespondWithJSON(http.StatusOK, "{}"***REMOVED***,
			***REMOVED***,
			// Create an upgrade policy
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"***REMOVED***,
				VerifyJQ(".version", "4.10.1"***REMOVED***,
				RespondWithJSON(http.StatusCreated, `
				{
					"kind": "UpgradePolicy",
					"id": "123",
					"href": "/api/clusters_mgmt/v1/clusters/123/upgrade_policies/123",
					"schedule_type": "manual",
					"upgrade_type": "OSD",
					"version": "4.10.1",
					"next_run": "2023-06-09T20:59:00Z",
					"cluster_id": "123",
					"enable_minor_version_upgrades": true
		***REMOVED***`***REMOVED***,
			***REMOVED***,
			// Patch the cluster (w/ no changes***REMOVED***
			CombineHandlers(
				VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, template***REMOVED***,
			***REMOVED***,
		***REMOVED***
		// Perform try the upgrade
		terraform.Source(`
		  resource "ocm_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = "",
				support_role_arn = "",
				instance_iam_roles = {
					master_role_arn = "",
					worker_role_arn = "",
		***REMOVED***
	***REMOVED***
			version = "openshift-v4.10.1"
***REMOVED***`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
