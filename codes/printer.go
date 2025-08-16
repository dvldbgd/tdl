package codes

import (
	"fmt"
	"sort"
)

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

