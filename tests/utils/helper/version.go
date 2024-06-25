package helper

import "regexp"

func GetMajorVersion(rawVersion string) string {
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+`)
	vResults := versionRegex.FindAllStringSubmatch(rawVersion, 1)
	vResult := ""
	if len(vResults) != 0 {
		vResult = vResults[0][0]
	}
	return vResult
}
