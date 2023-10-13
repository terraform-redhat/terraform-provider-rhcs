package ci

***REMOVED***
	"encoding/json"
***REMOVED***
***REMOVED***
	"sort"
	"strings"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	EXE "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/exec"
***REMOVED***
***REMOVED***

// The cfg will be used to define the testing environment
var cfg = CON.RHCS

func GetYAMLProfileFile(fileName string***REMOVED*** (filename string***REMOVED*** {
	fPath := cfg.YAMLProfilesDir
	filename = path.Join(fPath, fileName***REMOVED***
	return
}

// ConvertToStringArray will convert the []interface to []string
func ConvertToStringArray(interfaceList []interface{}***REMOVED*** []string {
	var stringList []string
	for _, inter := range interfaceList {
		if inter != nil {
			stringList = append(stringList, inter.(string***REMOVED******REMOVED***
***REMOVED***
	}
	return stringList
}

// ConvertToStringMap will convert the []interface to []string
func ConvertToStringMap(interfaceMap map[string]interface{}***REMOVED*** map[string]string {
	stringMap := map[string]string{}
	for key, value := range interfaceMap {
		stringMap[key] = value.(string***REMOVED***
	}
	return stringMap
}

func ConvertToMap(qs interface{}***REMOVED*** map[string]interface{} {
	quotaSearchMap := make(map[string]interface{}***REMOVED***
	j, _ := json.Marshal(qs***REMOVED***
	err := json.Unmarshal(j, &quotaSearchMap***REMOVED***
	if err != nil {
		panic(err***REMOVED***
	}
	return quotaSearchMap
}

func convertToString(searchMap map[string]interface{}***REMOVED*** (filter string***REMOVED*** {
	var parameterArry []string
	for key, value := range searchMap {
		switch value := value.(type***REMOVED*** {
		case string:
			if h.StartsWith(value, "like "***REMOVED*** {
				parameterArry = append(parameterArry, fmt.Sprintf("%s like '%s'", key, h.Lstrip(value, "like "***REMOVED******REMOVED******REMOVED***
	***REMOVED*** else {
				parameterArry = append(parameterArry, fmt.Sprintf("%s='%s'", key, value***REMOVED******REMOVED***
	***REMOVED***
		case map[string]string:
			for subKey, subValue := range value {
				if h.StartsWith(subValue, "like "***REMOVED*** {
					parameterArry = append(parameterArry, fmt.Sprintf("%s like '%s'", h.Join(key, subKey***REMOVED***, subValue***REMOVED******REMOVED***
		***REMOVED*** else {
					parameterArry = append(parameterArry, fmt.Sprintf("%s='%s'", h.Join(key, subKey***REMOVED***, h.Lstrip(subValue, "like "***REMOVED******REMOVED******REMOVED***
		***REMOVED***
	***REMOVED***
***REMOVED***

	}
	filter = strings.Join(parameterArry, " and "***REMOVED***
	return
}

// ConvertFilterToString will Convert a fileter struct to a string
// if "like" in the value will be keeped
// if no 'like' in value, the string will contains key=value
// if map in value the substring will be key.subkey='subvalue'
func ConvertFilterToString(qs interface{}***REMOVED*** (filter string***REMOVED*** {
	filterMap := ConvertToMap(qs***REMOVED***
	filter = convertToString(filterMap***REMOVED***
	return
}

func TrimName(name string***REMOVED*** string {
	if len(name***REMOVED*** >= EXE.MaxNameLength {
		name = name[0:EXE.MaxNameLength]
	}
	return name
}

func TrimVersion(version string, groupChannel string***REMOVED*** string {
	prefix := "openshift-v"
	suffix := ""
	if groupChannel != EXE.StableChannel {
		suffix = "-" + groupChannel
	}
	trimedVersion := h.Rstrip(h.Lstrip(version, prefix***REMOVED***, suffix***REMOVED***
	return trimedVersion
}

// check if one element is in the array. if yes, return true; or return false
func ElementInArray(target string, str_array []string***REMOVED*** bool {
	sort.Strings(str_array***REMOVED***
	index := sort.SearchStrings(str_array, target***REMOVED***
	if index < len(str_array***REMOVED*** && str_array[index] == target {
		return true
	}
	return false
}

// GetElements will return an array or a string get from the items based on the num
// If num 0 and itemkey id passed it will return the itemkey value with index 0
func GetElements(content map[string]interface{}, element string, num ...int***REMOVED*** interface{} {
	var result []interface{}
	keys := strings.Split(element, CON.DotConnector***REMOVED***
	var keysInterface []interface{}
	for _, key := range keys {
		keysInterface = append(keysInterface, key***REMOVED***
	}
	items := h.DigArray(content, "items"***REMOVED***
	if len(num***REMOVED*** == 1 {
		return h.Dig(items[num[0]], keysInterface***REMOVED***
	}
	if len(num***REMOVED*** == 0 {
		for i := 0; i < len(items***REMOVED***; i++ {
			result = append(result, h.Dig(items[i], keysInterface***REMOVED******REMOVED***
***REMOVED***
		return result
	}
	for i := range num {
		result = append(result, h.Dig(items[i], keysInterface***REMOVED******REMOVED***
	}
	return result
}
