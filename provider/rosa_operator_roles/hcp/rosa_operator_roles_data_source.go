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

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type RosaOperatorRolesDataSource struct {
	awsInquiries *cmv1.AWSInquiriesClient
}

var _ datasource.DataSource = &RosaOperatorRolesDataSource{}
var _ datasource.DataSourceWithConfigure = &RosaOperatorRolesDataSource{}

const (
	DefaultAccountRolePrefix = "ManagedOpenShift"
	serviceAccountFmt        = "system:serviceaccount:%s:%s"
)

func New() datasource.DataSource {
	return &RosaOperatorRolesDataSource{}
}

func (s *RosaOperatorRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hcp_rosa_operator_roles"
}

func (s *RosaOperatorRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of ROSA operator role for a specific cluster.",
		Attributes: map[string]schema.Attribute{
			"operator_role_prefix": schema.StringAttribute{
				Description: "Operator role prefix.",
				Required:    true,
			},
			"account_role_prefix": schema.StringAttribute{
				Description: "Account role prefix.",
				Optional:    true,
			},
			"operator_iam_roles": schema.ListNestedAttribute{
				Description: "Operator IAM Roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"operator_name": schema.StringAttribute{
							Description: "Operator Name",
							Computed:    true,
						},
						"operator_namespace": schema.StringAttribute{
							Description: "Kubernetes Namespace",
							Computed:    true,
						},
						"role_name": schema.StringAttribute{
							Description: "policy name",
							Computed:    true,
						},
						"policy_name": schema.StringAttribute{
							Description: "policy name",
							Computed:    true,
						},
						"service_accounts": schema.ListAttribute{
							Description: "service accounts",
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
		},
	}
	return
}

func (s *RosaOperatorRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection)

	// Get the collection of cloud providers:
	s.awsInquiries = connection.ClustersMgmt().V1().AWSInquiries()
}

func (s *RosaOperatorRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get the state:
	state := &RosaOperatorRolesState{}
	diags := req.Config.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stsOperatorRolesList, err := s.awsInquiries.STSCredentialRequests().
		List().Parameter("is_hypershift", true).Send()
	if err != nil {
		description := fmt.Sprintf("Failed to get STS Operator Roles list with error: %v", err)
		tflog.Error(ctx, description)
		resp.Diagnostics.AddError(
			description,
			"hint: validate the credetials (token) used to run this provider",
		)
		return
	}

	stsOperatorMap := make(map[string]*cmv1.STSOperator)
	roleNameSpaces := make([]string, 0)
	stsOperatorRolesList.Items().Each(func(stsCredentialRequest *cmv1.STSCredentialRequest) bool {
		// TODO: check the MinVersion of the operator role
		tflog.Debug(ctx, fmt.Sprintf("Operator name: %s, namespace %s, service account %s",
			stsCredentialRequest.Operator().Name(),
			stsCredentialRequest.Operator().Namespace(),
			stsCredentialRequest.Operator().ServiceAccounts(),
		))
		// The key can't be stsCredentialRequest.Operator().Name() because of constants between
		// the name of `ingress_operator_cloud_credentials` and `cloud_network_config_controller_cloud_credentials`
		// both of them includes the same Name `cloud-credentials` and it cannot be fixed
		roleNameSpaces = append(roleNameSpaces, stsCredentialRequest.Operator().Namespace())
		stsOperatorMap[stsCredentialRequest.Operator().Namespace()] = stsCredentialRequest.Operator()
		return true
	})

	accountRolePrefix := DefaultAccountRolePrefix
	if !common.IsStringAttributeUnknownOrEmpty(state.AccountRolePrefix) {
		accountRolePrefix = state.AccountRolePrefix.ValueString()
	}

	// TODO: use the sts.OperatorRolePrefix() if not empty
	// There is a bug in the return value of sts.OperatorRolePrefix() - it's always empty string
	sort.Strings(roleNameSpaces)
	for _, key := range roleNameSpaces {
		v := stsOperatorMap[key]
		r := OperatorIAMRole{
			Name:            types.StringValue(v.Name()),
			Namespace:       types.StringValue(v.Namespace()),
			RoleName:        types.StringValue(getRoleName(state.OperatorRolePrefix.ValueString(), v)),
			PolicyName:      types.StringValue(getPolicyName(accountRolePrefix, v.Namespace(), v.Name())),
			ServiceAccounts: buildServiceAccountsArray(stsOperatorMap[v.Namespace()].ServiceAccounts(), v.Namespace()),
		}
		state.OperatorIAMRoles = append(state.OperatorIAMRoles, &r)
	}

	// Save the state:
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// TODO: should be in a separate repo
func getRoleName(rolePrefix string, operatorRole *cmv1.STSOperator) string {
	role := fmt.Sprintf("%s-%s-%s", rolePrefix, operatorRole.Namespace(), operatorRole.Name())
	if len(role) > 64 {
		role = role[0:64]
	}
	return role
}

// TODO: should be in a separate repo
func getPolicyName(prefix string, namespace string, name string) string {
	policy := fmt.Sprintf("%s-%s-%s", prefix, namespace, name)
	if len(policy) > 64 {
		policy = policy[0:64]
	}
	return policy
}

func buildServiceAccountsArray(serviceAccountArr []string, operatorNamespace string) types.List {
	svcAcctList := []string{}
	for _, v := range serviceAccountArr {
		svcAcctList = append(svcAcctList, fmt.Sprintf(serviceAccountFmt, operatorNamespace, v))
	}

	serviceAccounts, _ := types.ListValueFrom(context.TODO(), types.StringType, svcAcctList)
	return serviceAccounts
}
