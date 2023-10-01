package clusterrosaclassic

***REMOVED***
	"context"
***REMOVED***
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
***REMOVED***

var availabilityZoneValidator = attrvalidators.NewStringValidator("AZ should be valid for cloud_region", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse***REMOVED*** {
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

var privateHZValidator = attrvalidators.NewObjectValidator("proxy map should not include an hard coded OCM proxy",
	func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse***REMOVED*** {
		if req.ConfigValue.IsNull(***REMOVED*** || req.ConfigValue.IsUnknown(***REMOVED*** {
			return
***REMOVED***

		privateHZ := PrivateHostedZone{}
		d := req.ConfigValue.As(ctx, &privateHZ, basetypes.ObjectAsOptions{}***REMOVED***
		if d.HasError(***REMOVED*** {
			// No attribute to validate
			return
***REMOVED***
		errSum := "Invalid private_hosted_zone attribute assignment"

		// validate ID and ARN are not empty
		if common.IsStringAttributeEmpty(privateHZ.ID***REMOVED*** || common.IsStringAttributeEmpty(privateHZ.RoleARN***REMOVED*** {
			resp.Diagnostics.AddError(errSum, "Invalid configuration. 'private_hosted_zone.id' and 'private_hosted_zone.arn' are required"***REMOVED***
			return
***REMOVED***

	}***REMOVED***

var propertiesValidator = attrvalidators.NewMapValidator("properties map should not include an hard coded OCM properties",
	func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse***REMOVED*** {
		if req.ConfigValue.IsNull(***REMOVED*** || req.ConfigValue.IsUnknown(***REMOVED*** {
			return
***REMOVED***
		propertiesElements := make(map[string]types.String, len(req.ConfigValue.Elements(***REMOVED******REMOVED******REMOVED***
		d := req.ConfigValue.ElementsAs(ctx, &propertiesElements, false***REMOVED***
		if d.HasError(***REMOVED*** {
			// No attribute to validate
			return
***REMOVED***

		for k, _ := range propertiesElements {
			if _, isDefaultKey := OCMProperties[k]; isDefaultKey {
				errHead := "Invalid property key."
				errDesc := fmt.Sprintf("Can not override reserved properties keys. %s is a reserved property key", k***REMOVED***
				resp.Diagnostics.AddError(errHead, errDesc***REMOVED***
				return
	***REMOVED***
***REMOVED***
	}***REMOVED***
