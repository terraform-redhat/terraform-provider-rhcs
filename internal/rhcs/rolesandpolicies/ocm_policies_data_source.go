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

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

const DataSourceID = "ocm-policies"

func OcmPoliciesDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: ocmPoliciesDataSourceRead,
		Schema:      ocmPoliciesSchema(),
	}
}

func ocmPoliciesDataSourceRead(ctx context.Context, resourceData *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	// Get the collection of AWSInquiries:
	awsInquiries := meta.(*sdk.Connection).ClustersMgmt().V1().AWSInquiries()

	policiesResponse, err := awsInquiries.STSPolicies().List().Send()
	if err != nil {
		tflog.Error(ctx, "Failed to get policies")
		return []diag.Diagnostic{
			{
				Severity: diag.Error,
				Summary:  "Failed to get policies",
				Detail:   err.Error(),
			}}
	}

	operatorRolePolicies := OperatorRolePolicies{}
	accountRolePolicies := AccountRolePolicies{}
	policiesResponse.Items().Each(func(awsPolicy *cmv1.AWSSTSPolicy) bool {
		tflog.Debug(ctx, fmt.Sprintf("policy id: %s ", awsPolicy.ID()))
		switch awsPolicy.ID() {
		// operator roles
		case CloudCred:
			operatorRolePolicies.CloudCred = awsPolicy.Details()
		case CloudNetwork:
			operatorRolePolicies.CloudNetwork = awsPolicy.Details()
		case ClusterCSI:
			operatorRolePolicies.ClusterCSI = awsPolicy.Details()
		case ImageRegistry:
			operatorRolePolicies.ImageRegistry = awsPolicy.Details()
		case IngressOperator:
			operatorRolePolicies.IngressOperator = awsPolicy.Details()
		case MachineAPI:
			operatorRolePolicies.MachineAPI = awsPolicy.Details()
		// account roles
		case Installer:
			accountRolePolicies.Installer = awsPolicy.Details()
		case Support:
			accountRolePolicies.Support = awsPolicy.Details()
		case InstanceWorker:
			accountRolePolicies.InstanceWorker = awsPolicy.Details()
		case InstanceControlPlane:
			accountRolePolicies.InstanceControlPlane = awsPolicy.Details()
		default:
			tflog.Debug(ctx, "This is neither operator role policy nor account role policy")
		}
		return true
	})

	resourceData.SetId(DataSourceID)

	resourceData.Set("operator_role_policies", FlatOperatorRolePolicies(operatorRolePolicies))
	resourceData.Set("account_role_policies", FlatAccountRolePolicies(accountRolePolicies))

	return
}
