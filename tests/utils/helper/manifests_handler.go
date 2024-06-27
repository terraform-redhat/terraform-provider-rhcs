package helper

import (
	"fmt"
	"os"
	"regexp"

	"path/filepath"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

func RHCSSourceAlignment(manifestsFilePath string, rhcsSource string, rhcsVersion string) error {
	sourceRegexrP := regexp.MustCompile(`(rhcs\s*=\s*{(?:\n*[\w\W]*?source\s*=\s*))"([a-z,A-Z,\/,\-,\.]*)("[^}]*})`)
	versionRegexP := regexp.MustCompile(`(rhcs\s*=\s*{(?:\n*[\w\W]*?version\s*=\s*))"([0-9.,=\-><\sA-Za-z]*)("[^}]*})`)
	var err error
	if rhcsSource != "" {
		err = ReplaceRegex(manifestsFilePath, *sourceRegexrP, rhcsSource)
		if err != nil {
			return err
		}

	}
	if rhcsVersion != "" {
		err = ReplaceRegex(manifestsFilePath, *versionRegexP, rhcsVersion)
	}

	return err
}
func ReplaceRegex(filePath string, regexP regexp.Regexp, replaceString string) error {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	stringContent := string(fileContent)
	if regexP.MatchString(stringContent) {
		Logger.Debugf("Find match string in file %s, going to replace", filePath)
		replacedContent := regexP.ReplaceAllString(stringContent, fmt.Sprintf(`$1"%s$3`, replaceString))
		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.WriteString(replacedContent)
		if err != nil {
			return err
		}
		Logger.Debugf("Replaced file %s with %s", filePath, replaceString)
	}
	return nil
}

func ScanManifestsDir(dir string) ([]string, error) {
	files := []string{}
	manifestR := regexp.MustCompile(`[\w\W]*\.+tf`)
	err := filepath.Walk(dir, func(filepath string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			if manifestR.MatchString(info.Name()) {
				files = append(files, filepath)
			}
		}
		return err
	})
	return files, err
}

func AlignRHCSSourceVersion(dir string) error {
	rhcsSource := os.Getenv(CON.RHCSSource)
	rhcsVersion := os.Getenv(CON.RHCSVersion)
	if rhcsSource == "" && rhcsVersion == "" {
		return nil
	}
	if rhcsSource != "" {
		Logger.Warnf("Got a global ENV variable %s set to %s. Going to replace all of the manifests files", CON.RHCSSource, rhcsSource)
	}
	if rhcsVersion != "" {
		Logger.Warnf("Got a global ENV variable %s set to %s. Going to replace all of the manifests files", CON.RHCSVersion, rhcsVersion)
	}
	files, err := ScanManifestsDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		err := RHCSSourceAlignment(file, rhcsSource, rhcsVersion)
		if err != nil {
			return err
		}
	}
	return nil
}
