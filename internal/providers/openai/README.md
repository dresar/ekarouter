# OpenAI Provider Subsystem

The `openai` package provides execution logic and protocol transformations for OpenAI and OpenAI-compatible services.

## Modules

- `adapter.go`: Core `Adapter` implementing `providers.Adapter`. Supports `openai` (`/v1/chat/completions`) and `codex` (`/v1/responses`).
- `reasoning.go`: Reasoning model adaptation (`o1`, `o3`, `o3-mini`, `deepseek-r1`). Transforms `system` role to `developer` role, applies `reasoning_effort` ("low", "medium", "high"), maps `max_tokens` to `max_completion_tokens`, and strips server item identifiers (`rs_*`, `fc_*`, `resp_*`, `msg_*`).
- `parser.go`: JSON response and SSE chunk parser for deltas, reasoning content, finish reasons, and token usage metadata.
- `capabilities.go`: Model capability matrix and context limits for OpenAI foundation models (`gpt-4o`, `gpt-4o-mini`, `o1`, `o3-mini`, etc.).
