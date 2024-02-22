package autoscaler

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceRange struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}
