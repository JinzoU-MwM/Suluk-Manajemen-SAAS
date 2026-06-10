<script>
  import { onMount } from "svelte";
  import { Plane, Building2, Globe, Heart, Truck, FileText, Plus } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MBadge from "../ui/MBadge.svelte";
  import MFormSheet from "../ui/MFormSheet.svelte";

  let { nav } = $props();
  let vendors = $state([]);
  let debt = $state({});
  let loading = $state(true);
  let formOpen = $state(false);
  let editing = $state(null);

  const TYPES = [
    { value: "maskapai", label: "Maskapai" }, { value: "hotel", label: "Hotel" },
    { value: "muassasah", label: "Muassasah" }, { value: "katering", label: "Katering" },
    { value: "transportasi", label: "Transportasi" }, { value: "visa", label: "Visa" },
    { value: "handling", label: "Handling" }, { value: "lainnya", label: "Lainnya" },
  ];
  const FIELDS = [
    { key: "name", label: "Nama Vendor", required: true },
    { key: "type", label: "Kategori", type: "select", options: TYPES, required: true },
    { key: "pic_name", label: "Nama PIC" },
    { key: "pic_phone", label: "No. HP PIC", type: "tel" },
    { key: "pic_email", label: "Email PIC", type: "email" },
    { key: "address", label: "Alamat / Kota" },
    { key: "npwp", label: "NPWP" },
    { key: "notes", label: "Catatan", type: "textarea" },
  ];

  const ICON = { Maskapai: Plane, maskapai: Plane, Hotel: Building2, hotel: Building2, Muassasah: Globe, Katering: Heart, katering: Heart, Transportasi: Truck, transportasi: Truck, Visa: FileText, visa: FileText };
  const debtOf = (v) => Number(debt[v.id] ?? v.utang ?? v.outstanding ?? 0);

  async function load() {
    try {
      const [vs, ds] = await Promise.all([ApiService.listVendors({ pageSize: 50 }), ApiService.getDebtSummary().catch(() => null)]);
      vendors = vs?.vendors || vs?.data || (Array.isArray(vs) ? vs : []) || [];
      const rows = ds?.by_vendor || ds?.vendors || [];
      if (Array.isArray(rows)) for (const r of rows) debt[r.vendor_id || r.id] = r.outstanding ?? r.utang ?? r.total;
    } catch {
      vendors = [];
    } finally {
      loading = false;
    }
  }
  onMount(load);

  function openCreate() {
    editing = null;
    formOpen = true;
  }
  function openEdit(v) {
    editing = v;
    formOpen = true;
  }
  async function submit(data) {
    if (editing) {
      await ApiService.updateVendor(editing.id, data);
      nav.toast("Vendor diperbarui");
    } else {
      await ApiService.createVendor(data);
      nav.toast("Vendor ditambahkan");
    }
    await load();
  }

  let totalDebt = $derived(vendors.reduce((s, v) => s + debtOf(v), 0));
</script>

{#snippet headerRight()}
  <button type="button" class="m-nav-btn" onclick={openCreate} aria-label="Tambah vendor"><Plus size={22} /></button>
{/snippet}

<MScreen title="Vendor & Pemasok" onBack={nav.back} right={headerRight}>
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
          <div class="m-row" role="button" tabindex="0" onclick={() => openEdit(v)} onkeydown={(e) => e.key === "Enter" && openEdit(v)}>
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

<MFormSheet open={formOpen} title={editing ? "Edit Vendor" : "Vendor Baru"} fields={FIELDS} initial={editing || {}} submitLabel={editing ? "Simpan" : "Tambah Vendor"} onClose={() => (formOpen = false)} onSubmit={submit} />
