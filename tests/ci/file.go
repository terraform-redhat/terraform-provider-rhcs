package ci

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

// The cfg will be used to define the testing environment
var cfg = RHCS

func GetYAMLProfileFile(fileName string) (filename string) {
	fPath := cfg.YAMLProfilesDir
	filename = path.Join(fPath, fileName)
	return
}

// ConvertToStringArray will convert the []interface to []string
func ConvertToStringArray(interfaceList []interface{}) []string {
	var stringList []string
	for _, inter := range interfaceList {
		if inter != nil {
			stringList = append(stringList, inter.(string))
		}
	}
	return stringList
}

// ConvertToStringMap will convert the []interface to []string
func ConvertToStringMap(interfaceMap map[string]interface{}) map[string]string {
	stringMap := map[string]string{}
	for key, value := range interfaceMap {
		stringMap[key] = value.(string)
	}
	return stringMap
}

func ConvertToMap(qs interface{}) map[string]interface{} {
	quotaSearchMap := make(map[string]interface{})
	j, _ := json.Marshal(qs)
	err := json.Unmarshal(j, &quotaSearchMap)
	if err != nil {
		panic(err)
	}
	return quotaSearchMap
}

func convertToString(searchMap map[string]interface{}) (filter string) {
	var parameterArry []string
	for key, value := range searchMap {
		switch value := value.(type) {
		case string:
			if h.StartsWith(value, "like ") {
				parameterArry = append(parameterArry, fmt.Sprintf("%s like '%s'", key, h.Lstrip(value, "like ")))
			} else {
				parameterArry = append(parameterArry, fmt.Sprintf("%s='%s'", key, value))
			}
		case map[string]string:
			for subKey, subValue := range value {
				if h.StartsWith(subValue, "like ") {
					parameterArry = append(parameterArry, fmt.Sprintf("%s like '%s'", h.Join(key, subKey), subValue))
				} else {
					parameterArry = append(parameterArry, fmt.Sprintf("%s='%s'", h.Join(key, subKey), h.Lstrip(subValue, "like ")))
				}
			}
		}

	}
	filter = strings.Join(parameterArry, " and ")
	return
}

// ConvertFilterToString will Convert a fileter struct to a string
// if "like" in the value will be keeped
// if no 'like' in value, the string will contains key=value
// if map in value the substring will be key.subkey='subvalue'
func ConvertFilterToString(qs interface{}) (filter string) {
	filterMap := ConvertToMap(qs)
	filter = convertToString(filterMap)
	return
}

func TrimName(name string) string {
	if len(name) >= EXE.MaxNameLength {
		name = name[0:EXE.MaxNameLength]
	}
	return name
}

func TrimVersion(version string, groupChannel string) string {
	prefix := "openshift-v"
	suffix := ""
	if groupChannel != EXE.StableChannel {
		suffix = "-" + groupChannel
	}
	trimedVersion := h.Rstrip(h.Lstrip(version, prefix), suffix)
	return trimedVersion
}

// check if one element is in the array. if yes, return true; or return false
func ElementInArray(target string, str_array []string) bool {
	sort.Strings(str_array)
	index := sort.SearchStrings(str_array, target)
	if index < len(str_array) && str_array[index] == target {
		return true
	}
	return false
}

// GetElements will return an array or a string get from the items based on the num
// If num 0 and itemkey id passed it will return the itemkey value with index 0
func GetElements(content map[string]interface{}, element string, num ...int) interface{} {
	var result []interface{}
	keys := strings.Split(element, CON.DotConnector)
	var keysInterface []interface{}
	for _, key := range keys {
		keysInterface = append(keysInterface, key)
	}
	items := h.DigArray(content, "items")
	if len(num) == 1 {
		return h.Dig(items[num[0]], keysInterface)
	}
	if len(num) == 0 {
		for i := 0; i < len(items); i++ {
			result = append(result, h.Dig(items[i], keysInterface))
		}
		return result
	}
	for i := range num {
		result = append(result, h.Dig(items[i], keysInterface))
	}
	return result
}
