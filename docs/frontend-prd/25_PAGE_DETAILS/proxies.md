# Page Specification: Outbound Proxies (`/proxies`)

## Purpose
Configure and monitor outbound proxy profiles (`http`, `socks5`, `relay`) used by EkaRouter to route requests to upstream providers through dedicated egress tunnels.

## Route
`/proxies`

## Access Requirement
Admin role.

## User Goals
- View configured outbound proxies with host, port, protocol, and status.
- Trigger an instant latency and connectivity test against any proxy.
- Add a new proxy profile via inline expandable card or dedicated page.
- Remove outdated proxy profiles.

## Page Layout
- **Header**: Title *"Outbound Proxy Profiles"*, Description *"Manage egress proxies and check network latency."*, Action `[+ Add Proxy]`.
- **Inline Create Form** (Hidden by default, slides open upon clicking `Add Proxy`):
  - Fields: Name, Scheme (`http`/`https`/`socks5`/`relay`), Host, Port, Username, Password.
- **Proxy Table**:
  - Columns: Name, Protocol Scheme, Host & Port, Username, Status, Latency Indicator, Actions (`Test Connection`, `Delete`).

## Live Connection Testing Flow
- Clicking `Test Connection` dispatches `POST /api/proxies/{id}/test`.
- The button enters a loading state with spinner.
- Upon response:
  - If `ok: true`: Displays green badge with latency (e.g. `Online (45ms)`).
  - If `ok: false`: Displays red badge with error tooltip (e.g. `Timeout (5000ms)`).

## Data Endpoints
- `GET /api/proxies`, `POST /api/proxies`, `DELETE /api/proxies/{id}`, `POST /api/proxies/{id}/test`.

## Required SVG Icons
- `Globe`, `Wifi`, `CheckCircle`, `XCircle`, `Plus`, `Trash2`, `RefreshCw`.
