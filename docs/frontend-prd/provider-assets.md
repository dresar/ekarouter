# EkaRouter Provider Brand Assets Documentation

**Generated**: 2026-09-15  
**Asset Directory**: `frontend/public/providers/`  
**Purpose**: High-fidelity vector logos for unified provider catalog and developer console.

---

## 1. Provider Logo Inventory

| Provider Name | File Name | Format | Source / Brand Reference | Dark / Light Adaptation | Usage Notes |
|---|---|---|---|---|---|
| **OpenAI** | `openai.svg` | SVG | Official OpenAI Brand Mark | Inverted with `currentColor` (off-white dark / near-black light) | Used for GPT-4o, o1, o3-mini |
| **Anthropic** | `anthropic.svg` | SVG | Official Anthropic Claude Mark | Inverted with `currentColor` (off-white dark / near-black light) | Used for Claude 3.5 Sonnet, Haiku, Opus |
| **Google Gemini** | `google.svg` | SVG | Google AI Official 4-Color Mark | Full-color on dark charcoal and light surfaces | Used for Gemini 1.5 Pro, Flash, 2.0 |
| **DeepSeek** | `deepseek.svg` | SVG | DeepSeek Official Brand Asset | Deep blue circular mark with white accent | Used for DeepSeek-V3, DeepSeek-R1 |
| **Alibaba Qwen** | `qwen.svg` | SVG | Alibaba Cloud / Tongyi Qwen Mark | Purple isometric node with light center | Used for Qwen 2.5, QwQ |
| **xAI Grok** | `xai.svg` | SVG | xAI Official Identity | Scalable vector mark (`currentColor`) | Used for Grok-2, Grok-Vision |
| **Moonshot Kimi** | `moonshot.svg` | SVG | Moonshot AI Official Logo | Dark sphere with luminous ocean-blue crescent | Used for Kimi Chat, Moonshot-v1 |
| **Groq** | `groq.svg` | SVG | Groq Official LPU Red Badge | High contrast red container with white typography | Used for LPU Ultra-fast inference |
| **Mistral AI** | `mistral.svg` | SVG | Mistral AI Official Cascading Steps | Orange pixel cascade on transparent background | Used for Mistral Large, Codestral, Pixtral |
| **Ollama** | `ollama.svg` | SVG | Ollama Official Llama Mark | Single-color vector with `currentColor` | Used for local instance proxy |
| **NVIDIA** | `nvidia.svg` | SVG | NVIDIA NIM Official Brand Green | NVIDIA Green (#76B900) rounded icon | Used for NVIDIA NIM inference |
| **Meta Llama** | `meta.svg` | SVG | Meta AI Infinity Loop | Meta Electric Blue (#0081FB) | Used for Llama 3.1, Llama 3.3 |
| **Cohere** | `cohere.svg` | SVG | Cohere Coral Mark | Forest green (#39594C) and lime (#D2E823) | Used for Command R, Command R+ |
| **Cloudflare** | `cloudflare.svg` | SVG | Cloudflare Official Cloud Icon | Cloudflare Orange (#F38020) | Used for Cloudflare AI & Workers |
| **GitHub** | `github.svg` | SVG | GitHub Invertocat | `currentColor` (off-white dark / near-black light) | Used for GitHub Models / Azure AI |
| **Together AI** | `together.svg` | SVG | Together AI Overlapping Circles | Electric blue & sky blue circles | Used for Together Inference API |
| **OpenRouter** | `openrouter.svg` | SVG | OpenRouter Diamond Node | Indigo (#6366F1) card with white node | Used for OpenRouter aggregator |
| **MiniMax** | `minimax.svg` | SVG | MiniMax Audio & Text Mark | Coral red (#FF4E36) with white wave | Used for MiniMax-01, Hailuo |
| **Zhipu GLM** | `zhipu.svg` | SVG | Zhipu AI BigModel Mark | Royal blue (#2955E7) geometric symbol | Used for GLM-4, GLM-4-Plus |
| **StepFun** | `stepfun.svg` | SVG | StepFun AI Step Vector | Emerald green (#18B984) stairs symbol | Used for Step-2, Step-1V |
| **Xiaomi MiMo** | `mimo.svg` | SVG | Xiaomi MiMo AI Brand Mark | Xiaomi Orange (#FF6900) | Used for MiMo Free Tier & TokenPlan |
| **Cerebras** | `cerebras.svg` | SVG | Cerebras Wafer Scale Engine Mark | Dark charcoal disk with Cerebras orange ring | Used for Cerebras 1800 tok/s inference |
| **Hugging Face** | `huggingface.svg` | SVG | Hugging Face Hug Emoji Emblem | Canary Yellow (#FFD21E) vector face | Used for HF Inference Endpoints |
| **Supabase** | `supabase.svg` | SVG | Supabase Lightning Bolt | Emerald Green (#3ECF8E) | Used for Supabase DB & Storage |
| **Neon** | `neon.svg` | SVG | Neon Postgres Slash Mark | Neon Green (#00E599) and black stroke | Used for Neon Serverless Postgres |
| **Vercel** | `vercel.svg` | SVG | Vercel Triangle Glyph | Pure black / off-white via `currentColor` | Used for Vercel AI Gateway & Deployments |
| **Generic Fallback** | `generic.svg` | SVG | Standard Server/Node outline | Crisp 2px stroke outline | Used as graceful fallback for custom endpoints |

---

## 2. Integration & Rendering Rules

1. **Local-First Asset Delivery**: All logos are served locally from `/providers/{name}.svg` without external CDN latency or privacy leakage.
2. **Graceful Fallback Pipeline**:
   - Primary: `<img src="/providers/{id}.svg" />`
   - Secondary: Fallback to initials badge with deterministic hue (e.g. `OP` for OpenRouter, `CC` for CommandCode) if file is missing.
   - Tertiary: `generic.svg` outline.
3. **Strict Zero-Malware Guarantee**: All SVGs are sanitised pure vector paths without `<script>`, `foreignObject`, or remote external URL references.
