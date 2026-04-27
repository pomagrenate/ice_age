<p align="center">
  <img src="https://em-content.zobj.net/source/apple/391/rock_1faa8.png" width="80" />
</p>

<h1 align="center">iceage-compress</h1>

<p align="center">
  <strong>shrink memory file. save token every session.</strong>
</p>

---

A Claude Code skill that compresses your project memory files (`CLAUDE.md`, todos, preferences) into iceage format — so every session loads fewer tokens automatically.

Claude read `CLAUDE.md` on every session start. If file big, cost big. Iceage make file small. Cost go down forever.

## What It Do

```
/iceage:compress CLAUDE.md
```

```
CLAUDE.md          ← compressed (Claude reads this — fewer tokens every session)
CLAUDE.original.md ← human-readable backup (you edit this)
```

Original never lost. You can read and edit `.original.md`. Run skill again to re-compress after edits.

## Benchmarks

Real results on real project files:

| File | Original | Compressed | Saved |
|------|----------:|----------:|------:|
| `claude-md-preferences.md` | 706 | 285 | **59.6%** |
| `project-notes.md` | 1145 | 535 | **53.3%** |
| `claude-md-project.md` | 1122 | 636 | **43.3%** |
| `todo-list.md` | 627 | 388 | **38.1%** |
| `mixed-with-code.md` | 888 | 560 | **36.9%** |
| **Average** | **898** | **481** | **46%** |

All validations passed ✅ — headings, code blocks, URLs, file paths preserved exactly.

## Before / After

<table>
<tr>
<td width="50%">

### 📄 Original (706 tokens)

> "I strongly prefer TypeScript with strict mode enabled for all new code. Please don't use `any` type unless there's genuinely no way around it, and if you do, leave a comment explaining the reasoning. I find that taking the time to properly type things catches a lot of bugs before they ever make it to runtime."

</td>
<td width="50%">

### 🪨 Iceage (285 tokens)

> "Prefer TypeScript strict mode always. No `any` unless unavoidable — comment why if used. Proper types catch bugs early."

</td>
</tr>
</table>

**Same instructions. 60% fewer tokens. Every. Single. Session.**

## Security

`iceage-compress` uses subprocess and file I/O patterns that may trigger static analysis warnings. This is expected — see [SECURITY.md](./SECURITY.md) for a full explanation of what the skill does and does not do.

## Install

Compress is built in with the `iceage` plugin. Install `iceage` once, then use `/iceage:compress`.

If you need local files, the compress skill lives at:

```bash
skills/compress/
```
