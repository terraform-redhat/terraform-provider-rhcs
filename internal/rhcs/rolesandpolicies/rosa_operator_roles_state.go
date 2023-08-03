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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
)

type RosaOperatorRolesState struct {
	// Required
	OperatorRolePrefix string `tfsdk:"operator_role_prefix"`

	//Optional
	AccountRolePrefix *string `tfsdk:"account_role_prefix"`

	// Computed
	OperatorIAMRoles []OperatorIAMRole `tfsdk:"operator_iam_roles"`
}

type OperatorIAMRole struct {
	Name            string   `tfsdk:"operator_name"`
	Namespace       string   `tfsdk:"operator_namespace"`
	RoleName        string   `tfsdk:"role_name"`
	PolicyName      string   `tfsdk:"policy_name"`
	ServiceAccounts []string `tfsdk:"service_accounts"`
}

func OperatorRolesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"operator_role_prefix": {
			Description: "Operator role prefix.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"account_role_prefix": {
			Description: "Account role prefix.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"operator_iam_roles": {
			Description: "Operator IAM Roles.",
			Type:        schema.TypeList,
			Elem: &schema.Resource{
				Schema: operatorRolesAttributes(),
			},
			Computed: true,
		},
	}
}

func operatorRolesAttributes() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"operator_name": {
			Description: "Operator Name",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"operator_namespace": {
			Description: "Kubernetes Namespace",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"role_name": {
			Description: "policy name",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"policy_name": {
			Description: "policy name",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"service_accounts": {
			Description: "service accounts",
			Type:        schema.TypeList,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Computed:    true,
		},
	}
}

func operatorRolesStateFromResourceData(resourceData *schema.ResourceData) *RosaOperatorRolesState {
	return &RosaOperatorRolesState{
		OperatorRolePrefix: resourceData.Get("operator_role_prefix").(string),
		AccountRolePrefix:  common.GetOptionalString(resourceData, "account_role_prefix"),
		OperatorIAMRoles:   []OperatorIAMRole{},
	}
}

func operatorRolesStateToResourceData(rosaOperatorRolesState *RosaOperatorRolesState, resourceData *schema.ResourceData) error {
	return resourceData.Set("operator_iam_roles", flatOperatorIAMRole(rosaOperatorRolesState))
}

func flatOperatorIAMRole(rosaOperatorRolesState *RosaOperatorRolesState) []interface{} {
	listOfOperatorIAMRole := make([]interface{}, len(rosaOperatorRolesState.OperatorIAMRoles))
	for i, operatorRole := range rosaOperatorRolesState.OperatorIAMRoles {
		operatorRoleMap := make(map[string]interface{})
		operatorRoleMap["operator_name"] = operatorRole.Name
		operatorRoleMap["operator_namespace"] = operatorRole.Namespace
		operatorRoleMap["role_name"] = operatorRole.RoleName
		operatorRoleMap["policy_name"] = operatorRole.PolicyName
		operatorRoleMap["service_accounts"] = operatorRole.ServiceAccounts
		listOfOperatorIAMRole[i] = operatorRoleMap
	}

	return listOfOperatorIAMRole
}
