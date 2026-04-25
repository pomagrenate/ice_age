// Read evals/snapshots/results.json and report token compression per skill
// against the terse control arm.
//
// Usage: go run ./cmd/measure
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

func evalsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

type Snapshot struct {
	Metadata struct {
		GeneratedAt      string `json:"generated_at"`
		Model            string `json:"model"`
		ClaudeCLIVersion string `json:"claude_cli_version"`
		NPrompts         int    `json:"n_prompts"`
	} `json:"metadata"`
	Arms map[string][]string `json:"arms"`
}

func countTokens(enc *tiktoken.Tiktoken, text string) int {
	return len(enc.Encode(text, nil, nil))
}

func median(vals []float64) float64 {
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

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stdev(vals []float64) float64 {
	if len(vals) <= 1 {
		return 0
	}
	m := mean(vals)
	sum := 0.0
	for _, v := range vals {
		d := v - m
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(vals)-1))
}

func sumInts(vals []int) int {
	s := 0
	for _, v := range vals {
		s += v
	}
	return s
}

func fmtPct(x float64) string {
	sign := "+"
	if x < 0 {
		sign = "−"
	}
	return fmt.Sprintf("%s%.0f%%", sign, math.Abs(x)*100)
}

func main() {
	snapshotPath := filepath.Join(evalsDir(), "snapshots", "results.json")

	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		fmt.Printf("No snapshot at %s. Run `go run ./cmd/llm_run` first.\n", snapshotPath)
		os.Exit(0)
	}

	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot parse snapshot: %v\n", err)
		os.Exit(1)
	}

	enc, err := tiktoken.GetEncoding("o200k_base")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot load tokenizer: %v\n", err)
		os.Exit(1)
	}

	tokenize := func(outputs []string) []int {
		counts := make([]int, len(outputs))
		for i, o := range outputs {
			counts[i] = countTokens(enc, o)
		}
		return counts
	}

	baselineTokens := tokenize(snap.Arms["__baseline__"])
	terseTokens := tokenize(snap.Arms["__terse__"])

	m := snap.Metadata
	fmt.Printf("_Generated: %s_\n", m.GeneratedAt)
	fmt.Printf("_Model: %s · CLI: %s_\n", m.Model, m.ClaudeCLIVersion)
	fmt.Println("_Tokenizer: tiktoken o200k_base (approximation of Claude's BPE)_")
	nPrompts := m.NPrompts
	if nPrompts == 0 {
		nPrompts = len(baselineTokens)
	}
	fmt.Printf("_n = %d prompts, single run per arm_\n", nPrompts)
	fmt.Println()
	fmt.Println("**Reference arms (no skill):**")
	totalBaseline := sumInts(baselineTokens)
	totalTerse := sumInts(terseTokens)
	fmt.Printf("- baseline (no system prompt): %d tokens total\n", totalBaseline)
	terseDelta := 0.0
	if totalBaseline > 0 {
		terseDelta = 1 - float64(totalTerse)/float64(totalBaseline)
	}
	fmt.Printf("- terse control (`Answer concisely.`): %d tokens total (%s vs baseline)\n",
		totalTerse, fmtPct(terseDelta))
	fmt.Println()
	fmt.Println("**Skills, measured as additional reduction on top of the terse control:**")
	fmt.Println()
	fmt.Println("| Skill | Median | Mean | Min | Max | Stdev | Tokens (skill / terse) |")
	fmt.Println("|-------|--------|------|-----|-----|-------|-------------------------|")

	type row struct {
		skill              string
		med, mn, lo, hi, sd float64
		skillTotal         int
		terseTotal         int
	}
	var rows []row

	for skill, outputs := range snap.Arms {
		if skill == "__baseline__" || skill == "__terse__" {
			continue
		}
		skillTokens := tokenize(outputs)
		savings := make([]float64, len(skillTokens))
		for i, s := range skillTokens {
			t := terseTokens[i]
			if t > 0 {
				savings[i] = 1 - float64(s)/float64(t)
			}
		}
		med := median(savings)
		mn := mean(savings)
		lo, hi := savings[0], savings[0]
		for _, v := range savings[1:] {
			if v < lo {
				lo = v
			}
			if v > hi {
				hi = v
			}
		}
		sd := stdev(savings)
		rows = append(rows, row{skill, med, mn, lo, hi, sd, sumInts(skillTokens), totalTerse})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].med > rows[j].med })
	for _, r := range rows {
		fmt.Printf("| **%s** | %s | %s | %s | %s | %.0f%% | %d / %d |\n",
			r.skill, fmtPct(r.med), fmtPct(r.mn), fmtPct(r.lo), fmtPct(r.hi),
			r.sd*100, r.skillTotal, r.terseTotal)
	}

	fmt.Println()
	fmt.Println("_Savings = `1 - skill_tokens / terse_tokens` per prompt._")
	fmt.Printf("_Source: %s. Refresh with `go run ./cmd/llm_run`._\n", filepath.Base(snapshotPath))
}
