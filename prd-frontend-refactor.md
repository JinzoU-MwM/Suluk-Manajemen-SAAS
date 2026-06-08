# PRD — Frontend Refactor Jamaah.in v2.1

## Svelte 5 SPA — Audit & Restructuring Plan

| Item | Detail |
|------|--------|
| **Product** | Jamaah.in — SaaS Administrasi Bisnis Travel Umrah & Haji |
| **Frontend** | Svelte 5 SPA, Vite 7, TailwindCSS 3, Lucide Icons |
| **Version Target** | v2.1 — Frontend Restructure |
| **Current Status** | ~14,000+ lines across ~30 page components, single App.svelte router |
| **Document Date** | 2026-06-08 |
| **Author** | Agent Analysis |

---

## 1. Executive Summary

The Jamaah.in frontend is a Svelte 5 SPA that has grown organically from v1 (AI Scanner + basic jamaah management) into v2 (full travel agency ERP). The codebase is functional but is approaching a complexity ceiling that will soon make development velocity unsustainable. This PRD documents the current state in detail and proposes a structured refactoring path.

**Key Finding**: The codebase is NOT broken — it's well-organized for its current scale. But with 28+ page components, a 954-line App.svelte acting as router/state-manager/SEO-manager, and pages reaching 1,820 lines (ProfilePage.svelte), the architecture needs decomposition before adding more features.

### Current Pain Points (by severity)

| # | Pain Point | Severity | Impact |
|---|-----------|----------|--------|
| 1 | App.svelte is a monolithic router (954 lines) with inline lazy-loading, SEO management, auth state, page-level routing, and modal control all in one file | Critical | Every navigation change touches this file |
| 2 | Pages average 500+ lines with no sub-component extraction (ProfilePage = 1,820, VendorsPage = 904, PackagesPage = 885) | High | Hard to test, review, or reuse UI pieces |
| 3 | No centralized route definitions — pages referenced via string literals in 3-4 places (App.svelte ensurePage, template {#if} chains, Sidebar) | High | Adding a page requires touching 3+ files in multiple places |
| 4 | Component library is under-utilized — shared components exist (TableResult, StatusBadge, SlideDrawer) but many pages have inline duplicate UI patterns | High | Inconsistent UX across modules |
| 5 | No TypeScript — all JavaScript with JSDoc annotations | Medium | Growing codebase will benefit from type safety |
| 6 | SEO strings duplicated as a massive object literal in App.svelte | Medium | 100+ lines of static SEO config polluting the router |
| 7 | State management is ad-hoc — single stores/app.js with 4 exports, most pages manage state inline with $state() | Medium | No cross-page state sharing pattern |
| 8 | API layer is well-factored but piCore.js has no request/response interceptors, retry logic, or offline queue | Low-Medium | Works but won't scale for offline/mobile use |

---

## 2. Current Architecture Deep-Dive

### 2.1 Tech Stack

| Layer | Technology | Version | Notes |
|-------|-----------|---------|-------|
| Framework | Svelte 5 | 5.43.8 | Runes mode ($state, $derived, $effect, $props) |
| Build | Vite | 7.2.4 | Dev proxy to Go API gateway on :8080 |
| CSS | TailwindCSS | 3.4.0 | Custom design system in 	ailwind.config.js + pp.css |
| Icons | Lucide Svelte | 0.563.0 | Used consistently across all components |
| Router | Manual (hash-based) | - | No library; currentPage state + window.location.hash + onhashchange |
| HTTP | Native etch | - | Wrapped in piCore.js; JWT Bearer auth; in-memory TTL cache |
| State | Svelte 5 runes | - | $state() for local; minimal global in stores/app.js |
| Testing | Vitest | 4.0.18 | Minimal coverage (1 test file: superAdminApi.test.js) |
| Lint/Check | svelte-check | 4.4.4 | TypeScript checking via jsconfig.json |
| Drag | @neodrag/svelte | 2.0.0 | Used in RoomingPage for drag-drop room assignments |

### 2.2 Directory Structure (Frontend)

\\\
frontend-svelte/
├── src/
│   ├── main.js                          # Entry: mounts App to #app
│   ├── App.svelte                       # 954L — Monolithic router + SEO + auth + modals
│   ├── app.css                          # 371L — Design system CSS variables + base/utility classes
│   └── lib/
│       ├── components/                  # 21 shared components
│       │   ├── BrandLogo.svelte
│       │   ├── Button.svelte
│       │   ├── Card.svelte
│       │   ├── EmptyState.svelte
│       │   ├── FileUpload.svelte
│       │   ├── GroupSelector.svelte
│       │   ├── IDRInput.svelte
│       │   ├── NotificationBell.svelte
│       │   ├── OnboardingModal.svelte
│       │   ├── Pager.svelte
│       │   ├── PageState.svelte
│       │   ├── ProGateScreen.svelte
│       │   ├── RegistrationLinkModal.svelte
│       │   ├── Sidebar.svelte           # 240L — Navigation + mobile drawer + upgrade banner
│       │   ├── SkeletonLoader.svelte
│       │   ├── SlideDrawer.svelte
│       │   ├── StatusBadge.svelte
│       │   ├── SubscriptionBanner.svelte
│       │   ├── SupportChatBubble.svelte
│       │   ├── TableResult.svelte
│       │   ├── Toast.svelte
│       │   ├── UpgradeModal.svelte
│       │   ├── WhatsAppBlast.svelte
│       │   └── super-admin/             # 4 components for super-admin panel
│       │       ├── Charts.svelte
│       │       ├── StatsCards.svelte
│       │       ├── TicketDetail.svelte
│       │       ├── TicketList.svelte
│       │       └── UserManagement.svelte
│       ├── pages/                       # 28 page components
│       │   ├── Dashboard.svelte         # 637L
│       │   ├── LandingPage.svelte       # 853L — Public marketing landing page
│       │   ├── Login.svelte             # 614L — Login + Register + OTP flow
│       │   ├── ProfilePage.svelte       # 1820L — User profile + subscription + team settings
│       │   ├── SuperAdminDashboard.svelte # 407L
│       │   ├── ScannerPage.svelte       # 657L — AI OCR document scanning
│       │   ├── DataJamaahPage.svelte    # 196L
│       │   ├── GroupsPage.svelte        # 172L
│       │   ├── ManifestPage.svelte      # 175L
│       │   ├── MutawwifManifest.svelte  # Public manifest view
│       │   ├── PublicRegistrationPage.svelte  # Self-service registration
│       │   ├── PublicPackagePage.svelte       # Public package detail
│       │   ├── PublicContractSigningPage.svelte # Public contract signing
│       │   ├── PackagesPage.svelte      # 885L — Package management + tiers
│       │   ├── CRMPage.svelte           # 402L — CRM + jamaah pipeline
│       │   ├── InvoicesPage.svelte      # 439L — Invoices + payments
│       │   ├── FinancePage.svelte       # 300L — Financial reports
│       │   ├── VendorsPage.svelte       # 904L — Vendor management
│       │   ├── AgentsPage.svelte        # 202L — Agent commissions
│       │   ├── ContractsPage.svelte     # 767L — E-contract templates
│       │   ├── CancellationPage.svelte  # 317L
│       │   ├── PayrollPage.svelte       # 271L
│       │   ├── InventoryPage.svelte     # 414L
│       │   ├── RoomingPage.svelte       # 688L
│       │   ├── TeamPage.svelte          # 502L
│       │   ├── AnalyticsPage.svelte
│       │   └── ItineraryPage.svelte
│       ├── services/
│       │   ├── apiCore.js               # 89L — fetch wrapper, JWT, error parsing
│       │   ├── api.js                   # 93L — ApiService factory + in-memory TTL cache
│       │   ├── toast.svelte.js          # 97L — Toast notification system (Svelte 5 runes)
│       │   ├── superAdminApi.js         # Super-admin API methods
│       │   ├── superAdminApi.test.js    # Single test file
│       │   └── apiDomains/              # 17 domain-specific API modules
│       │       ├── agentApi.js
│       │       ├── authSubscriptionApi.js   # ~200L — Auth + subscription + trial
│       │       ├── contentApi.js
│       │       ├── contractApi.js
│       │       ├── documentExcelApi.js
│       │       ├── exportApi.js
│       │       ├── groupOpsApi.js
│       │       ├── invoiceApi.js
│       │       ├── jamaahApi.js
│       │       ├── packageApi.js
│       │       ├── paymentApi.js
│       │       ├── payrollApi.js
│       │       ├── refundApi.js
│       │       ├── registrationApi.js
│       │       ├── supportTicketApi.js
│       │       ├── vendorApi.js
│       │       └── apiUtils.js          # Shared unwrapData helper
│       ├── stores/
│       │   └── app.js                   # 19L — minimal: currentUser, activePackageId
│       ├── utils/
│       │   ├── formatting.js            # 51L — formatRupiah, formatDate, formatDateTime, formatPct
│       │   └── errors.js
│       └── config/
│           └── pricing.js               # Pricing plan constants
├── public/                              # Static assets + SEO landing pages
├── package.json
├── vite.config.js                       # Dev server proxy + test config
├── tailwind.config.js                   # Custom colors (primary blue, emerald, gold) + animations
├── svelte.config.js
├── postcss.config.js
└── jsconfig.json
\\\


### 2.3 Page Size Analysis

\\\
ProfilePage.svelte        1820L  ████████████████████████████████████████  (outlier — needs decomposition)
VendorsPage.svelte         904L  ████████████████████████
PackagesPage.svelte        885L  ████████████████████████
LandingPage.svelte         853L  ████████████████████████
ContractsPage.svelte       767L  ██████████████████████
RoomingPage.svelte         688L  ███████████████████
ScannerPage.svelte         657L  ██████████████████
Dashboard.svelte           637L  █████████████████
Login.svelte               614L  █████████████████
TeamPage.svelte            502L  █████████████
InvoicesPage.svelte        439L  ████████████
InventoryPage.svelte       414L  ███████████
SuperAdminDashboard.svelte 407L  ███████████
CRMPage.svelte             402L  ███████████
CancellationPage.svelte    317L  ████████
FinancePage.svelte         300L  ████████
PayrollPage.svelte         271L  ███████
Sidebar.svelte             240L  ██████
AgentsPage.svelte          202L  █████
DataJamaahPage.svelte      196L  █████
ManifestPage.svelte        175L  ████
GroupsPage.svelte          172L  ████
App.svelte                 954L  ██████████████████████████  (critical — router bottleneck)
app.css                    371L  █████████
\\\

**Total**: ~14,000+ lines of frontend code (excluding node_modules and dist).

### 2.4 Routing Architecture (Current)

The current routing is a **hash-based manual system**:

\\\
URL hash pattern           →  Page                      Auth Required?
──────────────────────────────────────────────────────────────────────
#/                          →  LandingPage               No
#/                          →  Login (if not logged in via hash check) No
#/m/{token}                 →  MutawwifManifest          No
#/reg/{token}               →  PublicRegistrationPage     No
#/paket/{slug}              →  PublicPackagePage          No
#/kontrak/{token}           →  PublicContractSigningPage  No
#/super-admin               →  SuperAdminDashboard        Yes (super admin)
internal currentPage state  →  All authenticated pages    Yes (JWT)
\\\

**How it works**: 
1. getInitialPageAndTokens() runs synchronously before first render to parse window.location.hash
2. window.addEventListener("hashchange", ...) handles subsequent navigation
3. HandlePageChange(page) function: sets currentPage, calls ensurePage(page) for lazy import, updates updateSeoMeta(page), manages hash
4. Template uses massive {#if currentPage === "..."} chain (~200 lines) to render the correct component
5. Sidebar calls onPageChange callback (which is HandlePageChange)

**Problems with this approach**:
- Adding a new page requires changes in 4 locations: ensurePage(), {#if} template chain, sidebar 
avGroups, and SEO config
- No URL-based deep linking within authenticated pages (always shows #/ hash)
- No back/forward navigation support for admin pages
- Component lazy-loading and template conditional rendering are redundant (the template {#if} checks act as lazy-load triggers)

### 2.5 API Layer Architecture

The API layer is **the most well-architected part of the frontend**. It follows a clean factory pattern:

\\\
apiCore.js          → fetch wrapper, JWT auth, error parsing
api.js              → ApiService object (central API surface) + TTL cache
apiDomains/         → Domain-specific API modules (17 files)
    Each domain file exports a factory function that receives 
    cache utilities and returns an object of async methods.
    Methods are merged into ApiService via Object.assign()
\\\

**Good patterns**:
- unwrapData() helper strips {success: true, data: {...}} wrappers
- In-memory TTL cache with cacheGet, cacheSet, cacheInvalidate(prefix)
- Error parsing handles nested Go error structures gracefully
- Indonesian error mapping in 	oast.svelte.js

**Gaps**:
- No request/response interceptors (logging, token refresh)
- No retry logic for transient failures
- No request deduplication (concurrent identical requests both hit the network)
- No offline support or optimistic updates
- Cache invalidation is manual per-mutation (could miss related caches)

### 2.6 State Management Assessment

The project uses **Svelte 5 runes** consistently. State is:

- **Page-local**: $state() in each page component for UI state, form data, loading flags
- **Global**: stores/app.js with 4 exports (getCurrentUser, setCurrentUser, getActivePackageId, setActivePackageId)
- **Derived**: $derived() for computed values (e.g., isPro from subscription object)
- **Cross-component**: showToast() function from 	oast.svelte.js acts as global notification bus

This is appropriate for the current scale and follows Svelte 5 best practices. No action needed until we grow past ~15 developers or need complex cross-module state.

### 2.7 Design System Inventory

The design system is **partially defined** across 3 files:

| File | Contents |
|------|----------|
| 	ailwind.config.js | Color tokens (primary blue, emerald, gold), custom fonts (Plus Jakarta Sans, Playfair Display), custom animations (float, fadeUp, slideRight) |
| pp.css | CSS variables, base styles, utility classes (.card, .btn, .badge, .input, .table, .stat-card), scrollbar, focus states |
| Component files | Individual component styles (e.g., Sidebar-link styles are scoped in Sidebar.svelte \<style>\ block) |

**Issues**:
- Utility classes and component styles overlap (.btn in CSS + Tailwind classes used inline on buttons)
- Some components use scoped CSS, others rely entirely on Tailwind/global classes
- No design token documentation — new developers must reverse-engineer from CSS variables
- Dark mode is configured (darkMode: 'class' in Tailwind) but no dark mode styles implemented

---

## 3. Backend Microservices Map (Frontend Context)

Understanding the backend is essential for the frontend refactor plan:

| Service | Port | DB | Domain | Frontend API Module |
|---------|------|-----|--------|-------------------|
| api-gateway | 8080 (HTTP) | - | Routes to gRPC backends | All modules proxy through /api |
| auth-service | 50051 (gRPC) | postgres:auth_db | Auth, users, subscription | authSubscriptionApi.js |
| package-service | 50052 (gRPC) | postgres:package_db | Package CRUD, pricing tiers | packageApi.js |
| jamaah-service | 50053 (gRPC) | postgres:jamaah_db | Jamaah data, groups, passport | jamaahApi.js |
| invoice-service | 50054 (gRPC) | postgres:invoice_db | Invoices, payments, invoice items | invoiceApi.js, paymentApi.js |
| finance-service | 50055 (gRPC) | postgres:finance_db | Financial reports, P&L | (finance methods in ApiService) |
| ai-ocr-service | 50056 (gRPC) | - | Document scanning, OCR | (scanner methods) |
| vendor-service | 50057 (gRPC) | postgres:vendor_db | Vendors, expenses | vendorApi.js |
| contract-service | 50058 (gRPC) | postgres:contract_db | E-contracts, signatures | contractApi.js |
| inventory-service | 50059 (gRPC) | postgres:inventory_db | Stock, items | (inventory methods) |
| payroll-service | 50060 (gRPC) | postgres:payroll_db | Payroll, employees, salary | payrollApi.js |
| agent-service | 50061 (gRPC) | postgres:agent_db | Agents, commissions | agentApi.js |

**Key insight**: Each microservice has its own database. Frontend API calls that span domains (e.g., dashboard stats, cross-service reports) go through the gateway which aggregates from multiple gRPC backends.

### Infrastructure Dependencies

| Service | Purpose | Frontend Impact |
|---------|---------|-----------------|
| PostgreSQL 16 | Primary data store per service | - |
| Redis 7 | Caching, rate limiting, sessions | Auth token validation |
| NATS 2 | Inter-service messaging (JetStream) | - |
| MinIO | Object storage (documents, exports) | File uploads/downloads via presigned URLs |

---

## 4. Feature Module Inventory (Current vs Target)

### 4.1 v1 Modules (Existing — to be preserved)

| Module | Pages | Components | Status | Notes |
|--------|-------|------------|--------|-------|
| AI Scanner | ScannerPage.svelte | FileUpload, Toast | Stable | Core differentiator; OCR for KTP/KK/passport |
| Landing Page | LandingPage.svelte | - | Stable | Public marketing site |
| Auth | Login.svelte | - | Stable | Login, register, OTP, forgot password |
| Dashboard | Dashboard.svelte | - | Stable | Stats cards, trend charts, quick actions |
| Jamaah Data | DataJamaahPage, GroupsPage | TableResult, Pager, GroupSelector | Stable | CRUD for jamaah per group |
| Manifest | ManifestPage, MutawwifManifest | - | Legacy | Public manifest with PIN; low usage |
| Rooming | RoomingPage.svelte | @neodrag/svelte | Stable | Drag-drop room assignment |
| Analytics | AnalyticsPage.svelte | - | Simple | Basic stats view |
| Itinerary | ItineraryPage.svelte | - | Simple | Trip schedule view |
| Profile | ProfilePage.svelte | - | Needs split | 1820L monster — profile, subscription, team, security, notifications all in one page |

### 4.2 v2 Modules (New — partially implemented)

| Module | Pages | Status | Completeness |
|--------|-------|--------|-------------|
| Packages | PackagesPage.svelte | Implemented | ~80% — pricing tiers, publish flow done |
| CRM | CRMPage.svelte | Implemented | ~70% — pipeline, kanban, details done |
| Invoices | InvoicesPage.svelte | Implemented | ~75% — invoice CRUD, payment recording done |
| Finance | FinancePage.svelte | Implemented | ~60% — reports exist, P&L needs work |
| Vendors | VendorsPage.svelte | Implemented | ~85% — full CRUD + expense tracking |
| Agents | AgentsPage.svelte | Implemented | ~70% — commission tracking done |
| Contracts | ContractsPage.svelte | Implemented | ~80% — template + signing flow |
| Cancellation | CancellationPage.svelte | Implemented | ~80% — refund workflow |
| Payroll | PayrollPage.svelte | Implemented | ~70% — payroll period + slip generation |
| Inventory | InventoryPage.svelte | Implemented | ~75% — stock in/out tracking |
| Export | ExportPage.svelte | Partially done | ~50% — basic export only |
| Super Admin | SuperAdminDashboard + 4 components | Implemented | ~70% — tickets, users, analytics |

### 4.3 Public-Facing Pages (No auth)

| Page | Hash Pattern | Token/Slug | Frontend Module |
|------|-------------|------------|----------------|
| Landing | #/ | - | LandingPage.svelte |
| Public Package | #/paket/{slug} | slug | PublicPackagePage.svelte |
| Public Registration | #/reg/{token} | token | PublicRegistrationPage.svelte |
| Public Contract Signing | #/kontrak/{token} | token | PublicContractSigningPage.svelte |
| Mutawwif Manifest | #/m/{token} | token | MutawwifManifest.svelte |

---

## 5. Refactoring Plan

### 5.0 Guiding Principles

1. **Incremental, not rewrite**: Refactor one module at a time while keeping the app functional
2. **Component extraction first**: Pull sub-components out of pages before touching routing
3. **Backward compatible**: No breaking changes to existing API contracts
4. **Svelte 5 runes only**: No legacy Svelte 4 stores; use $state///
5. **Indonesian UI language**: All user-facing strings remain in Bahasa Indonesia
6. **No SvelteKit migration yet**: backend is still stabilizing; defer full framework migration

### 5.1 Phase 1: Stabilize Foundation (Week 1-2)

**Goal**: Reduce App.svelte complexity and establish routing patterns without breaking anything.

#### Step 1.1: Extract SEO Config

Move PAGE_SEO and upsertMeta/upsertLinkRel/updateSeoMeta out of App.svelte:

\\\
New file: src/lib/config/seo.js
- Export PAGE_SEO (the existing object)
- Export function updateSeoMeta(page)
- App.svelte imports and uses it, removing ~130 lines
\\\

#### Step 1.2: Extract Route Definitions

Create a centralized route registry:

\\\
New file: src/lib/config/routes.js
- Define all routes as an array of objects:
  { id, page, label, icon, hashPattern, authRequired, proRequired, ownerOnly }
- Used by: App.svelte (routing logic), Sidebar (nav rendering), ensurePage (lazy loading)
- Eliminates string duplication across 4+ files

New file: src/lib/config/navGroups.js
- Nav groups extracted from Sidebar.svelte into a config object
- Sidebar imports it instead of inline definition
\\\

#### Step 1.3: Extract Page Loader Logic

\\\
New file: src/lib/router/pageLoader.js
- Dynamic import map: { dashboard: () => import('../pages/Dashboard.svelte'), ... }
- Exports: ensurePage(page) and getPageComponent(page)
- Removes ~80 lines of if/else chain from App.svelte
\\\

#### Step 1.4: Extract Auth/Subscription Logic

\\\
New file: src/lib/stores/auth.js (rename/reorganize from stores/app.js)
- Current user state
- Subscription status
- Login/logout/refresh helpers
- Trial status helpers
- Exported as Svelte 5 runes-compatible getter/setter functions
\\\

**Expected Outcome**: App.svelte reduces from 954L to ~400L. Clear separation of concerns.

### 5.2 Phase 2: Component Extraction (Week 2-4)

**Goal**: Decompose oversized pages into reusable sub-components.

#### ProfilePage.svelte (1820L → ~200L + 8 sub-components)

| Sub-component | Content | Est. Lines |
|--------------|---------|------------|
| ProfileAvatar.svelte | Avatar color picker, name edit | ~150L |
| ProfileSecurity.svelte | Password change, 2FA, sessions | ~200L |
| ProfileSubscription.svelte | Plan display, upgrade banner, trial | ~200L |
| ProfileTeam.svelte | Team member list, invite management | ~250L |
| ProfileActivity.svelte | Activity log | ~120L |
| ProfileDangerZone.svelte | Account deletion flow | ~100L |
| ProfileNotifications.svelte | Notification preferences | ~150L |
| ProfilePhoneVerify.svelte | Phone verification with OTP | ~180L |

#### VendorsPage.svelte (904L → ~200L + 5 sub-components)

| Sub-component | Content | Est. Lines |
|--------------|---------|------------|
| VendorList.svelte | Vendor table/filter/search | ~200L |
| VendorForm.svelte | Create/edit vendor form | ~250L |
| VendorDetail.svelte | Vendor detail with expense history | ~200L |
| VendorExpenseForm.svelte | Record expense form | ~150L |
| VendorExpenseList.svelte | Expense list for a vendor | ~120L |

#### PackagesPage.svelte (885L → ~200L + 4 sub-components)

| Sub-component | Content | Est. Lines |
|--------------|---------|------------|
| PackageList.svelte | Package cards with status filter | ~200L |
| PackageForm.svelte | Multi-step create/edit form | ~300L |
| PackageDetail.svelte | Package detail panel | ~250L |
| PricingTiers.svelte | Editable pricing tiers table | ~200L |

#### ContractsPage.svelte (767L → ~150L + 4 sub-components)

| Sub-component | Content |
|--------------|---------|
| ContractList.svelte | Template + signed contract list |
| ContractTemplateForm.svelte | Template editor with placeholders |
| ContractSigningPanel.svelte | Signature pad + accept flow |
| ContractPreview.svelte | Contract preview renderer |

#### Remaining Pages (moderate extraction)

| Page | Sub-components to extract |
|------|--------------------------|
| Dashboard (637L) | DashboardStatCards, DashboardCharts, DashboardTrendChart, DashboardQuickActions |
| Login (614L) | LoginForm, RegisterForm, OtpVerification, ForgotPassword |
| ScannerPage (657L) | ScannerUploadZone, ScannerResultList, ScannerExportPanel |
| RoomingPage (688L) | RoomingHotelCard, RoomingJamaahList, RoomingAssignmentPanel |
| TeamPage (502L) | TeamMemberList, TeamInviteForm, TeamRoleEditor |
| InvoicesPage (439L) | InvoiceList, InvoiceForm, PaymentForm, InvoiceDetail |
| InventoryPage (414L) | InventoryList, InventoryForm, StockMovementLog |
| SuperAdminDashboard (407L) | Leverage existing super-admin/ sub-components; reduce main page |

### 5.3 Phase 3: Router Modernization (Week 4-5)

**Goal**: Replace the {#if currentPage === "..."} chain with a declarative route component.

#### New Router Component

\\\
src/lib/router/
├── Router.svelte          # Route matching + component rendering
├── RouteGuard.svelte      # Auth check + ProGateScreen integration
├── routes.js              # Route config (moved from Phase 1)
└── pageLoader.js          # Lazy import map (moved from Phase 1)
\\\

\\\svelte
<!-- Router.svelte usage in App.svelte -->
<Router {currentPage} {user} {subscription} {isPro}>
  <RouteGuard required={auth} fallback={LoginPage}>
    <Route page="dashboard" component={DashboardPage} />
    <Route page="scanner" component={ScannerPage} />
    <!-- ... -->
  </RouteGuard>
</Router>
\\\

#### Hash-based Navigation Improvements

- All admin pages get meaningful hash routes: #/dashboard, #/paket, #/crm, etc.
- Back/forward browser navigation works for admin SPA
- Deep-linking: #/paket/{id}/detail opens package detail directly
- Route guards handle auth redirect (unauthenticated → #/ with login modal)

### 5.4 Phase 4: Shared Component Enhancement (Week 5-6)

**Goal**: Standardize UI patterns and reduce inline duplication across pages.

#### Components to Create/Enhance

| Component | Purpose | Priority |
|-----------|---------|----------|
| DataTable.svelte | Sortable, paginated generic table (replace inline tables across 12+ pages) | High |
| ConfirmDialog.svelte | Standardized delete/confirm modal (currently duplicated in 8+ pages) | High |
| FormField.svelte | Label + input + error message wrapper | Medium |
| SearchBar.svelte | Debounced search input (duplicated in 6+ pages) | Medium |
| FilterTabs.svelte | Status filter tabs (duplicated in 5+ pages) | Medium |
| DetailDrawer.svelte | Extension of SlideDrawer with loading/error states | Medium |
| EmptyState.svelte | Already exists — enhance with action button slot | Low |
| PageHeader.svelte | Consistent page title + subtitle + action button bar | Low |
| StatCard.svelte | Stat display with icon + trend indicator (used in Dashboard + other pages) | Low |
| KeyValueList.svelte | Label-value pairs (used in detail views across 8+ pages) | Low |

### 5.5 Phase 5: API Layer Hardening (Week 6-7)

**Goal**: Improve reliability and developer experience.

#### Changes

1. **Request Interceptor Pipeline** in piCore.js:
   - Logging (dev mode only)
   - Automatic token refresh on 401
   - Request deduplication (concurrent identical GETs share one promise)
   - Rate limit awareness (respect Retry-After headers)

2. **Offline Queue** (optional, for mobile/PWA):
   - Queue failed mutations when offline
   - Replay when connectivity returns
   - Conflict resolution strategy (last-write-wins)

3. **Type Generation**:
   - If backend exposes OpenAPI/Swagger specs, generate JS type definitions
   - Otherwise, add JSDoc type annotations to API domain modules manually

### 5.6 Phase 6: Performance & UX Polish (Week 7-8)

#### Bundle Size Optimization

- Audit all page imports for unused Lucide icons (some pages import 20+ icons)
- Tree-shake and split icons per page
- Consider lazy loading icon sprites vs individual imports

#### Mobile Responsiveness

- Audit all pages for mobile usability
- Ensure SlideDrawer works correctly on mobile (touch-friendly close button, swipe-to-close)
- Fix any overflow/horizontal scroll issues in tables on small screens

#### Loading States

- Standardize: skeleton loaders for all data tables and stat cards
- Add optimistic UI updates for quick actions (status changes, toggles)
- Add transition animations between page navigations

#### Accessibility

- Add ria-label to icon-only buttons
- Ensure keyboard navigation works across all modals and drawers
- Add focus trap to drawers and modals
- Semantic heading hierarchy (h1 → h2 → h3)

---

## 6. Component Dependency Map

### 6.1 App.svelte Dependencies (Current)

\\\
App.svelte
├── LandingPage (always bundled — first import)
├── Login (always bundled)
├── Sidebar (always bundled)
├── ProGateScreen (always bundled)
├── SupportChatBubble (lazy)
├── UpgradeModal (always bundled)
├── Toast (always bundled)
├── ApiService (always bundled)
└── Lazy imports (28 pages, loaded on demand)
    ├── Dashboard.svelte
    ├── ProfilePage.svelte
    ├── SuperAdminDashboard.svelte
    ├── ... (all pages)
    └── [All pages imported via ensurePage()]
\\\

**Bundle impact**: The initial JavaScript bundle includes App.svelte, LandingPage, Login, Sidebar, ProGateScreen, UpgradeModal, Toast, ApiService, and all API domain modules (because pi.js statically imports all domain files). This is approximately 200-300KB gzipped for the critical path.

### 6.2 Shared Component Usage Matrix

| Component | Used In (Pages) |
|-----------|----------------|
| Sidebar | App.svelte (every authenticated page) |
| Toast | App.svelte (global) |
| UpgradeModal | App.svelte (global) |
| SlideDrawer | PackagesPage, CRMPage, VendorsPage, InvoicesPage, ContractsPage, InventoryPage, PayrollPage |
| StatusBadge | PackagesPage, CRMPage, InvoicesPage, VendorsPage, AgentsPage |
| IDRInput | PackagesPage, InvoicesPage, VendorsPage, FinancePage, PayrollPage |
| TableResult | DataJamaahPage, VendorsPage, AgentsPage, PayrollPage, InventoryPage |
| FileUpload | ScannerPage, DataJamaahPage, ContractsPage |
| Pager | DataJamaahPage, InvoicesPage, VendorsPage |
| PageState | Dashboard, PackagesPage, CRMPage, InvoicesPage |
| GroupSelector | DataJamaahPage, Dashboard |
| SkeletonLoader | Dashboard, PackagesPage |
| EmptyState | PackagesPage, InvoicesPage, VendorsPage |
| Button | Multiple pages (generic) |
| Card | Multiple pages (generic) |
| ProGateScreen | App.svelte (wraps pro-only pages) |
| BrandLogo | Sidebar, Login |
| WhatsAppBlast | DataJamaahPage |
| NotificationBell | Dashboard |
| OnboardingModal | Dashboard (first-time) |
| SubscriptionBanner | Dashboard, PackagesPage |
| RegistrationLinkModal | DataJamaahPage |

---

## 7. Risks & Mitigations

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|-----------|
| Breaking existing functionality during refactor | Medium | High | Phase-by-phase, keep app running at each step; comprehensive manual testing checklist |
| Backend API changes during refactor | Low | Medium | Freeze API contracts before Phase 3; API modules are already well-isolated |
| Scope creep (trying to fix everything) | High | Medium | Strict phase boundaries; each phase has defined acceptance criteria |
| Performance regression from lazy loading | Low | Low | Vite does code splitting natively; measure bundle size before/after each phase |
| Loss of institutional knowledge | Medium | Medium | Document patterns created; inline JSDoc on extracted components |
| User confusion from hash URL changes | Low | Medium | Keep backward-compatible hash patterns; redirect old hashes |
| Svelte 5 runes migration pitfalls | Low | Medium | Already on Svelte 5; no Svelte 4 store migration needed |

---

## 8. Success Metrics

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| App.svelte lines | 954 | < 400 | wc -l |
| Pages with sub-components | 4 of 28 | > 20 of 28 | File count in lib/components/ per page |
| Max page component size | 1,820L (ProfilePage) | < 500L for any page | wc -l |
| Route registration points | 4 (App.svelte template, ensurePage, Sidebar, SEO) | 1 (routes.js config) | Manual count |
| Initial JS bundle size | ~200-300KB gzipped (est.) | < 250KB gzipped | ite build output |
| Lighthouse performance score | Unknown | > 90 | Lighthouse audit |
| API layer test coverage | 1 test file | > 5 test files | Vitest coverage report |
| Time to add new page | ~15 min (4 files) | ~3 min (1 file + route config) | Developer self-timing |

---

## 9. Implementation Order (Priority Matrix)

### Must Have (Phase 1-2)

| # | Task | Effort | Impact | Risk |
|---|------|--------|--------|------|
| 1 | Extract SEO config from App.svelte | 1h | Low | Very Low |
| 2 | Extract route definitions to outes.js | 2h | High | Low |
| 3 | Extract page loader (ensurePage) to separate module | 2h | Medium | Low |
| 4 | Decompose ProfilePage.svelte | 6h | High | Medium |
| 5 | Decompose VendorsPage.svelte | 4h | High | Low |
| 6 | Decompose PackagesPage.svelte | 4h | High | Low |
| 7 | Decompose ContractsPage.svelte | 3h | High | Low |

### Should Have (Phase 3-4)

| # | Task | Effort | Impact | Risk |
|---|------|--------|--------|------|
| 8 | Create Router.svelte component | 5h | Very High | Medium |
| 9 | Create RouteGuard.svelte | 2h | High | Low |
| 10 | Add meaningful hash routes for all pages | 3h | Medium | Medium |
| 11 | Create DataTable.svelte shared component | 4h | High | Medium |
| 12 | Create ConfirmDialog.svelte | 2h | Medium | Low |
| 13 | Create FilterTabs.svelte, SearchBar.svelte | 2h | Medium | Low |

### Nice to Have (Phase 5-6)

| # | Task | Effort | Impact | Risk |
|---|------|--------|--------|------|
| 14 | API request interceptor pipeline | 4h | Medium | Medium |
| 15 | Token auto-refresh on 401 | 3h | Medium | Medium |
| 16 | Bundle size optimization | 3h | Low | Low |
| 17 | Mobile responsiveness audit | 4h | Medium | Low |
| 18 | Accessibility improvements | 3h | Medium | Low |
| 19 | Add Vitest tests for API modules | 6h | Medium | Low |

---

## 10. Appendix: Current Code Smells Catalog

### A. Duplicated Patterns (counted across pages)

| Pattern | Occurrences | Affected Files |
|---------|-------------|---------------|
| let isLoading = (true) + try/catch in onMount | 20+ | Every data-fetching page |
| Inline confirm delete dialog with let showDeleteConfirm = (false) | 8+ | Packages, Vendors, Contracts, Inventory, Payroll, Agents, Invoices, Team |
| Status filter tabs (All/Active/Archived etc.) | 6+ | Packages, Invoices, Vendors, Contracts, Agents, Payroll |
| Search input with debounce | 6+ | DataJamaah, CRM, Invoices, Vendors, Agents, Inventory |
| "Rp X.XXX.XXX" manual formatting (should use ormatRupiah) | 4+ | Old pages that predate formatting.js |
| {#if isLoading} → Skeleton → {:else if error} → Error → {:else} → Content | 15+ | Most pages |
| {#if !items || items.length === 0} → EmptyState | 10+ | Most list pages |

### B. Anti-patterns Found

| Anti-pattern | Location | Risk |
|-------------|----------|------|
| 1,820-line single component (ProfilePage) | ProfilePage.svelte | Maintainability nightmare |
| SEO strings as 130-line object literal in router | App.svelte:14-144 | Configuration in logic file |
| Sidebar nav items hardcoded in component | Sidebar.svelte:45-110 | Can't reuse nav config for breadcrumbs/mobile |
| ensurePage() as 80-line if/else chain | App.svelte:218-297 | Violates Open/Closed principle |
| {#if currentPage === "..."} chain (200+ lines) | App.svelte:720-920 | Template bloat, poor readability |
| unsubscribe from onhashchange not cleaned up | App.svelte | Potential memory leak on hot reload |
| Direct localStorage access in apiCore.js | apiCore.js:13 | Not SSR-compatible, no error handling for private browsing |
| Map cache with no max size limit | api.js:17-30 | Memory leak risk for long sessions |

### C. Missing Patterns

| Pattern | Importance | Notes |
|---------|-----------|-------|
| Breadcrumb navigation | Medium | Users can't see where they are in deep pages |
| Notification center | Low | NotificationBell exists but no full notification list page |
| Keyboard shortcuts | Low | Power users (owner/admin) would benefit |
| Bulk actions on data tables | Medium | Select-all + bulk delete/export not implemented consistently |
| Form validation library | Low | Currently manual validation per form; no shared validation helper |
| Error boundary component | Medium | Uncaught errors crash the whole SPA; no {#catch} fallback |
| Analytics/tracking | Low | No page view tracking or event analytics |

---

## 11. Appendix: File-by-File Assessment

### lib/components/ (Shared Components)

| File | Lines | Quality | Notes |
|------|-------|---------|-------|
| BrandLogo.svelte | ~40 | Good | Clean, single purpose |
| Button.svelte | ~60 | Good | Generic button with variants |
| Card.svelte | ~30 | Good | Simple wrapper |
| EmptyState.svelte | ~50 | Adequate | Needs action slot |
| FileUpload.svelte | ~120 | Good | Drag-drop + progress |
| GroupSelector.svelte | ~80 | Good | Group dropdown with search |
| IDRInput.svelte | ~80 | Good | Auto-format IDR input |
| NotificationBell.svelte | ~70 | Good | Real-time notification count |
| OnboardingModal.svelte | ~100 | Good | First-time user walkthrough |
| Pager.svelte | ~60 | Good | Pagination controls |
| PageState.svelte | ~50 | Good | Loading/error/empty wrapper |
| ProGateScreen.svelte | ~80 | Good | Upgrade prompt for pro features |
| RegistrationLinkModal.svelte | ~80 | Good | Share registration link |
| Sidebar.svelte | 240 | Good | Nav groups should be externalized |
| SkeletonLoader.svelte | ~40 | Good | Basic skeleton |
| SlideDrawer.svelte | ~100 | Good | Needs loading/error states |
| StatusBadge.svelte | ~50 | Good | Color-coded status badges |
| SubscriptionBanner.svelte | ~60 | Good | Upgrade CTA banner |
| SupportChatBubble.svelte | ~100 | Adequate | Chat widget |
| TableResult.svelte | ~100 | Good | Generic table with sorting |
| Toast.svelte | ~60 | Good | Toast display component |
| UpgradeModal.svelte | ~200 | Adequate | Could extract pricing display |
| WhatsAppBlast.svelte | ~100 | Good | WhatsApp template sender |

### lib/pages/ (Page Components)

| File | Lines | Refactor Priority | Key Issues |
|------|-------|-------------------|-----------|
| ProfilePage.svelte | 1820 | **Critical** | 8 concerns in 1 file; should be 8 sub-components |
| VendorsPage.svelte | 904 | **High** | Vendor CRUD + expense tracking in 1 file |
| PackagesPage.svelte | 885 | **High** | Package list + form + detail + pricing tiers inline |
| LandingPage.svelte | 853 | Low | Public page; acceptable as-is for marketing site |
| ContractsPage.svelte | 767 | **High** | Template editor + signing flow mixed |
| RoomingPage.svelte | 688 | Medium | Drag-drop logic is complex; extraction may be tricky |
| ScannerPage.svelte | 657 | Medium | File upload + OCR results + export mixed |
| Dashboard.svelte | 637 | Medium | Could extract stat cards and charts |
| Login.svelte | 614 | Medium | Login + register + OTP + forgot password in 1 file |
| TeamPage.svelte | 502 | Medium | Team list + invite + role management inline |
| InvoicesPage.svelte | 439 | Medium | Invoice list + payment form mixed |
| InventoryPage.svelte | 414 | Medium | Stock list + form + movement log |
| SuperAdminDashboard.svelte | 407 | Low | Uses sub-components; main page is thin |
| CRMPage.svelte | 402 | Medium | Kanban + detail drawer inline |
| CancellationPage.svelte | 317 | Low | Acceptable size |
| FinancePage.svelte | 300 | Low | Acceptable size |
| PayrollPage.svelte | 271 | Low | Acceptable size |
| AgentsPage.svelte | 202 | Low | Small; OK as-is |
| DataJamaahPage.svelte | 196 | Low | Small; OK as-is |
| ManifestPage.svelte | 175 | Low | Small; OK as-is |
| GroupsPage.svelte | 172 | Low | Small; OK as-is |

### lib/services/ (API Layer)

| File | Quality | Notes |
|------|---------|-------|
| apiCore.js | Good | Clean fetch wrapper; needs interceptors and retry |
| api.js | Good | Clean service composition; cache needs size limit |
| toast.svelte.js | Good | Well-structured notification system |
| apiDomains/*.js | Good | Consistent factory pattern across all 17 files |

---

## 12. Migration Notes for Developers

### File Naming Conventions

- **Components**: PascalCase, .svelte extension: DataTable.svelte
- **Page components**: PascalCase, *Page.svelte suffix: PackagesPage.svelte
- **JS modules**: camelCase, .js extension: packageApi.js
- **Config files**: camelCase, .js extension: outes.js

### Component Pattern (Standard)

\\\svelte
<script>
  // 1. Imports
  // 2. Props via ()
  // 3. Local state via ()
  // 4. Derived via ()
  // 5. Effects via ()
  // 6. Functions
</script>

<!-- Template -->

<style>
  /* Scoped styles when needed */
</style>
\\\

### API Module Pattern (Standard)

\\\js
// apiDomains/exampleApi.js
import { API_URL, authHeaders, parseError, apiFetch } from '../apiCore.js';

function unwrapData(json) {
    return json?.success === true && json.data !== undefined ? json.data : json;
}

export function createExampleApi({ cacheInvalidate }) {
    return {
        async list(params = {}) {
            const res = await apiFetch(\\/example?\\, {
                headers: authHeaders(),
            });
            if (!res.ok) throw new Error(await parseError(res));
            return unwrapData(await res.json());
        },
        async create(data) {
            const res = await apiFetch(\\/example\, {
                method: 'POST',
                headers: authHeaders({ 'Content-Type': 'application/json' }),
                body: JSON.stringify(data),
            });
            if (!res.ok) throw new Error(await parseError(res));
            cacheInvalidate?.('example:');
            return unwrapData(await res.json());
        },
    };
}
\\\

### Page Component Pattern (Post-Refactor)

\\\svelte
<script>
  import { onMount } from 'svelte';
  import PageState from '../components/PageState.svelte';
  import SubComponentA from '../components/this-module/SubComponentA.svelte';
  import SubComponentB from '../components/this-module/SubComponentB.svelte';
  import { ApiService } from '../services/api';

  let { user = null, onNavigate } = ();

  let data = (null);
  let loading = (true);
  let error = ('');

  onMount(async () => {
    try { data = await ApiService.getSomething(); }
    catch (e) { error = e.message; }
    finally { loading = false; }
  });
</script>

<PageState {loading} {error}>
  <SubComponentA {data} {user} />
  <SubComponentB {data} onAction={onNavigate} />
</PageState>
\\\

---

## 13. Conclusion

Jamaah.in v2.1 frontend is a well-structured but organically-grown SPA that has reached the complexity threshold where architectural refactoring becomes necessary. The code quality is good — consistent patterns, proper use of Svelte 5 runes, clean API layer — but the monolith App.svelte and oversized page components are becoming bottlenecks.

The recommended approach is **incremental decomposition** across 6 phases spanning 8 weeks, starting with the lowest-risk extractions (config, routing) and progressing to component extraction and finally performance polish. Each phase is independently shippable and testable.

The key insight is: **don't rewrite, decompose**. The existing patterns are correct; they just need to be separated into smaller, focused files.

---

*PRD Version: 1.0 | Date: 2026-06-08 | Author: Automated Codebase Analysis*
