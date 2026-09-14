# Page Specification: System Audit Logs (`/audit-logs`)

## Purpose
Inspection of immutable administrative audit records and system events.

## Route
`/audit-logs`

## Access Requirement
Admin role.

## User Goals
- Review administrative changes (who added a provider, deleted a route, rotated a key).
- Inspect request metadata and execution results.
- Filter audit trail by actor, action type, or date.

## Page Layout
- **Header**: Title *"System Audit Trail"*, Auto-Refresh Toggle (`10s` interval).
- **Filter Toolbar**: Action filter dropdown, Actor search input.
- **Audit Table**:
  - Columns: Timestamp (ISO format), Actor (`admin`, `user`), Action (`provider.create`, `route.update`), Resource Type, Resource ID, Result (`Success` / `Failure`), Details Button.
- **Side Drawer Inspector**:
  - Clicking "Details" opens a `440px` right-side drawer with formatted JSON metadata.

## Data Endpoints
- `GET /api/v1/audit-logs`.

## Required SVG Icons
- `FileText`, `Filter`, `RefreshCw`, `Eye`, `User`, `Calendar`.
