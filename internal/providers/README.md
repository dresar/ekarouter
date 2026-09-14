# Providers Package

## Purpose
Contains the abstract provider contract (`Adapter`), provider registry, concrete adapters (OpenAI, Anthropic, Gemini, Custom), and normalized error classification.

## Files
- `types.go`: Neutral `Request`, `Response`, `Message`, `ModelInfo`, `StreamEvent`, `Credentials`, and `Adapter` interface.
- `registry.go`: Thread-safe adapter registry, `ProviderError` classification, and transient error detection.
- `openai.go`: Adapter for OpenAI-compatible chat completions and SSE streaming.
- `anthropic.go`: Adapter translating neutral requests to Anthropic `/v1/messages` format and event streams.
- `gemini.go`: Adapter translating requests to Gemini `generateContent` and `streamGenerateContent`.
- `custom.go`: Configurable endpoint adapter for custom or self-hosted models.
- `providers_test.go`: Unit tests exercising mock server interactions, format conversions, streaming, and error classification.

## Allowed Responsibilities
- Protocol and format translation between neutral gateway structs and provider-specific wire APIs.
- Non-buffering SSE stream decoding.
- Error classification into transient vs client errors.

## Forbidden Responsibilities
- No database queries or global routing decisions.
- No direct storage or logging of plain API keys.
