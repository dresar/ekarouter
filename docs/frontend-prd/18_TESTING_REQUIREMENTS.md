# 18. Frontend Quality Assurance & Testing Standards

---

## 1. Testing Pyramid for Frontend

The frontend implementation must include automated testing across three tiers:

### 1. Static Analysis
- **TypeScript**: Strict mode enabled (`"strict": true` in `tsconfig.json`). Zero `any` types on backend entity models.
- **ESLint**: Zero lint warnings or errors before git commits.

### 2. Component Unit Testing (Vitest / Jest + React Testing Library)
- Test every reusable component in [`10_COMPONENT_SYSTEM.md`](./10_COMPONENT_SYSTEM.md).
- Verify all 6 UI states (Loading, Empty, Error, Success, Disabled, Retrying).
- Test accessible labels on icon-only buttons.
- Test theme switching logic and custom property binding.

### 3. Integration & API Contract Testing (MSW - Mock Service Worker)
- Mock Service Worker handlers matching the schemas in [`05_API_CONTRACT.md`](./05_API_CONTRACT.md).
- Test authentication expiry and automatic redirect to `/login`.
- Test optimistic UI updates and rollback on network failure.
- Test rate-limit backoff banner rendering on HTTP 429.

### 4. End-to-End Smoke Testing (Playwright)
- Test Login -> Overview dashboard metrics load.
- Test Provider creation -> Account addition -> Route assignment.
- Test Ingress API key generation -> Copy to clipboard -> Revocation.
- Test Proxy connectivity test action.
