# SEO Boosting Plan — suluk.site

> Living document. Tracks the SEO strategy for **Suluk** (ERP/SaaS for Indonesian umrah/hajj travel agencies).
> Context: freshly migrated off the old `jamaah.web.id` domain (now retired); the app is a **hash-routed Vite + Svelte 5 SPA** behind a Cloudflare tunnel, with a handful of hand-built static HTML landing pages in `frontend-svelte/public/`.

---

## The headline problem

The app uses **hash-based routing** (`suluk.site/#/software-travel-umrah`, `#/fitur/crm-jamaah`, `#/tentang`, …). Search engines collapse every `#/…` fragment into the single root URL, so **11 marketing/guide pages of content are effectively invisible to crawlers**. Only the 4 hand-built static `.html` files are actually indexable today:

- `software-travel-umrah.html`
- `fitur-ocr-siskopatuh.html`
- `fitur-rooming-jamaah.html`
- `manifest-mutawwif-digital.html`

**Fixing crawlability is the single biggest lever.** Everything else amplifies it. Also: we're on a brand-new domain (≈zero authority), so authority must be rebuilt deliberately.

---

## Current state assessment

**Strengths (already in place)**
- Solid baseline `<meta>` + Open Graph + Twitter cards in `index.html`
- 4 JSON-LD blocks (SoftwareApplication, Organization, WebSite, FAQPage)
- 4 crawlable static HTML landing pages on real keywords
- `sitemap.xml` + `robots.txt` present
- Cloudflare in front (HTTPS, HTTP/2/3, Brotli, CDN edge cache)
- Good niche + locale targeting (`lang="id"`, umrah-travel-software keywords)
- Per-route dynamic meta via `upsertMeta()` in `App.svelte`

**Problems**
- Domain migration done with **no equity transfer** — `jamaah.web.id` now hard-404s
- **Hash routing** → marketing/guide content not indexable
- **Blanket root canonical** — every SPA route reports `canonical = suluk.site/`
- **Soft 404s** — `try_files … /index.html` returns HTTP 200 for any path
- **Content duplication** — static `/software-travel-umrah.html` vs in-app `#/software-travel-umrah` (`Pillar.svelte`)
- **No real OG image** — `og:image` points at the 512×512 `icon-512.png` (tiny social cards)
- **Stale sitemap** (`lastmod 2026-03`, only 5 URLs)
- **No origin compression / cache headers** in `nginx.conf` (partly masked by Cloudflare)
- **No analytics / Search Console** verification in place

---

## Phase 0 — Measurement & migration recovery *(do first; mostly owner actions)*

- [ ] **P0** Verify `suluk.site` in **Google Search Console** + **Bing Webmaster Tools**; submit `sitemap.xml`. *(Claude can add the verification `<meta>` to `index.html` given the token.)*
- [ ] **P0** Add analytics — GA4 or privacy-friendly Plausible/Umami. *(Claude can wire the snippet.)*
- [ ] **P0 — DECISION:** Recover `jamaah.web.id` equity or fresh start?
  - If it had real traffic/backlinks → add a **Cloudflare 301 Redirect Rule** `jamaah.web.id/* → suluk.site/*` (dashboard only; no need to revive the old stack), then file a **Change of Address** in GSC.
  - If negligible → treat `suluk.site` as a clean start.

> **UPDATE 2026-06-10 — done on branch `refactor/sveltekit`** (see memory [[sveltekit-migration]]): full SvelteKit migration shipped Phase 2 **(C)** plus most of Phase 1 & on-page Phase 3. The frontend is now SvelteKit with `adapter-static`; all marketing pages are **prerendered HTML** with per-page canonical/title, breadcrumb + FAQ + Service JSON-LD, and a real OG image. Not yet merged/deployed.

## Phase 1 — Technical SEO quick wins *(code; high impact, low effort)*

- [x] **P1** **Per-page canonical tags.** Done via `Seo.svelte` — each route emits its own `canonical` (e.g. `/tentang` → `suluk.site/tentang`) instead of always `suluk.site/`.
- [x] **P1** **Real 1200×630 OG/Twitter image** (`/og-suluk.png`, brand green + logo). Referenced by `Seo.svelte` (every page) + the homepage head.
- [x] **P1** **Regenerate `sitemap.xml` + `robots.txt`** — clean paths, `lastmod 2026-06-10`, app/admin/token paths disallowed.
- [x] **P1** **nginx compression + cache headers** — `gzip on`, immutable cache on `/_app/immutable/`, no-cache HTML, SPA fallback to `/200.html`.
- [ ] **P2** **Fix soft-404s** — still HTTP 200 via the `200.html` SPA fallback (`+error.svelte` shows 404 visually). True 404 status needs SSR (adapter-node) — deferred.
- [x] **P2** **Resolve duplication** — deleted the legacy `public/software-travel-umrah.html`; the prerendered `/software-travel-umrah` route now owns that URL.

## Phase 2 — Make content crawlable *(the big win — pick ONE approach)*

- [x] **DONE — approach (C): SvelteKit migration.** Marketing pages prerendered to static HTML via `adapter-static`; authed app/mobile/token routes stay client-SPA behind the fallback. Hash routes (`#/…`) replaced by clean History-API paths; a legacy-hash redirect shim preserves old links.
  - ~~(A) Expand the static-HTML pattern~~
  - ~~(B) Build-time prerender on the old SPA~~
  - **(C) Migrate marketing to SvelteKit (SSG)** ← shipped.

## Phase 3 — On-page & content strategy *(owner-led; Claude assists)*

- [ ] **Keyword → page map** (one cluster per page). Core clusters: `software travel umrah`, `aplikasi umrah siskopatuh`, `OCR KTP/KK umrah`, `rooming jamaah`, `manifest mutawwif`, `invoice/keuangan umrah`, `e-kontrak jamaah`.
- [x] **Unique title/H1/meta per page** — each prerendered route now has a unique `<title>` + meta description (no more blanket "… - Suluk").
- [x] **Pillar + cluster structure** — `/software-travel-umrah` = pillar; each `/fitur/*` = cluster; BreadcrumbList JSON-LD on all (Beranda → Software Travel Umrah → Feature) + interlinking via `MarketingShell` + Pillar related modules.
- [ ] **Content depth + freshness** — add `/blog` or `/panduan` (umrah-operations how-tos) for long-tail; publish on a cadence. *(Now easy: add prerendered routes under `(marketing)`.)*
- [x] **FAQ schema** — FAQPage JSON-LD on the pillar + all 5 feature pages, mirroring the visible `<details>` FAQ (`lib/seo/faq.js`).

## Phase 4 — Off-page & local *(owner-led, ongoing)*

- [ ] **Google Business Profile** + Indonesian SaaS/travel directories.
- [ ] **Backlinks** — guest posts in Indonesian travel-agency/umrah communities, listings, association partnerships.
- [ ] **Social signals** — consistent posting; verify OG cards render (depends on Phase 1 image).

---

## Open decisions

1. **Phase 0:** 301 `jamaah.web.id → suluk.site` to recover equity, or fresh start?
2. **Phase 2:** approach (A) static-HTML expansion *(recommended)*, (B) prerender + routing refactor, or (C) SvelteKit migration?
3. **Kickoff:** start with Phase 1 code quick wins now, then deploy?

## Key files

| Concern | Path |
| --- | --- |
| SPA shell + global SEO/JSON-LD | `frontend-svelte/index.html` |
| Per-route SEO config + routing | `frontend-svelte/src/App.svelte` (`PAGE_SEO`, `MARKETING_HASHES`) |
| Static landing pages | `frontend-svelte/public/*.html` |
| Landing page styles | `frontend-svelte/public/landing-guides.css` |
| Sitemap / robots | `frontend-svelte/public/sitemap.xml`, `robots.txt` |
| In-app marketing pages | `frontend-svelte/src/lib/pages/marketing/*.svelte` |
| Web server (compression/cache/404) | `nginx.conf` |
| Build config | `frontend-svelte/vite.config.js` |

## Deploy reminder

Frontend changes ship via: commit → `git pull` on VPS (`/data/docker/suluk`) → `docker compose -f deployments/docker-compose.yml build frontend && up -d frontend`. Public site = `https://suluk.site` (Cloudflare tunnel → `127.0.0.1:8007`).
