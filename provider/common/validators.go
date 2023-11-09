package common

import (
	"fmt"
	"regexp"
	"strings"
)

func ValidateHTPasswdUsername(username string) error {
	if strings.ContainsAny(username, "/:%") {
		return fmt.Errorf("invalid username '%s': "+
			"username must not contain /, :, or %%", username)
	}
	return nil
}

func ValidateHTPasswdPassword(password string) error {
	notAsciiOnly, _ := regexp.MatchString(`[^\x20-\x7E]`, password)
	containsSpace := strings.Contains(password, " ")
	tooShort := len(password) < 14
	if notAsciiOnly || containsSpace || tooShort {
		return fmt.Errorf(
			"password must be at least 14 characters (ASCII-standard) without whitespaces")
	}
	hasUppercase, _ := regexp.MatchString(`[A-Z]`, password)
	hasLowercase, _ := regexp.MatchString(`[a-z]`, password)
	hasNumberOrSymbol, _ := regexp.MatchString(`[^a-zA-Z]`, password)
	if !hasUppercase || !hasLowercase || !hasNumberOrSymbol {
		return fmt.Errorf(
			"password must include uppercase letters, lowercase letters, and numbers " +
				"or symbols (ASCII-standard characters only)")
	}
	return nil
}
