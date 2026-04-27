---
name: iceage-batch
description: >
  Batch compress all natural language files in a directory using iceage compression.
  Parallel workers, progress tracking, dry-run mode. Use when user says
  "compress all files", "batch compress", "compress this directory", or invokes /iceage-batch.
---

Compress all compressible files in a directory. Uses Claude API per file.

## Usage

```bash
go run ./iceage-batch/go <directory> [flags]
```

Flags:
- `--dry-run` — list files that would be compressed, no API calls
- `--no-backup` — skip creating `.original.md` backups (good for version-controlled dirs)
- `--workers N` — parallel compression (default 1, increase carefully — API rate limits apply)
- `--recursive=false` — top-level only, no subdirectories

## Examples

```bash
go run ./iceage-batch/go docs/
go run ./iceage-batch/go docs/ --dry-run
go run ./iceage-batch/go docs/ --no-backup
go run ./iceage-batch/go docs/ --workers 3
go run ./iceage-batch/go . --recursive=false
```

## Requirements

`ANTHROPIC_API_KEY` in environment or `.env.local`. Falls back to `claude` CLI if no key.

## Output

```
Found 8 compressible file(s) in docs/

[1/8] docs/guide.md ... OK (1.2s)
[2/8] docs/api.md ... OK (0.9s)
[3/8] docs/old.md ... Skipped (backup exists)
[4/8] docs/broken.md ... FAILED (validation failed after retries)

Done
  Compressed: 6
  Skipped:    1
  Failed:     1
```

Exit codes: `0` = all OK, `1` = fatal error, `2` = one or more files failed.
