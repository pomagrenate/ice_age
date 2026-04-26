package main

import (
	"os"
	"regexp"
	"strings"
)

var (
	urlRegex    = regexp.MustCompile(`https?://[^\s)]+`)
	fenceOpenRe = regexp.MustCompile("^(\\s{0,3})(`{3,}|~{3,})(.*)$")
	headingRe   = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)`)
	bulletRe    = regexp.MustCompile(`(?m)^\s*[-*+]\s+`)
	pathRe      = regexp.MustCompile(`(?:./|\.\./|/|[A-Za-z]:\\)[\w\-/\\\.]+|[\w\-\.]+[/\\][\w\-/\\\.]+`)
)

type ValidationResult struct {
	IsValid  bool
	Errors   []string
	Warnings []string
}

func (v *ValidationResult) addError(msg string) {
	v.IsValid = false
	v.Errors = append(v.Errors, msg)
}

func (v *ValidationResult) addWarning(msg string) {
	v.Warnings = append(v.Warnings, msg)
}

type heading struct{ level, title string }

func extractHeadings(text string) []heading {
	matches := headingRe.FindAllStringSubmatch(text, -1)
	out := make([]heading, 0, len(matches))
	for _, m := range matches {
		out = append(out, heading{m[1], strings.TrimSpace(m[2])})
	}
	return out
}

func extractCodeBlocks(text string) []string {
	var blocks []string
	lines := strings.Split(text, "\n")
	i, n := 0, len(lines)
	for i < n {
		m := fenceOpenRe.FindStringSubmatch(lines[i])
		if m == nil {
			i++
			continue
		}
		fenceChar := rune(m[2][0])
		fenceLen := len(m[2])
		blockLines := []string{lines[i]}
		i++
		closed := false
		for i < n {
			cm := fenceOpenRe.FindStringSubmatch(lines[i])
			if cm != nil && rune(cm[2][0]) == fenceChar && len(cm[2]) >= fenceLen && strings.TrimSpace(cm[3]) == "" {
				blockLines = append(blockLines, lines[i])
				closed = true
				i++
				break
			}
			blockLines = append(blockLines, lines[i])
			i++
		}
		if closed {
			blocks = append(blocks, strings.Join(blockLines, "\n"))
		}
	}
	return blocks
}

func extractURLs(text string) map[string]bool {
	matches := urlRegex.FindAllString(text, -1)
	out := make(map[string]bool, len(matches))
	for _, m := range matches {
		out[m] = true
	}
	return out
}

func extractPaths(text string) map[string]bool {
	matches := pathRe.FindAllString(text, -1)
	out := make(map[string]bool, len(matches))
	for _, m := range matches {
		out[m] = true
	}
	return out
}

func countBullets(text string) int {
	return len(bulletRe.FindAllString(text, -1))
}

func setDiff(a, b map[string]bool) []string {
	var out []string
	for k := range a {
		if !b[k] {
			out = append(out, k)
		}
	}
	return out
}

func validateHeadings(orig, comp string, result *ValidationResult) {
	h1 := extractHeadings(orig)
	h2 := extractHeadings(comp)
	if len(h1) != len(h2) {
		result.addError("Heading count mismatch")
		return
	}
	for i := range h1 {
		if h1[i] != h2[i] {
			result.addWarning("Heading text/order changed")
			return
		}
	}
}

func validateCodeBlocks(orig, comp string, result *ValidationResult) {
	c1 := extractCodeBlocks(orig)
	c2 := extractCodeBlocks(comp)
	if len(c1) != len(c2) {
		result.addError("Code blocks not preserved exactly")
		return
	}
	for i := range c1 {
		if c1[i] != c2[i] {
			result.addError("Code blocks not preserved exactly")
			return
		}
	}
}

func validateURLs(orig, comp string, result *ValidationResult) {
	u1 := extractURLs(orig)
	u2 := extractURLs(comp)
	lost := setDiff(u1, u2)
	added := setDiff(u2, u1)
	if len(lost) > 0 || len(added) > 0 {
		result.addError("URL mismatch: lost=" + strings.Join(lost, ",") + " added=" + strings.Join(added, ","))
	}
}

func validatePaths(orig, comp string, result *ValidationResult) {
	p1 := extractPaths(orig)
	p2 := extractPaths(comp)
	lost := setDiff(p1, p2)
	added := setDiff(p2, p1)
	if len(lost) > 0 || len(added) > 0 {
		result.addWarning("Path mismatch: lost=" + strings.Join(lost, ",") + " added=" + strings.Join(added, ","))
	}
}

func validateBullets(orig, comp string, result *ValidationResult) {
	b1 := countBullets(orig)
	if b1 == 0 {
		return
	}
	b2 := countBullets(comp)
	diff := float64(b1-b2)
	if diff < 0 {
		diff = -diff
	}
	if diff/float64(b1) > 0.15 {
		result.addWarning("Bullet count changed too much")
	}
}

func validateTexts(orig, comp string) *ValidationResult {
	result := &ValidationResult{IsValid: true}
	validateHeadings(orig, comp, result)
	validateCodeBlocks(orig, comp, result)
	validateURLs(orig, comp, result)
	validatePaths(orig, comp, result)
	validateBullets(orig, comp, result)
	return result
}

func validate(originalPath, compressedPath string) *ValidationResult {
	result := &ValidationResult{IsValid: true}

	origData, err := os.ReadFile(originalPath)
	if err != nil {
		result.addError("Cannot read original: " + err.Error())
		return result
	}
	compData, err := os.ReadFile(compressedPath)
	if err != nil {
		result.addError("Cannot read compressed: " + err.Error())
		return result
	}

	return validateTexts(string(origData), string(compData))
}
