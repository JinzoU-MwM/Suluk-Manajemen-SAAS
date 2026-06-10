<script>
  import { onMount } from "svelte";
  import { Download } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MSection from "../ui/MSection.svelte";
  import MProgress from "../ui/MProgress.svelte";

  let { nav } = $props();
  let owner = $state(null);
  let expenses = $state([]); // [{label, value, color}]
  let loading = $state(true);

  const EXP_COLORS = ["#2563a8", "#C99A2E", "#7a5ae0", "#1B7F5A", "#a9842f", "#15564a"];

  onMount(async () => {
    try {
      const [ow, debt] = await Promise.all([ApiService.getOwnerDashboard().catch(() => null), ApiService.getDebtSummary().catch(() => null)]);
      owner = ow;
      const rows = debt?.by_category || debt?.categories || [];
      if (Array.isArray(rows) && rows.length) {
        expenses = rows.map((r, i) => ({ label: r.category || r.kategori || r.label || "Lainnya", value: Number(r.total ?? r.amount ?? r.outstanding ?? 0), color: EXP_COLORS[i % EXP_COLORS.length] }));
      }
    } catch {}
    finally {
      loading = false;
    }
  });

  let revenue = $derived(Number(owner?.summary?.total_revenue ?? 0));
  let grossProfit = $derived(Number(owner?.summary?.gross_profit_month ?? owner?.summary?.gross_profit ?? 0));
  let cost = $derived(Math.max(0, revenue - grossProfit));
  let margin = $derived(revenue > 0 ? ((grossProfit / revenue) * 100).toFixed(1).replace(".", ",") : "0");
  let maxExp = $derived(Math.max(1, ...expenses.map((e) => e.value)));
</script>

<MScreen title="Keuangan" onBack={nav.back}>
  <div style="padding:16px 18px 0">
    <div class="m-card m-card-pad" style="background:linear-gradient(150deg,var(--c-primary-deep),var(--c-primary));border:none;color:#fff">
      <div style="font-size:12.5px;opacity:.85;font-weight:600">Laba kotor bulan ini</div>
      <div class="tnum" style="font-size:30px;font-weight:800;margin-top:4px">{loading ? "…" : fmtRpShort(grossProfit)}</div>
      <div style="display:flex;gap:18px;margin-top:14px">
        <div><div style="font-size:11px;opacity:.8">Pendapatan</div><div class="tnum" style="font-size:15px;font-weight:700">{fmtRpShort(revenue)}</div></div>
        <div><div style="font-size:11px;opacity:.8">Biaya</div><div class="tnum" style="font-size:15px;font-weight:700">{fmtRpShort(cost)}</div></div>
        <div><div style="font-size:11px;opacity:.8">Margin</div><div class="tnum" style="font-size:15px;font-weight:700">{margin}%</div></div>
      </div>
    </div>
  </div>

  {#if expenses.length}
    <MSection label="Utang Vendor per Kategori" style="padding-top:20px">
      <div class="m-card m-card-pad" style="display:flex;flex-direction:column;gap:13px">
        {#each expenses as e (e.label)}
          <div>
            <div style="display:flex;justify-content:space-between;font-size:13px;margin-bottom:5px"><span style="color:var(--c-ink-soft);font-weight:600">{e.label}</span><span class="tnum" style="font-weight:700">{fmtRpShort(e.value)}</span></div>
            <MProgress value={e.value} max={maxExp} color={e.color} />
          </div>
        {/each}
      </div>
    </MSection>
  {/if}

  <div style="padding:20px 18px 0">
    <button type="button" class="m-btn m-btn-ghost" onclick={() => nav.toast("Laporan keuangan — buka versi desktop untuk unduh")}><Download size={18} />Unduh Laporan</button>
  </div>
</MScreen>
