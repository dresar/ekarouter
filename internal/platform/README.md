# Platform Package

## Purpose
The `platform` package manages the universal third-party provider integration framework, cataloging developer APIs across 11 distinct categories, discovering provider capabilities, verifying credentials, and executing external HTTP tools with strict SSRF protection.

## Responsibilities
- Universal provider registry for metadata, discovery, search, and capability declarations.
- Coverage across 11 categories: Developer, Communication, Monitoring, Automation, Scraping & Data, Storage, Payments, Analytics, Maps, Security, and Custom.
- Standardized `ProviderAdapter` contract for health checks, credential validation, and parameterized tool execution.
- Strict SSRF (Server-Side Request Forgery) protection rejecting loopback, link-local, RFC 1918 private ranges, and cloud metadata addresses.

## Public Interfaces
- `NewRegistry() *Registry`
- `RegisterDefaultProviders(r *Registry)`
- `(*Registry) Register(adapter ProviderAdapter) error`
- `(*Registry) Get(id string) (ProviderAdapter, bool)`
- `(*Registry) List() []ProviderMetadata`
- `(*Registry) ListByCategory(cat Category) []ProviderMetadata`
- `(*Registry) Search(query string) []ProviderMetadata`
- `ValidateSSRF(rawURL string) error`

## Security Considerations
- Prohibits arbitrary non-HTTP protocols (`file://`, `ftp://`, etc.).
- Rejects cloud metadata service endpoints (`169.254.169.254`).
- Sanitizes request headers and prevents raw API key leaks.
