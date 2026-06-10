<script>
  import { onMount } from "svelte";
  import { Users } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MScreen from "../ui/MScreen.svelte";

  let { nav } = $props();
  let groups = $state([]);
  let loading = $state(true);
  const COLORS = ["#a9842f", "#C99A2E", "#1B7F5A", "#2563a8", "#7a5ae0"];

  onMount(async () => {
    try {
      const res = await ApiService.listGroups();
      groups = res?.groups || res?.data || (Array.isArray(res) ? res : []) || [];
    } catch {
      groups = [];
    } finally {
      loading = false;
    }
  });
</script>

<MScreen title="Grup Keberangkatan" onBack={nav.back}>
  <div style="padding:16px 18px 0;display:flex;flex-direction:column;gap:12px">
    {#if loading}
      <div class="m-loading" style="padding:50px 0">Memuat…</div>
    {:else if groups.length}
      {#each groups as g, i (g.id)}
        {@const col = COLORS[i % COLORS.length]}
        <div class="m-card" role="button" tabindex="0" style="overflow:hidden" onclick={() => nav.toast("Detail " + (g.name || "grup"))} onkeydown={() => {}}>
          <div style="height:5px;background:{col}"></div>
          <div style="padding:16px">
            <div style="display:flex;justify-content:space-between;align-items:flex-start;gap:12px">
              <div style="flex:1;min-width:0">
                <div style="font-size:16px;font-weight:800">{g.name || g.nama || "Grup"}</div>
                <div style="font-size:12.5px;color:var(--c-muted);margin-top:2px">{g.package_name || g.paket || "—"}</div>
              </div>
              <div style="width:42px;height:42px;border-radius:12px;background:{col}18;color:{col};display:flex;align-items:center;justify-content:center;flex-shrink:0"><Users size={20} /></div>
            </div>
            <div style="display:flex;gap:22px;margin-top:14px;padding-top:14px;border-top:1px solid var(--c-line-soft)">
              <div><div class="tnum" style="font-size:18px;font-weight:800">{g.member_count ?? g.jamaah_count ?? 0}</div><div style="font-size:11px;color:var(--c-faint)">jamaah</div></div>
              {#if g.departure_date || g.tgl}<div><div style="font-size:13.5px;font-weight:700;margin-top:3px">{g.departure_date || g.tgl}</div><div style="font-size:11px;color:var(--c-faint)">berangkat</div></div>{/if}
              {#if g.mutawwif}<div style="flex:1;text-align:right"><div style="font-size:13px;font-weight:700;margin-top:3px">{String(g.mutawwif).replace("Ust. ", "")}</div><div style="font-size:11px;color:var(--c-faint)">muthawwif</div></div>{/if}
            </div>
          </div>
        </div>
      {/each}
    {:else}
      <div class="m-empty">Belum ada grup</div>
    {/if}
  </div>
</MScreen>
