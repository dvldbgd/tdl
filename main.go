package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"tdl/core"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected subcommand: init | destroy | scan")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initTdl()
	case "destroy":
		destroyTdl()
	case "scan":
		scanCodeBase(os.Args[2:])
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

// initTdl creates the .tdl directory
func initTdl() {
	dirName := ".tdl"

	info, err := os.Stat(dirName)
	if err == nil && info.IsDir() {
		fmt.Println("Directory already exists:", dirName)
		return
	}

	// Only create if it doesn't exist
	err = os.MkdirAll(dirName, 0755)
	if err != nil {
		fmt.Println("Error creating .tdl:", err)
		return
	}

	fmt.Println("Directory created:", dirName)
}

// destroyTdl deletes the .tdl directory recursively with confirmation
func destroyTdl() {
	dirName := ".tdl"

	info, err := os.Stat(dirName)
	if os.IsNotExist(err) || !info.IsDir() {
		fmt.Println("Directory does not exist:", dirName)
		return
	}

	fmt.Printf("Are you sure you want to destroy '%s'? (y/N): ", dirName)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		fmt.Println("Aborted")
		return
	}

	err = os.RemoveAll(dirName)
	if err != nil {
		fmt.Println("Error destroying .tdl:", err)
		return
	}

	fmt.Println("Destroyed:", dirName)
}

// scanCodeBase handles scanning files and extracting comments
func scanCodeBase(args []string) {
	// Command-line flags for scanning
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	dirpath := fs.String("dirpath", ".", "Directory to recursively scan for files")
	tag := fs.String("tag", "", "Comma-separated tags to filter by (TODO, FIXME, etc.)")
	color := fs.Bool("color", true, "Enable colorized output")
	ignore := fs.Bool("ignore", true, "Skip unsupported file extensions silently")
	workers := fs.Int("workers", runtime.NumCPU(), "Number of concurrent worker goroutines")
	output := fs.String("output", "", "Output format (json, yaml, text). Leave empty to skip file output.")
	outputdir := fs.String("outputdir", ".", "Output dir for the resulting output file")
	summarize := fs.Bool("summarize", false, "Print the frequency of each tag in the code base")

	fs.Usage = func() {
		fmt.Println("Usage: tdl scan [options]")
		fmt.Println("Options:")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	// Step 1: gather files
	files, err := core.GetAllFilePaths(*dirpath)
	if err != nil {
		fmt.Println("Error scanning directory:", err)
		return
	}

	// Step 2: extract comments concurrently
	results := core.RunExtractCommentsConcurrently(files, *workers, *tag, *ignore)

	// Step 3: write output file if requested
	if *output != "" {
		if err := core.PrepareOutputFile(results, *output, *outputdir); err != nil {
			fmt.Println("Error writing output:", err)
		}
		return
	}

	// Step 4: summarize tags if requested
	if *summarize {
		tags := map[string]int{
			"TODO": 0, "FIXME": 0, "NOTE": 0, "HACK": 0, "BUG": 0, "OPTIMIZE": 0, "DEPRECATE": 0,
		}

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

	// Step 5: normal pretty-print
	core.PrettyPrintComments(results, *color)

	totalComments := 0
	for _, cs := range results {
		totalComments += len(cs)
	}
	fmt.Printf("Scanned %d files, found %d comments.\n", len(files), totalComments)
}

