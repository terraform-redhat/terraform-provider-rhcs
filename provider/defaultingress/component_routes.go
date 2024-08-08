package defaultingress

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

type ComponentRoute struct {
	Hostname     types.String `tfsdk:"hostname"`
	TlsSecretRef types.String `tfsdk:"tls_secret_ref"`
}

var ComponentRouteAttributeTypes = map[string]attr.Type{
	"hostname":       types.StringType,
	"tls_secret_ref": types.StringType,
}

func FlattenComponentRoute(hostname, tlsSecretRef string) types.Object {
	if hostname == "" && tlsSecretRef == "" {
		return types.ObjectNull(ComponentRouteAttributeTypes)
	}

	attrs := map[string]attr.Value{
		"hostname":       types.StringValue(hostname),
		"tls_secret_ref": types.StringValue(tlsSecretRef),
	}

	return types.ObjectValueMust(ComponentRouteAttributeTypes, attrs)
}

func ExpandComponentRoute(ctx context.Context,
	object types.Object, diags diag.Diagnostics) (string, string) {
	if object.IsNull() {
		return "", ""
	}

	var componentRoute ComponentRoute
	diags.Append(object.As(ctx, &componentRoute, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return "", ""
	}

	return componentRoute.Hostname.ValueString(), componentRoute.TlsSecretRef.ValueString()
}

func ResetComponentRoutes() map[string]*cmv1.ComponentRouteBuilder {
	resetRoutes := map[string]*cmv1.ComponentRouteBuilder{}
	for _, route := range []string{"oauth", "downloads", "console"} {
		resetRoutes[route] = cmv1.NewComponentRoute().Hostname("").TlsSecretRef("")
	}
	return resetRoutes
}
