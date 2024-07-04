package helper

import (
	"os"
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
