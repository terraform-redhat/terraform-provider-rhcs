package hcp

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AutoScaling struct {
	Enabled     types.Bool  `tfsdk:"enabled"`
	MinReplicas types.Int64 `tfsdk:"min_replicas"`
	MaxReplicas types.Int64 `tfsdk:"max_replicas"`
}

func AutoscalingResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{
			Description: "Enables autoscaling. If `true`, this variable requires you to set a maximum and minimum replicas range using the `max_replicas` and `min_replicas` variables.",
			Required:    true,
		},
		"min_replicas": schema.Int64Attribute{
			Description: "The minimum number of replicas for autoscaling functionality.",
			Optional:    true,
		},
		"max_replicas": schema.Int64Attribute{
			Description: "The maximum number of replicas for autoscaling functionality.",
			Optional:    true,
		},
	}
}

func AutoscalingDatasource() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"enabled": schema.BoolAttribute{
			Description: "Enables autoscaling. If `true`, this variable requires you to set a maximum and minimum replicas range using the `max_replicas` and `min_replicas` variables.",
			Required:    true,
		},
		"min_replicas": schema.Int64Attribute{
			Description: "The minimum number of replicas for autoscaling functionality.",
			Optional:    true,
		},
		"max_replicas": schema.Int64Attribute{
			Description: "The maximum number of replicas for autoscaling functionality.",
			Optional:    true,
		},
	}
}
