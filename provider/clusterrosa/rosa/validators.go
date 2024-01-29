package rosa

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common/attrvalidators"
)

var AvailabilityZoneValidator = attrvalidators.NewStringValidator("AZ should be valid for cloud_region", func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
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

var PrivateHZValidator = attrvalidators.NewObjectValidator("proxy map should not include an hard coded OCM proxy",
	func(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}

		privateHZ := PrivateHostedZone{}
		d := req.ConfigValue.As(ctx, &privateHZ, basetypes.ObjectAsOptions{})
		if d.HasError() {
			// No attribute to validate
			return
		}
		errSum := "Invalid private_hosted_zone attribute assignment"

		// validate ID and ARN are not empty
		if common.IsStringAttributeKnownAndEmpty(privateHZ.ID) || common.IsStringAttributeKnownAndEmpty(privateHZ.RoleARN) {
			resp.Diagnostics.AddError(errSum, "Invalid configuration. 'private_hosted_zone.id' and 'private_hosted_zone.role_arn' are required")
			return
		}

	})

var PropertiesValidator = attrvalidators.NewMapValidator("properties map should not include an hard coded OCM properties",
	func(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
		if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
			return
		}
		propertiesElements := make(map[string]types.String, len(req.ConfigValue.Elements()))
		d := req.ConfigValue.ElementsAs(ctx, &propertiesElements, false)
		if d.HasError() {
			// No attribute to validate
			return
		}

		for k := range propertiesElements {
			if _, isDefaultKey := OCMProperties[k]; isDefaultKey {
				errHead := "Invalid property key."
				errDesc := fmt.Sprintf("Can not override reserved properties keys. %s is a reserved property key", k)
				resp.Diagnostics.AddError(errHead, errDesc)
				return
			}
		}
	})
