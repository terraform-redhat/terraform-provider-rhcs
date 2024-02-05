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

package hcp

import "github.com/hashicorp/terraform-plugin-framework/types"

type OcmPoliciesState struct {
	OperatorRolePolicies *OperatorRolePolicies `tfsdk:"operator_role_policies"`
	AccountRolePolicies  *AccountRolePolicies  `tfsdk:"account_role_policies"`
}

type OperatorRolePolicies struct {
	ImageRegistry                   types.String `tfsdk:"openshift_hcp_image_registry_installer_cloud_credentials_policy"`
	IngressOperator                 types.String `tfsdk:"openshift_hcp_ingress_operator_cloud_credentials_policy"`
	ClusterCSI                      types.String `tfsdk:"openshift_hcp_cluster_csi_drivers_ebs_cloud_credentials_policy"`
	CloudNetwork                    types.String `tfsdk:"openshift_hcp_cloud_network_config_controller_cloud_credentials_policy"`
	SharedVpcIngressOperator        types.String `tfsdk:"shared_vpc_openshift_ingress_operator_cloud_credentials_policy"`
	KubeControllerManagerKubeSystem types.String `tfsdk:"openshift_hcp_kube_controller_manager_credentials_policy"`
	CapaControllerManagerKubeSystem types.String `tfsdk:"openshift_hcp_capa_controller_manager_credentials_policy"`
	ControlPlaneOperatorKubeSystem  types.String `tfsdk:"openshift_hcp_control_plane_operator_credentials_policy"`
	KmsProviderKubeSystem           types.String `tfsdk:"openshift_hcp_kms_provider_credentials_policy"`
}

type AccountRolePolicies struct {
	Installer      types.String `tfsdk:"sts_hcp_installer_permission_policy"`
	Support        types.String `tfsdk:"sts_hcp_support_permission_policy"`
	InstanceWorker types.String `tfsdk:"sts_hcp_instance_worker_permission_policy"`
}
