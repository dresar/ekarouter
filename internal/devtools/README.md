# DevTools Package (`internal/devtools`)

## 1. Purpose
The `devtools` package handles developer productivity workflows, including OpenAPI 3.x specification parsing, request templating, and schema-driven tool generation.

## 2. Responsibilities
- Parse OpenAPI 3.x documents (JSON and YAML).
- Extract servers, paths, operations, query parameters, request bodies, and authentication schemes.
- Convert OpenAPI operations into executable EkaRouter Tool Definitions.
- Validate and interpolate request templates with variable substitution.

## 3. Public Interfaces
- `ParseOpenAPI(content []byte) (*OpenAPISpec, error)`: Parses OpenAPI 3.x document content.
- `GenerateTools(spec *OpenAPISpec, providerID string) ([]ToolBlueprint, error)`: Converts operations into tool definitions.
- `InterpolateTemplate(tpl string, vars map[string]string) string`: Safely injects variables into templates.

## 4. Dependencies
- Standard library: `encoding/json`, `fmt`, `strings`.
- Pure-Go YAML parser (or JSON-compatible AST parser).

## 5. Security Considerations
- Remote OpenAPI document imports pass through strict SSRF validation (`platform.ValidateSSRF`).
- Imported tools are marked inactive by default to allow manual administrative review prior to live execution.

## 6. Testing Instructions
Run devtools tests:
```bash
go test -v ./internal/devtools/...
```

## 7. Extension Instructions
To support OpenAPI 3.1 or custom schema extensions, extend the AST structures in `openapi.go`.
