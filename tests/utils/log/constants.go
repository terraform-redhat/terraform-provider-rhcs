// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"regexp"
)

const (
	RedactValue = "XXXXXXXX"
)

var RedactKeyList = []*regexp.Regexp{
	regexp.MustCompile(`(\\?"password\\?":\\?")([^"]*)(\\?")`),
	regexp.MustCompile(`(\\?"additional_trust_bundle\\?":\\?")([^"]*)(\\?")`),
	regexp.MustCompile(`(-----BEGIN CERTIFICATE-----)([^-----]*)(-----END CERTIFICATE-----)`),
	regexp.MustCompile(`(password\s*=\s*)([^\n\\\n]+)([\n\\\n]+)`),
	regexp.MustCompile(`(aws_(billing)?_?account_id[\s]*=[\s\\]*"?)([0-9]{12})([\\"]*)`),
	regexp.MustCompile(`(arn:aws:[^:]*:[a-z0-9-]*:)([0-9]{12})([^\n\"\\]*)`),
	// Account IDs in network interface diagnostics, such as owner_id="123456789012".
	// A value containing 12 consecutive digits is replaced entirely; other values,
	// such as service aliases or empty values, are kept.
	regexp.MustCompile(`(\b(?:instance_owner|owner|requester)_id=)("[^"]*[0-9]{12}[^"]*)(")`),
}
