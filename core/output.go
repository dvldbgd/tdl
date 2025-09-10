package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// PrepareOutputFile writes results in JSON, YAML, or text format.
func PrepareOutputFile(results map[string][]Comment, format, outputDir string) error {
	var all []Comment
	for _, list := range results {
		all = append(all, list...)
	}
	if len(all) == 0 {
		return fmt.Errorf("no comments found")
	}

	ext := strings.ToLower(format)
	outPath := filepath.Join(outputDir, "comments."+ext)

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

// PrettyPrintComments prints comments grouped by file with optional colors.
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

		sort.Slice(list, func(i, j int) bool { return list[i].LineNumber < list[j].LineNumber })
		for _, c := range list {
			line := fmt.Sprintf("%-5d", c.LineNumber)
			if color {
				col := colors[c.Tag]
				if col == "" {
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

// CreateReport counts frequency of tags in comments.json
func CreateReport(tag string) {
	type LogEntry struct {
		Tag           string `json:"tag"`
		Content       string `json:"content"`
		FilePath      string `json:"file"`
		LineNumber    int    `json:"line"`
		CreationStamp string `json:"stamp"`
		Author        string `json:"author"`
		Commit        string `json:"commit"`
	}

	path := filepath.Join(".tdl", "comments.json")
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var comments []LogEntry
	if err := json.Unmarshal(data, &comments); err != nil {
		panic(err)
	}

	count := map[string]int32{
		"TODO": 0, "FIXME": 0, "NOTE": 0,
		"HACK": 0, "BUG": 0, "OPTIMIZE": 0, "DEPRECATE": 0,
	}

	for _, c := range comments {
		if _, ok := count[strings.ToUpper(c.Tag)]; ok {
			count[strings.ToUpper(c.Tag)]++
		}
	}

	if strings.ToLower(tag) == "all" || tag == "" {
		for t, cnt := range count {
			fmt.Printf("%s => %d\n", t, cnt)
		}
	} else {
		for _, t := range strings.Split(tag, ",") {
			t = strings.ToUpper(strings.TrimSpace(t))
			if cnt, ok := count[t]; ok {
				fmt.Printf("%s => %d\n", t, cnt)
			}
		}
	}
}

