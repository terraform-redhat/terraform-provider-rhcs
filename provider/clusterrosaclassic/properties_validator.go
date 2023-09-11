package clusterrosaclassic

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type propertiesValidator struct {
}

// Description describes the validation in plain text formatting.
func (v propertiesValidator) Description(_ context.Context) string {
	return fmt.Sprintf("properties map should not include an hard coded OCM properties")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v propertiesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v propertiesValidator) ValidateMap(ctx context.Context, request validator.MapRequest, response *validator.MapResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}
	propertiesElements := make(map[string]types.String, len(request.ConfigValue.Elements()))
	d := request.ConfigValue.ElementsAs(ctx, &propertiesElements, false)
	if d.HasError() {
		// No attribute to validate
		return
	}

	for k, _ := range propertiesElements {
		if _, isDefaultKey := OCMProperties[k]; isDefaultKey {
			errHead := "Invalid property key."
			errDesc := fmt.Sprintf("Can not override reserved properties keys. %s is a reserved property key", k)
			response.Diagnostics.AddError(errHead, errDesc)
			return
		}
	}
}

func PropertiesValidator() validator.Map {
	return propertiesValidator{}
}
