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

var _ = Describe("Cluster creation", func(***REMOVED*** {
	// This is the cluster that will be returned by the server when asked to create or retrieve
	// a cluster.
	const template = `{
	  "id": "123",
	  "product": {
		"id": "rosa"
	  },
	  "name": "my-cluster",
	  "cloud_provider": {
	    "id": "aws"
	  },
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
	  "nodes": {
	    "compute": 3,
	    "compute_machine_type": {
	      "id": "r5.xlarge"
	    }
	  },
	  "ccs": {
	    "enabled": false
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
	  "state": "ready"
	}`

	It("Creates basic cluster", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				RespondWithJSON(http.StatusCreated, template***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster" "my_cluster" {
		    name           = "my-cluster"
			product		   = "rosa"
		    cloud_provider = "aws"			
		    cloud_region   = "us-west-1"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates cluster with http proxy", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.proxy.http_proxy`, "http://proxy.com"***REMOVED***,
				VerifyJQ(`.proxy.https_proxy`, "http://proxy.com"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/proxy",
					  "value": {						  
						  "http_proxy" : "http://proxy.com",
						  "https_proxy" : "http://proxy.com"
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster" "my_cluster" {
		    name           = "my-cluster"
			product		   = "rosa"
		    cloud_provider = "aws"			
		    cloud_region   = "us-west-1"
			proxy = {
				http_proxy = "http://proxy.com",
				https_proxy = "http://proxy.com",
	***REMOVED***			
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates cluster with aws subnet ids & private link", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
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
					***REMOVED***: "/availability_zones",
						"value": ["az1", "az2", "az3"]
			***REMOVED***,					
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "private_link": true,
						  "subnet_ids": ["id1", "id2", "id3"]
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster" "my_cluster" {
		    name           = "my-cluster"
			product		   = "rosa"
		    cloud_provider = "aws"			
		    cloud_region   = "us-west-1"
			availability_zones = ["az1","az2","az3"]
			aws_private_link = true
			aws_subnet_ids = [
				"id1", "id2", "id3"
			]
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates cluster when private link is false", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
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
						  "private_link": false
					  }
			***REMOVED***]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  resource "ocm_cluster" "my_cluster" {
		    name           = "my-cluster"
			product		   = "rosa"
		    cloud_provider = "aws"			
		    cloud_region   = "us-west-1"
			aws_private_link = false
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***

	It("Creates rosa sts cluster", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"***REMOVED***,
				VerifyJQ(`.name`, "my-cluster"***REMOVED***,
				VerifyJQ(`.cloud_provider.id`, "aws"***REMOVED***,
				VerifyJQ(`.region.id`, "us-west-1"***REMOVED***,
				VerifyJQ(`.product.id`, "rosa"***REMOVED***,
				VerifyJQ(`.aws.sts.role_arn`, "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role"***REMOVED***,
				VerifyJQ(`.aws.sts.support_role_arn`, "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role"***REMOVED***,
				VerifyJQ(`.aws.sts.instance_iam_roles.master_role_arn`, "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role"***REMOVED***,
				VerifyJQ(`.aws.sts.instance_iam_roles.worker_role_arn`, "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"***REMOVED***,
				VerifyJQ(`.aws.sts.operator_iam_roles.[0].role_arn`, "arn:aws:iam::account-id:role/cloud-credential"***REMOVED***,
				VerifyJQ(`.aws.sts.operator_iam_roles.[1].role_arn`, "arn:aws:iam::account-id:role/image-registry"***REMOVED***,
				VerifyJQ(`.aws.sts.operator_iam_roles.[2].role_arn`, "arn:aws:iam::account-id:role/ingress"***REMOVED***,
				VerifyJQ(`.aws.sts.operator_iam_roles.[3].role_arn`, "arn:aws:iam::account-id:role/ebs"***REMOVED***,
				VerifyJQ(`.aws.sts.operator_iam_roles.[4].role_arn`, "arn:aws:iam::account-id:role/cloud-network-config"***REMOVED***,
				VerifyJQ(`.aws.sts.operator_iam_roles.[5].role_arn`, "arn:aws:iam::account-id:role/machine-api"***REMOVED***,
				RespondWithPatchedJSON(http.StatusOK, template, `[
					{
					  "op": "add",
					  "path": "/aws",
					  "value": {
						  "sts" : {
							  "oidc_endpoint_url": "oidc_endpoint_url",
							  "role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
							  "support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
							  "instance_iam_roles" : {
								"master_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
								"worker_role_arn" : "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
							  },
							  "operator_iam_roles" : [
								{
									"name": "cloud-credential-operator-iam-ro-creds",
									"namespace": "openshift-cloud-credential-operator",
									"role_arn": "arn:aws:iam::account-id:role/cloud-credential"
								  },
								  {
									"name": "installer-cloud-credentials",
									"namespace": "openshift-image-registry",
									"role_arn": "arn:aws:iam::account-id:role/image-registry"
								  },
								  {
									"name": "cloud-credentials",
									"namespace": "openshift-ingress-operator",
									"role_arn": "arn:aws:iam::account-id:role/ingress"
								  },
								  {
									"name": "ebs-cloud-credentials",
									"namespace": "openshift-cluster-csi-drivers",
									"role_arn": "arn:aws:iam::account-id:role/ebs"
								  },
								  {
									"name": "cloud-credentials",
									"namespace": "openshift-cloud-network-config-controller",
									"role_arn": "arn:aws:iam::account-id:role/cloud-network-config"
								  },
								  {
									"name": "aws-cloud-credentials",
									"namespace": "openshift-machine-api",
									"role_arn": "arn:aws:iam::account-id:role/machine-api"
								  }								  
							  ]
						  }
					  }
			***REMOVED***
				  ]`***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		resource "ocm_cluster" "my_cluster" {
			name           = "my-cluster"
			product		   = "rosa"
			cloud_provider = "aws"			
			cloud_region   = "us-west-1"
			sts = {
				role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
				support_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
				operator_iam_roles = [
					{
						name =  "cloud-credential-operator-iam-ro-creds",
						namespace = "openshift-cloud-credential-operator",
						role_arn = "arn:aws:iam::account-id:role/cloud-credential",
			***REMOVED***,
					{
						name =  "installer-cloud-credentials",
						namespace = "openshift-image-registry",
						role_arn = "arn:aws:iam::account-id:role/image-registry",
			***REMOVED***,
					{
						name =  "cloud-credentials",
						namespace = "openshift-ingress-operator",
						role_arn = "arn:aws:iam::account-id:role/ingress",
			***REMOVED***,
					{
						name =  "ebs-cloud-credentials",
						namespace = "openshift-cluster-csi-drivers",
						role_arn = "arn:aws:iam::account-id:role/ebs",
			***REMOVED***,
					{
						name =  "cloud-credentials",
						namespace = "openshift-cloud-network-config-controller",
						role_arn = "arn:aws:iam::account-id:role/cloud-network-config",
			***REMOVED***,
					{
						name =  "aws-cloud-credentials",
						namespace = "openshift-machine-api",
						role_arn = "arn:aws:iam::account-id:role/machine-api",
			***REMOVED***,
				]
				instance_iam_roles = {
				  master_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
				  worker_role_arn = "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
		***REMOVED***,
	***REMOVED***
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***
	}***REMOVED***
}***REMOVED***
