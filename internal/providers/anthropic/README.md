# Anthropic Provider Subsystem

The `anthropic` package provides execution logic, thinking block translation, and protocol adapters for the Anthropic Claude Messages API (`/v1/messages`).

## Modules

- `adapter.go`: Core `Adapter` implementing `providers.Adapter`.
- `thinking.go`: Reasoning and thinking block configuration (`budget_tokens`), temperature enforcement, multi-turn system prompt concatenation, beta headers (`prompt-caching-2024-07-31`, `pdfs-2024-09-25`, `output-128k-2025-02-19`), and tool cloaking helpers (`_ide` suffix for anti-ban).
- `parser.go`: JSON response and SSE event parser handling `content_block_delta` (`text_delta` and `thinking_delta`), `message_delta`, and `message_stop`.
- `capabilities.go`: Model capability matrix and context limits for Anthropic Claude models (`claude-3-7-sonnet`, `claude-3-5-sonnet`, `claude-3-5-haiku`, etc.).
