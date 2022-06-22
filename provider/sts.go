package provider

***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
***REMOVED***

func stsResource(***REMOVED*** tfsdk.NestedAttributes {
	return tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"oidc_endpoint_url": {
			Description: "OIDC Endpoint URL",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"role_arn": {
			Description: "Installer Role",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"support_role_arn": {
			Description: "Support Role",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
		"instance_iam_roles": {
			Description: "Instance IAm Roles",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"master_role_arn": {
					Description: "Master/Controller Plane Role ARN",
					Type:        types.StringType,
					Required:    true,
		***REMOVED***,
				"worker_role_arn": {
					Description: "Worker Node Role ARN",
					Type:        types.StringType,
					Required:    true,
		***REMOVED***,
	***REMOVED******REMOVED***,
			Required: true,
***REMOVED***,
		"operator_iam_roles": {
			Description: "Operator IAM Roles",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"cloud_credential": {
					Description: "Cloud Credential ARN",
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"name": {
							Description: "Name of Cloud Credential Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"namespace": {
							Description: "Namespace of Cloud Credential Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"role_arn": {
							Description: "Name of Cloud Credential Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
			***REMOVED******REMOVED***,
					Required: true,
		***REMOVED***,
				"image_registry": {
					Description: "Image Registry ARN",
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"name": {
							Description: "Name of Image Registry Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"namespace": {
							Description: "Namespace of Image Registry Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"role_arn": {
							Description: "Name of Image Registry Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
			***REMOVED******REMOVED***,
					Required: true,
		***REMOVED***,
				"ingress": {
					Description: "Ingress ARN",
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"name": {
							Description: "Ingress Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"namespace": {
							Description: "Namespace of Ingress Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"role_arn": {
							Description: "Role ARN of Ingress Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
			***REMOVED******REMOVED***,
					Required: true,
		***REMOVED***,
				"ebs": {
					Description: "EBS ARN",
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"name": {
							Description: "EBS Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"namespace": {
							Description: "Namespace of EBS Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"role_arn": {
							Description: "Role ARN of EBS Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
			***REMOVED******REMOVED***,
					Required: true,
		***REMOVED***,
				"cloud_network_config": {
					Description: "Cloud Network Config ARN",
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"name": {
							Description: "Cloud Network Config Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"namespace": {
							Description: "Namespace of Cloud Network Config Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"role_arn": {
							Description: "Role ARN of Cloud Network Config Operator role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
			***REMOVED******REMOVED***,
					Required: true,
		***REMOVED***,
				"machine_api": {
					Description: "Machine API ARN",
					Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
						"name": {
							Description: "Machine API role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"namespace": {
							Description: "Namespace of Machine API role role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
						"role_arn": {
							Description: "Role ARN of Machine API role",
							Type:        types.StringType,
							Required:    true,
				***REMOVED***,
			***REMOVED******REMOVED***,
					Required: true,
		***REMOVED***,
	***REMOVED******REMOVED***,
			Required: true,
***REMOVED***,
	}***REMOVED***
}
