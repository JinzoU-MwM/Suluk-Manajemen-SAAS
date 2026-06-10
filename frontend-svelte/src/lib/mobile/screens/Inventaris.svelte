<script>
  import { onMount } from "svelte";
  import { Boxes } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MScreen from "../ui/MScreen.svelte";
  import MChips from "../ui/MChips.svelte";
  import MGroup from "../ui/MGroup.svelte";

  let { nav } = $props();
  let groups = $state([]);
  let groupId = $state(null);
  let forecast = $state(null);
  let loading = $state(true);
  let loadingItems = $state(false);

  onMount(async () => {
    try {
      const res = await ApiService.listGroups();
      groups = res?.groups || res?.data || (Array.isArray(res) ? res : []) || [];
      if (groups.length) {
        groupId = groups[0].id;
        await loadForecast();
      }
    } catch {}
    finally {
      loading = false;
    }
  });

  async function loadForecast() {
    loadingItems = true;
    forecast = null;
    try {
      forecast = await ApiService.getInventoryForecast(groupId);
    } finally {
      loadingItems = false;
    }
  }
  function pick(id) {
    groupId = id;
    loadForecast();
  }
  let chips = $derived(groups.map((g) => ({ value: g.id, label: g.name || g.nama || "Grup" })));
  let reqs = $derived(forecast?.requirements || forecast?.details || forecast?.items || []);
  const itemName = (r) => r.item || r.name || r.nama || r.item_name || "Item";
  const itemQty = (r) => r.quantity ?? r.qty ?? r.required ?? r.total ?? 0;
  const itemUnit = (r) => r.unit || r.satuan || "pcs";
</script>

<MScreen title="Inventaris" onBack={nav.back}>
  {#if loading}
    <div class="m-loading" style="padding:50px 0">Memuat…</div>
  {:else if !groups.length}
    <div class="m-empty" style="padding:50px 20px">Belum ada grup</div>
  {:else}
    <div style="padding:14px 0 6px"><MChips tabs={chips} value={groupId} onChange={pick} /></div>
    {#if forecast?.total_members != null}
      <div style="padding:6px 18px 0">
        <div class="m-card m-card-pad" style="background:var(--c-primary-tint);border:none">
          <div style="font-size:13.5px;font-weight:700">Kebutuhan Perlengkapan</div>
          <div style="font-size:12.5px;color:var(--c-muted);margin-top:2px">{forecast.total_members} jamaah · {reqs.length} jenis item</div>
        </div>
      </div>
    {/if}

    <div style="padding:16px 18px 0">
      {#if loadingItems}
        <div class="m-loading" style="padding:40px 0">Memuat…</div>
      {:else if reqs.length}
        <MGroup>
          {#each reqs as r, i (i)}
            <div class="m-row">
              <div class="m-row-ic" style="background:var(--c-bg-2);color:var(--c-ink-soft)"><Boxes size={18} /></div>
              <div class="m-row-main"><div class="m-row-title">{itemName(r)}</div>{#if r.category || r.kategori}<div class="m-row-sub">{r.category || r.kategori}</div>{/if}</div>
              <div style="text-align:right;flex-shrink:0"><div class="tnum" style="font-size:15px;font-weight:800">{itemQty(r)}</div><div style="font-size:11px;color:var(--c-faint)">{itemUnit(r)}</div></div>
            </div>
          {/each}
        </MGroup>
      {:else}
        <div class="m-empty" style="padding:30px 20px">Belum ada kebutuhan tercatat</div>
      {/if}
      <div style="height:24px"></div>
    </div>
  {/if}
</MScreen>
