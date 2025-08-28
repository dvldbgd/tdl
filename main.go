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
	case "print":
		printComments()
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

// initTdl creates the .tdl directory
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

// destroyTdl deletes the .tdl directory recursively with confirmation
func destroyTdl() {
	dirName := ".tdl"

	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		fmt.Println("Directory does not exist:", dirName)
		return
	}

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

// scanCodeBase handles scanning files and extracting comments
func scanCodeBase(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	dirpath := fs.String("dirpath", ".", "Directory to recursively scan for files")
	tag := fs.String("tag", "", "Comma-separated tags to filter by (TODO, FIXME, etc.)")
	color := fs.Bool("color", true, "Enable colorized output")
	printFlag := fs.Bool("print", false, "Also pretty-print comments after scanning")
	ignore := fs.Bool("ignore", true, "Skip unsupported file extensions silently")
	workers := fs.Int("workers", runtime.NumCPU(), "Number of concurrent worker goroutines")

	fs.Usage = func() {
		fmt.Println("Usage: tdl scan [options]")
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

	// Step 3: ensure .tdl exists
	if err := os.MkdirAll(".tdl", 0755); err != nil {
		fmt.Println("Failed to create .tdl directory:", err)
		return
	}

	// Step 4: write results to JSON inside .tdl
	if err := core.PrepareOutputFile(results, "json", ".tdl"); err != nil {
		fmt.Println("Error writing output:", err)
		return
	}

	// Step 5: pretty print if requested
	if *printFlag {
		core.PrettyPrintComments(results, *color)
	}

	totalComments := 0
	for _, cs := range results {
		totalComments += len(cs)
	}
	fmt.Printf("Scanned %d files, found %d comments.\n", len(files), totalComments)

	// Pretty prints comments.json
	// printComments pretty-prints existing comments.json from .tdl
}

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

	// Convert to map[filePath][]Comment for PrettyPrintComments
	results := make(map[string][]core.Comment)
	for _, c := range all {
		results[c.FilePath] = append(results[c.FilePath], c)
	}

	// Optionally, allow --color flag
	fs := flag.NewFlagSet("print", flag.ExitOnError)
	color := fs.Bool("color", true, "Enable colorized output")
	fs.Parse(os.Args[2:])

	core.PrettyPrintComments(results, *color)
}

