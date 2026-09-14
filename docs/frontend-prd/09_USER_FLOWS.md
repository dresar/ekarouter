# 09. Core User Flows & Interaction Models

---

## 1. Flow A: Adding an AI Provider & Target Account

```text
[ /providers ]
      │
      ▼ (Click "Add Provider")
[ /providers/new ] (Dedicated full page form)
      │
      ▼ (Submit Provider Metadata)
[ POST /api/providers ] ──► (Success 201)
      │
      ▼ (Redirect to Provider Detail)
[ /providers/[id] ]
      │
      ▼ (Expand "Add Account" Inline Panel)
[ Input: Account Name, Priority, API Key ]
      │
      ▼ (Submit Account Form)
[ POST /api/accounts ] ──► (AES-256-GCM Encrypted & Stored)
      │
      ▼
[ Account Appears in Accounts List with 'Active' Status Badge ]
```

---

## 2. Flow B: Creating a Model Routing Combo with Failover

```text
[ /routing ]
      │
      ▼ (Click "New Route")
[ /routing/new ]
      │
      ├── 1. Enter Route Name: "gpt-4o"
      ├── 2. Select Strategy: "Round Robin" or "Priority Fallback"
      └── 3. Add Targets:
             - Target 1: OpenAI Primary Account (Priority 1)
             - Target 2: OpenAI Backup Account  (Priority 1)
             - Target 3: OpenRouter Fallback    (Priority 2)
      │
      ▼ (Submit Route Form)
[ POST /api/routes ] ──► (Success 201)
      │
      ▼
[ Redirects to /routing/[id] with visual diagram of target priorities ]
```

---

## 3. Flow C: Generating Ingress API Key for Client Application

```text
[ /api-keys ]
      │
      ▼ (Click "Generate Key")
[ Inline Header Form Expands: Name + Scopes Input ]
      │
      ▼ (Submit)
[ POST /api/keys ] ──► (Success 201)
      │
      ▼
[ Persistent Alert Banner Displays Full Secret Key Exactly Once ]
      │
      ├── User clicks "Copy to Clipboard" button
      └── User clicks "I have stored this key safely" to dismiss banner
```

---

## 4. Flow D: Testing Outbound Proxy Connectivity

```text
[ /proxies ]
      │
      ▼ (Click "Test Connection" on Proxy Row)
[ POST /api/proxies/[id]/test ]
      │
      ├── Button enters Loading State (spinner, text: "Testing...")
      │
      ▼
[ Response Received: ok=true, status=200, latency=42ms ]
      │
      ▼
[ Inline Badge flashes Emerald Green: "Online (42ms)" ]
```

---

## 5. Flow E: Credential Rotation in Vault

```text
[ /vault/[id] ]
      │
      ▼ (Click "Rotate Secret" in Header Actions)
[ Right-side Drawer opens with 'New Secret Value' input ]
      │
      ▼ (Enter new API Key & Submit)
[ POST /api/v1/credentials/[id]/rotate ] ──► (Success 200)
      │
      ▼
[ Drawer closes; Audit entry appended; Status badge updates to 'Active' ]
```
