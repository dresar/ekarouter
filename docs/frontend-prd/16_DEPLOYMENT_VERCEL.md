# 16. Vercel Deployment & Environment Configuration

---

## 1. Environment Variable Architecture

The standalone frontend requires the following environment variables configured in Vercel:

| Variable Name | Environment | Example Value | Description |
| :--- | :--- | :--- | :--- |
| `NEXT_PUBLIC_API_BASE_URL` | Production | `https://api.ekarouter.io` | Live Go backend URL |
| `NEXT_PUBLIC_API_BASE_URL` | Preview | `https://staging-api.ekarouter.io` | Staging backend |
| `NEXT_PUBLIC_API_BASE_URL` | Development | `http://localhost:8080` | Local Go backend instance |
| `NEXT_PUBLIC_APP_NAME` | All | `EkaRouter Console` | Browser title brand |

*Rule: Never hardcode `http://localhost:8080` in frontend source code. Always reference the environment variable.*

---

## 2. Vercel Configuration (`vercel.json`)

To enable seamless client-side routing (Single Page Application fallback) and optional local reverse-proxy rewrites, include this `vercel.json`:

```json
{
  "rewrites": [
    {
      "source": "/api/backend/:path*",
      "destination": "https://api.ekarouter.io/api/:path*"
    },
    {
      "source": "/((?!api/|_next/|_static/|[\\w-]+\\.\\w+).*)",
      "destination": "/"
    }
  ]
}
```

---

## 3. Build & Production Verification

- **Build Command**: `npm run build` or `pnpm build`
- **Output Directory**: `.next` (Next.js) or `dist` (Vite)
- **Node Version**: Node 20.x or 22.x LTS
- **Build Invariants**:
  - `npm run lint` must pass with 0 errors.
  - TypeScript type check (`tsc --noEmit`) must pass with 0 errors.
  - No secret keys (`SECRET_KEY`, `ADMIN_PASSWORD`) should ever be prefixed with `NEXT_PUBLIC_` or bundled into client code.
