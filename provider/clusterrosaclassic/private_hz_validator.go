package clusterrosaclassic

***REMOVED***
	"context"
***REMOVED***
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type privateHZValidator struct {
}

// Description describes the validation in plain text formatting.
func (v privateHZValidator***REMOVED*** Description(_ context.Context***REMOVED*** string {
	return fmt.Sprintf("proxy map should not include an hard coded OCM proxy"***REMOVED***
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v privateHZValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.Description(ctx***REMOVED***
}

// Validate performs the validation.
func (v privateHZValidator***REMOVED*** ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse***REMOVED*** {
	if request.ConfigValue.IsNull(***REMOVED*** || request.ConfigValue.IsUnknown(***REMOVED*** {
		return
	}

	privateHZ := PrivateHostedZone{}
	d := request.ConfigValue.As(ctx, &privateHZ, basetypes.ObjectAsOptions{}***REMOVED***
	if d.HasError(***REMOVED*** {
		// No attribute to validate
		return
	}
	errSum := "Invalid private_hosted_zone attribute assignment"

	// validate ID and ARN are not empty
	if common.IsStringAttributeEmpty(privateHZ.ID***REMOVED*** || common.IsStringAttributeEmpty(privateHZ.RoleARN***REMOVED*** {
		response.Diagnostics.AddError(errSum, "Invalid configuration. 'private_hosted_zone.id' and 'private_hosted_zone.arn' are required"***REMOVED***
		return
	}

}

func PrivateHZValidator(***REMOVED*** validator.Object {
	return privateHZValidator{}
}
