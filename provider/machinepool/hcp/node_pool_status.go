package hcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.UseStateForUnknown(),
			},
		},
		"message": schema.StringAttribute{
			Description: "Message regarding status of the replica",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
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

func flattenNodePoolStatus(currentReplicas int64, message string) types.Object {
	attributeTypes := map[string]attr.Type{
		"current_replicas": types.Int64Type,
		"message":          types.StringType,
	}
	if currentReplicas == 0 && message == "" {
		return types.ObjectNull(attributeTypes)
	}

	attrs := map[string]attr.Value{
		"current_replicas": types.Int64Value(currentReplicas),
		"message":          types.StringValue(message),
	}

	return types.ObjectValueMust(attributeTypes, attrs)
}

func expandNodePoolStatus(ctx context.Context,
	object types.Object, diags diag.Diagnostics) (currentReplicas int64, message string) {
	if object.IsNull() {
		return 0, ""
	}

	var status NodePoolStatus
	diags.Append(object.As(ctx, &status, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return 0, ""
	}

	return status.CurrentReplicas.ValueInt64(), status.Message.ValueString()
}

func nodePoolStatusNull() types.Object {
	return flattenNodePoolStatus(0, "")
}
