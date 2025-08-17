package core

import (
	"os"
	"path/filepath"
)

// GetAllFilePaths recursively collects all file paths under the given root directory.
func GetAllFilePaths(rootDir string) ([]string, error) {
	var collectedFiles []string
	err := filepath.WalkDir(rootDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			collectedFiles = append(collectedFiles, path)
		}
		return nil
	})
	return collectedFiles, err
}

