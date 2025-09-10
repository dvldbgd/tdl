package core

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// isBinaryFile checks for null bytes to decide if a file is binary.
func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	return slices.Contains(buf[:n], byte(0))
}

// GetAllFilePaths walks a directory tree and returns all supported text files.
func GetAllFilePaths(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = strings.ToLower(filepath.Base(path))
		}
		// Skip unsupported or binary files
		if _, ok := extensionToChar[ext]; !ok || isBinaryFile(path) {
			return nil
		}
		out = append(out, path)
		return nil
	})
	return out, err
}

