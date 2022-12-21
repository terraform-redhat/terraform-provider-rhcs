package provider

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
)

const (
	// This is the cluster that will be returned by the server when asked to retrieve a cluster with ID 123
	getStsCredentialRequests = `{
	"page": 1,
	"size": 6,
	"total": 6,
	"items": [
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
		}
	]
}`
)

var _ = Describe("ROSA Operator IAM roles data source", func() {

	It("Can list Operator IAM roles", func() {
		// Prepare the server:
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
			  account_role_prefix = "TerraformAccountPrefix"
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("ocm_rosa_operator_roles", "operator_roles")
		//Expect(resource).To(MatchJQ(`.attributes.items | length`, 1))
		Expect(resource).To(MatchJQ(`.attributes.operator_role_prefix`, "terraform-operator"))
		Expect(resource).To(MatchJQ(`.attributes.account_role_prefix`, "TerraformAccountPrefix"))
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "123"))
		Expect(resource).To(MatchJQ(`.attributes.operator_iam_roles | length`, 2))
		compareResultOfRoles(resource, 0,
			"ebs-cloud-credentials",
			"openshift-cluster-csi-drivers",
			"TerraformAccountPrefix-openshift-cluster-csi-drivers-ebs-cloud-c",
			"terraform-operator-openshift-cluster-csi-drivers-ebs-cloud-crede",
			2,
			[]string{
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa",
			},
		)

		compareResultOfRoles(resource, 1,
			"cloud-credentials",
			"openshift-cloud-network-config-controller",
			"TerraformAccountPrefix-openshift-cloud-network-config-controller",
			"terraform-operator-openshift-cloud-network-config-controller-clo",
			1,
			[]string{"system:serviceaccount:openshift-cloud-network-config-controller:cloud-network-config-controller"},
		)
	})

	It("Can list Operator IAM roles without account role prefix", func() {
		// Prepare the server:
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
		  }
		`)
		Expect(terraform.Apply()).To(BeZero())

		// Check the state:
		resource := terraform.Resource("ocm_rosa_operator_roles", "operator_roles")
		//Expect(resource).To(MatchJQ(`.attributes.items | length`, 1))
		Expect(resource).To(MatchJQ(`.attributes.operator_role_prefix`, "terraform-operator"))
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "123"))
		Expect(resource).To(MatchJQ(`.attributes.operator_iam_roles | length`, 2))
		compareResultOfRoles(resource, 0,
			"ebs-cloud-credentials",
			"openshift-cluster-csi-drivers",
			"ManagedOpenShift-openshift-cluster-csi-drivers-ebs-cloud-credent",
			"terraform-operator-openshift-cluster-csi-drivers-ebs-cloud-crede",
			2,
			[]string{
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa",
			},
		)

		compareResultOfRoles(resource, 1,
			"cloud-credentials",
			"openshift-cloud-network-config-controller",
			"ManagedOpenShift-openshift-cloud-network-config-controller-cloud",
			"terraform-operator-openshift-cloud-network-config-controller-clo",
			1,
			[]string{"system:serviceaccount:openshift-cloud-network-config-controller:cloud-network-config-controller"},
		)
	})
})

func compareResultOfRoles(resource interface{}, index int, name, namespace, policyName, roleName string, serviceAccountLen int, serviceAccounts []string) {
	operatorNameFmt := ".attributes.operator_iam_roles[%v].operator_name"
	operatorNamespaceFmt := ".attributes.operator_iam_roles[%v].operator_namespace"
	policyNameFmt := ".attributes.operator_iam_roles[%v].policy_name"
	roleNameFmt := ".attributes.operator_iam_roles[%v].role_name"
	serviceAccountLenFmt := ".attributes.operator_iam_roles[%v].service_accounts\t|\tlength "
	serviceAccountFmt := ".attributes.operator_iam_roles[%v].service_accounts[%v]"

	Expect(resource).To(MatchJQ(fmt.Sprintf(operatorNameFmt, index), name))
	Expect(resource).To(MatchJQ(fmt.Sprintf(operatorNamespaceFmt, index), namespace))
	Expect(resource).To(MatchJQ(fmt.Sprintf(policyNameFmt, index), policyName))
	Expect(resource).To(MatchJQ(fmt.Sprintf(roleNameFmt, index), roleName))
	Expect(resource).To(MatchJQ(fmt.Sprintf(serviceAccountLenFmt, index), serviceAccountLen))
	for k, v := range serviceAccounts {
		Expect(resource).To(MatchJQ(fmt.Sprintf(serviceAccountFmt, index, k), v))
	}
}
