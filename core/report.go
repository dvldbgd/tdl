package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

