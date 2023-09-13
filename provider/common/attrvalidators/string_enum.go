package attrvalidators

***REMOVED***
	"context"
***REMOVED***
	"github.com/thoas/go-funk"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
***REMOVED***

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type enumValueValidator struct {
	allowedList []string
}

// Description describes the validation in plain text formatting.
func (v enumValueValidator***REMOVED*** Description(_ context.Context***REMOVED*** string {
	return fmt.Sprintf("proxy map should not include an hard coded OCM proxy"***REMOVED***
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v enumValueValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.Description(ctx***REMOVED***
}

// Validate performs the validation.
func (v enumValueValidator***REMOVED*** ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse***REMOVED*** {
	if request.ConfigValue.IsNull(***REMOVED*** || request.ConfigValue.IsUnknown(***REMOVED*** {
		return
	}

	value := request.ConfigValue.ValueString(***REMOVED***
	if funk.Contains(v.allowedList, value***REMOVED*** {
		return
	}

	lastStep, _ := request.Path.Steps(***REMOVED***.LastStep(***REMOVED***
	errorLastStep := ""
	if lastStep != nil {
		errorLastStep = fmt.Sprintf("Invalid %s.", lastStep***REMOVED***
	}
	response.Diagnostics.AddError(errorLastStep,
		fmt.Sprintf("Expected a valid param. Options are %s. Got %s.",
			v.allowedList, value***REMOVED***,
	***REMOVED***
}

func EnumValueValidator(enumList []string***REMOVED*** validator.String {
	return enumValueValidator{allowedList: enumList}
}
