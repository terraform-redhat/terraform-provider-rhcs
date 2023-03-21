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
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

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
***REMOVED***

type OcmPoliciesDataSourceType struct {
}

type OcmPoliciesDataSource struct {
	logger       logging.Logger
	awsInquiries *cmv1.AWSInquiriesClient
}

func (t *OcmPoliciesDataSourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "List of rosa operator role for a specific cluster.",
		Attributes: map[string]tfsdk.Attribute{
			"operator_role_policies": {
				Description: "Operator role policies.",
				Attributes:  operatorRolePoliciesNames(***REMOVED***,
				Computed:    true,
	***REMOVED***,
			"account_role_policies": {
				Description: "Account role policies.",
				Attributes:  accountRolePoliciesNames(***REMOVED***,
				Computed:    true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func accountRolePoliciesNames(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"sts_installer_permission_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"sts_support_permission_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"sts_instance_worker_permission_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"sts_instance_controlplane_permission_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
	}***REMOVED***
}

func operatorRolePoliciesNames(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"openshift_cloud_credential_operator_cloud_credential_operator_iam_ro_creds_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"openshift_cloud_network_config_controller_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"openshift_cluster_csi_drivers_ebs_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"openshift_image_registry_installer_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"openshift_ingress_operator_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
		"openshift_machine_api_aws_cloud_credentials_policy": {
			Type:     types.StringType,
			Computed: true,
***REMOVED***,
	}***REMOVED***
}

func (t *OcmPoliciesDataSourceType***REMOVED*** NewDataSource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.DataSource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	awsInquiries := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.AWSInquiries(***REMOVED***

	// Create the resource:
	result = &OcmPoliciesDataSource{
		logger:       parent.logger,
		awsInquiries: awsInquiries,
	}
	return
}

func (t *OcmPoliciesDataSource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadDataSourceRequest,
	response *tfsdk.ReadDataSourceResponse***REMOVED*** {
	// Get the state:
	state := &OcmPoliciesState{}
	diags := request.Config.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	policiesResponse, err := t.awsInquiries.STSPolicies(***REMOVED***.List(***REMOVED***.Send(***REMOVED***
	if err != nil {
		t.logger.Error(ctx, "Failed to get policies"***REMOVED***
		return
	}

	operatorRolePolicies := OperatorRolePolicies{}
	accountRolePolicies := AccountRolePolicies{}
	policiesResponse.Items(***REMOVED***.Each(func(awsPolicy *cmv1.AWSSTSPolicy***REMOVED*** bool {
		t.logger.Debug(ctx, "policy id: %s ", awsPolicy.ID(***REMOVED******REMOVED***
		switch awsPolicy.ID(***REMOVED*** {
		// operator roles
		case CloudCred:
			operatorRolePolicies.CloudCred = types.String{Value: awsPolicy.Details(***REMOVED***}
		case CloudNetwork:
			operatorRolePolicies.CloudNetwork = types.String{Value: awsPolicy.Details(***REMOVED***}
		case ClusterCSI:
			operatorRolePolicies.ClusterCSI = types.String{Value: awsPolicy.Details(***REMOVED***}
		case ImageRegistry:
			operatorRolePolicies.ImageRegistry = types.String{Value: awsPolicy.Details(***REMOVED***}
		case IngressOperator:
			operatorRolePolicies.IngressOperator = types.String{Value: awsPolicy.Details(***REMOVED***}
		case MachineAPI:
			operatorRolePolicies.MachineAPI = types.String{Value: awsPolicy.Details(***REMOVED***}
		// account roles
		case Installer:
			accountRolePolicies.Installer = types.String{Value: awsPolicy.Details(***REMOVED***}
		case Support:
			accountRolePolicies.Support = types.String{Value: awsPolicy.Details(***REMOVED***}
		case InstanceWorker:
			accountRolePolicies.InstanceWorker = types.String{Value: awsPolicy.Details(***REMOVED***}
		case InstanceControlPlane:
			accountRolePolicies.InstanceControlPlane = types.String{Value: awsPolicy.Details(***REMOVED***}
		default:
			t.logger.Debug(ctx, "This is neither operator role policy nor account role policy"***REMOVED***
***REMOVED***
		return true
	}***REMOVED***

	state.OperatorRolePolicies = &operatorRolePolicies
	state.AccountRolePolicies = &accountRolePolicies

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}
