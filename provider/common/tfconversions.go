package common

***REMOVED***
	"context"
***REMOVED***

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
***REMOVED***

func BoolWithFalseDefault(tfVal types.Bool***REMOVED*** bool {
	if HasValue(tfVal***REMOVED*** {
		return tfVal.ValueBool(***REMOVED***
	}
	return false
}

func OptionalInt64(tfVal types.Int64***REMOVED*** *int64 {
	if HasValue(tfVal***REMOVED*** {
		return tfVal.ValueInt64Pointer(***REMOVED***
	}
	return nil
}

func OptionalString(tfVal types.String***REMOVED*** *string {
	if HasValue(tfVal***REMOVED*** {
		return tfVal.ValueStringPointer(***REMOVED***
	}
	return nil
}
func OptionalMap(ctx context.Context, tfVal types.Map***REMOVED*** (map[string]string, error***REMOVED*** {
	if !HasValue(tfVal***REMOVED*** {
		return nil, nil
	}
	result := make(map[string]string, len(tfVal.Elements(***REMOVED******REMOVED******REMOVED***
	d := tfVal.ElementsAs(ctx, &result, false***REMOVED***
	if d.HasError(***REMOVED*** {
		return nil, fmt.Errorf("error converting to map object %v", d.Errors(***REMOVED***[0].Detail(***REMOVED******REMOVED***
	}

	return result, nil
}

func OptionalList(tfVal types.List***REMOVED*** []string {
	if tfVal.IsUnknown(***REMOVED*** || tfVal.IsNull(***REMOVED*** {
		return nil
	}
	result := make([]string, 0***REMOVED***
	for _, e := range tfVal.Elements(***REMOVED*** {
		result = append(result, e.(types.String***REMOVED***.ValueString(***REMOVED******REMOVED***
	}
	return result
}

func StringListToArray(ctx context.Context, tfVal types.List***REMOVED*** ([]string, error***REMOVED*** {
	if !HasValue(tfVal***REMOVED*** {
		return nil, nil
	}
	result := make([]string, len(tfVal.Elements(***REMOVED******REMOVED******REMOVED***
	d := tfVal.ElementsAs(ctx, &result, false***REMOVED***
	if d.HasError(***REMOVED*** {
		return nil, fmt.Errorf("error converting to map object %v", d.Errors(***REMOVED***[0].Detail(***REMOVED******REMOVED***
	}
	return result, nil
}

func ConvertStringMapToMapType(stringMap map[string]string***REMOVED*** (types.Map, error***REMOVED*** {
	elements := map[string]attr.Value{}
	for k, v := range stringMap {
		elements[k] = types.StringValue(v***REMOVED***
	}
	mapValue, diags := types.MapValue(types.StringType, elements***REMOVED***
	if diags != nil && diags.HasError(***REMOVED*** {
		return mapValue, fmt.Errorf("failed to convert to MapType %v", diags.Errors(***REMOVED***[0].Detail(***REMOVED******REMOVED***
	}
	return mapValue, nil
}

func StringArrayToList(stringList []string***REMOVED*** (types.List, error***REMOVED*** {
	elements := []attr.Value{}
	for _, e := range stringList {
		elements = append(elements, types.StringValue(e***REMOVED******REMOVED***
	}
	listValue, diags := types.ListValue(types.StringType, elements***REMOVED***
	if diags != nil && diags.HasError(***REMOVED*** {
		return listValue, fmt.Errorf("failed to convert to List type %v", diags.Errors(***REMOVED***[0].Detail(***REMOVED******REMOVED***
	}
	return listValue, nil
}
