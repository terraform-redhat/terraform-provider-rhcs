package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func BoolWithFalseDefault(tfVal types.Bool) bool {
	if HasValue(tfVal) {
		return tfVal.ValueBool()
	}
	return false
}

func OptionalInt64(tfVal types.Int64) *int64 {
	if HasValue(tfVal) {
		return tfVal.ValueInt64Pointer()
	}
	return nil
}

func OptionalString(tfVal types.String) *string {
	if HasValue(tfVal) {
		return tfVal.ValueStringPointer()
	}
	return nil
}
func OptionalMap(ctx context.Context, tfVal types.Map) (map[string]string, error) {
	if !HasValue(tfVal) {
		return nil, nil
	}
	result := make(map[string]string, len(tfVal.Elements()))
	d := tfVal.ElementsAs(ctx, &result, false)
	if d.HasError() {
		return nil, fmt.Errorf("error converting to map object %v", d.Errors()[0].Detail())
	}

	return result, nil
}

func StringListToArray(ctx context.Context, tfVal types.List) ([]string, error) {
	if !HasValue(tfVal) {
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
		return mapValue, fmt.Errorf("failed to convert to MapType %v", diags.Errors()[0].Detail())
	}
	return mapValue, nil
}

func StringArrayToList(stringList []string) (types.List, error) {
	elements := []attr.Value{}
	for _, e := range stringList {
		elements = append(elements, types.StringValue(e))
	}
	listValue, diags := types.ListValue(types.StringType, elements)
	if diags != nil && diags.HasError() {
		return listValue, fmt.Errorf("failed to convert to List type %v", diags.Errors()[0].Detail())
	}
	return listValue, nil
}
