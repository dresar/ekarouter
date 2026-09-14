# Proxy Package

## Purpose
Manages reusable outbound HTTP client transports, proxy profile resolution, connection pooling, and strict SSRF destination validation.

## Files
- `proxy.go`: Profile data structures, shared `http.Transport` caching by profile key, custom dialing context with private network IP validation.
- `proxy_test.go`: Unit tests for URL generation, private IP detection, SSRF policy blocking/allowing, and transport reuse.

## Allowed Responsibilities
- Creating and caching long-lived HTTP transports for direct and proxied connections.
- Validating outbound network targets to protect against SSRF.

## Forbidden Responsibilities
- No model payload translation or provider-specific headers.
