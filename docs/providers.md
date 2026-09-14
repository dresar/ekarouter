# Universal Provider Directory & Capabilities

EkaRouter categorizes external developer APIs into 11 distinct domains with explicit capability flags.

## Supported Provider Categories

### 1. Developer & Deployment (`developer`)
- **Cloudflare (`cloudflare`):** DNS zones, Worker deployments, Page rules, Edge caches.
- **GitHub (`github`):** Repositories, issues, Actions workflows, pull requests.
- **Vercel (`vercel`):** Deployments, project management, domains.
- **Supabase (`supabase`):** Database management, project status.
- **Neon (`neon`):** Serverless Postgres branching, compute scaling.

### 2. Email & Communication (`communication`)
- **Resend (`resend`):** Modern developer transactional emails.
- **SendGrid (`sendgrid`):** High-volume marketing and transactional mail.
- **Twilio (`twilio`):** Programmable SMS, MMS, and voice calls.
- **Discord (`discord`):** Bot messages, channel webhooks, community alerts.
- **Slack (`slack`):** Web API, chat.postMessage, team channels.

### 3. Monitoring & Observability (`monitoring`)
- **Sentry (`sentry`):** Real-time error tracking and issue reporting.
- **Better Stack (`betterstack`):** Uptime monitoring, heartbeat alerts, incident status.

### 4. Automation & Workflows (`automation`)
- **Generic Webhook (`webhook`):** Outbound event dispatch with signature verification.

### 5. Search, Scraping & Data (`scraping_and_data`)
- **Firecrawl (`firecrawl`):** Web scraping and clean Markdown extraction for AI.
- **Tavily (`tavily`):** AI-agent optimized web search engine.
- **SerpAPI (`serpapi`):** Structured Google and Bing search results.
- **Jina AI (`jina`):** Reader API converting URLs to Markdown.

### 6. Storage & Media (`storage`)
- **Cloudflare R2 (`cloudflare-r2`):** Zero-egress S3-compatible object storage.

### 7. Payments (`payments`)
- **Stripe (`stripe`):** Global card payments, subscriptions, and webhooks.
- **Midtrans (`midtrans`):** Southeast Asian payment gateway and QRIS.

### 8. Analytics (`analytics`)
- **PostHog (`posthog`):** Event ingestion, user cohorts, feature flags.
- **Umami (`umami`):** Privacy-compliant web analytics.

### 9. Maps & Location (`maps`)
- **Mapbox (`mapbox`):** Geocoding, vector tiles, static map generation.

### 10. Security & Threat Intel (`security`)
- **VirusTotal (`virustotal`):** IP, URL, and file malware analysis.
- **AbuseIPDB (`abuseipdb`):** Malicious IP address checking and reporting.
- **IPinfo (`ipinfo`):** IP geolocation and ASN data.

### 11. Custom & Generic (`custom`)
- **Generic REST (`generic-rest`):** User-defined HTTP API with SSRF guard and auth injection.
