# EkaRouter Frontend Integration Guide

## 1. Overview & Architectural Principles

This guide defines the engineering conventions and communication standards for frontend applications (React, Next.js, TanStack Start, SvelteKit, or Vue) communicating with the EkaRouter backend.

### Golden Rules for Frontend Engineers:
1. **Never Assume Plaintext Secrets**: The backend vault never returns unmasked API keys or tokens. Any credential form should display the `masked_value` (e.g. `sk-proj-***4b`) in read-only mode. To replace an existing key, provide a dedicated "Rotate Secret" modal or toggle.
2. **Include Credentials Everywhere**: Administrative requests rely on the `session_token` HTTP-only cookie. Always pass `credentials: 'include'` in `fetch()` or `withCredentials: true` in Axios.
3. **No Polling for Non-Changing Data**: Use TanStack Query / SWR with optimistic UI updates and targeted query invalidation.

---

## 2. API Base URLs & Routing Prefix

| Route Group | Base Prefix | Intended Caller | Auth Mechanism |
|---|---|---|---|
| **Admin Console** | `/api/*` | Admin UI Dashboard | Cookie (`session_token`) |
| **Platform API** | `/api/v1/*` | Developer Portal & Console | Cookie or Bearer Token |
| **AI Gateway** | `/v1/*` | Client apps & Playground | Bearer API Key (`eka_live_*`) |
| **System Probes** | `/health`, `/ready`, `/live` | Orchestrator & Monitoring | Public |

---

## 3. Authentication & Session Management

### 3.1 Login Flow
```typescript
export async function login(username: string, password: string) {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
    credentials: 'include', // Necessary to receive session_token cookie
  });

  if (!res.ok) {
    const errorData = await res.json().catch(() => ({ error: 'Login failed' }));
    throw new Error(errorData.error || 'Invalid credentials');
  }

  return res.json();
}
```

### 3.2 Session Verification on Mount (`useAuth`)
```typescript
import { useQuery } from '@tanstack/react-query';

export function useAuth() {
  return useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      const res = await fetch('/api/auth/me', { credentials: 'include' });
      if (res.status === 401) return null;
      if (!res.ok) throw new Error('Failed to verify session');
      return res.json() as Promise<{ user: string }>;
    },
    retry: false,
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  });
}
```

### 3.3 Logout
```typescript
export async function logout() {
  await fetch('/api/auth/logout', {
    method: 'POST',
    credentials: 'include',
  });
  window.location.href = '/login';
}
```

---

## 4. Query & Mutation Best Practices (TanStack Query)

### Fetching Credentials List
```typescript
export function useCredentials(providerId?: string) {
  const url = providerId 
    ? `/api/v1/credentials?provider_id=${encodeURIComponent(providerId)}`
    : '/api/v1/credentials';

  return useQuery({
    queryKey: ['credentials', providerId ?? 'all'],
    queryFn: async () => {
      const res = await fetch(url, { credentials: 'include' });
      if (!res.ok) throw new Error('Could not load credentials');
      return res.json();
    },
  });
}
```

### Rotating a Credential with Optimistic Invalidation
```typescript
export function useRotateCredential() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ id, newSecret }: { id: string; newSecret: string }) => {
      const res = await fetch(`/api/v1/credentials/${id}/rotate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ new_secret: newSecret }),
        credentials: 'include',
      });
      if (!res.ok) throw new Error('Failed to rotate credential');
      return res.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['credentials'] });
      queryClient.invalidateQueries({ queryKey: ['audit-logs'] });
    },
  });
}
```

---

## 5. UI Error Handling & Toast Mapping

When handling API errors in UI forms and buttons:

```typescript
export function getFriendlyErrorMessage(err: any): string {
  if (err.response?.status === 403 && err.response?.data?.error?.includes('SSRF')) {
    return 'Destination blocked: Local or private network access is forbidden for security.';
  }
  if (err.response?.status === 429) {
    return 'Rate limit exceeded: Provider accounts are cooling down. Please retry shortly.';
  }
  if (err.response?.status === 401) {
    return 'Session expired. Please sign in again.';
  }
  return err.message || 'An unexpected error occurred. Please check console logs.';
}
```

---

## 6. Live SSE Streaming in Interactive AI Playgrounds

When integrating the chat completion playground:

```typescript
export async function streamChatCompletion({
  model,
  messages,
  apiKey,
  onChunk,
  onDone,
}: {
  model: string;
  messages: Array<{ role: string; content: string }>;
  apiKey: string;
  onChunk: (token: string) => void;
  onDone: () => void;
}) {
  const response = await fetch('/v1/chat/completions', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${apiKey}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ model, messages, stream: true }),
  });

  const reader = response.body?.getReader();
  if (!reader) throw new Error('Streaming not supported by browser');

  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop() || '';

    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed || trimmed === 'data: [DONE]') continue;
      if (trimmed.startsWith('data: ')) {
        try {
          const parsed = JSON.parse(trimmed.slice(6));
          const delta = parsed.choices?.[0]?.delta?.content;
          if (delta) onChunk(delta);
        } catch {
          // Ignore partial chunk parsing errors
        }
      }
    }
  }

  onDone();
}
```
