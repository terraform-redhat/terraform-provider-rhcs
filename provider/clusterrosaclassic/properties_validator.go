package clusterrosaclassic

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
***REMOVED***

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type propertiesValidator struct {
}

// Description describes the validation in plain text formatting.
func (v propertiesValidator***REMOVED*** Description(_ context.Context***REMOVED*** string {
	return fmt.Sprintf("properties map should not include an hard coded OCM properties"***REMOVED***
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v propertiesValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.Description(ctx***REMOVED***
}

// Validate performs the validation.
func (v propertiesValidator***REMOVED*** ValidateMap(ctx context.Context, request validator.MapRequest, response *validator.MapResponse***REMOVED*** {
	if request.ConfigValue.IsNull(***REMOVED*** || request.ConfigValue.IsUnknown(***REMOVED*** {
		return
	}
	propertiesElements := make(map[string]types.String, len(request.ConfigValue.Elements(***REMOVED******REMOVED******REMOVED***
	d := request.ConfigValue.ElementsAs(ctx, &propertiesElements, false***REMOVED***
	if d.HasError(***REMOVED*** {
		// No attribute to validate
		return
	}

	for k, _ := range propertiesElements {
		if _, isDefaultKey := OCMProperties[k]; isDefaultKey {
			errHead := "Invalid property key."
			errDesc := fmt.Sprintf("Can not override reserved properties keys. %s is a reserved property key", k***REMOVED***
			response.Diagnostics.AddError(errHead, errDesc***REMOVED***
			return
***REMOVED***
	}
}

func PropertiesValidator(***REMOVED*** validator.Map {
	return propertiesValidator{}
}
