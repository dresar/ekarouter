# Page Specification: System Settings & Backup (`/settings` & `/backup`)

## Purpose
Manage global runtime settings, inspect server environment variables, and manage SQLite database backup archives.

## Route
- Settings: `/settings`
- Backups: `/backup`

## Access Requirement
Admin role.

## User Goals
- View active system settings (token saver mode, retention periods).
- Trigger an instant online SQLite backup snapshot (`VACUUM INTO`).
- Download or inspect existing database backup archives.

## Page Layout
- **Tabs**: `General Settings`, `Database & Backup`, `System Information`.
- **General Settings Tab**:
  - Key-value form for system settings (`PUT /api/settings`).
- **Database & Backup Tab**:
  - Top Action: `[Create Backup Now]` button.
  - Backup Archives Table: File Name, Size, Created Timestamp.
- **System Information Tab**:
  - Read-only table showing Go version, runtime architecture, SQLite version, active port.

## Data Endpoints
- `GET /api/settings`, `PUT /api/settings`.
- `GET /api/backup`, `POST /api/backup`.
- `GET /health`, `GET /ready`.

## Required SVG Icons
- `Settings`, `Database`, `HardDrive`, `Download`, `Play`, `Check`.
