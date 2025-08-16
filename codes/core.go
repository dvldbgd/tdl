package codes

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// singleLineCommentMap defines the mapping between comment characters and supported file extensions.
var singleLineCommentMap = map[string][]string{
	"//": {".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".swift", ".kt", ".rs", ".scala",
		".ts", ".js", ".jsx", ".tsx"},
	"#": {".py", ".rb", ".sh", ".bash", ".zsh", ".yml", ".yaml", ".toml", ".pl", ".pm", ".mk",
		"makefile", "dockerfile", ".ini"},
	";":   {".lisp", ".clj", ".scm", ".s", ".asm"},
	"--":  {".lua", ".hs", ".sql", ".adb"},
	"'":   {".vb", ".vbs"},
	".. ": {".rst"},
}

// SupportedTags defines the set of recognized comment tags.
var SupportedTags = map[string]struct{}{
	"TODO":      {},
	"FIXME":     {},
	"NOTE":      {},
	"HACK":      {},
	"BUG":       {},
	"OPTIMIZE":  {},
	"DEPRECATE": {},
}

// extensionToCommentChar maps a file extension to its comment character for quick lookup.
var extensionToCommentChar map[string]string

// Comment represents a tagged comment found in a file.
type Comment struct {
	Tag        string // The tag type (TODO, FIXME, etc.)
	Content    string // The full comment content including the tag
	FilePath   string // Absolute or relative path of the file
	LineNumber int    // Line number in the file where comment was found
}

// init initializes the extensionToCommentChar mapping.
func init() {
	extensionToCommentChar = make(map[string]string)
	for commentChar, extensions := range singleLineCommentMap {
		for _, extension := range extensions {
			extensionToCommentChar[strings.ToLower(extension)] = commentChar
		}
	}
}

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

// RunExtractCommentsConcurrently runs ExtractComments in parallel for multiple files.
func RunExtractCommentsConcurrently(files []string, maxWorkers int, tags string, ignoreErrors bool) map[string][]Comment {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if maxWorkers > len(files) {
		maxWorkers = len(files)
	}

	results := make(map[string][]Comment)
	var resultsMutex sync.Mutex
	var waitGroup sync.WaitGroup
	fileChannel := make(chan string)

	// Worker goroutine for extracting comments
	worker := func() {
		defer waitGroup.Done()
		for file := range fileChannel {
			extractedComments, err := ExtractComments(file, tags)
			if err != nil && !ignoreErrors {
				fmt.Printf("Error processing %s: %v\n", file, err)
			}
			if extractedComments != nil {
				resultsMutex.Lock()
				results[file] = extractedComments
				resultsMutex.Unlock()
			}
		}
	}

	waitGroup.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	// Feed files into the worker pool
	go func() {
		for _, filePath := range files {
			fileChannel <- filePath
		}
		close(fileChannel)
	}()

	waitGroup.Wait()
	return results
}

// PrettyPrintComments displays extracted comments in a readable format, optionally with colors.
func PrettyPrintComments(commentsMap map[string][]Comment, colorize bool) {
	const resetColor = "\033[0m"
	tagColors := map[string]string{
		"TODO":      "\033[33m",
		"FIXME":     "\033[31m",
		"NOTE":      "\033[36m",
		"HACK":      "\033[35m",
		"BUG":       "\033[91m",
		"OPTIMIZE":  "\033[32m",
		"DEPRECATE": "\033[90m",
	}

	for filePath, comments := range commentsMap {
		if colorize {
			fmt.Printf("%sFile: %s%s\n", "\033[36m", filePath, resetColor)
		} else {
			fmt.Printf("File: %s\n", filePath)
		}

		if len(comments) == 0 {
			fmt.Println("    No tagged comments found")
			fmt.Println("")
			continue
		}

		// Sort comments by line number
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].LineNumber < comments[j].LineNumber
		})

		for _, comment := range comments {
			lineNumberStr := fmt.Sprintf("%-5d", comment.LineNumber)
			if colorize {
				tagColor, ok := tagColors[comment.Tag]
				if !ok {
					tagColor = resetColor
				}
				fmt.Printf("    %s %s%s%s\n", lineNumberStr, tagColor, comment.Content, resetColor)
			} else {
				fmt.Printf("    %s %s\n", lineNumberStr, comment.Content)
			}
		}
		fmt.Println()
	}
}

