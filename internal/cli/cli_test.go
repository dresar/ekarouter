package cli

import (
	"strings"
	"testing"
)

func TestStripAnsiLength(t *testing.T) {
	colored := Styled("hello world", ColorCyan, Bold)
	length := stripAnsiLength(colored)
	if length != 11 {
		t.Errorf("expected 11, got %d", length)
	}
}

func TestBadge(t *testing.T) {
	b := Badge("ONLINE", ColorWhite, BgGreen)
	if !strings.Contains(b, "ONLINE") {
		t.Errorf("badge missing text: %s", b)
	}
}

func TestBox(t *testing.T) {
	lines := []string{"Line 1", "Line 2"}
	box := Box("TEST BOX", lines, 40, ColorCyan)
	if !strings.Contains(box, "TEST BOX") {
		t.Errorf("box missing title")
	}
	if !strings.Contains(box, "Line 1") {
		t.Errorf("box missing content")
	}
}

func TestToolConfigs(t *testing.T) {
	claude := BuildClaudeCodeConfig(8999, "test-key")
	if !strings.Contains(claude.Snippets["PowerShell"], "8999") {
		t.Errorf("expected port in claude config")
	}

	cursor := BuildCursorConfig(8999, "test-key")
	if !strings.Contains(cursor.Snippets["settings.json"], "MY-COMBO") {
		t.Errorf("expected MY-COMBO in cursor config")
	}

	cline := BuildClineConfig(8999, "test-key")
	if !strings.Contains(cline.Snippets["cline_custom_modes.json"], "openai-compatible") {
		t.Errorf("expected openai-compatible in cline config")
	}

	aider := BuildAiderConfig(8999, "test-key")
	if !strings.Contains(aider.Snippets["Terminal Command"], "aider") {
		t.Errorf("expected aider in command")
	}
}
