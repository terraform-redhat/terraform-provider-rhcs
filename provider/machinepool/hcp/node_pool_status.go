package hcp

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NodePoolStatus struct {
	CurrentReplicas types.Int64  `tfsdk:"current_replicas"`
	Message         types.String `tfsdk:"message"`
}

func NodePoolStatusResource() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"current_replicas": schema.Int64Attribute{
			Description: "The current number of replicas.",
			Computed:    true,
		},
		"message": schema.StringAttribute{
			Description: "Message regarding status of the replica",
			Computed:    true,
		},
	}
}

func NodePoolStatusDatasource() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"current_replicas": schema.Int64Attribute{
			Description: "The current number of replicas.",
			Computed:    true,
		},
		"message": schema.StringAttribute{
			Description: "Message regarding status of the replica",
			Computed:    true,
		},
	}
}
