// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

var awsRegionRegexFmt = "(?:af|ap|ca|eu|me|sa|us)(?:-gov)?-(?:central|north|(?:north(?:east|west))|south|south(?:east|west)|east|west)-\\d+(?:[a-z]{1})?"

func RemoveNoProxyZeroEgressDefaultDomains(
	input string, separator string, defaultDomains []string, awsAccountID string,
) string {
	domains := strings.Split(input, separator)
	defaultDomainsMap := make(map[string]string)
	for _, item := range defaultDomains {
		defaultDomainsMap[item] = ""
	}

	var deprecatedEcrRegex *regexp.Regexp
	if awsAccountID != "" {
		deprecatedEcrRegex = regexp.MustCompile(fmt.Sprintf(
			"^\\.?%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com$",
			regexp.QuoteMeta(awsAccountID),
			awsRegionRegexFmt,
		))
	}

	domains = slices.DeleteFunc(domains, func(item string) bool {
		// Check exact match first
		if _, exists := defaultDomainsMap[item]; exists {
			return true
		}
		// Check deprecated zero egress default domains for backward compatibility
		if deprecatedEcrRegex != nil && deprecatedEcrRegex.MatchString(item) {
			return true
		}
		return false
	})
	return strings.Join(domains, separator)
}

func Contains[T comparable](slice []T, element T) bool {
	for _, sliceElement := range slice {
		if reflect.DeepEqual(sliceElement, element) {
			return true
		}
	}

	return false
}
