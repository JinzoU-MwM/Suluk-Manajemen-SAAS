<script>
  import { onMount } from "svelte";
  import { Building2 } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MGroup from "../ui/MGroup.svelte";

  let { nav } = $props();
  let agents = $state([]);
  let loading = $state(true);

  const outstanding = (a) => Number(a.total_outstanding ?? (a.omzet ? a.omzet * a.komisi : 0));

  onMount(async () => {
    try {
      const res = await ApiService.listAgents({ limit: 50 });
      agents = res?.agents || res?.data || (Array.isArray(res) ? res : []) || [];
    } catch {
      agents = [];
    } finally {
      loading = false;
    }
  });

  let active = $derived(agents.filter((a) => (a.status || "Aktif") !== "Nonaktif" && a.is_active !== false).length);
  let totalOwed = $derived(agents.reduce((s, a) => s + outstanding(a), 0));
</script>

<MScreen title="Agen & Mitra" onBack={nav.back}>
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:16px 18px 0">
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800">{active}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Agen aktif</div></div>
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800">{fmtRpShort(totalOwed)}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Komisi terutang</div></div>
  </div>
  <div style="padding:16px 18px 0">
    {#if loading}
      <div class="m-loading" style="padding:50px 0">Memuat…</div>
    {:else if agents.length}
      <MGroup>
        {#each agents as a (a.id)}
          <div class="m-row" role="button" tabindex="0" onclick={() => nav.toast("Detail " + (a.name || "agen"))} onkeydown={() => {}}>
            <div class="m-row-ic" style="background:var(--c-primary-soft);color:var(--c-primary-deep)"><Building2 size={18} /></div>
            <div class="m-row-main">
              <div class="m-row-title">{a.name || a.nama}</div>
              <div class="m-row-sub">{(a.pic_name || a.pic || a.phone || "—") + " · " + (a.total_jamaah ?? a.jamaah ?? 0) + " jamaah"}</div>
            </div>
            <div style="text-align:right;flex-shrink:0">
              <div class="tnum" style="font-size:13px;font-weight:800;color:var(--c-success)">{fmtRpShort(outstanding(a))}</div>
              <div style="font-size:11px;color:var(--c-faint)">komisi</div>
            </div>
          </div>
        {/each}
      </MGroup>
    {:else}
      <div class="m-empty">Belum ada agen</div>
    {/if}
    <div style="height:24px"></div>
  </div>
</MScreen>
