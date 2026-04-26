---
name: iceage-export
description: >
  Exports the current chat session to a clean markdown file. Reconstructs the
  visible conversation from context — user messages, assistant responses, code
  blocks, tool output — formatted as readable .md. Use when user says
  "export chat", "save this conversation", "export to markdown", or invokes
  /iceage-export. Accepts optional filename argument.
---

Export current conversation to a markdown file. Reconstruct from visible context.

## Invocation

```
/iceage-export                        → chat-export-YYYY-MM-DD.md (current dir)
/iceage-export <filename>             → exact path/name given
/iceage-export <filename> --no-code   → strip code blocks from output
```

## What to Export

Reconstruct the full visible conversation in order:

- Every **user message** — exact wording, no paraphrasing
- Every **assistant response** — full content, preserve all code blocks, lists, tables
- **Tool calls** — include only if they produced user-visible output (e.g. file reads, search results). Skip internal plumbing (file existence checks, etc.)
- **System context** visible to user — skip hidden hook injections

## Output Format

```markdown
# Chat Export — YYYY-MM-DD HH:MM

**Session:** <working directory or project name if known>

---

## User

<exact user message text>

---

## Assistant

<full assistant response>

---

## User

...
```

Rules:
- Preserve all code blocks exactly — language tag, indentation, content unchanged
- Preserve all markdown (headings, bullets, tables, bold, links)
- Timestamp in filename uses local date: `chat-export-2026-04-26.md`
- If file already exists, append `_2`, `_3` etc — never overwrite silently
- Confirm filename to user after writing: `Exported: <path>`

## What NOT to Include

- Session hook injections (ICEAGE MODE ACTIVE banners, system reminders)
- Internal tool calls that weren't user-visible (stat checks, glob searches with no result shown)
- Your own internal reasoning or `<thinking>` blocks
- Duplicate content (don't include both a tool call and its already-rendered output)

## Boundaries

One-shot operation. Does not modify conversation. Does not affect iceage mode. "stop iceage-export": cancel and do nothing.

After writing the file, report: `Exported: <absolute path> (<N> exchanges)`
