package tokensaver

import (
	"strings"
	"testing"
)

func TestTokenSaverModes(t *testing.T) {
	tsOff := New("off")
	if tsOff.Mode() != ModeOff {
		t.Errorf("expected off mode, got %s", tsOff.Mode())
	}

	tsSafe := New("safe")
	if tsSafe.Mode() != ModeSafe {
		t.Errorf("expected safe mode, got %s", tsSafe.Mode())
	}

	tsBalanced := New("balanced")
	if tsBalanced.Mode() != ModeBalanced {
		t.Errorf("expected balanced mode, got %s", tsBalanced.Mode())
	}

	tsAggressive := New("aggressive")
	if tsAggressive.Mode() != ModeAggressive {
		t.Errorf("expected aggressive mode, got %s", tsAggressive.Mode())
	}

	tsDefault := New("unknown")
	if tsDefault.Mode() != ModeSafe {
		t.Errorf("expected safe mode as default, got %s", tsDefault.Mode())
	}
}

func TestCompactRepeatedLines(t *testing.T) {
	input := "log line 1\nerror reading config\nerror reading config\nerror reading config\nerror reading config\nerror reading config\nlog line 2"
	compacted := CompactRepeatedLines(input)

	if !strings.Contains(compacted, "repeated 4 times") {
		t.Fatalf("expected repeated count in output, got:\n%s", compacted)
	}
}

func TestCompactGitDiff(t *testing.T) {
	var diffBuilder strings.Builder
	diffBuilder.WriteString("diff --git a/pkg/service.go b/pkg/service.go\n")
	diffBuilder.WriteString("index 123..456 100644\n")
	diffBuilder.WriteString("--- a/pkg/service.go\n")
	diffBuilder.WriteString("+++ b/pkg/service.go\n")
	diffBuilder.WriteString("@@ -1,5 +1,150 @@\n")
	for i := 0; i < 150; i++ {
		diffBuilder.WriteString("+ added line here for testing diff truncation\n")
	}

	diff := diffBuilder.String()
	compacted := CompactGitDiff(diff, 500)

	if !strings.Contains(compacted, "truncated") {
		t.Fatalf("expected diff hunk truncation, got:\n%s", compacted)
	}
}

func TestFailOpenBehavior(t *testing.T) {
	ts := New("safe")
	small := "hello world short text"
	out := ts.Compact(small)
	if out != small {
		t.Fatalf("expected exact original text for small inputs, got %s", out)
	}
}

func TestSyntaxCharProtection(t *testing.T) {
	code := "func foo() {\n    if true {\n        x := 1\n        _ = x\n    }\n}\n"
	compacted := CompactRepeatedLines(code)
	if strings.Contains(compacted, "repeated") {
		t.Errorf("syntax braces should never be collapsed, got:\n%s", compacted)
	}
}

func TestErrorTracePreservation(t *testing.T) {
	ts := New("safe")
	errTrace := "panic: runtime error: invalid memory address or nil pointer dereference\n[signal SIGSEGV: segmentation violation]\ngoroutine 1 [running]:\nmain.main()\n\t/app/main.go:10 +0x20\nrepeated line\nrepeated line\nrepeated line\nrepeated line\nrepeated line"
	out := ts.Compact(errTrace)
	if out != errTrace {
		t.Errorf("expected error trace to be preserved 100%% untouched")
	}
}

func TestCompactJSONStructures(t *testing.T) {
	jsonInput := "{\n  \"name\": \"ekarouter\",\n  \"version\": \"2.0\",\n  \"enabled\": true\n}"
	compacted, count := CompactJSONStructures(jsonInput)
	if count != 1 {
		t.Errorf("expected 1 json structure minified, got %d", count)
	}
	if strings.Contains(compacted, "\n  \"name\"") {
		t.Errorf("expected formatted indentation to be stripped, got %s", compacted)
	}

	markdownJson := "Here is the response:\n```json\n{\n  \"status\": \"ok\",\n  \"code\": 200\n}\n```\nDone."
	compactedMd, mdCount := CompactJSONStructures(markdownJson)
	if mdCount != 1 {
		t.Errorf("expected 1 json markdown block minified, got %d", mdCount)
	}
	if strings.Contains(compactedMd, "  \"status\"") {
		t.Errorf("expected indented json inside markdown to be compacted, got %s", compactedMd)
	}
}

func TestCompactBlankLines(t *testing.T) {
	input := "line 1\n\n\n\n\n\nline 2"
	compacted, count := CompactBlankLines(input, false)
	if count == 0 {
		t.Errorf("expected blank lines to be collapsed")
	}
	if strings.Contains(compacted, "\n\n\n\n") {
		t.Errorf("expected no more than 2 consecutive blank lines in balanced mode")
	}

	compactedAggro, countAggro := CompactBlankLines(input, true)
	if countAggro == 0 {
		t.Errorf("expected aggressive blank lines to be collapsed")
	}
	if strings.Contains(compactedAggro, "\n\n\n") {
		t.Errorf("expected at most 1 blank line in aggressive mode")
	}
}

func TestTrimTrailingWhitespace(t *testing.T) {
	input := "line with space   \nline with tabs\t\t\nclean line"
	compacted, lines := TrimTrailingWhitespace(input)
	if lines != 2 {
		t.Errorf("expected 2 lines trimmed, got %d", lines)
	}
	if strings.Contains(compacted, "   \n") || strings.Contains(compacted, "\t\t\n") {
		t.Errorf("expected trailing whitespace removed, got %q", compacted)
	}
}

func TestCompactWithMode(t *testing.T) {
	ts := New("safe")
	logMsg := "2026-09-15 15:00:00 [ERROR] worker process failed with timeout connecting to remote cluster node\n"
	input := "header\n" + strings.Repeat(logMsg, 6) + "footer"
	out, transformations := ts.CompactWithMode(input, ModeSafe)
	if len(transformations) == 0 {
		t.Errorf("expected at least one transformation recorded")
	}
	if !strings.Contains(out, "repeated") {
		t.Errorf("expected repeated lines collapsed in output, got:\n%s", out)
	}
}
