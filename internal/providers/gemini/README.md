# Gemini Provider Subsystem

The `gemini` package provides adapters and execution logic for Google Gemini services across multiple transport protocols.

## Modules

- `adapter.go`: Core `Adapter` implementing `providers.Adapter`. Supports `gemini` (public REST), `gemini-cli` (CodeAssist CLI), and `antigravity` (Google Internal CloudCode IDE).
- `agy.go`: Antigravity IDE protocol wrapper. Handles internal CloudCode endpoints (`/v1internal:generateContent`), project ID wrapping, thought signature caching, tool name sanitization regex, and aspect ratio detection.
- `cli.go`: Gemini CLI / CodeAssist converter. Handles project envelope wrapping and companion endpoint headers.
- `parser.go`: Streaming and non-streaming response JSON parser for candidate parts, thoughts, finish reasons, and token usage metadata.
- `capabilities.go`: Model capability matrix and context limits for Gemini models (`gemini-2.5-pro`, `gemini-2.5-flash`, `gemini-2.0-flash`, etc.).
