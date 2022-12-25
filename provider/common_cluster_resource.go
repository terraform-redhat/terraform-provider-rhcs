package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func commonAttributesForClusterResources() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Description: "Unique identifier of the cluster.",
			Type:        types.StringType,
			Computed:    true,
		},
		"name": {
			Description: "Name of the cluster.",
			Type:        types.StringType,
			Required:    true,
		},
		"cloud_region": {
			Description: "Cloud region identifier, for example 'us-east-1'.",
			Type:        types.StringType,
			Required:    true,
		},
		"multi_az": {
			Description: "Indicates if the cluster should be deployed to " +
				"multiple availability zones. Default value is 'false'.",
			Type:     types.BoolType,
			Optional: true,
			Computed: true,
			PlanModifiers: []tfsdk.AttributePlanModifier{
				tfsdk.RequiresReplace(),
			},
		},
		"properties": {
			Description: "User defined properties.",
			Type: types.MapType{
				ElemType: types.StringType,
			},
			Optional: true,
			Computed: true,
		},
		"api_url": {
			Description: "URL of the API server.",
			Type:        types.StringType,
			Computed:    true,
		},
		"console_url": {
			Description: "URL of the console.",
			Type:        types.StringType,
			Computed:    true,
		},
		"compute_nodes": {
			Description: "Number of compute nodes of the cluster.",
			Type:        types.Int64Type,
			Optional:    true,
			Computed:    true,
		},
		"compute_machine_type": {
			Description: "Identifier of the machine type used by the compute nodes, " +
				"for example `r5.xlarge`. Use the `ocm_machine_types` data " +
				"source to find the possible values.",
			Type:     types.StringType,
			Optional: true,
			Computed: true,
			PlanModifiers: []tfsdk.AttributePlanModifier{
				tfsdk.RequiresReplace(),
			},
		},
		"aws_subnet_ids": {
			Description: "aws subnet ids",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"aws_private_link": {
			Description: "aws subnet ids",
			Type:        types.BoolType,
			Optional:    true,
			Computed:    true,
			PlanModifiers: []tfsdk.AttributePlanModifier{
				tfsdk.RequiresReplace(),
			},
		},
		"availability_zones": {
			Description: "availability zones",
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Optional: true,
		},
		"machine_cidr": {
			Description: "Block of IP addresses for nodes.",
			Type:        types.StringType,
			Optional:    true,
			Computed:    true,
		},
		"proxy": {
			Description: "proxy",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"http_proxy": {
					Description: "http proxy",
					Type:        types.StringType,
					Required:    true,
				},
				"https_proxy": {
					Description: "https proxy",
					Type:        types.StringType,
					Required:    true,
				},
				"no_proxy": {
					Description: "no proxy",
					Type:        types.StringType,
					Optional:    true,
				},
			}),
			Optional: true,
		},
		"service_cidr": {
			Description: "Block of IP addresses for services.",
			Type:        types.StringType,
			Optional:    true,
			Computed:    true,
		},
		"pod_cidr": {
			Description: "Block of IP addresses for pods.",
			Type:        types.StringType,
			Optional:    true,
			Computed:    true,
		},
		"host_prefix": {
			Description: "Length of the prefix of the subnet assigned to each node.",
			Type:        types.Int64Type,
			Optional:    true,
			Computed:    true,
		},
		"version": {
			Description: "Identifier of the version of OpenShift, for example 'openshift-v4.1.0'.",
			Type:        types.StringType,
			Optional:    true,
			Computed:    true,
		},
		"state": {
			Description: "State of the cluster.",
			Type:        types.StringType,
			Computed:    true,
		},
	}
}
