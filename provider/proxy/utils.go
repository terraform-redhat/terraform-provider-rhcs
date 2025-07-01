package proxy

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var awsRegionRegexFmt = "(?:af|ap|ca|eu|me|sa|us)(?:-gov)?-(?:central|north|(?:north(?:east|west))|south|south(?:east|west)|east|west)-\\d+(?:[a-z]{1})?"
var awsAccountIdRegexFmt = "\\d{12}"
var zeroEgressDefaultDomainFmts = []*regexp.Regexp{
	regexp.MustCompile(fmt.Sprintf("s3.dualstack.%s.amazonaws.com", awsRegionRegexFmt)),
	regexp.MustCompile(fmt.Sprintf("sts.%s.amazonaws.com", awsRegionRegexFmt)),
	regexp.MustCompile(fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com", awsAccountIdRegexFmt, awsRegionRegexFmt)),
}

func RemoveNoProxyZeroEgressDefaultDomains(input string, separator string) string {
	splits := strings.Split(input, separator)
	result := []string{}
	for _, split := range splits {
		var matched bool
		for _, format := range zeroEgressDefaultDomainFmts {
			matched = format.Match([]byte(split))
			if matched {
				break
			}
		}
		if !matched {
			result = append(result, split)
		}
	}
	return strings.Join(result, ",")
}

func Contains[T comparable](slice []T, element T) bool {
	for _, sliceElement := range slice {
		if reflect.DeepEqual(sliceElement, element) {
			return true
		}
	}

	return false
}
