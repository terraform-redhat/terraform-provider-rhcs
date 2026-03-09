package imagemirror

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ImageMirrorState struct {
	ClusterID           types.String `tfsdk:"cluster_id"`
	Type                types.String `tfsdk:"type"`
	Source              types.String `tfsdk:"source"`
	Mirrors             types.List   `tfsdk:"mirrors"`
	ID                  types.String `tfsdk:"id"`
	CreationTimestamp   types.String `tfsdk:"creation_timestamp"`
	LastUpdateTimestamp types.String `tfsdk:"last_update_timestamp"`
}
