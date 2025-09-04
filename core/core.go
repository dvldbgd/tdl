package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Comment represents one tagged comment (TODO/FIXME/etc.) found in a source file.
// It keeps the tag, the comment content, its location, and Git blame metadata.
type Comment struct {
	Tag           string `json:"tag" yaml:"tag"`         // The tag (TODO, FIXME, etc.)
	Content       string `json:"content" yaml:"content"` // The full comment text
	FilePath      string `json:"file" yaml:"file"`       // Path to the file containing this comment
	LineNumber    int    `json:"line" yaml:"line"`       // Line number in the file
	CreationStamp string `json:"stamp" yaml:"stamp"`     // RFC3339 timestamp from Git blame
	Author        string `json:"author" yaml:"author"`   // Author of the commit that introduced this line
	Commit        string `json:"commit" yaml:"commit"`   // Commit hash from Git blame
}

var (
	// Maps single-line comment delimiters to file extensions that use them.
	// This lets the scanner know how to detect comments in different languages.
	singleLineCommentMap = map[string][]string{
		"//": {
			".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".swift", ".kt", ".rs", ".scala",
			".ts", ".js", ".jsx", ".tsx",
		},
		"#": {
			".py", ".rb", ".sh", ".bash", ".zsh", ".yml", ".yaml", ".toml", ".pl", ".pm", ".mk",
			"makefile", "dockerfile", ".ini",
		},
		";":  {".lisp", ".clj", ".scm", ".s", ".asm"},
		"--": {".lua", ".hs", ".sql", ".adb"},
		"'":  {".vb", ".vbs"},
		"..": {".rst"},
	}

	// List of supported tags to detect
	SupportedTags       = []string{"TODO", "FIXME", "NOTE", "HACK", "BUG", "OPTIMIZE", "DEPRECATE"}
	supportedTagsLookup = make(map[string]struct{}) // fast lookup map of tags
	extensionToChar     = make(map[string]string)   // maps file extension -> comment delimiter

	// Scanner buffer limit (1MB)
	maxScanCapacity = 1024 * 1024
)

// init builds lookup maps for faster scanning at runtime.
func init() {
	// Build set of supported tags for O(1) lookups.
	for _, tag := range SupportedTags {
		supportedTagsLookup[tag] = struct{}{}
	}
	// Map each extension to its comment delimiter (e.g. ".go" -> "//")
	for ch, exts := range singleLineCommentMap {
		for _, e := range exts {
			extensionToChar[strings.ToLower(e)] = ch
		}
	}
}

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

// RunExtractCommentsConcurrently processes multiple files in parallel.
// Uses worker goroutines to avoid bottlenecks on large repos.
func RunExtractCommentsConcurrently(
	files []string, maxWorkers int, tags string, ignoreErrors bool,
) map[string][]Comment {
	results := make(map[string][]Comment)
	if len(files) == 0 {
		return results
	}

	// Default worker count to CPU cores
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if maxWorkers > len(files) {
		maxWorkers = len(files) // don’t spawn more workers than files
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	ch := make(chan string, len(files)) // buffered channel holds all files

	// Worker: consumes file paths, extracts comments, stores them in results
	worker := func() {
		defer wg.Done()
		for file := range ch {
			cmts, err := ExtractComments(file, tags)
			if err != nil && !ignoreErrors {
				fmt.Printf("Error processing %s: %v\n", file, err)
			}
			if len(cmts) > 0 {
				mu.Lock()
				results[file] = cmts
				mu.Unlock()
			}
		}
	}

	// Start workers
	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	// Send all files into channel
	for _, f := range files {
		ch <- f
	}
	close(ch)
	wg.Wait()

	return results
}

// PrepareOutputFile saves results to disk in JSON, YAML, or plain text format.
func PrepareOutputFile(results map[string][]Comment, format, outputDir string) error {
	var all []Comment
	for _, list := range results {
		all = append(all, list...)
	}
	if len(all) == 0 {
		return fmt.Errorf("no comments found")
	}

	ext := strings.ToLower(format)
	fileName := "comments." + ext
	outPath := filepath.Join(outputDir, fileName)

	// Create output directory if it doesn’t exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write in requested format
	switch ext {
	case "json":
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(all); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
	case "yaml", "yml":
		enc := yaml.NewEncoder(f)
		defer enc.Close()
		if err := enc.Encode(all); err != nil {
			return fmt.Errorf("failed to write YAML: %w", err)
		}
	case "text", "txt":
		for _, c := range all {
			if _, err := fmt.Fprintf(f, "%s:%d [%s] %s\n",
				c.FilePath, c.LineNumber, c.Tag, c.Content); err != nil {
				return fmt.Errorf("failed to write text: %w", err)
			}
		}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	fmt.Printf("Extracted %d comments written to %s\n", len(all), outPath)
	return nil
}

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

// PrettyPrintComments outputs results to stdout with optional ANSI colors.
func PrettyPrintComments(m map[string][]Comment, color bool) {
	const reset = "\033[0m"
	colors := map[string]string{
		"TODO":      "\033[33m", // yellow
		"FIXME":     "\033[31m", // red
		"NOTE":      "\033[36m", // cyan
		"HACK":      "\033[35m", // magenta
		"BUG":       "\033[91m", // bright red
		"OPTIMIZE":  "\033[32m", // green
		"DEPRECATE": "\033[90m", // grey
	}

	files := make([]string, 0, len(m))
	for f := range m {
		files = append(files, f)
	}
	sort.Strings(files)

	for _, file := range files {
		list := m[file]
		if color {
			fmt.Printf("\033[36mFile: %s%s\n", file, reset)
		} else {
			fmt.Printf("File: %s\n", file)
		}
		if len(list) == 0 {
			fmt.Println("    No tagged comments found")
			continue
		}
		// Sort comments in file by line number
		sort.Slice(list, func(i, j int) bool { return list[i].LineNumber < list[j].LineNumber })
		for _, c := range list {
			line := fmt.Sprintf("%-5d", c.LineNumber)
			if color {
				col, ok := colors[c.Tag]
				if !ok {
					col = reset
				}
				fmt.Printf("    %s %s%s%s\n", line, col, c.Content, reset)
			} else {
				fmt.Printf("    %s %s\n", line, c.Content)
			}
		}
		fmt.Println()
	}
}

