package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type BatchOptions struct {
	NoBackup  bool
	DryRun    bool
	Workers   int
	Recursive bool
}

type fileResult struct {
	path    string
	status  CompressStatus
	err     error
	elapsed time.Duration
}

func collectFiles(dir string, recursive bool) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != dir && !recursive {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldCompress(path) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func runBatch(dir string, opts BatchOptions) int {
	files, err := collectFiles(dir, opts.Recursive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Walk error: %v\n", err)
		return 1
	}

	if len(files) == 0 {
		fmt.Println("No compressible files found.")
		return 0
	}

	if opts.DryRun {
		fmt.Printf("Found %d compressible file(s) in %s (dry run — no changes)\n\n", len(files), dir)
		for _, f := range files {
			stem := strings.TrimSuffix(f, filepath.Ext(f))
			if _, err := os.Stat(stem + ".original.md"); err == nil {
				fmt.Printf("  %s  (would skip — backup exists)\n", f)
			} else {
				fmt.Printf("  %s\n", f)
			}
		}
		fmt.Printf("\n%d file(s) would be processed.\n", len(files))
		return 0
	}

	total := len(files)
	fmt.Printf("Found %d compressible file(s) in %s\n\n", total, dir)

	var counter atomic.Int32
	var mu sync.Mutex
	resultCh := make(chan fileResult, total)
	sem := make(chan struct{}, opts.Workers)
	var wg sync.WaitGroup

	for _, f := range files {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			n := int(counter.Add(1))
			start := time.Now()
			status, err := compressFile(path, opts.NoBackup, true)
			elapsed := time.Since(start)

			mu.Lock()
			switch status {
			case StatusOK:
				fmt.Printf("[%d/%d] %s ... OK (%.1fs)\n", n, total, path, elapsed.Seconds())
			case StatusSkippedType:
				fmt.Printf("[%d/%d] %s ... Skipped (not natural language)\n", n, total, path)
			case StatusSkippedBackup:
				fmt.Printf("[%d/%d] %s ... Skipped (backup exists)\n", n, total, path)
			case StatusFailed:
				if err != nil {
					fmt.Printf("[%d/%d] %s ... FAILED (%v)\n", n, total, path, err)
				} else {
					fmt.Printf("[%d/%d] %s ... FAILED (validation failed after retries)\n", n, total, path)
				}
			}
			mu.Unlock()

			resultCh <- fileResult{path, status, err, elapsed}
		}(f)
	}

	wg.Wait()
	close(resultCh)

	var nOK, nSkipped, nFailed int
	for r := range resultCh {
		switch r.status {
		case StatusOK:
			nOK++
		case StatusSkippedType, StatusSkippedBackup:
			nSkipped++
		case StatusFailed:
			nFailed++
		}
	}

	fmt.Printf("\nDone\n")
	fmt.Printf("  Compressed: %d\n", nOK)
	if nSkipped > 0 {
		fmt.Printf("  Skipped:    %d\n", nSkipped)
	}
	if nFailed > 0 {
		fmt.Printf("  Failed:     %d\n", nFailed)
		return 2
	}
	return 0
}
