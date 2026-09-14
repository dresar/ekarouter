# Page Specification: Interactive API Documentation (`/api-docs`)

## Purpose
In-console interactive API explorer allowing developers to view endpoint schemas, test gateway responses, and inspect curl code snippets directly in the web UI.

## Route
`/api-docs`

## Access Requirement
Authenticated.

## User Goals
- Browse all public AI gateway (`/v1/*`) and management endpoints (`/api/*`).
- View exact request body schemas and expected status codes.
- Copy generated curl snippets for external terminal testing.

## Page Layout
- **Split View**:
  - Left: Endpoint Navigation tree grouped by tag (Health, Ingress, Providers, Routes, Keys, Proxies, Vault, Tools).
  - Center: Parameter table, Request schema viewer, Response status codes.
  - Right: Interactive curl command generator with current ingress base URL.

## Data Endpoints
- Renders statically from `docs/frontend-prd/23_OPENAPI/openapi.json`.

## Required SVG Icons
- `BookOpen`, `Terminal`, `Copy`, `Check`, `Code`, `ExternalLink`.
