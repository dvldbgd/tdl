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
	// CLI entrypoint — dispatch based on subcommand
	if len(os.Args) < 2 {
		fmt.Println("Expected subcommand: init | destroy | scan | print | report")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initTdl() // ensure .tdl exists
	case "destroy":
		destroyTdl() // remove .tdl after confirmation
	case "scan":
		scanCodeBase(os.Args[2:]) // scan project for tagged comments
	case "print":
		printComments() // pretty-print from .tdl/comments.json
	case "report":
		reportCodebase() // frequency summary of tags
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

// initTdl creates .tdl if missing
func initTdl() {
	const dirName = ".tdl"

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

// destroyTdl deletes .tdl after explicit confirmation
func destroyTdl() {
	const dirName = ".tdl"

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

// scanCodeBase:
// 1. parse flags
// 2. collect files
// 3. extract comments concurrently
// 4. save JSON output to .tdl
// 5. optional pretty-print + stats
func scanCodeBase(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	dirpath := fs.String("dirpath", ".", "Directory to recursively scan")
	tag := fs.String("tag", "", "Comma-separated tags to filter by")
	color := fs.Bool("color", true, "Enable color output")
	printFlag := fs.Bool("print", false, "Pretty-print after scanning")
	ignore := fs.Bool("ignore", true, "Skip unsupported file extensions")
	workers := fs.Int("workers", runtime.NumCPU(), "Number of concurrent workers")

	fs.Usage = func() {
		fmt.Println("Usage: tdl scan [options]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	files, err := core.GetAllFilePaths(*dirpath)
	if err != nil {
		fmt.Println("Error scanning directory:", err)
		return
	}

	results := core.RunExtractCommentsConcurrently(files, *workers, *tag, *ignore)

	if err := os.MkdirAll(".tdl", 0755); err != nil {
		fmt.Println("Failed to create .tdl:", err)
		return
	}

	if err := core.PrepareOutputFile(results, "json", ".tdl"); err != nil {
		fmt.Println("Error writing output:", err)
		return
	}

	if *printFlag {
		core.PrettyPrintComments(results, *color)
	}

	total := 0
	for _, cs := range results {
		total += len(cs)
	}
	fmt.Printf("Scanned %d files, found %d comments.\n", len(files), total)
}

// printComments loads and pretty-prints comments.json
func printComments() {
	filePath := ".tdl/comments.json"
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening comments file:", err)
		return
	}
	defer f.Close()

	var all []core.Comment
	if err := json.NewDecoder(f).Decode(&all); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	// regroup by file for printing
	results := make(map[string][]core.Comment)
	for _, c := range all {
		results[c.FilePath] = append(results[c.FilePath], c)
	}

	fs := flag.NewFlagSet("print", flag.ExitOnError)
	color := fs.Bool("color", true, "Enable colorized output")
	fs.Parse(os.Args[2:])

	core.PrettyPrintComments(results, *color)
}

// reportCodebase prints tag frequency stats
func reportCodebase() {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	tag := fs.String("tag", "all", "Comma-separated list of tags (default: all)")
	fs.Parse(os.Args[2:])

	core.CreateReport(*tag)
}

