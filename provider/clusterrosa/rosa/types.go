package rosa

import "github.com/hashicorp/terraform-plugin-framework/types"

type AdminCredentials struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

type PrivateHostedZone struct {
	ID      types.String `tfsdk:"id"`
	RoleARN types.String `tfsdk:"role_arn"`
}
