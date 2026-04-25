---
name: iceage-help
description: >
  Quick-reference card for all iceage modes, skills, and commands.
  One-shot display, not a persistent mode. Trigger: /iceage-help,
  "iceage help", "what iceage commands", "how do I use iceage".
---

# Iceage Help

Display this reference card when invoked. One-shot — do NOT change mode, write flag files, or persist anything. Output in iceage style.

## Modes

| Mode | Trigger | What change |
|------|---------|-------------|
| **Lite** | `/iceage lite` | Drop filler. Keep sentence structure. |
| **Full** | `/iceage` | Drop articles, filler, pleasantries, hedging. Fragments OK. Default. |
| **Ultra** | `/iceage ultra` | Extreme compression. Bare fragments. Tables over prose. |
| **Wenyan-Lite** | `/iceage wenyan-lite` | Classical Chinese style, light compression. |
| **Wenyan-Full** | `/iceage wenyan` | Full 文言文. Maximum classical terseness. |
| **Wenyan-Ultra** | `/iceage wenyan-ultra` | Extreme. Ancient scholar on a budget. |

Mode stick until changed or session end.

## Skills

| Skill | Trigger | What it do |
|-------|---------|-----------|
| **iceage-commit** | `/iceage-commit` | Terse commit messages. Conventional Commits. ≤50 char subject. |
| **iceage-review** | `/iceage-review` | One-line PR comments: `L42: bug: user null. Add guard.` |
| **iceage-compress** | `/iceage:compress <file>` | Compress .md files to iceage prose. Saves ~46% input tokens. |
| **iceage-help** | `/iceage-help` | This card. |

## Deactivate

Say "stop iceage" or "normal mode". Resume anytime with `/iceage`.

## Configure Default Mode

Default mode = `full`. Change it:

**Environment variable** (highest priority):
```bash
export ICEAGE_DEFAULT_MODE=ultra
```

**Config file** (`~/.config/iceage/config.json`):
```json
{ "defaultMode": "lite" }
```

Set `"off"` to disable auto-activation on session start. User can still activate manually with `/iceage`.

Resolution: env var > config file > `full`.
