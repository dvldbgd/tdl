package core

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
	"//": {
		".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".swift", ".kt", ".rs", ".scala",
		".ts", ".js", ".jsx", ".tsx",
	},
	"#": {
		".py", ".rb", ".sh", ".bash", ".zsh", ".yml", ".yaml", ".toml", ".pl", ".pm", ".mk",
		"makefile", "dockerfile", ".ini",
	},
	";": {
		".lisp", ".clj", ".scm", ".s", ".asm",
	},
	"--": {
		".lua", ".hs", ".sql", ".adb",
	},
	"'": {
		".vb", ".vbs",
	},
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

// init initializes the extensionToCommentChar mapping for fast extension lookup.
func init() {
	extensionToCommentChar = make(map[string]string)
	for commentChar, extensions := range singleLineCommentMap {
		for _, extension := range extensions {
			extensionToCommentChar[strings.ToLower(extension)] = commentChar
		}
	}
}

// RunExtractCommentsConcurrently runs ExtractComments in parallel for multiple files using a worker pool.
func RunExtractCommentsConcurrently(
	files []string,
	maxWorkers int,
	tags string,
	ignoreErrors bool,
) map[string][]Comment {
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

	// Worker goroutine: processes files from channel and extracts comments.
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

	// Launch workers
	waitGroup.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	// Feed file paths into the channel
	go func() {
		for _, filePath := range files {
			fileChannel <- filePath
		}
		close(fileChannel)
	}()

	waitGroup.Wait()
	return results
}

// ExtractComments reads a file line by line and extracts comments containing specified tags.
func ExtractComments(filePath string, tags string) ([]Comment, error) {
	// Determine file extension or fallback to filename for cases like "Makefile"
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

	// Increase scanner buffer to handle very long lines safely (default 64KB)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	lineNumber := 0
	tagsToCheck := parseTags(tags)

	for scanner.Scan() {
		lineNumber++
		lineText := scanner.Text()

		// Find the comment start marker in the line
		commentStartIndex := strings.Index(lineText, commentChar)
		if commentStartIndex == -1 {
			continue
		}

		// Extract the comment substring
		commentText := strings.TrimSpace(lineText[commentStartIndex:])
		foundTag := findTag(commentText, tagsToCheck)
		if foundTag == "" {
			continue
		}

		// Append comment metadata to results
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
// If empty, defaults to all supported tags.
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

// isBinaryFile checks the first few KB of a file for NUL bytes to detect binaries.
func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false // fail open → treat as text, caller decides
	}
	defer f.Close()

	buf := make([]byte, 8000) // read up to 8KB
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

// GetAllFilePaths recursively collects supported file paths under the given root directory.
// It filters by known extensions and skips binary files.
func GetAllFilePaths(rootDir string) ([]string, error) {
	var collectedFiles []string
	err := filepath.WalkDir(rootDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		// Match extension or special filename (e.g., Makefile)
		ext := strings.ToLower(filepath.Ext(path))
		if ext == "" {
			ext = strings.ToLower(filepath.Base(path))
		}
		if _, ok := extensionToCommentChar[ext]; !ok {
			return nil // unsupported extension → skip
		}

		// Skip binary files
		if isBinaryFile(path) {
			return nil
		}

		collectedFiles = append(collectedFiles, path)
		return nil
	})
	return collectedFiles, err
}

// PrettyPrintComments displays extracted comments in a formatted way, optionally colorized.
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

		// Sort comments by line number before printing
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

// Comment represents a tagged comment found in a file.
type Comment struct {
	Tag        string // Tag type (TODO, FIXME, etc.)
	Content    string // Full comment text
	FilePath   string // File path
	LineNumber int    // Line number in the file
}

