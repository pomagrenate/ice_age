// Generate a boxplot HTML showing token compression distribution per skill.
// Reads evals/snapshots/results.json, writes evals/snapshots/results.html.
//
// Uses Plotly.js (CDN) — no Python/kaleido dependency.
//
// Usage: go run ./cmd/plot
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
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
		GeneratedAt string `json:"generated_at"`
		Model       string `json:"model"`
		NPrompts    int    `json:"n_prompts"`
	} `json:"metadata"`
	Arms map[string][]string `json:"arms"`
}

type skillData struct {
	Name    string    `json:"name"`
	Savings []float64 `json:"savings"`
	Median  float64   `json:"median"`
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

func maxFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

var plotTemplate = template.Must(template.New("plot").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Iceage Skill Compression Distribution</title>
<script src="https://cdn.jsdelivr.net/npm/plotly.js@2/dist/plotly.min.js"></script>
<style>
  body { font-family: sans-serif; margin: 24px; background: #fff; }
  h2 { color: #2c3e50; }
  .subtitle { color: #555; font-size: 13px; margin-top: -8px; }
</style>
</head>
<body>
<h2>How much shorter does each skill make Claude's answers?</h2>
<p class="subtitle">Distribution of per-prompt savings vs system prompt = <em>"Answer concisely."</em><br>
{{.Meta.Model}} &middot; n={{.Meta.NPrompts}} prompts &middot; single run per arm &middot; generated {{.Meta.GeneratedAt}}</p>
<div id="chart"></div>
<script>
var skills = {{.SkillsJSON}};
var traces = skills.map(function(s) {
  return {
    type: "box",
    y: s.savings,
    name: s.name,
    boxpoints: "all",
    jitter: 0.4,
    pointpos: 0,
    marker: {color: "#2ca02c", size: 7, opacity: 0.7},
    line: {color: "#2c3e50", width: 2},
    fillcolor: "rgba(76,120,168,0.25)",
    boxmean: true,
    hovertemplate: "<b>%{x}</b><br>%{y:.1f}%<extra></extra>"
  };
});
var annotations = skills.map(function(s) {
  return {
    x: s.name,
    y: Math.max.apply(null, s.savings),
    text: "<b>" + (s.median >= 0 ? "+" : "") + Math.round(s.median) + "%</b>",
    showarrow: false,
    yshift: 22,
    font: {size: 16, color: "#2c3e50"}
  };
});
annotations.push({
  x: 0.5, y: -0.22, xref: "paper", yref: "paper",
  showarrow: false,
  font: {size: 11, color: "#555"},
  text: "<b>box</b> = IQR (middle 50%) &middot; <b>line in box</b> = median &middot; <b>dashed line</b> = mean &middot; <b>green dots</b> = individual prompts"
});
Plotly.newPlot("chart", traces, {
  shapes: [{type: "line", x0: 0, x1: 1, xref: "paper", y0: 0, y1: 0,
             line: {color: "black", width: 1.5, dash: "dash"}}],
  annotations: annotations,
  xaxis: {title: "", automargin: true},
  yaxis: {title: "↑ shorter  ·  vs control  ·  longer ↓", ticksuffix: "%",
          zeroline: false, gridcolor: "rgba(0,0,0,0.08)", range: [-30, 115]},
  plot_bgcolor: "white",
  height: 560,
  width: 980,
  margin: {l: 140, r: 80, t: 40, b: 120},
  showlegend: false
});
</script>
</body>
</html>
`))

type tmplData struct {
	Meta struct {
		Model       string
		NPrompts    int
		GeneratedAt string
	}
	SkillsJSON template.JS
}

func main() {
	snapshotPath := filepath.Join(evalsDir(), "snapshots", "results.json")
	htmlOut := filepath.Join(evalsDir(), "snapshots", "results.html")

	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read snapshot: %v\n", err)
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

	terseOutputs := snap.Arms["__terse__"]
	terseTokens := make([]int, len(terseOutputs))
	for i, o := range terseOutputs {
		terseTokens[i] = countTokens(enc, o)
	}

	var skills []skillData
	for skill, outputs := range snap.Arms {
		if skill == "__baseline__" || skill == "__terse__" {
			continue
		}
		savings := make([]float64, len(outputs))
		for i, o := range outputs {
			t := terseTokens[i]
			if t > 0 {
				savings[i] = (1 - float64(countTokens(enc, o))/float64(t)) * 100
			}
		}
		skills = append(skills, skillData{
			Name:    skill,
			Savings: savings,
			Median:  median(savings),
		})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Median > skills[j].Median })

	// Clamp y-range: ensure annotations don't overflow
	for i := range skills {
		for j := range skills[i].Savings {
			skills[i].Savings[j] = math.Round(skills[i].Savings[j]*10) / 10
		}
	}

	skillsJSON, err := json.Marshal(skills)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot marshal skills: %v\n", err)
		os.Exit(1)
	}

	td := tmplData{}
	td.Meta.Model = snap.Metadata.Model
	td.Meta.NPrompts = snap.Metadata.NPrompts
	td.Meta.GeneratedAt = snap.Metadata.GeneratedAt
	td.SkillsJSON = template.JS(skillsJSON)

	f, err := os.Create(htmlOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot create output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := plotTemplate.Execute(f, td); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: template error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %s\n", htmlOut)
}
