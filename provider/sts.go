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
			Optional:    true,
			Computed:    true,
***REMOVED***,
		"oidc_config_id": {
			Description: "OIDC Configuration ID",
			Type:        types.StringType,
			Optional:    true,
***REMOVED***,
		"thumbprint": {
			Description: "SHA1-hash value of the root CA of the issuer URL",
			Type:        types.StringType,
			Computed:    true,
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
			Description: "Instance IAM Roles",
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
		"operator_role_prefix": {
			Description: "Operator IAM Role prefix",
			Type:        types.StringType,
			Required:    true,
***REMOVED***,
	}***REMOVED***

}
