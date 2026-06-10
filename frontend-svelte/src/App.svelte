<script>
  import { onMount } from "svelte";
  import LandingPage from "./lib/pages/LandingPage.svelte";
  import Login from "./lib/pages/Login.svelte";
  import Sidebar from "./lib/components/Sidebar.svelte";
  import Topbar from "./lib/components/Topbar.svelte";
  import MarketingRouter from "./lib/pages/marketing/MarketingRouter.svelte";

  // Public marketing/content routes (hash <-> slug).
  const MARKETING_HASHES = {
    "#/tentang": "about",
    "#/kontak": "contact",
    "#/privasi": "privacy",
    "#/ketentuan": "terms",
    "#/software-travel-umrah": "software",
    "#/fitur/invoice-umrah": "fitur-invoice",
    "#/fitur/crm-jamaah": "fitur-crm",
    "#/fitur/laporan-keuangan": "fitur-keuangan",
    "#/fitur/e-kontrak": "fitur-kontrak",
    "#/fitur/penggajian": "fitur-payroll",
  };
  const SLUG_TO_HASH = {
    about: "#/tentang",
    contact: "#/kontak",
    privacy: "#/privasi",
    terms: "#/ketentuan",
    software: "#/software-travel-umrah",
    "fitur-invoice": "#/fitur/invoice-umrah",
    "fitur-crm": "#/fitur/crm-jamaah",
    "fitur-keuangan": "#/fitur/laporan-keuangan",
    "fitur-kontrak": "#/fitur/e-kontrak",
    "fitur-payroll": "#/fitur/penggajian",
  };
  function marketingSlugForHash(h) {
    if (MARKETING_HASHES[h]) return MARKETING_HASHES[h];
    if (h === "#/404") return "404";
    return null;
  }
  function goMarketing(slug) {
    marketingSlug = slug;
    currentPage = "marketing";
    window.location.hash = SLUG_TO_HASH[slug] || "#/404";
  }
  function marketingGoApp() {
    window.location.hash = "";
    currentPage = user ? "dashboard" : "register";
  }
  function marketingGoHome() {
    window.location.hash = "";
    currentPage = "landing";
  }
  import ProGateScreen from "./lib/components/ProGateScreen.svelte";
  import SupportChatBubble from "./lib/components/SupportChatBubble.svelte";
  import UpgradeModal from "./lib/components/UpgradeModal.svelte";
  import Toast from "./lib/components/Toast.svelte";
  import { ApiService } from "./lib/services/api";
  import { isProOrHigher } from "./lib/config/pricing.js";

  const BASE_URL = "https://jamaah.web.id";
  const DEFAULT_SEO = {
    title: "Software Travel Umrah | Scan KTP/KK ke Siskopatuh - Jamaah.in",
    description:
      "Otomatisasi input data jamaah 10x lebih cepat. Scan KTP/KK langsung jadi Excel Siskopatuh 32 kolom, auto rooming hotel, manifest digital mutawwif.",
    robots: "index,follow",
    canonical: `${BASE_URL}/`,
  };

  const PAGE_SEO = {
    landing: DEFAULT_SEO,
    login: {
      title: "Login - Jamaah.in",
      description: "Login ke dashboard Jamaah.in untuk kelola data jamaah dan operasional umrah.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    register: {
      title: "Daftar - Jamaah.in",
      description: "Daftar akun Jamaah.in untuk mulai otomatisasi data jamaah.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    mutawwif: {
      title: "Manifest Mutawwif - Jamaah.in",
      description: "Manifest jamaah untuk operasional mutawwif.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    registration: {
      title: "Form Pendaftaran Jamaah - Jamaah.in",
      description: "Form pendaftaran jamaah online.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    package: {
      title: "Paket Umrah - Jamaah.in",
      description: "Lihat detail paket umrah, harga per tipe kamar, dan sisa kursi.",
      robots: "index,follow",
      canonical: `${BASE_URL}/`,
    },
    dashboard: {
      title: "Dashboard - Jamaah.in",
      description: "Dashboard operasional jamaah.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    scanner: {
      title: "Scanner Dokumen - Jamaah.in",
      description: "Scan KTP/KK, paspor, dan visa untuk ekstraksi data otomatis.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    jamaah: {
      title: "Data Jamaah - Jamaah.in",
      description: "Kelola data jamaah per grup keberangkatan.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    grup: {
      title: "Grup & Hotel - Jamaah.in",
      description: "Kelola grup keberangkatan dan persiapan hotel.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    manifest: {
      title: "Manifest Digital - Jamaah.in",
      description: "Bagikan manifest digital ber-PIN untuk mutawwif.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    analytics: {
      title: "Analytics - Jamaah.in",
      description: "Statistik operasional jamaah dan grup.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    profile: {
      title: "Profil - Jamaah.in",
      description: "Pengaturan akun Jamaah.in.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    inventory: {
      title: "Inventory - Jamaah.in",
      description: "Manajemen inventory perlengkapan jamaah.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    rooming: {
      title: "Rooming - Jamaah.in",
      description: "Pengaturan rooming jamaah.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    team: {
      title: "Tim - Jamaah.in",
      description: "Manajemen tim operasional.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    itinerary: {
      title: "Itinerary - Jamaah.in",
      description: "Jadwal perjalanan jamaah.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    contracts: {
      title: "E-Kontrak - Jamaah.in",
      description: "Kelola template kontrak digital untuk jamaah.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    "contract-sign": {
      title: "Tanda Tangan Kontrak - Jamaah.in",
      description: "Baca dan tanda tangani kontrak perjalanan secara digital dari HP.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
    "super-admin": {
      title: "Super Admin - Jamaah.in",
      description: "Panel super admin.",
      robots: "noindex,nofollow",
      canonical: `${BASE_URL}/`,
    },
  };

  function upsertMeta(selector, attrs) {
    let el = document.head.querySelector(selector);
    if (!el) {
      el = document.createElement("meta");
      document.head.appendChild(el);
    }
    Object.entries(attrs).forEach(([k, v]) => el.setAttribute(k, v));
  }

  function upsertLinkRel(rel, href) {
    let el = document.head.querySelector(`link[rel="${rel}"]`);
    if (!el) {
      el = document.createElement("link");
      el.setAttribute("rel", rel);
      document.head.appendChild(el);
    }
    el.setAttribute("href", href);
  }

  function updateSeoMeta(page) {
    if (typeof document === "undefined") return;
    const seo = PAGE_SEO[page] || DEFAULT_SEO;
    document.title = seo.title;
    upsertMeta('meta[name="description"]', {
      name: "description",
      content: seo.description,
    });
    upsertMeta('meta[name="robots"]', { name: "robots", content: seo.robots });
    upsertMeta('meta[property="og:title"]', {
      property: "og:title",
      content: seo.title,
    });
    upsertMeta('meta[property="og:description"]', {
      property: "og:description",
      content: seo.description,
    });
    upsertMeta('meta[property="og:url"]', {
      property: "og:url",
      content: seo.canonical,
    });
    upsertMeta('meta[name="twitter:title"]', {
      name: "twitter:title",
      content: seo.title,
    });
    upsertMeta('meta[name="twitter:description"]', {
      name: "twitter:description",
      content: seo.description,
    });
    upsertLinkRel("canonical", seo.canonical);
  }

  // Lazy-loaded page components (reduces initial bundle for mobile/Landing)
  let DashboardPage = $state(null);
  let MobileApp = $state(null);
  let ProfilePage = $state(null);
  let SuperAdminDashboardPage = $state(null);
  let InventoryPage = $state(null);
  let RoomingPage = $state(null);
  let TeamPage = $state(null);
  let ScannerPage = $state(null);
  let DataJamaahPage = $state(null);
  let GroupsPage = $state(null);
  let ManifestPage = $state(null);
  let AnalyticsPage = $state(null);
  let ItineraryPage = $state(null);
  let MutawwifManifestPage = $state(null);
  let PublicRegistrationPage = $state(null);
  let PublicPackagePage = $state(null);
  let PublicContractSigningPage = $state(null);
  // v2 pages
  let PackagesPage = $state(null);
  let CRMPage = $state(null);
  let InvoicesPage = $state(null);
  let FinancePage = $state(null);
  let VendorsPage = $state(null);
  let AgentsPage = $state(null);
  let DocumentsPage = $state(null);
  let ContractsPage = $state(null);
  let CancellationPage = $state(null);
  let PayrollPage = $state(null);
  let ExportPage = $state(null);

  async function ensurePage(page) {
    if (page === "dashboard" && !DashboardPage) {
      DashboardPage = (await import("./lib/pages/Dashboard.svelte")).default;
    } else if (page === "mobile" && !MobileApp) {
      MobileApp = (await import("./lib/mobile/MobileApp.svelte")).default;
    } else if (page === "profile" && !ProfilePage) {
      ProfilePage = (await import("./lib/pages/ProfilePage.svelte")).default;
    } else if (page === "super-admin" && !SuperAdminDashboardPage) {
      SuperAdminDashboardPage = (await import("./lib/pages/SuperAdminDashboard.svelte")).default;
    } else if (page === "inventory" && !InventoryPage) {
      InventoryPage = (await import("./lib/pages/InventoryPage.svelte")).default;
    } else if (page === "rooming" && !RoomingPage) {
      RoomingPage = (await import("./lib/pages/RoomingPage.svelte")).default;
    } else if (page === "team" && !TeamPage) {
      TeamPage = (await import("./lib/pages/TeamPage.svelte")).default;
    } else if (page === "scanner" && !ScannerPage) {
      ScannerPage = (await import("./lib/pages/ScannerPage.svelte")).default;
    } else if (page === "jamaah" && !DataJamaahPage) {
      DataJamaahPage = (await import("./lib/pages/DataJamaahPage.svelte")).default;
    } else if (page === "grup" && !GroupsPage) {
      GroupsPage = (await import("./lib/pages/GroupsPage.svelte")).default;
    } else if (page === "manifest" && !ManifestPage) {
      ManifestPage = (await import("./lib/pages/ManifestPage.svelte")).default;
    } else if (page === "analytics" && !AnalyticsPage) {
      AnalyticsPage = (await import("./lib/pages/AnalyticsPage.svelte")).default;
    } else if (page === "itinerary" && !ItineraryPage) {
      ItineraryPage = (await import("./lib/pages/ItineraryPage.svelte")).default;
    } else if (page === "mutawwif" && !MutawwifManifestPage) {
      MutawwifManifestPage = (await import("./lib/pages/MutawwifManifest.svelte")).default;
    } else if (page === "registration" && !PublicRegistrationPage) {
      PublicRegistrationPage = (await import("./lib/pages/PublicRegistrationPage.svelte")).default;
    } else if (page === "package" && !PublicPackagePage) {
      PublicPackagePage = (await import("./lib/pages/PublicPackagePage.svelte")).default;
    } else if (page === "contract-sign" && !PublicContractSigningPage) {
      PublicContractSigningPage = (await import("./lib/pages/PublicContractSigningPage.svelte")).default;
    } else if (page === "packages" && !PackagesPage) {
      PackagesPage = (await import("./lib/pages/PackagesPage.svelte")).default;
    } else if (page === "crm" && !CRMPage) {
      CRMPage = (await import("./lib/pages/CRMPage.svelte")).default;
    } else if (page === "invoices" && !InvoicesPage) {
      InvoicesPage = (await import("./lib/pages/InvoicesPage.svelte")).default;
    } else if (page === "finance" && !FinancePage) {
      FinancePage = (await import("./lib/pages/FinancePage.svelte")).default;
    } else if (page === "vendors" && !VendorsPage) {
      VendorsPage = (await import("./lib/pages/VendorsPage.svelte")).default;
    } else if (page === "agents" && !AgentsPage) {
      AgentsPage = (await import("./lib/pages/AgentsPage.svelte")).default;
    } else if (page === "documents" && !DocumentsPage) {
      DocumentsPage = (await import("./lib/pages/DocumentsPage.svelte")).default;
    } else if (page === "contracts" && !ContractsPage) {
      ContractsPage = (await import("./lib/pages/ContractsPage.svelte")).default;
    } else if (page === "cancellation" && !CancellationPage) {
      CancellationPage = (await import("./lib/pages/CancellationPage.svelte")).default;
    } else if (page === "payroll" && !PayrollPage) {
      PayrollPage = (await import("./lib/pages/PayrollPage.svelte")).default;
    } else if (page === "export" && !ExportPage) {
      ExportPage = (await import("./lib/pages/ExportPage.svelte")).default;
    } else if (page === "stock" && !InventoryPage) {
      InventoryPage = (await import("./lib/pages/InventoryPage.svelte")).default;
    }
  }

  // Check hash synchronously BEFORE first render to avoid flash of wrong page
  function getInitialPageAndTokens() {
    if (typeof window === "undefined") return { page: "landing", sharedToken: "", registrationToken: "", packageSlug: "", contractToken: "" };
    const hash = window.location.hash;
    const manifestMatch = hash.match(/^#\/m\/([a-f0-9]+)$/i);
    const registrationMatch = hash.match(/^#\/reg\/([a-zA-Z0-9_-]+)$/i);
    const packageMatch = hash.match(/^#\/paket\/([a-zA-Z0-9_-]+)$/i);
    const contractMatch = hash.match(/^#\/kontrak\/([a-zA-Z0-9_-]+)$/i);
    const superAdminMatch = hash === "#/super-admin";
    const mslug = marketingSlugForHash(hash);
    if (mslug) {
      return { page: "marketing", marketingSlug: mslug, sharedToken: "", registrationToken: "", packageSlug: "", contractToken: "" };
    }
    if (superAdminMatch) {
      return { page: "super-admin", sharedToken: "", registrationToken: "", packageSlug: "", contractToken: "" };
    }
    if (hash === "#/app") {
      return { page: "mobile", sharedToken: "", registrationToken: "", packageSlug: "", contractToken: "" };
    }
    if (manifestMatch) {
      return { page: "mutawwif", sharedToken: manifestMatch[1], registrationToken: "", packageSlug: "", contractToken: "" };
    }
    if (registrationMatch) {
      return { page: "registration", sharedToken: "", registrationToken: registrationMatch[1], packageSlug: "", contractToken: "" };
    }
    if (packageMatch) {
      return { page: "package", sharedToken: "", registrationToken: "", packageSlug: packageMatch[1], contractToken: "" };
    }
    if (contractMatch) {
      return { page: "contract-sign", sharedToken: "", registrationToken: "", packageSlug: "", contractToken: contractMatch[1] };
    }
    return { page: "landing", sharedToken: "", registrationToken: "", packageSlug: "", contractToken: "" };
  }
  const initial = getInitialPageAndTokens();

  // Mobile-app mode: when launched at #/app (the Capacitor wrapper / mobile PWA),
  // ALL auth flows must return to the mobile shell ("mobile"), not the desktop
  // "dashboard". Persisted so it survives the login round-trip and reloads; the
  // "Buka Versi Desktop" action and logout clear it.
  function readMobileMode() {
    if (typeof window === "undefined") return false;
    if (initial.page === "mobile") return true;
    try {
      return localStorage.getItem("suluk_mobile") === "1";
    } catch {
      return false;
    }
  }
  let mobileMode = $state(readMobileMode());
  function setMobileMode(on) {
    mobileMode = on;
    try {
      if (on) localStorage.setItem("suluk_mobile", "1");
      else localStorage.removeItem("suluk_mobile");
    } catch {}
  }
  if (mobileMode) setMobileMode(true); // persist on first mobile launch
  // Where to land an authenticated user, honoring mobile mode + super-admin.
  const homePage = () => (mobileMode ? "mobile" : "dashboard");

  // Pages: 'landing' | 'login' | 'register' | 'dashboard' | 'profile' | 'jamaah' | 'grup' | 'scanner' | 'inventory' | 'rooming' | 'manifest' | 'analytics' | 'mutawwif' | 'super-admin' | 'mobile'
  // v2 pages: 'packages' | 'crm' | 'invoices' | 'finance' | 'vendors' | 'agents' | 'documents' | 'contracts' | 'stock' | 'payroll'
  let currentPage = $state(initial.page);
  let user = $state(null);
  let subscription = $state(null);
  let sidebarCollapsed = $state(false);
  let sidebarOpen = $state(false); // mobile drawer
  let marketingSlug = $state(initial.marketingSlug || "");
  let showGlobalUpgradeModal = $state(false);
  let checkingSuperAdminAuth = $state(false); // Loading state for super admin auth check
  let sharedToken = $state(initial.sharedToken); // For /#/m/{token} public manifest
  let registrationToken = $state(initial.registrationToken); // For /#/reg/{token} public registration
  let packageSlug = $state(initial.packageSlug); // For /#/paket/{slug} public package page
  let contractToken = $state(initial.contractToken); // For /#/kontrak/{token} public contract signing

  // Derived
  let isPro = $derived(
    isProOrHigher(subscription?.plan) && subscription?.status !== "expired",
  );
  let isSuperAdmin = $derived(
    user?.is_super_admin === true,
  );
  let trialAvailable = $state(false);

  // Feature gate configs for Pro-only pages
  const proFeatures = {
    inventory: {
      name: "Inventory Checklist",
      desc: "Kelola stok koper, ihram, mukena, dan perlengkapan jamaah. Distribusi tercatat rapi, tidak ada yang terlewat.",
      highlights: [
        "Kelola stok semua item di satu tempat",
        "Forecast kebutuhan otomatis berdasarkan data jamaah",
        "Tracking distribusi per jamaah - siapa sudah terima",
      ],
    },
    rooming: {
      name: "Smart Auto-Rooming",
      desc: "Bagi kamar hotel secara otomatis. Gender dipisahkan, keluarga disatukan - cukup 1 klik.",
      highlights: [
        "Algoritma cerdas mengisi kamar Quad/Triple/Double",
        "Laki-laki & perempuan otomatis dipisahkan",
        "Drag & drop untuk penyesuaian manual",
      ],
    },
    team: {
      name: "Manajemen Tim",
      desc: "Kelola tim Mutawwif dan staf operasional Anda. Bagikan akses manifest digital yang aman.",
      highlights: [
        "Undang anggota tim via email",
        "Atur hak akses per role (Admin / Mutawwif)",
        "Manifest digital aman dengan PIN untuk Mutawwif",
      ],
    },
    manifest: {
      name: "Manifest Digital",
      desc: "Bagikan manifest digital aman untuk mutawwif, lengkap dengan PIN dan link khusus per grup.",
      highlights: [
        "Link manifest ber-PIN untuk setiap grup",
        "Data jamaah dan rooming tampil di HP mutawwif",
        "Akses bisa dibatasi masa berlakunya",
      ],
    },
    analytics: {
      name: "Analytics Operasional",
      desc: "Pantau tren jamaah, kelengkapan perlengkapan, gender, dan performa grup dalam satu tempat.",
      highlights: [
        "Ringkasan data jamaah dan grup",
        "Trend jamaah per bulan",
        "Komposisi gender dan kelengkapan perlengkapan",
      ],
    },
    itinerary: {
      name: "Jadwal Perjalanan",
      desc: "Atur jadwal perjalanan umrah dari hari ke hari. Jamaah dan Mutawwif bisa melihat agenda via manifest digital.",
      highlights: [
        "Buat jadwal harian lengkap dari keberangkatan sampai pulang",
        "Otomatis tampil di manifest digital Mutawwif",
        "Template jadwal untuk reuse di grup berikutnya",
      ],
    },
    contracts: {
      name: "E-Kontrak Digital",
      desc: "Kelola template kontrak digital dan siapkan fondasi penandatanganan jamaah.",
      highlights: [
        "Template kontrak dengan variabel otomatis",
        "Preview isi kontrak sebelum dikirim",
        "Fondasi untuk monitoring tanda tangan",
      ],
    },
  };
  let groups = $state([]);
  let totalJamaahCount = $derived(
    groups.reduce((total, group) => total + Number(group.member_count || 0), 0),
  );

  async function checkSuperAdminAuth() {
    checkingSuperAdminAuth = true;
    const savedUser = localStorage.getItem("user");

    if (savedUser) {
      try {
        user = JSON.parse(savedUser);
      } catch {
        user = null;
      }
    }

    try {
      // Try to get current user and verify super admin status
      const me = await ApiService.getMe();
      user = me;
      localStorage.setItem("user", JSON.stringify(me));

      if (!me.is_super_admin) {
        // Authenticated but not super admin, redirect to dashboard
        checkingSuperAdminAuth = false;
        window.location.hash = "#/dashboard";
        currentPage = "dashboard";
        return;
      }

      // User is super admin, fetch additional data
      const [sub, groupsData, trial] = await Promise.all([
        ApiService.getSubscriptionStatus(),
        ApiService.listGroups(),
        ApiService.getTrialStatus().catch(() => null),
      ]);
      subscription = sub;
      groups = groupsData.groups || [];
      trialAvailable = trial?.trial_available ?? false;
    } catch (err) {
      checkingSuperAdminAuth = false;
      user = null;
      localStorage.removeItem("user");
      window.location.hash = "#/login";
      currentPage = "login";
    } finally {
      checkingSuperAdminAuth = false;
    }
  }

  onMount(async () => {
    // Clean up any leftover dark mode from previous versions
    document.documentElement.classList.remove("dark");
    localStorage.removeItem("darkMode");

    // Public marketing/content pages: route on hash change (back/forward + direct nav).
    window.addEventListener("hashchange", () => {
      const ms = marketingSlugForHash(window.location.hash);
      if (ms) {
        marketingSlug = ms;
        currentPage = "marketing";
      }
    });

    // If already on public page (set by getInitialPageAndTokens), skip auth flow
    if (currentPage === "mutawwif" || currentPage === "registration" || currentPage === "package" || currentPage === "contract-sign" || currentPage === "marketing") {
      return;
    }

    // Special handling for super-admin page: check authentication first
    if (currentPage === "super-admin") {
      await checkSuperAdminAuth();
      return;
    }

    // Listen for hash changes (e.g., user navigates to /#/m/... or /#/reg/... or /#/super-admin)
    window.addEventListener("hashchange", () => {
      const h = window.location.hash;
      const m = h.match(/^#\/m\/([a-f0-9]+)$/i);
      const r = h.match(/^#\/reg\/([a-zA-Z0-9_-]+)$/i);
      const p = h.match(/^#\/paket\/([a-zA-Z0-9_-]+)$/i);
      const ct = h.match(/^#\/kontrak\/([a-zA-Z0-9_-]+)$/i);
      if (h === "#/super-admin") {
        currentPage = "super-admin";
        checkSuperAdminAuth(); // Check auth when navigating to super-admin
      }
      if (h === "#/app") {
        setMobileMode(true);
        ensurePage("mobile");
        currentPage = "mobile";
      }
      if (m) {
        sharedToken = m[1];
        currentPage = "mutawwif";
      }
      if (r) {
        registrationToken = r[1];
        currentPage = "registration";
      }
      if (p) {
        packageSlug = p[1];
        currentPage = "package";
      }
      if (ct) {
        contractToken = ct[1];
        currentPage = "contract-sign";
      }
    });

    // Check for existing cookie session - optimistic navigation from cached profile
    const savedUser = localStorage.getItem("user");
    if (savedUser) {
      try {
        user = JSON.parse(savedUser);
      } catch {
        user = { name: "User", email: "" };
      }
      currentPage = homePage();
    }

    try {
      const me = await ApiService.getMe();
      user = me;
      localStorage.setItem("user", JSON.stringify(me));

      const [sub, groupsData, trial] = await Promise.all([
        ApiService.getSubscriptionStatus().catch(() => null),
        ApiService.listGroups().catch(() => ({ groups: [] })),
        ApiService.getTrialStatus().catch(() => null),
      ]);
      subscription = sub;
      groups = groupsData?.groups || [];
      trialAvailable = trial?.trial_available ?? false;
      currentPage = currentPage === "super-admin" ? "super-admin" : homePage();
    } catch {
      if (savedUser) {
        user = null;
        localStorage.removeItem("user");
      }
      // In mobile mode an unauthenticated launch should show the in-app login
      // (the "mobile" branch renders <Login> when !user), not the marketing site.
      if (mobileMode) {
        currentPage = "mobile";
      } else if (currentPage !== "landing") {
        currentPage = "landing";
      }
    }

    // Handle return from Pakasir payment redirect
    const params = new URLSearchParams(window.location.search);
    if (params.get("payment") === "success" && user) {
      currentPage = homePage();
      // Re-fetch subscription — the Pakasir webhook may have just upgraded the
      // plan, so the status loaded above can be stale. Without this the user
      // returns to a dashboard still showing the old (free/expired) tier.
      try {
        const freshSub = await ApiService.getSubscriptionStatus();
        if (freshSub) subscription = freshSub;
      } catch {}
      // Clean the URL
      window.history.replaceState({}, "", window.location.pathname);
    }
  });

  function handleLoginSuccess(userData) {
    user = userData;
    currentPage = homePage();

    // Fetch subscription and groups
    loadUserData();
  }

  async function loadUserData() {
    try {
      // P0: Parallel fetch all user data
      const [sub, groupsData, trial] = await Promise.all([
        ApiService.getSubscriptionStatus(),
        ApiService.listGroups(),
        ApiService.getTrialStatus().catch(() => null),
      ]);
      subscription = sub;
      groups = groupsData.groups || [];
      trialAvailable = trial?.trial_available ?? false;
    } catch (e) {
      // non-blocking refresh
    }
  }

  async function handleLogout() {
    try {
      await ApiService.logout();
    } catch {
      // no-op
    }
    user = null;
    subscription = null;
    groups = [];
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("user");
    // Stay in the mobile app (show its login) when in mobile mode.
    if (mobileMode) {
      window.location.hash = "#/app";
      currentPage = "mobile";
    } else {
      currentPage = "landing";
    }
  }

  function handlePageChange(page) {
    if (page === "profile:upgrade" || page === "header:upgrade" || page === "trial:activate") {
      // Trigger the global upgrade modal immediately without navigating anywhere
      showGlobalUpgradeModal = true;
      // If we are on landing, login, etc., we probably still want to stay there
      // or we can let the user upgrade from within the dashboard context.
    } else if (page === "super-admin") {
      // Check auth before showing super admin
      currentPage = page;
      checkSuperAdminAuth();
    } else {
      currentPage = page;
    }

    // Refresh data when navigating to certain pages
    if (page === "inventory" || page === "rooming" || page === "jamaah" || page === "grup" || page === "manifest") {
      loadUserData();
    }
  }

  function toggleSidebar() {
    sidebarCollapsed = !sidebarCollapsed;
  }

  // Authenticated layout with sidebar
  const SIDEBAR_PAGES = new Set([
    "dashboard", "jamaah", "grup", "scanner", "profile",
    "inventory", "rooming", "manifest", "analytics", "team", "itinerary",
    // v2
    "packages", "crm", "invoices", "finance", "vendors", "agents", "documents",
    "contracts", "stock", "payroll",
    // v3
    "cancellation", "export",
  ]);

  let showSidebar = $derived(SIDEBAR_PAGES.has(currentPage));

  $effect(() => {
    updateSeoMeta(currentPage);
  });

  // Super admin page (full screen, no sidebar)
  let showSuperAdminPage = $derived(currentPage === "super-admin");
  let showSupportChat = $derived(
    !!user &&
      !showSuperAdminPage &&
      currentPage !== "landing" &&
      currentPage !== "login" &&
      currentPage !== "register" &&
      currentPage !== "mutawwif" &&
      currentPage !== "registration" &&
      currentPage !== "contract-sign",
  );

  $effect(() => {
    // Prefetch only current route and likely next route to reduce render-blocking JS.
    ensurePage(currentPage);
    if (currentPage === "dashboard") {
      ensurePage("scanner");
      ensurePage("jamaah");
      ensurePage("profile");
    }
  });
</script>

<main class="min-h-screen flex flex-col">
  {#if showSuperAdminPage}
    {#if checkingSuperAdminAuth}
      <div class="min-h-screen flex items-center justify-center">
        <div class="text-center">
          <div class="animate-spin rounded-full h-12 w-12 border-4 border-emerald-500 border-t-transparent mx-auto"></div>
          <p class="mt-4 text-slate-600">Verifying access...</p>
        </div>
      </div>
    {:else if user?.is_super_admin}
      {#if SuperAdminDashboardPage}
        <SuperAdminDashboardPage />
      {:else}
        <div class="min-h-screen flex items-center justify-center text-slate-500">Loading dashboard...</div>
      {/if}
    {:else}
      <div class="min-h-screen flex items-center justify-center">
        <div class="text-center">
          <div class="text-6xl mb-4">🔒</div>
          <h1 class="text-2xl font-bold text-red-600 mb-2">Access Denied</h1>
          <p class="text-slate-600">You don't have permission to access this page.</p>
        </div>
      </div>
    {/if}
  {:else if currentPage === "landing"}
    <LandingPage
      onGoToLogin={() => (currentPage = "login")}
      onGoToRegister={() => (currentPage = "register")}
      onNavigate={goMarketing}
    />
  {:else if currentPage === "marketing"}
    <MarketingRouter
      slug={marketingSlug}
      onNav={goMarketing}
      onGoToApp={marketingGoApp}
      onHome={marketingGoHome}
    />
  {:else if currentPage === "login" || currentPage === "register"}
    <Login
      initialMode={currentPage}
      onLoginSuccess={handleLoginSuccess}
      onBack={() => (currentPage = "landing")}
    />
  {:else if currentPage === "mobile"}
    {#if user}
      {#if MobileApp}
        <MobileApp
          {user}
          onExit={() => {
            setMobileMode(false);
            window.location.hash = "#/dashboard";
            currentPage = "dashboard";
            ensurePage("dashboard");
          }}
          onLogout={handleLogout}
        />
      {:else}
        <div class="min-h-screen flex items-center justify-center text-slate-500">Memuat aplikasi…</div>
      {/if}
    {:else}
      <Login initialMode="login" onLoginSuccess={handleLoginSuccess} onBack={() => { setMobileMode(false); currentPage = "landing"; }} />
    {/if}
  {:else if showSidebar}
    <!-- Layout with Sidebar (matches Claude design app shell) -->
    <div class="suluk-shell">
      <Sidebar
        {currentPage}
        onPageChange={handlePageChange}
        {user}
        {isPro}
        open={sidebarOpen}
        onClose={() => (sidebarOpen = false)}
        onUpgrade={() => (showGlobalUpgradeModal = true)}
      />

      <!-- Main column: topbar + scrollable content -->
      <div class="suluk-main">
        <Topbar
          {currentPage}
          {user}
          onMenu={() => (sidebarOpen = !sidebarOpen)}
          onProfile={() => handlePageChange("profile")}
        />
        <div class="suluk-content">
        {#if currentPage === "dashboard"}
          {#if DashboardPage}
            <DashboardPage
              {user}
              {subscription}
              onLogout={handleLogout}
              onSubscriptionChange={loadUserData}
              onNavigate={handlePageChange}
            />
          {:else}
            <div class="p-6 text-slate-500">Loading dashboard...</div>
          {/if}
        {:else if currentPage === "scanner"}
          {#if ScannerPage}
            <ScannerPage
              {user}
              {subscription}
              onLogout={handleLogout}
              onSubscriptionChange={loadUserData}
            />
          {:else}
            <div class="p-6 text-slate-500">Loading scanner...</div>
          {/if}
        {:else if currentPage === "jamaah"}
          {#if DataJamaahPage}
            <DataJamaahPage onNavigate={handlePageChange} />
          {:else}
            <div class="p-6 text-slate-500">Loading data jamaah...</div>
          {/if}
        {:else if currentPage === "grup"}
          {#if GroupsPage}
            <GroupsPage onNavigate={handlePageChange} />
          {:else}
            <div class="p-6 text-slate-500">Loading grup...</div>
          {/if}
        {:else if currentPage === "profile"}
          {#if ProfilePage}
            <ProfilePage
              onLogout={handleLogout}
              {user}
              {subscription}
              {trialAvailable}
              {groups}
              onUpgradeRequest={() => handlePageChange("header:upgrade")}
            />
          {:else}
            <div class="p-6 text-slate-500">Loading profile...</div>
          {/if}
        {:else if currentPage === "inventory"}
          {#if isPro}
            {#if InventoryPage}
              <InventoryPage
                isOpen={true}
                onClose={() => (currentPage = "dashboard")}
                {groups}
                {isPro}
              />
            {:else}
              <div class="p-6 text-slate-500">Loading inventory...</div>
            {/if}
          {:else}
            <ProGateScreen
              featureName={proFeatures.inventory.name}
              featureDescription={proFeatures.inventory.desc}
              highlights={proFeatures.inventory.highlights}
              {trialAvailable}
              onUpgrade={() => handlePageChange("profile:upgrade")}
              onTrial={() => handlePageChange("profile:upgrade")}
            />
          {/if}
        {:else if currentPage === "rooming"}
          {#if isPro}
            {#if RoomingPage}
              <RoomingPage
                isOpen={true}
                onClose={() => (currentPage = "dashboard")}
                {groups}
                {isPro}
              />
            {:else}
              <div class="p-6 text-slate-500">Loading rooming...</div>
            {/if}
          {:else}
            <ProGateScreen
              featureName={proFeatures.rooming.name}
              featureDescription={proFeatures.rooming.desc}
              highlights={proFeatures.rooming.highlights}
              {trialAvailable}
              onUpgrade={() => handlePageChange("profile:upgrade")}
              onTrial={() => handlePageChange("profile:upgrade")}
            />
          {/if}
        {:else if currentPage === "team"}
          {#if isPro}
            {#if TeamPage}
              <TeamPage {isPro} />
            {:else}
              <div class="p-6 text-slate-500">Loading team...</div>
            {/if}
          {:else}
            <ProGateScreen
              featureName={proFeatures.team.name}
              featureDescription={proFeatures.team.desc}
              highlights={proFeatures.team.highlights}
              {trialAvailable}
              onUpgrade={() => handlePageChange("profile:upgrade")}
              onTrial={() => handlePageChange("profile:upgrade")}
            />
          {/if}
        {:else if currentPage === "manifest"}
          {#if isPro}
            {#if ManifestPage}
              <ManifestPage />
            {:else}
              <div class="p-6 text-slate-500">Loading manifest...</div>
            {/if}
          {:else}
            <ProGateScreen
              featureName={proFeatures.manifest.name}
              featureDescription={proFeatures.manifest.desc}
              highlights={proFeatures.manifest.highlights}
              {trialAvailable}
              onUpgrade={() => handlePageChange("profile:upgrade")}
              onTrial={() => handlePageChange("profile:upgrade")}
            />
          {/if}
        {:else if currentPage === "analytics"}
          {#if isPro}
            {#if AnalyticsPage}
              <AnalyticsPage />
            {:else}
              <div class="p-6 text-slate-500">Loading analytics...</div>
            {/if}
          {:else}
            <ProGateScreen
              featureName={proFeatures.analytics.name}
              featureDescription={proFeatures.analytics.desc}
              highlights={proFeatures.analytics.highlights}
              {trialAvailable}
              onUpgrade={() => handlePageChange("profile:upgrade")}
              onTrial={() => handlePageChange("profile:upgrade")}
            />
          {/if}
        {:else if currentPage === "itinerary"}
          {#if isPro}
            {#if ItineraryPage}
              <ItineraryPage {groups} {isPro} />
            {:else}
              <div class="p-6 text-slate-500">Loading itinerary...</div>
            {/if}
          {:else}
            <ProGateScreen
              featureName={proFeatures.itinerary.name}
              featureDescription={proFeatures.itinerary.desc}
              highlights={proFeatures.itinerary.highlights}
              {trialAvailable}
              onUpgrade={() => handlePageChange("profile:upgrade")}
              onTrial={() => handlePageChange("profile:upgrade")}
            />
          {/if}
        {:else if currentPage === "packages"}
          {#if PackagesPage}
            <PackagesPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading paket...</div>
          {/if}
        {:else if currentPage === "crm"}
          {#if CRMPage}
            <CRMPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading CRM...</div>
          {/if}
        {:else if currentPage === "invoices"}
          {#if InvoicesPage}
            <InvoicesPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading invoice...</div>
          {/if}
        {:else if currentPage === "finance"}
          {#if FinancePage}
            <FinancePage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading keuangan...</div>
          {/if}
        {:else if currentPage === "vendors"}
          {#if VendorsPage}
            <VendorsPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading vendor...</div>
          {/if}
        {:else if currentPage === "agents"}
          {#if AgentsPage}
            <AgentsPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading agen...</div>
          {/if}
        {:else if currentPage === "documents"}
          {#if DocumentsPage}
            <DocumentsPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading dokumen...</div>
          {/if}
        {:else if currentPage === "contracts"}
          {#if isPro}
            {#if ContractsPage}
              <ContractsPage {user} />
            {:else}
              <div class="p-6 text-slate-500">Loading e-kontrak...</div>
            {/if}
          {:else}
            <ProGateScreen
              featureName={proFeatures.contracts.name}
              featureDescription={proFeatures.contracts.desc}
              highlights={proFeatures.contracts.highlights}
              {trialAvailable}
              onUpgrade={() => handlePageChange("profile:upgrade")}
              onTrial={() => handlePageChange("profile:upgrade")}
            />
          {/if}
        {:else if currentPage === "cancellation"}
          {#if CancellationPage}
            <CancellationPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading pembatalan...</div>
          {/if}
        {:else if currentPage === "payroll"}
          {#if PayrollPage}
            <PayrollPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading penggajian...</div>
          {/if}
        {:else if currentPage === "export"}
          {#if ExportPage}
            <ExportPage onNavigate={handlePageChange} {user} />
          {:else}
            <div class="p-6 text-slate-500">Loading export...</div>
          {/if}
        {:else if currentPage === "stock"}
          {#if InventoryPage}
            <InventoryPage isOpen={true} onClose={() => handlePageChange("dashboard")} groups={groups} {isPro} />
          {:else}
            <div class="p-6 text-slate-500">Loading inventory...</div>
          {/if}
        {/if}
        </div>
      </div>
    </div>
  {:else if currentPage === "profile"}
    <!-- Fallback profile without sidebar -->
    {#if ProfilePage}
      <ProfilePage
        onLogout={handleLogout}
        {user}
        {subscription}
        {trialAvailable}
        {groups}
        onUpgradeRequest={() => handlePageChange("header:upgrade")}
      />
    {:else}
      <div class="p-6 text-slate-500">Loading profile...</div>
    {/if}
  {:else if currentPage === "mutawwif"}
    <!-- Public Mutawwif Manifest (no auth, no sidebar) -->
    {#if MutawwifManifestPage}
      <MutawwifManifestPage token={sharedToken} />
    {:else}
      <div class="min-h-screen flex items-center justify-center text-slate-500">Loading manifest...</div>
    {/if}
  {:else if currentPage === "registration"}
    <!-- Public Self-Service Registration (no auth, no sidebar) -->
    {#if PublicRegistrationPage}
      <PublicRegistrationPage token={registrationToken} />
    {:else}
      <div class="min-h-screen flex items-center justify-center text-slate-500">Loading registration...</div>
    {/if}
  {:else if currentPage === "package"}
    {#if PublicPackagePage}
      <PublicPackagePage slug={packageSlug} />
    {:else}
      <div class="min-h-screen flex items-center justify-center text-slate-500">Loading package...</div>
    {/if}
  {:else if currentPage === "contract-sign"}
    {#if PublicContractSigningPage}
      <PublicContractSigningPage token={contractToken} />
    {:else}
      <div class="min-h-screen flex items-center justify-center text-slate-500">Loading contract...</div>
    {/if}
  {/if}

  <!-- Global Upgrade Modal -->
  <UpgradeModal
    show={showGlobalUpgradeModal}
    onClose={() => (showGlobalUpgradeModal = false)}
    onSuccess={async (newSub) => {
      subscription = newSub;
      showGlobalUpgradeModal = false;
      await loadUserData();
    }}
  />
  {#if showSupportChat}
    <SupportChatBubble />
  {/if}
</main>
<Toast />
