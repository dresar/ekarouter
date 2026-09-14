# Audit Report: Frontend-Backend Integration Gaps

---

## 1. Identified Integration Discrepancies

### A. Timestamp Formats
- **Observation**: SQLite stores timestamps via `CURRENT_TIMESTAMP` as strings (e.g. `"2026-09-15 04:30:00"`), while some Go struct fields output RFC3339 formatted strings (e.g. `"2026-09-15T04:30:00Z"`).
- **Frontend Mitigation**: Use standard date parsing (`new Date(str)`) or `date-fns` parseISO with fallback to handle space-separated SQLite dates.

### B. Boolean Storage in SQLite
- **Observation**: SQLite does not have a native boolean type. Boolean flags (`enabled`, `streaming`) are stored as integer `1` or `0`.
- **Backend Behavior**: The Go JSON encoders serialize these as JSON booleans (`true` / `false`).
- **Frontend Verification**: TypeScript schemas in `05_API_CONTRACT.md` correctly type these as `boolean`.

### C. CORS Credentials Header
- **Observation**: If the frontend on Vercel sends `credentials: 'include'`, the backend must respond with `Access-Control-Allow-Credentials: true` and cannot use wildcard `Access-Control-Allow-Origin: *`.
- **Backend Readiness**: Configured via `CORS_ALLOWED_ORIGINS` environment variable in `internal/config/config.go`.
