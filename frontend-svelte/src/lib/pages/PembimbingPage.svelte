<script>
  import { onMount } from "svelte";
  import { Plus, Search, Pencil, Trash2, UserCog, BadgeCheck, Loader2 } from "lucide-svelte";
  import PageHeader from "../components/PageHeader.svelte";
  import StatCard from "../components/StatCard.svelte";
  import EmptyState from "../components/EmptyState.svelte";
  import SlideDrawer from "../components/SlideDrawer.svelte";
  import Card from "../components/ui/Card.svelte";
  import Button from "../components/ui/Button.svelte";
  import Avatar from "../components/Avatar.svelte";
  import { showToast, mapError } from "../services/toast.svelte.js";
  import { formatDate } from "../utils/formatting.js";
  import { ApiService } from "../services/api.js";

  let guides = $state([]);
  let loading = $state(true);
  let search = $state("");
  let searchDebounce;

  let showForm = $state(false);
  let editing = $state(null);
  const empty = { name: "", phone: "", email: "", type: "mutawwif", license_no: "", license_expiry: "", notes: "" };
  let form = $state({ ...empty });
  let saving = $state(false);

  const TYPES = [
    { value: "mutawwif", label: "Mutawwif" },
    { value: "tour_leader", label: "Tour Leader" },
    { value: "kesehatan", label: "Tim Kesehatan" },
  ];
  function typeLabel(t) { return TYPES.find((x) => x.value === t)?.label || t; }

  let stats = $derived({
    total: guides.length,
    active: guides.filter((g) => g.is_active).length,
    assigned: guides.reduce((s, g) => s + (g.assignment_count || 0), 0),
  });

  onMount(load);

  async function load() {
    loading = true;
    try {
      const r = await ApiService.listGuides(search);
      guides = r.guides || [];
    } catch (e) { showToast(mapError(e.message), "error"); } finally { loading = false; }
  }

  function onSearch() {
    clearTimeout(searchDebounce);
    searchDebounce = setTimeout(load, 300);
  }

  function openNew() { editing = null; form = { ...empty }; showForm = true; }
  function openEdit(g) {
    editing = g;
    form = { name: g.name, phone: g.phone, email: g.email, type: g.type, license_no: g.license_no, license_expiry: (g.license_expiry || "").slice(0, 10), notes: g.notes };
    showForm = true;
  }

  async function save() {
    if (!form.name.trim()) { showToast("Nama wajib diisi", "warning"); return; }
    saving = true;
    try {
      if (editing) await ApiService.updateGuide(editing.id, form);
      else await ApiService.createGuide(form);
      showToast("Pembimbing disimpan", "success");
      showForm = false;
      await load();
    } catch (e) { showToast(mapError(e.message), "error"); } finally { saving = false; }
  }

  async function remove(g) {
    if (!confirm(`Hapus pembimbing ${g.name}?`)) return;
    try { await ApiService.deleteGuide(g.id); showToast("Dihapus", "success"); await load(); }
    catch (e) { showToast(mapError(e.message), "error"); }
  }

  function licenseExpired(g) {
    return g.license_expiry && new Date(g.license_expiry) < new Date();
  }
</script>

<div class="flex flex-col gap-6 p-4 lg:p-8" style="background:var(--c-bg);min-height:100vh">
  <PageHeader kicker="Operasional" title="Pembimbing &amp; Mutawwif" subtitle="Kelola roster pembimbing dan tugaskan ke kloter keberangkatan.">
    {#snippet actions()}
      <Button variant="primary" icon={Plus} onclick={openNew}>Tambah Pembimbing</Button>
    {/snippet}
  </PageHeader>

  <div class="grid grid-cols-2 gap-4 lg:grid-cols-3">
    <StatCard icon={UserCog} label="Total Pembimbing" value={String(stats.total)} accent="var(--c-primary)" />
    <StatCard icon={BadgeCheck} label="Aktif" value={String(stats.active)} accent="var(--c-success)" />
    <StatCard icon={UserCog} label="Penugasan Kloter" value={String(stats.assigned)} accent="var(--c-info)" />
  </div>

  <div class="relative max-w-xs">
    <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" style="color:var(--c-faint)" />
    <input type="text" bind:value={search} oninput={onSearch} placeholder="Cari pembimbing..." class="w-full rounded-xl py-2 pl-9 pr-3 text-sm outline-none" style="border:1px solid var(--c-line);color:var(--c-ink)" />
  </div>

  {#if loading}
    <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
  {:else if guides.length === 0}
    <Card><EmptyState icon={UserCog} title="Belum ada pembimbing" text="Tambahkan mutawwif / tour leader untuk ditugaskan ke kloter." /></Card>
  {:else}
    <Card pad={false} style="padding:8px 4px">
      <div class="overflow-x-auto">
        <table class="w-full" style="font-size:13.5px">
          <thead>
            <tr class="text-left text-[11px] font-bold uppercase tracking-wider" style="color:var(--c-faint)">
              <th class="px-4 py-3">Nama</th>
              <th class="px-4 py-3">Tipe</th>
              <th class="hidden px-4 py-3 md:table-cell">Lisensi</th>
              <th class="px-4 py-3 text-center">Kloter</th>
              <th class="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody>
            {#each guides as g (g.id)}
              <tr style="border-top:1px solid var(--c-line-soft)">
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2.5">
                    <Avatar name={g.name} size={34} />
                    <div>
                      <p class="font-bold" style="color:var(--c-ink)">{g.name}{#if !g.is_active}<span class="ml-1 text-[10px]" style="color:var(--c-faint)">(nonaktif)</span>{/if}</p>
                      <p class="text-xs" style="color:var(--c-faint)">{g.phone || "—"}</p>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3"><span class="rounded-full px-2 py-0.5 text-[11px] font-semibold" style="background:var(--c-primary-tint);color:var(--c-primary)">{typeLabel(g.type)}</span></td>
                <td class="hidden px-4 py-3 md:table-cell" style="color:var(--c-muted)">
                  {#if g.license_no}{g.license_no}{#if g.license_expiry} · <span style="color:{licenseExpired(g) ? 'var(--c-danger)' : 'var(--c-muted)'}">{formatDate(g.license_expiry)}</span>{/if}{:else}—{/if}
                </td>
                <td class="px-4 py-3 text-center font-bold" style="color:var(--c-ink)">{g.assignment_count || 0}</td>
                <td class="px-4 py-3 text-right">
                  <div class="flex justify-end gap-1.5">
                    <button type="button" onclick={() => openEdit(g)} class="rounded-lg p-1.5" style="color:var(--c-ink-soft)" aria-label="Edit"><Pencil class="h-4 w-4" /></button>
                    <button type="button" onclick={() => remove(g)} class="rounded-lg p-1.5" style="color:var(--c-danger)" aria-label="Hapus"><Trash2 class="h-4 w-4" /></button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </Card>
  {/if}
</div>

<SlideDrawer open={showForm} title={editing ? "Edit Pembimbing" : "Tambah Pembimbing"} width="460px" onClose={() => (showForm = false)}>
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1"><label for="g-name" class="text-xs font-medium text-slate-700">Nama <span class="text-red-500">*</span></label><input id="g-name" type="text" bind:value={form.name} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1"><label for="g-phone" class="text-xs font-medium text-slate-700">Telepon</label><input id="g-phone" type="tel" bind:value={form.phone} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="g-type" class="text-xs font-medium text-slate-700">Tipe</label><select id="g-type" bind:value={form.type} class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400">{#each TYPES as t}<option value={t.value}>{t.label}</option>{/each}</select></div>
    </div>
    <div class="flex flex-col gap-1"><label for="g-email" class="text-xs font-medium text-slate-700">Email</label><input id="g-email" type="email" bind:value={form.email} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1"><label for="g-lic" class="text-xs font-medium text-slate-700">No. Lisensi/SKP</label><input id="g-lic" type="text" bind:value={form.license_no} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="g-exp" class="text-xs font-medium text-slate-700">Masa Berlaku</label><input id="g-exp" type="date" bind:value={form.license_expiry} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
    </div>
    <div class="flex flex-col gap-1"><label for="g-notes" class="text-xs font-medium text-slate-700">Catatan</label><input id="g-notes" type="text" bind:value={form.notes} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => (showForm = false)} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={save} disabled={saving} class="flex flex-1 items-center justify-center gap-2 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{#if saving}<Loader2 class="h-4 w-4 animate-spin" />{/if}Simpan</button>
    </div>
  </div>
</SlideDrawer>
