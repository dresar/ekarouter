# Providers Subsystem

## Architecture & Responsibilities

The `providers` subsystem connects the internal routing gateway with upstream AI model services. Each upstream ecosystem is decoupled into its own modular subpackage with dedicated protocol transformation, streaming parsers, and capability flags ("otak sendiri").

## Core Files

- `types.go`: Canonical request, response, streaming event, capability model, and adapter interfaces.
- `registry.go`: Thread-safe provider registry, HTTP error classification, and transient error heuristics.
- `providers_test.go`: Core registry and error classification test suite.

## Provider Subpackages

- `gemini/`: Google Gemini ecosystem handler.
  - `adapter.go`: Core adapter supporting standard Gemini REST, Gemini CLI, and Antigravity modes.
  - `agy.go`: Antigravity IDE protocol wrapper (`daily-cloudcode-pa.googleapis.com`), project envelope, tool sanitization regex, and aspect ratio detector.
  - `cli.go`: Gemini CLI / CodeAssist protocol converter and companion headers.
  - `parser.go`: Streaming and non-streaming response candidate parser.
  - `capabilities.go`: Model capability matrix and context window definitions.
- `openai/`: OpenAI ecosystem handler.
  - `adapter.go`: Core adapter for `/v1/chat/completions` and `/v1/responses` (Codex).
  - `reasoning.go`: Reasoning models (`o1`, `o3`, `o3-mini`, `deepseek-r1`), `system` to `developer` role transformation, and `reasoning_effort`.
  - `parser.go`: Response and SSE chunk parser for deltas, reasoning content, and token usage.
  - `capabilities.go`: OpenAI foundation model capability matrix.
- `anthropic/`: Anthropic Claude Messages API (`/v1/messages`) handler.
  - `adapter.go`: Core Claude Messages API adapter.
  - `thinking.go`: Extended thinking budget configuration, temperature enforcement, multi-turn system prompt concatenation, beta headers, and tool cloaking.
  - `parser.go`: SSE event parser for text deltas, thinking deltas, and usage.
  - `capabilities.go`: Anthropic Claude capability matrix.
- `custom/`: Self-hosted and alternative backends.
  - `adapter.go`: Flexible adapter dispatching requests with automatic base URL fallback.
  - `backends.go`: Registry and default base URLs for Ollama, vLLM, DeepSeek, LocalAI, Groq, OpenRouter, Mistral, and Together.
  - `capabilities.go`: Backend capability mappings.
