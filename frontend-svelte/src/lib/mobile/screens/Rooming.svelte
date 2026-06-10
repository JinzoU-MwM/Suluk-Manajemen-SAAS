<script>
  import { onMount } from "svelte";
  import { BedDouble } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MScreen from "../ui/MScreen.svelte";
  import MChips from "../ui/MChips.svelte";
  import MAvatar from "../ui/MAvatar.svelte";

  let { nav } = $props();
  let groups = $state([]);
  let groupId = $state(null);
  let rooms = $state([]);
  let summary = $state(null);
  let loading = $state(true);
  let loadingRooms = $state(false);

  onMount(async () => {
    try {
      const res = await ApiService.listGroups();
      groups = res?.groups || res?.data || (Array.isArray(res) ? res : []) || [];
      if (groups.length) {
        groupId = groups[0].id;
        await loadRooms();
      }
    } catch {}
    finally {
      loading = false;
    }
  });

  async function loadRooms() {
    loadingRooms = true;
    rooms = [];
    try {
      const [rs, sm] = await Promise.all([ApiService.getGroupRooms(groupId).catch(() => null), ApiService.getRoomingSummary(groupId).catch(() => null)]);
      rooms = rs?.rooms || rs?.data || (Array.isArray(rs) ? rs : []) || [];
      summary = sm;
    } finally {
      loadingRooms = false;
    }
  }
  function pick(id) {
    groupId = id;
    loadRooms();
  }
  let chips = $derived(groups.map((g) => ({ value: g.id, label: g.name || g.nama || "Grup" })));
</script>

<MScreen title="Rooming List" onBack={nav.back}>
  {#if loading}
    <div class="m-loading" style="padding:50px 0">Memuat…</div>
  {:else if !groups.length}
    <div class="m-empty" style="padding:50px 20px">Belum ada grup</div>
  {:else}
    <div style="padding:14px 0 6px"><MChips tabs={chips} value={groupId} onChange={pick} /></div>
    {#if summary}
      <div style="padding:6px 18px 0">
        <div class="m-card m-card-pad" style="background:var(--c-primary-tint);border:none">
          <div style="font-size:13.5px;font-weight:700">{(groups.find((g) => g.id === groupId)?.name) || "Grup"}</div>
          <div style="font-size:12.5px;color:var(--c-muted);margin-top:2px">{(summary.total_rooms ?? rooms.length) + " kamar · " + (summary.unassigned_count ?? 0) + " belum ditempatkan"}</div>
        </div>
      </div>
    {/if}

    <div style="padding:16px 18px 0;display:flex;flex-direction:column;gap:12px">
      {#if loadingRooms}
        <div class="m-loading" style="padding:40px 0">Memuat kamar…</div>
      {:else if rooms.length}
        {#each rooms as r (r.id)}
          {@const occ = r.members?.length ?? r.occupied ?? 0}
          {@const cap = r.capacity ?? 0}
          {@const full = r.is_full ?? occ >= cap}
          <div class="m-card m-card-pad">
            <div style="display:flex;align-items:center;gap:10px;margin-bottom:11px">
              <div style="width:36px;height:36px;border-radius:10px;background:var(--c-primary-soft);color:var(--c-primary-deep);display:flex;align-items:center;justify-content:center"><BedDouble size={18} /></div>
              <div style="flex:1"><div style="font-size:14.5px;font-weight:800">{r.room_number || r.id}</div><div style="font-size:11.5px;color:var(--c-faint)">{r.room_type || r.gender_type || r.tipe || ""}</div></div>
              <span class="m-chip" style="background:{full ? 'var(--c-success-soft)' : 'var(--c-bg-2)'};color:{full ? 'var(--c-success)' : 'var(--c-muted)'}">{occ}/{cap}</span>
            </div>
            <div style="display:flex;flex-direction:column;gap:7px">
              {#each r.members || [] as m, i (i)}
                {@const nm = m.name || m.nama || m.member_name || "Jamaah"}
                <div style="display:flex;align-items:center;gap:9px"><MAvatar name={nm} size={28} /><span style="font-size:13px;font-weight:500">{nm}</span></div>
              {/each}
              {#each Array(Math.max(0, cap - occ)) as _, i (i)}
                <div style="border:1.5px dashed var(--c-line);border-radius:10px;padding:8px;text-align:center;font-size:12.5px;color:var(--c-faint)">Kosong</div>
              {/each}
            </div>
          </div>
        {/each}
      {:else}
        <div class="m-empty" style="padding:30px 20px">Belum ada kamar untuk grup ini</div>
      {/if}
    </div>
  {/if}
</MScreen>
