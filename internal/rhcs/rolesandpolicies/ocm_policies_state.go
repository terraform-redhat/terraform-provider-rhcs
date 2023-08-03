/*
Copyright (c) 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rolesandpolicies

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// Policy IDs from type operator roles
	CloudCred       = "openshift_cloud_credential_operator_cloud_credential_operator_iam_ro_creds_policy"
	CloudNetwork    = "openshift_cloud_network_config_controller_cloud_credentials_policy"
	ClusterCSI      = "openshift_cluster_csi_drivers_ebs_cloud_credentials_policy"
	ImageRegistry   = "openshift_image_registry_installer_cloud_credentials_policy"
	IngressOperator = "openshift_ingress_operator_cloud_credentials_policy"
	MachineAPI      = "openshift_machine_api_aws_cloud_credentials_policy"

	// Policy IDs from type account roles
	Installer            = "sts_installer_permission_policy"
	Support              = "sts_support_permission_policy"
	InstanceWorker       = "sts_instance_worker_permission_policy"
	InstanceControlPlane = "sts_instance_controlplane_permission_policy"
)

type OperatorRolePolicies struct {
	CloudCred       string `tfsdk:"openshift_cloud_credential_operator_cloud_credential_operator_iam_ro_creds_policy"`
	CloudNetwork    string `tfsdk:"openshift_cloud_network_config_controller_cloud_credentials_policy"`
	ClusterCSI      string `tfsdk:"openshift_cluster_csi_drivers_ebs_cloud_credentials_policy"`
	ImageRegistry   string `tfsdk:"openshift_image_registry_installer_cloud_credentials_policy"`
	IngressOperator string `tfsdk:"openshift_ingress_operator_cloud_credentials_policy"`
	MachineAPI      string `tfsdk:"openshift_machine_api_aws_cloud_credentials_policy"`
}

type AccountRolePolicies struct {
	Installer            string `tfsdk:"sts_installer_permission_policy"`
	Support              string `tfsdk:"sts_support_permission_policy"`
	InstanceWorker       string `tfsdk:"sts_instance_worker_permission_policy"`
	InstanceControlPlane string `tfsdk:"sts_instance_controlplane_permission_policy"`
}

func ocmPoliciesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"operator_role_policies": {
			Description: "Operator role policies.",
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: operatorRolePoliciesNames(),
			},
			Computed: true,
		},
		"account_role_policies": {
			Description: "Account role policies.",
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: accountRolePoliciesNames(),
			},
			Computed: true,
		},
	}
}

func operatorRolePoliciesNames() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		CloudCred: {
			Type:     schema.TypeString,
			Computed: true,
		},
		CloudNetwork: {
			Type:     schema.TypeString,
			Computed: true,
		},
		ClusterCSI: {
			Type:     schema.TypeString,
			Computed: true,
		},
		ImageRegistry: {
			Type:     schema.TypeString,
			Computed: true,
		},
		IngressOperator: {
			Type:     schema.TypeString,
			Computed: true,
		},
		MachineAPI: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func accountRolePoliciesNames() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		Installer: {
			Type:     schema.TypeString,
			Computed: true,
		},
		Support: {
			Type:     schema.TypeString,
			Computed: true,
		},
		InstanceWorker: {
			Type:     schema.TypeString,
			Computed: true,
		},
		InstanceControlPlane: {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func FlatOperatorRolePolicies(operatorRolePolicies OperatorRolePolicies) []interface{} {
	result := make(map[string]interface{})
	result[CloudCred] = operatorRolePolicies.CloudCred
	result[CloudNetwork] = operatorRolePolicies.CloudNetwork
	result[ClusterCSI] = operatorRolePolicies.ClusterCSI
	result[ImageRegistry] = operatorRolePolicies.ImageRegistry
	result[IngressOperator] = operatorRolePolicies.IngressOperator
	result[MachineAPI] = operatorRolePolicies.MachineAPI
	return []interface{}{result}
}

func FlatAccountRolePolicies(accountRolePolicies AccountRolePolicies) []interface{} {
	result := make(map[string]interface{})
	result[Installer] = accountRolePolicies.Installer
	result[Support] = accountRolePolicies.Support
	result[InstanceWorker] = accountRolePolicies.InstanceWorker
	result[InstanceControlPlane] = accountRolePolicies.InstanceControlPlane
	return []interface{}{result}
}
