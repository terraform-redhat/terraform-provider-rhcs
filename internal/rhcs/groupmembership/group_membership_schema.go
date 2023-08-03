package groupmembership

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GroupMembershipFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster": {
			Description: "Identifier of the cluster.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"group": {
			Description: "Identifier of the group.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"user": {
			Description: "user name.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

type GroupMembershipState struct {
	// required
	Cluster string `tfsdk:"cluster"`
	Group   string `tfsdk:"group"`
	User    string `tfsdk:"user"`

	// computed
	ID string `tfsdk:"id"`
}
