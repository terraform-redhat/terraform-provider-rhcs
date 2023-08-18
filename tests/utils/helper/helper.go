package helper

***REMOVED***
	"bytes"
	"encoding/json"
***REMOVED***
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega"
***REMOVED***
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***

// Parse parses the given JSON data and returns a map of strings containing the result.
func Parse(data []byte***REMOVED*** map[string]interface{} {
	var object map[string]interface{}
	err := json.Unmarshal(data, &object***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	return object
}

// If there is no attribute with the given path then the return value will be an empty string.
func DigString(object interface{}, keys ...interface{}***REMOVED*** string {
	switch result := Dig(object, keys***REMOVED***.(type***REMOVED*** {
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

// NewRand returns a rand with the time seed
func NewRand(***REMOVED*** *rand.Rand {
	return rand.New(rand.NewSource(time.Now(***REMOVED***.UnixNano(***REMOVED******REMOVED******REMOVED***
}

// DigStringArray tries to find an array inside the given object with the given path, and returns its
// value. If there is no attribute with the given path then the test will be aborted with an error.
func DigStringArray(object interface{}, keys ...interface{}***REMOVED*** []string {
	value := Dig(object, keys***REMOVED***
	gomega.ExpectWithOffset(1, value***REMOVED***.ToNot(gomega.BeNil(***REMOVED******REMOVED***
	var result []string
	gomega.ExpectWithOffset(1, value***REMOVED***.To(gomega.BeAssignableToTypeOf(result***REMOVED******REMOVED***
	result = value.([]string***REMOVED***
	return result
}

// DigArray tries to find an array inside the given object with the given path, and returns its
// value. If there is no attribute with the given path then the test will be aborted with an error.
func DigArray(object interface{}, keys ...interface{}***REMOVED*** []interface{} {
	value := Dig(object, keys***REMOVED***
	//ExpectWithOffset(1, value***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	var result []interface{}
	//ExpectWithOffset(1, value***REMOVED***.To(BeAssignableToTypeOf(result***REMOVED******REMOVED***
	result = value.([]interface{}***REMOVED***
	return result
}

// DigInt tries to find an attribute inside the given object with the given path, and returns its
// value, assuming that it is an integer. If there is no attribute with the given path then the test
// will be aborted with an error.
func DigInt(object interface{}, keys ...interface{}***REMOVED*** int {
	value := Dig(object, keys***REMOVED***
	ExpectWithOffset(1, value***REMOVED***.ToNot(BeNil(***REMOVED******REMOVED***
	var result float64
	ExpectWithOffset(1, value***REMOVED***.To(BeAssignableToTypeOf(result***REMOVED******REMOVED***
	result = value.(float64***REMOVED***
	return int(result***REMOVED***
}

func DigBool(object interface{}, keys ...interface{}***REMOVED*** bool {
	switch result := Dig(object, keys***REMOVED***.(type***REMOVED*** {
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

func Dig(object interface{}, keys []interface{}***REMOVED*** interface{} {
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
			return Dig(value, keys[1:]***REMOVED***
***REMOVED***
	case int:
		switch data := object.(type***REMOVED*** {
		case []interface{}:
			value := data[key]
			if len(keys***REMOVED*** == 1 {
				return value
	***REMOVED***
			return Dig(value, keys[1:]***REMOVED***
***REMOVED***
	}
	return nil
}

func RunCMD(cmd string***REMOVED*** (stdout string, stderr string, err error***REMOVED*** {
	fmt.Println("[>>] Running CMD: ", cmd***REMOVED***
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

// NeedFiltered will return the attribute that should be filtered
// filterList should be array with regex like ["excluded\..+","excluded_[\s\S]+"]
func NeedFiltered(filterList []string, key string***REMOVED*** bool {
	for _, regex := range filterList {
		pattern := regexp.MustCompile(regex***REMOVED***
		if pattern.MatchString(key***REMOVED*** {
			if key == "network.type" {
				fmt.Printf(">>>> network.type matched regex: %s\n", regex***REMOVED***
	***REMOVED***
			return true
***REMOVED***
	}
	return false
}

func BoolPoint(b bool***REMOVED*** *bool {
	boolVar := b
	return &boolVar
}

func subfix(***REMOVED*** string {
	subfix := make([]byte, 3***REMOVED***
	r := rand.New(rand.NewSource(time.Now(***REMOVED***.UnixNano(***REMOVED******REMOVED******REMOVED***
	for i := range subfix {
		subfix[i] = CON.CharsBytes[r.Intn(len(CON.CharsBytes***REMOVED******REMOVED***]
	}

	return string(subfix***REMOVED***
}

func GenerateClusterName(profileName string***REMOVED*** string {

	clusterPrefix := CON.RHCSPrefix + CON.HyphenConnector + profileName[5:]
	return clusterPrefix + CON.HyphenConnector + subfix(***REMOVED***
}
