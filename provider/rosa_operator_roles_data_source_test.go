package provider

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

const (
	// This is the cluster that will be returned by the server when asked to retrieve a cluster with ID 123
	getClusterJson = `{
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
	"nodes": {
		"compute": 3,
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
	"state": "ready",
	"aws": {
		"private_link": false,
		"sts": {
			"enabled": true,
			"role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Installer-Role",
			"support_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Support-Role",
			"oidc_endpoint_url": "https://oidc_endpoint_url",
			"thumbprint": "111111",
			"instance_iam_roles": {
				"master_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-ControlPlane-Role",
				"worker_role_arn": "arn:aws:iam::account-id:role/ManagedOpenShift-Worker-Role"
			},
			"operator_role_prefix": "terraform-operator",
			"operator_iam_roles": [{
					"id": "",
					"name": "cloud-credential-operator-iam-ro-creds",
					"namespace": "openshift-cloud-credential-operator",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cloud-credential-operator-cloud-cred",
					"service_account": ""
				},
				{
					"id": "",
					"name": "installer-cloud-credentials",
					"namespace": "openshift-image-registry",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-image-registry-installer-cloud-crede",
					"service_account": ""
				},
				{
					"id": "",
					"name": "cloud-credentials",
					"namespace": "openshift-ingress-operator",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-ingress-operator-cloud-credentials",
					"service_account": ""
				},
				{
					"id": "",
					"name": "ebs-cloud-credentials",
					"namespace": "openshift-cluster-csi-drivers",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cluster-csi-drivers-ebs-cloud-creden",
					"service_account": ""
				},
				{
					"id": "",
					"name": "cloud-credentials",
					"namespace": "openshift-cloud-network-config-controller",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cloud-network-config-controller-clou",
					"service_account": ""
				},
				{
					"id": "",
					"name": "aws-cloud-credentials",
					"namespace": "openshift-machine-api",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-machine-api-aws-cloud-credentials",
					"service_account": ""
				}
			]
		}
	}
}`
	getStsCredentialRequests = `{
	"page": 1,
	"size": 6,
	"total": 6,
	"items": [{
			"kind": "STSOperator",
			"name": "ingress_operator_cloud_credentials",
			"operator": {
				"name": "cloud-credentials",
				"namespace": "openshift-ingress-operator",
				"service_accounts": [
					"ingress-operator"
				],
				"min_version": "",
				"max_version": ""
			}
		},
		{
			"kind": "STSOperator",
			"name": "cluster_csi_drivers_ebs_cloud_credentials",
			"operator": {
				"name": "ebs-cloud-credentials",
				"namespace": "openshift-cluster-csi-drivers",
				"service_accounts": [
					"aws-ebs-csi-driver-operator",
					"aws-ebs-csi-driver-controller-sa"
				],
				"min_version": "",
				"max_version": ""
			}
		},
		{
			"kind": "STSOperator",
			"name": "cloud_network_config_controller_cloud_credentials",
			"operator": {
				"name": "cloud-credentials",
				"namespace": "openshift-cloud-network-config-controller",
				"service_accounts": [
					"cloud-network-config-controller"
				],
				"min_version": "4.10",
				"max_version": ""
			}
		},
		{
			"kind": "STSOperator",
			"name": "machine_api_aws_cloud_credentials",
			"operator": {
				"name": "aws-cloud-credentials",
				"namespace": "openshift-machine-api",
				"service_accounts": [
					"machine-api-controllers"
				],
				"min_version": "",
				"max_version": ""
			}
		},
		{
			"kind": "STSOperator",
			"name": "cloud_credential_operator_cloud_credential_operator_iam_ro_creds",
			"operator": {
				"name": "cloud-credential-operator-iam-ro-creds",
				"namespace": "openshift-cloud-credential-operator",
				"service_accounts": [
					"cloud-credential-operator"
				],
				"min_version": "",
				"max_version": ""
			}
		},
		{
			"kind": "STSOperator",
			"name": "image_registry_installer_cloud_credentials",
			"operator": {
				"name": "installer-cloud-credentials",
				"namespace": "openshift-image-registry",
				"service_accounts": [
					"cluster-image-registry-operator",
					"registry"
				],
				"min_version": "",
				"max_version": ""
			}
		}
	]
}`
)

var _ = Describe("ROSA Operator IAM roles data source", func() {

	It("Can list Operator IAM roles", func() {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"),
				RespondWithJSON(http.StatusOK, getClusterJson),
			),
		)
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/aws_inquiries/sts_credential_requests"),
				RespondWithJSON(http.StatusOK, getStsCredentialRequests),
			),
		)

		// Run the apply command:
		terraform.Source(`
		  data "ocm_rosa_operator_roles" "operator_roles" {
			  cluster_id = "123"
			  operator_role_prefix = "terraform-operator"
			  account_role_prefix = "ManagedOpenShift"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("ocm_rosa_operator_roles", "operator_roles")
		Expect(resource).To(MatchJQ(`.attributes.items | length`, 1))
		Expect(resource).To(MatchJQ(`.attributes.items[0].operator_role_prefix`, "terraform-operator"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].account_role_prefix`, "ManagedOpenShift"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].operator_iam_roles.items | length`, "6"))
		println(resource)
	})
})
