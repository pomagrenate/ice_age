package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const maxRetries = 2
const maxFileSize = 500_000

var outerFenceRe = regexp.MustCompile("(?s)\\A\\s*(`{3,}|~{3,})[^\\n]*\\n(.*)\\n\\1\\s*\\z")

var sensitiveBasenameRe = regexp.MustCompile(`(?ix)^(` +
	`\.env(\..+)?` +
	`|\.netrc` +
	`|credentials(\..+)?` +
	`|secrets?(\..+)?` +
	`|passwords?(\..+)?` +
	`|id_(rsa|dsa|ecdsa|ed25519)(\.pub)?` +
	`|authorized_keys` +
	`|known_hosts` +
	`|.*\.(pem|key|p12|pfx|crt|cer|jks|keystore|asc|gpg)` +
	`)$`)

var sensitivePathComponents = map[string]bool{
	".ssh": true, ".aws": true, ".gnupg": true, ".kube": true, ".docker": true,
}

var sensitiveNameTokens = []string{
	"secret", "credential", "password", "passwd",
	"apikey", "accesskey", "token", "privatekey",
}

func isSensitivePath(path string) bool {
	base := filepath.Base(path)
	if sensitiveBasenameRe.MatchString(base) {
		return true
	}
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		if sensitivePathComponents[strings.ToLower(part)] {
			return true
		}
	}
	lower := strings.ToLower(base)
	lower = regexp.MustCompile(`[_\-\s.]`).ReplaceAllString(lower, "")
	for _, tok := range sensitiveNameTokens {
		if strings.Contains(lower, tok) {
			return true
		}
	}
	return false
}

func stripLLMWrapper(text string) string {
	m := outerFenceRe.FindStringSubmatch(text)
	if m != nil {
		return m[2]
	}
	return text
}

func callClaude(prompt string) (string, error) {
	model := os.Getenv("ICEAGE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-5"
	}

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		client := anthropic.NewClient(option.WithAPIKey(apiKey))
		msg, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
			Model:     anthropic.F(anthropic.Model(model)),
			MaxTokens: anthropic.F(int64(8192)),
			Messages: anthropic.F([]anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			}),
		})
		if err != nil {
			return "", fmt.Errorf("claude API call failed: %w", err)
		}
		text := msg.Content[0].Text
		return stripLLMWrapper(strings.TrimSpace(text)), nil
	}

	// Fallback: claude CLI
	cmd := exec.Command("claude", "--print")
	cmd.Stdin = bytes.NewBufferString(prompt)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if ok := false; !ok {
			_ = exitErr
		}
		return "", fmt.Errorf("claude CLI failed: %w", err)
	}
	return stripLLMWrapper(strings.TrimSpace(string(out))), nil
}

func buildCompressPrompt(original string) string {
	return `Compress this markdown into iceage format.

STRICT RULES:
- Do NOT modify anything inside ` + "```" + ` code blocks
- Do NOT modify anything inside inline backticks
- Preserve ALL URLs exactly
- Preserve ALL headings exactly
- Preserve file paths and commands
- Return ONLY the compressed markdown body — do NOT wrap the entire output in a ` + "```" + `markdown fence or any other fence. Inner code blocks from the original stay as-is; do not add a new outer fence around the whole file.

Only compress natural language.

TEXT:
` + original
}

func buildFixPrompt(original, compressed string, errors []string) string {
	errLines := make([]string, len(errors))
	for i, e := range errors {
		errLines[i] = "- " + e
	}
	return `You are fixing an iceage-compressed markdown file. Specific validation errors were found.

CRITICAL RULES:
- DO NOT recompress or rephrase the file
- ONLY fix the listed errors — leave everything else exactly as-is
- The ORIGINAL is provided as reference only (to restore missing content)
- Preserve iceage style in all untouched sections

ERRORS TO FIX:
` + strings.Join(errLines, "\n") + `

HOW TO FIX:
- Missing URL: find it in ORIGINAL, restore it exactly where it belongs in COMPRESSED
- Code block mismatch: find the exact code block in ORIGINAL, restore it in COMPRESSED
- Heading mismatch: restore the exact heading text from ORIGINAL into COMPRESSED
- Do not touch any section not mentioned in the errors

ORIGINAL (reference only):
` + original + `

COMPRESSED (fix this):
` + compressed + `

Return ONLY the fixed compressed file. No explanation.`
}

func compressFile(path string, noBackup bool) (bool, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return false, fmt.Errorf("file not found: %s", absPath)
	}
	if info.Size() > maxFileSize {
		return false, fmt.Errorf("file too large (max 500KB): %s", absPath)
	}
	if isSensitivePath(absPath) {
		return false, fmt.Errorf(
			"refusing to compress %s: filename looks sensitive "+
				"(credentials, keys, secrets, or known private paths). "+
				"Compression sends file contents to the Anthropic API. "+
				"Rename the file if this is a false positive.", absPath)
	}

	fmt.Printf("Processing: %s\n", absPath)

	if !shouldCompress(absPath) {
		fmt.Println("Skipping (not natural language)")
		return false, nil
	}

	origData, err := os.ReadFile(absPath)
	if err != nil {
		return false, err
	}
	originalText := string(origData)

	var backupPath string
	if !noBackup {
		stem := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))
		backupPath = filepath.Join(filepath.Dir(absPath), stem+".original.md")

		if _, err := os.Stat(backupPath); err == nil {
			fmt.Printf("Backup file already exists: %s\n", backupPath)
			fmt.Println("Aborting to prevent data loss. Remove or rename the backup file to proceed.")
			return false, nil
		}
	}

	fmt.Println("Compressing with Claude...")
	compressed, err := callClaude(buildCompressPrompt(originalText))
	if err != nil {
		return false, err
	}

	if !noBackup {
		if err := os.WriteFile(backupPath, origData, 0600); err != nil {
			return false, err
		}
	}
	if err := os.WriteFile(absPath, []byte(compressed), 0600); err != nil {
		return false, err
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		fmt.Printf("\nValidation attempt %d\n", attempt+1)

		result := validateTexts(originalText, compressed)
		if result.IsValid {
			fmt.Println("Validation passed")
			return true, nil
		}

		fmt.Println("Validation failed:")
		for _, e := range result.Errors {
			fmt.Printf("   - %s\n", e)
		}

		if attempt == maxRetries-1 {
			_ = os.WriteFile(absPath, origData, 0600)
			if !noBackup {
				_ = os.Remove(backupPath)
			}
			fmt.Println("Failed after retries — original restored")
			return false, nil
		}

		fmt.Println("Fixing with Claude...")
		compressed, err = callClaude(buildFixPrompt(originalText, compressed, result.Errors))
		if err != nil {
			return false, err
		}
		if err := os.WriteFile(absPath, []byte(compressed), 0600); err != nil {
			return false, err
		}
	}

	return true, nil
}
