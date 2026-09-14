# 22. AI Coding Agent Behavioral Constraints & Directives

---

## 1. Operating Rules for AI Coding Agents

When an AI agent (Cursor, Copilot, Claude Code) implements the EkaRouter frontend, the following rules govern its autonomy:

1. **Verify Before Coding**: Always check `04_ENDPOINT_INVENTORY.md` before writing a data fetch hook. If an endpoint does not appear there, do not invent it.
2. **Preserve Semantic Token Names**: Never hardcode hex values (`#2563EB`) in inline styles. Always reference CSS variables (`var(--brand-primary)`) or configured Tailwind classes.
3. **Handle Edge Cases Early**: Write loading and empty states before styling the populated data state. Every table must handle the `items.length === 0` case gracefully.
4. **Never Bypass Authentication**: Do not create temporary hardcoded user sessions or mock bypasses that could accidentally leak into production deployments.
5. **Report Backend Blockers**: If an interaction requires backend data not currently supplied by an endpoint, do not fake the data on the client. Mark the component with a clear notification banner: `Requires Backend Support: [Feature Description]`.
