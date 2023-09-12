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

package rosa_operator_roles

***REMOVED***
	"context"
***REMOVED***
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type RosaOperatorRolesDataSource struct {
	awsInquiries *cmv1.AWSInquiriesClient
}

var _ datasource.DataSource = &RosaOperatorRolesDataSource{}
var _ datasource.DataSourceWithConfigure = &RosaOperatorRolesDataSource{}

const (
	DefaultAccountRolePrefix = "ManagedOpenShift"
	serviceAccountFmt        = "system:serviceaccount:%s:%s"
***REMOVED***

func New(***REMOVED*** datasource.DataSource {
	return &RosaOperatorRolesDataSource{}
}

func (s *RosaOperatorRolesDataSource***REMOVED*** Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse***REMOVED*** {
	resp.TypeName = req.ProviderTypeName + "_rosa_operator_roles"
}

func (s *RosaOperatorRolesDataSource***REMOVED*** Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse***REMOVED*** {
	resp.Schema = schema.Schema{
		Description: "List of rosa operator role for a specific cluster.",
		Attributes: map[string]schema.Attribute{
			"operator_role_prefix": schema.StringAttribute{
				Description: "Operator role prefix.",
				Required:    true,
	***REMOVED***,
			"account_role_prefix": schema.StringAttribute{
				Description: "Account role prefix.",
				Optional:    true,
	***REMOVED***,
			"operator_iam_roles": schema.ListNestedAttribute{
				Description: "Operator IAM Roles.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"operator_name": schema.StringAttribute{
							Description: "Operator Name",
							Computed:    true,
				***REMOVED***,
						"operator_namespace": schema.StringAttribute{
							Description: "Kubernetes Namespace",
							Computed:    true,
				***REMOVED***,
						"role_name": schema.StringAttribute{
							Description: "policy name",
							Computed:    true,
				***REMOVED***,
						"policy_name": schema.StringAttribute{
							Description: "policy name",
							Computed:    true,
				***REMOVED***,
						"service_accounts": schema.ListAttribute{
							Description: "service accounts",
							ElementType: types.StringType,
							Computed:    true,
				***REMOVED***,
			***REMOVED***,
		***REMOVED***,
				Computed: true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (s *RosaOperatorRolesDataSource***REMOVED*** Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse***REMOVED*** {
	// Prevent panic if the provider has not been configured:
	if req.ProviderData == nil {
		return
	}

	// Cast the provider data to the specific implementation:
	connection := req.ProviderData.(*sdk.Connection***REMOVED***

	// Get the collection of cloud providers:
	s.awsInquiries = connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.AWSInquiries(***REMOVED***
}

func (s *RosaOperatorRolesDataSource***REMOVED*** Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse***REMOVED*** {
	// Get the state:
	state := &RosaOperatorRolesState{}
	diags := req.Config.Get(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
	if resp.Diagnostics.HasError(***REMOVED*** {
		return
	}

	stsOperatorRolesList, err := s.awsInquiries.STSCredentialRequests(***REMOVED***.List(***REMOVED***.Send(***REMOVED***
	if err != nil {
		description := fmt.Sprintf("Failed to get STS Operator Roles list with error: %v", err***REMOVED***
		tflog.Error(ctx, description***REMOVED***
		resp.Diagnostics.AddError(
			description,
			"hint: validate the credetials (token***REMOVED*** used to run this provider",
		***REMOVED***
		return
	}

	stsOperatorMap := make(map[string]*cmv1.STSOperator***REMOVED***
	roleNameSpaces := make([]string, 0***REMOVED***
	stsOperatorRolesList.Items(***REMOVED***.Each(func(stsCredentialRequest *cmv1.STSCredentialRequest***REMOVED*** bool {
		// TODO: check the MinVersion of the operator role
		tflog.Debug(ctx, fmt.Sprintf("Operator name: %s, namespace %s, service account %s",
			stsCredentialRequest.Operator(***REMOVED***.Name(***REMOVED***,
			stsCredentialRequest.Operator(***REMOVED***.Namespace(***REMOVED***,
			stsCredentialRequest.Operator(***REMOVED***.ServiceAccounts(***REMOVED***,
		***REMOVED******REMOVED***
		// The key can't be stsCredentialRequest.Operator(***REMOVED***.Name(***REMOVED*** because of constants between
		// the name of `ingress_operator_cloud_credentials` and `cloud_network_config_controller_cloud_credentials`
		// both of them includes the same Name `cloud-credentials` and it cannot be fixed
		roleNameSpaces = append(roleNameSpaces, stsCredentialRequest.Operator(***REMOVED***.Namespace(***REMOVED******REMOVED***
		stsOperatorMap[stsCredentialRequest.Operator(***REMOVED***.Namespace(***REMOVED***] = stsCredentialRequest.Operator(***REMOVED***
		return true
	}***REMOVED***

	accountRolePrefix := DefaultAccountRolePrefix
	if !common.IsStringAttributeEmpty(state.AccountRolePrefix***REMOVED*** {
		accountRolePrefix = state.AccountRolePrefix.ValueString(***REMOVED***
	}

	// TODO: use the sts.OperatorRolePrefix(***REMOVED*** if not empty
	// There is a bug in the return value of sts.OperatorRolePrefix(***REMOVED*** - it's always empty string
	sort.Strings(roleNameSpaces***REMOVED***
	for _, key := range roleNameSpaces {
		v := stsOperatorMap[key]
		r := OperatorIAMRole{
			Name:            types.StringValue(v.Name(***REMOVED******REMOVED***,
			Namespace:       types.StringValue(v.Namespace(***REMOVED******REMOVED***,
			RoleName:        types.StringValue(getRoleName(state.OperatorRolePrefix.ValueString(***REMOVED***, v***REMOVED******REMOVED***,
			PolicyName:      types.StringValue(getPolicyName(accountRolePrefix, v.Namespace(***REMOVED***, v.Name(***REMOVED******REMOVED******REMOVED***,
			ServiceAccounts: buildServiceAccountsArray(stsOperatorMap[v.Namespace(***REMOVED***].ServiceAccounts(***REMOVED***, v.Namespace(***REMOVED******REMOVED***,
***REMOVED***
		state.OperatorIAMRoles = append(state.OperatorIAMRoles, &r***REMOVED***
	}

	// Save the state:
	diags = resp.State.Set(ctx, state***REMOVED***
	resp.Diagnostics.Append(diags...***REMOVED***
}

// TODO: should be in a separate repo
func getRoleName(rolePrefix string, operatorRole *cmv1.STSOperator***REMOVED*** string {
	role := fmt.Sprintf("%s-%s-%s", rolePrefix, operatorRole.Namespace(***REMOVED***, operatorRole.Name(***REMOVED******REMOVED***
	if len(role***REMOVED*** > 64 {
		role = role[0:64]
	}
	return role
}

// TODO: should be in a separate repo
func getPolicyName(prefix string, namespace string, name string***REMOVED*** string {
	policy := fmt.Sprintf("%s-%s-%s", prefix, namespace, name***REMOVED***
	if len(policy***REMOVED*** > 64 {
		policy = policy[0:64]
	}
	return policy
}

func buildServiceAccountsArray(serviceAccountArr []string, operatorNamespace string***REMOVED*** types.List {
	svcAcctList := []string{}
	for _, v := range serviceAccountArr {
		svcAcctList = append(svcAcctList, fmt.Sprintf(serviceAccountFmt, operatorNamespace, v***REMOVED******REMOVED***
	}

	serviceAccounts, _ := types.ListValueFrom(context.TODO(***REMOVED***, types.StringType, svcAcctList***REMOVED***
	return serviceAccounts
}
