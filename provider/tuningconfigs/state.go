package tuningconfigs

import "github.com/hashicorp/terraform-plugin-framework/types"

type TuningConfig struct {
	Id      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Cluster types.String `tfsdk:"cluster"`
	Spec    types.String `tfsdk:"spec"`
}
