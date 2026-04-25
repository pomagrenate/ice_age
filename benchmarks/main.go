package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const scriptVersion = "1.0.0"
const normalSystem = "You are a helpful assistant."
const benchmarkStart = "<!-- BENCHMARK-TABLE-START -->"
const benchmarkEnd = "<!-- BENCHMARK-TABLE-END -->"

// ---------- paths ----------

func scriptDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}

func repoDir() string { return filepath.Dir(scriptDir()) }

// ---------- types ----------

type PromptEntry struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Prompt   string `json:"prompt"`
}

type PromptsFile struct {
	Prompts []PromptEntry `json:"prompts"`
}

type TrialResult struct {
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	Text         string `json:"text"`
	StopReason   string `json:"stop_reason"`
}

type BenchmarkEntry struct {
	ID       string        `json:"id"`
	Category string        `json:"category"`
	Prompt   string        `json:"prompt"`
	Normal   []TrialResult `json:"normal"`
	Iceage   []TrialResult `json:"iceage"`
}

type StatRow struct {
	ID            string  `json:"id"`
	Category      string  `json:"category"`
	Prompt        string  `json:"prompt"`
	NormalMedian  int     `json:"normal_median"`
	IceageMedian  int     `json:"iceage_median"`
	SavingsPct    int     `json:"savings_pct"`
}

type Summary struct {
	AvgSavings int `json:"avg_savings"`
	MinSavings int `json:"min_savings"`
	MaxSavings int `json:"max_savings"`
	AvgNormal  int `json:"avg_normal"`
	AvgIceage  int `json:"avg_iceage"`
}

type ResultsOutput struct {
	Metadata struct {
		ScriptVersion  string `json:"script_version"`
		Model          string `json:"model"`
		Date           string `json:"date"`
		Trials         int    `json:"trials"`
		SkillMDSHA256  string `json:"skill_md_sha256"`
	} `json:"metadata"`
	Summary Summary      `json:"summary"`
	Rows    []StatRow    `json:"rows"`
	Raw     []BenchmarkEntry `json:"raw"`
}

// ---------- env ----------

func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		key, rest, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		value := strings.TrimSpace(rest)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

// ---------- stats ----------

func medianInts(vals []int64) float64 {
	if len(vals) == 0 {
		return 0
	}
	cp := append([]int64(nil), vals...)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	n := len(cp)
	if n%2 == 0 {
		return float64(cp[n/2-1]+cp[n/2]) / 2
	}
	return float64(cp[n/2])
}

func medianFloats(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	cp := append([]float64(nil), vals...)
	sort.Float64s(cp)
	n := len(cp)
	if n%2 == 0 {
		return (cp[n/2-1] + cp[n/2]) / 2
	}
	return cp[n/2]
}

func meanFloats(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func roundInt(f float64) int {
	return int(math.Round(f))
}

// ---------- API ----------

func callAPI(ctx context.Context, client *anthropic.Client, model, system, prompt string) (*TrialResult, error) {
	delays := []time.Duration{5 * time.Second, 10 * time.Second, 20 * time.Second}
	maxRetries := 3
	for attempt := 0; attempt <= maxRetries; attempt++ {
		msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.F(anthropic.Model(model)),
			MaxTokens: anthropic.F(int64(4096)),
			System: anthropic.F([]anthropic.TextBlockParam{
				{Text: anthropic.F(system)},
			}),
			Messages: anthropic.F([]anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			}),
		})
		if err != nil {
			var apiErr *anthropic.Error
			if errors.As(err, &apiErr) && apiErr.StatusCode == 429 && attempt < maxRetries {
				delay := delays[min(attempt, len(delays)-1)]
				fmt.Fprintf(os.Stderr, "  Rate limited, retrying in %v...\n", delay)
				time.Sleep(delay)
				continue
			}
			return nil, err
		}
		text := ""
		if len(msg.Content) > 0 {
			text = msg.Content[0].Text
		}
		return &TrialResult{
			InputTokens:  msg.Usage.InputTokens,
			OutputTokens: msg.Usage.OutputTokens,
			Text:         text,
			StopReason:   string(msg.StopReason),
		}, nil
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------- benchmark ----------

func runBenchmarks(ctx context.Context, client *anthropic.Client, model string, prompts []PromptEntry, iceageSystem string, trials int) []BenchmarkEntry {
	results := make([]BenchmarkEntry, 0, len(prompts))
	total := len(prompts)
	for i, p := range prompts {
		entry := BenchmarkEntry{
			ID:       p.ID,
			Category: p.Category,
			Prompt:   p.Prompt,
		}
		for _, mode := range []string{"normal", "iceage"} {
			system := normalSystem
			if mode == "iceage" {
				system = iceageSystem
			}
			for t := 1; t <= trials; t++ {
				fmt.Fprintf(os.Stderr, "  [%d/%d] %s | %s | trial %d/%d\n", i+1, total, p.ID, mode, t, trials)
				result, err := callAPI(ctx, client, model, system, p.Prompt)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
					result = &TrialResult{}
				}
				if mode == "normal" {
					entry.Normal = append(entry.Normal, *result)
				} else {
					entry.Iceage = append(entry.Iceage, *result)
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
		results = append(results, entry)
	}
	return results
}

func computeStats(results []BenchmarkEntry) ([]StatRow, Summary) {
	rows := make([]StatRow, 0, len(results))
	allSavings := make([]float64, 0, len(results))

	for _, entry := range results {
		normalTokens := make([]int64, len(entry.Normal))
		for i, t := range entry.Normal {
			normalTokens[i] = t.OutputTokens
		}
		iceageTokens := make([]int64, len(entry.Iceage))
		for i, t := range entry.Iceage {
			iceageTokens[i] = t.OutputTokens
		}
		normalMed := medianInts(normalTokens)
		iceageMed := medianInts(iceageTokens)
		savings := 0.0
		if normalMed > 0 {
			savings = 1 - (iceageMed / normalMed)
		}
		allSavings = append(allSavings, savings)
		rows = append(rows, StatRow{
			ID:           entry.ID,
			Category:     entry.Category,
			Prompt:       entry.Prompt,
			NormalMedian: roundInt(normalMed),
			IceageMedian: roundInt(iceageMed),
			SavingsPct:   roundInt(savings * 100),
		})
	}

	normalMeds := make([]float64, len(rows))
	iceageMeds := make([]float64, len(rows))
	for i, r := range rows {
		normalMeds[i] = float64(r.NormalMedian)
		iceageMeds[i] = float64(r.IceageMedian)
	}

	minS, maxS := allSavings[0], allSavings[0]
	for _, v := range allSavings[1:] {
		if v < minS {
			minS = v
		}
		if v > maxS {
			maxS = v
		}
	}

	summary := Summary{
		AvgSavings: roundInt(meanFloats(allSavings) * 100),
		MinSavings: roundInt(minS * 100),
		MaxSavings: roundInt(maxS * 100),
		AvgNormal:  roundInt(meanFloats(normalMeds)),
		AvgIceage:  roundInt(meanFloats(iceageMeds)),
	}
	return rows, summary
}

// ---------- labels + table ----------

var promptLabels = map[string]string{
	"react-rerender":        "Explain React re-render bug",
	"auth-middleware-fix":   "Fix auth middleware token expiry",
	"postgres-pool":         "Set up PostgreSQL connection pool",
	"git-rebase-merge":      "Explain git rebase vs merge",
	"async-refactor":        "Refactor callback to async/await",
	"microservices-monolith": "Architecture: microservices vs monolith",
	"pr-security-review":    "Review PR for security issues",
	"docker-multi-stage":    "Docker multi-stage build",
	"race-condition-debug":  "Debug PostgreSQL race condition",
	"error-boundary":        "Implement React error boundary",
}

func formatPromptLabel(id string) string {
	if label, ok := promptLabels[id]; ok {
		return label
	}
	return id
}

func formatTable(rows []StatRow, summary Summary) string {
	var sb strings.Builder
	sb.WriteString("| Task | Normal (tokens) | Iceage (tokens) | Saved |\n")
	sb.WriteString("|------|---------------:|----------------:|------:|\n")
	for _, r := range rows {
		fmt.Fprintf(&sb, "| %s | %d | %d | %d%% |\n",
			formatPromptLabel(r.ID), r.NormalMedian, r.IceageMedian, r.SavingsPct)
	}
	fmt.Fprintf(&sb, "| **Average** | **%d** | **%d** | **%d%%** |\n",
		summary.AvgNormal, summary.AvgIceage, summary.AvgSavings)
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "*Range: %d%%–%d%% savings across prompts.*", summary.MinSavings, summary.MaxSavings)
	return sb.String()
}

// ---------- save + readme ----------

func saveResults(results []BenchmarkEntry, rows []StatRow, summary Summary, model string, trials int, skillHash string) (string, error) {
	var out ResultsOutput
	out.Metadata.ScriptVersion = scriptVersion
	out.Metadata.Model = model
	out.Metadata.Date = time.Now().UTC().Format(time.RFC3339)
	out.Metadata.Trials = trials
	out.Metadata.SkillMDSHA256 = skillHash
	out.Summary = summary
	out.Rows = rows
	out.Raw = results

	ts := time.Now().UTC().Format("20060102_150405")
	resultsDir := filepath.Join(scriptDir(), "results")
	_ = os.MkdirAll(resultsDir, 0755)
	path := filepath.Join(resultsDir, fmt.Sprintf("benchmark_%s.json", ts))

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0644)
}

func updateReadme(tableMarkdown string) error {
	readmePath := filepath.Join(repoDir(), "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	text := string(content)
	startIdx := strings.Index(text, benchmarkStart)
	endIdx := strings.Index(text, benchmarkEnd)
	if startIdx == -1 || endIdx == -1 {
		return fmt.Errorf("benchmark markers not found in README.md")
	}
	before := text[:startIdx+len(benchmarkStart)]
	after := text[endIdx:]
	newContent := before + "\n" + tableMarkdown + "\n" + after
	return os.WriteFile(readmePath, []byte(newContent), 0644)
}

func sha256File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// ---------- dry run ----------

func dryRun(prompts []PromptEntry, model string, trials int) {
	fmt.Printf("Model:  %s\n", model)
	fmt.Printf("Trials: %d\n", trials)
	fmt.Printf("Prompts: %d\n", len(prompts))
	fmt.Printf("Total API calls: %d\n", len(prompts)*2*trials)
	fmt.Println()
	for _, p := range prompts {
		preview := p.Prompt
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		fmt.Printf("  [%s] (%s)\n    %s\n", p.ID, p.Category, preview)
	}
	fmt.Println()
	fmt.Println("Dry run complete. No API calls made.")
}

// ---------- main ----------

func main() {
	trials := flag.Int("trials", 3, "Trials per prompt per mode")
	doDryRun := flag.Bool("dry-run", false, "Print config, no API calls")
	doUpdateReadme := flag.Bool("update-readme", false, "Update README.md benchmark table")
	model := flag.String("model", "claude-sonnet-4-20250514", "Model to use")
	flag.Parse()

	loadEnvFile(filepath.Join(repoDir(), ".env.local"))

	promptsPath := filepath.Join(scriptDir(), "prompts.json")
	data, err := os.ReadFile(promptsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read prompts.json: %v\n", err)
		os.Exit(1)
	}
	var pf PromptsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot parse prompts.json: %v\n", err)
		os.Exit(1)
	}

	if *doDryRun {
		dryRun(pf.Prompts, *model, *trials)
		return
	}

	skillPath := filepath.Join(repoDir(), "skills", "iceage", "SKILL.md")
	skillData, err := os.ReadFile(skillPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read SKILL.md: %v\n", err)
		os.Exit(1)
	}
	iceageSystem := string(skillData)
	skillHash, _ := sha256File(skillPath)

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	var client *anthropic.Client
	if apiKey != "" {
		c := anthropic.NewClient(option.WithAPIKey(apiKey))
		client = c
	} else {
		c := anthropic.NewClient()
		client = c
	}

	fmt.Fprintf(os.Stderr, "Running benchmarks: %d prompts x 2 modes x %d trials\n", len(pf.Prompts), *trials)
	fmt.Fprintf(os.Stderr, "Model: %s\n\n", *model)

	ctx := context.Background()
	results := runBenchmarks(ctx, client, *model, pf.Prompts, iceageSystem, *trials)
	rows, summary := computeStats(results)
	tableMarkdown := formatTable(rows, summary)

	jsonPath, err := saveResults(results, rows, summary, *model, *trials, skillHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: could not save results: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "\nResults saved to %s\n", jsonPath)
	}

	if *doUpdateReadme {
		if err := updateReadme(tableMarkdown); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR updating README: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "README.md updated.")
	}

	fmt.Println(tableMarkdown)
}
