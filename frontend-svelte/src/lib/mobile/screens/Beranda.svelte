<script>
  import { onMount } from "svelte";
  import { Bell, ArrowUp, ScanLine, Wallet, UserPlus, CircleCheck, Users, Package, Clock, ChevronRight } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRpShort } from "../format.js";
  import MCard from "../ui/MCard.svelte";
  import MSection from "../ui/MSection.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MSpark from "../ui/MSpark.svelte";
  import MCountUp from "../ui/MCountUp.svelte";

  let { nav } = $props();

  let stats = $state(null);
  let owner = $state(null);
  let packages = $state([]);
  let pendingApprovals = $state(0);

  const MONTHS = ["Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"];
  function dateChip(s) {
    if (!s) return { d: "--", m: "" };
    const t = new Date(s);
    if (isNaN(t.getTime())) return { d: String(s).slice(0, 2), m: "" };
    return { d: String(t.getDate()), m: MONTHS[t.getMonth()] };
  }

  onMount(async () => {
    const [st, ow, pk] = await Promise.all([
      ApiService.getDashboardStats().catch(() => null),
      ApiService.getOwnerDashboard().catch(() => null),
      ApiService.listPackages({ status: "published", pageSize: 5 }).catch(() => null),
    ]);
    stats = st;
    owner = ow;
    const list = pk?.packages || pk?.data || pk || [];
    packages = (Array.isArray(list) ? list : []).slice(0, 3);
    try {
      const rf = await ApiService.listRefunds({ status: "pending", limit: 50 });
      const items = rf?.data || rf?.refunds || rf || [];
      pendingApprovals = Array.isArray(items) ? items.length : 0;
    } catch {
      pendingApprovals = 0;
    }
  });

  let greeting = $derived(nav.user?.name || "Admin");
  let revenue = $derived(owner?.summary?.total_revenue ?? 0);
  let spark = $derived((owner?.revenue_chart || []).map((d) => Number(d.revenue ?? d.value ?? d.v ?? 0)));
  let kpis = $derived([
    { ic: Users, val: stats?.total_jamaah ?? 0, fmt: (v) => Math.round(v).toLocaleString("id-ID"), l: "Jamaah aktif", c: "#1B7F5A" },
    { ic: Package, val: owner?.summary?.total_packages ?? 0, fmt: (v) => Math.round(v), l: "Paket aktif", c: "#2563a8" },
    { ic: Clock, val: owner?.summary?.total_piutang ?? 0, fmt: (v) => fmtRpShort(v), l: "Tertunggak", c: "#b8860b" },
    { ic: CircleCheck, val: pendingApprovals, fmt: (v) => Math.round(v), l: "Perlu approval", c: "#7a5ae0" },
  ]);
  const acts = [
    { ic: ScanLine, label: "Scan", color: "#2563a8", go: () => nav.switchTab("scan") },
    { ic: Wallet, label: "Bayar", color: "#1B7F5A", go: () => nav.go("bayar") },
    { ic: UserPlus, label: "Jamaah", color: "#C99A2E", go: () => nav.switchTab("jamaah") },
    { ic: CircleCheck, label: "Approval", color: "#7a5ae0", go: () => nav.go("approval") },
  ];
</script>

<div class="m-screen-root">
  <div class="m-hero m-hero-in">
    <div style="display:flex;justify-content:space-between;align-items:flex-start;position:relative">
      <div>
        <div style="font-size:13px;opacity:.8">Assalamualaikum,</div>
        <div style="font-size:21px;font-weight:800;margin-top:2px">{greeting}</div>
      </div>
      <button
        type="button"
        class="m-bell"
        onclick={() => nav.go("notifikasi")}
        style="position:relative;width:42px;height:42px;border-radius:13px;background:rgba(255,255,255,.14);display:flex;align-items:center;justify-content:center;color:#fff"
      >
        <Bell size={20} />
        <span style="position:absolute;top:9px;right:10px;width:8px;height:8px;border-radius:999px;background:#C99A2E;border:2px solid var(--c-primary-deep)"></span>
      </button>
    </div>
    <div style="margin-top:20px;position:relative">
      <div style="font-size:12.5px;opacity:.8;font-weight:600">Pendapatan bulan ini</div>
      <div style="font-size:34px;font-weight:800;letter-spacing:-.02em;margin-top:2px">
        <MCountUp to={revenue} dur={1100} format={(v) => fmtRpShort(v)} />
      </div>
      {#if spark.length}
        <div style="margin-top:14px"><MSpark data={spark} height={44} /></div>
      {/if}
    </div>
  </div>

  <div class="m-scroll m-stagger">
    <!-- KPI cards -->
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:16px 18px 4px">
      {#each kpis as s}
        {@const Icon = s.ic}
        <MCard pad>
          <div style="display:flex;align-items:center;gap:8px">
            <div style="width:32px;height:32px;border-radius:9px;background:{s.c}1c;color:{s.c};display:flex;align-items:center;justify-content:center">
              <Icon size={17} />
            </div>
          </div>
          <div style="font-size:21px;font-weight:800;margin-top:10px;letter-spacing:-.02em">
            <MCountUp to={s.val} dur={900} format={s.fmt} />
          </div>
          <div style="font-size:12px;color:var(--c-muted);margin-top:1px">{s.l}</div>
        </MCard>
      {/each}
    </div>

    <!-- Quick actions -->
    <MSection label="Aksi Cepat" style="padding-top:18px">
      <div style="display:flex;gap:10px">
        {#each acts as a}
          {@const Icon = a.ic}
          <button type="button" onclick={a.go} class="m-qa" style="flex:1;background:var(--c-surface);border:1px solid var(--c-line);border-radius:16px;padding:14px 6px;display:flex;flex-direction:column;align-items:center;gap:8px">
            <div style="width:42px;height:42px;border-radius:12px;background:{a.color}1c;color:{a.color};display:flex;align-items:center;justify-content:center">
              <Icon size={20} />
            </div>
            <span style="font-size:11.5px;font-weight:600;color:var(--c-ink)">{a.label}</span>
          </button>
        {/each}
      </div>
    </MSection>

    <!-- Departures -->
    <MSection label="Keberangkatan Mendatang" style="padding-top:22px">
      {#snippet action()}
        <button type="button" onclick={() => nav.switchTab("lainnya")} style="font-size:12.5px;font-weight:700;color:var(--c-primary)">Semua</button>
      {/snippet}
      <MGroup>
        {#if packages.length}
          {#each packages as p}
            {@const dc = dateChip(p.departure_date)}
            <div class="m-row">
              <div style="width:44px;height:44px;border-radius:12px;background:#1B7F5A18;color:#1B7F5A;display:flex;flex-direction:column;align-items:center;justify-content:center;flex-shrink:0">
                <span style="font-size:15px;font-weight:800;line-height:1">{dc.d}</span>
                <span style="font-size:9px;font-weight:700;text-transform:uppercase">{dc.m}</span>
              </div>
              <div class="m-row-main">
                <div class="m-row-title">{p.name || p.nama || "Paket"}</div>
                <div class="m-row-sub">{(p.airline || p.maskapai || "") + " · " + (p.reserved_seats ?? 0) + "/" + (p.total_seats ?? 0) + " kursi"}</div>
              </div>
              <ChevronRight size={18} class="m-chev" />
            </div>
          {/each}
        {:else}
          <div class="m-empty" style="padding:28px 20px">Belum ada keberangkatan</div>
        {/if}
      </MGroup>
    </MSection>

    <div style="height:24px"></div>
  </div>
</div>
