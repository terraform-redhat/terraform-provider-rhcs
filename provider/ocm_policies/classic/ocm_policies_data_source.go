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

package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/terraform-redhat/terraform-provider-rhcs/provider/ocm_policies/common"
)

const (
	// Policy IDs from type operator roles
	CloudCred                = "openshift_cloud_credential_operator_cloud_credential_operator_iam_ro_creds_policy"
	CloudNetwork             = "openshift_cloud_network_config_controller_cloud_credentials_policy"
	ClusterCSI               = "openshift_cluster_csi_drivers_ebs_cloud_credentials_policy"
	ImageRegistry            = "openshift_image_registry_installer_cloud_credentials_policy"
	IngressOperator          = "openshift_ingress_operator_cloud_credentials_policy"
	SharedVpcIngressOperator = "shared_vpc_openshift_ingress_operator_cloud_credentials_policy"
	MachineAPI               = "openshift_machine_api_aws_cloud_credentials_policy"
	AvoCredentials           = "openshift_aws_vpce_operator_avo_aws_creds_policy"

	// Policy IDs from type account roles
	Installer            = "sts_installer_permission_policy"
	Support              = "sts_support_permission_policy"
	SupportRhSreRole     = "sts_support_rh_sre_role"
	InstanceWorker       = "sts_instance_worker_permission_policy"
	InstanceControlPlane = "sts_instance_controlplane_permission_policy"

	// Policy IDs from type OCM role
	OCMTrustPolicy               = "sts_ocm_trust_policy"
	OCMPermissionPolicy          = "sts_ocm_permission_policy"
	OCMAdminPermissionPolicy     = "sts_ocm_admin_permission_policy"
	OCMNoConsolePermissionPolicy = "sts_ocm_no_console_permission_policy"
)

type OcmPoliciesDataSource struct {
	awsInquiries *cmv1.AWSInquiriesClient
}

var _ datasource.DataSource = &OcmPoliciesDataSource{}
var _ datasource.DataSourceWithConfigure = &OcmPoliciesDataSource{}

func New() datasource.DataSource {
	return &OcmPoliciesDataSource{}
}

func (s *OcmPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policies"
}

func (s *OcmPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of ROSA operator role policies and account role policies.",
		Attributes: map[string]schema.Attribute{
			"operator_role_policies": schema.SingleNestedAttribute{
				Description: "Operator role policies.",
				Attributes: map[string]schema.Attribute{
					CloudCred: schema.StringAttribute{
						Computed: true,
					},
					CloudNetwork: schema.StringAttribute{
						Computed: true,
					},
					ClusterCSI: schema.StringAttribute{
						Computed: true,
					},
					ImageRegistry: schema.StringAttribute{
						Computed: true,
					},
					IngressOperator: schema.StringAttribute{
						Computed: true,
					},
					SharedVpcIngressOperator: schema.StringAttribute{
						Computed: true,
					},
					MachineAPI: schema.StringAttribute{
						Computed: true,
					},
					AvoCredentials: schema.StringAttribute{
						Computed: true,
					},
				},
				Computed: true,
			},
			"account_role_policies": schema.SingleNestedAttribute{
				Description: "Account role policies.",
				Attributes: map[string]schema.Attribute{
					Installer: schema.StringAttribute{
						Computed: true,
					},
					Support: schema.StringAttribute{
						Computed: true,
					},
					SupportRhSreRole: schema.StringAttribute{
						Computed: true,
					},
					InstanceWorker: schema.StringAttribute{
						Computed: true,
					},
					InstanceControlPlane: schema.StringAttribute{
						Computed: true,
					},
				},
				Computed: true,
			},
			"ocm_role_policies": schema.SingleNestedAttribute{
				Description: "OCM role permission policies and trust policy.",
				Attributes: map[string]schema.Attribute{
					OCMTrustPolicy: schema.StringAttribute{
						Description: "Trust policy for the OCM role.",
						Computed:    true,
					},
					OCMPermissionPolicy: schema.StringAttribute{
						Description: "Permission policy for the standard OCM role.",
						Computed:    true,
					},
					OCMAdminPermissionPolicy: schema.StringAttribute{
						Description: "Permission policy for the admin OCM role.",
						Computed:    true,
					},
					OCMNoConsolePermissionPolicy: schema.StringAttribute{
						Description: "Permission policy for the no-console OCM role.",
						Computed:    true,
					},
				},
				Computed: true,
			},
		},
	}
}

func (s *OcmPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection)

	// Get the collection of cloud providers:
	s.awsInquiries = connection.ClustersMgmt().V1().AWSInquiries()
}

func (s *OcmPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the state:
	state := &OcmPoliciesState{}
	diags := req.Config.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policiesResponse, err := s.awsInquiries.STSPolicies().List().Send()
	if err != nil {
		description := fmt.Sprintf("Failed to get policies: %v", err)
		tflog.Error(ctx, description)
		resp.Diagnostics.AddError(
			description,
			"Verify your OCM credentials and connectivity to the OCM API",
		)
		return
	}

	operatorRolePolicies := OperatorRolePolicies{}
	accountRolePolicies := AccountRolePolicies{}
	ocmRolePolicies := OCMRolePolicies{}
	policiesResponse.Items().Each(func(awsPolicy *cmv1.AWSSTSPolicy) bool {
		tflog.Debug(ctx, fmt.Sprintf("policy id: %s ", awsPolicy.ID()))
		switch awsPolicy.ID() {
		// operator roles
		case CloudCred:
			operatorRolePolicies.CloudCred = types.StringValue(awsPolicy.Details())
		case CloudNetwork:
			operatorRolePolicies.CloudNetwork = types.StringValue(awsPolicy.Details())
		case ClusterCSI:
			operatorRolePolicies.ClusterCSI = types.StringValue(awsPolicy.Details())
		case ImageRegistry:
			operatorRolePolicies.ImageRegistry = types.StringValue(awsPolicy.Details())
		case IngressOperator:
			operatorRolePolicies.IngressOperator = types.StringValue(awsPolicy.Details())
		case SharedVpcIngressOperator:
			operatorRolePolicies.SharedVpcIngressOperator = types.StringValue(awsPolicy.Details())
		case MachineAPI:
			operatorRolePolicies.MachineAPI = types.StringValue(awsPolicy.Details())
		case AvoCredentials:
			operatorRolePolicies.AvoCredentials = types.StringValue(awsPolicy.Details())
		// account roles
		case Installer:
			accountRolePolicies.Installer = types.StringValue(awsPolicy.Details())
		case Support:
			accountRolePolicies.Support = types.StringValue(awsPolicy.Details())
		case "sts_support_trust_policy":
			jitRole, err := common.ParseRhSupportRole(ctx, awsPolicy.Details())
			if err != nil {
				resp.Diagnostics.AddError("failed to fetch Classic policies", fmt.Sprintf("%v", err))
			}
			accountRolePolicies.SupportRhSreRole = types.StringValue(jitRole)
		case InstanceWorker:
			accountRolePolicies.InstanceWorker = types.StringValue(awsPolicy.Details())
		case InstanceControlPlane:
			accountRolePolicies.InstanceControlPlane = types.StringValue(awsPolicy.Details())
		// OCM role policies
		case OCMTrustPolicy:
			ocmRolePolicies.TrustPolicy = types.StringValue(awsPolicy.Details())
		case OCMPermissionPolicy:
			ocmRolePolicies.PermissionPolicy = types.StringValue(awsPolicy.Details())
		case OCMAdminPermissionPolicy:
			ocmRolePolicies.AdminPermissionPolicy = types.StringValue(awsPolicy.Details())
		case OCMNoConsolePermissionPolicy:
			ocmRolePolicies.NoConsolePermissionPolicy = types.StringValue(awsPolicy.Details())
		default:
			tflog.Debug(ctx, fmt.Sprintf("Unknown policy ID: %s", awsPolicy.ID()))
		}
		return true
	})
	if resp.Diagnostics.HasError() {
		return
	}
	state.OperatorRolePolicies = &operatorRolePolicies
	state.AccountRolePolicies = &accountRolePolicies
	state.OCMRolePolicies = &ocmRolePolicies

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
