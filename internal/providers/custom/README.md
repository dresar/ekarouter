# Custom & Alternative Provider Subsystem

The `custom` package provides execution adapters for OpenAI-compatible self-hosted and alternative upstream AI backends.

## Supported Backends

- Generic OpenAI-compatible (`custom`)
- Ollama (`ollama`, default `http://localhost:11434/v1`)
- vLLM (`vllm`, default `http://localhost:8000/v1`)
- DeepSeek (`deepseek`, default `https://api.deepseek.com/v1`)
- LocalAI (`localai`, default `http://localhost:8080/v1`)
- Groq (`groq`, default `https://api.groq.com/openai/v1`)
- OpenRouter (`openrouter`, default `https://openrouter.ai/api/v1`)
- Mistral (`mistral`, default `https://api.mistral.ai/v1`)
- Together (`together`, default `https://api.together.xyz/v1`)

## Modules

- `adapter.go`: Core `Adapter` implementing `providers.Adapter`. Handles automatic fallback to default backend base URLs when credentials do not supply an explicit URL.
- `backends.go`: Backend registry constants, default base URL resolvers, and provider-specific header injection (e.g. OpenRouter attribution headers).
- `capabilities.go`: Capability matrix per backend kind (reasoning support, prompt caching, streaming).
