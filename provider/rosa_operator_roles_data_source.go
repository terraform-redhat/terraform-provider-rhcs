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
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

type RosaOperatorRolesDataSourceType struct {
}

type RosaOperatorRolesDataSource struct {
	logger         logging.Logger
	clustersClient *cmv1.ClustersClient
	awsInquiries   *cmv1.AWSInquiriesClient
}

const DefaultAccountRolePrefix = "ManagedOpenShift"

func (t *RosaOperatorRolesDataSourceType***REMOVED*** GetSchema(ctx context.Context***REMOVED*** (result tfsdk.Schema,
	diags diag.Diagnostics***REMOVED*** {
	result = tfsdk.Schema{
		Description: "List of rosa operator role for a specific cluster.",
		Attributes: map[string]tfsdk.Attribute{
			"cluster_id": {
				Description: "Cluster id.",
				Type:        types.StringType,
				Required:    true,
	***REMOVED***,
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
		"role_arn": {
			Description: "AWS Role ARN",
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
	clustersClient := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Clusters(***REMOVED***
	awsInquiries := parent.connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.AWSInquiries(***REMOVED***

	// Create the resource:
	result = &RosaOperatorRolesDataSource{
		logger:         parent.logger,
		clustersClient: clustersClient,
		awsInquiries:   awsInquiries,
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
		t.logger.Error(ctx, "Failed to get operator list"***REMOVED***
		return
	}

	stsOperatorMap := make(map[string]*cmv1.STSOperator***REMOVED***
	stsOperatorRolesList.Items(***REMOVED***.Each(func(stsCredentialRequest *cmv1.STSCredentialRequest***REMOVED*** bool {
		t.logger.Debug(ctx, "Operator name: %s, namespace %s, service account %s",
			stsCredentialRequest.Operator(***REMOVED***.Name(***REMOVED***,
			stsCredentialRequest.Operator(***REMOVED***.Namespace(***REMOVED***,
			stsCredentialRequest.Operator(***REMOVED***.ServiceAccounts(***REMOVED***,
		***REMOVED***
		// The key can't be stsCredentialRequest.Operator(***REMOVED***.Name(***REMOVED*** because of constants between
		// the name of `ingress_operator_cloud_credentials` and `cloud_network_config_controller_cloud_credentials`
		// both of them includes the same Name `cloud-credentials` and it cannot be fixed
		stsOperatorMap[stsCredentialRequest.Operator(***REMOVED***.Namespace(***REMOVED***] = stsCredentialRequest.Operator(***REMOVED***
		return true
	}***REMOVED***

	get, err := t.clustersClient.Cluster(state.ClusterID.Value***REMOVED***.Get(***REMOVED***.SendContext(ctx***REMOVED***
	if err != nil {
		response.Diagnostics.AddError(
			"Can't find cluster",
			fmt.Sprintf(
				"Can't find cluster with identifier '%s': %v",
				state.ClusterID.Value, err,
			***REMOVED***,
		***REMOVED***
		return
	}
	object := get.Body(***REMOVED***
	sts, ok := object.AWS(***REMOVED***.GetSTS(***REMOVED***
	if ok {
		accountRolePrefix := DefaultAccountRolePrefix
		if !state.AccountRolePrefix.Unknown && !state.AccountRolePrefix.Null {
			accountRolePrefix = state.AccountRolePrefix.Value
***REMOVED***

		// TODO: use the sts.OperatorRolePrefix(***REMOVED*** if not empty
		// There is a bug in the return value of sts.OperatorRolePrefix(***REMOVED*** - it's always empty string
		for _, operatorRole := range sts.OperatorIAMRoles(***REMOVED*** {
			r := OperatorIAMRole{
				Name: types.String{
					Value: operatorRole.Name(***REMOVED***,
		***REMOVED***,
				Namespace: types.String{
					Value: operatorRole.Namespace(***REMOVED***,
		***REMOVED***,
				RoleARN: types.String{
					Value: operatorRole.RoleARN(***REMOVED***,
		***REMOVED***,
				RoleName: types.String{
					Value: getRoleName(state.OperatorRolePrefix.Value, operatorRole.Namespace(***REMOVED***, operatorRole.Name(***REMOVED******REMOVED***,
		***REMOVED***,
				PolicyName: types.String{
					Value: getRoleName(accountRolePrefix, operatorRole.Namespace(***REMOVED***, operatorRole.Name(***REMOVED******REMOVED***,
		***REMOVED***,
				ServiceAccounts: getServiceAccount(stsOperatorMap[operatorRole.Namespace(***REMOVED***].ServiceAccounts(***REMOVED***, operatorRole.Namespace(***REMOVED******REMOVED***,
	***REMOVED***
			state.OperatorIAMRoles = append(state.OperatorIAMRoles, &r***REMOVED***
***REMOVED***
	}
	// Save the state:
	diags = response.State.Set(ctx, state***REMOVED***
	response.Diagnostics.Append(diags...***REMOVED***
}

// TODO: should be in a separate repo
func getRoleName(prefix string, namespace string, name string***REMOVED*** string {
	roleName := fmt.Sprintf("%s-%s-%s", prefix, namespace, name***REMOVED***
	if len(roleName***REMOVED*** > 64 {
		roleName = roleName[0:64]
	}
	return roleName
}

func getServiceAccount(serviceAccountArr []string, operatorNamespace string***REMOVED*** types.List {
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
