package hcp

import "github.com/hashicorp/terraform-plugin-framework/types"

const (
	autoNodeModeEnabled = "enabled"
)

type AutoNode struct {
	Mode    types.String `tfsdk:"mode"`
	RoleARN types.String `tfsdk:"role_arn"`
}

func autoNodeMode(autoNode *AutoNode) types.String {
	if autoNode == nil {
		return types.StringNull()
	}

	return autoNode.Mode
}

func autoNodeRoleARN(autoNode *AutoNode) types.String {
	if autoNode == nil {
		return types.StringNull()
	}

	return autoNode.RoleARN
}
