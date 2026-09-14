# Page Specification: Routing Combos & Fallback (`/routing`)

## Purpose
Configure virtual model routes (e.g. `gpt-4o`, `claude-code`, `code-master`), assigning multiple provider accounts with priority-based failover or round-robin balancing.

## Route
- List: `/routing`
- Create / Edit: `/routing/new` or `/routing/[id]`

## Access Requirement
Admin role.

## User Goals
- Inspect all defined routes and the count of assigned target accounts.
- View target sequence and priority tiers (Tier 1 Primary -> Tier 2 Fallback).
- Configure rotation strategies (`priority`, `round_robin`).
- Add or delete routing targets.

## Page Layout
- **Header**: Title *"Model Routing & Combos"*, Action `[+ New Route]` navigating to `/routing/new`.
- **Routes Grid**: Card matrix displaying each route:
  - Route Slug (e.g. `gpt-4o` in monospace bold).
  - Strategy Tag: `Priority Fallback` or `Round Robin`.
  - Target Pipeline: Visual mini-cards showing Provider -> Account Name -> Priority.
  - Active Switch & Delete Button.
- **Route Editor (`/routing/new` & `/routing/[id]`)**:
  - Route Name input.
  - Strategy selector (`Priority`, `Round Robin`).
  - Target Accounts Builder: Reorderable rows where user selects Provider, Account, Priority, and Weight.

## Data Endpoints
- `GET /api/routes`, `POST /api/routes`, `DELETE /api/routes/{id}`.
- `GET /api/providers`, `GET /api/accounts`.

## Required SVG Icons
- `GitFork`, `Shuffle`, `Layers`, `Plus`, `Trash2`, `ArrowRight`, `Sliders`.
