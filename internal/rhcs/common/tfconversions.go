package common

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Bool(tfVal *bool) bool {
	return tfVal != nil && *tfVal
}

// ExpandStringValueList takes the result of flatmap.Expand for an array of strings
// and returns a []string
func ExpandStringValueList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}

// ExpandStringValueMap expands a string map of interfaces to a string map of strings
func ExpandStringValueMap(m map[string]interface{}) map[string]string {
	stringMap := make(map[string]string, len(m))
	for k, v := range m {
		stringMap[k] = v.(string)
	}
	return stringMap
}

func GetOptionalString(resourceData *schema.ResourceData, key string) *string {
	if v, ok := resourceData.GetOk(key); ok {
		return Pointer(v.(string))
	}
	return nil
}

func GetOptionalInt(resourceData *schema.ResourceData, key string) *int {
	if v, ok := resourceData.GetOk(key); ok {
		return Pointer(v.(int))
	}
	return nil
}

func GetOptionalBool(resourceData *schema.ResourceData, key string) *bool {
	if v, ok := resourceData.GetOkExists(key); ok {
		return Pointer(v.(bool))
	}
	return nil
}

func GetOptionalFloat(resourceData *schema.ResourceData, key string) *float64 {
	if v, ok := resourceData.GetOk(key); ok {
		return Pointer(v.(float64))
	}
	return nil
}

func GetOptionalListOfValueStringsFromResourceData(resourceData *schema.ResourceData, key string) []string {
	if v, ok := resourceData.GetOk(key); ok && len(v.([]interface{})) > 0 {
		return ExpandStringValueList(v.([]interface{}))
	}
	return nil
}

func GetOptionalMapStringFromResourceData(resourceData *schema.ResourceData, key string) map[string]string {
	if v, ok := resourceData.GetOk(key); ok && len(v.(map[string]interface{})) > 0 {
		return ExpandStringValueMap(v.(map[string]interface{}))
	}
	return map[string]string{}
}

func GetOptionalMapString(resourceMap map[string]interface{}, key string) map[string]string {
	if v, ok := resourceMap[key]; ok && len(v.(map[string]interface{})) > 0 {
		return ExpandStringValueMap(v.(map[string]interface{}))
	}
	return nil
}

func Pointer[T any](src T) *T {
	return &src
}

func GetOptionalStringFromMapString(mapString map[string]interface{}, key string) *string {
	if value, ok := mapString[key]; ok && value.(string) != "" {
		return Pointer(value.(string))
	}
	return nil
}

func GetOptionalBoolFromMapString(mapString map[string]interface{}, key string) *bool {
	if value, ok := mapString[key]; ok {
		return Pointer(value.(bool))
	}
	return nil
}

func GetOptionalListOfValueStrings(mapString map[string]interface{}, key string) []string {
	if value, ok := mapString[key]; ok && value != nil {
		return ExpandStringValueList(value.([]interface{}))
	}
	return nil
}
func GetOptionalStringValue(val interface{}) string {
	if val == nil {
		return ""
	}
	return val.(string)
}

// GetOptionalStringBool the default value is false
func GetOptionalInterfaceBool(val interface{}) bool {
	if val == nil {
		return false
	}
	return val.(bool)
}
