<script>
  import { onMount } from "svelte";
  import { Building2, Plus } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MFormSheet from "../ui/MFormSheet.svelte";

  let { nav } = $props();
  let agents = $state([]);
  let loading = $state(true);
  let formOpen = $state(false);
  let editing = $state(null);

  const FIELDS = [
    { key: "name", label: "Nama Agen / Travel", required: true },
    { key: "phone", label: "No. HP", type: "tel" },
    { key: "email", label: "Email", type: "email" },
    { key: "commission_rate", label: "Komisi (desimal, mis. 0.04)", inputmode: "decimal", placeholder: "0.04" },
    { key: "address", label: "Alamat / Kota" },
  ];

  const outstanding = (a) => Number(a.total_outstanding ?? (a.omzet ? a.omzet * a.komisi : 0));

  async function load() {
    try {
      const res = await ApiService.listAgents({ limit: 50 });
      agents = res?.agents || res?.data || (Array.isArray(res) ? res : []) || [];
    } catch {
      agents = [];
    } finally {
      loading = false;
    }
  }
  onMount(load);

  function openCreate() {
    editing = null;
    formOpen = true;
  }
  function openEdit(a) {
    editing = { ...a, commission_rate: a.commission_rate ?? "" };
    formOpen = true;
  }
  async function submit(data) {
    const payload = { ...data };
    if (payload.commission_rate !== "" && payload.commission_rate != null) payload.commission_rate = Number(payload.commission_rate) || 0;
    if (editing) {
      await ApiService.updateAgent(editing.id, payload);
      nav.toast("Agen diperbarui");
    } else {
      await ApiService.createAgent(payload);
      nav.toast("Agen ditambahkan");
    }
    await load();
  }

  let active = $derived(agents.filter((a) => (a.status || "Aktif") !== "Nonaktif" && a.is_active !== false).length);
  let totalOwed = $derived(agents.reduce((s, a) => s + outstanding(a), 0));
</script>

{#snippet headerRight()}
  <button type="button" class="m-nav-btn" onclick={openCreate} aria-label="Tambah agen"><Plus size={22} /></button>
{/snippet}

<MScreen title="Agen & Mitra" onBack={nav.back} right={headerRight}>
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
          <div class="m-row" role="button" tabindex="0" onclick={() => openEdit(a)} onkeydown={(e) => e.key === "Enter" && openEdit(a)}>
            <div class="m-row-ic" style="background:var(--c-primary-soft);color:var(--c-primary-deep)"><Building2 size={18} /></div>
            <div class="m-row-main">
              <div class="m-row-title">{a.name || a.nama}</div>
              <div class="m-row-sub">{(a.parent_name ? "↳ " + a.parent_name : (a.pic_name || a.pic || a.phone || "—")) + " · " + (a.total_jamaah ?? a.jamaah ?? 0) + " jamaah"}</div>
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

<MFormSheet open={formOpen} title={editing ? "Edit Agen" : "Agen Baru"} fields={FIELDS} initial={editing || {}} submitLabel={editing ? "Simpan" : "Tambah Agen"} onClose={() => (formOpen = false)} onSubmit={submit} />
