<script>
  import { onMount } from "svelte";
  import { Plus, Search as SearchIcon, ScanLine } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MSearch from "../ui/MSearch.svelte";
  import MChips from "../ui/MChips.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MAvatar from "../ui/MAvatar.svelte";
  import MBadge from "../ui/MBadge.svelte";
  import MSheet from "../ui/MSheet.svelte";
  import MFormSheet from "../ui/MFormSheet.svelte";

  let { nav } = $props();

  let all = $state([]);
  let total = $state(0);
  let loading = $state(true);
  let q = $state("");
  let tab = $state("Semua");
  let addOpen = $state(false); // choice sheet: scan vs manual
  let formOpen = $state(false);
  const tabs = ["Semua", "Lunas", "Cicilan", "DP", "Verifikasi"];

  const FIELDS = [
    { key: "nama", label: "Nama Lengkap", required: true },
    { key: "no_hp", label: "No. HP", type: "tel" },
    { key: "jenis_identitas", label: "Jenis Identitas", type: "select", options: [{ value: "ktp", label: "KTP" }, { value: "paspor", label: "Paspor" }] },
    { key: "no_identitas", label: "No. Identitas" },
    { key: "gender", label: "Jenis Kelamin", type: "select", options: [{ value: "L", label: "Laki-laki" }, { value: "P", label: "Perempuan" }] },
    { key: "kota", label: "Kota Asal" },
  ];

  async function load() {
    try {
      const res = await ApiService.listJamaah({ pageSize: 50 });
      all = res?.data || res?.jamaah || (Array.isArray(res) ? res : []) || [];
      total = res?.meta?.total ?? all.length;
    } catch {
      all = [];
    } finally {
      loading = false;
    }
  }
  onMount(load);

  async function submit(data) {
    const payload = Object.fromEntries(Object.entries(data).filter(([, v]) => v !== "" && v != null));
    await ApiService.createProfile(payload);
    nav.toast("Jamaah ditambahkan");
    await load();
  }

  const nm = (j) => j.nama || j.name || "Tanpa Nama";
  const sub = (j) => j.paket || j.package_name || j.kota || j.no_hp || j.id || "";
  let list = $derived(
    all.filter((j) => {
      if (tab !== "Semua" && j.status !== tab) return false;
      if (!q) return true;
      const s = q.toLowerCase();
      return [nm(j), j.kota, j.id, j.no_hp].some((v) => String(v || "").toLowerCase().includes(s));
    }),
  );
</script>

<div class="m-screen-root">
  <div class="m-head">
    <div class="m-head-row">
      <div style="flex:1">
        <div class="m-head-title">Jamaah</div>
        <div class="m-head-sub">{total.toLocaleString("id-ID")} jamaah terdaftar</div>
      </div>
      <button type="button" onclick={() => (addOpen = true)} aria-label="Tambah jamaah" style="width:42px;height:42px;border-radius:13px;background:var(--c-primary);color:#fff;display:flex;align-items:center;justify-content:center">
        <Plus size={22} />
      </button>
    </div>
    <div style="margin-top:14px"><MSearch bind:value={q} placeholder="Cari nama, kota, ID…" /></div>
  </div>

  <div style="padding-bottom:8px"><MChips {tabs} value={tab} onChange={(v) => (tab = v)} /></div>

  <div class="m-scroll">
    <div style="padding:6px 18px 0">
      {#if loading}
        <div class="m-loading" style="padding:60px 0">Memuat…</div>
      {:else if list.length}
        <MGroup class="m-stagger">
          {#each list as j (j.id)}
            <div class="m-row" role="button" tabindex="0" onclick={() => nav.go("jamaah-detail", { id: j.id, jamaah: j })}
              onkeydown={(e) => e.key === "Enter" && nav.go("jamaah-detail", { id: j.id, jamaah: j })}>
              <MAvatar name={nm(j)} size={42} />
              <div class="m-row-main">
                <div class="m-row-title">{nm(j)}</div>
                <div class="m-row-sub">{sub(j)}</div>
              </div>
              {#if j.status}
                <div style="text-align:right;flex-shrink:0"><MBadge status={j.status} /></div>
              {/if}
            </div>
          {/each}
        </MGroup>
      {:else}
        <div class="m-empty">
          <SearchIcon size={30} class="ic" />
          <div style="margin-top:10px;font-size:14px">Tidak ada jamaah</div>
        </div>
      {/if}
    </div>
    <div style="height:24px"></div>
  </div>
</div>

<MSheet open={addOpen} title="Tambah Jamaah" onClose={() => (addOpen = false)}>
  <div style="display:flex;flex-direction:column;gap:10px;padding-bottom:6px">
    <button type="button" class="m-btn m-btn-primary" onclick={() => { addOpen = false; nav.switchTab("scan"); }}><ScanLine size={18} />Pindai Dokumen (AI Scanner)</button>
    <button type="button" class="m-btn m-btn-ghost" onclick={() => { addOpen = false; formOpen = true; }}><Plus size={18} />Input Manual</button>
  </div>
</MSheet>

<MFormSheet open={formOpen} title="Jamaah Baru" fields={FIELDS} submitLabel="Tambah Jamaah" onClose={() => (formOpen = false)} onSubmit={submit} />
