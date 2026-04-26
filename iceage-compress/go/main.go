package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	noBackup := flag.Bool("no-backup", false, "skip creating .original.md backup")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: iceage-compress [--no-backup] <filepath>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	path := flag.Arg(0)

	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "File not found: %s\n", path)
		os.Exit(1)
	}
	if !info.Mode().IsRegular() {
		fmt.Fprintf(os.Stderr, "Not a file: %s\n", path)
		os.Exit(1)
	}

	absPath, _ := filepath.Abs(path)
	ft := detectFileType(absPath)
	fmt.Printf("Detected: %s\n", ft)

	if !shouldCompress(absPath) {
		fmt.Println("Skipping: file is not natural language (code/config)")
		os.Exit(0)
	}

	fmt.Println("Starting iceage compression...\n")

	success, err := compressFile(absPath, *noBackup)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if success {
		fmt.Println("\nCompression completed successfully")
		fmt.Printf("Compressed: %s\n", absPath)
		if !*noBackup {
			stem := absPath[:len(absPath)-len(filepath.Ext(absPath))]
			fmt.Printf("Original:   %s\n", stem+".original.md")
		}
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "Compression failed after retries")
	os.Exit(2)
}
