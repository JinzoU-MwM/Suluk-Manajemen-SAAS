<script>
  import { onMount } from "svelte";
  import { MapPin } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MScreen from "../ui/MScreen.svelte";
  import MChips from "../ui/MChips.svelte";

  let { nav } = $props();
  let groups = $state([]);
  let groupId = $state(null);
  let items = $state([]);
  let loading = $state(true);
  let loadingItems = $state(false);

  onMount(async () => {
    try {
      const res = await ApiService.listGroups();
      groups = res?.groups || res?.data || (Array.isArray(res) ? res : []) || [];
      if (groups.length) {
        groupId = groups[0].id;
        await loadItems();
      }
    } catch {}
    finally {
      loading = false;
    }
  });

  async function loadItems() {
    loadingItems = true;
    items = [];
    try {
      const res = await ApiService.getItinerary(groupId);
      const list = res?.items || res?.data || res?.itinerary || (Array.isArray(res) ? res : []) || [];
      items = Array.isArray(list) ? list : [];
    } finally {
      loadingItems = false;
    }
  }
  function pick(id) {
    groupId = id;
    loadItems();
  }
  const dayLabel = (it, i) => it.day_label || (it.day_number ? "Hari " + it.day_number : it.day ? "Hari " + it.day : "Hari " + (i + 1));
  const timeRange = (it) => [it.time_start, it.time_end].filter(Boolean).join("–");
  let chips = $derived(groups.map((g) => ({ value: g.id, label: g.name || g.nama || "Grup" })));
</script>

<MScreen title="Itinerary" onBack={nav.back}>
  {#if loading}
    <div class="m-loading" style="padding:50px 0">Memuat…</div>
  {:else if !groups.length}
    <div class="m-empty" style="padding:50px 20px">Belum ada grup</div>
  {:else}
    <div style="padding:14px 0 10px"><MChips tabs={chips} value={groupId} onChange={pick} /></div>

    {#if loadingItems}
      <div class="m-loading" style="padding:40px 0">Memuat…</div>
    {:else if items.length}
      <div style="padding:4px 18px 0 26px;position:relative">
        {#each items as it, i (it.id ?? i)}
          <div style="display:flex;gap:16px;padding-bottom:{i < items.length - 1 ? 22 : 0}px;position:relative">
            {#if i < items.length - 1}<div style="position:absolute;left:19px;top:42px;bottom:0;width:2px;background:var(--c-line)"></div>{/if}
            <div style="width:40px;height:40px;border-radius:50%;background:var(--c-primary-soft);color:var(--c-primary-deep);display:flex;align-items:center;justify-content:center;flex-shrink:0;z-index:1;border:3px solid var(--c-bg)"><MapPin size={18} /></div>
            <div style="flex:1;padding-top:1px">
              <div style="font-size:11px;font-weight:700;color:var(--c-accent);text-transform:uppercase;letter-spacing:.04em">{dayLabel(it, i)}{timeRange(it) ? " · " + timeRange(it) : ""}</div>
              <div style="font-size:15px;font-weight:800;margin-top:2px">{it.activity || it.title || it.t || "Kegiatan"}</div>
              {#if it.notes || it.description}<div style="font-size:12.5px;color:var(--c-muted);margin-top:3px;line-height:1.5">{it.notes || it.description}</div>{/if}
            </div>
          </div>
        {/each}
        <div style="height:24px"></div>
      </div>
    {:else}
      <div class="m-empty" style="padding:30px 20px">Itinerary belum dibuat untuk grup ini</div>
    {/if}
  {/if}
</MScreen>
