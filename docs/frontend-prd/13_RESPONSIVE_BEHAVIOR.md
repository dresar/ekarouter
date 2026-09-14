# 13. Responsive Breakpoints & Multi-Device Layouts

---

## 1. Breakpoint Definitions

| Token | Min Width | Target Viewport | Navigation Adaptation | Table Behavior |
| :--- | :--- | :--- | :--- | :--- |
| `sm` | `640px` | Large Phones | Floating bottom bar or hamburger | Horizontal scroll container |
| `md` | `768px` | Tablets | Compact icon sidebar | Dense table with essential columns |
| `lg` | `1024px` | Laptops | Full 240px sidebar | Full columns with metadata |
| `xl` | `1280px` | Desktops | Fixed sidebar | Full data grids with right inspect panel |
| `2xl` | `1536px` | Ultra-Wide | Max container width `1440px` | Centered with balanced side margins |

---

## 2. Component Adaptations by Screen Size

### 1. Sidebar Navigation
- **`>= 1024px` (Desktop / Laptop)**: Permanent left column (`width: 240px`), grouped navigation items with titles, badge counts, and user profile pill at bottom.
- **`< 1024px` (Tablet / Mobile)**: Collapses into a backdrop drawer toggled via topbar hamburger icon (`aria-expanded`). Navigation links have a minimum touch target of `44x44px`.

### 2. Data Tables & Grids
- **Desktop**: Full data columns (e.g. Provider, Name, Model, Strategy, Status, Priority, Last Used, Actions).
- **Mobile**: Secondary metadata columns (`Last Used`, `Created At`, `Priority`) hide via responsive CSS utilities. The primary key and status badge remain sticky. An expandable arrow reveals secondary rows inline.

### 3. Metric Cards Ribbon
- **Desktop**: 4-column grid (`grid-template-columns: repeat(4, 1fr)`).
- **Tablet**: 2-column grid (`grid-template-columns: repeat(2, 1fr)`).
- **Mobile**: Horizontal swipeable scroll-snap container or 2x2 compact matrix.
