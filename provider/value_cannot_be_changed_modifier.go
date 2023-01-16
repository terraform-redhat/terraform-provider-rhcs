package provider

***REMOVED***
	"context"
	"errors"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/openshift-online/ocm-sdk-go/logging"
***REMOVED***

type valueCannotBeChangedModifier struct {
	logger logging.Logger
}

func ValueCannotBeChangedModifier(logger logging.Logger***REMOVED*** tfsdk.AttributePlanModifier {
	return valueCannotBeChangedModifier{
		logger: logger,
	}
}
func (m valueCannotBeChangedModifier***REMOVED*** Description(ctx context.Context***REMOVED*** string {
	return "The value cannot be changed after the resource was created."
}

func (m valueCannotBeChangedModifier***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return m.Description(ctx***REMOVED***
}

func (m valueCannotBeChangedModifier***REMOVED*** Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse***REMOVED*** {

	if req.AttributeConfig == nil || req.AttributeState == nil || req.AttributePlan == nil {
		// shouldn't happen, but let's not panic if it does
		return
	}

	if req.State.Raw.IsNull(***REMOVED*** {
		// if we're creating the resource, no need to delete and
		// recreate it
		return
	}

	if req.Plan.Raw.IsNull(***REMOVED*** {
		// if we're deleting the resource, no need to delete and
		// recreate it
		return
	}

	attrSchema, err := req.State.Schema.AttributeAtPath(req.AttributePath***REMOVED***
	if err != nil && !errors.Is(err, errors.New("path leads to block, not an attribute"***REMOVED******REMOVED*** {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error finding attribute schema",
			fmt.Sprintf("An unexpected error was encountered retrieving the schema for this attribute. This is always a bug in the provider.\n\nError: %s", err***REMOVED***,
		***REMOVED***
		return
	}

	configRaw, err := req.AttributeConfig.ToTerraformValue(ctx***REMOVED***
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath,
			"Error converting config value",
			fmt.Sprintf("An unexpected error was encountered converting a %s to its equivalent Terraform representation. This is always a bug in the provider.\n\nError: %s", req.AttributeConfig.Type(ctx***REMOVED***, err***REMOVED***,
		***REMOVED***
		return
	}

	if configRaw == nil && attrSchema.Computed {
		// if the config is null and the attribute is computed, this
		// could be an out-of-band change, don't require blocking
		return
	}

	if req.AttributeState.Equal(req.AttributePlan***REMOVED*** {
		m.logger.Debug(ctx, "attribute state and attribute plan have the same value"***REMOVED***
		return
	}

	// the attribute value was changes
	m.logger.Debug(ctx, "attribute plan was changed"***REMOVED***
	resp.Diagnostics.AddAttributeError(req.AttributePath, "Value cannot be changed", "This attribute is blocked for updating"***REMOVED***
	return

}
