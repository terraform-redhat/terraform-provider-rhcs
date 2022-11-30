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
	getClusterJson = `{
	"kind": "",
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
***REMOVED***
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
			"role_arn": "arn:aws:iam::account-id:role/TerraformAccount-Installer-Role",
			"support_role_arn": "arn:aws:iam::account-id:role/TerraformAccount-Support-Role",
			"oidc_endpoint_url": "https://oidc_endpoint_url",
			"thumbprint": "111111",
			"instance_iam_roles": {
				"master_role_arn": "arn:aws:iam::account-id:role/TerraformAccount-ControlPlane-Role",
				"worker_role_arn": "arn:aws:iam::account-id:role/TerraformAccount-Worker-Role"
	***REMOVED***,
			"operator_role_prefix": "terraform-operator",
			"operator_iam_roles": [
				{
					"id": "",
					"name": "ebs-cloud-credentials",
					"namespace": "openshift-cluster-csi-drivers",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cluster-csi-drivers-ebs-cloud-creden",
					"service_account": ""
		***REMOVED***,
				{
					"id": "",
					"name": "cloud-credentials",
					"namespace": "openshift-cloud-network-config-controller",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cloud-network-config-controller-clou",
					"service_account": ""
		***REMOVED***
			]
***REMOVED***
	}
}`
	getClusterWithDefaultAccountPrefixJson = `{
	"kind": "",
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
***REMOVED***
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
	***REMOVED***,
			"operator_role_prefix": "terraform-operator",
			"operator_iam_roles": [
				{
					"id": "",
					"name": "ebs-cloud-credentials",
					"namespace": "openshift-cluster-csi-drivers",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cluster-csi-drivers-ebs-cloud-creden",
					"service_account": ""
		***REMOVED***,
				{
					"id": "",
					"name": "cloud-credentials",
					"namespace": "openshift-cloud-network-config-controller",
					"role_arn": "arn:aws:iam::765374464689:role/terrafom-operator-openshift-cloud-network-config-controller-clou",
					"service_account": ""
		***REMOVED***
			]
***REMOVED***
	}
}`
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
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, getClusterJson***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_rosa_operator_roles" "operator_roles" {
			  cluster_id = "123"
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
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.cluster_id`, "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.operator_iam_roles | length`, 2***REMOVED******REMOVED***
		compareResultOfRoles(resource, 0,
			"ebs-cloud-credentials",
			"openshift-cluster-csi-drivers",
			"TerraformAccountPrefix-openshift-cluster-csi-drivers-ebs-cloud-c",
			"arn:aws:iam::765374464689:role/terrafom-operator-openshift-cluster-csi-drivers-ebs-cloud-creden",
			"terraform-operator-openshift-cluster-csi-drivers-ebs-cloud-crede",
			2,
			[]string{
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa",
	***REMOVED***,
		***REMOVED***

		compareResultOfRoles(resource, 1,
			"cloud-credentials",
			"openshift-cloud-network-config-controller",
			"TerraformAccountPrefix-openshift-cloud-network-config-controller",
			"arn:aws:iam::765374464689:role/terrafom-operator-openshift-cloud-network-config-controller-clou",
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
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/123"***REMOVED***,
				RespondWithJSON(http.StatusOK, getClusterWithDefaultAccountPrefixJson***REMOVED***,
			***REMOVED***,
		***REMOVED***

		// Run the apply command:
		terraform.Source(`
		  data "ocm_rosa_operator_roles" "operator_roles" {
			  cluster_id = "123"
			  operator_role_prefix = "terraform-operator"
		  }
		`***REMOVED***
		Expect(terraform.Apply(***REMOVED******REMOVED***.To(BeZero(***REMOVED******REMOVED***

		// Check the state:
		resource := terraform.Resource("ocm_rosa_operator_roles", "operator_roles"***REMOVED***
		//Expect(resource***REMOVED***.To(MatchJQ(`.attributes.items | length`, 1***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.operator_role_prefix`, "terraform-operator"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.cluster_id`, "123"***REMOVED******REMOVED***
		Expect(resource***REMOVED***.To(MatchJQ(`.attributes.operator_iam_roles | length`, 2***REMOVED******REMOVED***
		compareResultOfRoles(resource, 0,
			"ebs-cloud-credentials",
			"openshift-cluster-csi-drivers",
			"ManagedOpenShift-openshift-cluster-csi-drivers-ebs-cloud-credent",
			"arn:aws:iam::765374464689:role/terrafom-operator-openshift-cluster-csi-drivers-ebs-cloud-creden",
			"terraform-operator-openshift-cluster-csi-drivers-ebs-cloud-crede",
			2,
			[]string{
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-operator",
				"system:serviceaccount:openshift-cluster-csi-drivers:aws-ebs-csi-driver-controller-sa",
	***REMOVED***,
		***REMOVED***

		compareResultOfRoles(resource, 1,
			"cloud-credentials",
			"openshift-cloud-network-config-controller",
			"ManagedOpenShift-openshift-cloud-network-config-controller-cloud",
			"arn:aws:iam::765374464689:role/terrafom-operator-openshift-cloud-network-config-controller-clou",
			"terraform-operator-openshift-cloud-network-config-controller-clo",
			1,
			[]string{"system:serviceaccount:openshift-cloud-network-config-controller:cloud-network-config-controller"},
		***REMOVED***
	}***REMOVED***
}***REMOVED***

func compareResultOfRoles(resource interface{}, index int, name, namespace, policyName, roleArn, roleName string, serviceAccountLen int, serviceAccounts []string***REMOVED*** {
	operatorNameFmt := ".attributes.operator_iam_roles[%v].operator_name"
	operatorNamespaceFmt := ".attributes.operator_iam_roles[%v].operator_namespace"
	policyNameFmt := ".attributes.operator_iam_roles[%v].policy_name"
	roleArnFmt := ".attributes.operator_iam_roles[%v].role_arn"
	roleNameFmt := ".attributes.operator_iam_roles[%v].role_name"
	serviceAccountLenFmt := ".attributes.operator_iam_roles[%v].service_accounts\t|\tlength "
	serviceAccountFmt := ".attributes.operator_iam_roles[%v].service_accounts[%v]"

	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(operatorNameFmt, index***REMOVED***, name***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(operatorNamespaceFmt, index***REMOVED***, namespace***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(policyNameFmt, index***REMOVED***, policyName***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(roleArnFmt, index***REMOVED***, roleArn***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(roleNameFmt, index***REMOVED***, roleName***REMOVED******REMOVED***
	Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(serviceAccountLenFmt, index***REMOVED***, serviceAccountLen***REMOVED******REMOVED***
	for k, v := range serviceAccounts {
		Expect(resource***REMOVED***.To(MatchJQ(fmt.Sprintf(serviceAccountFmt, index, k***REMOVED***, v***REMOVED******REMOVED***
	}
}
