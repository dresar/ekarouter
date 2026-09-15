package tokensaver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Mode string

const (
	ModeOff        Mode = "off"
	ModeSafe       Mode = "safe"
	ModeBalanced   Mode = "balanced"
	ModeAggressive Mode = "aggressive"
)

type TokenSaver struct {
	mode Mode
}

func New(mode string) *TokenSaver {
	m := Mode(strings.ToLower(strings.TrimSpace(mode)))
	switch m {
	case ModeOff, ModeBalanced, ModeAggressive:
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
	compacted, _ := ts.CompactWithMode(input, ts.mode)
	if len(compacted) >= len(input) {
		return input
	}
	return compacted
}

func (ts *TokenSaver) CompactWithMode(input string, mode Mode) (string, []string) {
	if mode == ModeOff || len(input) == 0 {
		return input, nil
	}

	if isErrorTrace(input) {
		return input, []string{"Fail-safe active: Error trace preserved untouched"}
	}

	result := input
	var applied []string

	if strings.Contains(result, "diff --git") {
		compacted := CompactGitDiff(result, 500)
		if len(compacted) < len(result) {
			diffSaved := len(result) - len(compacted)
			result = compacted
			applied = append(applied, fmt.Sprintf("Git diff hunks compressed (-%d bytes)", diffSaved))
		}
	}

	if strings.Contains(result, "\n") {
		compacted, count := CompactRepeatedLinesCount(result)
		if count > 0 && len(compacted) < len(result) {
			saved := len(result) - len(compacted)
			result = compacted
			applied = append(applied, fmt.Sprintf("Repeated lines collapsed: %d blocks (-%d bytes)", count, saved))
		}
	}

	if mode == ModeBalanced || mode == ModeAggressive {
		compacted, count := CompactBlankLines(result, mode == ModeAggressive)
		if count > 0 && len(compacted) < len(result) {
			saved := len(result) - len(compacted)
			result = compacted
			applied = append(applied, fmt.Sprintf("Excessive blank lines collapsed: %d lines (-%d bytes)", count, saved))
		}

		compacted, trimmedLines := TrimTrailingWhitespace(result)
		if trimmedLines > 0 && len(compacted) < len(result) {
			saved := len(result) - len(compacted)
			result = compacted
			applied = append(applied, fmt.Sprintf("Trailing whitespace trimmed across %d lines (-%d bytes)", trimmedLines, saved))
		}
	}

	if mode == ModeAggressive {
		compacted, jsonBlocks := CompactJSONStructures(result)
		if jsonBlocks > 0 && len(compacted) < len(result) {
			saved := len(result) - len(compacted)
			result = compacted
			applied = append(applied, fmt.Sprintf("JSON structures minified: %d blocks (-%d bytes)", jsonBlocks, saved))
		}
	}

	if mode == ModeBalanced && len(result) > 20000 {
		compacted := SmartTruncate(result, 15000)
		if len(compacted) < len(result) {
			saved := len(result) - len(compacted)
			result = compacted
			applied = append(applied, fmt.Sprintf("Smart boundary truncation: preserved head/tail (-%d bytes)", saved))
		}
	} else if mode == ModeAggressive && len(result) > 8000 {
		compacted := SmartTruncate(result, 6000)
		if len(compacted) < len(result) {
			saved := len(result) - len(compacted)
			result = compacted
			applied = append(applied, fmt.Sprintf("Aggressive boundary truncation: preserved head/tail (-%d bytes)", saved))
		}
	}

	if len(result) >= len(input) {
		if len(applied) == 0 {
			applied = append(applied, "No redundant patterns detected; input already optimal")
		}
		return input, applied
	}

	return result, applied
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
	res, _ := CompactRepeatedLinesCount(text)
	return res
}

func CompactRepeatedLinesCount(text string) (string, int) {
	lines := strings.Split(text, "\n")
	if len(lines) < 4 {
		return text, 0
	}

	var result []string
	var lastLine string
	var lastRaw string
	repeatCount := 0
	blocksCollapsed := 0

	flushRepeats := func() {
		if repeatCount >= 3 {
			result = append(result, fmt.Sprintf("  [... repeated %d times]", repeatCount))
			blocksCollapsed++
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

	return strings.Join(result, "\n"), blocksCollapsed
}

func CompactBlankLines(text string, aggressive bool) (string, int) {
	lines := strings.Split(text, "\n")
	var result []string
	consecutiveBlank := 0
	collapsedCount := 0

	maxAllowed := 1
	if !aggressive {
		maxAllowed = 2
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			consecutiveBlank++
			if consecutiveBlank <= maxAllowed {
				result = append(result, "")
			} else {
				collapsedCount++
			}
		} else {
			consecutiveBlank = 0
			result = append(result, line)
		}
	}

	if collapsedCount > 0 {
		return strings.Join(result, "\n"), collapsedCount
	}
	return text, 0
}

func TrimTrailingWhitespace(text string) (string, int) {
	lines := strings.Split(text, "\n")
	trimmedLines := 0
	changed := false
	for i, line := range lines {
		trimmed := strings.TrimRight(line, " \t\r")
		if len(trimmed) < len(line) {
			lines[i] = trimmed
			trimmedLines++
			changed = true
		}
	}
	if changed {
		return strings.Join(lines, "\n"), trimmedLines
	}
	return text, 0
}

func CompactJSONStructures(text string) (string, int) {
	trimmed := strings.TrimSpace(text)
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(trimmed)); err == nil && buf.Len() < len(trimmed) {
			return buf.String(), 1
		}
	}

	count := 0
	pattern := "```json"
	if !strings.Contains(text, pattern) {
		return text, 0
	}

	var sb strings.Builder
	remaining := text
	for {
		start := strings.Index(remaining, pattern)
		if start == -1 {
			sb.WriteString(remaining)
			break
		}
		sb.WriteString(remaining[:start+len(pattern)])
		remaining = remaining[start+len(pattern):]

		end := strings.Index(remaining, "```")
		if end == -1 {
			sb.WriteString(remaining)
			break
		}

		jsonBody := remaining[:end]
		trimmedJson := strings.TrimSpace(jsonBody)
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(trimmedJson)); err == nil && buf.Len() < len(trimmedJson) {
			sb.WriteString("\n" + buf.String() + "\n")
			count++
		} else {
			sb.WriteString(jsonBody)
		}

		sb.WriteString("```")
		remaining = remaining[end+3:]
	}

	if count > 0 {
		return sb.String(), count
	}
	return text, 0
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
