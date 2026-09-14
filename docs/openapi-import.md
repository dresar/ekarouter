# OpenAPI 3.x Import Engine

EkaRouter includes an integrated OpenAPI 3.x importer capable of parsing JSON specifications, extracting API paths, operations, parameters, and schemas, and generating corresponding tool definitions.

## Workflow

1. **Upload or Fetch:**
   Provide the OpenAPI specification JSON to `POST /api/devtools/import-openapi`:
   ```bash
   curl -X POST http://localhost:8080/api/devtools/import-openapi \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <session_token>" \
     -d '{"spec_json":"{...}"}'
   ```

2. **Analysis & Safety Review:**
   The parser categorizes endpoints into safe operations (GET) and potentially destructive operations (POST, PUT, DELETE).

3. **Confirmation & Activation:**
   Operators confirm the imported definitions before they become active in `POST /api/devtools/confirm-openapi`. Tools are never activated without explicit confirmation.
