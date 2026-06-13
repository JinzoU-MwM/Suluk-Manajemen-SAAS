// Global session store — replaces the auth/data orchestration that used to live
// in App.svelte's onMount. Holds the current user, subscription, groups and the
// derived Pro / super-admin flags, plus the global upgrade-modal flag. Pages read
// these reactively (session.user, session.isPro, ...) and pass them to the
// existing leaf components as props, so those components stay unchanged.
import { goto } from "$app/navigation";
import { browser } from "$app/environment";
import { ApiService } from "../services/api.js";
import { isProOrHigher } from "../config/pricing.js";

class SessionStore {
  user = $state(null);
  subscription = $state(null);
  subLoaded = $state(false); // true once a subscription fetch has resolved
  groups = $state([]);
  trialAvailable = $state(false);
  upgradeOpen = $state(false); // global UpgradeModal visibility
  loaded = $state(false); // true once the initial getMe round-trip settled

  isPro = $derived(
    isProOrHigher(this.subscription?.plan) && this.subscription?.status !== "expired",
  );
  isSuperAdmin = $derived(this.user?.is_super_admin === true);
  totalJamaahCount = $derived(
    this.groups.reduce((t, g) => t + Number(g.member_count || 0), 0),
  );

  /** Optimistically hydrate user from localStorage (avoids a flash of logged-out). */
  hydrateFromCache() {
    if (!browser) return;
    try {
      const saved = localStorage.getItem("user");
      if (saved) this.user = JSON.parse(saved);
    } catch {
      this.user = null;
    }
  }

  /** Verify the session against the API and load subscription/groups/trial. */
  async loadSession() {
    try {
      const me = await ApiService.getMe();
      this.user = me;
      try {
        localStorage.setItem("user", JSON.stringify(me));
      } catch {}
      const [sub, groupsData, trial] = await Promise.all([
        ApiService.getSubscriptionStatus().catch(() => null),
        ApiService.listGroups().catch(() => ({ groups: [] })),
        ApiService.getTrialStatus().catch(() => null),
      ]);
      this.subscription = sub;
      this.subLoaded = true;
      this.groups = groupsData?.groups || [];
      this.trialAvailable = trial?.trial_available ?? false;
      this.loaded = true;
      return true;
    } catch {
      this.user = null;
      try {
        localStorage.removeItem("user");
      } catch {}
      this.loaded = true;
      return false;
    }
  }

  /** Non-blocking refresh of subscription/groups/trial (after navigation/upgrade). */
  async refresh() {
    try {
      const [sub, groupsData, trial] = await Promise.all([
        ApiService.getSubscriptionStatus(),
        ApiService.listGroups(),
        ApiService.getTrialStatus().catch(() => null),
      ]);
      this.subscription = sub;
      this.subLoaded = true;
      this.groups = groupsData.groups || [];
      this.trialAvailable = trial?.trial_available ?? false;
    } catch {
      /* non-blocking */
    }
  }

  /** Called by Login on success — set user then load the rest. */
  setUserAndLoad(userData) {
    this.user = userData;
    return this.refresh();
  }

  async logout() {
    try {
      await ApiService.logout();
    } catch {}
    this.user = null;
    this.subscription = null;
    this.groups = [];
    this.subLoaded = false;
    try {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
      localStorage.removeItem("user");
    } catch {}
  }

  openUpgrade() {
    this.upgradeOpen = true;
  }
  closeUpgrade() {
    this.upgradeOpen = false;
  }

  // --- mobile mode (Capacitor / PWA shell at /mobile) ---
  get mobileMode() {
    if (!browser) return false;
    try {
      return localStorage.getItem("suluk_mobile") === "1";
    } catch {
      return false;
    }
  }
  setMobileMode(on) {
    if (!browser) return;
    try {
      if (on) localStorage.setItem("suluk_mobile", "1");
      else localStorage.removeItem("suluk_mobile");
    } catch {}
  }
}

export const session = new SessionStore();

// Where an authenticated user lands (honoring mobile mode).
export function homePath() {
  return session.mobileMode ? "/mobile" : "/app";
}

// Logout then send the user back to the right entry point.
export async function logoutAndRedirect() {
  const mobile = session.mobileMode;
  await session.logout();
  goto(mobile ? "/mobile" : "/");
}

// Old sidebar page-keys -> SvelteKit paths. Keeps existing components' callbacks
// (onNavigate("jamaah"), Sidebar onPageChange, etc.) working unchanged.
const PAGE_PATHS = {
  dashboard: "/app",
  jamaah: "/app/jamaah",
  grup: "/app/grup",
  scanner: "/app/scanner",
  profile: "/app/profile",
  inventory: "/app/inventory",
  rooming: "/app/rooming",
  manifest: "/app/manifest",
  analytics: "/app/analytics",
  team: "/app/team",
  itinerary: "/app/itinerary",
  packages: "/app/packages",
  crm: "/app/crm",
  invoices: "/app/invoices",
  finance: "/app/finance",
  akuntansi: "/app/akuntansi",
  vendors: "/app/vendors",
  agents: "/app/agents",
  documents: "/app/documents",
  contracts: "/app/contracts",
  cancellation: "/app/cancellation",
  payroll: "/app/payroll",
  export: "/app/export",
  stock: "/app/stock",
  "super-admin": "/super-admin",
};

export function pagePath(pageKey) {
  return PAGE_PATHS[pageKey] || "/app";
}

// Drop-in for the old handlePageChange — translates upgrade pseudo-pages into the
// modal and everything else into navigation.
export function navigate(pageKey) {
  if (
    pageKey === "profile:upgrade" ||
    pageKey === "header:upgrade" ||
    pageKey === "trial:activate"
  ) {
    session.openUpgrade();
    return;
  }
  goto(pagePath(pageKey));
}

// Marketing slug (About/Pillar/etc.) -> clean path, mirroring the old hash map.
const MARKETING_PATHS = {
  about: "/tentang",
  contact: "/kontak",
  privacy: "/privasi",
  terms: "/ketentuan",
  software: "/software-travel-umrah",
  "fitur-invoice": "/fitur/invoice-umrah",
  "fitur-crm": "/fitur/crm-jamaah",
  "fitur-keuangan": "/fitur/laporan-keuangan",
  "fitur-kontrak": "/fitur/e-kontrak",
  "fitur-payroll": "/fitur/penggajian",
  unduh: "/unduh",
};

export function marketingPath(slug) {
  return MARKETING_PATHS[slug] || "/";
}

export function gotoMarketing(slug) {
  goto(marketingPath(slug));
}

// Marketing "Coba Gratis" / app entry — logged-in users go to the app, others
// to the register page (mirrors the old marketingGoApp behavior).
export function gotoApp() {
  let target = "/daftar";
  if (browser) {
    try {
      if (localStorage.getItem("user")) target = homePath();
    } catch {}
  }
  goto(target);
}
