// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package seaplane

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/itchyny/gojq"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func verifyJQ(filter string, expected any) http.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		Expect(err).To(Succeed())
		Expect(req.Body.Close()).To(Succeed())
		req.Body = io.NopCloser(bytes.NewReader(body))

		var input any
		Expect(json.Unmarshal(body, &input)).To(Succeed())
		results, err := runJQ(filter, input)
		Expect(err).NotTo(HaveOccurred())
		Expect(results).NotTo(BeEmpty(), "jq filter %q returned no results", filter)
		for _, result := range results {
			Expect(result).To(Equal(expected))
		}
	}
}

func matchJQ(filter string, expected any) types.GomegaMatcher {
	return &jqMatcher{filter: filter, expected: expected}
}

type jqMatcher struct {
	filter   string
	expected any
	results  []any
}

func (m *jqMatcher) Match(actual any) (bool, error) {
	results, err := runJQ(m.filter, actual)
	if err != nil {
		return false, err
	}
	m.results = results
	if len(results) == 0 {
		return false, nil
	}
	for _, result := range results {
		if !reflect.DeepEqual(result, m.expected) {
			return false, nil
		}
	}
	return true, nil
}

func (m *jqMatcher) FailureMessage(actual any) string {
	return fmt.Sprintf("expected JQ filter %q on %v to return %v, got %v", m.filter, actual, m.expected, m.results)
}

func (m *jqMatcher) NegatedFailureMessage(actual any) string {
	return fmt.Sprintf("expected JQ filter %q on %v not to return %v", m.filter, actual, m.expected)
}

func runJQ(filter string, input any) ([]any, error) {
	query, err := gojq.Parse(filter)
	if err != nil {
		return nil, err
	}
	iterator := query.Run(input)
	var results []any
	for {
		result, ok := iterator.Next()
		if !ok {
			return results, nil
		}
		if err, isError := result.(error); isError {
			return nil, err
		}
		results = append(results, result)
	}
}
