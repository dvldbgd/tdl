package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExtractComments reads a file line by line and extracts comments containing specified tags.
func ExtractComments(filePath string, tags string) ([]Comment, error) {
	// Determine file extension or fallback to filename for files like "Makefile"
	fileExtension := strings.ToLower(filepath.Ext(filePath))
	if fileExtension == "" {
		fileExtension = strings.ToLower(filepath.Base(filePath))
	}

	commentChar, ok := extensionToCommentChar[fileExtension]
	if !ok {
		return nil, fmt.Errorf("unsupported file extension: %s", fileExtension)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var extractedComments []Comment
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	tagsToCheck := parseTags(tags)

	for scanner.Scan() {
		lineNumber++
		lineText := scanner.Text() // preserve original spacing

		// Locate comment start in the line
		commentStartIndex := strings.Index(lineText, commentChar)
		if commentStartIndex == -1 {
			continue
		}

		commentText := strings.TrimSpace(lineText[commentStartIndex:]) // extract comment text only
		foundTag := findTag(commentText, tagsToCheck)
		if foundTag == "" {
			continue
		}

		extractedComments = append(extractedComments, Comment{
			Tag:        foundTag,
			Content:    commentText,
			FilePath:   filePath,
			LineNumber: lineNumber,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return extractedComments, nil
}

// parseTags converts a comma-separated string into a set of uppercase tags.
// If empty, it defaults to all SupportedTags.
func parseTags(tags string) map[string]struct{} {
	tagSet := make(map[string]struct{})
	if strings.TrimSpace(tags) == "" {
		for tag := range SupportedTags {
			tagSet[tag] = struct{}{}
		}
	} else {
		for _, tag := range strings.Split(tags, ",") {
			trimmedTag := strings.ToUpper(strings.TrimSpace(tag))
			if trimmedTag != "" {
				tagSet[trimmedTag] = struct{}{}
			}
		}
	}
	return tagSet
}

// findTag returns the first tag found in a comment, case-insensitive.
func findTag(commentText string, tags map[string]struct{}) string {
	upperComment := strings.ToUpper(commentText)
	for tag := range tags {
		if strings.Contains(upperComment, tag) {
			return tag
		}
	}
	return ""
}

