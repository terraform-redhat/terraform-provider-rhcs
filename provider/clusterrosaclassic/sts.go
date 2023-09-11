package clusterrosaclassic

***REMOVED***
	tfrschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
***REMOVED***

func stsResource(***REMOVED*** map[string]tfrschema.Attribute {
	return map[string]tfrschema.Attribute{
		"oidc_endpoint_url": tfrschema.StringAttribute{
			Description: "OIDC Endpoint URL",
			Optional:    true,
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				// This passes the state through to the plan, so that the OIDC
				// provider URL will not be "unknown" at plan-time, resulting in
				// the oidc provider being needlessly replaced. This should be
				// OK, since the OIDC provider URL is not expected to change.
				stringplanmodifier.UseStateForUnknown(***REMOVED***,
	***REMOVED***,
***REMOVED***,
		"oidc_config_id": tfrschema.StringAttribute{
			Description: "OIDC Configuration ID",
			Optional:    true,
***REMOVED***,
		"thumbprint": tfrschema.StringAttribute{
			Description: "SHA1-hash value of the root CA of the issuer URL",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				// This passes the state through to the plan, so that the OIDC
				// thumbprint will not be "unknown" at plan-time, resulting in
				// the oidc provider being needlessly replaced. This should be
				// OK, since the thumbprint is not expected to change.
				stringplanmodifier.UseStateForUnknown(***REMOVED***,
	***REMOVED***,
***REMOVED***,
		"role_arn": tfrschema.StringAttribute{
			Description: "Installer Role",
			Required:    true,
***REMOVED***,
		"support_role_arn": tfrschema.StringAttribute{
			Description: "Support Role",
			Required:    true,
***REMOVED***,
		"instance_iam_roles": tfrschema.SingleNestedAttribute{
			Description: "Instance IAM Roles",
			Attributes: map[string]tfrschema.Attribute{
				"master_role_arn": tfrschema.StringAttribute{
					Description: "Master/Controller Plane Role ARN",
					Required:    true,
		***REMOVED***,
				"worker_role_arn": tfrschema.StringAttribute{
					Description: "Worker Node Role ARN",
					Required:    true,
		***REMOVED***,
	***REMOVED***,
			Required: true,
***REMOVED***,
		"operator_role_prefix": tfrschema.StringAttribute{
			Description: "Operator IAM Role prefix",
			Required:    true,
***REMOVED***,
	}
}
