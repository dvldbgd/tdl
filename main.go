package main

import (
	"flag"
	"fmt"
	"runtime"
	"tdl/codes"
)

// Entry point of the TODO/Comment scanner application.
// This program recursively scans a specified directory for source files
// and extracts comments matching certain tags (e.g., TODO, FIXME).
func main() {
	// Command-line flags:
	dirpath := flag.String("dirpath", ".", "Directory to recursively scan for files")          // Root directory to scan
	tag := flag.String("tag", "", "Comma-separated tags to filter by (TODO, FIXME, etc.)")     // Filter comments by these tags
	color := flag.Bool("color", true, "Enable colorized output")                               // Enable colored terminal output
	ignore := flag.Bool("ignore", true, "Skip unsupported file extensions silently")           // Ignore unsupported file types without error
	workers := flag.Int("workers", runtime.NumCPU(), "Number of concurrent worker goroutines") // Concurrency level

	// Custom usage message
	flag.Usage = func() {
		fmt.Println("Usage: tdl [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}

	flag.Parse() // Parse command-line arguments

	// Get all file paths under the specified directory
	files, err := codes.GetAllFilePaths(*dirpath)
	if err != nil {
		fmt.Println("Error scanning directory:", err)
		return
	}

	// Extract comments concurrently from all files
	results := codes.RunExtractCommentsConcurrently(files, *workers, *tag, *ignore)

	// Display extracted comments with optional color
	codes.PrettyPrintComments(results, *color)

	// Optional summary: total files scanned and comments found
	totalComments := 0
	for _, cs := range results {
		totalComments += len(cs)
	}
	fmt.Printf("Scanned %d files, found %d comments.\n", len(files), totalComments)
}

