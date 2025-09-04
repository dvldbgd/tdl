package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"tdl/core"
)

func main() {
	// Basic CLI entrypoint — dispatches based on first argument
	if len(os.Args) < 2 {
		fmt.Println("Expected subcommand: init | destroy | scan | print")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initTdl() // create .tdl dir if not present
	case "destroy":
		destroyTdl() // remove .tdl after confirmation
	case "scan":
		scanCodeBase(os.Args[2:]) // scan project and extract tagged comments
	case "print":
		printComments() // read .tdl/comments.json and pretty-print
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

// initTdl ensures .tdl exists (creates if missing)
func initTdl() {
	dirName := ".tdl"

	if _, err := os.Stat(dirName); err == nil {
		// directory already exists
		fmt.Println("Directory already exists:", dirName)
		return
	}

	// try to create directory
	if err := os.MkdirAll(dirName, 0755); err != nil {
		fmt.Println("Error creating .tdl:", err)
		return
	}

	fmt.Println("Directory created:", dirName)
}

// destroyTdl deletes .tdl after user types yes/y confirmation
func destroyTdl() {
	dirName := ".tdl"

	// if not present, nothing to do
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		fmt.Println("Directory does not exist:", dirName)
		return
	}

	// ask user before destructive action
	fmt.Printf("Are you sure you want to destroy '%s'? (y/N): ", dirName)
	var input string
	fmt.Scanln(&input)

	switch input {
	case "y", "Y", "yes", "YES":
		// confirmed — remove entire directory recursively
		if err := os.RemoveAll(dirName); err != nil {
			fmt.Println("Error destroying .tdl:", err)
			return
		}
		fmt.Println("Destroyed:", dirName)
	default:
		// anything else cancels
		fmt.Println("Aborted")
	}
}

// scanCodeBase:
// 1. parse flags
// 2. collect all files
// 3. run comment extraction concurrently
// 4. write JSON results to .tdl
// 5. optionally pretty-print and show stats
func scanCodeBase(args []string) {
	// setup CLI flags
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	dirpath := fs.String("dirpath", ".", "Directory to recursively scan")
	tag := fs.String("tag", "", "Comma-separated tags to filter by")
	color := fs.Bool("color", true, "Enable color output")
	printFlag := fs.Bool("print", false, "Also pretty-print after scanning")
	ignore := fs.Bool("ignore", true, "Skip unsupported file extensions silently")
	workers := fs.Int("workers", runtime.NumCPU(), "Number of concurrent workers")

	// custom usage info
	fs.Usage = func() {
		fmt.Println("Usage: tdl scan [options]")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	// Step 1: recursively collect files under dirpath
	files, err := core.GetAllFilePaths(*dirpath)
	if err != nil {
		fmt.Println("Error scanning directory:", err)
		return
	}

	// Step 2: run extraction using multiple goroutines
	results := core.RunExtractCommentsConcurrently(files, *workers, *tag, *ignore)

	// Step 3: ensure .tdl exists before writing
	if err := os.MkdirAll(".tdl", 0755); err != nil {
		fmt.Println("Failed to create .tdl directory:", err)
		return
	}

	// Step 4: save comments to JSON
	if err := core.PrepareOutputFile(results, "json", ".tdl"); err != nil {
		fmt.Println("Error writing output:", err)
		return
	}

	// Step 5: optional pretty-print after scan
	if *printFlag {
		core.PrettyPrintComments(results, *color)
	}

	// Show quick stats
	totalComments := 0
	for _, cs := range results {
		totalComments += len(cs)
	}
	fmt.Printf("Scanned %d files, found %d comments.\n", len(files), totalComments)
}

// printComments loads .tdl/comments.json and prints with optional coloring
func printComments() {
	filePath := ".tdl/comments.json"
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening comments file:", err)
		return
	}
	defer f.Close()

	// decode raw JSON into []core.Comment
	var all []core.Comment
	dec := json.NewDecoder(f)
	if err := dec.Decode(&all); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// regroup by file for PrettyPrintComments
	results := make(map[string][]core.Comment)
	for _, c := range all {
		results[c.FilePath] = append(results[c.FilePath], c)
	}

	// parse optional flags for print
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	color := fs.Bool("color", true, "Enable colorized output")
	fs.Parse(os.Args[2:])

	// pretty print the comments
	core.PrettyPrintComments(results, *color)
}

