# TokenSaver Package

## Purpose
Reduces prompt context token consumption through safe, deterministic heuristic transformations without altering semantic code structure or error messages.

## Files
- `tokensaver.go`: TokenSaver coordinator, Git diff hunk truncation, repeated lines deduplication, smart text boundary truncation.
- `tokensaver_test.go`: Unit tests for modes, diff compaction, log deduplication, and fail-open guarantees.

## Allowed Responsibilities
- Fast, CPU-bounded byte and string transformations.
- Preserving file names, line numbers, and error traces.
- Failing open: retaining original input if compaction fails or produces larger text.

## Forbidden Responsibilities
- No network requests or heavy external tokenizers in the core path.
