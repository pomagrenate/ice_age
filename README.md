<div align="center">
  <img src="assets/iceage.svg" alt="iceage" width="120" />
  <h1>iceage</h1>
  <p><strong>AI agents that speak like mammoth hunters. Less words. Same brain.</strong></p>

  ![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
  ![Node](https://img.shields.io/badge/Node-hooks-339933?logo=node.js&logoColor=white)
  ![Claude Code](https://img.shields.io/badge/Claude_Code-plugin-D97706?logo=anthropic&logoColor=white)
  ![Cursor](https://img.shields.io/badge/Cursor-rule-black?logo=cursor&logoColor=white)
  ![Windsurf](https://img.shields.io/badge/Windsurf-rule-0EA5E9)
  ![Codex](https://img.shields.io/badge/Codex-plugin-412991?logo=openai&logoColor=white)
  ![Gemini](https://img.shields.io/badge/Gemini_CLI-extension-4285F4?logo=google&logoColor=white)
</div>

---

## What It Does

Iceage makes your AI coding agent respond in compressed, terse prose — no filler, no hedging, no pleasantries. Technical accuracy stays. Token cost drops. Code, commits, and security warnings stay normal.

**Before:**
> "Sure! I'd be happy to help you with that. The issue you're experiencing is most likely caused by the fact that you're creating a new object reference on every render, which React interprets as a changed prop even though the values are the same. To resolve this, you should wrap the object in a `useMemo` hook."

**After (full mode):**
> "New object ref each render. Inline object prop = new ref = re-render. Wrap in `useMemo`."

Same fix. ~75% fewer tokens. Every response.

---

## Install

### Claude Code (auto-activates, hooks + statusline badge)

```bash
claude plugin install pomagrenate/ice_age
```

Or standalone hooks only (no plugin system needed):

```bash
bash hooks/install.sh        # macOS / Linux
hooks\install.ps1            # Windows PowerShell
```

### Cursor / Windsurf / Cline / Copilot (always-on rule)

Rules are already in this repo. Clone it into your project or workspace:

```bash
git clone https://github.com/pomagrenate/ice_age
```

Cursor picks up `.cursor/rules/iceage.mdc`, Windsurf picks up `.windsurf/rules/iceage.md`, Cline reads `.clinerules/iceage.md`, Copilot reads `.github/copilot-instructions.md` — automatically.

### Codex

```bash
claude plugin install pomagrenate/ice_age   # same plugin, Codex-compatible
```

### Gemini CLI

```bash
gemini extensions install pomagrenate/ice_age
```

### Any other agent (`npx skills`)

```bash
npx skills add pomagrenate/ice_age
```

Then say `/iceage` at the start of each session to activate.

---

## What You Get

| Feature | What it does |
|---------|-------------|
| **Terse prose** | Drops articles, filler, pleasantries, hedging every response |
| **6 intensity levels** | `lite` → `full` → `ultra` → three wenyan (classical Chinese) modes |
| **Auto-clarity** | Reverts to normal for security warnings, destructive ops, confusion — resumes after |
| **Statusline badge** | `[ICEAGE]` / `[ICEAGE:ULTRA]` in Claude Code status bar |
| **Mode persistence** | Stays active across turns, survives context compression |
| **iceage-compress** | Compresses your existing `.md` files to iceage style via Claude API |
| **iceage-commit** | Conventional Commits, ≤50 char subject, terse body |
| **iceage-review** | One-line code review comments: `L42: error Token check inverted. Use \`exp * 1000\`.` |

---

## Intensity Levels

Switch anytime mid-conversation:

```
/iceage lite      — tight but full sentences, no filler
/iceage           — default full: fragments OK, drop articles
/iceage ultra     — max compression, abbreviations, arrows for causality
/iceage wenyan    — classical Chinese (文言文) mode
```

**Example — "Explain database connection pooling."**

| Level | Response |
|-------|----------|
| `lite` | Connection pooling reuses open connections instead of creating new ones per request. Avoids repeated handshake overhead. |
| `full` | Pool reuse open DB connections. No new connection per request. Skip handshake overhead. |
| `ultra` | Pool = reuse DB conn. Skip handshake → fast under load. |
| `wenyan-full` | 池reuse open connection。不每req新開。skip handshake overhead。 |

---

## Commands

| Command | Effect |
|---------|--------|
| `/iceage` | Activate full mode |
| `/iceage lite` | Activate lite mode |
| `/iceage ultra` | Activate ultra mode |
| `/iceage wenyan` | Activate wenyan-full mode |
| `stop iceage` | Return to normal |
| `normal mode` | Return to normal |
| `/iceage-commit` | One-shot: terse Conventional Commit message |
| `/iceage-review` | One-shot: terse code review |
| `/iceage:compress <file>` | Compress a markdown file in place |

Natural language works too: "activate iceage", "talk like iceage", "turn off iceage".

---

## iceage-compress

Compresses existing markdown documentation to iceage style. Sends file to Claude API, validates headings/code blocks/URLs are preserved, retries on failure.

```bash
# With backup (default)
go run ./iceage-compress/go README.md

# No backup — good for version-controlled files
go run ./iceage-compress/go --no-backup README.md
```

Requires `ANTHROPIC_API_KEY` in environment or `.env.local`. Falls back to `claude` CLI if no key set.

---

## Want It Always On?

Paste this into any agent's rule file, `CLAUDE.md`, or system prompt:

```
Respond terse like smart mammoth hunter. All technical substance stay. Only fluff die.
Drop: articles, filler (just/really/basically), pleasantries, hedging.
Fragments OK. Short synonyms. Technical terms exact. Code unchanged.
Pattern: [thing] [action] [reason]. [next step].
Stop: "stop iceage" or "normal mode".
```

---

## How Claude Code Hooks Work

Three hooks activate automatically on every session:

```
SessionStart hook
  → writes ~/.claude/.iceage-active
  → injects full ruleset as hidden system context

UserPromptSubmit hook
  → watches for /iceage commands, updates mode
  → sends per-turn reminder to prevent style drift

caveman-statusline.sh / .ps1
  → reads flag file → shows [ICEAGE] badge in status bar
```

Install: `bash hooks/install.sh` · Uninstall: `bash hooks/uninstall.sh`

---

## Benchmarks

Benchmark harness in `benchmarks/` measures real output token counts across 10 prompts × normal vs iceage modes × 3 trials via the Anthropic API.

To run:

```bash
echo "ANTHROPIC_API_KEY=sk-ant-..." > .env.local
cd benchmarks && go run . --trials 3
```

Results saved to `benchmarks/results/`. Evals (token ratio vs terse-only baseline) in `evals/snapshots/`.

---

## Agent Support

| Agent | How | Auto-activates? |
|-------|-----|----------------|
| Claude Code | Plugin (hooks + skills) | Yes — every session |
| Codex | Plugin + `.codex/hooks.json` | Yes — SessionStart hook |
| Gemini CLI | Extension + `GEMINI.md` | Yes — context file |
| Cursor | `.cursor/rules/iceage.mdc` | Yes — always-on rule |
| Windsurf | `.windsurf/rules/iceage.md` | Yes — always-on rule |
| Cline | `.clinerules/iceage.md` | Yes — auto-discovered |
| Copilot | `.github/copilot-instructions.md` | Yes — repo instructions |
| Others | `npx skills add pomagrenate/ice_age` | No — say `/iceage` each session |

<details>
<summary>Claude Code plugin details</summary>

Plugin installs hooks, skills, and statusline automatically. No manual settings.json edits needed.

```bash
claude plugin install pomagrenate/ice_age
```

To uninstall:
```bash
claude plugin disable ice_age
```

</details>

<details>
<summary>Standalone hooks (Claude Code without plugin)</summary>

```bash
bash hooks/install.sh        # macOS / Linux
.\hooks\install.ps1          # Windows
```

Installs hook files to `~/.claude/hooks/`, registers SessionStart + UserPromptSubmit hooks, configures statusline badge. Backs up `settings.json` before touching it.

To uninstall:
```bash
bash hooks/uninstall.sh
.\hooks\uninstall.ps1
```

</details>

<details>
<summary>Codex plugin details</summary>

Plugin config in `plugins/iceage/`. Contains skills and `.codex-plugin/plugin.json`. Hooks wired via `.codex/hooks.json` at repo root for Codex SessionStart injection.

</details>

<details>
<summary>npx skills (Cursor, Windsurf, Cline, Copilot, others)</summary>

```bash
npx skills add pomagrenate/ice_age     # install
npx skills remove pomagrenate/ice_age  # remove
```

Copies rule files to the right location for each detected agent. Then say `/iceage` to activate each session.

</details>
