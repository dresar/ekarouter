# 11. Design System & Visual Language

---

## 1. Aesthetic Direction: Ocean Blue Developer Console

EkaRouter follows a restrained, dense, professional developer console design language inspired by modern infrastructure monitoring tools (Cloudflare, Supabase, Datadog). 

Core characteristics:
- **Primary Color**: Deep Ocean Blue (`#2563EB` / `#3B82F6`).
- **Surface Geometry**: Precise, compact rounded rectangles.
- **Card Radius**: `8px` to `12px` (strictly non-pill).
- **Control Radius**: `5px` to `8px`.
- **Information Density**: High density (compact padding, tabular numbers, clean data grids).
- **Prohibited Aesthetics**: No neon cyberpunk effects, no saturated rainbow gradients, no glassmorphism, no emoji icons, no giant floating cards.

---

## 2. Master Color Tokens

### Dark Theme Palette (Primary Console View)
```css
:root[data-theme="dark"] {
  /* Surface & Background */
  --bg-app: #0F172A;              /* Deep Slate Navy (Slate 900) */
  --bg-surface: #111827;          /* Card & Table Surface (Gray 900) */
  --bg-panel: #1E293B;            /* Sidebar & Elevated Panels (Slate 800) */
  --bg-input: #0B0F19;            /* Deep Inset Form Fields */
  --border-subtle: #1E293B;       /* Internal dividers */
  --border-strong: #334155;       /* Card & Container Borders (Slate 700) */

  /* Ocean Blue Brand Palette */
  --brand-primary: #2563EB;       /* Vibrant Ocean Blue (Blue 600) */
  --brand-hover: #1D4ED8;         /* Deep Ocean (Blue 700) */
  --brand-subtle: rgba(37, 99, 235, 0.12); /* Selected & Highlight Tints */
  --brand-text: #60A5FA;          /* Readable Blue Text on Dark (Blue 400) */

  /* Typography */
  --text-primary: #F8FAFC;        /* Crisp High-Contrast White (Slate 50) */
  --text-secondary: #CBD5E1;      /* Secondary Field Labels (Slate 300) */
  --text-muted: #64748B;          /* Metadata & Helpers (Slate 500) */

  /* Status Colors */
  --status-success: #10B981;      /* Emerald 500 */
  --status-warning: #F59E0B;      /* Amber 500 */
  --status-danger: #EF4444;       /* Rose 500 */
  --status-info: #0284C7;         /* Sky 600 */
}
```

### Light Theme Palette (Clean Neutral Contrast)
```css
:root[data-theme="light"] {
  /* Surface & Background */
  --bg-app: #F8FAFC;              /* Clean Slate Canvas (Slate 50) */
  --bg-surface: #FFFFFF;          /* Pure White Card Surface */
  --bg-panel: #F1F5F9;            /* Sidebar & Panels (Slate 100) */
  --bg-input: #FFFFFF;            /* Form Input Background */
  --border-subtle: #E2E8F0;       /* Light Dividers (Slate 200) */
  --border-strong: #CBD5E1;       /* Container Outlines (Slate 300) */

  /* Ocean Blue Brand Palette */
  --brand-primary: #2563EB;       /* Ocean Blue (Blue 600) */
  --brand-hover: #1D4ED8;         /* Darker Blue on Hover */
  --brand-subtle: #EFF6FF;        /* Light Blue Tint (Blue 50) */
  --brand-text: #1E40AF;          /* High-Contrast Blue (Blue 800) */

  /* Typography */
  --text-primary: #0F172A;        /* Near Black (Slate 900) */
  --text-secondary: #334155;      /* Slate 700 */
  --text-muted: #64748B;          /* Slate 500 */

  /* Status Colors */
  --status-success: #059669;      /* Emerald 600 */
  --status-warning: #D97706;      /* Amber 600 */
  --status-danger: #DC2626;       /* Red 600 */
  --status-info: #0284C7;         /* Sky 600 */
}
```

---

## 3. Typography Hierarchy

- **Font Family**: Modern clean sans-serif (Inter, system-ui, -apple-system, sans-serif).
- **Code & Numbers**: Tabular monospace (JetBrains Mono, SF Mono, Menlo, monospace).

| Element | Size | Weight | Line Height | Letter Spacing |
| :--- | :--- | :--- | :--- | :--- |
| **Page Title** | `20px` | `600` (Semi-bold) | `28px` | `-0.015em` |
| **Section Header** | `15px` | `600` (Semi-bold) | `22px` | `-0.01em` |
| **Table Head** | `11px` | `600` (Semi-bold) | `16px` | `+0.04em` (Uppercase) |
| **Body Regular** | `13px` | `400` (Regular) | `20px` | `0` |
| **Meta / Helpers** | `12px` | `400` (Regular) | `16px` | `0` |
| **Code / Keys** | `12px` | `500` (Medium Mono)| `18px` | `0` |
