<script>
  import { onMount } from "svelte";
  import {
    AlertTriangle,
    ArrowDown,
    ArrowUp,
    Building2,
    CheckCircle,
    Download,
    FileText,
    Loader2,
    Package,
    Plane,
    Plus,
    Receipt,
    ScanLine,
    TrendingUp,
    UserPlus,
    Users,
    Wallet,
  } from "lucide-svelte";
  import { ApiService } from "../services/api";
  import { formatRupiah, formatDate, formatPct } from "../utils/formatting.js";
  import { isProOrHigher } from "../config/pricing.js";

  let { user = null, subscription = null, onNavigate = null } = $props();

  let stats = $state(null);
  let ownerDash = $state(null);
  let isLoading = $state(true);
  let ownerLoading = $state(false);
  let error = $state("");

  let isOwner = $derived(user?.role === "owner" || user?.role === "admin");
  let isPro = $derived(isProOrHigher(subscription?.plan) && subscription?.status !== "expired");
  let displayName = $derived(user?.name || "Admin Travel");

  const today = new Date().toLocaleDateString("id-ID", {
    weekday: "long",
    day: "numeric",
    month: "long",
    year: "numeric",
  });

  onMount(async () => {
    try {
      stats = await ApiService.getDashboardStats();
    } catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
    }
  });

  $effect(() => {
    if (isOwner && ownerDash === null && !ownerLoading) {
      ownerLoading = true;
      ApiService.getOwnerDashboard()
        .then((d) => { ownerDash = d; })
        .catch(() => {})
        .finally(() => { ownerLoading = false; });
    }
  });

  // ── Gender donut (brand green / gold / slate) ──
  let gender = $derived(stats?.gender_breakdown || { male: 0, female: 0, unknown: 0 });
  let genderTotal = $derived(gender.male + gender.female + gender.unknown);
  let malePct = $derived(genderTotal > 0 ? Math.round((gender.male / genderTotal) * 100) : 0);
  let femalePct = $derived(genderTotal > 0 ? Math.round((gender.female / genderTotal) * 100) : 0);
  let unknownPct = $derived(Math.max(0, 100 - malePct - femalePct));
  let genderDonut = $derived(
    `conic-gradient(#1B7F5A 0 ${malePct + 0.4}%, #C99A2E ${malePct + 0.4}% ${malePct + femalePct + 0.4}%, #d7e0db ${malePct + femalePct + 0.4}% 100%)`,
  );

  // ── Trend series: owner -> revenue; otherwise -> jamaah registrations ──
  let trend = $derived(stats?.monthly_trend || []);
  let revTrend = $derived(ownerDash?.revenue_chart || []);
  let chartSeries = $derived(
    isOwner && revTrend.length
      ? {
          title: "Tren Pendapatan",
          points: revTrend.map((r) => ({ label: r.month, value: Number(r.total ?? 0) })),
          total: formatRupiah(revTrend.reduce((s, r) => s + Number(r.total ?? 0), 0)),
          totalLabel: "total pendapatan",
        }
      : {
          title: "Tren Pendaftaran Jamaah",
          points: trend.map((t) => ({ label: t.label, value: Number(t.count ?? 0) })),
          total: `${trend.reduce((s, t) => s + Number(t.count ?? 0), 0)} jamaah`,
          totalLabel: "6 bulan terakhir",
        },
  );

  let recentGroups = $derived(stats?.recent_groups || []);
  let activePackages = $derived(ownerDash?.active_packages || []);

  let statCards = $derived(
    isOwner && ownerDash
      ? [
          { icon: Users, label: "Total Jamaah Aktif", value: `${stats?.total_jamaah ?? 0}`, accent: "#1B7F5A",
            delta: stats?.jamaah_this_month > 0 ? `${stats.jamaah_this_month}` : null, deltaUp: true,
            sub: stats?.jamaah_this_month > 0 ? `${stats.jamaah_this_month} jamaah baru bulan ini` : "Total jamaah terdaftar" },
          { icon: Wallet, label: "Pendapatan", value: formatRupiah(ownerDash.summary?.total_revenue ?? 0), accent: "#C99A2E",
            sub: "Total terkumpul" },
          { icon: Package, label: "Paket Aktif", value: `${ownerDash.summary?.total_packages ?? 0}`, accent: "#2563c9",
            sub: "Paket berjalan" },
          { icon: AlertTriangle, label: "Tagihan Tertunggak", value: formatRupiah(ownerDash.summary?.total_piutang ?? 0), accent: "#c0392b",
            delta: ownerDash.summary?.overdue_invoices > 0 ? `${ownerDash.summary.overdue_invoices}` : null, deltaUp: false,
            sub: `${ownerDash.summary?.overdue_invoices ?? 0} invoice belum lunas` },
        ]
      : [
          { icon: Users, label: "Total Jamaah", value: `${stats?.total_jamaah ?? 0}`, accent: "#1B7F5A",
            delta: stats?.jamaah_this_month > 0 ? `${stats.jamaah_this_month}` : null, deltaUp: true,
            sub: stats?.jamaah_this_month > 0 ? `${stats.jamaah_this_month} jamaah baru bulan ini` : "Total jamaah terdaftar" },
          { icon: Building2, label: "Total Grup", value: `${stats?.total_groups ?? 0}`, accent: "#2563c9", sub: "Grup keberangkatan" },
          { icon: CheckCircle, label: "Perlengkapan", value: `${formatPct(stats?.equipment_rate ?? 0)}%`, accent: "#C99A2E", sub: "Perlengkapan terpenuhi" },
          { icon: AlertTriangle, label: "Paspor Segera Habis", value: `${stats?.passport_expiring_soon ?? 0}`, accent: "#c0392b", sub: "Berlaku ≤ 90 hari" },
        ],
  );

  const quickActions = [
    { label: "Scan Dokumen", icon: ScanLine, color: "#2563c9", page: "scanner" },
    { label: "Tambah Jamaah", icon: UserPlus, color: "#1B7F5A", page: "crm" },
    { label: "Buat Invoice", icon: Receipt, color: "#C99A2E", page: "invoices" },
    { label: "Paket Baru", icon: Package, color: "#7a5ae0", page: "packages" },
  ];

  let alerts = $derived(ownerDash?.alerts || null);
  let hasAlerts = $derived(
    alerts && ((alerts.passport_expiring_soon ?? 0) + (alerts.incomplete_documents ?? 0) + (alerts.overdue_payments ?? 0) > 0),
  );
</script>

{#snippet statCard(card)}
  {@const Icon = card.icon}
  <div class="rounded-2xl border border-slate-200/70 bg-white p-5 shadow-sm">
    <div class="mb-3.5 flex items-start justify-between">
      <div class="flex h-[42px] w-[42px] items-center justify-center rounded-xl" style="background:{card.accent}18;color:{card.accent}">
        <Icon class="h-5 w-5" />
      </div>
      {#if card.delta != null}
        <span
          class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-bold"
          style={card.deltaUp ? "color:#1B7F5A;background:#E8F4EF" : "color:#c0392b;background:#fbe9e7"}
        >
          {#if card.deltaUp}<ArrowUp class="h-3 w-3" />{:else}<ArrowDown class="h-3 w-3" />{/if}
          {card.delta}
        </span>
      {/if}
    </div>
    <p class="tabular text-[27px] font-extrabold leading-none tracking-tight text-[#10211c]">{card.value}</p>
    <p class="mt-1.5 text-[13.5px] font-medium text-slate-500">{card.label}</p>
    {#if card.sub}<p class="mt-1 text-xs text-slate-400">{card.sub}</p>{/if}
  </div>
{/snippet}

{#snippet areaChart(points)}
  {@const n = points.length}
  {@const max = Math.max(...points.map((p) => p.value), 1)}
  {@const W = 600}
  {@const H = 180}
  {@const pad = 10}
  {@const xs = (i) => (n <= 1 ? W / 2 : (i / (n - 1)) * (W - 2 * pad) + pad)}
  {@const ys = (v) => H - pad - (v / max) * (H - 2 * pad)}
  {@const line = points.map((p, i) => `${i === 0 ? "M" : "L"}${xs(i).toFixed(1)},${ys(p.value).toFixed(1)}`).join(" ")}
  {@const area = `${line} L${xs(n - 1).toFixed(1)},${H} L${xs(0).toFixed(1)},${H} Z`}
  <svg viewBox={`0 0 ${W} ${H}`} class="w-full" style="height:200px" preserveAspectRatio="none" role="img" aria-label="Tren chart">
    <defs>
      <linearGradient id="sulukArea" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="#1B7F5A" stop-opacity="0.22" />
        <stop offset="100%" stop-color="#1B7F5A" stop-opacity="0" />
      </linearGradient>
    </defs>
    <path d={area} fill="url(#sulukArea)" />
    <path d={line} fill="none" stroke="#1B7F5A" stroke-width="2.5" vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
  <div class="mt-1.5 flex justify-between px-1 text-[11px] font-medium text-slate-400">
    {#each points as p}<span class="truncate">{p.label}</span>{/each}
  </div>
{/snippet}

<div class="page-enter min-h-screen bg-[#f6f8f7] p-4 lg:p-8">
  <!-- Page header -->
  <div class="mb-6 flex flex-wrap items-end justify-between gap-4">
    <div>
      <p class="mb-1.5 text-xs font-bold uppercase tracking-[0.08em] text-primary-600">Assalamualaikum, {displayName}</p>
      <h1 class="font-serif text-[26px] font-extrabold tracking-tight text-[#10211c]">Dashboard</h1>
      <p class="mt-1.5 max-w-xl text-sm text-slate-500">Ringkasan operasional travel umrah &amp; haji Anda hari ini, {today}.</p>
    </div>
    <div class="flex items-center gap-2.5">
      <button type="button" onclick={() => onNavigate?.("export")} class="inline-flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 transition-colors hover:bg-slate-50">
        <Download class="h-4 w-4" /> Ekspor
      </button>
      <button type="button" onclick={() => onNavigate?.("scanner")} class="inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm shadow-primary-600/30 transition-colors hover:bg-primary-700">
        <Plus class="h-4 w-4" /> Daftarkan Jamaah
      </button>
    </div>
  </div>

  {#if isLoading}
    <div class="flex min-h-[400px] items-center justify-center rounded-2xl border border-slate-200 bg-white text-slate-400 shadow-sm">
      <Loader2 class="mr-2 h-6 w-6 animate-spin text-primary-500" /> Memuat dashboard...
    </div>
  {:else if error}
    <div class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>
  {:else}
    <!-- Owner alerts -->
    {#if hasAlerts}
      <div class="mb-5 flex flex-wrap gap-2.5">
        {#if alerts.passport_expiring_soon > 0}
          <button type="button" onclick={() => onNavigate?.("documents")} class="inline-flex items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-3.5 py-2 text-xs font-semibold text-amber-800">
            <AlertTriangle class="h-4 w-4 text-amber-600" /> {alerts.passport_expiring_soon} paspor segera habis
          </button>
        {/if}
        {#if alerts.incomplete_documents > 0}
          <button type="button" onclick={() => onNavigate?.("documents")} class="inline-flex items-center gap-2 rounded-xl border border-orange-200 bg-orange-50 px-3.5 py-2 text-xs font-semibold text-orange-800">
            <FileText class="h-4 w-4 text-orange-600" /> {alerts.incomplete_documents} dokumen belum lengkap
          </button>
        {/if}
        {#if alerts.overdue_payments > 0}
          <button type="button" onclick={() => onNavigate?.("invoices")} class="inline-flex items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-3.5 py-2 text-xs font-semibold text-red-800">
            <Receipt class="h-4 w-4 text-red-600" /> {alerts.overdue_payments} pembayaran jatuh tempo
          </button>
        {/if}
      </div>
    {/if}

    <!-- Stat cards -->
    <div class="mb-5 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {#each statCards as card}{@render statCard(card)}{/each}
    </div>

    <!-- Main grid -->
    <div class="grid grid-cols-1 gap-5 xl:grid-cols-[1.7fr_1fr] xl:items-start">
      <!-- Left column -->
      <div class="flex flex-col gap-5">
        <!-- Trend -->
        <section class="rounded-2xl border border-slate-200/70 bg-white p-5 shadow-sm">
          <div class="mb-3 flex items-center justify-between gap-3">
            <h2 class="text-[15.5px] font-extrabold text-[#10211c]">{chartSeries.title}</h2>
            <span class="inline-flex items-center gap-1 rounded-full bg-primary-50 px-2.5 py-1 text-xs font-semibold text-primary-700"><TrendingUp class="h-3.5 w-3.5" /> Live</span>
          </div>
          <div class="mb-1 flex items-baseline gap-3">
            <span class="tabular text-3xl font-extrabold tracking-tight text-[#10211c]">{chartSeries.total}</span>
            <span class="text-[13px] text-slate-500">{chartSeries.totalLabel}</span>
          </div>
          {#if chartSeries.points.length >= 2}
            {@render areaChart(chartSeries.points)}
          {:else}
            <div class="flex h-[200px] items-center justify-center rounded-xl bg-[#f6f8f7] text-sm text-slate-400">Belum cukup data untuk menampilkan tren.</div>
          {/if}
        </section>

        <!-- Quick actions -->
        <section class="rounded-2xl border border-slate-200/70 bg-white p-5 shadow-sm">
          <h2 class="mb-3.5 text-[15.5px] font-extrabold text-[#10211c]">Aksi Cepat</h2>
          <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {#each quickActions as action}
              {@const ActionIcon = action.icon}
              <button type="button" onclick={() => onNavigate?.(action.page)} class="quick-action flex flex-col gap-2.5 rounded-xl border border-slate-200 bg-[#f6f8f7] p-4 text-left transition-all">
                <span class="flex h-[38px] w-[38px] items-center justify-center rounded-lg" style="background:{action.color}1c;color:{action.color}">
                  <ActionIcon class="h-[19px] w-[19px]" />
                </span>
                <span class="text-[13px] font-bold text-[#10211c]">{action.label}</span>
              </button>
            {/each}
          </div>
        </section>

        <!-- Departures (owner) / Recent groups -->
        {#if isOwner && activePackages.length > 0}
          <section class="rounded-2xl border border-slate-200/70 bg-white shadow-sm">
            <div class="flex items-center justify-between px-5 py-4">
              <h2 class="text-[15.5px] font-extrabold text-[#10211c]">Keberangkatan Mendatang</h2>
              <button type="button" onclick={() => onNavigate?.("packages")} class="text-xs font-bold text-primary-600">Kelola paket</button>
            </div>
            <div>
              {#each activePackages.slice(0, 5) as pkg, i}
                <div class="flex items-center gap-3.5 px-5 py-3.5 {i > 0 ? 'border-t border-slate-100' : ''}">
                  <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-primary-50 text-primary-700">
                    <Package class="h-5 w-5" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="truncate text-[13.5px] font-bold text-[#10211c]">{pkg.name}</p>
                    <p class="mt-1 flex items-center gap-1.5 text-xs text-slate-500"><Plane class="h-3 w-3" /> {pkg.reserved_seats ?? 0}/{pkg.total_seats ?? 0} kursi</p>
                  </div>
                  <div class="w-20">
                    <div class="h-1.5 overflow-hidden rounded-full bg-slate-100">
                      <div class="h-full rounded-full bg-primary-500" style:width={`${pkg.total_seats ? Math.min(100, Math.round((pkg.reserved_seats / pkg.total_seats) * 100)) : 0}%`}></div>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </section>
        {/if}
      </div>

      <!-- Right column -->
      <div class="flex flex-col gap-5">
        <!-- Composition donut -->
        <section class="rounded-2xl border border-slate-200/70 bg-white p-5 shadow-sm">
          <h2 class="mb-4 text-[15.5px] font-extrabold text-[#10211c]">Komposisi Jamaah</h2>
          {#if genderTotal > 0}
            <div class="flex items-center gap-5">
              <div class="relative flex h-[130px] w-[130px] flex-shrink-0 items-center justify-center rounded-full" style:background={genderDonut}>
                <div class="flex h-[86px] w-[86px] flex-col items-center justify-center rounded-full bg-white shadow-inner">
                  <span class="tabular text-2xl font-extrabold text-[#10211c]">{genderTotal}</span>
                  <span class="text-[11px] font-semibold text-slate-400">jamaah</span>
                </div>
              </div>
              <div class="flex flex-1 flex-col gap-2.5">
                <div class="flex items-center gap-2.5 text-sm">
                  <span class="h-2.5 w-2.5 flex-shrink-0 rounded-[3px] bg-primary-600"></span>
                  <span class="flex-1 text-slate-600">Laki-laki</span>
                  <span class="tabular font-bold text-slate-800">{gender.male}</span>
                </div>
                <div class="flex items-center gap-2.5 text-sm">
                  <span class="h-2.5 w-2.5 flex-shrink-0 rounded-[3px] bg-gold-500"></span>
                  <span class="flex-1 text-slate-600">Perempuan</span>
                  <span class="tabular font-bold text-slate-800">{gender.female}</span>
                </div>
                {#if gender.unknown > 0}
                  <div class="flex items-center gap-2.5 text-sm">
                    <span class="h-2.5 w-2.5 flex-shrink-0 rounded-[3px] bg-slate-300"></span>
                    <span class="flex-1 text-slate-600">Belum diisi</span>
                    <span class="tabular font-bold text-slate-800">{gender.unknown}</span>
                  </div>
                {/if}
              </div>
            </div>
          {:else}
            <div class="flex flex-col items-center justify-center py-8 text-center">
              <div class="flex h-16 w-16 items-center justify-center rounded-full border-2 border-dashed border-slate-200 bg-slate-50">
                <Users class="h-7 w-7 text-slate-300" />
              </div>
              <p class="mt-3 text-xs text-slate-400">Belum ada data jamaah.</p>
            </div>
          {/if}
        </section>

        <!-- Recent groups -->
        <section class="rounded-2xl border border-slate-200/70 bg-white shadow-sm">
          <div class="flex items-center justify-between px-5 py-4">
            <h2 class="text-[15.5px] font-extrabold text-[#10211c]">Grup Terbaru</h2>
            <button type="button" onclick={() => onNavigate?.("jamaah")} class="text-xs font-bold text-primary-600">Lihat semua</button>
          </div>
          {#if recentGroups.length > 0}
            <div>
              {#each recentGroups.slice(0, 6) as group, i}
                <div class="flex items-center gap-3 px-5 py-3 {i > 0 ? 'border-t border-slate-100' : ''}">
                  <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-primary-50 text-primary-700">
                    <Building2 class="h-4 w-4" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="truncate text-[13.5px] font-semibold text-[#10211c]">{group.name}</p>
                    <p class="text-xs text-slate-400">{formatDate(group.created_at)}</p>
                  </div>
                  <span class="rounded-full bg-primary-50 px-2 py-0.5 text-[11px] font-bold text-primary-700">{group.member_count} jamaah</span>
                </div>
              {/each}
            </div>
          {:else}
            <div class="px-5 pb-6 pt-2 text-center text-xs text-slate-400">Belum ada grup keberangkatan.</div>
          {/if}
        </section>
      </div>
    </div>
  {/if}
</div>

<style>
  .quick-action:hover {
    background: #ffffff;
    transform: translateY(-2px);
    box-shadow: 0 4px 14px rgba(16, 33, 28, 0.07);
    border-color: #cdd6d2;
  }
  .tabular { font-variant-numeric: tabular-nums; }
</style>
