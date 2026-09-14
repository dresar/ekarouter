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
