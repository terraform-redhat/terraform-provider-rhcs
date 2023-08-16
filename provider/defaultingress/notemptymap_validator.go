package defaultingress

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type notEmptyMapValidator struct {
}

// Description describes the validation in plain text formatting.
func (v notEmptyMapValidator) Description(_ context.Context) string {
	return fmt.Sprintf("properties map should not include an hard coded OCM properties")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v notEmptyMapValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v notEmptyMapValidator) ValidateMap(ctx context.Context, request validator.MapRequest, response *validator.MapResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}
	lastStep, _ := request.Path.Steps().LastStep()

	if len(request.ConfigValue.Elements()) == 0 {
		response.Diagnostics.AddError(fmt.Sprintf("Invalid %s.", lastStep),
			fmt.Sprintf("Expected at least one value in map for %s.",
				lastStep),
		)
	}
}

func NotEmptyMapValidator() validator.Map {
	return notEmptyMapValidator{}
}
