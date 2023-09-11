package clusterrosaclassic

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
***REMOVED***

// atLeastValidator validates that an integer Attribute's value is at least a certain value.
type adminCredsValidator struct {
}

// Description describes the validation in plain text formatting.
func (v adminCredsValidator***REMOVED*** Description(_ context.Context***REMOVED*** string {
	return fmt.Sprintf("proxy map should not include an hard coded OCM proxy"***REMOVED***
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v adminCredsValidator***REMOVED*** MarkdownDescription(ctx context.Context***REMOVED*** string {
	return v.Description(ctx***REMOVED***
}

// Validate performs the validation.
func (v adminCredsValidator***REMOVED*** ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse***REMOVED*** {
	if request.ConfigValue.IsNull(***REMOVED*** || request.ConfigValue.IsUnknown(***REMOVED*** {
		return
	}

	var creds *AdminCredentials
	d := request.ConfigValue.As(ctx, &creds, basetypes.ObjectAsOptions{}***REMOVED***
	if d.HasError(***REMOVED*** {
		// No attribute to validate
		return
	}
	errSum := "Invalid admin_creedntials"
	if creds == nil {
		return
	}
	if common.IsStringAttributeEmpty(creds.Username***REMOVED*** {
		response.Diagnostics.AddError(errSum, "Usename can't be empty"***REMOVED***
		return
	}
	if err := common.ValidateHTPasswdUsername(creds.Username.ValueString(***REMOVED******REMOVED***; err != nil {
		response.Diagnostics.AddError(errSum, err.Error(***REMOVED******REMOVED***
		return
	}
	if common.IsStringAttributeEmpty(creds.Password***REMOVED*** {
		response.Diagnostics.AddError(errSum, "Usename can't be empty"***REMOVED***
		return
	}
	if err := common.ValidateHTPasswdPassword(creds.Password.ValueString(***REMOVED******REMOVED***; err != nil {
		response.Diagnostics.AddError(errSum, err.Error(***REMOVED******REMOVED***
		return
	}

}

func AdminCredsValidator(***REMOVED*** validator.Object {
	return adminCredsValidator{}
}
