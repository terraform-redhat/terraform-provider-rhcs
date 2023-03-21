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

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
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

type OcmPoliciesDataSourceType struct {
}

type OcmPoliciesDataSource struct {
	logger       logging.Logger
	awsInquiries *cmv1.AWSInquiriesClient
}

func (t *OcmPoliciesDataSourceType) GetSchema(ctx context.Context) (result tfsdk.Schema,
	diags diag.Diagnostics) {
	result = tfsdk.Schema{
		Description: "List of rosa operator role for a specific cluster.",
		Attributes: map[string]tfsdk.Attribute{
			"operator_role_policies": {
				Description: "Operator role policies.",
				Attributes:  operatorRolePoliciesNames(),
				Computed:    true,
			},
			"account_role_policies": {
				Description: "Account role policies.",
				Attributes:  accountRolePoliciesNames(),
				Computed:    true,
			},
		},
	}
	return
}

func accountRolePoliciesNames() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"sts_installer_permission_policy": {
			Type:     types.StringType,
			Computed: true,
		},
		"sts_support_permission_policy": {
			Type:     types.StringType,
			Computed: true,
		},
		"sts_instance_worker_permission_policy": {
			Type:     types.StringType,
			Computed: true,
		},
		"sts_instance_controlplane_permission_policy": {
			Type:     types.StringType,
			Computed: true,
		},
	})
}

func operatorRolePoliciesNames() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"openshift_cloud_credential_operator_cloud_credential_operator_iam_ro_creds_policy": {
			Type:     types.StringType,
			Computed: true,
		},
		"openshift_cloud_network_config_controller_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
		},
		"openshift_cluster_csi_drivers_ebs_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
		},
		"openshift_image_registry_installer_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
		},
		"openshift_ingress_operator_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
		},
		"openshift_machine_api_aws_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
		},
	})
}

func (t *OcmPoliciesDataSourceType) NewDataSource(ctx context.Context,
	p tfsdk.Provider) (result tfsdk.DataSource, diags diag.Diagnostics) {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider)

	// Get the collection of clusters:
	awsInquiries := parent.connection.ClustersMgmt().V1().AWSInquiries()

	// Create the resource:
	result = &OcmPoliciesDataSource{
		logger:       parent.logger,
		awsInquiries: awsInquiries,
	}
	return
}

func (t *OcmPoliciesDataSource) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest,
	response *tfsdk.ReadDataSourceResponse) {
	// Get the state:
	state := &OcmPoliciesState{}
	diags := request.Config.Get(ctx, state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	policiesResponse, err := t.awsInquiries.STSPolicies().List().Send()
	if err != nil {
		t.logger.Error(ctx, "Failed to get policies")
		return
	}

	operatorRolePolicies := OperatorRolePolicies{}
	accountRolePolicies := AccountRolePolicies{}
	policiesResponse.Items().Each(func(awsPolicy *cmv1.AWSSTSPolicy) bool {
		t.logger.Debug(ctx, "policy id: %s ", awsPolicy.ID())
		switch awsPolicy.ID() {
		// operator roles
		case CloudCred:
			operatorRolePolicies.CloudCred = types.String{Value: awsPolicy.Details()}
		case CloudNetwork:
			operatorRolePolicies.CloudNetwork = types.String{Value: awsPolicy.Details()}
		case ClusterCSI:
			operatorRolePolicies.ClusterCSI = types.String{Value: awsPolicy.Details()}
		case ImageRegistry:
			operatorRolePolicies.ImageRegistry = types.String{Value: awsPolicy.Details()}
		case IngressOperator:
			operatorRolePolicies.IngressOperator = types.String{Value: awsPolicy.Details()}
		case MachineAPI:
			operatorRolePolicies.MachineAPI = types.String{Value: awsPolicy.Details()}
		// account roles
		case Installer:
			accountRolePolicies.Installer = types.String{Value: awsPolicy.Details()}
		case Support:
			accountRolePolicies.Support = types.String{Value: awsPolicy.Details()}
		case InstanceWorker:
			accountRolePolicies.InstanceWorker = types.String{Value: awsPolicy.Details()}
		case InstanceControlPlane:
			accountRolePolicies.InstanceControlPlane = types.String{Value: awsPolicy.Details()}
		default:
			t.logger.Debug(ctx, "This is neither operator role policy nor account role policy")
		}
		return true
	})

	state.OperatorRolePolicies = &operatorRolePolicies
	state.AccountRolePolicies = &accountRolePolicies

	// Save the state:
	diags = response.State.Set(ctx, state)
	response.Diagnostics.Append(diags...)
}
