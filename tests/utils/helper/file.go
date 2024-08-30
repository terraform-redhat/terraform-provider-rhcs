package helper

import (
	"encoding/json"
	"os"

	. "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/log"
)

// Delete a file
func DeleteFile(filename string) error {
	return os.Remove(filename)
}

func IsFileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func CreateTempFileWithContent(fileContent string) (string, error) {
	return CreateTempFileWithPrefixAndContent("tmpfile", fileContent)
}

func CreateTempFileWithPrefixAndContent(prefix string, fileContent string) (string, error) {
	f, err := os.CreateTemp("", prefix+"-")
	if err != nil {
		return "", err
	}
	return CreateFileWithContent(f.Name(), fileContent)
}

// Write string to a file
func CreateFileWithContent(fileAbsPath string, content interface{}) (string, error) {
	var err error
	switch content := content.(type) {
	case string:
		err = os.WriteFile(fileAbsPath, []byte(content), 0644) // #nosec G306
	case []byte:
		err = os.WriteFile(fileAbsPath, content, 0644) // #nosec G306
	case interface{}:
		var marshedContent []byte
		marshedContent, err = json.Marshal(content)
		if err != nil {
			return fileAbsPath, err
		}
		err = os.WriteFile(fileAbsPath, marshedContent, 0644) // #nosec G306
	}

	if err != nil {
		Logger.Errorf("Failed to write to file: %s", err)
		return "", err
	}
	return fileAbsPath, err
}
