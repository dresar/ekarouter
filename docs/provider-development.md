# Provider Development Guide

Adding a new third-party API provider to EkaRouter is straightforward using the `platform.ProviderAdapter` interface.

## ProviderAdapter Interface

```go
type ProviderAdapter interface {
    Metadata() ProviderMetadata
    ValidateCredential(ctx context.Context, secret string) (bool, string, error)
    HealthCheck(ctx context.Context, secret string) (HealthStatus, error)
    Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error)
}
```

## Step-by-Step Implementation

1. **Define Metadata:**
   Specify provider ID, display name, category, base URL, auth type, docs links, and capabilities:

```go
meta := platform.ProviderMetadata{
    ID:                  "my-service",
    Name:                "My Service API",
    Category:            platform.CategoryDeveloper,
    Description:         "Developer API for my service",
    BaseURL:             "https://api.myservice.com/v1",
    AuthType:            platform.AuthTypeBearerToken,
    RequiredCredentials: []string{"api_key"},
    SupportedOperations: []string{"get_status", "create_resource"},
    Capabilities:        []platform.Capability{
        platform.CapBearerAuth,
        platform.CapHealthCheck,
        platform.CapRequestProxy,
    },
    WebsiteURL:          "https://myservice.com",
    DocsURL:             "https://myservice.com/docs",
    FreeTierStatus:      "available",
    Enabled:             true,
}
```

2. **Instantiate Base Adapter:**
   Use `platform.NewBaseAdapter` or wrap it with custom logic:

```go
adapter := platform.NewBaseAdapter(meta, 30*time.Second)
```

3. **Register in the Registry:**
   In `internal/platform/catalog.go`:

```go
_ = r.Register(adapter)
```

4. **Verify with Tests:**
   Add unit tests in `internal/platform/platform_test.go` verifying registration, search, and mock execution.
