package helper

/*
Copyright (c) 2018 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	// nolint
	. "github.com/onsi/gomega"
)

// Parse parses the given JSON data and returns a map of strings containing the result.
func Parse(data []byte) map[string]interface{} {
	var object map[string]interface{}
	err := json.Unmarshal(data, &object)
	Expect(err).ToNot(HaveOccurred())
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
//		}
//	}
//
// The the 'id' attribute can be obtained like this:
//
//	clusterID := DigString(object, "id")
//
// And the 'id' attribute inside the 'flavour' can be obtained like this:
//
//	flavourID := DigString(object, "flavour", "id")
//
// If there is no attribute with the given path then the return value will be an empty string.
func DigString(object interface{}, keys ...interface{}) string {
	switch result := dig(object, keys).(type) {
	case nil:
		return ""
	case string:
		return result
	case fmt.Stringer:
		return result.String()
	default:
		return fmt.Sprintf("%s", result)
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
//		}
//	}
//
// The the 'hasId' attribute can be obtained like this:
//
//	hasID := DigBool(object, "hasId")
//
// And the 'hasId' attribute inside the 'flavour' can be obtained like this:
//
//	flavourHasID := DigBool(object, "flavour", "hasId")
//
// If there is no attribute with the given path then the return value will be false.
func DigBool(object interface{}, keys ...interface{}) bool {
	switch result := dig(object, keys).(type) {
	case nil:
		return false
	case bool:
		return result
	case string:
		b, err := strconv.ParseBool(result)
		if err != nil {
			return false
		}
		return b
	default:
		return false
	}
}

// DigInt tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is an integer. If there is no attribute with the given path then the test
// will be aborted with an error.
func DigInt(object interface{}, keys ...interface{}) int {
	value := dig(object, keys)
	ExpectWithOffset(1, value).ToNot(BeNil())
	var result float64
	ExpectWithOffset(1, value).To(BeAssignableToTypeOf(result))
	result = value.(float64)
	return int(result)
}

// DigFloat tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is an floating point number. If there is no attribute with the given path
// then the test will be aborted with an error.
func DigFloat(object interface{}, keys ...interface{}) float64 {
	value := dig(object, keys)
	ExpectWithOffset(1, value).ToNot(BeNil())
	var result float64
	ExpectWithOffset(1, value).To(BeAssignableToTypeOf(result))
	result = value.(float64)
	return result
}

// DigObject tries to find an attribute inside the given object with the given path, and returns its
// value. If there is no attribute with the given path then the test will be aborted with an error.
func DigObject(object interface{}, keys ...interface{}) interface{} {
	value := dig(object, keys)
	ExpectWithOffset(1, value).ToNot(BeNil())
	return value
}

// DigArray tries to find an array inside the given object with the given path, and returns its
// value. If there is no attribute with the given path then the test will be aborted with an error.
func DigArray(object interface{}, keys ...interface{}) []interface{} {
	value := dig(object, keys)
	//ExpectWithOffset(1, value).ToNot(BeNil())
	var result []interface{}
	//ExpectWithOffset(1, value).To(BeAssignableToTypeOf(result))
	result = value.([]interface{})
	return result
}
func DigStringArray(object interface{}, keys ...interface{}) []string {
	value := dig(object, keys)
	//ExpectWithOffset(1, value).ToNot(BeNil())
	var result []interface{}
	//ExpectWithOffset(1, value).To(BeAssignableToTypeOf(result))
	result = value.([]interface{})
	stringRes := []string{}
	for _, r := range result {
		stringRes = append(stringRes, r.(string))
	}
	return stringRes
}

func dig(object interface{}, keys []interface{}) interface{} {
	if object == nil || len(keys) == 0 {
		return nil
	}
	switch key := keys[0].(type) {
	case string:
		switch data := object.(type) {
		case map[string]interface{}:
			value := data[key]
			if len(keys) == 1 {
				return value
			}
			return dig(value, keys[1:])
		}
	case int:
		switch data := object.(type) {
		case []interface{}:
			value := data[key]
			if len(keys) == 1 {
				return value
			}
			return dig(value, keys[1:])
		}
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
//			}
//		}
//		`,
//		"Name", "mycluster",
//		"Flavour", "4",
//	)
//
// Produces the following result:
//
//	{
//		"name": "mycluster",
//		"flavour": {
//			"id": "4"
//		}
//	}
func Template(source string, args ...interface{}) string {
	// Check that there is an even number of args, and that the first of each pair is an string:
	count := len(args)
	ExpectWithOffset(1, count%2).To(
		Equal(0),
		"Template '%s' should have an even number of arguments, but it has %d",
		source, count,
	)
	for i := 0; i < count; i = i + 2 {
		name := args[i]
		_, ok := name.(string)
		ExpectWithOffset(1, ok).To(
			BeTrue(),
			"Argument %d of template '%s' is a key, so it should be a string, "+
				"but its type is %T",
			i, source, name,
		)
	}

	// Put the variables in the map that will be passed as the data object for the execution of
	// the template:
	data := make(map[string]interface{})
	for i := 0; i < count; i = i + 2 {
		name := args[i].(string)
		value := args[i+1]
		data[name] = value
	}

	// Parse the template:
	tmpl, err := template.New("").Parse(source)
	ExpectWithOffset(1, err).ToNot(
		HaveOccurred(),
		"Can't parse template '%s': %v",
		source, err,
	)

	// Execute the template:
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, data)
	ExpectWithOffset(1, err).ToNot(
		HaveOccurred(),
		"Can't execute template '%s': %v",
		source, err,
	)
	return buffer.String()
}

func isLocal(url string) bool {
	return strings.Contains(url, "localhost") ||
		strings.Contains(url, "127.0.0.1") ||
		strings.Contains(url, "::1")
}

// runAttempt will run a function (attempt), until the function returns false - meaning no further attempts should
// follow, or until the number of attempts reached maxAttempts. Between each 2 attempts, it will wait for a given
// delay time.
// In case maxAttempts have been reached, an error will be returned, with the latest attempt result.
// The attempt function should return true as long as another attempt should be made, false when no further attempts
// are required - i.e. the attempt succeeded, and the result is available as returned value.
func runAttempt(attempt func() (interface{}, bool), maxAttempts int, delay time.Duration) (interface{}, error) {
	var result interface{}
	for i := 0; i < maxAttempts; i++ {

		result, toContinue := attempt()
		if toContinue {

		} else {

			return result, nil
		}

		time.Sleep(delay)
	}
	return result, fmt.Errorf("Got to max attempts %d", maxAttempts)
}

func RunCMD(cmd string) (stdout string, stderr string, err error) {
	// logger.Info("[>>] Running CMD: ", cmd)
	var stdoutput bytes.Buffer
	var stderroutput bytes.Buffer
	CMD := exec.Command("bash", "-c", cmd)
	CMD.Stderr = &stderroutput
	CMD.Stdout = &stdoutput
	err = CMD.Run()
	stdout = strings.Trim(stdoutput.String(), "\n")
	stderr = strings.Trim(stderroutput.String(), "\n")
	return
}

// NewRand returns a rand with the time seed
func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Join will link the strings with "."
func Join(s ...string) string {
	return strings.Join(s, ".")
}

// IsSorted will return whether the array is sorted by mode
func IsSorted(arry []string, mode string) (flag bool) {
	switch mode {
	case "desc":
		for i := 0; i < len(arry)-1; i++ {
			if arry[i] < arry[i+1] {
				flag = false
			}
		}
		flag = true
	case "asc":
		for i := 0; i < len(arry)-1; i++ {
			if arry[i] > arry[i+1] {
				flag = false
			}
		}
		flag = true
	default:
		for i := 0; i < len(arry)-1; i++ {
			if arry[i] > arry[i+1] {
				flag = false
			}
		}
		flag = true
	}
	return
}

// Min will return the minimize value
func Min(a int, b int) int {
	if a >= b {
		return b
	}
	return a
}

// Min will return the minimize value
func Max(a int, b int) int {
	if a <= b {
		return b
	}
	return a
}

// Contains will return bool balue that whether the arry contains string val
func Contains(arry []string, val string) (index int, flag bool) {
	var i int
	flag = false
	if len(arry) == 0 {
		return
	}
	for i = 0; i < len(arry); i++ {
		if arry[i] == val {
			index = i
			flag = true
			break
		}
	}
	return
}

// EndsWith will return the bool value that whether the st is end with substring
func EndsWith(st string, substring string) (flag bool) {

	if len(st) < len(substring) {
		return
	}
	flag = (st[len(st)-len(substring):] == substring)
	return
}

// Strip will return the value striped with substring
func Strip(st string, substring string) string {
	st = Lstrip(st, substring)
	st = Rstrip(st, substring)
	return st
}

// Lstrip will return the string left striped with substring
func Lstrip(st string, substring string) string {
	if StartsWith(st, substring) {
		st = st[len(substring):]
	}
	return st
}

// Rstrip will return the string right striped with substring
func Rstrip(st string, substring string) string {
	if EndsWith(st, substring) {
		st = st[:len(st)-len(substring)]
	}
	return st
}

// StartsWith return bool whether st start with substring
func StartsWith(st string, substring string) (flag bool) {
	if len(st) < len(substring) {
		return
	}
	flag = (st[:len(substring)] == substring)
	return
}

// NegateBoolToString reverts the boolean to its oppositely value as a string.
func NegateBoolToString(value bool) string {
	boolString := "true"
	if value {
		boolString = "false"
	}
	return boolString
}

// ConvertMapToJSONString converts the map to a json string.
func ConvertMapToJSONString(inputMap map[string]interface{}) string {
	jsonBytes, _ := json.Marshal(inputMap)
	return string(jsonBytes)
}

func ConvertStringToInt(mystring string) int {
	myint, _ := strconv.Atoi(mystring)
	return myint
}

func ConvertStructToMap(s interface{}) map[string]interface{} {
	structMap := make(map[string]interface{})

	j, _ := json.Marshal(s)
	err := json.Unmarshal(j, &structMap)
	if err != nil {
		panic(err)
	}

	return structMap
}

func ConvertStructToString(s interface{}) string {
	structMap := ConvertStructToMap(s)
	return ConvertMapToJSONString(structMap)
}

func IsInMap(inputMap map[string]interface{}, key string) bool {
	_, contain := inputMap[key]
	return contain
}

func BoolPoint(b bool) *bool {
	boolVar := b
	return &boolVar
}

const (
	UnderscoreConnector string = "_"
	DotConnector        string = "."
	HyphenConnector     string = "-"
)

// Dig Expose dig to used by others
func Dig(object interface{}, keys []interface{}) interface{} {
	return dig(object, keys)
}

func DigByConnector(object interface{}, key string) interface{} {
	keys := strings.Split(key, DotConnector)
	var keysInterface []interface{}
	for _, key := range keys {
		keysInterface = append(keysInterface, key)
	}

	return Dig(object, keysInterface)
}

// FlatMap parses the data, and stores all the attributes and the sub attributes at the same level with the prefix key.
func FlatMap(inputMap map[string]interface{}, outputMap map[string]interface{}, key string, connector string) {
	if len(inputMap) == 0 {
		outputMap[key] = ""
	} else {
		for k, v := range inputMap {
			outKey := fmt.Sprintf("%s%s%s", key, connector, k)
			if key == "" {
				outKey = k
			}
			switch v := v.(type) {
			case map[string]interface{}:
				FlatMap(v, outputMap, outKey, connector)
			case []interface{}:
				// If the keys are in an array, it will flat the array with the index, etc, items_0_id or items.0.id
				for nk, nv := range v {
					newOutKey := fmt.Sprintf("%s%s%d", outKey, connector, nk)

					switch nv := nv.(type) {
					case map[string]interface{}:
						FlatMap(nv, outputMap, newOutKey, connector)
					default:
						outputMap[newOutKey] = nv
					}
				}
			default:
				outputMap[outKey] = v
			}
		}
	}

}

// FlatInitialMap parses the data to store all the attributes at the first level.
func FlatInitialMap(inputMap map[string]interface{}, connector string) map[string]interface{} {
	outputMap := make(map[string]interface{})
	FlatMap(inputMap, outputMap, "", connector)
	return outputMap
}

func FlatInterface(input []interface{}, outputMap map[string]interface{}, key string, connector string) {
	if len(input) == 0 {
		outputMap[key] = ""
	} else {
		for n, v := range input {
			outKey := fmt.Sprintf("%s%s%d", key, connector, n)
			switch v := v.(type) {
			case map[string]interface{}:
				for mk, mv := range v {
					newOutKey := fmt.Sprintf("%s%s%s", outKey, connector, mk)
					switch mv := mv.(type) {
					case []interface{}:
						FlatInterface(mv, outputMap, newOutKey, connector)
					default:
						outputMap[newOutKey] = mv
					}
				}
			case []interface{}:
				// If the keys are in an array, it will flat the array with the index, etc, items_0_id or items.0.id
				FlatInterface(v, outputMap, outKey, connector)
			default:
				outputMap[outKey] = v
			}
		}
	}

}

// FlatInitialMap parses the data to store all the attributes at the first level.
func FlatDetails(input []interface{}, connector string) map[string]interface{} {
	outputMap := make(map[string]interface{})
	FlatInterface(input, outputMap, "details", connector)
	return outputMap
}

// MapStructure will map the map to the address of the structre *i
func MapStructure(m map[string]interface{}, i interface{}) error {
	jsonbody, err := json.Marshal(m)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonbody, i)
	if err != nil {
		return err
	}
	return nil
}

// ConvertRequestBodyByAttr will return the request bodies by attributes
func ConvertRequestBodyByAttr(inputString string, connector string) (outArray []string, err error) {
	inputMap := Parse([]byte(inputString))
	outMap := FlatInitialMap(inputMap, connector)
	outArray = ConvertFlatMapToArray(outMap, connector)
	return
}

// ConvertFlatMapToArray will return the request body converted from flatmap
// {"id":"xxx", "product.id":"xxxx", "managed":true}
// will be converted to
// ["{"id":"xxxx"}","{"product":{"id":"xxxx"}}",...]
func ConvertFlatMapToArray(flatMap map[string]interface{}, connector string) (outArray []string) {
	for key, value := range flatMap {
		keys := strings.Split(key, connector)
		resultmap := make(map[string]interface{})
		for i := len(keys) - 1; i >= 0; i-- {
			middleMap := make(map[string]interface{})
			if i == len(keys)-1 {
				middleMap[keys[i]] = value

			} else {
				middleMap[keys[i]] = resultmap
			}
			resultmap = middleMap

		}
		requestBody, err := json.Marshal(resultmap)
		if err != nil {
			return
		}
		outArray = append(outArray, string(requestBody))
	}
	return
}

// LookUpKey support find a key in a map
func LookUpKey(sourceMap interface{}, key interface{}) (result interface{}) {
	switch sourceMap.(type) {
	case map[string]interface{}:
		for k, value := range sourceMap.(map[string]interface{}) {
			if k == Lstrip(key.(string), "^") {
				result = value
				return result
			}
			if !StartsWith(key.(string), "^") {
				switch value.(type) {
				case map[string]interface{}:
					result = LookUpKey(value, key)
					if result != nil {
						return
					}
				}
			}

		}
	case []interface{}:
		switch key.(type) {
		case int:
			result = sourceMap.([]interface{})[key.(int)]
			return
		}
	}
	return
}

// LookUpKeys support find keys that has not an abosolute path in a map
func LookUpKeys(sourceMap interface{}, keys ...interface{}) (result interface{}) {
	for _, key := range keys {
		sourceMap = LookUpKey(sourceMap, key)
	}
	result = sourceMap
	return
}

// MixCases converts a string of lowercase into a string of mixed case
func MixCases(s string) string {
	var res string
	for i, c := range s {
		if i%2 == 0 {
			res += strings.ToUpper(string(c))
		} else {
			res += strings.ToLower(string(c))
		}
	}
	return res
}
