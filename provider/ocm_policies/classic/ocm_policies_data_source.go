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

	// Policy IDs from type account roles
	Installer            = "sts_installer_permission_policy"
	Support              = "sts_support_permission_policy"
	InstanceWorker       = "sts_instance_worker_permission_policy"
	InstanceControlPlane = "sts_instance_controlplane_permission_policy"
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
					InstanceWorker: schema.StringAttribute{
						Computed: true,
					},
					InstanceControlPlane: schema.StringAttribute{
						Computed: true,
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
		tflog.Error(ctx, "Failed to get policies")
		return
	}

	operatorRolePolicies := OperatorRolePolicies{}
	accountRolePolicies := AccountRolePolicies{}
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
		// account roles
		case Installer:
			accountRolePolicies.Installer = types.StringValue(awsPolicy.Details())
		case Support:
			accountRolePolicies.Support = types.StringValue(awsPolicy.Details())
		case InstanceWorker:
			accountRolePolicies.InstanceWorker = types.StringValue(awsPolicy.Details())
		case InstanceControlPlane:
			accountRolePolicies.InstanceControlPlane = types.StringValue(awsPolicy.Details())
		default:
			tflog.Debug(ctx, "This is neither operator role policy nor account role policy")
		}
		return true
	})

	state.OperatorRolePolicies = &operatorRolePolicies
	state.AccountRolePolicies = &accountRolePolicies

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
