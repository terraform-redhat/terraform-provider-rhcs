/*
Copyright (c***REMOVED*** 2021 Red Hat, Inc.

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

package tests

***REMOVED***
	"bytes"
	"encoding/json"
***REMOVED***
	"reflect"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/onsi/gomega/types"
***REMOVED***

// JQ runs the given `jq` filter on the given object and returns the list of results. The returned
// slice will never be nil; if there are no results it will be empty.
func JQ(filter string, input interface{}***REMOVED*** (results []interface{}, err error***REMOVED*** {
	query, err := gojq.Parse(filter***REMOVED***
	if err != nil {
		return
	}
	iterator := query.Run(input***REMOVED***
	for {
		result, ok := iterator.Next(***REMOVED***
		if !ok {
			break
***REMOVED***
		results = append(results, result***REMOVED***
	}
	return
}

// MatchJQ creates a matcher that checks that the all the results of applying a `jq` filter to the
// actual value is the given expected value.
func MatchJQ(filter string, expected interface{}***REMOVED*** types.GomegaMatcher {
	return &jqMatcher{
		filter:   filter,
		expected: expected,
	}
}

type jqMatcher struct {
	filter   string
	expected interface{}
	results  []interface{}
}

func (m *jqMatcher***REMOVED*** Match(actual interface{}***REMOVED*** (success bool, err error***REMOVED*** {
	// Run the query:
	m.results, err = JQ(m.filter, actual***REMOVED***
	if err != nil {
		return
	}

	// We consider the match sucessful if all the results returned by the JQ filter are exactly
	// equal to the expected value.
	success = true
	for _, result := range m.results {
		if !reflect.DeepEqual(result, m.expected***REMOVED*** {
			success = false
			break
***REMOVED***
	}
	return
}

func (m *jqMatcher***REMOVED*** FailureMessage(actual interface{}***REMOVED*** string {
	return fmt.Sprintf(
		"Expected all results of running JQ filter\n\t%s\n"+
			"on input\n\t%s\n"+
			"to be\n\t%s\n"+
			"but at list one of the following results isn't\n\t%s\n",
		m.filter, m.pretty(actual***REMOVED***, m.pretty(m.expected***REMOVED***, m.pretty(m.results***REMOVED***,
	***REMOVED***
}

func (m *jqMatcher***REMOVED*** NegatedFailureMessage(actual interface{}***REMOVED*** (message string***REMOVED*** {
	return fmt.Sprintf(
		"Expected results of running JQ filter\n\t%s\n"+
			"on\n\t%s\n"+
			"to not be\n\t%s\n",
		m.filter, m.pretty(actual***REMOVED***, m.pretty(m.expected***REMOVED***,
	***REMOVED***
}

func (m *jqMatcher***REMOVED*** pretty(object interface{}***REMOVED*** string {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer***REMOVED***
	encoder.SetIndent("\t", "  "***REMOVED***
	err := encoder.Encode(object***REMOVED***
	if err != nil {
		return fmt.Sprintf("\t%v", object***REMOVED***
	}
	return strings.TrimRight(buffer.String(***REMOVED***, "\n"***REMOVED***
}
