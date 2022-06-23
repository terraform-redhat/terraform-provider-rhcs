package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stsResource() tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"oidc_endpoint_url": {
			Description: "OIDC Endpoint URL",
			Type:        types.StringType,
			Computed:    true,
			Optional:    true,
		},
		"role_arn": {
			Description: "Installer Role",
			Type:        types.StringType,
			Required:    true,
		},
		"support_role_arn": {
			Description: "Support Role",
			Type:        types.StringType,
			Required:    true,
		},
		"instance_iam_roles": {
			Description: "Instance IAm Roles",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"master_role_arn": {
					Description: "Master/Controller Plane Role ARN",
					Type:        types.StringType,
					Required:    true,
				},
				"worker_role_arn": {
					Description: "Worker Node Role ARN",
					Type:        types.StringType,
					Required:    true,
				},
			}),
			Required: true,
		},
		"operator_iam_roles": {
			Description: "Operator IAM Roles",
			Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
				"name": {
					Description: "Operator Name",
					Type:        types.StringType,
					Required:    true,
				},
				"namespace": {
					Description: "Kubernetes Namespace",
					Type:        types.StringType,
					Required:    true,
				},
				"role_arn": {
					Description: "AWS Role ARN",
					Type:        types.StringType,
					Required:    true,
				},
			}, tfsdk.ListNestedAttributesOptions{
				MinItems: 6,
				MaxItems: 6}),
			// Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			// 	"cloud_credential": {
			// 		Description: "Cloud Credential ARN",
			// 		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			// 			"name": {
			// 				Description: "Name of Cloud Credential Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"namespace": {
			// 				Description: "Namespace of Cloud Credential Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"role_arn": {
			// 				Description: "Name of Cloud Credential Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 		}),
			// 		Required: true,
			// 	},
			// 	"image_registry": {
			// 		Description: "Image Registry ARN",
			// 		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			// 			"name": {
			// 				Description: "Name of Image Registry Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"namespace": {
			// 				Description: "Namespace of Image Registry Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"role_arn": {
			// 				Description: "Name of Image Registry Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 		}),
			// 		Required: true,
			// 	},
			// 	"ingress": {
			// 		Description: "Ingress ARN",
			// 		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			// 			"name": {
			// 				Description: "Ingress Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"namespace": {
			// 				Description: "Namespace of Ingress Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"role_arn": {
			// 				Description: "Role ARN of Ingress Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 		}),
			// 		Required: true,
			// 	},
			// 	"ebs": {
			// 		Description: "EBS ARN",
			// 		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			// 			"name": {
			// 				Description: "EBS Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"namespace": {
			// 				Description: "Namespace of EBS Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"role_arn": {
			// 				Description: "Role ARN of EBS Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 		}),
			// 		Required: true,
			// 	},
			// 	"cloud_network_config": {
			// 		Description: "Cloud Network Config ARN",
			// 		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			// 			"name": {
			// 				Description: "Cloud Network Config Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"namespace": {
			// 				Description: "Namespace of Cloud Network Config Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"role_arn": {
			// 				Description: "Role ARN of Cloud Network Config Operator role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 		}),
			// 		Required: true,
			// 	},
			// 	"machine_api": {
			// 		Description: "Machine API ARN",
			// 		Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
			// 			"name": {
			// 				Description: "Machine API role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"namespace": {
			// 				Description: "Namespace of Machine API role role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 			"role_arn": {
			// 				Description: "Role ARN of Machine API role",
			// 				Type:        types.StringType,
			// 				Required:    true,
			// 			},
			// 		}),
			// 		Required: true,
			// 	},
			// }),
			Required: true,
		},
	})
}
