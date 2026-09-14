# EkaRouter Health Probes & Monitoring API

## 1. Overview

EkaRouter provides multi-tier health and diagnostic probes designed for container orchestrators (Kubernetes, Docker Swarm, Nomad), reverse proxies (Nginx, Traefik, Cloudflare), and administrator dashboards.

The health system is split into two primary layers:
1. **Infrastructure Probes (Public / Unauthenticated)**: Lightweight endpoints for liveness, readiness, and load-balancer health checks.
2. **Platform Health API (Authenticated / Admin & Platform)**: Comprehensive status of upstream AI providers, credential vault keys, and rotation status.

---

## 2. Infrastructure Health Probes

### 2.1 System Liveness: `GET /live`
Validates that the Go HTTP server process is running and accepting incoming network connections.

- **URL**: `/live`
- **Method**: `GET`
- **Authentication**: None (Public)
- **Response Code**: `200 OK`
- **Response Headers**: `Content-Type: application/json`
- **Response Body**:
```json
{
  "status": "alive",
  "uptime_seconds": 86420,
  "timestamp": "2026-09-15T04:55:00Z"
}
```

### 2.2 System Readiness: `GET /ready`
Validates that core dependencies—primarily the SQLite database connection pool and filesystem storage—are healthy and capable of serving queries.

- **URL**: `/ready`
- **Method**: `GET`
- **Authentication**: None (Public)
- **Responses**:
  - `200 OK`: Database connected, read/write verification passed.
  - `503 Service Unavailable`: Database locked, migration failed, or filesystem inaccessible.
- **Response Body (200 OK)**:
```json
{
  "status": "ready",
  "database": "connected",
  "writable": true,
  "timestamp": "2026-09-15T04:55:00Z"
}
```
- **Response Body (503 Service Unavailable)**:
```json
{
  "status": "degraded",
  "database": "error: database table is locked",
  "writable": false,
  "timestamp": "2026-09-15T04:55:00Z"
}
```

### 2.3 General Health: `GET /health`
A unified status probe returning overall application health.

- **URL**: `/health`
- **Method**: `GET`
- **Authentication**: None (Public)
- **Response Body (200 OK)**:
```json
{
  "status": "ok",
  "app": "ekarouter",
  "timestamp": "2026-09-15T04:55:00Z"
}
```

---

## 3. Platform Health Probes (Authenticated)

### 3.1 Platform Health Summary: `GET /api/v1/health`
Retrieves aggregated health status across all registered providers and credentials.

- **URL**: `/api/v1/health`
- **Method**: `GET`
- **Authentication**: Session Cookie (`session_token`) or `Authorization: Bearer <token>`
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "total_providers": 27,
    "healthy_providers": 26,
    "degraded_providers": 1,
    "total_credentials": 48,
    "active_credentials": 45,
    "cooldown_credentials": 3,
    "last_check_at": "2026-09-15T04:54:12Z"
  }
}
```

### 3.2 Live Provider Health Check: `POST /api/v1/providers/{id}/health`
Sends a synthetic ping or lightweight query to the specific upstream provider API to verify end-to-end network reachability and authentication validity.

- **URL**: `/api/v1/providers/{id}/health`
- **Method**: `POST`
- **Parameters**: `id` (string, path parameter, e.g. `openai`, `anthropic`, `cloudflare`)
- **Authentication**: Admin Session / Platform Token
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": {
    "provider_id": "openai",
    "status": "healthy",
    "latency_ms": 142,
    "http_status": 200,
    "checked_at": "2026-09-15T04:55:01Z"
  }
}
```

### 3.3 Credential Health Check: `GET /api/v1/credentials/{id}/health`
Returns historical health logs, consecutive failure counts, and cooldown state for an individual credential.

- **URL**: `/api/v1/credentials/{id}/health`
- **Method**: `GET`
- **Parameters**: `id` (string, path parameter, e.g. `cred_01j8k9m0`)
- **Response Body (200 OK)**:
```json
{
  "success": true,
  "data": {
    "credential_id": "cred_01j8k9m0",
    "provider_id": "openai",
    "status": "active",
    "consecutive_failures": 0,
    "is_in_cooldown": false,
    "cooldown_expires_at": null,
    "last_successful_request": "2026-09-15T04:50:23Z",
    "last_error": null
  }
}
```

---

## 4. Kubernetes & Docker Orchestration Configurations

### Docker Compose Example
```yaml
services:
  ekarouter:
    image: ekarouter:latest
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/ready"]
      interval: 10s
      timeout: 3s
      retries: 3
      start_period: 5s
```

### Kubernetes Pod Spec Example
```yaml
livenessProbe:
  httpGet:
    path: /live
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 2
  periodSeconds: 5
```
