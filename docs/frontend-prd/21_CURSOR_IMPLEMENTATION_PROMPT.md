# 21. Master Cursor / AI Agent Implementation Prompt

> **Instructions for the User**: Copy and paste the entire prompt block below directly into Cursor, Claude Code, or your chosen AI frontend coding agent to build the EkaRouter frontend.

```markdown
# MISSION PROMPT: BUILD EKAROUTER FRONTEND CONSOLE

You are an expert Senior Frontend Architect and UI/UX Systems Engineer. Your mission is to build the complete, production-ready frontend for **EkaRouter**—an enterprise Universal AI Gateway and Developer API Platform.

You must build this frontend completely decoupled from the backend, designed for seamless deployment on Vercel, communicating with the live EkaRouter Go backend over HTTP/HTTPS.

---

## 1. MANDATORY READING SEQUENCE (DO THIS FIRST)

Before writing any code or generating components, you MUST read the comprehensive PRD specifications located in `docs/frontend-prd/` in this exact order:

1. `docs/frontend-prd/00_README.md`
2. `docs/frontend-prd/01_PROJECT_OVERVIEW.md`
3. `docs/frontend-prd/02_BACKEND_AUDIT.md`
4. `docs/frontend-prd/03_FEATURE_INVENTORY.md`
5. `docs/frontend-prd/04_ENDPOINT_INVENTORY.md`
6. `docs/frontend-prd/05_API_CONTRACT.md`
7. `docs/frontend-prd/06_AUTHENTICATION_AND_SECURITY.md`
8. `docs/frontend-prd/07_FRONTEND_SITEMAP.md`
9. `docs/frontend-prd/08_PAGE_SPECIFICATIONS.md`
10. `docs/frontend-prd/10_COMPONENT_SYSTEM.md`
11. `docs/frontend-prd/11_DESIGN_SYSTEM.md`
12. `docs/frontend-prd/12_DARK_LIGHT_THEME.md`
13. `docs/frontend-prd/16_DEPLOYMENT_VERCEL.md`

Do NOT guess endpoints, response fields, or UI patterns. Everything is strictly defined in these files.

---

## 2. NON-NEGOTIABLE CORE RULES

1. **Zero Hallucinated Endpoints**: Every API call you implement must match the verified endpoints in `docs/frontend-prd/04_ENDPOINT_INVENTORY.md`. If a feature is marked `NOT_IMPLEMENTED` (such as direct WebSocket log streaming or native Anthropic `/v1/messages`), do not invent fake client workarounds or mock backends.
2. **No Modal Trap**: Modals are PROHIBITED for primary CRUD workflows (creating providers, editing routes, inspecting credentials). Use dedicated full pages (e.g. `/providers/new`), expandable table rows, inline forms, or right-side slide-over drawers. Modals are permitted solely for irreversible destructive confirmations.
3. **No AI-Slop Styling**:
   - Prohibited: Neon cyberpunk accents, floating glassmorphism cards, oversaturated rainbow gradients, oversized rounded corners (`rounded-3xl`, `rounded-full` for cards), and generic template layouts.
   - Required: Dense, restrained, professional infrastructure console styling inspired by Cloudflare, Supabase, and Datadog.
4. **Design Palette**: Deep Ocean Blue (`#2563EB` brand primary, `#0F172A` dark background canvas, `#111827` surface, `#334155` borders).
5. **Card & Button Geometry**:
   - Card radius: Compact `8px` to `12px` with subtle border dividers.
   - Button radius: Compact `5px` to `8px`, height `32px` to `36px` (compact) or `36px` to `40px` (standard). Non-pill.
6. **Pure SVG Icons**: Never use emojis as icons. Never use inconsistent icon fonts. Use a single unified SVG stroke icon library (such as Lucide React or Tabler Icons) with explicit `aria-hidden="true"` or accessible labels.
7. **Secret Security**: Never render secrets in plaintext by default. Always display masked tokens (`sk-****abcd`) with one-click copy to clipboard. Never output secrets to browser console logs.
8. **Theme Support**: Flawless Dark Mode and Light Mode with high-contrast text meeting WCAG AA (`4.5:1` ratio).

---

## 3. RECOMMENDED FRONTEND TECH STACK

- **Framework**: Next.js (App Router) or Vite + React (SPA).
- **Language**: TypeScript (Strict mode enabled, zero `any` types).
- **Styling**: Tailwind CSS v4 / Vanilla CSS using the design tokens defined in `11_DESIGN_SYSTEM.md`.
- **Data Fetching & State**: TanStack Query (React Query) or SWR for automatic revalidation, caching, and retry logic.
- **Icons**: Lucide Icons (SVG stroke icons, 16px to 20px).
- **Deployment Target**: Vercel (read `16_DEPLOYMENT_VERCEL.md`).

---

## 4. PAGE IMPLEMENTATION SEQUENCE

Build the application incrementally in this exact order:

### Phase 1: Core Foundation & Shell
1. Setup dynamic API client (`NEXT_PUBLIC_API_BASE_URL` env variable).
2. Implement CSS design tokens for Dark and Light mode (`11_DESIGN_SYSTEM.md`).
3. Build `AppShell`, `Sidebar`, `Topbar`, and `StatusBadge` components.
4. Implement `/login` page with session state persistence and auth guard.

### Phase 2: AI Gateway Management
5. `/overview`: Command Center dashboard with active providers, token metrics, active failover status, and live health probes.
6. `/providers`: Provider catalog, account list, health checks, and `/providers/new` creation page.
7. `/providers/[id]`: Provider detail with multi-account priority list and credential inputs.
8. `/routing`: Route combos list with strategy indicators (`priority`, `round_robin`).
9. `/routing/new` & `/routing/[id]`: Route editor with target account sequencing and retries.
10. `/models`: Registered model aliases and capability inspection.
11. `/token-saver`: Interactive prompt compaction test playground.

### Phase 3: Developer Platform & Vault
12. `/vault`: Encrypted credential vault, masked secrets, environment tags, and rotation triggers.
13. `/vault/[id]`: Credential detail with health checks and secret rotation drawer.
14. `/tools`: Generic parameterized tool definitions and request templates.
15. `/tools/[id]`: Tool execution testing panel with variable interpolation.
16. `/proxies`: Outbound proxy profile manager with live connection latency tester.
17. `/free-tiers`: Community free tier catalog browser.

### Phase 4: Access, Observability & Settings
18. `/api-keys`: Ingress gateway client key management (one-time full key modal upon creation).
19. `/usage`: Aggregate token consumption charts and provider breakdown.
20. `/audit-logs`: Immutable system audit log table with raw JSON inspection drawer.
21. `/settings`: Runtime environment inspection and system settings editor.
22. `/backup`: Online SQLite backup manager with instant backup trigger.
23. `/api-docs`: Embedded interactive OpenAPI documentation viewer.

---

## 5. TESTING & VERIFICATION CHECKLIST

Before claiming completion:
- [ ] Run TypeScript type check (`tsc --noEmit`) and ensure 0 errors.
- [ ] Run linter and ensure 0 errors.
- [ ] Verify that every button triggers a real API call or defined navigation action. Zero inert/dummy buttons.
- [ ] Test Dark Mode and Light Mode across all 15+ pages.
- [ ] Test responsive viewports: Mobile (375px), Tablet (768px), and Desktop (1280px+).
- [ ] Test network failure handling: When backend is offline, topbar displays a persistent reconnection banner.
- [ ] Verify build passes (`npm run build`).

Deliver clean, concise, production-ready code with clean commit messages.
```
