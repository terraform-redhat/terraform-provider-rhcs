package helper

import (
	"os"

	ini "gopkg.in/ini.v1"
)

// Create a file for usage
func TouchFile(filename string) (*os.File, error) {
	return os.Create(filename)
}

// Delete a file
func DeleteFile(filename string) error {
	return os.Remove(filename)
}

// IniConnection builds the connection of the ini file
func IniConnection(filename string) (*ini.File, error) {
	if _, err := os.Stat(filename); err != nil {
		_, err = TouchFile(filename)
		if err != nil {
			return nil, err
		}
	}
	return ini.Load(filename)
}
