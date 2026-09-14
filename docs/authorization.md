# Role-Based Access Control (RBAC) & Scopes

EkaRouter enforces fine-grained authorization across all platform management and tool execution APIs.

## User Roles

| Role | Scope | Permissions Included |
|---|---|---|
| `admin` | Global Superuser | All permissions (`*`), user management, secret decryption for test, system settings. |
| `developer` | Project & Tool Operator | `providers.read`, `credentials.read`, `credentials.create`, `credentials.update`, `credentials.test`, `tools.read`, `tools.execute`, `projects.read`, `usage.read`. |
| `viewer` | Read-Only Observability | `providers.read`, `tools.read`, `projects.read`, `usage.read`. |

## Granular Permissions
- `providers.read`: List and view metadata/docs of registered API providers.
- `providers.manage`: Register, update, and toggle custom providers.
- `credentials.read`: List masked credentials, health states, and usage stats.
- `credentials.create`: Add new third-party credentials to the encrypted vault.
- `credentials.update`: Modify priorities, tags, and notes on existing credentials.
- `credentials.rotate`: Perform key rotation with new secret values.
- `credentials.test`: Trigger real-time connectivity health checks against upstream APIs.
- `tools.read`: View tool definitions and schemas.
- `tools.execute`: Execute external API tools through the platform engine.
- `projects.manage`: Create and modify development projects and environments.
- `audit.read`: View audit log events.
- `system.manage`: Trigger system backups, migrations, and global configuration changes.
