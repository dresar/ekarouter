# 15. API Client Architecture & Integration Rules

---

## 1. Unified HTTP Client Configuration

The frontend must instantiate a singleton HTTP client wrapper (e.g., using `fetch` or `ky` or `axios`) configured with:

1. **Dynamic Base URL**: Read from `process.env.NEXT_PUBLIC_API_BASE_URL` or `import.meta.env.VITE_API_BASE_URL`.
2. **Default Headers**:
   ```http
   Content-Type: application/json
   Accept: application/json
   ```
3. **Authorization Interceptor**: Automatically attaches `Authorization: Bearer <token>` if a token is present in client state.
4. **Timeout Guarantee**: Default request timeout of 15 seconds (except for SSE chat streaming which has unlimited timeout).

---

## 2. Cross-Origin Resource Sharing (CORS) Protocol

Because the frontend is deployed to Vercel and the backend runs on a separate domain (or local machine):
- **Credentials Mode**: Set `credentials: 'include'` on all `fetch` calls to allow session cookies across origins.
- **Preflight Handling**: Ensure headers sent by the frontend do not trigger unexpected CORS rejection.
- **Fallback Proxying**: During development or preview deployments, use Next.js / Vite rewrites to proxy `/api/*` to the backend if CORS is restricted on upstream networks.

---

## 3. Streaming Chat Completions (`/v1/chat/completions`)

When consuming streaming AI chat in playground widgets:
1. Use standard `fetch` with `ReadableStream` reader or `@microsoft/fetch-event-source`.
2. Decode UTF-8 chunks line-by-line looking for prefix `data: `.
3. Check for completion sentinel `data: [DONE]`.
4. Parse JSON chunks containing `choices[0].delta.content` and `choices[0].delta.reasoning_content`.
