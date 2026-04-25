// Run each prompt through Claude in three arms and snapshot the outputs.
//
// Arms:
//   __baseline__  — no system prompt
//   __terse__     — "Answer concisely."
//   <skill>       — "Answer concisely.\n\n{SKILL.md}"
//
// Writes evals/snapshots/results.json.
// Env: ICEAGE_EVAL_MODEL overrides the model passed to claude CLI.
//
// Usage: go run ./cmd/llm_run
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const tersePrefix = "Answer concisely."

func evalsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func runClaude(prompt string, system string) (string, error) {
	args := []string{"-p"}
	if system != "" {
		args = append(args, "--system-prompt", system)
	}
	if model := os.Getenv("ICEAGE_EVAL_MODEL"); model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, prompt)
	out, err := exec.Command("claude", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func claudeVersion() string {
	out, err := exec.Command("claude", "--version").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func loadPrompts(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var prompts []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			prompts = append(prompts, line)
		}
	}
	return prompts, sc.Err()
}

func discoverSkills(skillsDir string) ([]string, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}
	var skills []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(skillsDir, e.Name(), "SKILL.md")); err == nil {
			skills = append(skills, e.Name())
		}
	}
	sort.Strings(skills)
	return skills, nil
}

type Snapshot struct {
	Metadata struct {
		GeneratedAt      string `json:"generated_at"`
		ClaudeCLIVersion string `json:"claude_cli_version"`
		Model            string `json:"model"`
		NPrompts         int    `json:"n_prompts"`
		TersePrefix      string `json:"terse_prefix"`
	} `json:"metadata"`
	Prompts []string            `json:"prompts"`
	Arms    map[string][]string `json:"arms"`
}

func runArm(label string, prompts []string, system string) []string {
	results := make([]string, len(prompts))
	for i, p := range prompts {
		fmt.Printf("  [%d/%d] %s\n", i+1, len(prompts), label)
		out, err := runClaude(p, system)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  WARNING: claude call failed for prompt %d: %v\n", i+1, err)
		}
		results[i] = out
	}
	return results
}

func main() {
	base := evalsDir()
	promptsPath := filepath.Join(base, "prompts", "en.txt")
	skillsDir := filepath.Join(base, "..", "skills")
	snapshotPath := filepath.Join(base, "snapshots", "results.json")

	prompts, err := loadPrompts(promptsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read prompts: %v\n", err)
		os.Exit(1)
	}

	skills, err := discoverSkills(skillsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot discover skills: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== %d prompts × (%d skills + 2 control arms) ===\n", len(prompts), len(skills))

	model := os.Getenv("ICEAGE_EVAL_MODEL")
	if model == "" {
		model = "default"
	}

	snap := Snapshot{}
	snap.Metadata.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	snap.Metadata.ClaudeCLIVersion = claudeVersion()
	snap.Metadata.Model = model
	snap.Metadata.NPrompts = len(prompts)
	snap.Metadata.TersePrefix = tersePrefix
	snap.Prompts = prompts
	snap.Arms = make(map[string][]string)

	fmt.Println("baseline (no system prompt)")
	snap.Arms["__baseline__"] = runArm("baseline", prompts, "")

	fmt.Println("terse (control: terse instruction only, no skill)")
	snap.Arms["__terse__"] = runArm("terse", prompts, tersePrefix)

	for _, skill := range skills {
		skillMD, err := os.ReadFile(filepath.Join(skillsDir, skill, "SKILL.md"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "  WARNING: cannot read %s/SKILL.md: %v\n", skill, err)
			continue
		}
		system := tersePrefix + "\n\n" + string(skillMD)
		fmt.Printf("  %s\n", skill)
		snap.Arms[skill] = runArm(skill, prompts, system)
	}

	if err := os.MkdirAll(filepath.Dir(snapshotPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot create snapshots dir: %v\n", err)
		os.Exit(1)
	}
	data, _ := json.MarshalIndent(snap, "", "  ")
	if err := os.WriteFile(snapshotPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot write snapshot: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nWrote %s\n", snapshotPath)
}
