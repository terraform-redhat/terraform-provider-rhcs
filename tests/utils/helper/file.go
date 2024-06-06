package helper

import (
	"os"
)

// Delete a file
func DeleteFile(filename string) error {
	return os.Remove(filename)
}
