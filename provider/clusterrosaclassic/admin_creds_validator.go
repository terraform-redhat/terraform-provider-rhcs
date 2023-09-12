package clusterrosaclassic

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type adminCredsValidator struct {
}

// Description describes the validation in plain text formatting.
func (v adminCredsValidator) Description(_ context.Context) string {
	return fmt.Sprintf("admin creds validators")
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v adminCredsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v adminCredsValidator) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	var creds *AdminCredentials
	d := request.ConfigValue.As(ctx, &creds, basetypes.ObjectAsOptions{})
	if d.HasError() {
		// No attribute to validate
		return
	}
	errSum := "Invalid admin_creedntials"
	if creds == nil {
		return
	}
	if common.IsStringAttributeEmpty(creds.Username) {
		response.Diagnostics.AddError(errSum, "Usename can't be empty")
		return
	}
	if err := common.ValidateHTPasswdUsername(creds.Username.ValueString()); err != nil {
		response.Diagnostics.AddError(errSum, err.Error())
		return
	}
	if common.IsStringAttributeEmpty(creds.Password) {
		response.Diagnostics.AddError(errSum, "Usename can't be empty")
		return
	}
	if err := common.ValidateHTPasswdPassword(creds.Password.ValueString()); err != nil {
		response.Diagnostics.AddError(errSum, err.Error())
		return
	}

}

func AdminCredsValidator() validator.Object {
	return adminCredsValidator{}
}
