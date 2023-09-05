package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClusterHCPWaiterState struct {
	Cluster types.String `tfsdk:"cluster"`
	Ready   types.Bool   `tfsdk:"ready"`
	Timeout types.Int64  `tfsdk:"timeout"`
}
