<div align="center">
  <img src="assets/iceage.png" alt="iceage" width="120" />
  <h1>iceage</h1>
  <p><strong>AI agents that speak like mammoth hunters. Less words. Same brain.</strong></p>

  ![Node](https://img.shields.io/badge/Node-hooks-339933?logo=node.js&logoColor=white)
  ![Claude Code](https://img.shields.io/badge/Claude_Code-hooks+skills-D97706?logo=anthropic&logoColor=white)
  ![Cursor](https://img.shields.io/badge/Cursor-rule-black?logo=cursor&logoColor=white)
  ![Windsurf](https://img.shields.io/badge/Windsurf-rule-0EA5E9)
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

Clone the repo first:

```bash
git clone https://github.com/pomagrenate/iceage.git
cd iceage
```

Then follow the section for your agent below.

---

### Claude Code

Claude Code needs three things: **hooks** (auto-activates on every session), **skills** (gives you `/iceage`, `/iceage-review`, etc. as slash commands), and optionally **commands** (alternative slash command style).

#### Step 1 — Hooks (auto-activation + statusline badge)

Hooks run automatically on every session start. They inject the iceage ruleset as hidden context and track which mode you're in.

**macOS / Linux:**
```bash
bash hooks/install.sh
```

**Windows (PowerShell):**
```powershell
powershell -ExecutionPolicy Bypass -File hooks\install.ps1
```

This copies hook files to `~/.claude/hooks/`, registers `SessionStart` and `UserPromptSubmit` hooks in `~/.claude/settings.json`, and configures the `[ICEAGE]` statusline badge. Backs up `settings.json` before touching it.

To uninstall:
```bash
bash hooks/uninstall.sh          # macOS / Linux
.\hooks\uninstall.ps1            # Windows
```

**Requirements:** Node.js must be installed (`node --version` should work).

#### Step 2 — Skills (slash commands)

Skills give you `/iceage`, `/iceage lite`, `/iceage ultra`, `/iceage-review`, `/iceage-commit`, and `/iceage-help` as slash commands in Claude Code.

**macOS / Linux:**
```bash
mkdir -p ~/.claude/skills/iceage \
         ~/.claude/skills/iceage-review \
         ~/.claude/skills/iceage-commit \
         ~/.claude/skills/iceage-help

cp skills/iceage/SKILL.md        ~/.claude/skills/iceage/SKILL.md
cp skills/iceage-review/SKILL.md ~/.claude/skills/iceage-review/SKILL.md
cp skills/iceage-commit/SKILL.md ~/.claude/skills/iceage-commit/SKILL.md
cp skills/iceage-help/SKILL.md   ~/.claude/skills/iceage-help/SKILL.md
```

**Windows (PowerShell):**
```powershell
$d = "$env:USERPROFILE\.claude\skills"
New-Item -ItemType Directory -Force "$d\iceage","$d\iceage-review","$d\iceage-commit","$d\iceage-help"

Copy-Item skills\iceage\SKILL.md        "$d\iceage\SKILL.md"
Copy-Item skills\iceage-review\SKILL.md "$d\iceage-review\SKILL.md"
Copy-Item skills\iceage-commit\SKILL.md "$d\iceage-commit\SKILL.md"
Copy-Item skills\iceage-help\SKILL.md   "$d\iceage-help\SKILL.md"
```

Restart Claude Code after copying. Skills appear when you type `/` in the chat input.

#### Step 3 — Commands (optional)

Commands are an alternative slash-command mechanism. Install alongside skills or instead of them.

**macOS / Linux:**
```bash
mkdir -p ~/.claude/commands
cp commands/*.toml ~/.claude/commands/
```

**Windows (PowerShell):**
```powershell
New-Item -ItemType Directory -Force "$env:USERPROFILE\.claude\commands"
Copy-Item commands\*.toml "$env:USERPROFILE\.claude\commands\"
```

#### Step 4 — Always-on rule (optional, per project)

To activate iceage automatically in a specific project without hooks, copy the rule file into your project:

```bash
mkdir -p .claude/rules
cp rules/iceage-activate.md .claude/rules/iceage-activate.md
```

Claude Code loads all `.claude/rules/*.md` files as system context on every session in that project.

---

### Cursor

```bash
mkdir -p .cursor/rules
cp .cursor/rules/iceage.mdc .cursor/rules/iceage.mdc
```

Or copy to your global Cursor rules directory (`~/.cursor/rules/` on macOS/Linux). The rule has `alwaysApply: true` — activates on every session automatically.

---

### Windsurf

```bash
mkdir -p .windsurf/rules
cp .windsurf/rules/iceage.md .windsurf/rules/iceage.md
```

The rule has `trigger: always_on` — no manual activation needed.

---

### Cline

```bash
mkdir -p .clinerules
cp .clinerules/iceage.md .clinerules/iceage.md
```

Cline auto-discovers all `.clinerules/*.md` files in the workspace root.

---

### Copilot

```bash
mkdir -p .github
cp .github/copilot-instructions.md .github/copilot-instructions.md
```

Or append the contents of `rules/iceage-activate.md` to your existing `.github/copilot-instructions.md`.

---

### Gemini CLI

```bash
cp GEMINI.md /path/to/your/project/GEMINI.md
```

Or append the contents of `rules/iceage-activate.md` to your existing `GEMINI.md`.

---

### Any other agent

Paste this into the agent's system prompt, rule file, or `CLAUDE.md`:

```
Respond terse like smart mammoth hunter. All technical substance stay. Only fluff die.
Drop: articles, filler (just/really/basically), pleasantries, hedging.
Fragments OK. Short synonyms. Technical terms exact. Code unchanged.
Pattern: [thing] [action] [reason]. [next step].
Stop: "stop iceage" or "normal mode".
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
| **iceage-compress** | Compresses your existing `.md` files to iceage style |
| **iceage-commit** | Conventional Commits, ≤50 char subject, terse body |
| **iceage-review** | One-line code review comments: `L42: 🔴 bug: token check inverted. Use \`exp * 1000\`.` |

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
| `wenyan-full` | 池復用連線。不每請求新開。故速。 |

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
| `/iceage-help` | Show quick reference card |
| `/iceage:compress <file>` | Compress a markdown file in place |

Natural language works too: "activate iceage", "talk like iceage", "turn off iceage".

---

## iceage-compress

Compresses existing markdown documentation to iceage style. Validates headings, code blocks, and URLs are preserved. Retries on failure.

```bash
# With backup (default)
go run ./skills/compress/go README.md

# No backup — good for version-controlled files
go run ./skills/compress/go --no-backup README.md
```

Requires `ANTHROPIC_API_KEY` in environment or `.env.local`. Falls back to `claude` CLI if no key set.

---

## How Claude Code Hooks Work

Three hooks activate automatically on every session:

```
SessionStart hook (hooks/iceage-activate.js)
  → writes ~/.claude/.iceage-active   (tracks current mode)
  → injects full ruleset as hidden system context (invisible to user)
  → nudges setup if statusline not yet configured

UserPromptSubmit hook (hooks/iceage-mode-tracker.js)
  → watches for /iceage commands → updates mode flag file
  → matches natural-language triggers ("activate iceage", "stop iceage")
  → sends per-turn reminder to prevent style drift mid-conversation

iceage-statusline.sh / iceage-statusline.ps1
  → reads flag file → outputs [ICEAGE] or [ICEAGE:ULTRA] badge in status bar
```

All hooks silent-fail on filesystem errors. Session start is never blocked.

**Configure default mode** (optional):

```bash
# Environment variable (highest priority)
export ICEAGE_DEFAULT_MODE=lite   # or: ultra, wenyan, off

# Config file
echo '{ "defaultMode": "lite" }' > ~/.config/iceage/config.json
```

Set `"off"` to disable auto-activation. You can still say `/iceage` manually each session.

---

## Agent Support

| Agent | Mechanism | Auto-activates? |
|-------|-----------|----------------|
| Claude Code | Hooks (`hooks/install.sh`) + Skills (`~/.claude/skills/`) | Yes — every session |
| Cursor | `.cursor/rules/iceage.mdc` (`alwaysApply: true`) | Yes — always-on rule |
| Windsurf | `.windsurf/rules/iceage.md` (`trigger: always_on`) | Yes — always-on rule |
| Cline | `.clinerules/iceage.md` (auto-discovered) | Yes — auto-discovered |
| Copilot | `.github/copilot-instructions.md` | Yes — repo instructions |
| Gemini CLI | `GEMINI.md` context file | Yes — context file |
| Others | Paste `rules/iceage-activate.md` into system prompt | No — say `/iceage` each session |

---

## Benchmarks

Benchmark harness in `benchmarks/` measures real output token counts across 10 prompts × normal vs iceage modes × 3 trials via the Anthropic API.

To run:

```bash
echo "ANTHROPIC_API_KEY=sk-ant-..." > .env.local
cd benchmarks && go run . --trials 3
```

Results saved to `benchmarks/results/`. Evals (token ratio vs terse-only baseline) in `evals/snapshots/`.
