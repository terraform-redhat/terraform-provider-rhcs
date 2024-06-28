package helper

import (
	"bytes"
	r "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

// Parse parses the given JSON data and returns a map of strings containing the result.
func Parse(data []byte) map[string]interface{} {
	var object map[string]interface{}
	err := json.Unmarshal(data, &object)
	Expect(err).ToNot(HaveOccurred())
	return object
}

func ParseStringToMap(input string) (map[string]string, error) {

	// Attempt to parse input as JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(input), &jsonData); err == nil {
		// If successful, convert the map to map[string]string
		result := make(map[string]string)
		for key, value := range jsonData {
			// Convert each value to string
			if strValue, ok := value.(string); ok {
				result[key] = strValue
			} else {
				return nil, fmt.Errorf("non-string value found in JSON for key %s", key)
			}
		}
		return result, nil
	}

	// Fallback to regex-based parsing if JSON parsing fails
	dataMap := make(map[string]string)

	// Define regex patterns for key-value pairs
	pattern := regexp.MustCompile(`(\w+)\s*=\s*"([^"]+)"`)

	// Find all matches in the input string
	matches := pattern.FindAllStringSubmatch(input, -1)

	// Check if there are any matches
	if len(matches) == 0 {
		return nil, fmt.Errorf("no key-value pairs found in the input string")
	}

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
	if value == nil {
		return nil
	}
	var result []string
	result, ok := value.([]string)
	if !ok {
		return nil
	}
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

func DigMapToString(object interface{}, keys ...interface{}) map[string]string {
	value := Dig(object, keys)
	if value == nil {
		return nil
	}
	result, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	strMap := map[string]string{}
	for k, v := range result {
		strMap[k] = v.(string)
	}
	return strMap
}

// DigInt tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is an integer. If there is no attribute with the given path then the test
// will be aborted with an error.
func DigInt(object interface{}, keys ...interface{}) int {
	value := Dig(object, keys)
	if value == nil {
		return 0
	}
	result, ok := value.(float64)
	if !ok {
		return 0
	}
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

func GenerateRandomStringWithSymbols(length int) string {
	b := make([]byte, length)
	_, err := r.Read(b)
	if err != nil {
		panic(err)
	}
	randomString := base64.StdEncoding.EncodeToString(b)[:length]
	f := func(r rune) bool {
		return r < 'A' || r > 'z'
	}
	// Verify that the string contains special character or number
	if strings.IndexFunc(randomString, f) == -1 {
		randomString = randomString[:len(randomString)-1] + "!"
	}
	return randomString
}

// Generate random string
func GenerateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	s := make([]byte, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func GenerateRandomName(prefix string, n int) string {
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(GenerateRandomString(n)))
}

func RandomInt(max int) int {
	val, err := r.Int(r.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}
	return int(val.Int64())
}

func Subfix(length int) string {
	subfix := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range subfix {
		subfix[i] = constants.CharsBytes[r.Intn(len(constants.CharsBytes))]
	}

	return string(subfix)
}

func GenerateClusterName(profileName string) string {
	var clusterNameParts []string
	if constants.RHCS.RHCSClusterNamePrefix != "" {
		clusterNameParts = append(clusterNameParts, constants.RHCS.RHCSClusterNamePrefix)
	}
	clusterNameParts = append(clusterNameParts, constants.RHCSPrefix, profileName[5:], Subfix(3))
	if constants.RHCS.RHCSClusterNameSuffix != "" {
		clusterNameParts = append(clusterNameParts, constants.RHCS.RHCSClusterNameSuffix)
	}
	return strings.Join(clusterNameParts, constants.HyphenConnector)
}

func GetClusterAdminPassword() string {
	path := fmt.Sprintf(path.Join(constants.GetRHCSOutputDir(), constants.ClusterAdminUser))
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Print(err)
	}
	return string(b)
}

var (
	emptyStringValue      = ""
	emptyStringSliceValue = []string{}
)

var EmptyStringPointer = StringPointer(emptyStringValue)
var EmptyStringSlicePointer = StringSlicePointer(emptyStringSliceValue)

// Return a bool pointer of the input bool value
func BoolPointer(b bool) *bool {
	return &b
}

// Return a string pointer of the input string value
func StringPointer(s string) *string {
	return &s
}

// Return a pointer of the input int value
func IntPointer(i int) *int {
	return &i
}

func Float64Pointer(f float64) *float64 {
	return &f
}

func StringSlicePointer(f []string) *[]string {
	return &f
}

func IntSlicePointer(f []int) *[]int {
	return &f
}

func StringMapPointer(f map[string]string) *map[string]string {
	return &f
}

func Pointer(f interface{}) *interface{} {
	return &f
}

func GetTFErrorMessage(err error) string {
	return strings.ReplaceAll(err.Error(), "\n", " ")
}
