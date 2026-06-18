// Copyright Red Hat
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/config"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

func ModuleSourceAlignment(manifestsFilePath string, ref string) error {
	regexpP := regexp.MustCompile(
		`"(git::https:\/\/github\.com\/terraform-redhat\/terraform-rhcs-rosa-(classic|hcp)\/\/modules\/[^?]+\?ref=)[^"]+"`,
	)
	var err error
	if ref != "" {
		err = ReplaceRegex(manifestsFilePath, *regexpP, fmt.Sprintf(`"$1%s"`, ref))
	}
	return err
}

func RHCSSourceAlignment(manifestsFilePath string, rhcsSource string, rhcsVersion string) error {
	sourceRegexrP := regexp.MustCompile(`(rhcs\s*=\s*{(?:\n*[\w\W]*?source\s*=\s*))"([a-z,A-Z,\/,\-,\.]*)("[^}]*})`)
	versionRegexP := regexp.MustCompile(`(rhcs\s*=\s*{(?:\n*[\w\W]*?version\s*=\s*))"([0-9.,=\-><\sA-Za-z]*)("[^}]*})`)
	var err error
	if rhcsSource != "" {
		err = ReplaceRegex(manifestsFilePath, *sourceRegexrP, fmt.Sprintf(`$1"%s$3`, rhcsSource))
		if err != nil {
			return err
		}

	}
	if rhcsVersion != "" {
		err = ReplaceRegex(manifestsFilePath, *versionRegexP, fmt.Sprintf(`$1"%s$3`, rhcsVersion))
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
		replacedContent := regexP.ReplaceAllString(stringContent, replaceString)
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
	var foundEnv bool
	retrieveRHCSEnvVar := func(name string, retriever func() string) string {
		value := retriever()
		if value != "" {
			foundEnv = true
			Logger.Warnf("Got a global ENV variable for %s set to %s. Going to replace all of the manifests files", name, value)
		}
		return value
	}

	rhcsSource := retrieveRHCSEnvVar("RHCS Source", config.GetRHCSSource)
	rhcsVersion := retrieveRHCSEnvVar("RHCS Version", config.GetRHCSVersion)
	rhcsModuleRef := retrieveRHCSEnvVar("RHCS Module Ref", config.GetRHCSModuleRef)
	if !foundEnv {
		return nil
	}

	Logger.Warnf("RHCS Source: %s, RHCS Version: %s", rhcsSource, rhcsVersion)
	Logger.Warnf("Module ref: %s", rhcsModuleRef)
	files, err := ScanManifestsDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		err := RHCSSourceAlignment(file, rhcsSource, rhcsVersion)
		if err != nil {
			return err
		}
		err = ModuleSourceAlignment(file, rhcsModuleRef)
		if err != nil {
			return err
		}
	}
	return nil
}
