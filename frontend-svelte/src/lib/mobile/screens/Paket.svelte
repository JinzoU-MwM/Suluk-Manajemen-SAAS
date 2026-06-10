<script>
  import { onMount } from "svelte";
  import { Calendar, Clock, ChevronRight, Moon } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MChips from "../ui/MChips.svelte";
  import MProgress from "../ui/MProgress.svelte";
  import MBadge from "../ui/MBadge.svelte";

  let { nav } = $props();
  let all = $state([]);
  let loading = $state(true);
  let tab = $state("Semua");

  const COLORS = ["#1B7F5A", "#C99A2E", "#2563a8", "#a9842f", "#7a5ae0"];
  const minPrice = (p) => {
    const tiers = p.pricing_tiers || [];
    const prices = tiers.map((t) => Number(t.price)).filter((n) => n > 0);
    return prices.length ? Math.min(...prices) : Number(p.price ?? p.base_price ?? 0);
  };
  const ptype = (p) => p.package_type || p.tipe || "Umrah";

  onMount(async () => {
    try {
      const res = await ApiService.listPackages({ pageSize: 50 });
      all = res?.packages || res?.data || (Array.isArray(res) ? res : []) || [];
    } catch {
      all = [];
    } finally {
      loading = false;
    }
  });

  let list = $derived(all.filter((p) => tab === "Semua" || ptype(p).toLowerCase().includes(tab.toLowerCase())));
</script>

<MScreen title="Paket Perjalanan" onBack={nav.back}>
  <div style="padding:14px 0 10px"><MChips tabs={["Semua", "Umrah", "Haji"]} value={tab} onChange={(v) => (tab = v)} /></div>

  <div style="padding:0 18px;display:flex;flex-direction:column;gap:14px">
    {#if loading}
      <div class="m-loading" style="padding:50px 0">Memuat…</div>
    {:else if list.length}
      {#each list as p, i (p.id)}
        {@const col = COLORS[i % COLORS.length]}
        <div class="m-card" role="button" tabindex="0" style="overflow:hidden" onclick={() => nav.go("paket-detail", { id: p.id, pkg: p })} onkeydown={(e) => e.key === "Enter" && nav.go("paket-detail", { id: p.id, pkg: p })}>
          <div style="height:76px;background:linear-gradient(120deg,{col},{col}cc);padding:14px;position:relative;display:flex;flex-direction:column;justify-content:flex-end">
            <div style="position:absolute;top:12px;left:14px;color:#fff;opacity:.9;font-size:11px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;display:flex;align-items:center;gap:5px"><Moon size={12} />{ptype(p)}</div>
            {#if p.status}<div style="position:absolute;top:12px;right:14px"><MBadge status={p.status === "published" ? "Published" : p.status} /></div>{/if}
            <div style="color:#fff;font-size:16px;font-weight:800;line-height:1.15;text-shadow:0 1px 6px rgba(0,0,0,.2)">{p.name || p.nama}</div>
          </div>
          <div style="padding:14px">
            <div style="display:flex;gap:14px;font-size:12px;color:var(--c-muted);margin-bottom:12px;flex-wrap:wrap">
              <span style="display:flex;align-items:center;gap:4px"><Calendar size={13} />{p.departure_date || p.tgl || "—"}</span>
              {#if p.duration_days || p.durasi}<span style="display:flex;align-items:center;gap:4px"><Clock size={13} />{p.duration_days || p.durasi} hari</span>{/if}
            </div>
            <div style="display:flex;justify-content:space-between;font-size:11.5px;margin-bottom:5px"><span style="color:var(--c-muted);font-weight:600">Kuota terisi</span><span class="tnum" style="font-weight:700">{(p.reserved_seats ?? p.terisi ?? 0) + "/" + (p.total_seats ?? p.kuota ?? 0)}</span></div>
            <MProgress value={p.reserved_seats ?? p.terisi ?? 0} max={p.total_seats ?? p.kuota ?? 1} color={col} />
            <div style="display:flex;justify-content:space-between;align-items:flex-end;margin-top:12px">
              <div><div style="font-size:10.5px;color:var(--c-faint);font-weight:600">Mulai dari</div><div class="tnum" style="font-size:17px;font-weight:800">{fmtRp(minPrice(p))}</div></div>
              <span style="font-size:12.5px;font-weight:700;color:var(--c-primary);display:flex;align-items:center;gap:2px">Detail<ChevronRight size={15} /></span>
            </div>
          </div>
        </div>
      {/each}
    {:else}
      <div class="m-empty">Belum ada paket</div>
    {/if}
    <div style="height:8px"></div>
  </div>
</MScreen>
