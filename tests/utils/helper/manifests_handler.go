package helper

import (
	"fmt"
	"os"
	"regexp"

	"path/filepath"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

func ModuleSourceAlignment(manifestsFilePath string, source string, sourceLocal string, version string) error {
	regexpP := regexp.MustCompile(`(source\s*=\s*")(terraform-redhat/rosa-classic/rhcs/)(.*"\s*\n)(\s*version\s*=\s*")(.*)"`)
	var err error
	if source != "" {
		// err = ReplaceRegex(manifestsFilePath, *regexpP, fmt.Sprintf(`$1%s$3$4$5"`, source))
		if sourceLocal != "" {
			err = ReplaceRegex(manifestsFilePath, *regexpP, fmt.Sprintf(`source = "%s"`, source))
		} else {
			err = ReplaceRegex(manifestsFilePath, *regexpP, fmt.Sprintf(`source = "%s$3$4$5`, source))
		}
		if err != nil {
			return err
		}
	}
	if version != "" && sourceLocal == "" {
		err = ReplaceRegex(manifestsFilePath, *regexpP, fmt.Sprintf(`source = "$2$3  version = "%s"`, version))
		if err != nil {
			return err
		}
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
	rhcsSource := os.Getenv(CON.RHCSSource)
	rhcsVersion := os.Getenv(CON.RHCSVersion)
	moduleSource := os.Getenv(CON.ModuleSource)
	moduleSourceLocal := os.Getenv(CON.ModuleSourceLocal)
	moduleVersion := os.Getenv(CON.ModuleVersion)
	if rhcsSource == "" && rhcsVersion == "" && moduleSource == "" && moduleVersion == "" && moduleSourceLocal == "" {
		return nil
	}
	if rhcsSource != "" {
		Logger.Warnf("Got a global ENV variable %s set to %s. Going to replace all of the manifests files", CON.RHCSSource, rhcsSource)
	}
	if rhcsVersion != "" {
		Logger.Warnf("Got a global ENV variable %s set to %s. Going to replace all of the manifests files", CON.RHCSVersion, rhcsVersion)
	}
	if moduleSource != "" {
		Logger.Warnf("Got a global ENV variable %s set to %s. Going to replace all of the manifests files", CON.ModuleSource, moduleSource)
	}
	if moduleSourceLocal != "" {
		Logger.Warnf("Got a global ENV variable %s set. Going to replace all of the manifests files", CON.ModuleVersion, moduleVersion)
	}
	if moduleVersion != "" {
		Logger.Warnf("Got a global ENV variable %s set to %s. Going to replace all of the manifests files", CON.ModuleVersion, moduleVersion)
	}
	Logger.Warnf("Source: %s, Version: %s", moduleSource, moduleVersion)
	files, err := ScanManifestsDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		err := RHCSSourceAlignment(file, rhcsSource, rhcsVersion)
		if err != nil {
			return err
		}
		err = ModuleSourceAlignment(file, moduleSource, moduleSourceLocal, moduleVersion)
		if err != nil {
			return err
		}
	}
	return nil
}
