# 10. Reusable Component System

This document outlines the core reusable UI component library required for the EkaRouter console. Every component adheres to the compact, non-modal, restrained ocean blue design principles.

---

## 1. Layout Components

### `AppShell`
- **Purpose**: Main chrome wrapping all authenticated views.
- **Sub-components**: `Sidebar`, `Topbar`, `ContentArea`.
- **Props**: `user: UserProfile`, `health: SystemHealthState`.
- **Behavior**: Sidebar collapses to icons on screens below `1024px`; full toggleable mobile drawer below `768px`.

### `PageHeader`
- **Purpose**: Unified page title, breadcrumb hierarchy, and contextual actions.
- **Props**: `title: string`, `description?: string`, `breadcrumbs: BreadcrumbItem[]`, `actions?: ReactNode`.

### `SectionCard`
- **Purpose**: Compact container for grouped settings or metrics.
- **Styling**: `border-radius: 8px`, `border: 1px solid var(--border-color)`, `padding: 16px`.
- **Rules**: Zero glow, subtle elevation shadow, no oversized padding.

---

## 2. Data Display Components

### `DataTable`
- **Purpose**: Dense tabular presentation for accounts, proxies, logs, and routes.
- **Features**: Sortable headers, row selection, inline actions, pagination footer.
- **Row Density**: Height `40px` per row; sharp typography (`13px`).
- **States**: `isLoading` (skeleton rows), `isEmpty` (icon + description + action), `isError` (inline retry banner).

### `StatusBadge`
- **Purpose**: Visual indicator for system and account health.
- **Variants**:
  - `active` / `healthy` / `success`: Emerald (`#10B981` bg with `#064E3B` text / dark equivalent).
  - `cooling_down` / `warning`: Amber (`#F59E0B`).
  - `disabled` / `neutral`: Slate (`#64748B`).
  - `error` / `revoked`: Rose (`#EF4444`).
- **Shape**: Compact rectangle with `rounded-[4px]`, `px-2 py-0.5`, `text-[11px]`, font weight medium. Non-pill.

### `MetricCard`
- **Purpose**: Display top-level KPIs on the Command Center.
- **Layout**: Label (uppercase muted `11px`), Primary Value (`24px` bold mono/tabular), Secondary change indicator (`12px`).
- **Dimensions**: Compact height (`84px` to `96px`).

### `SecretViewer`
- **Purpose**: Display masked API keys with one-click copy.
- **Props**: `maskedValue: string`, `copyPayload?: string`.
- **Behavior**: Renders monospace masked string (`sk-****abcd`) accompanied by a stroke SVG copy icon button.

---

## 3. Form Controls

### `Button`
- **Purpose**: Primary interactive control.
- **Variants**: `primary` (ocean blue), `secondary` (slate surface), `danger` (rose outline/fill), `ghost`.
- **Sizes**:
  - `compact`: Height `32px`, font size `12px`, padding `0 12px`, radius `6px`.
  - `standard`: Height `36px`, font size `13px`, padding `0 16px`, radius `6px`.
- **States**: `default`, `hover`, `active`, `disabled`, `loading` (inline SVG spinner).

### `TextInput` & `Select`
- **Purpose**: Form data input.
- **Height**: `34px` (dense developer layout).
- **Styling**: Dark background (`#111827`), subtle border (`#334155`), focus ring (`#2563EB` 2px ring).

---

## 4. Non-Modal Interaction Components

### `InlineConfirm`
- **Purpose**: Safe inline confirmation for destructive actions (e.g. deleting an account or revoking a key).
- **Behavior**: Replaces the action button with "Confirm Delete?" and "Cancel" buttons within the table row itself, avoiding jarring modal popups.

### `RightDrawer`
- **Purpose**: Contextual inspector and form editor for complex sub-resources (e.g. rotating a secret or inspecting raw audit log JSON).
- **Width**: `440px` on desktop, `100%` on mobile. Smooth slide-in transition from right.
