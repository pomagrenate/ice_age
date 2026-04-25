package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: iceage-compress <filepath>")
		os.Exit(1)
	}

	path := os.Args[1]

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

	success, err := compressFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if success {
		fmt.Println("\nCompression completed successfully")
		stem := absPath[:len(absPath)-len(filepath.Ext(absPath))]
		backupPath := stem + ".original.md"
		fmt.Printf("Compressed: %s\n", absPath)
		fmt.Printf("Original:   %s\n", backupPath)
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "Compression failed after retries")
	os.Exit(2)
}
