package provider

***REMOVED***
***REMOVED***
***REMOVED***

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
***REMOVED***                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
***REMOVED***

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
	***REMOVED***
***REMOVED***,
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
	***REMOVED***
***REMOVED***
	]
}`
***REMOVED***

var _ = Describe("ROSA Operator IAM roles data source", func(***REMOVED*** {

	It("Can list Operator IAM roles", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/aws_inquiries/sts_credential_requests"***REMOVED***,
				RespondWithJSON(http.StatusOK, getStsCredentialRequests***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_rosa_operator_roles" "operator_roles" {
			  operator_role_prefix = "terraform-operator"
			  account_role_prefix = "TerraformAccountPrefix"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_rosa_operator_roles", "operator_roles"***REMOVED***
		//Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items | length`, 1***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.operator_role_prefix`, "terraform-operator"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.account_role_prefix`, "TerraformAccountPrefix"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.operator_iam_roles | length`, 2***REMOVED******REMOVED***
		compareResultOfRoles(resource, 1,
			"ebs-cloud-credentials",
			"openshift-cluster-csi-drivers",
			"TerraformAccountPrefix-openshift-cluster-csi-drivers-ebs-cloud-c",
			"terraform-operator-openshift-cluster-csi-drivers-ebs-cloud-crede",
			2,
			[]string{
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa",
	***REMOVED***,
		***REMOVED***

		compareResultOfRoles(resource, 0,
			"cloud-credentials",
			"openshift-cloud-network-config-controller",
			"TerraformAccountPrefix-openshift-cloud-network-config-controller",
			"terraform-operator-openshift-cloud-network-config-controller-clo",
			1,
			[]string{"system:serviceaccount:openshift-cloud-network-config-controller:cloud-network-config-controller"},
		***REMOVED***
	}***REMOVED***

	It("Can list Operator IAM roles without account role prefix", func(***REMOVED*** {
		// Prepare the server:
		server.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/aws_inquiries/sts_credential_requests"***REMOVED***,
				RespondWithJSON(http.StatusOK, getStsCredentialRequests***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_rosa_operator_roles" "operator_roles" {
			  operator_role_prefix = "terraform-operator"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_rosa_operator_roles", "operator_roles"***REMOVED***
		//Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items | length`, 1***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.operator_role_prefix`, "terraform-operator"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.operator_iam_roles | length`, 2***REMOVED******REMOVED***
		compareResultOfRoles(resource, 1,
			"ebs-cloud-credentials",
			"openshift-cluster-csi-drivers",
			"ManagedOpenShift-openshift-cluster-csi-drivers-ebs-cloud-credent",
			"terraform-operator-openshift-cluster-csi-drivers-ebs-cloud-crede",
			2,
			[]string{
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa",
	***REMOVED***,
		***REMOVED***

		compareResultOfRoles(resource, 0,
			"cloud-credentials",
			"openshift-cloud-network-config-controller",
			"ManagedOpenShift-openshift-cloud-network-config-controller-cloud",
			"terraform-operator-openshift-cloud-network-config-controller-clo",
			1,
			[]string{"system:serviceaccount:openshift-cloud-network-config-controller:cloud-network-config-controller"},
		***REMOVED***
	}***REMOVED***
}***REMOVED***

func compareResultOfRoles(resource interface{}, index int, name, namespace, policyName, roleName string, serviceAccountLen int, serviceAccounts []string***REMOVED*** {
	operatorNameFmt := ".attributes.operator_iam_roles[%v].operator_name"
	operatorNamespaceFmt := ".attributes.operator_iam_roles[%v].operator_namespace"
	policyNameFmt := ".attributes.operator_iam_roles[%v].policy_name"
	roleNameFmt := ".attributes.operator_iam_roles[%v].role_name"
	serviceAccountLenFmt := ".attributes.operator_iam_roles[%v].service_accounts\t|\tlength "
	serviceAccountFmt := ".attributes.operator_iam_roles[%v].service_accounts[%v]"

	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(operatorNameFmt, index***REMOVED***, name***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(operatorNamespaceFmt, index***REMOVED***, namespace***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(policyNameFmt, index***REMOVED***, policyName***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(roleNameFmt, index***REMOVED***, roleName***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(serviceAccountLenFmt, index***REMOVED***, serviceAccountLen***REMOVED******REMOVED***
	for k, v := range serviceAccounts {
		Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(serviceAccountFmt, index, k***REMOVED***, v***REMOVED******REMOVED***
	}
}
