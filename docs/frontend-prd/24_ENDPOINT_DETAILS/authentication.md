# Endpoint Details: Authentication & Sessions

---

## 1. POST `/api/auth/login`

- **Purpose**: Authenticate administrative session using username and password.
- **Access**: Public.
- **Request Headers**: `Content-Type: application/json`
- **Request Body**:
  ```json
  {
    "username": "admin",
    "password": "admin12345"
  }
  ```
- **Validation**:
  - `username`: Required string.
  - `password`: Required string (validated against PBKDF2 hash in `rbac_users` or config default).
- **Success Status**: `200 OK`
- **Success Response**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "username": "admin",
      "role": "admin"
    }
  }
  ```
- **Side Effects**: Sets `httpOnly` cookie `ekarouter_session`.
- **Error Codes**:
  - `400 Bad Request`: `{"error":"invalid json"}`
  - `401 Unauthorized`: `{"error":"invalid credentials"}`

---

## 2. POST `/api/auth/logout`

- **Purpose**: Invalidate administrative session.
- **Access**: Authenticated.
- **Success Status**: `200 OK`
- **Success Response**:
  ```json
  {
    "status": "logged_out"
  }
  ```
- **Side Effects**: Clears `ekarouter_session` cookie.

---

## 3. GET `/api/auth/me`

- **Purpose**: Fetch profile metadata for current authenticated session.
- **Access**: Authenticated (`Bearer <token>` or cookie).
- **Success Status**: `200 OK`
- **Success Response**:
  ```json
  {
    "authenticated": true,
    "user": {
      "username": "admin",
      "role": "admin"
    }
  }
  ```
- **Error Codes**:
  - `401 Unauthorized`: `{"error":"unauthorized"}`
