package clusterrosaclassic

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type privateHZValidator struct {
}

// Description describes the validation in plain text formatting.
func (v privateHZValidator) Description(_ context.Context) string {
	return fmt.Sprintf("proxy map should not include an hard coded OCM proxy")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v privateHZValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v privateHZValidator) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	privateHZ := PrivateHostedZone{}
	d := request.ConfigValue.As(ctx, &privateHZ, basetypes.ObjectAsOptions{})
	if d.HasError() {
		// No attribute to validate
		return
	}
	errSum := "Invalid private_hosted_zone attribute assignment"

	// validate ID and ARN are not empty
	if common.IsStringAttributeEmpty(privateHZ.ID) || common.IsStringAttributeEmpty(privateHZ.RoleARN) {
		response.Diagnostics.AddError(errSum, "Invalid configuration. 'private_hosted_zone.id' and 'private_hosted_zone.arn' are required")
		return
	}

}

func PrivateHZValidator() validator.Object {
	return privateHZValidator{}
}
