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
}
