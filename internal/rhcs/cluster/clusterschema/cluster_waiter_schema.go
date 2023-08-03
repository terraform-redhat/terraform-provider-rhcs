package clusterschema

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-redhat/terraform-provider-rhcs/internal/rhcs/common"
)

const (
	DefaultTimeoutInMinutes = 60
)

func ClusterWaiterFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster": {
			Description: "Identifier of the cluster.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"timeout": {
			Description: "An optional timeout till the cluster is ready. The timeout value should be in minutes." +
				" the default value is 60 minutes",
			Type:             schema.TypeInt,
			Optional:         true,
			ValidateDiagFunc: timeoutValidators,
		},
		"ready": {
			Description: "Whether the cluster is ready",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}
}

type ClusterWaiterState struct {
	// Required
	Cluster string `tfsdk:"cluster"`

	// Optional
	Timeout int64 `tfsdk:"timeout"`

	// Computed
	Ready bool `tfsdk:"ready"`
}

func ExpandClusterWaiterFromResourceData(resourceData *schema.ResourceData) *ClusterWaiterState {
	result := &ClusterWaiterState{
		Cluster: resourceData.Get("cluster").(string),
	}

	timeout := common.GetOptionalInt(resourceData, "timeout")
	if timeout != nil {
		result.Timeout = int64(*timeout)
	}

	ready := common.GetOptionalBool(resourceData, "ready")
	if ready != nil {
		result.Ready = *ready
	} else {
		result.Ready = false
	}

	return result
}

func timeoutValidators(i interface{}, path cty.Path) diag.Diagnostics {
	timeout, ok := i.(int)
	if !ok {
		return diag.Errorf("expected type to be int")
	}
	if timeout <= 0 {
		return diag.Errorf("Invalid timeout configuration. timeout must be positive")
	}

	return nil
}
