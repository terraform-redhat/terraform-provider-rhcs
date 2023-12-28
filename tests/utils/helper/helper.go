package helper

import (
	"bytes"
	r "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

type m = map[string]string

// combine two strings maps to one,
// if key already exists - do nothing
func MergeMaps(map1, map2 m) m {
	for k, v := range map2 {
		_, ok := map1[k]
		if !ok {
			map1[k] = v
		}
	}
	return map1
}

// Parse parses the given JSON data and returns a map of strings containing the result.
func Parse(data []byte) map[string]interface{} {
	var object map[string]interface{}
	err := json.Unmarshal(data, &object)
	Expect(err).ToNot(HaveOccurred())
	return object
}

func ParseStringToMap(input string) (map[string]string, error) {
	dataMap := make(map[string]string)

	// Define regex patterns for key-value pairs
	pattern := regexp.MustCompile(`(\w+)\s*=\s*"([^"]+)"`)

	// Find all matches in the input string
	matches := pattern.FindAllStringSubmatch(input, -1)

	// Populate the map with key-value pairs
	for _, match := range matches {
		key := match[1]
		value := match[2]
		dataMap[key] = value
	}

	return dataMap, nil
}

// If there is no attribute with the given path then the return value will be an empty string.
func DigString(object interface{}, keys ...interface{}) string {
	switch result := Dig(object, keys).(type) {
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

// NewRand returns a rand with the time seed
func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// DigStringArray tries to find an array inside the given object with the given path, and returns its
// value. If there is no attribute with the given path then the test will be aborted with an error.
func DigStringArray(object interface{}, keys ...interface{}) []string {
	value := Dig(object, keys)
	gomega.ExpectWithOffset(1, value).ToNot(gomega.BeNil())
	var result []string
	gomega.ExpectWithOffset(1, value).To(gomega.BeAssignableToTypeOf(result))
	result = value.([]string)
	return result
}

// DigArray tries to find an array inside the given object with the given path, and returns its
// value. If there is no attribute with the given path then the test will be aborted with an error.
func DigArray(object interface{}, keys ...interface{}) []interface{} {
	value := Dig(object, keys)
	result := value.([]interface{})
	return result
}

func DigArrayToString(object interface{}, keys ...interface{}) []string {
	value := Dig(object, keys)
	var result []interface{}
	if value == nil {
		return nil
	}
	result = value.([]interface{})
	strR := []string{}
	for _, r := range result {
		strR = append(strR, r.(string))
	}
	return strR
}

// DigInt tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is an integer. If there is no attribute with the given path then the test
// will be aborted with an error.
func DigInt(object interface{}, keys ...interface{}) int {
	value := Dig(object, keys)
	ExpectWithOffset(1, value).ToNot(BeNil())
	var result float64
	ExpectWithOffset(1, value).To(BeAssignableToTypeOf(result))
	result = value.(float64)
	return int(result)
}

func DigBool(object interface{}, keys ...interface{}) bool {
	switch result := Dig(object, keys).(type) {
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

func DigObject(object interface{}, keys ...interface{}) interface{} {
	value := Dig(object, keys)
	return value
}

func Dig(object interface{}, keys []interface{}) interface{} {
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
			return Dig(value, keys[1:])
		}
	case int:
		switch data := object.(type) {
		case []interface{}:
			value := data[key]
			if len(keys) == 1 {
				return value
			}
			return Dig(value, keys[1:])
		}
	}
	return nil
}

func RunCMD(cmd string) (stdout string, stderr string, err error) {
	Logger.Infof("[>>] Running CMD: %s", cmd)
	var stdoutput bytes.Buffer
	var stderroutput bytes.Buffer
	CMD := exec.Command("bash", "-c", cmd)
	CMD.Stderr = &stderroutput
	CMD.Stdout = &stdoutput
	err = CMD.Run()
	if err != nil {
		Logger.Errorf("Got error status: %v", err)
	}

	stdout = strings.Trim(stdoutput.String(), "\n")
	stderr = strings.Trim(stderroutput.String(), "\n")
	Logger.Infof("Got output %s", stdout)
	if stderr != "" {
		Logger.Errorf("Got error output %s", stderr)
	}
	return
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

// Join will link the strings with "."
func Join(s ...string) string {
	return strings.Join(s, ".")
}

func JoinStringWithArray(s string, strArray []string) []string {

	// create tmp map for joining strings
	var tmpMap = make(map[int]string)
	for i, str := range strArray {
		tmpMap[i] = s + str
	}

	// create array from map
	var newArray = make([]string, len(tmpMap))
	for i, value := range tmpMap {
		newArray[i] = value
	}
	return newArray
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

// NeedFiltered will return the attribute that should be filtered
// filterList should be array with regex like ["excluded\..+","excluded_[\s\S]+"]
func NeedFiltered(filterList []string, key string) bool {
	for _, regex := range filterList {
		pattern := regexp.MustCompile(regex)
		if pattern.MatchString(key) {
			if key == "network.type" {
				Logger.Infof(">>>> network.type matched regex: %s\n", regex)
			}
			return true
		}
	}
	return false
}

func BoolPoint(b bool) *bool {
	boolVar := b
	return &boolVar
}

func GenerateRandomStringWithSymbols(length int) string {
	b := make([]byte, length)
	_, err := r.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)[:length]
}

func Subfix(length int) string {
	subfix := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range subfix {
		subfix[i] = CON.CharsBytes[r.Intn(len(CON.CharsBytes))]
	}

	return string(subfix)
}

func GenerateClusterName(profileName string) string {

	clusterPrefix := CON.RHCSPrefix + CON.HyphenConnector + profileName[5:]
	return clusterPrefix + CON.HyphenConnector + Subfix(3)
}

func GetClusterAdminPassword() string {
	path := fmt.Sprintf(path.Join(CON.GetRHCSOutputDir(), CON.ClusterAdminUser))
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Print(err)
	}
	return string(b)
}
