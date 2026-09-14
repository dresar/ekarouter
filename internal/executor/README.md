# Executor Package

## Purpose
The `executor` package manages generic tool definitions and safe parameter execution for third-party API operations.

## Responsibilities
- Tool definition registration with SQLite persistence.
- Safe template string interpolation `{{variable}}`.
- Strict header key and value sanitization preventing CRLF injection and HTTP response splitting.
- Tool invocation with timeout control, SSRF validation, credential injection, and execution logging in `tool_executions`.

## Public Interfaces
- `InterpolateString(template string, vars map[string]string) string`
- `SanitizeHeaderKey(key string) (string, error)`
- `SanitizeHeaderValue(value string) (string, error)`
- `NewExecutor(db *sql.DB, registry *platform.Registry, vaultStore *vault.Store) *Executor`
- `(*Executor) CreateTool(ctx context.Context, tool *ToolDefinition) error`
- `(*Executor) GetTool(ctx context.Context, id string) (*ToolDefinition, error)`
- `(*Executor) ListTools(ctx context.Context, providerID string) ([]*ToolDefinition, error)`
- `(*Executor) ExecuteTool(ctx context.Context, params *ExecutionParams) (*ExecutionResult, error)`

## Security Considerations
- Prohibits header values containing carriage return or newline characters.
- Validates all destination URLs against SSRF protection policies prior to HTTP dispatch.
