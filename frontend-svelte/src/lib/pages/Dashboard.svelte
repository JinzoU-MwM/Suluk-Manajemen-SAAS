<script>
  import { onMount } from "svelte";
  import {
    AlertTriangle,
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
    Smartphone,
    TrendingUp,
    UserPlus,
    Users,
    Wallet,
  } from "lucide-svelte";
  import { ApiService } from "../services/api";
  import { formatRupiah, formatDate, formatPct } from "../utils/formatting.js";
  import { isProOrHigher } from "../config/pricing.js";
  import PageHeader from "../components/PageHeader.svelte";
  import StatCard from "../components/StatCard.svelte";
  import Avatar from "../components/Avatar.svelte";
  import Card from "../components/ui/Card.svelte";
  import Button from "../components/ui/Button.svelte";
  import ProgressBar from "../components/ui/ProgressBar.svelte";

  let { user = null, subscription = null, onNavigate = null, onUpgrade = null } = $props();

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
  let genderDonut = $derived(
    `conic-gradient(#1B7F5A 0 ${malePct + 0.4}%, #C99A2E ${malePct + 0.4}% ${malePct + femalePct + 0.4}%, #d7e0db ${malePct + femalePct + 0.4}% 100%)`,
  );
  let mixLegend = $derived([
    { label: "Laki-laki", v: gender.male, c: "#1B7F5A" },
    { label: "Perempuan", v: gender.female, c: "#C99A2E" },
    ...(gender.unknown > 0 ? [{ label: "Belum diisi", v: gender.unknown, c: "#d7e0db" }] : []),
  ]);

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
  const DEP_COLORS = ["#1B7F5A", "#C99A2E", "#2563c9", "#7a5ae0"];

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

{#snippet areaChart(points)}
  {@const n = points.length}
  {@const max = Math.max(...points.map((p) => p.value), 1) * 1.15}
  {@const W = 640}
  {@const H = 210}
  {@const padL = 8}
  {@const padT = 16}
  {@const padB = 28}
  {@const xs = (i) => (n <= 1 ? W / 2 : padL + (i / (n - 1)) * (W - 2 * padL))}
  {@const ys = (v) => padT + (1 - v / max) * (H - padT - padB)}
  {@const line = points.map((p, i) => `${i === 0 ? "M" : "L"}${xs(i).toFixed(1)} ${ys(p.value).toFixed(1)}`).join(" ")}
  {@const area = `${line} L${xs(n - 1).toFixed(1)} ${H - padB} L${xs(0).toFixed(1)} ${H - padB} Z`}
  <div style="position:relative">
    <svg viewBox={`0 0 ${W} ${H}`} style="width:100%;height:{H}px;overflow:visible" role="img" aria-label="Tren chart">
      <defs>
        <linearGradient id="sulukArea" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="var(--c-primary)" stop-opacity="0.22" />
          <stop offset="100%" stop-color="var(--c-primary)" stop-opacity="0" />
        </linearGradient>
      </defs>
      {#each [0.25, 0.5, 0.75, 1] as g}
        <line x1={padL} x2={W - padL} y1={padT + g * (H - padT - padB)} y2={padT + g * (H - padT - padB)} stroke="var(--c-line-soft)" stroke-width="1" />
      {/each}
      <path d={area} fill="url(#sulukArea)" />
      <path d={line} fill="none" stroke="var(--c-primary)" stroke-width="2.5" vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" />
      {#each points as p, i}
        <circle cx={xs(i)} cy={ys(p.value)} r="4" fill="var(--c-surface)" stroke="var(--c-primary)" stroke-width="2.5" />
      {/each}
      {#each points as p, i}
        <text x={xs(i)} y={H - 8} text-anchor="middle" font-size="12" font-weight="600" fill="var(--c-faint)">{p.label}</text>
      {/each}
    </svg>
  </div>
{/snippet}

<div class="page-enter min-h-screen" style="background:var(--c-bg);padding:clamp(16px,2.5vw,32px)">
  <PageHeader
    kicker={`Assalamualaikum, ${displayName}`}
    title="Dashboard"
    subtitle={`Ringkasan operasional travel umrah & haji Anda hari ini, ${today}.`}
  >
    {#snippet actions()}
      <Button variant="ghost" icon={Download} onclick={() => onNavigate?.("export")}>Ekspor</Button>
      <Button variant="primary" icon={Plus} onclick={() => onNavigate?.("scanner")}>Daftarkan Jamaah</Button>
    {/snippet}
  </PageHeader>

  {#if isLoading}
    <Card style="min-height:400px;display:flex;align-items:center;justify-content:center;color:var(--c-faint)">
      <Loader2 class="mr-2 h-6 w-6 animate-spin" style="color:var(--c-primary)" /> Memuat dashboard...
    </Card>
  {:else if error}
    <div style="border-radius:var(--radius-lg);border:1px solid var(--c-danger-soft);background:var(--c-danger-soft);padding:12px 16px;font-size:14px;color:var(--c-danger)">{error}</div>
  {:else}
    <!-- Owner alerts -->
    {#if hasAlerts}
      <div class="mb-5 flex flex-wrap" style="gap:10px">
        {#if alerts.passport_expiring_soon > 0}
          <button type="button" onclick={() => onNavigate?.("documents")} class="inline-flex items-center gap-2" style="border-radius:var(--radius);border:1px solid var(--c-warning-soft);background:var(--c-warning-soft);padding:8px 14px;font-size:12.5px;font-weight:700;color:var(--c-warning)">
            <AlertTriangle class="h-4 w-4" /> {alerts.passport_expiring_soon} paspor segera habis
          </button>
        {/if}
        {#if alerts.incomplete_documents > 0}
          <button type="button" onclick={() => onNavigate?.("documents")} class="inline-flex items-center gap-2" style="border-radius:var(--radius);border:1px solid var(--c-warning-soft);background:var(--c-warning-soft);padding:8px 14px;font-size:12.5px;font-weight:700;color:var(--c-warning)">
            <FileText class="h-4 w-4" /> {alerts.incomplete_documents} dokumen belum lengkap
          </button>
        {/if}
        {#if alerts.overdue_payments > 0}
          <button type="button" onclick={() => onNavigate?.("invoices")} class="inline-flex items-center gap-2" style="border-radius:var(--radius);border:1px solid var(--c-danger-soft);background:var(--c-danger-soft);padding:8px 14px;font-size:12.5px;font-weight:700;color:var(--c-danger)">
            <Receipt class="h-4 w-4" /> {alerts.overdue_payments} pembayaran jatuh tempo
          </button>
        {/if}
      </div>
    {/if}

    <!-- Mobile app banner -->
    <div class="dash-appbar" style="margin-bottom:var(--gap)">
      <div class="dash-appbar-ic"><Smartphone class="h-6 w-6" /></div>
      <div style="flex:1;min-width:0">
        <div style="font-size:15px;font-weight:800;display:flex;align-items:center;gap:8px;flex-wrap:wrap">
          Aplikasi Mobile Suluk
          <span style="font-size:10.5px;font-weight:800;letter-spacing:.06em;background:var(--c-accent-soft);color:var(--c-accent);padding:2px 8px;border-radius:999px">PRO</span>
        </div>
        <div style="font-size:13px;color:var(--c-muted);margin-top:2px">Kelola jamaah, pembayaran & AI Scanner langsung dari ponsel.</div>
      </div>
      {#if isPro}
        <a class="dash-app-btn" href="/unduh">Unduh Aplikasi</a>
      {:else}
        <button type="button" class="dash-app-btn" onclick={() => onUpgrade?.()}>Upgrade ke Pro</button>
      {/if}
    </div>

    <!-- Stat cards -->
    <div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:var(--gap);margin-bottom:var(--gap)">
      {#each statCards as card}
        <StatCard icon={card.icon} label={card.label} value={card.value} accent={card.accent} delta={card.delta} deltaUp={card.deltaUp} sub={card.sub} />
      {/each}
    </div>

    <!-- Main grid: left 1.7fr / right 1fr -->
    <div class="dash-grid" style="display:grid;gap:var(--gap);align-items:start">
      <!-- Left column -->
      <div style="display:flex;flex-direction:column;gap:var(--gap)">
        <!-- Trend / revenue -->
        <Card pad={false}>
          <div style="display:flex;justify-content:space-between;align-items:center;padding:18px 22px 14px;gap:12px">
            <div style="font-size:15.5px;font-weight:800;color:var(--c-ink);white-space:nowrap">{chartSeries.title}</div>
            <div style="display:flex;gap:6px;align-items:center;font-size:12.5px;font-weight:600;color:var(--c-success);background:var(--c-success-soft);padding:5px 10px;border-radius:999px">
              <TrendingUp size={14} /> Live
            </div>
          </div>
          <div style="padding:0 16px 8px">
            <div style="padding:0 8px 10px;display:flex;align-items:baseline;gap:12px">
              <div class="tabular" style="font-size:30px;font-weight:800;letter-spacing:-.02em;white-space:nowrap;color:var(--c-ink)">{chartSeries.total}</div>
              <div style="font-size:13px;color:var(--c-muted)">{chartSeries.totalLabel}</div>
            </div>
            {#if chartSeries.points.length >= 2}
              {@render areaChart(chartSeries.points)}
            {:else}
              <div style="height:210px;display:flex;align-items:center;justify-content:center;border-radius:var(--radius);background:var(--c-bg);font-size:14px;color:var(--c-faint)">Belum cukup data untuk menampilkan tren.</div>
            {/if}
          </div>
        </Card>

        <!-- Quick actions -->
        <Card>
          <div style="font-size:15.5px;font-weight:800;color:var(--c-ink);margin-bottom:14px">Aksi Cepat</div>
          <div style="display:flex;gap:12px;flex-wrap:wrap">
            {#each quickActions as action}
              {@const ActionIcon = action.icon}
              <button
                type="button"
                onclick={() => onNavigate?.(action.page)}
                class="quick-action"
                style="--qa-color:{action.color};display:flex;flex-direction:column;gap:10px;padding:16px 14px;border-radius:var(--radius);background:var(--c-bg);border:1px solid var(--c-line);text-align:left;transition:all .15s;flex:1;min-width:0"
              >
                <span style="width:38px;height:38px;border-radius:var(--radius-sm);background:{action.color}1c;color:{action.color};display:flex;align-items:center;justify-content:center">
                  <ActionIcon size={19} />
                </span>
                <span style="font-size:13px;font-weight:700;color:var(--c-ink)">{action.label}</span>
              </button>
            {/each}
          </div>
        </Card>

        <!-- Departures (owner active packages) -->
        {#if isOwner && activePackages.length > 0}
          <Card pad={false}>
            <div style="display:flex;justify-content:space-between;align-items:center;padding:18px 22px 14px">
              <div style="font-size:15.5px;font-weight:800;color:var(--c-ink);white-space:nowrap">Keberangkatan Mendatang</div>
              <button type="button" onclick={() => onNavigate?.("packages")} style="font-size:12.5px;font-weight:700;color:var(--c-primary)">Kelola paket</button>
            </div>
            <div style="display:flex;flex-direction:column">
              {#each activePackages.slice(0, 4) as pkg, i}
                {@const color = DEP_COLORS[i % DEP_COLORS.length]}
                <div style="display:flex;align-items:center;gap:14px;padding:13px 22px;{i ? 'border-top:1px solid var(--c-line-soft)' : ''}">
                  <div style="width:44px;height:44px;border-radius:var(--radius);background:{color}18;color:{color};display:flex;align-items:center;justify-content:center;flex-shrink:0">
                    <Package size={20} />
                  </div>
                  <div style="flex:1;min-width:0">
                    <div style="font-size:13.5px;font-weight:700;color:var(--c-ink);overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{pkg.name}</div>
                    <div style="font-size:12px;color:var(--c-muted);margin-top:4px;display:flex;align-items:center;gap:6px">
                      <Plane size={12} /> {pkg.reserved_seats ?? 0}/{pkg.total_seats ?? 0} kursi
                    </div>
                  </div>
                  <div style="width:90px">
                    <ProgressBar value={pkg.reserved_seats ?? 0} max={pkg.total_seats || 1} {color} />
                  </div>
                </div>
              {/each}
            </div>
          </Card>
        {/if}
      </div>

      <!-- Right column -->
      <div style="display:flex;flex-direction:column;gap:var(--gap)">
        <!-- Composition donut -->
        <Card pad={false}>
          <div style="display:flex;justify-content:space-between;align-items:center;padding:18px 22px 14px">
            <div style="font-size:15.5px;font-weight:800;color:var(--c-ink);white-space:nowrap">Komposisi Jamaah</div>
          </div>
          {#if genderTotal > 0}
            <div style="padding:6px 22px 22px;display:flex;gap:22px;align-items:center;flex-wrap:wrap">
              <div style="position:relative;display:flex;align-items:center;justify-content:center;width:150px;height:150px;flex-shrink:0;border-radius:50%;background:{genderDonut}">
                <div style="width:102px;height:102px;border-radius:50%;background:var(--c-surface);display:flex;flex-direction:column;align-items:center;justify-content:center;box-shadow:inset 0 0 0 1px var(--c-line-soft)">
                  <span class="tabular" style="font-size:26px;font-weight:800;color:var(--c-ink)">{genderTotal}</span>
                  <span style="font-size:11.5px;color:var(--c-muted);font-weight:600">jamaah</span>
                </div>
              </div>
              <div style="flex:1;min-width:130px;display:flex;flex-direction:column;gap:10px">
                {#each mixLegend as d}
                  <div style="display:flex;align-items:center;gap:9px">
                    <span style="width:10px;height:10px;border-radius:3px;background:{d.c};flex-shrink:0"></span>
                    <span style="font-size:12.5px;color:var(--c-ink-soft);flex:1">{d.label}</span>
                    <span class="tabular" style="font-size:12.5px;font-weight:700;color:var(--c-ink)">{d.v}</span>
                  </div>
                {/each}
              </div>
            </div>
          {:else}
            <div style="display:flex;flex-direction:column;align-items:center;justify-content:center;padding:32px 22px;text-align:center">
              <div style="width:64px;height:64px;border-radius:50%;background:var(--c-bg-2);color:var(--c-faint);display:flex;align-items:center;justify-content:center">
                <Users size={28} />
              </div>
              <p style="margin-top:12px;font-size:12.5px;color:var(--c-faint)">Belum ada data jamaah.</p>
            </div>
          {/if}
        </Card>

        <!-- Recent groups (activity feed) -->
        <Card pad={false}>
          <div style="display:flex;justify-content:space-between;align-items:center;padding:18px 22px 14px">
            <div style="font-size:15.5px;font-weight:800;color:var(--c-ink);white-space:nowrap">Grup Terbaru</div>
            <button type="button" onclick={() => onNavigate?.("jamaah")} style="font-size:12.5px;font-weight:700;color:var(--c-primary)">Lihat semua</button>
          </div>
          {#if recentGroups.length > 0}
            <div style="padding-bottom:8px">
              {#each recentGroups.slice(0, 6) as group}
                <div style="display:flex;gap:13px;padding:13px 22px;align-items:center">
                  <Avatar name={group.name} size={36} />
                  <div style="flex:1;min-width:0">
                    <div style="font-size:13.5px;color:var(--c-ink);line-height:1.4;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"><strong style="font-weight:700">{group.name}</strong></div>
                    <div style="font-size:12.5px;color:var(--c-muted);margin-top:2px">{formatDate(group.created_at)}</div>
                  </div>
                  <span style="font-size:12px;font-weight:700;color:var(--c-primary);background:var(--c-primary-soft);padding:4px 10px;border-radius:999px;white-space:nowrap">{group.member_count} jamaah</span>
                </div>
              {/each}
            </div>
          {:else}
            <div style="padding:8px 22px 24px;text-align:center;font-size:12.5px;color:var(--c-faint)">Belum ada grup keberangkatan.</div>
          {/if}
        </Card>
      </div>
    </div>
  {/if}
</div>

<style>
  .dash-grid {
    grid-template-columns: 1fr;
  }
  @media (min-width: 1024px) {
    .dash-grid {
      grid-template-columns: 1.7fr 1fr;
    }
  }
  .quick-action:hover {
    background: var(--c-surface) !important;
    transform: translateY(-2px);
    box-shadow: var(--shadow);
    border-color: var(--qa-color) !important;
  }
  .tabular {
    font-variant-numeric: tabular-nums;
  }
  .dash-appbar {
    display: flex; align-items: center; gap: 14px;
    background: linear-gradient(120deg, var(--c-primary-deep), var(--c-primary));
    color: #fff; border-radius: var(--radius-lg, 16px); padding: 16px 18px;
  }
  .dash-appbar-ic {
    flex-shrink: 0; width: 44px; height: 44px; border-radius: 12px;
    background: rgba(255,255,255,.16); display: flex; align-items: center; justify-content: center;
  }
  .dash-app-btn {
    flex-shrink: 0; background: #fff; color: var(--c-primary-deep); font-weight: 700; font-size: 13.5px;
    padding: 10px 18px; border-radius: 11px; border: none; cursor: pointer; text-decoration: none; white-space: nowrap;
    transition: transform .12s;
  }
  .dash-app-btn:hover { transform: translateY(-1px); }
  @media (max-width: 640px) {
    .dash-appbar { flex-wrap: wrap; }
    .dash-app-btn { width: 100%; text-align: center; }
  }
</style>
