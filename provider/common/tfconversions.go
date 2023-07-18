package common

import "github.com/hashicorp/terraform-plugin-framework/types"

func OptionalInt64(tfVal types.Int64***REMOVED*** *int64 {
	if tfVal.Unknown || tfVal.Null {
		return nil
	}
	return &tfVal.Value
}

func Bool(tfVal types.Bool***REMOVED*** bool {
	return !tfVal.Unknown && !tfVal.Null && tfVal.Value
}

func OptionalString(tfVal types.String***REMOVED*** *string {
	if tfVal.Unknown || tfVal.Null {
		return nil
	}
	return &tfVal.Value
}

func OptionalMap(tfVal types.Map***REMOVED*** map[string]string {
	if tfVal.Unknown || tfVal.Null {
		return nil
	}
	result := map[string]string{}
	for k, v := range tfVal.Elems {
		result[k] = v.(types.String***REMOVED***.Value
	}
	return result
}

func OptionalList(tfVal types.List***REMOVED*** []string {
	if tfVal.Unknown || tfVal.Null {
		return nil
	}
	result := make([]string, 0***REMOVED***
	for _, e := range tfVal.Elems {
		result = append(result, e.(types.String***REMOVED***.Value***REMOVED***
	}
	return result
}
