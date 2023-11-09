package attrvalidators

import (
	"context"
	"fmt"
	"github.com/thoas/go-funk"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type enumValueValidator struct {
	allowedList []string
}

// Description describes the validation in plain text formatting.
func (v enumValueValidator) Description(_ context.Context) string {
	return fmt.Sprintf("proxy map should not include an hard coded OCM proxy")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v enumValueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v enumValueValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()
	if funk.Contains(v.allowedList, value) {
		return
	}

	lastStep, _ := request.Path.Steps().LastStep()
	errorLastStep := ""
	if lastStep != nil {
		errorLastStep = fmt.Sprintf("Invalid %s.", lastStep)
	}
	response.Diagnostics.AddError(errorLastStep,
		fmt.Sprintf("Expected a valid param. Options are %s. Got %s.",
			v.allowedList, value),
	)
}

func EnumValueValidator(enumList []string) validator.String {
	return enumValueValidator{allowedList: enumList}
}
