# 12. Dark & Light Theme Implementation Protocol

---

## 1. Theme Strategy & Contrast Standards

EkaRouter supports **Dark Mode**, **Light Mode**, and **System Preference**.

Theme switching is implemented via standard CSS custom properties attached to `html[data-theme="dark"]` and `html[data-theme="light"]`.

### Contrast Guardrails
- **WCAG AA Compliance**: All normal body text must maintain a minimum contrast ratio of `4.5:1` against its background surface.
- **Data Tables**: Table header text must have a minimum ratio of `4.5:1` in both themes.
- **Interactive Focus Rings**: `outline: 2px solid var(--brand-primary); outline-offset: 2px;` must be visible against both light and dark backgrounds.

---

## 2. Preventing Inversion Flaws

Common AI frontend generation traps to avoid:
1. **Never use pure black `#000000`**: Pure black destroys visual depth and makes dark border dividers invisible. Always use deep slate navy (`#0F172A`).
2. **Never invert status semantics**: Emerald green for success must remain green in light mode, not blue.
3. **No saturated backgrounds**: Background surfaces in light mode must remain calm (`#F8FAFC`), not vibrant blue or tinted purple.
4. **Chart Gridlines**: In dark mode, chart grid lines must use `rgba(255, 255, 255, 0.08)`. In light mode, chart grid lines must use `rgba(0, 0, 0, 0.06)`.

---

## 3. Client Storage & Hydration Script

To prevent Flash of Incorrect Theme (FOIT) on page load, the frontend must execute a zero-dependency hydration script in `<head>`:

```javascript
(function() {
  var stored = localStorage.getItem('ekarouter_theme');
  var theme = stored || (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
  document.documentElement.setAttribute('data-theme', theme);
})();
```
