package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func OptionalInt64(tfVal types.Int64) *int64 {
	if !tfVal.IsNull() && !tfVal.IsUnknown() {
		return pointer(tfVal.ValueInt64())
	}
	return nil
}

func OptionalBool(tfVal types.Bool) *bool {
	if !tfVal.IsNull() && !tfVal.IsUnknown() {
		return pointer(tfVal.ValueBool())
	}
	return nil
}

func BoolWithFalseDefault(tfVal types.Bool) bool {
	if !tfVal.IsNull() && !tfVal.IsUnknown() {
		return tfVal.ValueBool()
	}
	return false
}

func pointer[T any](src T) *T {
	return &src
}

func OptionalString(tfVal types.String) *string {
	if tfVal.IsNull() || tfVal.IsUnknown() {
		return nil
	}
	return pointer(tfVal.ValueString())
}

func OptionalMap(ctx context.Context, tfVal types.Map) (map[string]string, error) {
	if tfVal.IsNull() || tfVal.IsUnknown() {
		return nil, nil
	}
	result := make(map[string]string, len(tfVal.Elements()))
	d := tfVal.ElementsAs(ctx, &result, false)
	if d.HasError() {
		return nil, fmt.Errorf("error converting to map object %v", d.Errors()[0].Detail())
	}

	return result, nil
}

func OptionalList(ctx context.Context, tfVal types.List) ([]string, error) {
	if tfVal.IsNull() || tfVal.IsUnknown() {
		return nil, nil
	}
	result := make([]string, len(tfVal.Elements()))
	d := tfVal.ElementsAs(ctx, &result, false)
	if d.HasError() {
		return nil, fmt.Errorf("error converting to map object %v", d.Errors()[0].Detail())
	}
	return result, nil
}

func ConvertStringMapToMapType(stringMap map[string]string) (types.Map, error) {
	elements := map[string]attr.Value{}
	for k, v := range stringMap {
		elements[k] = types.StringValue(v)
	}
	mapValue, diags := types.MapValue(types.StringType, elements)
	if diags != nil && diags.HasError() {
		fmt.Errorf("failed to convert to MapType %v", diags.Errors()[0].Detail())
	}
	return mapValue, nil
}

func ConvertStringListToListType(stringList []string) (types.List, error) {
	elements := []attr.Value{}
	for _, e := range stringList {
		elements = append(elements, types.StringValue(e))
	}
	listValue, diags := types.ListValue(types.StringType, elements)
	if diags != nil && diags.HasError() {
		fmt.Errorf("failed to convert to List type %v", diags.Errors()[0].Detail())
	}
	return listValue, nil
}
