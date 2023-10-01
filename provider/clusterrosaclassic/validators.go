package clusterrosaclassic

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

var availabilityZoneValidator = attrvalidators.NewStringValidator("", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	az := req.ConfigValue.ValueString()
	regionAttr := basetypes.StringValue{}
	err := req.Config.GetAttribute(ctx, path.Root("cloud_region"), &regionAttr)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid AZ", "Failed to fetch cloud_region")
		return
	}
	region := regionAttr.ValueString()
	if !strings.Contains(az, region) {
		msg := fmt.Sprintf("Invalid AZ '%s' for region '%s'.", az, region)
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid AZ", msg)
		return
	}
})
