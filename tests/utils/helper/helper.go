package helper

/*
Copyright (c***REMOVED*** 2018 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file contains helper functions for the tests.

***REMOVED***
	"bytes"
	"encoding/json"
***REMOVED***
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	// nolint
***REMOVED***
***REMOVED***

// Parse parses the given JSON data and returns a map of strings containing the result.
func Parse(data []byte***REMOVED*** map[string]interface{} {
	var object map[string]interface{}
	err := json.Unmarshal(data, &object***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	return object
}

// DigString tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is an string. For example, if the object is the result of parsing the
// following JSON document:
//
//	{
//		"kind": "Cluster",
//		"id": "123",
//		"flavour": {
//			"kind": "Flavour",
//			"id": "456",
//			"href": "/api/clusters_mgmt/v1/flavours/456"
//***REMOVED***
//	}
//
// The the 'id' attribute can be obtained like this:
//
//	clusterID := DigString(object, "id"***REMOVED***
//
// And the 'id' attribute inside the 'flavour' can be obtained like this:
//
//	flavourID := DigString(object, "flavour", "id"***REMOVED***
//
// If there is no attribute with the given path then the return value will be an empty string.
func DigString(object interface{}, keys ...interface{}***REMOVED*** string {
	switch result := dig(object, keys***REMOVED***.(type***REMOVED*** {
	case nil:
		return ""
	case string:
		return result
	case fmt.Stringer:
		return result.String(***REMOVED***
	default:
		return fmt.Sprintf("%s", result***REMOVED***
	}
}

// DigBool tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is a boolean. For example, if the object is the result of parsing the
// following JSON document:
//
//	{
//		"kind": "Cluster",
//		"id": "123",
//		"hasId": true,
//		"flavour": {
//			"kind": "Flavour",
//			"hasId": false,
//			"href": "/api/clusters_mgmt/v1/flavours/456"
//***REMOVED***
//	}
//
// The the 'hasId' attribute can be obtained like this:
//
//	hasID := DigBool(object, "hasId"***REMOVED***
//
// And the 'hasId' attribute inside the 'flavour' can be obtained like this:
//
//	flavourHasID := DigBool(object, "flavour", "hasId"***REMOVED***
//
// If there is no attribute with the given path then the return value will be false.
func DigBool(object interface{}, keys ...interface{}***REMOVED*** bool {
	switch result := dig(object, keys***REMOVED***.(type***REMOVED*** {
	case nil:
		return false
	case bool:
		return result
	case string:
		b, err := strconv.ParseBool(result***REMOVED***
		if err != nil {
			return false
***REMOVED***
		return b
	default:
		return false
	}
}

// DigInt tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is an integer. If there is no attribute with the given path then the test
// will be aborted with an error.
func DigInt(object interface{}, keys ...interface{}***REMOVED*** int {
	value := dig(object, keys***REMOVED***
	ExpectWithOffset(1, value***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	var result float64
	ExpectWithOffset(1, value***REMOVED***.To(BeAssignableToTypeOf(result***REMOVED******REMOVED***
	result = value.(float64***REMOVED***
	return int(result***REMOVED***
}

// DigFloat tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is an floating point number. If there is no attribute with the given path
// then the test will be aborted with an error.
func DigFloat(object interface{}, keys ...interface{}***REMOVED*** float64 {
	value := dig(object, keys***REMOVED***
	ExpectWithOffset(1, value***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	var result float64
	ExpectWithOffset(1, value***REMOVED***.To(BeAssignableToTypeOf(result***REMOVED******REMOVED***
	result = value.(float64***REMOVED***
	return result
}

// DigObject tries to find an attribute inside the given object with the given path, and returns its
// value. If there is no attribute with the given path then the test will be aborted with an error.
func DigObject(object interface{}, keys ...interface{}***REMOVED*** interface{} {
	value := dig(object, keys***REMOVED***
	ExpectWithOffset(1, value***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	return value
}

// DigArray tries to find an array inside the given object with the given path, and returns its
// value. If there is no attribute with the given path then the test will be aborted with an error.
func DigArray(object interface{}, keys ...interface{}***REMOVED*** []interface{} {
	value := dig(object, keys***REMOVED***
	//ExpectWithOffset(1, value***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	var result []interface{}
	//ExpectWithOffset(1, value***REMOVED***.To(BeAssignableToTypeOf(result***REMOVED******REMOVED***
	result = value.([]interface{}***REMOVED***
	return result
}
func DigStringArray(object interface{}, keys ...interface{}***REMOVED*** []string {
	value := dig(object, keys***REMOVED***
	//ExpectWithOffset(1, value***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	var result []interface{}
	//ExpectWithOffset(1, value***REMOVED***.To(BeAssignableToTypeOf(result***REMOVED******REMOVED***
	result = value.([]interface{}***REMOVED***
	stringRes := []string{}
	for _, r := range result {
		stringRes = append(stringRes, r.(string***REMOVED******REMOVED***
	}
	return stringRes
}

func dig(object interface{}, keys []interface{}***REMOVED*** interface{} {
	if object == nil || len(keys***REMOVED*** == 0 {
		return nil
	}
	switch key := keys[0].(type***REMOVED*** {
	case string:
		switch data := object.(type***REMOVED*** {
		case map[string]interface{}:
			value := data[key]
			if len(keys***REMOVED*** == 1 {
				return value
	***REMOVED***
			return dig(value, keys[1:]***REMOVED***
***REMOVED***
	case int:
		switch data := object.(type***REMOVED*** {
		case []interface{}:
			value := data[key]
			if len(keys***REMOVED*** == 1 {
				return value
	***REMOVED***
			return dig(value, keys[1:]***REMOVED***
***REMOVED***
	}
	return nil
}

// Template processes the given template using as data the set of name value pairs that are given as
// arguments. For example, to the following code:
//
//	result, err := Template(`
//		{
//			"name": "{{ .Name }}",
//			"flavour": {
//				"id": "{{ .Flavour }}"
//	***REMOVED***
//***REMOVED***
//		`,
//		"Name", "mycluster",
//		"Flavour", "4",
//	***REMOVED***
//
// Produces the following result:
//
//	{
//		"name": "mycluster",
//		"flavour": {
//			"id": "4"
//***REMOVED***
//	}
func Template(source string, args ...interface{}***REMOVED*** string {
	// Check that there is an even number of args, and that the first of each pair is an string:
	count := len(args***REMOVED***
	ExpectWithOffset(1, count%2***REMOVED***.To(
		Equal(0***REMOVED***,
		"Template '%s' should have an even number of arguments, but it has %d",
		source, count,
	***REMOVED***
	for i := 0; i < count; i = i + 2 {
		name := args[i]
		_, ok := name.(string***REMOVED***
		ExpectWithOffset(1, ok***REMOVED***.To(
			BeTrue(***REMOVED***,
			"Argument %d of template '%s' is a key, so it should be a string, "+
				"but its type is %T",
			i, source, name,
		***REMOVED***
	}

	// Put the variables in the map that will be passed as the data object for the execution of
	// the template:
	data := make(map[string]interface{}***REMOVED***
	for i := 0; i < count; i = i + 2 {
		name := args[i].(string***REMOVED***
		value := args[i+1]
		data[name] = value
	}

	// Parse the template:
	tmpl, err := template.New(""***REMOVED***.Parse(source***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(
		HaveOccurred(***REMOVED***,
		"Can't parse template '%s': %v",
		source, err,
	***REMOVED***

	// Execute the template:
	buffer := new(bytes.Buffer***REMOVED***
	err = tmpl.Execute(buffer, data***REMOVED***
	ExpectWithOffset(1, err***REMOVED***.ToNot(
		HaveOccurred(***REMOVED***,
		"Can't execute template '%s': %v",
		source, err,
	***REMOVED***
	return buffer.String(***REMOVED***
}

func isLocal(url string***REMOVED*** bool {
	return strings.Contains(url, "localhost"***REMOVED*** ||
		strings.Contains(url, "127.0.0.1"***REMOVED*** ||
		strings.Contains(url, "::1"***REMOVED***
}

// runAttempt will run a function (attempt***REMOVED***, until the function returns false - meaning no further attempts should
// follow, or until the number of attempts reached maxAttempts. Between each 2 attempts, it will wait for a given
// delay time.
// In case maxAttempts have been reached, an error will be returned, with the latest attempt result.
// The attempt function should return true as long as another attempt should be made, false when no further attempts
// are required - i.e. the attempt succeeded, and the result is available as returned value.
func runAttempt(attempt func(***REMOVED*** (interface{}, bool***REMOVED***, maxAttempts int, delay time.Duration***REMOVED*** (interface{}, error***REMOVED*** {
	var result interface{}
	for i := 0; i < maxAttempts; i++ {

		result, toContinue := attempt(***REMOVED***
		if toContinue {

***REMOVED*** else {

			return result, nil
***REMOVED***

		time.Sleep(delay***REMOVED***
	}
	return result, fmt.Errorf("Got to max attempts %d", maxAttempts***REMOVED***
}

func RunCMD(cmd string***REMOVED*** (stdout string, stderr string, err error***REMOVED*** {
	// logger.Info("[>>] Running CMD: ", cmd***REMOVED***
	var stdoutput bytes.Buffer
	var stderroutput bytes.Buffer
	CMD := exec.Command("bash", "-c", cmd***REMOVED***
	CMD.Stderr = &stderroutput
	CMD.Stdout = &stdoutput
	err = CMD.Run(***REMOVED***
	stdout = strings.Trim(stdoutput.String(***REMOVED***, "\n"***REMOVED***
	stderr = strings.Trim(stderroutput.String(***REMOVED***, "\n"***REMOVED***
	return
}

// NewRand returns a rand with the time seed
func NewRand(***REMOVED*** *rand.Rand {
	return rand.New(rand.NewSource(time.Now(***REMOVED***.UnixNano(***REMOVED******REMOVED******REMOVED***
}

// Join will link the strings with "."
func Join(s ...string***REMOVED*** string {
	return strings.Join(s, "."***REMOVED***
}

// IsSorted will return whether the array is sorted by mode
func IsSorted(arry []string, mode string***REMOVED*** (flag bool***REMOVED*** {
	switch mode {
	case "desc":
		for i := 0; i < len(arry***REMOVED***-1; i++ {
			if arry[i] < arry[i+1] {
				flag = false
	***REMOVED***
***REMOVED***
		flag = true
	case "asc":
		for i := 0; i < len(arry***REMOVED***-1; i++ {
			if arry[i] > arry[i+1] {
				flag = false
	***REMOVED***
***REMOVED***
		flag = true
	default:
		for i := 0; i < len(arry***REMOVED***-1; i++ {
			if arry[i] > arry[i+1] {
				flag = false
	***REMOVED***
***REMOVED***
		flag = true
	}
	return
}

// Min will return the minimize value
func Min(a int, b int***REMOVED*** int {
	if a >= b {
		return b
	}
	return a
}

// Min will return the minimize value
func Max(a int, b int***REMOVED*** int {
	if a <= b {
		return b
	}
	return a
}

// Contains will return bool balue that whether the arry contains string val
func Contains(arry []string, val string***REMOVED*** (index int, flag bool***REMOVED*** {
	var i int
	flag = false
	if len(arry***REMOVED*** == 0 {
		return
	}
	for i = 0; i < len(arry***REMOVED***; i++ {
		if arry[i] == val {
			index = i
			flag = true
			break
***REMOVED***
	}
	return
}

// EndsWith will return the bool value that whether the st is end with substring
func EndsWith(st string, substring string***REMOVED*** (flag bool***REMOVED*** {

	if len(st***REMOVED*** < len(substring***REMOVED*** {
		return
	}
	flag = (st[len(st***REMOVED***-len(substring***REMOVED***:] == substring***REMOVED***
	return
}

// Strip will return the value striped with substring
func Strip(st string, substring string***REMOVED*** string {
	st = Lstrip(st, substring***REMOVED***
	st = Rstrip(st, substring***REMOVED***
	return st
}

// Lstrip will return the string left striped with substring
func Lstrip(st string, substring string***REMOVED*** string {
	if StartsWith(st, substring***REMOVED*** {
		st = st[len(substring***REMOVED***:]
	}
	return st
}

// Rstrip will return the string right striped with substring
func Rstrip(st string, substring string***REMOVED*** string {
	if EndsWith(st, substring***REMOVED*** {
		st = st[:len(st***REMOVED***-len(substring***REMOVED***]
	}
	return st
}

// StartsWith return bool whether st start with substring
func StartsWith(st string, substring string***REMOVED*** (flag bool***REMOVED*** {
	if len(st***REMOVED*** < len(substring***REMOVED*** {
		return
	}
	flag = (st[:len(substring***REMOVED***] == substring***REMOVED***
	return
}

// NegateBoolToString reverts the boolean to its oppositely value as a string.
func NegateBoolToString(value bool***REMOVED*** string {
	boolString := "true"
	if value {
		boolString = "false"
	}
	return boolString
}

// ConvertMapToJSONString converts the map to a json string.
func ConvertMapToJSONString(inputMap map[string]interface{}***REMOVED*** string {
	jsonBytes, _ := json.Marshal(inputMap***REMOVED***
	return string(jsonBytes***REMOVED***
}

func ConvertStringToInt(mystring string***REMOVED*** int {
	myint, _ := strconv.Atoi(mystring***REMOVED***
	return myint
}

func ConvertStructToMap(s interface{}***REMOVED*** map[string]interface{} {
	structMap := make(map[string]interface{}***REMOVED***

	j, _ := json.Marshal(s***REMOVED***
	err := json.Unmarshal(j, &structMap***REMOVED***
	if err != nil {
		panic(err***REMOVED***
	}

	return structMap
}

func ConvertStructToString(s interface{}***REMOVED*** string {
	structMap := ConvertStructToMap(s***REMOVED***
	return ConvertMapToJSONString(structMap***REMOVED***
}

func IsInMap(inputMap map[string]interface{}, key string***REMOVED*** bool {
	_, contain := inputMap[key]
	return contain
}

func BoolPoint(b bool***REMOVED*** *bool {
	boolVar := b
	return &boolVar
}

const (
	UnderscoreConnector string = "_"
	DotConnector        string = "."
	HyphenConnector     string = "-"
***REMOVED***

// Dig Expose dig to used by others
func Dig(object interface{}, keys []interface{}***REMOVED*** interface{} {
	return dig(object, keys***REMOVED***
}

func DigByConnector(object interface{}, key string***REMOVED*** interface{} {
	keys := strings.Split(key, DotConnector***REMOVED***
	var keysInterface []interface{}
	for _, key := range keys {
		keysInterface = append(keysInterface, key***REMOVED***
	}

	return Dig(object, keysInterface***REMOVED***
}

// FlatMap parses the data, and stores all the attributes and the sub attributes at the same level with the prefix key.
func FlatMap(inputMap map[string]interface{}, outputMap map[string]interface{}, key string, connector string***REMOVED*** {
	if len(inputMap***REMOVED*** == 0 {
		outputMap[key] = ""
	} else {
		for k, v := range inputMap {
			outKey := fmt.Sprintf("%s%s%s", key, connector, k***REMOVED***
			if key == "" {
				outKey = k
	***REMOVED***
			switch v := v.(type***REMOVED*** {
			case map[string]interface{}:
				FlatMap(v, outputMap, outKey, connector***REMOVED***
			case []interface{}:
				// If the keys are in an array, it will flat the array with the index, etc, items_0_id or items.0.id
				for nk, nv := range v {
					newOutKey := fmt.Sprintf("%s%s%d", outKey, connector, nk***REMOVED***

					switch nv := nv.(type***REMOVED*** {
					case map[string]interface{}:
						FlatMap(nv, outputMap, newOutKey, connector***REMOVED***
					default:
						outputMap[newOutKey] = nv
			***REMOVED***
		***REMOVED***
			default:
				outputMap[outKey] = v
	***REMOVED***
***REMOVED***
	}

}

// FlatInitialMap parses the data to store all the attributes at the first level.
func FlatInitialMap(inputMap map[string]interface{}, connector string***REMOVED*** map[string]interface{} {
	outputMap := make(map[string]interface{}***REMOVED***
	FlatMap(inputMap, outputMap, "", connector***REMOVED***
	return outputMap
}

func FlatInterface(input []interface{}, outputMap map[string]interface{}, key string, connector string***REMOVED*** {
	if len(input***REMOVED*** == 0 {
		outputMap[key] = ""
	} else {
		for n, v := range input {
			outKey := fmt.Sprintf("%s%s%d", key, connector, n***REMOVED***
			switch v := v.(type***REMOVED*** {
			case map[string]interface{}:
				for mk, mv := range v {
					newOutKey := fmt.Sprintf("%s%s%s", outKey, connector, mk***REMOVED***
					switch mv := mv.(type***REMOVED*** {
					case []interface{}:
						FlatInterface(mv, outputMap, newOutKey, connector***REMOVED***
					default:
						outputMap[newOutKey] = mv
			***REMOVED***
		***REMOVED***
			case []interface{}:
				// If the keys are in an array, it will flat the array with the index, etc, items_0_id or items.0.id
				FlatInterface(v, outputMap, outKey, connector***REMOVED***
			default:
				outputMap[outKey] = v
	***REMOVED***
***REMOVED***
	}

}

// FlatInitialMap parses the data to store all the attributes at the first level.
func FlatDetails(input []interface{}, connector string***REMOVED*** map[string]interface{} {
	outputMap := make(map[string]interface{}***REMOVED***
	FlatInterface(input, outputMap, "details", connector***REMOVED***
	return outputMap
}

// MapStructure will map the map to the address of the structre *i
func MapStructure(m map[string]interface{}, i interface{}***REMOVED*** error {
	jsonbody, err := json.Marshal(m***REMOVED***
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonbody, i***REMOVED***
	if err != nil {
		return err
	}
	return nil
}

// ConvertRequestBodyByAttr will return the request bodies by attributes
func ConvertRequestBodyByAttr(inputString string, connector string***REMOVED*** (outArray []string, err error***REMOVED*** {
	inputMap := Parse([]byte(inputString***REMOVED******REMOVED***
	outMap := FlatInitialMap(inputMap, connector***REMOVED***
	outArray = ConvertFlatMapToArray(outMap, connector***REMOVED***
	return
}

// ConvertFlatMapToArray will return the request body converted from flatmap
// {"id":"xxx", "product.id":"xxxx", "managed":true}
// will be converted to
// ["{"id":"xxxx"}","{"product":{"id":"xxxx"}}",...]
func ConvertFlatMapToArray(flatMap map[string]interface{}, connector string***REMOVED*** (outArray []string***REMOVED*** {
	for key, value := range flatMap {
		keys := strings.Split(key, connector***REMOVED***
		resultmap := make(map[string]interface{}***REMOVED***
		for i := len(keys***REMOVED*** - 1; i >= 0; i-- {
			middleMap := make(map[string]interface{}***REMOVED***
			if i == len(keys***REMOVED***-1 {
				middleMap[keys[i]] = value

	***REMOVED*** else {
				middleMap[keys[i]] = resultmap
	***REMOVED***
			resultmap = middleMap

***REMOVED***
		requestBody, err := json.Marshal(resultmap***REMOVED***
		if err != nil {
			return
***REMOVED***
		outArray = append(outArray, string(requestBody***REMOVED******REMOVED***
	}
	return
}

// LookUpKey support find a key in a map
func LookUpKey(sourceMap interface{}, key interface{}***REMOVED*** (result interface{}***REMOVED*** {
	switch sourceMap.(type***REMOVED*** {
	case map[string]interface{}:
		for k, value := range sourceMap.(map[string]interface{}***REMOVED*** {
			if k == Lstrip(key.(string***REMOVED***, "^"***REMOVED*** {
				result = value
				return result
	***REMOVED***
			if !StartsWith(key.(string***REMOVED***, "^"***REMOVED*** {
				switch value.(type***REMOVED*** {
				case map[string]interface{}:
					result = LookUpKey(value, key***REMOVED***
					if result != nil {
						return
			***REMOVED***
		***REMOVED***
	***REMOVED***

***REMOVED***
	case []interface{}:
		switch key.(type***REMOVED*** {
		case int:
			result = sourceMap.([]interface{}***REMOVED***[key.(int***REMOVED***]
			return
***REMOVED***
	}
	return
}

// LookUpKeys support find keys that has not an abosolute path in a map
func LookUpKeys(sourceMap interface{}, keys ...interface{}***REMOVED*** (result interface{}***REMOVED*** {
	for _, key := range keys {
		sourceMap = LookUpKey(sourceMap, key***REMOVED***
	}
	result = sourceMap
	return
}

// MixCases converts a string of lowercase into a string of mixed case
func MixCases(s string***REMOVED*** string {
	var res string
	for i, c := range s {
		if i%2 == 0 {
			res += strings.ToUpper(string(c***REMOVED******REMOVED***
***REMOVED*** else {
			res += strings.ToLower(string(c***REMOVED******REMOVED***
***REMOVED***
	}
	return res
}
