package common

import "github.com/hashicorp/terraform-plugin-framework/types"

func OptionalInt64(tfVal types.Int64) *int64 {
	if tfVal.Unknown || tfVal.Null {
		return nil
	}
	return &tfVal.Value
}

func Bool(tfVal types.Bool) bool {
	return !tfVal.Unknown && !tfVal.Null && tfVal.Value
}
