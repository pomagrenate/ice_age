package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	noBackup := flag.Bool("no-backup", false, "skip creating .original.md backups")
	dryRun := flag.Bool("dry-run", false, "list what would be compressed, no API calls")
	workers := flag.Int("workers", 1, "parallel workers")
	recursive := flag.Bool("recursive", true, "recurse into subdirectories")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: iceage-batch [flags] <directory>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  iceage-batch docs/")
		fmt.Fprintln(os.Stderr, "  iceage-batch docs/ --no-backup")
		fmt.Fprintln(os.Stderr, "  iceage-batch docs/ --dry-run")
		fmt.Fprintln(os.Stderr, "  iceage-batch docs/ --workers 3")
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	dir := flag.Arg(0)
	info, err := os.Stat(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", dir)
		os.Exit(1)
	}

	if *workers < 1 {
		*workers = 1
	}

	os.Exit(runBatch(dir, BatchOptions{
		NoBackup:  *noBackup,
		DryRun:    *dryRun,
		Workers:   *workers,
		Recursive: *recursive,
	}))
}
