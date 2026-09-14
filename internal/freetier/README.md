# Free Tier Catalog (`internal/freetier`)

## 1. Purpose
The `freetier` package catalogs free, freemium, and developer-tier third-party API providers, tracking quotas, rate limits, and verification requirements.

## 2. Responsibilities
- Maintain curated definitions of free and trial tiers across developer, AI, and SaaS services.
- Seed the local database with verified free-tier limits, reset cycles, and authentication requirements.
- Distinguish between completely free, credit-card-required, and temporary trial tiers.

## 3. Public Interfaces
- `Catalog`: Registry of free-tier provider offerings and metadata.
- `SeedDefaultCatalog(ctx context.Context, db *sql.DB) error`: Populates database tables with default entries.
- `GetProviderTier(providerID string) (*TierInfo, bool)`: Resolves tier metadata for a provider.

## 4. Dependencies
- Standard library: `context`, `database/sql`, `time`.
- `github.com/dresar/ekarouter/internal/db`: SQLite database handle.

## 5. Security Considerations
- Zero sensitive data stored: Only public terms, limits, and reference URLs are tracked.
- Quota values are explicitly marked as `estimated` or `verified` to avoid false assumptions.

## 6. Testing Instructions
Run freetier tests:
```bash
go test -v ./internal/freetier/...
```

## 7. Extension Instructions
To add a new provider to the catalog, add an entry to the seed array in `seed.go` with official documentation references.
