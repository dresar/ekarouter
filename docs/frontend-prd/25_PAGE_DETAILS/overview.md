# Page Specification: Command Center (`/overview`)

## Purpose
Primary mission control dashboard displaying real-time AI gateway traffic, provider availability, active failover cooldowns, and system health telemetry.

## Route
`/overview` (Default authenticated landing page).

## Access Requirement
Authenticated (`role: admin`, `developer`, or `viewer`).

## User Goals
- Monitor overall gateway health and uptime at a glance.
- Detect if any AI provider is currently in cooldown due to rate limits or auth errors.
- Inspect aggregate token usage and request volumes.
- Quickly generate an ingress curl snippet to test gateway completions.

## Page Layout
- **Topbar**: System Health Indicator Pill (`Healthy` / `Degraded`), Active Ingress Port (`8080`), Theme Toggle, User Profile.
- **Top Metrics Ribbon**: 4 compact MetricCards:
  1. Active Providers (`N / Total`)
  2. Total Requests Today (`N` requests)
  3. Total Tokens Processed (`N` tokens)
  4. Active Cooldowns (`N` providers in backoff)
- **Main Grid**:
  - Left (60%): Active Failover Status & Provider Availability Matrix.
  - Right (40%): Recent Ingress Traffic Stream & Ingress Test Snippet.

## Data Sources & API Endpoints
- `GET /health`: Uptime and database connection status.
- `GET /api/usage`: Total requests, success/failure counts, token totals.
- `GET /api/providers`: Provider availability and enabled status.
- `GET /api/accounts`: Cooldown status and active target states.

## States
- **Loading**: Skeleton placeholder blocks with subtle shimmer.
- **Empty**: Renders *"No traffic recorded yet. Connect an AI client to start routing."*
- **Error**: Displays inline reconnection banner with auto-retry countdown.

## Required SVG Icons
- `Activity`, `Cpu`, `Server`, `AlertTriangle`, `CheckCircle2`, `Clock`, `ArrowUpRight`.
