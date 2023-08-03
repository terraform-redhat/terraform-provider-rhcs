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
	"context"
	"fmt"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
	"sort"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const (
	DefaultAccountRolePrefix = "ManagedOpenShift"
	serviceAccountFmt        = "system:serviceaccount:%s:%s"
)

func OperatorRolesDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: operatorRolesDataSourceRead,
		Schema:      OperatorRolesSchema(),
	}
}

func operatorRolesDataSourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	// Get the collection of AWSInquiries:
	awsInquiries := meta.(*sdk.Connection).ClustersMgmt().V1().AWSInquiries()

	operatorRolesState := operatorRolesStateFromResourceData(resourceData)
	stsOperatorRolesList, err := awsInquiries.STSCredentialRequests().List().Send()
	if err != nil {
		description := fmt.Sprintf("Failed to get STS Operator Roles list with error: %v", err)
		tflog.Error(ctx, description)
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  description,
				Detail:   "hint: validate the credetials (token) used to run this clusterservice",
			}}
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
	if !common.IsStringAttributeEmpty(operatorRolesState.AccountRolePrefix) {
		accountRolePrefix = *operatorRolesState.AccountRolePrefix
	}

	// TODO: use the sts.OperatorRolePrefix() if not empty
	// There is a bug in the return value of sts.OperatorRolePrefix() - it's always empty string
	sort.Strings(roleNameSpaces)
	for _, key := range roleNameSpaces {

		r := OperatorIAMRole{
			Name:            stsOperatorMap[key].Name(),
			Namespace:       stsOperatorMap[key].Namespace(),
			RoleName:        getRoleName(operatorRolesState.OperatorRolePrefix, stsOperatorMap[key]),
			PolicyName:      getPolicyName(accountRolePrefix, stsOperatorMap[key].Namespace(), stsOperatorMap[key].Name()),
			ServiceAccounts: buildServiceAccountsArray(stsOperatorMap[stsOperatorMap[key].Namespace()].ServiceAccounts(), stsOperatorMap[key].Namespace()),
		}
		operatorRolesState.OperatorIAMRoles = append(operatorRolesState.OperatorIAMRoles, r)
	}

	resourceData.SetId(operatorRolesState.OperatorRolePrefix)

	err = operatorRolesStateToResourceData(operatorRolesState, resourceData)
	if err != nil {
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "failed to save the state in  resourceData",
				Detail:   err.Error(),
			}}
	}

	return
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

func buildServiceAccountsArray(serviceAccountArr []string, operatorNamespace string) []string {
	serviceAccounts := []string{}

	for _, v := range serviceAccountArr {
		serviceAccount := fmt.Sprintf(serviceAccountFmt, operatorNamespace, v)
		serviceAccounts = append(serviceAccounts, serviceAccount)
	}

	return serviceAccounts
}
