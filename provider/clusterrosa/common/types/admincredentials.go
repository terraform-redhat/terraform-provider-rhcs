package types

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/terraform-redhat/terraform-provider-rhcs/provider/common"
)

type AdminCredentials struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func FlattenAdminCredentials(username string, password string) types.Object {
	attributeTypes := map[string]attr.Type{
		"username": types.StringType,
		"password": types.StringType,
	}
	if username == "" && password == "" {
		return types.ObjectNull(attributeTypes)
	}

	attrs := map[string]attr.Value{
		"username": types.StringValue(username),
		"password": types.StringValue(password),
	}

	return types.ObjectValueMust(attributeTypes, attrs)
}

func ExpandAdminCredentials(ctx context.Context, object types.Object, diags diag.Diagnostics) (username string, password string) {
	if object.IsNull() {
		return "", ""
	}

	var conf AdminCredentials
	diags.Append(object.As(ctx, &conf, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return "", ""
	}

	return conf.Username.ValueString(), conf.Password.ValueString()
}

func AdminCredentialsNull() types.Object {
	return FlattenAdminCredentials("", "")
}

func AdminCredentialsEqual(state, plan types.Object) bool {
	if state.IsNull() {
		if common.HasValue(plan) {
			return false
		}
		return true
	}
	return reflect.DeepEqual(state, plan)
}
