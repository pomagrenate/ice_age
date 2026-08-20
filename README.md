<div align="center">
  <img src="assets/iceage.png" alt="Iceage AI Agent Compression" width="120" />

  <h1>iceage</h1>

  <p>
    <strong>Make AI coding agents terse, precise, and token-efficient.</strong>
  </p>

  <p>
    Less filler. Same technical depth. Lower token usage.
  </p>

  <p>
    <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go" />
    <img src="https://img.shields.io/badge/Claude_Code-hooks_%2B_skills-D97706?logo=anthropic&logoColor=white" alt="Claude Code" />
    <img src="https://img.shields.io/badge/Cursor-rules-000000?logo=cursor&logoColor=white" alt="Cursor" />
    <img src="https://img.shields.io/badge/Windsurf-rules-0EA5E9" alt="Windsurf" />
    <img src="https://img.shields.io/badge/Gemini_CLI-extension-4285F4?logo=google&logoColor=white" alt="Gemini CLI" />
  </p>
  <p>
    <a href="#features">Features</a> ·
    <a href="#installation">Installation</a> ·
    <a href="#supported-ai-coding-agents">AI Agents</a> ·
    <a href="#modes">Modes</a> ·
    <a href="#commands">Commands</a> ·
    <a href="#benchmarks">Benchmarks</a>
  </p>
</div>

---

## What is Iceage?

**Iceage** is a lightweight prompt and agent-style optimization toolkit that makes AI coding assistants communicate in **compressed, technical, high-signal prose**.

Modern AI coding agents often spend a surprising number of tokens on:

- Pleasantries
- Repeated explanations
- Filler words
- Hedging
- Redundant context
- Long-form explanations when a short technical answer is enough

Iceage removes that overhead while preserving the information that matters.

### Before

> "Sure! I'd be happy to help you with that. The issue you're experiencing is most likely caused by the fact that you're creating a new object reference on every render, which React interprets as a changed prop even though the values are the same. To resolve this, you should wrap the object in a `useMemo` hook."

### After

> "New object ref each render. Inline object prop = new ref = re-render. Wrap in `useMemo`."

**Same diagnosis. Same fix. Significantly fewer tokens.**

Iceage is designed specifically for developers who use AI coding agents throughout the day and want **less conversational overhead without sacrificing technical accuracy**.

---

## Why Iceage?

AI agents are powerful, but their default communication style is often optimized for general conversation rather than high-frequency software engineering.

Iceage changes the communication layer.

```text
Normal AI Agent
      │
      ▼
Long explanations
Pleasantries
Hedging
Repetition
      │
      ▼
     Iceage
      │
      ▼
Compressed technical response
      │
      ├── Less filler
      ├── Fewer tokens
      ├── Faster scanning
      └── Same technical substance
````

The goal is simple:

> **Remove words. Keep meaning.**

---

## Features

| Feature                           | Description                                                                                            |
| --------------------------------- | ------------------------------------------------------------------------------------------------------ |
| **Terse AI responses**            | Removes filler, pleasantries, unnecessary articles, and hedging                                        |
| **Token-efficient communication** | Compresses natural-language responses while preserving technical meaning                               |
| **6 intensity levels**            | Choose between `lite`, `full`, `ultra`, and Wenyan modes                                               |
| **Automatic clarity fallback**    | Returns to normal communication for security warnings, destructive operations, or ambiguous situations |
| **Session persistence**           | Keeps Iceage active across turns and context compression                                               |
| **Claude Code hooks**             | Automatically activates Iceage and tracks the current mode                                             |
| **Claude Code skills**            | Provides `/iceage`, `/iceage-review`, `/iceage-commit`, and more                                       |
| **Cursor rules**                  | Always-on Iceage rules for Cursor                                                                      |
| **Windsurf rules**                | Always-on Iceage rules for Windsurf                                                                    |
| **Cline support**                 | Workspace-level `.clinerules` integration                                                              |
| **GitHub Copilot support**        | Repository-level Copilot instructions                                                                  |
| **Gemini CLI support**            | Project-level `GEMINI.md` integration                                                                  |
| **Code review mode**              | Generates concise, actionable review comments                                                          |
| **Commit mode**                   | Generates terse Conventional Commit messages                                                           |
| **Markdown compression**          | Compress existing Markdown documentation while preserving structure                                    |
| **Statusline indicator**          | Shows the current Iceage mode directly in Claude Code                                                  |

---

## Installation

Clone the repository:

```bash
git clone https://github.com/pomagrenate/iceage.git
cd iceage
```

Then install Iceage for the AI coding agent you use.

---

# Supported AI Coding Agents

## Claude Code

Claude Code has the deepest Iceage integration through:

* Hooks
* Skills
* Commands
* Project rules
* Statusline integration

### 1. Install Hooks

Hooks automatically activate Iceage when a Claude Code session starts.

#### macOS / Linux

```bash
bash hooks/install.sh
```

#### Windows PowerShell

```powershell
powershell -ExecutionPolicy Bypass -File hooks\install.ps1
```

The installer:

* Copies hooks into `~/.claude/hooks/`
* Registers `SessionStart`
* Registers `UserPromptSubmit`
* Configures the `[ICEAGE]` statusline badge
* Backs up your existing `settings.json`

### Uninstall

#### macOS / Linux

```bash
bash hooks/uninstall.sh
```

#### Windows PowerShell

```powershell
.\hooks\uninstall.ps1
```

### Requirement

Node.js must be available:

```bash
node --version
```

---

## 2. Install Claude Code Skills

Iceage provides slash commands for Claude Code.

Available skills:

```text
/iceage
/iceage-review
/iceage-commit
/iceage-help
```

### macOS / Linux

```bash
mkdir -p ~/.claude/skills/iceage \
         ~/.claude/skills/iceage-review \
         ~/.claude/skills/iceage-commit \
         ~/.claude/skills/iceage-help

cp skills/iceage/SKILL.md \
   ~/.claude/skills/iceage/SKILL.md

cp skills/iceage-review/SKILL.md \
   ~/.claude/skills/iceage-review/SKILL.md

cp skills/iceage-commit/SKILL.md \
   ~/.claude/skills/iceage-commit/SKILL.md

cp skills/iceage-help/SKILL.md \
   ~/.claude/skills/iceage-help/SKILL.md
```

### Windows PowerShell

```powershell
$d = "$env:USERPROFILE\.claude\skills"

New-Item -ItemType Directory -Force `
  "$d\iceage",
  "$d\iceage-review",
  "$d\iceage-commit",
  "$d\iceage-help"

Copy-Item skills\iceage\SKILL.md `
  "$d\iceage\SKILL.md"

Copy-Item skills\iceage-review\SKILL.md `
  "$d\iceage-review\SKILL.md"

Copy-Item skills\iceage-commit\SKILL.md `
  "$d\iceage-commit\SKILL.md"

Copy-Item skills\iceage-help\SKILL.md `
  "$d\iceage-help\SKILL.md"
```

Restart Claude Code.

Then type `/` in the chat input to see the available Iceage commands.

---

## 3. Install Claude Code Commands

Commands are an alternative to Claude Code skills.

### macOS / Linux

```bash
mkdir -p ~/.claude/commands
cp commands/*.toml ~/.claude/commands/
```

### Windows PowerShell

```powershell
New-Item -ItemType Directory -Force "$env:USERPROFILE\.claude\commands"
Copy-Item commands\*.toml "$env:USERPROFILE\.claude\commands\"
```

You can use commands alongside skills or instead of them.

---

## 4. Enable Iceage Per Project

If you want Iceage automatically enabled for a specific project without using global hooks:

```bash
mkdir -p .claude/rules
cp rules/iceage-activate.md .claude/rules/iceage-activate.md
```

Claude Code automatically loads `.claude/rules/*.md` as project context.

---

# Cursor

Iceage can run as an always-on Cursor rule.

```bash
mkdir -p .cursor/rules
cp .cursor/rules/iceage.mdc .cursor/rules/iceage.mdc
```

You can also install the rule globally:

```text
~/.cursor/rules/
```

The rule uses:

```yaml
alwaysApply: true
```

This activates Iceage automatically for Cursor sessions.

---

# Windsurf

Install the Iceage rule:

```bash
mkdir -p .windsurf/rules
cp .windsurf/rules/iceage.md .windsurf/rules/iceage.md
```

The rule uses:

```yaml
trigger: always_on
```

No manual activation is required.

---

# Cline

Install Iceage into your workspace:

```bash
mkdir -p .clinerules
cp .clinerules/iceage.md .clinerules/iceage.md
```

Cline automatically discovers `.clinerules/*.md` files in the workspace root.

---

# GitHub Copilot

Install the repository-level Copilot instructions:

```bash
mkdir -p .github
cp .github/copilot-instructions.md .github/copilot-instructions.md
```

Alternatively, append the contents of:

```text
rules/iceage-activate.md
```

to your existing:

```text
.github/copilot-instructions.md
```

---

# Gemini CLI

Copy the Iceage context file into your project:

```bash
cp GEMINI.md /path/to/your/project/GEMINI.md
```

Alternatively, append the contents of:

```text
rules/iceage-activate.md
```

to your existing `GEMINI.md`.

---

# Other AI Coding Agents

Iceage can also work with other AI coding assistants that support system prompts, rule files, or project instructions.

Use the following core rules:

```text
Respond terse like smart mammoth hunter. All technical substance stay. Only fluff die.

Drop: articles, filler (just/really/basically), pleasantries, hedging.

Fragments OK. Short synonyms. Technical terms exact. Code unchanged.

Pattern: [thing] [action] [reason]. [next step].

Stop: "stop iceage" or "normal mode".
```

Then activate it with:

```text
/iceage
```

---

# Modes

Iceage supports multiple communication intensity levels.

Switch modes at any time during a conversation.

```text
/iceage lite
/iceage
/iceage ultra
/iceage wenyan
```

## Lite

Tight writing while keeping complete sentences.

```text
Connection pooling reuses open connections instead of creating new ones per request. Avoids repeated handshake overhead.
```

## Full

Default Iceage mode.

```text
Pool reuse open DB connections. No new connection per request. Skip handshake overhead.
```

## Ultra

Maximum compression.

```text
Pool = reuse DB conn. Skip handshake → fast under load.
```

## Wenyan

Classical Chinese compression mode.

```text
池復用連線。不每請求新開。故速。
```

---

# Response Compression Example

### Normal

```text
Sure! I'd be happy to help you understand this issue.

The reason this happens is that you're creating a new object reference on every render. React sees that the reference has changed, even if the actual values inside the object remain the same.

To fix this issue, you can use the useMemo hook to memoize the object and prevent a new reference from being created on every render.
```

### Iceage

```text
New object ref each render.

Inline object prop = new ref → React sees changed prop → re-render.

Use `useMemo` to stabilize the reference.
```

The technical meaning stays.

The unnecessary prose disappears.

---

# Commands

| Command                   | Description                                  |
| ------------------------- | -------------------------------------------- |
| `/iceage`                 | Activate full mode                           |
| `/iceage lite`            | Activate lite mode                           |
| `/iceage ultra`           | Activate ultra mode                          |
| `/iceage wenyan`          | Activate Wenyan mode                         |
| `stop iceage`             | Disable Iceage                               |
| `normal mode`             | Return to normal communication               |
| `/iceage-commit`          | Generate a terse Conventional Commit message |
| `/iceage-review`          | Generate concise code review comments        |
| `/iceage-help`            | Display the Iceage quick reference           |
| `/iceage:compress <file>` | Compress a Markdown file                     |

Natural language activation also works:

```text
activate iceage
talk like iceage
turn off iceage
```

---

# iceage-compress

`iceage-compress` compresses existing Markdown documentation into Iceage's concise writing style.

It preserves important document structure, including:

* Headings
* Code blocks
* URLs

The compressor also validates the result and retries when necessary.

### With backup

```bash
go run ./skills/compress/go README.md
```

### Without backup

Useful for files already tracked by Git:

```bash
go run ./skills/compress/go --no-backup README.md
```

### API configuration

Set:

```text
ANTHROPIC_API_KEY
```

in your environment or `.env.local`.

If no API key is configured, Iceage can fall back to the `claude` CLI.

---

# How Claude Code Hooks Work

Iceage uses three components to maintain its active mode across Claude Code sessions.

```text
SessionStart
    │
    ├── Track current Iceage mode
    ├── Inject Iceage rules
    └── Configure statusline
            │
            ▼
      Claude Code Session
            │
            ▼
UserPromptSubmit
    │
    ├── Detect /iceage commands
    ├── Detect natural-language triggers
    ├── Update current mode
    └── Prevent style drift
            │
            ▼
      Iceage Statusline
            │
            ├── [ICEAGE]
            └── [ICEAGE:ULTRA]
```

### SessionStart Hook

```text
hooks/iceage-activate.js
```

Responsible for:

* Tracking the current mode
* Injecting Iceage rules as hidden context
* Nudging initial statusline setup

### UserPromptSubmit Hook

```text
hooks/iceage-mode-tracker.js
```

Responsible for:

* Detecting `/iceage` commands
* Detecting natural-language activation
* Updating the mode state
* Sending per-turn reminders

### Statusline

```text
iceage-statusline.sh
iceage-statusline.ps1
```

Displays the active mode:

```text
[ICEAGE]
```

or:

```text
[ICEAGE:ULTRA]
```

All hooks fail silently on filesystem errors so Iceage never blocks a Claude Code session.

---

# Default Mode

You can configure the default Iceage mode through an environment variable:

```bash
export ICEAGE_DEFAULT_MODE=lite
```

Supported values include:

```text
lite
ultra
wenyan
off
```

Or use the configuration file:

```bash
mkdir -p ~/.config/iceage

echo '{ "defaultMode": "lite" }' \
  > ~/.config/iceage/config.json
```

Use:

```text
off
```

to disable automatic activation while keeping manual `/iceage` activation available.

---

# Supported AI Coding Agents

| AI Coding Agent    | Integration                       | Auto-Activation |
| ------------------ | --------------------------------- | --------------- |
| **Claude Code**    | Hooks + Skills                    | Yes             |
| **Cursor**         | `.cursor/rules/iceage.mdc`        | Yes             |
| **Windsurf**       | `.windsurf/rules/iceage.md`       | Yes             |
| **Cline**          | `.clinerules/iceage.md`           | Yes             |
| **GitHub Copilot** | `.github/copilot-instructions.md` | Yes             |
| **Gemini CLI**     | `GEMINI.md`                       | Yes             |
| **Other agents**   | System prompt / rule file         | Manual          |

---

# Code Review Mode

Iceage can compress code review feedback into short, actionable comments.

Example:

```text
L42: 🔴 bug: token check inverted. Use `exp * 1000`.
```

Instead of a long explanation, the reviewer gets:

```text
location → severity → issue → fix
```

This makes large code reviews easier to scan.

---

# Commit Mode

Use:

```text
/iceage-commit
```

to generate concise Conventional Commit messages.

The goal is:

```text
type(scope): concise description
```

with a terse body when additional context is required.

---

# Documentation Compression

Iceage can also be used beyond AI conversations.

Use:

```bash
go run ./skills/compress/go README.md
```

to compress existing Markdown documentation.

Useful for:

* README files
* Developer documentation
* Technical guides
* Project notes
* Architecture documentation
* Internal engineering docs

The compressor is designed to reduce unnecessary prose while preserving the structure developers need.

---

# Benchmarks

Iceage includes a benchmark harness for measuring real output token counts.

The benchmark evaluates:

```text
10 prompts
×
Normal vs Iceage modes
×
3 trials
```

using the Anthropic API.

Run the benchmark:

```bash
echo "ANTHROPIC_API_KEY=sk-ant-..." > .env.local

cd benchmarks
go run . --trials 3
```

Results are stored in:

```text
benchmarks/results/
```

Evaluation snapshots are stored in:

```text
evals/snapshots/
```

The evaluation compares token ratios against a terse-only baseline.

---

# Philosophy

Iceage is built around one principle:

> **Communication should carry information, not ceremony.**

AI coding assistants already have the intelligence.

The problem is often the interface between that intelligence and the developer.

For engineering workflows, responses should be:

```text
Precise
    ↓
Technical
    ↓
Scannable
    ↓
Minimal
```

Not:

```text
Greeting
    ↓
Disclaimer
    ↓
Repetition
    ↓
Explanation
    ↓
Conclusion
    ↓
Actual answer
```

Iceage keeps the answer.

Everything else is optional.

---

# Who Is Iceage For?

Iceage is designed for developers who:

* Use AI coding agents daily
* Work with Claude Code, Cursor, Windsurf, Cline, Copilot, or Gemini CLI
* Want faster AI interactions
* Prefer concise technical communication
* Want to reduce unnecessary token usage
* Read large volumes of AI-generated code and explanations
* Want AI responses optimized for engineering workflows

If you already think in:

```text
problem → cause → fix
```

Iceage is designed for you.

---

# Design Goals

Iceage prioritizes:

1. **Technical accuracy**
2. **Information density**
3. **Low verbosity**
4. **Fast human scanning**
5. **Predictable behavior**
6. **Minimal configuration**
7. **Compatibility with existing AI coding workflows**

Iceage does **not** aim to make AI responses blindly short.

When additional detail is necessary for correctness, safety, debugging, or ambiguity, Iceage can return to normal communication.

--- 

# Quick Start

```bash
git clone https://github.com/pomagrenate/iceage.git
cd iceage

# Claude Code
bash hooks/install.sh

# Then restart Claude Code
```

Activate:

```text
/iceage
```

Change intensity:

```text
/iceage lite
/iceage ultra
/iceage wenyan
```

Disable:

```text
stop iceage
```

That's it.
