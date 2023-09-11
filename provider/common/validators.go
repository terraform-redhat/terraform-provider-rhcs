package common

***REMOVED***
***REMOVED***
	"regexp"
	"strings"
***REMOVED***

func ValidateHTPasswdUsername(username string***REMOVED*** error {
	if strings.ContainsAny(username, "/:%"***REMOVED*** {
		return fmt.Errorf("invalid username '%s': "+
			"username must not contain /, :, or %%", username***REMOVED***
	}
	return nil
}

func ValidateHTPasswdPassword(password string***REMOVED*** error {
	notAsciiOnly, _ := regexp.MatchString(`[^\x20-\x7E]`, password***REMOVED***
	containsSpace := strings.Contains(password, " "***REMOVED***
	tooShort := len(password***REMOVED*** < 14
	if notAsciiOnly || containsSpace || tooShort {
		return fmt.Errorf(
			"password must be at least 14 characters (ASCII-standard***REMOVED*** without whitespaces"***REMOVED***
	}
	hasUppercase, _ := regexp.MatchString(`[A-Z]`, password***REMOVED***
	hasLowercase, _ := regexp.MatchString(`[a-z]`, password***REMOVED***
	hasNumberOrSymbol, _ := regexp.MatchString(`[^a-zA-Z]`, password***REMOVED***
	if !hasUppercase || !hasLowercase || !hasNumberOrSymbol {
		return fmt.Errorf(
			"password must include uppercase letters, lowercase letters, and numbers " +
				"or symbols (ASCII-standard characters only***REMOVED***"***REMOVED***
	}
	return nil
}
