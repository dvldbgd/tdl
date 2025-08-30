package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Comment is the struct we use to store tagged comments found in source files.
// Keeps track of tag, content, file path, and line number.
type Comment struct {
	Tag        string `json:"tag" yaml:"tag"`
	Content    string `json:"content" yaml:"content"`
	FilePath   string `json:"file" yaml:"file"`
	LineNumber int    `json:"line" yaml:"line"`
}

var (
	// Mapping of single-line comment syntax to file extensions that use it
	singleLineCommentMap = map[string][]string{
		"//": {
			".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".swift", ".kt", ".rs", ".scala",
			".ts", ".js", ".jsx", ".tsx",
		},
		"#": {
			"py", ".rb", ".sh", ".bash", ".zsh", ".yml", ".yaml", ".toml", ".pl", ".pm", ".mk",
			"makefile", "dockerfile", ".ini",
		},
		";":  {".lisp", ".clj", ".scm", ".s", ".asm"},
		"--": {".lua", ".hs", ".sql", ".adb"},
		"'":  {".vb", ".vbs"},
		"..": {".rst"},
	}

	// Tags we care about scanning for
	SupportedTags       = []string{"TODO", "FIXME", "NOTE", "HACK", "BUG", "OPTIMIZE", "DEPRECATE"}
	supportedTagsLookup = make(map[string]struct{}) // quick lookup
	extensionToChar     = make(map[string]string)   // map extension to comment char

	maxScanCapacity = 1024 * 1024 // scanner buffer limit, 1MB
)

// init precomputes lookup maps for tags and file extensions to comment chars
func init() {
	for _, tag := range SupportedTags {
		supportedTagsLookup[tag] = struct{}{}
	}
	for ch, exts := range singleLineCommentMap {
		for _, e := range exts {
			extensionToChar[strings.ToLower(e)] = ch
		}
	}
}

// RunExtractCommentsConcurrently scans multiple files at once.
// It spins up maxWorkers goroutines and distributes files to them.
// Results are stored in a map keyed by file path.
func RunExtractCommentsConcurrently(
	files []string, maxWorkers int, tags string, ignoreErrors bool,
) map[string][]Comment {
	results := make(map[string][]Comment)
	if len(files) == 0 {
		return results
	}

	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU() // default to number of CPU cores
	}
	if maxWorkers > len(files) {
		maxWorkers = len(files) // don't spawn more workers than files
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	ch := make(chan string, len(files)) // buffered channel to hold all file paths

	// worker reads file paths from channel and extracts comments
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

	wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go worker()
	}

	for _, f := range files {
		ch <- f
	}
	close(ch) // no more files, workers will exit
	wg.Wait()

	return results
}

// PrepareOutputFile writes the extracted comments to JSON, YAML, or plain text.
// Automatically creates the output directory if needed.
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

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

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
			if _, err := fmt.Fprintf(f, "%s:%d [%s] %s\n", c.FilePath, c.LineNumber, c.Tag, c.Content); err != nil {
				return fmt.Errorf("failed to write text: %w", err)
			}
		}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	fmt.Printf("Extracted %d comments written to %s\n", len(all), outPath)
	return nil
}

// ExtractComments scans a single file line by line for tags we care about.
// Returns all matches or an error if the file can't be opened/read.
func ExtractComments(filePath, tags string) ([]Comment, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		ext = strings.ToLower(filepath.Base(filePath)) // fallback for files without extension
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
			continue // line doesn't contain a comment
		}
		text := strings.TrimSpace(line[pos+len(char):])
		if tag := findTag(text, tagSet); tag != "" {
			out = append(out, Comment{
				Tag:        tag,
				Content:    text,
				FilePath:   filePath,
				LineNumber: lineNum,
			})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// parseTags converts a comma-separated string into a set for quick lookup
func parseTags(tags string) map[string]struct{} {
	// if empty, return a copy of all supported tags
	if strings.TrimSpace(tags) == "" {
		set := make(map[string]struct{}, len(supportedTagsLookup))
		for tag := range supportedTagsLookup {
			set[tag] = struct{}{}
		}
		return set
	}

	parts := strings.Split(tags, ",")
	set := make(map[string]struct{}, len(parts)) // preallocate
	for _, t := range parts {                    // range over slice
		trimmed := strings.ToUpper(strings.TrimSpace(t))
		if trimmed != "" {
			set[trimmed] = struct{}{}
		}
	}
	return set
}

// findTag checks if any of the allowed tags exist in the text
func findTag(text string, tags map[string]struct{}) string {
	upper := strings.ToUpper(text)
	for t := range tags {
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(t) + `\b`) // exact word match
		if re.MatchString(upper) {
			return t
		}
	}
	return ""
}

// isBinaryFile does a cheap check if a file contains null bytes
func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 8192)
	n, _ := f.Read(buf)

	// check if 0 exists in the slice of read bytes
	return slices.Contains(buf[:n], byte(0))
}

// GetAllFilePaths recursively collects all files that can contain comments
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
		if _, ok := extensionToChar[ext]; !ok {
			return nil
		}
		if isBinaryFile(path) {
			return nil
		}
		out = append(out, path)
		return nil
	})
	return out, err
}

// PrettyPrintComments prints all comments to stdout, optionally with ANSI colors
func PrettyPrintComments(m map[string][]Comment, color bool) {
	const reset = "\033[0m"
	colors := map[string]string{
		"TODO":      "\033[33m",
		"FIXME":     "\033[31m",
		"NOTE":      "\033[36m",
		"HACK":      "\033[35m",
		"BUG":       "\033[91m",
		"OPTIMIZE":  "\033[32m",
		"DEPRECATE": "\033[90m",
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

