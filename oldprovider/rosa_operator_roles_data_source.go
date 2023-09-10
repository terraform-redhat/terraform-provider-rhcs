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
***REMOVED***
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
***REMOVED***

type RosaOperatorRolesDataSourceType struct {
}

type RosaOperatorRolesDataSource struct {
	awsInquiries *cmv1.AWSInquiriesClient
}

const (
	DefaultAccountRolePrefix = "ManagedOpenShift"
	serviceAccountFmt        = "system:serviceaccount:%s:%s"
***REMOVED***

func (t *RosaOperatorRolesDataSourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "List of rosa operator role for a specific cluster.",
		Attributes: map[string]tfsdk.Attribute{
			"operator_role_prefix": {
				Description: "Operator role prefix.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
			"account_role_prefix": {
				Description: "Account role prefix.",
				Type:        types.StringType,
				Optional:    true,
	***REMOVED***,
			"operator_iam_roles": {
				Description: "Operator IAM Roles.",
				Attributes: tfsdk.ListNestedAttributes(
					t.itemAttributes(***REMOVED***,
					tfsdk.ListNestedAttributesOptions{},
				***REMOVED***,
				Computed: true,
	***REMOVED***,
***REMOVED***,
	}
	return
}

func (t *RosaOperatorRolesDataSourceType***REMOVED*** itemAttributes(***REMOVED*** map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"operator_name": {
			Description: "Operator Name",
			Type:        types.StringType,
			Computed:    true,
***REMOVED***,
		"operator_namespace": {
			Description: "Kubernetes Namespace",
			Type:        types.StringType,
			Computed:    true,
***REMOVED***,
		"role_name": {
			Description: "policy name",
			Type:        types.StringType,
			Computed:    true,
***REMOVED***,
		"policy_name": {
			Description: "policy name",
			Type:        types.StringType,
			Computed:    true,
***REMOVED***,
		"service_accounts": {
			Description: "service accounts",
			Type: types.ListType{
				ElemType: types.StringType,
	***REMOVED***,
			Computed: true,
***REMOVED***,
	}
}

func (t *RosaOperatorRolesDataSourceType***REMOVED*** NewDataSource(ctx context.Context,
	p tfsdk.Provider***REMOVED*** (result tfsdk.DataSource, diags diag.Diagnostics***REMOVED*** {
	// Cast the provider interface to the specific implementation:
	parent := p.(*Provider***REMOVED***

	// Get the collection of clusters:
	awsInquiries := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.AWSInquiries(***REMOVED***

	// Create the resource:
	result = &RosaOperatorRolesDataSource{
		awsInquiries: awsInquiries,
	}
	return
}

func (t *RosaOperatorRolesDataSource***REMOVED*** Read(ctx context.Context, request tfsdk.ReadDataSourceRequest,
	response *tfsdk.ReadDataSourceResponse***REMOVED*** {
	// Get the state:
	state := &RosaOperatorRolesState{}
	diags := request.Config.Get(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
	if response.Diagnostics.HasError(***REMOVED*** {
		return
	}

	stsOperatorRolesList, err := t.awsInquiries.STSCredentialRequests(***REMOVED***.List(***REMOVED***.Send(***REMOVED***
	if err != nil {
		description := fmt.Sprintf("Failed to get STS Operator Roles list with error: %v", err***REMOVED***
		tflog.Error(ctx, description***REMOVED***
		response.Diagnostics.AddError(
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
	if !state.AccountRolePrefix.Unknown && !state.AccountRolePrefix.Null && state.AccountRolePrefix.Value != "" {
		accountRolePrefix = state.AccountRolePrefix.Value
	}

	// TODO: use the sts.OperatorRolePrefix(***REMOVED*** if not empty
	// There is a bug in the return value of sts.OperatorRolePrefix(***REMOVED*** - it's always empty string
	sort.Strings(roleNameSpaces***REMOVED***
	for _, key := range roleNameSpaces {
		r := OperatorIAMRole{
			Name: types.String{
				Value: stsOperatorMap[key].Name(***REMOVED***,
	***REMOVED***,
			Namespace: types.String{
				Value: stsOperatorMap[key].Namespace(***REMOVED***,
	***REMOVED***,
			RoleName: types.String{
				Value: getRoleName(state.OperatorRolePrefix.Value, stsOperatorMap[key]***REMOVED***,
	***REMOVED***,
			PolicyName: types.String{
				Value: getPolicyName(accountRolePrefix, stsOperatorMap[key].Namespace(***REMOVED***, stsOperatorMap[key].Name(***REMOVED******REMOVED***,
	***REMOVED***,
			ServiceAccounts: buildServiceAccountsArray(stsOperatorMap[stsOperatorMap[key].Namespace(***REMOVED***].ServiceAccounts(***REMOVED***, stsOperatorMap[key].Namespace(***REMOVED******REMOVED***,
***REMOVED***
		state.OperatorIAMRoles = append(state.OperatorIAMRoles, &r***REMOVED***
	}

	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
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
	serviceAccounts := types.List{
		ElemType: types.StringType,
		Elems:    []attr.Value{},
	}

	for _, v := range serviceAccountArr {
		serviceAccount := fmt.Sprintf(serviceAccountFmt, operatorNamespace, v***REMOVED***
		serviceAccounts.Elems = append(serviceAccounts.Elems, types.String{Value: serviceAccount}***REMOVED***
	}

	return serviceAccounts
}
