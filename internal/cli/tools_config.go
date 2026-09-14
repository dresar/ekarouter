package cli

import (
	"database/sql"
	"fmt"
)

type ToolConfig struct {
	Name        string
	Description string
	Snippets    map[string]string
}

func GetDefaultGatewayKey(db *sql.DB) string {
	var prefix string
	err := db.QueryRow("SELECT prefix FROM api_keys WHERE enabled = 1 LIMIT 1").Scan(&prefix)
	if err == nil {
		if prefix == "sk-7016dd" || prefix == "sk-7016dd1" || prefix == "sk-7016dd12" || prefix == "sk-7016dd1291" {
			return "sk-7016dd129191d903-u8emvi-c0b721b9"
		}
		if len(prefix) > 20 {
			return prefix
		}
	}
	return "sk-7016dd129191d903-u8emvi-c0b721b9"
}

func BuildClaudeCodeConfig(port int, apiKey string) ToolConfig {
	psScript := fmt.Sprintf(`$env:ANTHROPIC_BASE_URL="http://127.0.0.1:%d"
$env:ANTHROPIC_AUTH_TOKEN="%s"
$env:ANTHROPIC_DEFAULT_SONNET_MODEL="MY-COMBO"
$env:ANTHROPIC_DEFAULT_OPUS_MODEL="MY-COMBO"
$env:ANTHROPIC_DEFAULT_HAIKU_MODEL="cli-combo"
$env:API_TIMEOUT_MS="600000"`, port, apiKey)

	bashScript := fmt.Sprintf(`export ANTHROPIC_BASE_URL="http://127.0.0.1:%d"
export ANTHROPIC_AUTH_TOKEN="%s"
export ANTHROPIC_DEFAULT_SONNET_MODEL="MY-COMBO"
export ANTHROPIC_DEFAULT_OPUS_MODEL="MY-COMBO"
export ANTHROPIC_DEFAULT_HAIKU_MODEL="cli-combo"
export API_TIMEOUT_MS=600000`, port, apiKey)

	return ToolConfig{
		Name:        "Claude Code CLI",
		Description: "Configure Anthropic Claude Code CLI to route through EkaRouter combos",
		Snippets: map[string]string{
			"PowerShell": psScript,
			"Bash/Zsh":   bashScript,
		},
	}
}

func BuildCursorConfig(port int, apiKey string) ToolConfig {
	jsonSnippet := fmt.Sprintf(`{
  "models.openai.baseUrl": "http://127.0.0.1:%d/v1",
  "models.openai.apiKey": "%s",
  "models.openai.customModels": [
    { "name": "MY-COMBO" },
    { "name": "cli-combo" }
  ]
}`, port, apiKey)

	return ToolConfig{
		Name:        "Cursor IDE",
		Description: "Configure Cursor AI Settings to route models through EkaRouter",
		Snippets: map[string]string{
			"settings.json": jsonSnippet,
		},
	}
}

func BuildClineConfig(port int, apiKey string) ToolConfig {
	jsonSnippet := fmt.Sprintf(`{
  "apiProvider": "openai-compatible",
  "openAiBaseUrl": "http://127.0.0.1:%d/v1",
  "openAiApiKey": "%s",
  "openAiModelId": "MY-COMBO"
}`, port, apiKey)

	return ToolConfig{
		Name:        "Cline / Roo-Code",
		Description: "VS Code Cline extension configuration",
		Snippets: map[string]string{
			"cline_custom_modes.json": jsonSnippet,
		},
	}
}

func BuildAiderConfig(port int, apiKey string) ToolConfig {
	cmd := fmt.Sprintf("aider --openai-api-base http://127.0.0.1:%d/v1 --openai-api-key %s --model openai/MY-COMBO", port, apiKey)

	return ToolConfig{
		Name:        "Aider CLI",
		Description: "Command line arguments to invoke Aider with EkaRouter combo",
		Snippets: map[string]string{
			"Terminal Command": cmd,
		},
	}
}
