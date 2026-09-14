# Endpoint Details: Settings, Backups & Maintenance

---

## 1. GET `/api/settings`

- **Purpose**: Retrieve system-wide configuration keys and runtime flags.
- **Access**: Admin.
- **Success Status**: `200 OK`
- **Response Schema**:
  ```json
  {
    "theme": "dark",
    "token_saver_mode": "safe",
    "backup_retention_days": "14"
  }
  ```

---

## 2. PUT `/api/settings`

- **Purpose**: Update a key-value setting.
- **Access**: Admin.
- **Request Body**:
  ```json
  {
    "key": "token_saver_mode",
    "value": "aggressive"
  }
  ```
- **Success Status**: `200 OK`
- **Response**: `{"status":"updated"}`

---

## 3. GET `/api/backup` & POST `/api/backup`

- **GET `/api/backup`**: Lists available SQLite backup files stored in `data/backups/`.
- **POST `/api/backup`**: Triggers atomic SQLite `VACUUM INTO` snapshot.
- **Success Status**: `201 Created`
- **Response**:
  ```json
  {
    "status": "created",
    "file": "backup_20260915_043000.db"
  }
  ```
