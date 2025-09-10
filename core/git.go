package core

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// fetchGitBlameInfo gets commit, author, and timestamp for a given file+line.
// It runs: git blame -L <line>,<line> --porcelain -- <file>
func fetchGitBlameInfo(filePath string, line int) (commit, author, stamp string, err error) {
	cmd := exec.Command("git", "blame", "-L", fmt.Sprintf("%d,%d", line, line),
		"--porcelain", "--", filePath)
	out, err := cmd.Output()
	if err != nil {
		return "", "", "", err
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", "", "", fmt.Errorf("empty blame output")
	}

	// First line contains commit hash
	parts := strings.Fields(lines[0])
	if len(parts) > 0 {
		commit = parts[0]
	}

	// Parse metadata: author and author-time
	for _, l := range lines {
		if strings.HasPrefix(l, "author ") {
			author = strings.TrimPrefix(l, "author ")
		}
		if strings.HasPrefix(l, "author-time ") {
			// Convert UNIX timestamp -> RFC3339 format
			ts := strings.TrimPrefix(l, "author-time ")
			i, err := strconv.ParseInt(ts, 10, 64)
			if err == nil {
				stamp = time.Unix(i, 0).Format(time.RFC3339)
			}
		}
	}

	return commit, author, stamp, nil
}

