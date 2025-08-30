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
	// simple CLI dispatch based on first arg
	if len(os.Args) < 2 {
		fmt.Println("Expected subcommand: init | destroy | scan")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initTdl() // create .tdl dir
	case "destroy":
		destroyTdl() // remove .tdl with confirmation
	case "scan":
		scanCodeBase(os.Args[2:]) // scan directory and extract comments
	case "print":
		printComments() // pretty-print existing comments.json
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

// initTdl creates the .tdl directory if it doesn't exist
func initTdl() {
	dirName := ".tdl"

	if _, err := os.Stat(dirName); err == nil {
		fmt.Println("Directory already exists:", dirName)
		return
	}

	if err := os.MkdirAll(dirName, 0755); err != nil {
		fmt.Println("Error creating .tdl:", err)
		return
	}

	fmt.Println("Directory created:", dirName)
}

// destroyTdl removes the .tdl directory after user confirmation
func destroyTdl() {
	dirName := ".tdl"

	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		fmt.Println("Directory does not exist:", dirName)
		return
	}

	// confirm deletion
	fmt.Printf("Are you sure you want to destroy '%s'? (y/N): ", dirName)
	var input string
	fmt.Scanln(&input)

	switch input {
	case "y", "Y", "yes", "YES":
		if err := os.RemoveAll(dirName); err != nil {
			fmt.Println("Error destroying .tdl:", err)
			return
		}
		fmt.Println("Destroyed:", dirName)
	default:
		fmt.Println("Aborted")
	}
}

// scanCodeBase is the main scanning workflow:
// 1. collect files
// 2. extract tagged comments concurrently
// 3. write JSON to .tdl
// 4. optionally pretty-print to console
func scanCodeBase(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	dirpath := fs.String("dirpath", ".", "Directory to recursively scan")
	tag := fs.String("tag", "", "Comma-separated tags to filter by")
	color := fs.Bool("color", true, "Enable color output")
	printFlag := fs.Bool("print", false, "Also pretty-print after scanning")
	ignore := fs.Bool("ignore", true, "Skip unsupported file extensions silently")
	workers := fs.Int("workers", runtime.NumCPU(), "Number of concurrent workers")

	// usage info if user messes up flags
	fs.Usage = func() {
		fmt.Println("Usage: tdl scan [options]")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	// Step 1: recursively collect files
	files, err := core.GetAllFilePaths(*dirpath)
	if err != nil {
		fmt.Println("Error scanning directory:", err)
		return
	}

	// Step 2: extract comments using multiple goroutines
	results := core.RunExtractCommentsConcurrently(files, *workers, *tag, *ignore)

	// Step 3: ensure .tdl exists
	if err := os.MkdirAll(".tdl", 0755); err != nil {
		fmt.Println("Failed to create .tdl directory:", err)
		return
	}

	// Step 4: save all extracted comments to JSON in .tdl
	if err := core.PrepareOutputFile(results, "json", ".tdl"); err != nil {
		fmt.Println("Error writing output:", err)
		return
	}

	// Step 5: optionally pretty-print
	if *printFlag {
		core.PrettyPrintComments(results, *color)
	}

	// quick stats for user
	totalComments := 0
	for _, cs := range results {
		totalComments += len(cs)
	}
	fmt.Printf("Scanned %d files, found %d comments.\n", len(files), totalComments)
}

// printComments reads comments.json from .tdl and pretty-prints them
func printComments() {
	filePath := ".tdl/comments.json"
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening comments file:", err)
		return
	}
	defer f.Close()

	var all []core.Comment
	dec := json.NewDecoder(f)
	if err := dec.Decode(&all); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// convert to map[filePath][]Comment for PrettyPrintComments
	results := make(map[string][]core.Comment)
	for _, c := range all {
		results[c.FilePath] = append(results[c.FilePath], c)
	}

	// optional --color flag
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	color := fs.Bool("color", true, "Enable colorized output")
	fs.Parse(os.Args[2:])

	core.PrettyPrintComments(results, *color)
}

