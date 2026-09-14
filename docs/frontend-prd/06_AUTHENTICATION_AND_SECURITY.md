# 06. Authentication, Authorization & Secret Security

---

## 1. Authentication Mechanisms

EkaRouter supports three distinct authentication pathways:

### A. Management Session (Dashboard UI)
- **Endpoint**: `POST /api/auth/login`
- **Request Body**:
  ```json
  {
    "username": "admin",
    "password": "your-password"
  }
  ```
- **Response**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "username": "admin",
      "role": "admin"
    }
  }
  ```
- **Cookie Setup**: The backend sets an `httpOnly` cookie named `ekarouter_session`.
- **Header Fallback**: If cookies are blocked cross-origin, the client must transmit:
  ```http
  Authorization: Bearer eyJhbGciOi...
  ```

### B. Developer Platform API Access
- Authenticated via Developer Personal Access Tokens (`eka_pat_*`) passed via:
  ```http
  Authorization: Bearer eka_pat_abc123...
  ```
  or header:
  ```http
  X-API-Key: eka_pat_abc123...
  ```

### C. Ingress AI Gateway Authentication
- Public OpenAI clients authenticate against `/v1/chat/completions` using the standard OpenAI client bearer:
  ```http
  Authorization: Bearer eka_live_xyz789...
  ```

---

## 2. Session Management Rules for Frontend

1. **Token Persistence**: Store the JWT token in memory / session storage. On app mount, verify session health by calling `GET /api/auth/me`.
2. **Automatic 401 Interception**: When any API call returns HTTP `401 Unauthorized`, clear client auth state and redirect to `/login` with return URL.
3. **Logout Flow**: Invoke `POST /api/auth/logout`, clear local storage tokens, and invalidate SWR/React Query caches.

---

## 3. Secret Security & Masking Protocols

### Invariant Rules for Secrets
1. **Never Display Raw Secrets**: Plaintext API keys and secrets are never returned in list endpoints. They are always returned in masked form (e.g. `sk-****8f3a`).
2. **One-Time Plaintext Display**: When creating a new API key (`POST /api/keys`), the full key is returned exactly once in the response. The frontend must prompt the user with an explicit copy dialog before dismissing.
3. **No Secret in Client Logs**: Never print raw request bodies containing API keys to browser developer tools or error tracking systems.
4. **No Secrets in Query Parameters**: All secrets must travel exclusively in JSON request bodies over encrypted HTTPS connections.
5. **Memory Sanitization**: Input fields of type `password` or secret inputs must disable autocomplete (`autocomplete="new-password"`).
