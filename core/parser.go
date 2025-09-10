package core

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ExtractComments scans one file line by line for tagged comments.
func ExtractComments(filePath, tags string) ([]Comment, error) {
	// Resolve extension or basename (e.g. Makefile)
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		ext = strings.ToLower(filepath.Base(filePath))
	}

	char, ok := extensionToChar[ext]
	if !ok {
		return nil, nil // unsupported file type
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tagSet := parseTags(tags)
	sc := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	sc.Buffer(buf, maxScanCapacity)

	var out []Comment
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := sc.Text()
		pos := strings.Index(line, char)
		if pos == -1 {
			continue // skip lines without comment delimiter
		}
		text := strings.TrimSpace(line[pos+len(char):])

		// If the comment contains a supported tag, capture it
		if tag := findTag(text, tagSet); tag != "" {
			commit, author, stamp, _ := fetchGitBlameInfo(filePath, lineNum)
			out = append(out, Comment{
				Tag:           tag,
				Content:       text,
				FilePath:      filePath,
				LineNumber:    lineNum,
				Commit:        commit,
				Author:        author,
				CreationStamp: stamp,
			})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// parseTags converts a comma-separated string into a lookup map of tags.
func parseTags(tags string) map[string]struct{} {
	if strings.TrimSpace(tags) == "" {
		// If no tags provided, use default SupportedTags
		set := make(map[string]struct{}, len(supportedTagsLookup))
		for tag := range supportedTagsLookup {
			set[tag] = struct{}{}
		}
		return set
	}

	parts := strings.Split(tags, ",")
	set := make(map[string]struct{}, len(parts))
	for _, t := range parts {
		trimmed := strings.ToUpper(strings.TrimSpace(t))
		if trimmed != "" {
			set[trimmed] = struct{}{}
		}
	}
	return set
}

// findTag checks if any supported tag exists in the given text.
func findTag(text string, tags map[string]struct{}) string {
	upper := strings.ToUpper(text)
	for t := range tags {
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(t) + `\b`)
		if re.MatchString(upper) {
			return t
		}
	}
	return ""
}

