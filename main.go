package main

import (
	"flag"
	"fmt"
	"runtime"
	"tdl/core"
)

func main() {
	// Command-line flags
	dirpath := flag.String("dirpath", ".", "Directory to recursively scan for files")
	tag := flag.String("tag", "", "Comma-separated tags to filter by (TODO, FIXME, etc.)")
	color := flag.Bool("color", true, "Enable colorized output")
	ignore := flag.Bool("ignore", true, "Skip unsupported file extensions silently")
	workers := flag.Int("workers", runtime.NumCPU(), "Number of concurrent worker goroutines")
	output := flag.String("output", "", "Output format (json, yaml, text). Leave empty to skip file output.")
	outputdir := flag.String("outputdir", ".", "Output dir for the resulting ouput file")
	summarize := flag.Bool("summarize", false, "Print the frequency of each tag in the code base")

	// Custom usage message
	flag.Usage = func() {
		fmt.Println("Usage: tdl [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Step 1: gather files
	files, err := core.GetAllFilePaths(*dirpath)
	if err != nil {
		fmt.Println("Error scanning directory:", err)
		return
	}

	// Step 2: extract comments concurrently
	results := core.RunExtractCommentsConcurrently(files, *workers, *tag, *ignore)

	// Step 3: if output flag is set, write file and exit
	if *output != "" {
		if err := core.PrepareOutputFile(results, *output, *outputdir); err != nil {
			fmt.Println("Error writing output:", err)
		}
		return
	}

	if *summarize {
		tags := map[string]int{"TODO": 0, "FIXME": 0, "NOTE": 0, "HACK": 0, "BUG": 0, "OPTIMIZE": 0, "DEPRECATE": 0}

		for _, result := range results {
			for _, c := range result {
				tags[c.Tag]++
			}
		}

		for tag, count := range tags {
			fmt.Printf("%s : %d\n", tag, count)
		}
		return

	}

	// Normal flow: pretty-print and summary
	core.PrettyPrintComments(results, *color)

	totalComments := 0
	for _, cs := range results {
		totalComments += len(cs)
	}
	fmt.Printf("Scanned %d files, found %d comments.\n", len(files), totalComments)
}

