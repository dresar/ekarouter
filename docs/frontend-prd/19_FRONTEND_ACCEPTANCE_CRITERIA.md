# 19. Frontend Acceptance Criteria & Definition of Done

---

## 1. Definition of Done (DoD)

A page or component is considered complete only when all criteria below are satisfied:

### Functional Criteria
- [ ] Connects to actual backend endpoint without 404 or schema mismatch.
- [ ] Successfully performs CRUD operations with immediate state revalidation.
- [ ] Form validations highlight erroneous fields with clear error messages.
- [ ] Destructive actions utilize inline confirmation or safe confirmation flows.
- [ ] Sensitive secrets are never exposed in unmasked form by default.

### UX & Aesthetic Criteria
- [ ] Styled in restrained ocean blue theme (`#2563EB` brand, `#0F172A` dark canvas).
- [ ] Compact card corners (`8px` to `12px`); compact buttons (`5px` to `8px`, `32-38px` height).
- [ ] Zero emojis used as icons; only consistent stroke SVG icons.
- [ ] Zero unprompted modal traps; create/edit uses dedicated pages or inline panels.
- [ ] Responsive across mobile (375px), tablet (768px), and desktop (1280px+).
- [ ] Fully functional and readable in both Dark Mode and Light Mode.

### Code & Technical Criteria
- [ ] TypeScript passes with 0 type errors (`tsc --noEmit`).
- [ ] ESLint passes with 0 warnings or errors.
- [ ] Automated tests pass for components and critical user flows.
- [ ] Vercel preview build succeeds with clean production bundle.
