package ocmrolelink

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OCMRoleLinksState struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	RoleArns       types.List   `tfsdk:"role_arns"`
}
