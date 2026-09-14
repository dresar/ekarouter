# EkaRouter Frontend PRD & API Contract Suite

Master documentation directory for the standalone frontend implementation of **EkaRouter**. This suite serves as the complete technical blueprint for frontend developers and AI coding agents (such as Cursor) to build, test, and deploy the user interface completely decoupled from the backend.

---

## 1. Directory Purpose & Audience

- **Target Audience**: AI coding agents (Cursor, Claude Code, Copilot Workspace) and senior frontend engineers.
- **Primary Objective**: Provide an authoritative, verified, and unambiguous contract of the backend application programming interface, information architecture, design tokens, interaction models, and deployment constraints.
- **Implementation Scope**: Standalone frontend application targeting deployment on **Vercel** communicating with an independent EkaRouter Go backend instance.

---

## 2. Reading Sequence for AI Coding Agents

Before writing any frontend code, the AI agent must read the documentation in this exact order:

1. [`01_PROJECT_OVERVIEW.md`](./01_PROJECT_OVERVIEW.md): System classification, operational architecture, and mission.
2. [`02_BACKEND_AUDIT.md`](./02_BACKEND_AUDIT.md): Verified backend stack, configuration, runtime flags, and database tables.
3. [`03_FEATURE_INVENTORY.md`](./03_FEATURE_INVENTORY.md): Available features versus pending features (`PASS`, `NOT_IMPLEMENTED`).
4. [`04_ENDPOINT_INVENTORY.md`](./04_ENDPOINT_INVENTORY.md): Complete list of all routes, parameters, and verification status.
5. [`05_API_CONTRACT.md`](./05_API_CONTRACT.md): Master request and response payload schemas.
6. [`06_AUTHENTICATION_AND_SECURITY.md`](./06_AUTHENTICATION_AND_SECURITY.md): JWT tokens, PBKDF2 hashing, session life-cycles, and secret masking rules.
7. [`07_FRONTEND_SITEMAP.md`](./07_FRONTEND_SITEMAP.md): Complete route tree and navigation structure.
8. [`08_PAGE_SPECIFICATIONS.md`](./08_PAGE_SPECIFICATIONS.md) & [`25_PAGE_DETAILS/`](./25_PAGE_DETAILS/): Granular functional blueprints for every page.
9. [`10_COMPONENT_SYSTEM.md`](./10_COMPONENT_SYSTEM.md) & [`11_DESIGN_SYSTEM.md`](./11_DESIGN_SYSTEM.md): Layout structures, non-modal patterns, and ocean blue design tokens.
10. [`16_DEPLOYMENT_VERCEL.md`](./16_DEPLOYMENT_VERCEL.md): Environment variables, CORS setup, and proxy rewrites.
11. [`21_CURSOR_IMPLEMENTATION_PROMPT.md`](./21_CURSOR_IMPLEMENTATION_PROMPT.md): The executable prompt instructing the frontend build.

---

## 3. Core Constraints & Non-Negotiables

| Category | Inflexible Rule | Rationale |
| :--- | :--- | :--- |
| **No Hallucinated APIs** | Every endpoint, field, and query parameter must exist in the backend source code. | Prevent runtime 404/422 failures during frontend integration. |
| **No Modal Trap** | Modals are prohibited for creation, editing, and viewing details. | Complex developer dashboards degrade when forced into dialog overlays. Use dedicated pages, inline cards, and right-side drawers. |
| **No Pill / Giant Cards** | Border-radius must remain compact (`8px` to `12px` for cards; `5px` to `8px` for buttons). | Maintain high information density suitable for developer infrastructure tooling. |
| **No AI Slop Styling** | No neon gradients, no glow effects, no generic template heroes, no emoji icons. | Produce a professional enterprise control center. |
| **Pure SVG Icons** | Only use coherent stroke SVG icons with explicit `aria-hidden` or accessible labels. | Consistent iconography without font-icon mismatch. |
| **Secret Protection** | Secrets are never displayed unmasked by default (`sk-****abcd`). | Prevent credential leakage across screen shares and logs. |

---

## 4. Current System Status

- **Backend Runtime**: Go 1.24+ (100% pure Go, CGO-free, embedded SQLite with WAL).
- **Backend Port**: `8080` default (`http://127.0.0.1:8080`).
- **Management API Path**: `/api/*` and `/api/v1/*`.
- **Public AI Gateway Path**: `/v1/*` (OpenAI compatibility).
- **Test Suite Pass Rate**: 100% across all 30+ internal packages.
- **Frontend Status**: Green-field implementation. Zero legacy frontend code exists in this repository.
