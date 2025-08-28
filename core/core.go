package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Comment represents a tagged comment found in a file.
type Comment struct {
	Tag        string `json:"tag" yaml:"tag"`
	Content    string `json:"content" yaml:"content"`
	FilePath   string `json:"file" yaml:"file"`
	LineNumber int    `json:"line" yaml:"line"`
}

var (
	singleLineCommentMap = map[string][]string{
		"//": {".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".swift", ".kt", ".rs", ".scala", ".ts", ".js", ".jsx", ".tsx"},
		"#":  {".py", ".rb", ".sh", ".bash", ".zsh", ".yml", ".yaml", ".toml", ".pl", ".pm", ".mk", "makefile", "dockerfile", ".ini"},
		";":  {".lisp", ".clj", ".scm", ".s", ".asm"},
		"--": {".lua", ".hs", ".sql", ".adb"},
		"'":  {".vb", ".vbs"},
		"..": {".rst"},
	}

	SupportedTags       = []string{"TODO", "FIXME", "NOTE", "HACK", "BUG", "OPTIMIZE", "DEPRECATE"}
	supportedTagsLookup = make(map[string]struct{})
	extensionToChar     = make(map[string]string)

	maxScanCapacity = 1024 * 1024 // 1 MB scanner buffer
)

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

// RunExtractCommentsConcurrently runs ExtractComments in parallel for multiple files.
func RunExtractCommentsConcurrently(files []string, maxWorkers int, tags string, ignoreErrors bool) map[string][]Comment {
	results := make(map[string][]Comment)
	if len(files) == 0 {
		return results
	}

	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if maxWorkers > len(files) {
		maxWorkers = len(files)
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	ch := make(chan string, len(files))

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
	close(ch)

	wg.Wait()
	return results
}

// PrepareOutputFile writes results to a file (json/yaml/text)
func PrepareOutputFile(results map[string][]Comment, format string, outputDir string) error {
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

// ExtractComments reads a file and extracts tagged comments
func ExtractComments(filePath, tags string) ([]Comment, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == "" {
		ext = strings.ToLower(filepath.Base(filePath))
	}
	char, ok := extensionToChar[ext]
	if !ok {
		return nil, nil
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
			continue
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

// parseTags parses tag filter string
func parseTags(tags string) map[string]struct{} {
	if strings.TrimSpace(tags) == "" {
		set := make(map[string]struct{}, len(supportedTagsLookup))
		for k := range supportedTagsLookup {
			set[k] = struct{}{}
		}
		return set
	}
	set := make(map[string]struct{})
	for _, t := range strings.Split(tags, ",") {
		trimmed := strings.ToUpper(strings.TrimSpace(t))
		if trimmed != "" {
			set[trimmed] = struct{}{}
		}
	}
	return set
}

// findTag matches text against allowed tags
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

// isBinaryFile detects binary files
func isBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

// GetAllFilePaths recursively collects files
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

// PrettyPrintComments displays comments nicely
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

