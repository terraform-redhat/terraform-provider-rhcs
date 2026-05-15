// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package clusterwaiter

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClusterWaiterState struct {
	Cluster types.String `tfsdk:"cluster"`
	Ready   types.Bool   `tfsdk:"ready"`
	Timeout types.Int64  `tfsdk:"timeout"`
}
