// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import "github.com/hashicorp/terraform-plugin-framework/types"

// DNSDomainSeaplaneState holds Terraform state for
// rhcs_dns_domain_seaplane.
type DNSDomainSeaplaneState struct {
	Name       types.String `tfsdk:"name"`
	ID         types.String `tfsdk:"id"`
	BaseDomain types.String `tfsdk:"base_domain"`
	Phase      types.String `tfsdk:"phase"`
}
