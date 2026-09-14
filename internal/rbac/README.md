# RBAC Package

## Purpose
The `rbac` package provides user management, PBKDF2-HMAC-SHA256 password hashing, role-based access control, project isolation, and developer client token generation (`eka_pat_*`).

## Responsibilities
- Secure password hashing using PBKDF2 with SHA-256 (100,000 rounds, 16-byte random salt, constant-time verification).
- User authentication and role-permission evaluation (`admin`, `developer`, `viewer`).
- Project and environment management.
- Scoped developer client access tokens (`eka_pat_*`) with SHA-256 hash persistence.

## Public Interfaces
- `HashPassword(password string) (string, error)`
- `CheckPassword(password, encoded string) bool`
- `NewService(db *sql.DB) *Service`
- `(*Service) CreateUser(ctx context.Context, email, password, displayName, role string) (*User, error)`
- `(*Service) AuthenticateUser(ctx context.Context, email, password string) (*User, error)`
- `(*Service) CheckPermission(role, requiredPermission string) bool`
- `(*Service) CreateProject(ctx context.Context, name, ownerID, env, desc string) (*Project, error)`
- `(*Service) CreateClientToken(ctx context.Context, name, userID, projectID, scopes string, expiresDays int) (string, *ClientToken, error)`
- `(*Service) VerifyClientToken(ctx context.Context, rawToken string) (*ClientToken, error)`

## Security Considerations
- Passwords are never stored in plaintext and never logged.
- Client tokens are hashed with SHA-256 before storage; the plaintext token is only returned once upon creation.
