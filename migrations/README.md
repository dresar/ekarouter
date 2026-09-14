# Migrations Directory

## Purpose
This directory contains versioned SQL migrations for the SQLite database.

## Files
- `0001_initial.sql`: Sets up all core relational tables, indexes, and constraints for settings, providers, accounts, credentials, models, routes, route items, API keys, sessions, proxy profiles, usage logs, request logs, quota snapshots, and OAuth states.

## Rules
- Migrations are applied in sequential numeric order.
- Applied versions are tracked in the `schema_migrations` table.
- Modifying previously applied migrations is strictly forbidden; write a new numbered migration instead.
