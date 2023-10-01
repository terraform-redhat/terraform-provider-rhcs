package clusterrosaclassic

***REMOVED***
	"context"
***REMOVED***
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
***REMOVED***

var availabilityZoneValidator = attrvalidators.NewStringValidator("", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
	az := req.ConfigValue.ValueString(***REMOVED***
	regionAttr := basetypes.StringValue{}
	err := req.Config.GetAttribute(ctx, path.Root("cloud_region"***REMOVED***, &regionAttr***REMOVED***
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid AZ", "Failed to fetch cloud_region"***REMOVED***
		return
	}
	region := regionAttr.ValueString(***REMOVED***
	if !strings.Contains(az, region***REMOVED*** {
		msg := fmt.Sprintf("Invalid AZ '%s' for region '%s'.", az, region***REMOVED***
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid AZ", msg***REMOVED***
		return
	}
}***REMOVED***
