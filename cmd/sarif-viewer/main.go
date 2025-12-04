package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/drew/devpipe/internal/sarif"
)

func main() {
	// Command-line flags
	verbose := flag.Bool("v", false, "Verbose output (show rule names)")
	summary := flag.Bool("s", false, "Show summary grouped by rule")
	dir := flag.String("d", "", "Directory to search for SARIF files")
	flag.Parse()

	// Get SARIF file(s)
	var files []string
	if *dir != "" {
		// Search directory for SARIF files
		found, err := sarif.FindSARIFFiles(*dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching directory: %v\n", err)
			os.Exit(1)
		}
		files = found
	} else if flag.NArg() > 0 {
		// Use files from arguments
		files = flag.Args()
	} else {
		// Default: look for SARIF files in tmp/codeql
		defaultPath := "tmp/codeql/results.sarif"
		if _, err := os.Stat(defaultPath); err == nil {
			files = []string{defaultPath}
		} else {
			fmt.Fprintf(os.Stderr, "Usage: %s [options] <sarif-file> [<sarif-file>...]\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "   or: %s -d <directory>\n\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "Options:\n")
			flag.PrintDefaults()
			fmt.Fprintf(os.Stderr, "\nIf no file is specified, looks for tmp/codeql/results.sarif\n")
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No SARIF files found\n")
		os.Exit(1)
	}

	// Process all files
	var allFindings []sarif.Finding
	for _, file := range files {
		// Parse SARIF file
		doc, err := sarif.Parse(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file, err)
			continue
		}

		findings := doc.GetFindings()
		
		// If multiple files, show which file we're processing
		if len(files) > 1 && len(findings) > 0 {
			fmt.Printf("\n📄 %s:\n", filepath.Base(file))
		}

		allFindings = append(allFindings, findings...)
	}

	// Display results
	if *summary {
		sarif.PrintSummary(allFindings)
	} else {
		sarif.PrintFindings(allFindings, *verbose)
	}

	// Exit with error code if issues found
	if len(allFindings) > 0 {
		os.Exit(1)
	}
}
