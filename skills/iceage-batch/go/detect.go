package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileType string

const (
	FileTypeNaturalLanguage FileType = "natural_language"
	FileTypeCode            FileType = "code"
	FileTypeConfig          FileType = "config"
	FileTypeUnknown         FileType = "unknown"
)

var compressibleExts = map[string]bool{
	".md": true, ".txt": true, ".markdown": true, ".rst": true,
}

var skipExts = map[string]bool{
	".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
	".json": true, ".yaml": true, ".yml": true, ".toml": true, ".env": true,
	".lock": true, ".css": true, ".scss": true, ".html": true, ".xml": true,
	".sql": true, ".sh": true, ".bash": true, ".zsh": true, ".go": true,
	".rs": true, ".java": true, ".c": true, ".cpp": true, ".h": true,
	".hpp": true, ".rb": true, ".php": true, ".swift": true, ".kt": true,
	".lua": true, ".dockerfile": true, ".makefile": true, ".csv": true,
	".ini": true, ".cfg": true,
}

var configExts = map[string]bool{
	".json": true, ".yaml": true, ".yml": true, ".toml": true,
	".ini": true, ".cfg": true, ".env": true,
}

var codePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*(import |from .+ import |require\(|const |let |var )`),
	regexp.MustCompile(`^\s*(def |class |function |async function |export )`),
	regexp.MustCompile(`^\s*(if\s*\(|for\s*\(|while\s*\(|switch\s*\(|try\s*\{)`),
	regexp.MustCompile(`^\s*[\}\]\);]+\s*$`),
	regexp.MustCompile(`^\s*@\w+`),
	regexp.MustCompile(`^\s*"[^"]+"\s*:\s*`),
	regexp.MustCompile(`^\s*\w+\s*=\s*[{\[\("']`),
}

var yamlKeyVal = regexp.MustCompile(`^\w[\w\s]*:\s`)
var yamlListItem = regexp.MustCompile(`^- .+:.+`)

func isCodeLine(line string) bool {
	for _, p := range codePatterns {
		if p.MatchString(line) {
			return true
		}
	}
	return false
}

func isJSONContent(text string) bool {
	var v any
	return json.Unmarshal([]byte(text), &v) == nil
}

func isYAMLContent(lines []string) bool {
	limit := 30
	if len(lines) < limit {
		limit = len(lines)
	}
	indicators, nonEmpty := 0, 0
	for _, line := range lines[:limit] {
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		nonEmpty++
		if strings.HasPrefix(s, "---") || yamlKeyVal.MatchString(s) || yamlListItem.MatchString(s) {
			indicators++
		}
	}
	return nonEmpty > 0 && float64(indicators)/float64(nonEmpty) > 0.6
}

func detectFileType(path string) FileType {
	ext := strings.ToLower(filepath.Ext(path))

	if compressibleExts[ext] {
		return FileTypeNaturalLanguage
	}
	if skipExts[ext] {
		if configExts[ext] {
			return FileTypeConfig
		}
		return FileTypeCode
	}

	if ext == "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return FileTypeUnknown
		}
		text := string(data)
		lines := strings.SplitN(text, "\n", 51)
		if len(lines) > 50 {
			lines = lines[:50]
		}

		sample := text
		if len(sample) > 10000 {
			sample = sample[:10000]
		}
		if isJSONContent(sample) {
			return FileTypeConfig
		}
		if isYAMLContent(lines) {
			return FileTypeConfig
		}

		codeLines, nonEmpty := 0, 0
		for _, l := range lines {
			if strings.TrimSpace(l) == "" {
				continue
			}
			nonEmpty++
			if isCodeLine(l) {
				codeLines++
			}
		}
		if nonEmpty > 0 && float64(codeLines)/float64(nonEmpty) > 0.4 {
			return FileTypeCode
		}
		return FileTypeNaturalLanguage
	}

	return FileTypeUnknown
}

func shouldCompress(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return false
	}
	if strings.HasSuffix(filepath.Base(path), ".original.md") {
		return false
	}
	return detectFileType(path) == FileTypeNaturalLanguage
}
