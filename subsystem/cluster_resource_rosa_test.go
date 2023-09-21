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
	"github.com/terraform-redhat/terraform-provider-rhcs/build"
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

var _ = Describe("rhcs_cluster_rosa_classic - create", func(***REMOVED*** {
	// This is the cluster that will be returned by the server when asked to create or retrieve
	// a cluster.
	template := `{
	  "id": "123",
	  "name": "my-cluster",
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
	  },
      "nodes": {
	    "compute": 3,
        "availability_zones": ["us-west-1a"],
	    "compute_machine_type": {
	      "id": "r5.xlarge"
	    }
	  },
      "dns" : {
          "base_domain": "mycluster-api.example.com"
      }
	}`

	Context("rhcs_cluster_rosa_classic - create", func(***REMOVED*** {
		It("invalid az for region", func(***REMOVED*** {
			terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("version with unsupported prefix error", func(***REMOVED*** {
			terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
			  name           = "my-cluster"
			  cloud_region   = "us-west-1"
              availability_zones = ["us-east-1a"]
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
			  version = "4.11.1"
	***REMOVED***
		  `***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

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
							  "ec2_metadata_http_tokens": "optional",
							  "sts" : {
								  "oidc_endpoint_url": "https://127.0.0.2",
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
							"op": "replace",
						***REMOVED***: "/version/id",
							"value": "openshift-v4.11.1"
				***REMOVED***
						]`***REMOVED***,
					***REMOVED***,
				***REMOVED***
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			  version = "4.11.1"
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
                              "ec2_metadata_http_tokens": "optional",
							  "sts" : {
								  "oidc_endpoint_url": "https://127.0.0.2",
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
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			  version = "4.50.0"
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
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			  version = "4.99.99"
	***REMOVED***
		  `***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***

		Context("Test wait attribute", func(***REMOVED*** {
			It("Create cluster and wait till it will be in error state", func(***REMOVED*** {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
						RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
					***REMOVED***,
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
						RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
						RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/state",
					  "value": "error"
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
				***REMOVED***

				// Run the apply command:
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			wait_for_create_complete = true
		  }
		`***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
				resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
				Expect(resource***REMOVED***.To(MatchJQ(".attributes.state", "error"***REMOVED******REMOVED***
	***REMOVED******REMOVED***

			It("Create cluster and wait till it will be in ready state", func(***REMOVED*** {
				// Prepare the server:
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
						RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
					***REMOVED***,
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
						RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
						RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/state",
					  "value": "ready"
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
				***REMOVED***

				// Run the apply command:
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			wait_for_create_complete = true
		  }
		`***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
				resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
				Expect(resource***REMOVED***.To(MatchJQ(".attributes.state", "ready"***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "openshift-4.8.0"***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Creates basic cluster returned empty az list", func(***REMOVED*** {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
						  "compute": 3,
	            "compute_machine_type": {
	              "id": "r5.xlarge"
	            }
					  }
			***REMOVED***
					]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "openshift-4.8.0"***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Creates basic cluster with admin user", func(***REMOVED*** {
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
					VerifyJQ(`.htpasswd.users.items[0].username`, "cluster_admin"***REMOVED***,
					VerifyJQ(`.htpasswd.users.items[0].password`, "1234AbB2341234"***REMOVED***,
					VerifyJQ(`.properties.rosa_tf_version`, build.Version***REMOVED***,
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit***REMOVED***,
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            admin_credentials = {
                username = "cluster_admin"
                password = "1234AbB2341234"
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
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "openshift-4.8.0"***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Creates basic cluster - and reconcile on a 404", func(***REMOVED*** {
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
                          "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "openshift-4.8.0"***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "123"***REMOVED******REMOVED*** // cluster has id 123

			// Prepare the server for reconcile
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusNotFound, "{}"***REMOVED***,
				***REMOVED***,
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
                      "op": "replace",
                      "path": "/id",
                      "value": "1234"
                    },
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			resource = terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "openshift-4.8.0"***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.id", "1234"***REMOVED******REMOVED*** // reconciled cluster has id of 1234
***REMOVED******REMOVED***

		It("Creates basic cluster with custom worker disk size", func(***REMOVED*** {
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
					VerifyJQ(`.nodes.compute_root_volume.aws.size`, 400.0***REMOVED***,
					RespondWithPatchedJSON(http.StatusCreated, template, `[
						{
						  "op": "add",
						  "path": "/aws",
						  "value": {
							  "ec2_metadata_http_tokens": "optional",
							  "sts" : {
								  "oidc_endpoint_url": "https://127.0.0.2",
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
							"availability_zones": ["az"],
							"compute_machine_type": {
								"id": "r5.xlarge"
					***REMOVED***,
							"compute_root_volume": {
								"aws": {
									"size": 400
						***REMOVED***
					***REMOVED***
						  }
				***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
			  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
				worker_disk_size = 400
			  }
			`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "openshift-4.8.0"***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.worker_disk_size", 400.0***REMOVED******REMOVED***
***REMOVED******REMOVED***

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
                          "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/properties",
                      "value": {
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

		It("Creates basic cluster with properties and update them", func(***REMOVED*** {
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
                          "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.properties.`+prop_key, prop_val***REMOVED******REMOVED***

			// Prepare server for update
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
                          "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					VerifyJQ(`.properties.rosa_tf_version`, build.Version***REMOVED***,
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit***REMOVED***,
					VerifyJQ(`.properties.`+prop_key, prop_val+"_1"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`_1"
                      }
                    }]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            properties = { ` +
				prop_key + ` = "` + prop_val + `_1"` +
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
			resource = terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val+"_1"***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.properties.`+prop_key, prop_val+"_1"***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Creates basic cluster with properties and delete them", func(***REMOVED*** {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.`+prop_key, prop_val***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.properties.`+prop_key, prop_val***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.properties| keys | length`, 1***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties| keys | length`, 3***REMOVED******REMOVED***

			// Prepare server for update
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`",
                        "`+prop_key+`": "`+prop_val+`"
                      }
                    }]`***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					VerifyJQ(`.properties.rosa_tf_version`, build.Version***REMOVED***,
					VerifyJQ(`.properties.rosa_tf_commit`, build.Commit***REMOVED***,
					VerifyJQ(`.properties.`+prop_key, nil***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                      "path": "/properties",
                      "value": {
                        "rosa_tf_commit":"`+build.Commit+`",
                        "rosa_tf_version":"`+build.Version+`"
                      }
                    }]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
            properties = {}
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
			resource = terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.rosa_tf_commit`, build.Commit***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties.rosa_tf_version`, build.Version***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.properties | keys | length`, 0***REMOVED******REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.ocm_properties | keys | length`, 2***REMOVED******REMOVED***
***REMOVED******REMOVED***

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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

		It("Should fail cluster creation when cluster name length is more than 15", func(***REMOVED*** {
			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster-234567"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			properties = {
   				cluster_name = "too_long"
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

***REMOVED******REMOVED***

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
					      "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

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
					      "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
                        "availability_zones": ["us-west-1a"],
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			terraform.Source(`
				  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.disable_workload_monitoring`, false***REMOVED******REMOVED***

***REMOVED******REMOVED***

		Context("Test Proxy", func(***REMOVED*** {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
				***REMOVED***

				// Run the apply command:
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						  "additional_trust_bundle" : "REDUCTED"
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
                          "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						  "additional_trust_bundle" : "REDUCTED"
					  }
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
				***REMOVED***

				// update the attribute "proxy"
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
				resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
				Expect(resource***REMOVED***.To(MatchJQ(`.attributes.proxy.https_proxy`, "https://proxy2.com"***REMOVED******REMOVED***
				Expect(resource***REMOVED***.To(MatchJQ(`.attributes.proxy.no_proxy`, "test"***REMOVED******REMOVED***
				Expect(resource***REMOVED***.To(MatchJQ(`.attributes.proxy.additional_trust_bundle`, "123"***REMOVED******REMOVED***
	***REMOVED******REMOVED***
			It("Creates cluster without http proxy and update trust bundle - should fail", func(***REMOVED*** {
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
						RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
				***REMOVED***

				// Run the apply command:
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
						VerifyJQ(`.additional_trust_bundle`, "123"***REMOVED***,
						RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						  "additional_trust_bundle" : "123"
					  }
			***REMOVED***]`***REMOVED***,
					***REMOVED***,
				***REMOVED***
				// update the attribute "proxy"
				terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			proxy = {
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
				Expect(terraform.Apply(***REMOVED******REMOVED***.ToNot(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
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
                          "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
                        "compute_labels": {
                            "label_key1": "label_value1"
                        },
						"availability_zones": [
      						"us-west-1a"
    					],
						"compute_machine_type": {
						   "id": "r5.xlarge"
	    		***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
                        "compute_labels": {
                            "changed_label": "changed"
                        },
						"availability_zones": [
      						"us-west-1a"
    					],
						"compute_machine_type": {
						   "id": "r5.xlarge"
	    		***REMOVED***
					  }
			***REMOVED***
					]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// update the attribute "proxy"
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(`.attributes.default_mp_labels.changed_label`, "changed"***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Except to fail on proxy validators", func(***REMOVED*** {
			// Expected at least one of the following: http-proxy, https-proxy
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			 resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***
		It("Creates private cluster with aws subnet ids without private link", func(***REMOVED*** {
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
					VerifyJQ(`.aws.private_link`, false***REMOVED***,
					VerifyJQ(`.nodes.availability_zones.[0]`, "us-west-1a"***REMOVED***,
					VerifyJQ(`.api.listening`, "internal"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": false,
						  "subnet_ids": ["id1", "id2", "id3"],
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
					  "path": "/api",
					  "value": {
					  	"listening": "internal"
					  }
			***REMOVED***,
					{
						"op": "add",
					***REMOVED***: "/availability_zones",
						"value": ["us-west-1a"]
			***REMOVED***,
					{
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
						"availability_zones": [
      						"us-west-1a"
    					],
						"compute_machine_type": {
						   "id": "r5.xlarge"
	    		***REMOVED***
					  }
			***REMOVED***
					]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			availability_zones = ["us-west-1a"]
			aws_private_link = false
			private = true
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
***REMOVED******REMOVED***
		It("Creates private cluster with aws subnet ids & private link", func(***REMOVED*** {
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
					VerifyJQ(`.nodes.availability_zones.[0]`, "us-west-1a"***REMOVED***,
					VerifyJQ(`.api.listening`, "internal"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": true,
						  "subnet_ids": ["id1", "id2", "id3"],
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
					  "path": "/api",
					  "value": {
					  	"listening": "internal"
					  }
			***REMOVED***,
					{
					  "op": "replace",
					  "path": "/nodes",
					  "value": {
						"availability_zones": [
      						"us-west-1a"
    					],
						"compute_machine_type": {
						   "id": "r5.xlarge"
	    		***REMOVED***
					  }
			***REMOVED***
					]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			availability_zones = ["us-west-1a"]
			private = true
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
***REMOVED******REMOVED***

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
					VerifyJQ(`.api.listening`, "external"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": false,
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

		It("Creates cluster with shared VPC", func(***REMOVED*** {
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
					VerifyJQ(`.dns.base_domain`, "mydomain.openshift.dev"***REMOVED***,
					VerifyJQ(`.aws.subnet_ids.[0]`, "id1"***REMOVED***,
					VerifyJQ(`.aws.private_hosted_zone_id`, "1234"***REMOVED***,
					VerifyJQ(`.aws.private_hosted_zone_role_arn`, "arn:aws:iam::111111111111:role/test-shared-vpc"***REMOVED***,
					VerifyJQ(`.nodes.availability_zones.[0]`, "us-west-1a"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "subnet_ids": ["id1", "id2", "id3"],
						  "ec2_metadata_http_tokens": "optional",
                          "private_hosted_zone_id": "1234",
                          "private_hosted_zone_role_arn": "arn:aws:iam::111111111111:role/test-shared-vpc",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
					***REMOVED***: "/dns",
                        "value": {"base_domain": "mydomain.openshift.dev"}
			***REMOVED***,
					{
						"op": "add",
					***REMOVED***: "/availability_zones",
						"value": ["us-west-1a"]
			***REMOVED***,
					{
					  "op": "add",
					  "path": "/nodes",
					  "value": {
						"compute": 3,
            "availability_zones": ["us-west-1a"],
						"compute_machine_type": {
							"id": "r5.xlarge"
				***REMOVED***
					  }
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			availability_zones = ["us-west-1a"]
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
            private_hosted_zone = {
                id = "1234"
                role_arn = "arn:aws:iam::111111111111:role/test-shared-vpc"
            }
            base_dns_domain = "mydomain.openshift.dev"
		  }
		`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				***REMOVED***,
						"availability_zones": [
							"az"
						]
					  }
			***REMOVED***
				  ]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
				***REMOVED***,
                        "availability_zones": [
							"az"
						]
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						"availability_zones": ["az"],
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
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						"availability_zones": ["az"],
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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
						"availability_zones": ["az"],
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
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
							  "oidc_config": {
								"id": "aaa",
								"secret_arn": "aaa",
								"issuer_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			// Run the apply command:
			terraform.Source(`
		resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

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
						  "ec2_metadata_http_tokens": "optional",
						  "sts" : {
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

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
							  "oidc_endpoint_url": "https://127.0.0.2",
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
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***

			// Run the apply command:
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
		    name           = "my-cluster"
		    cloud_region   = "us-west-1"
			aws_account_id = "123"
			ec2_metadata_http_tokens = "bad_string"
			version = "4.12"
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
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("rhcs_cluster_rosa_classic - upgrade", func(***REMOVED*** {
	const template = `{
		"id": "123",
		"name": "my-cluster",
		"state": "ready",
		"region": {
		  "id": "us-west-1"
***REMOVED***,
		"aws": {
			"ec2_metadata_http_tokens": "optional",
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
	        "availability_zones": ["az"],
			"compute_machine_type": {
				"id": "r5.xlarge"
	***REMOVED***
***REMOVED***,
		"version": {
			"id": "4.10.0"
***REMOVED***
	}`
	const versionList = `{
		"kind": "VersionList",
		"page": 1,
		"size": 3,
		"total": 3,
		"items": [{
				"kind": "Version",
				"id": "openshift-v4.10.0",
				"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.0",
				"raw_id": "4.10.0"
	***REMOVED***,
			{
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
	const v4_10_0Info = `{
		"kind": "Version",
		"id": "openshift-v4.10.0",
		"href": "/api/clusters_mgmt/v1/versions/openshift-v4.10.0",
		"raw_id": "4.10.0",
		"enabled": true,
		"default": false,
		"channel_group": "stable",
		"available_upgrades": ["4.10.1"],
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
	const upgradePoliciesEmpty = `{
		"kind": "UpgradePolicyList",
		"page": 1,
		"size": 0,
		"total": 0,
		"items": []
	}`
	BeforeEach(func(***REMOVED*** {
		Expect(json.Valid([]byte(template***REMOVED******REMOVED******REMOVED***.To(BeTrue(***REMOVED******REMOVED***
		Expect(json.Valid([]byte(v4_10_0Info***REMOVED******REMOVED******REMOVED***.To(BeTrue(***REMOVED******REMOVED***

		// Create a cluster for us to upgrade:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				RespondWithJSON(http.StatusOK, versionList***REMOVED***,
			***REMOVED***,
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/properties",
					  "value": {
						"rosa_tf_commit": "123",
						"rosa_tf_version": "123"
					  }
			***REMOVED***
				]`***REMOVED***,
			***REMOVED***,
		***REMOVED***
		terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			version = "4.10.0"
***REMOVED***
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Verify initial cluster version
		resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "4.10.0"***REMOVED******REMOVED***
	}***REMOVED***
	Context("rhcs_cluster_rosa_classic - create", func(***REMOVED*** {
		It("Upgrades cluster", func(***REMOVED*** {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, template***REMOVED***,
				***REMOVED***,
				// Validate upgrade versions
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"***REMOVED***,
					RespondWithJSON(http.StatusOK, v4_10_0Info***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"***REMOVED***,
					RespondWithJSON(http.StatusOK, v4_10_1Info***REMOVED***,
				***REMOVED***,
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"***REMOVED***,
					RespondWithJSON(http.StatusOK, upgradePoliciesEmpty***REMOVED***,
				***REMOVED***,
				// Look for gate agreements by posting an upgrade policy w/ dryRun
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies", "dryRun=true"***REMOVED***,
					VerifyJQ(".version", "4.10.1"***REMOVED***,
					RespondWithJSON(http.StatusBadRequest, `{
					"kind": "Error",
					"id": "400",
					"href": "/api/clusters_mgmt/v1/errors/400",
					"code": "CLUSTERS-MGMT-400",
					"reason": "There are missing version gate agreements for this cluster. See details.",
					"details": [
					  {
						"kind": "VersionGate",
						"id": "999",
						"href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
						"version_raw_id_prefix": "4.10",
						"label": "api.openshift.com/gate-sts",
						"value": "4.10",
						"warning_message": "STS roles must be updated blah blah blah",
						"description": "OpenShift STS clusters include new required cloud provider permissions in OpenShift 4.YY.",
						"documentation_url": "https://access.redhat.com/solutions/0000000",
						"sts_only": true,
						"creation_timestamp": "2023-04-03T06:39:57.057613Z"
					  }
					],
					"operation_id": "8f2d2946-c4ef-4c2f-877b-c19eb17dc918"
				  }`***REMOVED***,
				***REMOVED***,
				// Send acks for all gate agreements
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/gate_agreements"***REMOVED***,
					VerifyJQ(".version_gate.id", "999"***REMOVED***,
					RespondWithJSON(http.StatusCreated, `{
					"kind": "VersionGateAgreement",
					"id": "888",
					"href": "/api/clusters_mgmt/v1/clusters/24g9q8jhdhv66fi41jfiuup5lsvu61fi/gate_agreements/d2e8d371-1033-11ee-9f05-0a580a820bdb",
					"version_gate": {
					  "kind": "VersionGate",
					  "id": "999",
					  "href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
					  "version_raw_id_prefix": "4.10",
					  "label": "api.openshift.com/gate-sts",
					  "value": "4.10",
					  "warning_message": "STS blah blah blah",
					  "description": "OpenShift STS clusters include new required cloud provider permissions in OpenShift 4.YY.",
					  "documentation_url": "https://access.redhat.com/solutions/0000000",
					  "sts_only": true,
					  "creation_timestamp": "2023-04-03T06:39:57.057613Z"
			***REMOVED***,
					"creation_timestamp": "2023-06-21T13:02:06.291443Z"
				  }`***REMOVED***,
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
					RespondWithPatchedJSON(http.StatusCreated, template, `[
						{
						  "op": "add",
						  "path": "/properties",
						  "value": {
							"rosa_tf_commit": "123",
							"rosa_tf_version": "123"
						  }
				***REMOVED***
					]`,
					***REMOVED******REMOVED***,
			***REMOVED***
			// Perform upgrade w/ auto-ack of sts-only gate agreements
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
			name           = "my-cluster"
			cloud_region   = "us-west-1"
			aws_account_id = "123"
			sts = {
				operator_role_prefix = "test"
				role_arn = ""
				support_role_arn = ""
				instance_iam_roles = {
					master_role_arn = ""
					worker_role_arn = ""
		***REMOVED***
	***REMOVED***
			version = "4.10.1"
***REMOVED***`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Upgrades cluster support old version format", func(***REMOVED*** {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, template***REMOVED***,
				***REMOVED***,
				// Validate upgrade versions
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"***REMOVED***,
					RespondWithJSON(http.StatusOK, v4_10_0Info***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"***REMOVED***,
					RespondWithJSON(http.StatusOK, v4_10_1Info***REMOVED***,
				***REMOVED***,
				// Look for existing upgrade policies
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"***REMOVED***,
					RespondWithJSON(http.StatusOK, upgradePoliciesEmpty***REMOVED***,
				***REMOVED***,
				// Look for gate agreements by posting an upgrade policy w/ dryRun
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies", "dryRun=true"***REMOVED***,
					VerifyJQ(".version", "4.10.1"***REMOVED***,
					RespondWithJSON(http.StatusBadRequest, `{
					"kind": "Error",
					"id": "400",
					"href": "/api/clusters_mgmt/v1/errors/400",
					"code": "CLUSTERS-MGMT-400",
					"reason": "There are missing version gate agreements for this cluster. See details.",
					"details": [
					  {
						"kind": "VersionGate",
						"id": "999",
						"href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
						"version_raw_id_prefix": "4.10",
						"label": "api.openshift.com/gate-sts",
						"value": "4.10",
						"warning_message": "STS roles must be updated blah blah blah",
						"description": "OpenShift STS clusters include new required cloud provider permissions in OpenShift 4.YY.",
						"documentation_url": "https://access.redhat.com/solutions/0000000",
						"sts_only": true,
						"creation_timestamp": "2023-04-03T06:39:57.057613Z"
					  }
					],
					"operation_id": "8f2d2946-c4ef-4c2f-877b-c19eb17dc918"
				  }`***REMOVED***,
				***REMOVED***,
				// Send acks for all gate agreements
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/gate_agreements"***REMOVED***,
					VerifyJQ(".version_gate.id", "999"***REMOVED***,
					RespondWithJSON(http.StatusCreated, `{
					"kind": "VersionGateAgreement",
					"id": "888",
					"href": "/api/clusters_mgmt/v1/clusters/24g9q8jhdhv66fi41jfiuup5lsvu61fi/gate_agreements/d2e8d371-1033-11ee-9f05-0a580a820bdb",
					"version_gate": {
					  "kind": "VersionGate",
					  "id": "999",
					  "href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
					  "version_raw_id_prefix": "4.10",
					  "label": "api.openshift.com/gate-sts",
					  "value": "4.10",
					  "warning_message": "STS blah blah blah",
					  "description": "OpenShift STS clusters include new required cloud provider permissions in OpenShift 4.YY.",
					  "documentation_url": "https://access.redhat.com/solutions/0000000",
					  "sts_only": true,
					  "creation_timestamp": "2023-04-03T06:39:57.057613Z"
			***REMOVED***,
					"creation_timestamp": "2023-06-21T13:02:06.291443Z"
				  }`***REMOVED***,
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
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/properties",
					  "value": {
						"rosa_tf_commit": "123",
						"rosa_tf_version": "123"
					  }
			***REMOVED***
				]`***REMOVED***,
				***REMOVED******REMOVED***
			// Perform upgrade w/ auto-ack of sts-only gate agreements
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
***REMOVED******REMOVED***

		It("Does nothing if upgrade is in progress", func(***REMOVED*** {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, template***REMOVED***,
				***REMOVED***,
				// Validate upgrade versions
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"***REMOVED***,
					RespondWithJSON(http.StatusOK, v4_10_0Info***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"***REMOVED***,
					RespondWithJSON(http.StatusOK, v4_10_1Info***REMOVED***,
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
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			version = "4.10.1"
***REMOVED***`***REMOVED***
			// Will fail due to upgrade in progress
			Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Cancels and upgrade for the wrong version & schedules new", func(***REMOVED*** {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, template***REMOVED***,
				***REMOVED***,
				// Validate upgrade versions
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"***REMOVED***,
					RespondWithJSON(http.StatusOK, v4_10_0Info***REMOVED***,
				***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"***REMOVED***,
					RespondWithJSON(http.StatusOK, v4_10_1Info***REMOVED***,
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
				// Look for gate agreements by posting an upgrade policy w/ dryRun (no gates necessary***REMOVED***
				CombineHandlers(
					VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies", "dryRun=true"***REMOVED***,
					VerifyJQ(".version", "4.10.1"***REMOVED***,
					RespondWithJSON(http.StatusNoContent, ""***REMOVED***,
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
					RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/properties",
					  "value": {
						"rosa_tf_commit": "123",
						"rosa_tf_version": "123"
					  }
			***REMOVED***
				]`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			// Perform try the upgrade
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			version = "4.10.1"
***REMOVED***`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("Cancels upgrade if version=current_version", func(***REMOVED*** {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, template***REMOVED***,
				***REMOVED***,
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
							"version": "4.10.1",
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
				// Patch the cluster (w/ no changes***REMOVED***
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, template***REMOVED***,
				***REMOVED***,
			***REMOVED***
			// Set version to match current cluster version
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			version = "4.10.0"
***REMOVED***`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("is an error to request a version older than current", func(***REMOVED*** {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template,
						`[
					{
						"op": "replace",
					***REMOVED***: "/version/id",
						"value": "openshift-v4.11.0"
			***REMOVED***]`***REMOVED***,
				***REMOVED***,
			***REMOVED***
			// Set version to before current cluster version, but after version from create
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			version = "4.10.1"
***REMOVED***`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		It("older than current is allowed as long as not changed", func(***REMOVED*** {
			server.AppendHandlers(
				// Refresh cluster state
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template,
						`[
						{
							"op": "replace",
						***REMOVED***: "/version/id",
							"value": "openshift-v4.11.0"
				***REMOVED***]`***REMOVED***,
				***REMOVED***,
				// Patch the cluster (w/ no changes***REMOVED***
				CombineHandlers(
					VerifyRequest(http.MethodPatch, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithJSON(http.StatusOK, template***REMOVED***,
				***REMOVED***,
			***REMOVED***
			// Set version to before current cluster version, but matching what was
			// used during creation (i.e. in state file***REMOVED***
			terraform.Source(`
		  resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
			version = "4.10.0"
***REMOVED***`***REMOVED***
			Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
***REMOVED******REMOVED***

		Context("Un-acked gates", func(***REMOVED*** {
			BeforeEach(func(***REMOVED*** {
				server.AppendHandlers(
					// Refresh cluster state
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
						RespondWithJSON(http.StatusOK, template***REMOVED***,
					***REMOVED***,
					// Validate upgrade versions
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.0"***REMOVED***,
						RespondWithJSON(http.StatusOK, v4_10_0Info***REMOVED***,
					***REMOVED***,
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions/openshift-v4.10.1"***REMOVED***,
						RespondWithJSON(http.StatusOK, v4_10_1Info***REMOVED***,
					***REMOVED***,
					// Look for existing upgrade policies
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies"***REMOVED***,
						RespondWithJSON(http.StatusOK, upgradePoliciesEmpty***REMOVED***,
					***REMOVED***,
					// Look for gate agreements by posting an upgrade policy w/ dryRun
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/upgrade_policies", "dryRun=true"***REMOVED***,
						VerifyJQ(".version", "4.10.1"***REMOVED***,
						RespondWithJSON(http.StatusBadRequest, `{
						"kind": "Error",
						"id": "400",
						"href": "/api/clusters_mgmt/v1/errors/400",
						"code": "CLUSTERS-MGMT-400",
						"reason": "There are missing version gate agreements for this cluster. See details.",
						"details": [
						{
							"kind": "VersionGate",
							"id": "999",
							"href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
							"version_raw_id_prefix": "4.10",
							"label": "api.openshift.com/ackme",
							"value": "4.10",
							"warning_message": "user gotta ack",
							"description": "deprecations... blah blah blah",
							"documentation_url": "https://access.redhat.com/solutions/0000000",
							"sts_only": false,
							"creation_timestamp": "2023-04-03T06:39:57.057613Z"
				***REMOVED***
						],
						"operation_id": "8f2d2946-c4ef-4c2f-877b-c19eb17dc918"
			***REMOVED***`***REMOVED***,
					***REMOVED***,
				***REMOVED***
	***REMOVED******REMOVED***
			It("Fails upgrade for un-acked gates", func(***REMOVED*** {
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
				Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
			It("Fails upgrade if wrong version is acked", func(***REMOVED*** {
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
				upgrade_acknowledgements_for = "1.1"
	***REMOVED***`***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.NotTo(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
			It("It acks gates if correct ack is provided", func(***REMOVED*** {
				server.AppendHandlers(
					// Send acks for all gate agreements
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/123/gate_agreements"***REMOVED***,
						VerifyJQ(".version_gate.id", "999"***REMOVED***,
						RespondWithJSON(http.StatusCreated, `{
					"kind": "VersionGateAgreement",
					"id": "888",
					"href": "/api/clusters_mgmt/v1/clusters/24g9q8jhdhv66fi41jfiuup5lsvu61fi/gate_agreements/d2e8d371-1033-11ee-9f05-0a580a820bdb",
					"version_gate": {
					  "kind": "VersionGate",
					  "id": "999",
					  "href": "/api/clusters_mgmt/v1/version_gates/596326fb-d1ea-11ed-9f29-0a580a8312f9",
					  "version_raw_id_prefix": "4.10",
					  "label": "api.openshift.com/gate-sts",
					  "value": "4.10",
					  "warning_message": "blah blah blah",
					  "description": "whatever",
					  "documentation_url": "https://access.redhat.com/solutions/0000000",
					  "sts_only": false,
					  "creation_timestamp": "2023-04-03T06:39:57.057613Z"
			***REMOVED***,
					"creation_timestamp": "2023-06-21T13:02:06.291443Z"
				  }`***REMOVED***,
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
						RespondWithPatchedJSON(http.StatusCreated, template, `[
					{
					  "op": "add",
					  "path": "/properties",
					  "value": {
						"rosa_tf_commit": "123",
						"rosa_tf_version": "123"
					  }
			***REMOVED***
				]`***REMOVED***,
					***REMOVED***,
				***REMOVED***
				terraform.Source(`
			resource "rhcs_cluster_rosa_classic" "my_cluster" {
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
				version = "4.10.1"
				upgrade_acknowledgements_for = "4.10"
	***REMOVED***`***REMOVED***
				Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	***REMOVED******REMOVED***
***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***

var _ = Describe("rhcs_cluster_rosa_classic - import", func(***REMOVED*** {
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
			"availability_zones": [
				"us-west-1a",
				"us-west-1b",
				"us-west-1c"
			],
			"compute": 3,
			"compute_machine_type": {
				"id": "r5.xlarge"
	***REMOVED***
***REMOVED***,
		"version": {
			"id": "4.10.0"
***REMOVED***
	}`
	Context("rhcs_cluster_rosa_classic - create", func(***REMOVED*** {
		It("can import a cluster", func(***REMOVED*** {
			// Prepare the server:
			server.AppendHandlers(
				// CombineHandlers(
				// 	VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"***REMOVED***,
				// 	RespondWithJSON(http.StatusOK, versionListPage1***REMOVED***,
				// ***REMOVED***,
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
					RespondWithPatchedJSON(http.StatusOK, template, `[
						{
						  "op": "add",
						  "path": "/aws",
						  "value": {
							  "sts" : {
								  "oidc_endpoint_url": "https://127.0.0.2",
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
							"availability_zones": [
								"us-west-1a",
								"us-west-1b",
								"us-west-1c"
							],
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
			  resource "rhcs_cluster_rosa_classic" "my_cluster" { }
			`***REMOVED***
			Expect(terraform.Import("rhcs_cluster_rosa_classic.my_cluster", "123"***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
			resource := terraform.Resource("rhcs_cluster_rosa_classic", "my_cluster"***REMOVED***
			Expect(resource***REMOVED***.To(MatchJQ(".attributes.current_version", "4.10.0"***REMOVED******REMOVED***
***REMOVED******REMOVED***

	}***REMOVED***
}***REMOVED***
