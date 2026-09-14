# 17. Accessibility (a11y) & Semantic Standards

---

## 1. Compliance Level: WCAG 2.1 AA

All frontend components must meet WCAG 2.1 Level AA accessibility standards.

### Invariant Rules
1. **Semantic HTML Elements**: Use `<main>`, `<nav>`, `<aside>`, `<header>`, `<section>`, `<table>`, `<button>` rather than generic `<div>` soup.
2. **Keyboard Focus Rings**: Never remove outline without providing an explicit focus state:
   ```css
   :focus-visible {
     outline: 2px solid var(--brand-primary);
     outline-offset: 2px;
   }
   ```
3. **Interactive Touch Targets**: Every button, link, and interactive icon must have a minimum bounding box of `36x36px` on desktop and `44x44px` on touch screens.
4. **Icon-Only Buttons**: Any button containing solely an SVG icon (e.g., copy button, delete button, theme switch) **must** provide an accessible label via `aria-label` or visually hidden screen reader text:
   ```html
   <button aria-label="Copy API Key to Clipboard">
     <svg aria-hidden="true">...</svg>
   </button>
   ```
5. **Form Label Association**: Every `<input>` must be explicitly associated with a `<label>` via `htmlFor` / `id` attributes.
6. **Live Regions for Async Telemetry**: Status changes (such as proxy health updates or failover alerts) should announce to screen readers using `aria-live="polite"`.
