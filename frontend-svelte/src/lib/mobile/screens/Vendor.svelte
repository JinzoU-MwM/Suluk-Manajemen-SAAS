<script>
  import { onMount } from "svelte";
  import { Plane, Building2, Globe, Heart, Truck, FileText } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MBadge from "../ui/MBadge.svelte";

  let { nav } = $props();
  let vendors = $state([]);
  let debt = $state({});
  let loading = $state(true);

  const ICON = { Maskapai: Plane, maskapai: Plane, Hotel: Building2, hotel: Building2, Muassasah: Globe, Katering: Heart, katering: Heart, Transportasi: Truck, transportasi: Truck, Visa: FileText, visa: FileText };
  const debtOf = (v) => Number(debt[v.id] ?? v.utang ?? v.outstanding ?? 0);

  onMount(async () => {
    try {
      const [vs, ds] = await Promise.all([ApiService.listVendors({ pageSize: 50 }), ApiService.getDebtSummary().catch(() => null)]);
      vendors = vs?.vendors || vs?.data || (Array.isArray(vs) ? vs : []) || [];
      // build per-vendor debt map if the summary returns one
      const rows = ds?.by_vendor || ds?.vendors || [];
      if (Array.isArray(rows)) for (const r of rows) debt[r.vendor_id || r.id] = r.outstanding ?? r.utang ?? r.total;
    } catch {
      vendors = [];
    } finally {
      loading = false;
    }
  });

  let totalDebt = $derived(vendors.reduce((s, v) => s + debtOf(v), 0));
</script>

<MScreen title="Vendor & Pemasok" onBack={nav.back}>
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:16px 18px 0">
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800">{vendors.length}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Total vendor</div></div>
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800">{fmtRpShort(totalDebt)}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Total utang</div></div>
  </div>
  <div style="padding:16px 18px 0">
    {#if loading}
      <div class="m-loading" style="padding:50px 0">Memuat…</div>
    {:else if vendors.length}
      <MGroup>
        {#each vendors as v (v.id)}
          {@const Icon = ICON[v.type] || ICON[v.kategori] || Building2}
          {@const d = debtOf(v)}
          <div class="m-row" role="button" tabindex="0" onclick={() => nav.toast("Detail " + (v.name || "vendor"))} onkeydown={() => {}}>
            <div class="m-row-ic" style="background:var(--c-accent-soft);color:var(--c-accent)"><Icon size={18} /></div>
            <div class="m-row-main">
              <div class="m-row-title">{v.name || v.nama}</div>
              <div class="m-row-sub">{(v.type || v.kategori || "Vendor") + (v.city || v.kota ? " · " + (v.city || v.kota) : "")}</div>
            </div>
            <div style="text-align:right;flex-shrink:0">
              {#if d > 0}<div class="tnum" style="font-size:13px;font-weight:700;color:var(--c-danger)">{fmtRpShort(d)}</div>{:else}<MBadge status="Aktif" />{/if}
            </div>
          </div>
        {/each}
      </MGroup>
    {:else}
      <div class="m-empty">Belum ada vendor</div>
    {/if}
    <div style="height:24px"></div>
  </div>
</MScreen>
