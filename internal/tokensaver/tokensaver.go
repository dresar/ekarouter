package tokensaver

import (
	"fmt"
	"strings"
)

type Mode string

const (
	ModeOff      Mode = "off"
	ModeSafe     Mode = "safe"
	ModeBalanced Mode = "balanced"
)

type TokenSaver struct {
	mode Mode
}

func New(mode string) *TokenSaver {
	m := Mode(strings.ToLower(strings.TrimSpace(mode)))
	switch m {
	case ModeOff, ModeBalanced:
	default:
		m = ModeSafe
	}
	return &TokenSaver{mode: m}
}

func (ts *TokenSaver) Mode() Mode {
	return ts.mode
}

func (ts *TokenSaver) Compact(input string) string {
	if ts.mode == ModeOff || len(input) < 100 {
		return input
	}

	if isErrorTrace(input) {
		return input
	}

	result := input

	if strings.Contains(input, "diff --git") {
		compacted := CompactGitDiff(input, 500)
		if len(compacted) < len(result) {
			result = compacted
		}
	}

	if strings.Contains(input, "\n") {
		compacted := CompactRepeatedLines(result)
		if len(compacted) < len(result) {
			result = compacted
		}
	}

	if ts.mode == ModeBalanced && len(result) > 20000 {
		compacted := SmartTruncate(result, 15000)
		if len(compacted) < len(result) {
			result = compacted
		}
	}

	if len(result) >= len(input) {
		return input
	}

	return result
}

func CompactGitDiff(diff string, maxLines int) string {
	lines := strings.Split(diff, "\n")
	var result []string
	currentFile := ""
	added := 0
	removed := 0
	inHunk := false
	hunkShown := 0
	hunkSkipped := 0
	const maxHunkLines = 100

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if hunkSkipped > 0 {
				result = append(result, fmt.Sprintf("  ... (%d lines truncated)", hunkSkipped))
				hunkSkipped = 0
			}
			if currentFile != "" && (added > 0 || removed > 0) {
				result = append(result, fmt.Sprintf("  +%d -%d", added, removed))
			}
			parts := strings.Split(line, " b/")
			if len(parts) > 1 {
				currentFile = strings.Join(parts[1:], " b/")
			} else {
				currentFile = "unknown"
			}
			result = append(result, "", currentFile)
			added = 0
			removed = 0
			inHunk = false
			hunkShown = 0
		} else if strings.HasPrefix(line, "@@") {
			if hunkSkipped > 0 {
				result = append(result, fmt.Sprintf("  ... (%d lines truncated)", hunkSkipped))
				hunkSkipped = 0
			}
			inHunk = true
			hunkShown = 0
			result = append(result, fmt.Sprintf("  %s", line))
		} else if inHunk {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				added++
				if hunkShown < maxHunkLines {
					result = append(result, fmt.Sprintf("  %s", line))
					hunkShown++
				} else {
					hunkSkipped++
				}
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				removed++
				if hunkShown < maxHunkLines {
					result = append(result, fmt.Sprintf("  %s", line))
					hunkShown++
				} else {
					hunkSkipped++
				}
			} else if hunkShown < maxHunkLines && !strings.HasPrefix(line, "\\") {
				if hunkShown > 0 {
					result = append(result, fmt.Sprintf("  %s", line))
					hunkShown++
				}
			}
		} else {
			result = append(result, line)
		}

		if len(result) >= maxLines {
			result = append(result, "... (more changes truncated)")
			break
		}
	}

	if hunkSkipped > 0 {
		result = append(result, fmt.Sprintf("  ... (%d lines truncated)", hunkSkipped))
	}
	if currentFile != "" && (added > 0 || removed > 0) {
		result = append(result, fmt.Sprintf("  +%d -%d", added, removed))
	}

	return strings.Join(result, "\n")
}

func CompactRepeatedLines(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) < 5 {
		return text
	}

	var result []string
	var lastLine string
	var lastRaw string
	repeatCount := 0

	flushRepeats := func() {
		if repeatCount >= 3 {
			result = append(result, fmt.Sprintf("  [... repeated %d times]", repeatCount))
		} else {
			for i := 0; i < repeatCount; i++ {
				result = append(result, lastRaw)
			}
		}
		repeatCount = 0
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !isSyntaxChar(trimmed) && trimmed == lastLine {
			repeatCount++
			continue
		}

		flushRepeats()
		result = append(result, line)
		lastLine = trimmed
		lastRaw = line
	}

	flushRepeats()

	return strings.Join(result, "\n")
}

func isSyntaxChar(s string) bool {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) <= 4 {
		switch trimmed {
		case "{", "}", "[", "]", "(", ")", ";", ",", `""`, "''", "`", "end", "fi", "done":
			return true
		}
	}
	return false
}

func isErrorTrace(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "panic:") ||
		strings.Contains(lower, "fatal error:") ||
		strings.Contains(lower, "traceback (most recent call last):") ||
		strings.Contains(lower, "stack trace:") ||
		strings.Contains(lower, "syntaxerror:")
}

func SmartTruncate(text string, maxChars int) string {
	if len(text) <= maxChars {
		return text
	}
	half := maxChars / 2
	head := text[:half]
	tail := text[len(text)-half:]
	return head + "\n\n... [output truncated for brevity] ...\n\n" + tail
}
