package utils

import "os"

// IsValidFolder checks if the provided path is a valid directory
func IsValidFolder(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
