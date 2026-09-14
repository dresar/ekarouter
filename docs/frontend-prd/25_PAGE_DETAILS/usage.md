# Page Specification: Usage & Telemetry (`/usage`)

## Purpose
Visual analytics of token consumption, request velocity, error distributions, and provider cost estimations.

## Route
`/usage`

## Access Requirement
Authenticated.

## User Goals
- View prompt tokens vs completion tokens processed over time.
- Identify the most heavily utilized AI models and accounts.
- Inspect request success-to-failure ratios.

## Page Layout
- **Header**: Title *"Usage & Token Telemetry"*, Time Range Selector (`Today`, `7 Days`, `30 Days`).
- **Summary Cards**:
  - Total Tokens (`850k prompt + 312k completion = 1.16M total`).
  - Total Requests (`1,420`).
  - Success Rate (`98.4%`).
  - Estimated Spend (Client-side calculated based on token counts with *"Estimate"* badge).
- **Charts**:
  - Daily Token Consumption Area Chart (Prompt vs Completion).
  - Provider Share Donut Chart (OpenAI, Anthropic, Gemini).
- **Recent Requests Table**:
  - Columns: Timestamp, Route, Provider, Model, Tokens, Status Code, Latency.

## Data Endpoints
- `GET /api/usage`, `GET /api/v1/usage/summary`, `GET /api/v1/usage/providers`.

## Required SVG Icons
- `BarChart3`, `PieChart`, `TrendingUp`, `Zap`, `Calendar`.
