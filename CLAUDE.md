# CLAUDE.md — iceage

## README is a product artifact

README = product front door. Non-technical people read it to decide if iceage worth install. Treat like UI copy.

**Rules for any README change:**

- Readable by non-AI-agent users. If you write "SessionStart hook injects system context," invisible to most — translate it.
- Keep Before/After examples first. That the pitch.
- Install table always complete + accurate. One broken install command costs real user.
- What You Get table must sync with actual code. Feature ships or removed → update table.
- Preserve voice. Iceage speak in README on purpose. "Brain still big." "Cost go down forever." "One rock. That it." — intentional brand. Don't normalize.
- Benchmark numbers from real runs in `benchmarks/` and `evals/`. Never invent or round. Re-run if doubt.
- Adding new agent to install table → add detail block in `<details>` section below.
- Readability check before any README commit: would non-programmer understand + install within 60 seconds?

---

## Project overview

Iceage makes AI coding agents respond in compressed iceage-style prose — cuts ~65-75% output tokens, full technical accuracy. Ships as Claude Code plugin, Codex plugin, Gemini CLI extension, agent rule files for Cursor, Windsurf, Cline, Copilot, 40+ others via `npx skills`.

---

## File structure and what owns what

### Single source of truth files — edit only these

| File | What it controls |
|------|-----------------|
| `skills/iceage/SKILL.md` | Iceage behavior: intensity levels, rules, wenyan mode, auto-clarity, persistence. Only file to edit for behavior changes. |
| `rules/iceage-activate.md` | Always-on auto-activation rule body. CI injects into Cursor, Windsurf, Cline, Copilot rule files. Edit here, not agent-specific copies. |
| `skills/iceage-commit/SKILL.md` | Iceage commit message behavior. Fully independent skill. |
| `skills/iceage-review/SKILL.md` | Iceage code review behavior. Fully independent skill. |
| `skills/iceage-help/SKILL.md` | Quick-reference card. One-shot display, not a persistent mode. |
| `skills/iceage-batch/SKILL.md` | Batch compress behavior. Fully independent skill. |
| `skills/compress/SKILL.md` | Compress sub-skill behavior. |

### Auto-generated / auto-synced — do not edit directly

Overwritten by CI on push to main when sources change. Edits here lost.

| File | Synced from |
|------|-------------|
| `iceage/SKILL.md` | `skills/iceage/SKILL.md` |
| `plugins/iceage/skills/iceage/SKILL.md` | `skills/iceage/SKILL.md` |
| `.cursor/skills/iceage/SKILL.md` | `skills/iceage/SKILL.md` |
| `.windsurf/skills/iceage/SKILL.md` | `skills/iceage/SKILL.md` |
| `iceage.skill` | ZIP of `skills/iceage/` directory |
| `.clinerules/iceage.md` | `rules/iceage-activate.md` |
| `.github/copilot-instructions.md` | `rules/iceage-activate.md` |
| `.cursor/rules/iceage.mdc` | `rules/iceage-activate.md` + Cursor frontmatter |
| `.windsurf/rules/iceage.md` | `rules/iceage-activate.md` + Windsurf frontmatter |

---

## CI sync workflow

`.github/workflows/sync-skill.yml` triggers on main push when `skills/iceage/SKILL.md` or `rules/iceage-activate.md` changes.

What it does:
1. Copies `skills/iceage/SKILL.md` to all agent-specific SKILL.md locations
2. Rebuilds `iceage.skill` as a ZIP of `skills/iceage/`
3. Rebuilds all agent rule files from `rules/iceage-activate.md`, prepending agent-specific frontmatter (Cursor needs `alwaysApply: true`, Windsurf needs `trigger: always_on`)
4. Commits and pushes with `[skip ci]` to avoid loops

CI bot commits as `github-actions[bot]`. After PR merge, wait for workflow before declaring release complete.

---

## Hook system (Claude Code)

Three hooks in `hooks/` plus a `iceage-config.js` shared module and a `package.json` CommonJS marker. Communicate via flag file at `$CLAUDE_CONFIG_DIR/.iceage-active` (falls back to `~/.claude/.iceage-active`).

```
SessionStart hook ──writes "full"──▶ $CLAUDE_CONFIG_DIR/.iceage-active ◀──writes mode── UserPromptSubmit hook
                                                       │
                                                    reads
                                                       ▼
                                              iceage-statusline.sh
                                            [CAVEMAN] / [CAVEMAN:ULTRA] / ...
```

`hooks/package.json` pins the directory to `{"type": "commonjs"}` so the `.js` hooks resolve as CJS even when an ancestor `package.json` (e.g. `~/.claude/package.json` from another plugin) declares `"type": "module"`. Without this, `require()` blows up with `ReferenceError: require is not defined in ES module scope`.

All hooks honor `CLAUDE_CONFIG_DIR` for non-default Claude Code config locations.

### `hooks/iceage-config.js` — shared module

Exports:
- `getDefaultMode()` — resolves default mode from `CAVEMAN_DEFAULT_MODE` env var, then `$XDG_CONFIG_HOME/iceage/config.json` / `~/.config/iceage/config.json` / `%APPDATA%\iceage\config.json`, then `'full'`
- `safeWriteFlag(flagPath, content)` — symlink-safe flag write. Refuses if flag target or its immediate parent is a symlink. Opens with `O_NOFOLLOW` where supported. Atomic temp + rename. Creates with `0600`. Protects against local attackers replacing the predictable flag path with a symlink to clobber files writable by the user. Used by both write hooks. Silent-fails on all filesystem errors.

### `hooks/iceage-activate.js` — SessionStart hook

Runs once per Claude Code session start. Three things:
1. Writes the active mode to `$CLAUDE_CONFIG_DIR/.iceage-active` via `safeWriteFlag` (creates if missing)
2. Emits iceage ruleset as hidden stdout — Claude Code injects SessionStart hook stdout as system context, invisible to user
3. Checks `settings.json` for statusline config; if missing, appends nudge to offer setup on first interaction

Silent-fails on all filesystem errors — never blocks session start.

### `hooks/iceage-mode-tracker.js` — UserPromptSubmit hook

Reads JSON from stdin. Three responsibilities:

**1. Slash-command activation.** If prompt starts with `/iceage`, writes mode to flag file via `safeWriteFlag`:
- `/iceage` → configured default (see `iceage-config.js`, defaults to `full`)
- `/iceage lite` → `lite`
- `/iceage ultra` → `ultra`
- `/iceage wenyan` or `/iceage wenyan-full` → `wenyan`
- `/iceage wenyan-lite` → `wenyan-lite`
- `/iceage wenyan-ultra` → `wenyan-ultra`
- `/iceage-commit` → `commit`
- `/iceage-review` → `review`
- `/iceage-compress` → `compress`

**2. Natural-language activation/deactivation.** Matches phrases like "activate iceage", "turn on iceage mode", "talk like iceage" and writes the configured default mode. Matches "stop iceage", "disable iceage", "normal mode", "deactivate iceage" etc. and deletes the flag file. README promises these triggers, the hook enforces them.

**3. Per-turn reinforcement.** When flag is set to a non-independent mode (i.e. not `commit`/`review`/`compress`), emits a small `hookSpecificOutput` JSON reminder so the model keeps iceage style after other plugins inject competing instructions mid-conversation. The full ruleset still comes from SessionStart — this is just an attention anchor.

### `hooks/iceage-statusline.sh` — Statusline badge

Reads flag file at `$CLAUDE_CONFIG_DIR/.iceage-active`. Outputs colored badge string for Claude Code statusline:
- `full` or empty → `[CAVEMAN]` (orange)
- anything else → `[CAVEMAN:<MODE_UPPERCASED>]` (orange)

Configured in `settings.json` under `statusLine.command`. PowerShell counterpart at `hooks/iceage-statusline.ps1` for Windows.

### Hook installation

**Plugin install** — hooks wired automatically by plugin system.

**Standalone install** — `hooks/install.sh` (macOS/Linux) or `hooks/install.ps1` (Windows) copies hook files into `~/.claude/hooks/` and patches `~/.claude/settings.json` to register SessionStart and UserPromptSubmit hooks plus statusline.

**Uninstall** — `hooks/uninstall.sh` / `hooks/uninstall.ps1` removes hook files and patches settings.json.

---

## Skill system

Skills = Markdown files with YAML frontmatter consumed by Claude Code's skill/plugin system and by `npx skills` for other agents.

### Intensity levels

Defined in `skills/iceage/SKILL.md`. Six levels: `lite`, `full` (default), `ultra`, `wenyan-lite`, `wenyan-full`, `wenyan-ultra`. Persists until changed or session ends.

### Auto-clarity rule

Iceage drops to normal prose for: security warnings, irreversible action confirmations, multi-step sequences where fragment ambiguity risks misread, user confused or repeating question. Resumes after. Defined in skill — preserve in any SKILL.md edit.

### iceage-compress

Sub-skill in `skills/compress/SKILL.md`. Takes file path, compresses prose to iceage style, writes to original path, saves backup at `<filename>.original.md`. Validates headings, code blocks, URLs, file paths, commands preserved. Retries up to 2 times on failure with targeted patches only. Requires Python 3.10+.

### iceage-commit / iceage-review

Independent skills in `skills/iceage-commit/SKILL.md` and `skills/iceage-review/SKILL.md`. Both have own `description` and `name` frontmatter so they load independently. iceage-commit: Conventional Commits, ≤50 char subject. iceage-review: one-line comments in `L<line>: <severity> <problem>. <fix>.` format.

---

## Agent distribution

How iceage reaches each agent type:

| Agent | Mechanism | Auto-activates? |
|-------|-----------|----------------|
| Claude Code | Plugin (hooks + skills) or standalone hooks | Yes — SessionStart hook injects rules |
| Codex | Plugin in `plugins/iceage/` plus repo `.codex/hooks.json` and `.codex/config.toml` | Yes on macOS/Linux — SessionStart hook |
| Gemini CLI | Extension with `GEMINI.md` context file | Yes — context file loads every session |
| Cursor | `.cursor/rules/iceage.mdc` with `alwaysApply: true` | Yes — always-on rule |
| Windsurf | `.windsurf/rules/iceage.md` with `trigger: always_on` | Yes — always-on rule |
| Cline | `.clinerules/iceage.md` (auto-discovered) | Yes — Cline injects all .clinerules files |
| Copilot | `.github/copilot-instructions.md` + `AGENTS.md` | Yes — repo-wide instructions |
| Others | `npx skills add pomagrenate/iceage` | No — user must say `/iceage` each session |

For agents without hook systems, minimal always-on snippet lives in README under "Want it always on?" — keep current with `rules/iceage-activate.md`.

---

## Evals

`evals/` has three-arm harness:
- `__baseline__` — no system prompt
- `__terse__` — `Answer concisely.`
- `<skill>` — `Answer concisely.\n\n{SKILL.md}`

Honest delta = **skill vs terse**, not skill vs baseline. Baseline comparison conflates skill with generic terseness — that cheating. Harness designed to prevent this.

`llm_run.py` calls `claude -p --system-prompt ...` per (prompt, arm), saves to `evals/snapshots/results.json`. `measure.py` reads snapshot offline with tiktoken (OpenAI BPE — approximates Claude tokenizer, ratios meaningful, absolute numbers approximate).

Add skill: drop `skills/<name>/SKILL.md`. Harness auto-discovers. Add prompt: append line to `evals/prompts/en.txt`.

Snapshots committed to git. CI reads without API calls. Only regenerate when SKILL.md or prompts change.

---

## Benchmarks

`benchmarks/` runs real prompts through Claude API (not Claude Code CLI), records raw token counts. Results committed as JSON in `benchmarks/results/`. Benchmark table in README generated from results — update when regenerating.

To reproduce: `uv run python benchmarks/run.py` (needs `ANTHROPIC_API_KEY` in `.env.local`).

---

## Key rules for agents working here

- Edit `skills/iceage/SKILL.md` for behavior changes. Never edit synced copies.
- Edit `rules/iceage-activate.md` for auto-activation rule changes. Never edit agent-specific rule copies.
- README most important file for user-facing impact. Optimize for non-technical readers. Preserve iceage voice.
- Benchmark and eval numbers must be real. Never fabricate or estimate.
- CI workflow commits back to main after merge. Account for when checking branch state.
- Hook files must silent-fail on all filesystem errors. Never let hook crash block session start.
- Any new flag file write must go through `safeWriteFlag()` in `iceage-config.js`. Direct `fs.writeFileSync` on predictable user-owned paths reopens the symlink-clobber attack surface.
- Hooks must respect `CLAUDE_CONFIG_DIR` env var, not hardcode `~/.claude`. Same for `install.sh` / `install.ps1` / statusline scripts.
