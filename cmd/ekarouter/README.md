# EkaRouter CLI Command

## Purpose
The primary binary entry point for EkaRouter.

## Files
- `main.go`: CLI flag definitions (`-port`, `-host`, `-db`, `-migrations`, `-backup`, `-version`), signal handling (`SIGINT`, `SIGTERM`), and bootstrap execution.

## Allowed Responsibilities
- Command-line parsing, environment overrides, version display, and process lifecycle management.

## Forbidden Responsibilities
- No embedded business logic or database queries directly in `main.go`.
