// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import "github.com/hashicorp/terraform-plugin-framework/types"

// DNSDomainSeaplaneState holds Terraform state for
// rhcs_dns_domain_seaplane.
type DNSDomainSeaplaneState struct {
	ID          types.String `tfsdk:"id"`
	ClusterArch types.String `tfsdk:"cluster_arch"`
}
