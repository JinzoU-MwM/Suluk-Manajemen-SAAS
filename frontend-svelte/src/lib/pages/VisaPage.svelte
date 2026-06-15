<script>
  import { onMount } from 'svelte';
  import { Plus, Search, FileCheck, Loader2, Clock, X } from 'lucide-svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import Button from '../components/ui/Button.svelte';
  import Avatar from '../components/Avatar.svelte';
  import { showToast, mapError } from '../services/toast.svelte.js';
  import { formatDate } from '../utils/formatting.js';
  import { ApiService } from '../services/api.js';

  let visas = $state([]);
  let loading = $state(true);
  let error = $state('');

  // Columns = the visa state machine stages.
  const STAGES = [
    { id: 'draft',     label: 'Draf',        color: 'var(--c-muted)' },
    { id: 'submitted', label: 'Diajukan',    color: 'var(--c-info)' },
    { id: 'approved',  label: 'Disetujui',   color: 'var(--c-success)' },
    { id: 'rejected',  label: 'Ditolak',     color: 'var(--c-danger)' },
    { id: 'expired',   label: 'Kedaluwarsa', color: 'var(--c-warning)' },
  ];

  // Create-draft drawer
  let showCreate = $state(false);
  let jamaahSearch = $state('');
  let jamaahResults = $state([]);
  let searching = $state(false);
  let picked = $state(null);
  let createForm = $state({ provider: '', reference_no: '', expiry_date: '', notes: '' });
  let saving = $state(false);
  let searchDebounce;

  // Approve drawer
  let showApprove = $state(false);
  let approveTarget = $state(null);
  let approveForm = $state({ reference_no: '', expiry_date: '' });

  function colItems(id) { return visas.filter(v => v.status === id); }
  let stats = $derived([
    { label: 'Diajukan', value: colItems('submitted').length, icon: Clock, accent: 'var(--c-info)' },
    { label: 'Disetujui', value: colItems('approved').length, icon: FileCheck, accent: 'var(--c-success)' },
    { label: 'Ditolak', value: colItems('rejected').length, icon: X, accent: 'var(--c-danger)' },
    { label: 'Kedaluwarsa', value: colItems('expired').length, icon: Clock, accent: 'var(--c-warning)' },
  ]);

  onMount(load);

  async function load() {
    loading = true; error = '';
    try {
      const res = await ApiService.listVisas({ pageSize: 200 });
      visas = res.data || [];
    } catch (e) { error = e.message; showToast(mapError(e.message), 'error'); }
    finally { loading = false; }
  }

  function openCreate() {
    picked = null; jamaahSearch = ''; jamaahResults = [];
    createForm = { provider: '', reference_no: '', expiry_date: '', notes: '' };
    showCreate = true;
  }

  function onJamaahSearch() {
    clearTimeout(searchDebounce);
    searchDebounce = setTimeout(async () => {
      if (jamaahSearch.trim().length < 2) { jamaahResults = []; return; }
      searching = true;
      try {
        const res = await ApiService.listJamaah({ search: jamaahSearch, pageSize: 8 });
        jamaahResults = res.jamaah || res.data || [];
      } catch { jamaahResults = []; } finally { searching = false; }
    }, 300);
  }

  async function saveDraft() {
    if (!picked) { showToast('Pilih jamaah dulu', 'warning'); return; }
    saving = true;
    try {
      await ApiService.upsertVisa(picked.id, createForm);
      showToast('Pengajuan visa dibuat (draf)', 'success');
      showCreate = false;
      await load();
    } catch (e) { showToast(mapError(e.message), 'error'); } finally { saving = false; }
  }

  async function transition(v, status, extra = {}) {
    try {
      await ApiService.transitionVisa(v.jamaah_id, { status, ...extra });
      showToast('Status visa diperbarui', 'success');
      await load();
    } catch (e) { showToast(mapError(e.message), 'error'); }
  }

  function openApprove(v) {
    approveTarget = v;
    approveForm = { reference_no: v.reference_no || '', expiry_date: '' };
    showApprove = true;
  }
  async function confirmApprove() {
    if (!approveTarget) return;
    await transition(approveTarget, 'approved', approveForm);
    showApprove = false; approveTarget = null;
  }
  async function reject(v) {
    const reason = prompt('Alasan penolakan visa?');
    if (reason === null) return;
    await transition(v, 'rejected', { reason });
  }
</script>

<div class="flex h-full min-h-0 flex-col" style="background:var(--c-bg)">
  <div class="flex-shrink-0 px-6 pt-6">
    <PageHeader title="Visa & Dokumen" subtitle="Lacak siklus pengajuan visa jamaah dari draf hingga terbit, dengan pengingat masa berlaku otomatis.">
      {#snippet actions()}
        <Button variant="primary" icon={Plus} onclick={openCreate}>Pengajuan Visa</Button>
      {/snippet}
    </PageHeader>
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {#each stats as s}<StatCard icon={s.icon} label={s.label} value={String(s.value)} accent={s.accent} />{/each}
    </div>
  </div>

  {#if loading}
    <div class="flex-1 p-6"><div class="h-40 animate-pulse rounded-2xl" style="background:var(--c-bg-2)"></div></div>
  {:else if error}
    <div class="m-6 rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{mapError(error)}</div>
  {:else}
    <div class="min-h-0 flex-1 overflow-x-auto px-6 py-5">
      <div class="flex h-full items-start gap-3.5">
        {#each STAGES as col}
          {@const items = colItems(col.id)}
          <div class="flex h-full min-h-[120px] w-[280px] flex-shrink-0 flex-col rounded-2xl p-2.5" style="background:var(--c-bg-2)">
            <div class="flex items-center gap-2 px-1.5 pb-3 pt-1.5">
              <span class="h-2.5 w-2.5 rounded-full" style="background:{col.color}"></span>
              <span class="text-[13.5px] font-extrabold" style="color:var(--c-ink)">{col.label}</span>
              <span class="rounded-full px-2 py-0.5 text-[11px] font-bold" style="background:var(--c-surface);color:var(--c-faint)">{items.length}</span>
            </div>
            <div class="flex min-h-[60px] flex-1 flex-col gap-2.5 overflow-y-auto">
              {#each items as v (v.id)}
                <div class="rounded-xl p-3" style="background:var(--c-surface);border:1px solid var(--c-line);border-left:3px solid {col.color};box-shadow:var(--shadow-sm)">
                  <div class="mb-2 flex items-center gap-2.5">
                    <Avatar name={v.jamaah_name} size={28} />
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-[13px] font-bold" style="color:var(--c-ink)">{v.jamaah_name}</p>
                      {#if v.provider}<p class="truncate text-[11px]" style="color:var(--c-faint)">{v.provider}</p>{/if}
                    </div>
                  </div>
                  {#if v.reference_no}<p class="mb-1 text-[11px]" style="color:var(--c-muted)">No: {v.reference_no}</p>{/if}
                  {#if v.expiry_date && col.id === 'approved'}<p class="mb-1 text-[11px]" style="color:var(--c-muted)">Berlaku s/d {formatDate(v.expiry_date)}</p>{/if}
                  {#if v.reject_reason && col.id === 'rejected'}<p class="mb-1 text-[11px]" style="color:var(--c-danger)">{v.reject_reason}</p>{/if}

                  <div class="mt-2 flex flex-wrap gap-1.5">
                    {#if col.id === 'draft'}
                      <button type="button" class="visa-act" onclick={() => transition(v, 'submitted')}>Ajukan</button>
                    {:else if col.id === 'submitted'}
                      <button type="button" class="visa-act visa-act-ok" onclick={() => openApprove(v)}>Setujui</button>
                      <button type="button" class="visa-act visa-act-no" onclick={() => reject(v)}>Tolak</button>
                    {:else if col.id === 'rejected' || col.id === 'expired'}
                      <button type="button" class="visa-act" onclick={() => transition(v, 'submitted')}>Ajukan Ulang</button>
                    {/if}
                  </div>
                </div>
              {/each}
              {#if items.length === 0}<p class="px-1 py-2 text-[11px]" style="color:var(--c-faint)">Belum ada</p>{/if}
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<!-- Create draft drawer -->
<SlideDrawer open={showCreate} title="Pengajuan Visa Baru" width="480px" onClose={() => (showCreate = false)}>
  <div class="space-y-4 p-6">
    {#if picked}
      <div class="flex items-center justify-between rounded-xl p-3" style="background:var(--c-bg-2)">
        <div class="flex items-center gap-2"><Avatar name={picked.nama || picked.name} size={32} /><span class="text-sm font-bold" style="color:var(--c-ink)">{picked.nama || picked.name}</span></div>
        <button type="button" onclick={() => (picked = null)} style="color:var(--c-faint)"><X class="h-4 w-4" /></button>
      </div>
    {:else}
      <div class="relative">
        <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" style="color:var(--c-faint)" />
        <input type="text" bind:value={jamaahSearch} oninput={onJamaahSearch} placeholder="Cari jamaah (nama/NIK)..." class="w-full rounded-xl border border-slate-200 py-2.5 pl-9 pr-3 text-sm outline-none focus:border-primary-400" />
      </div>
      {#if searching}<p class="text-xs" style="color:var(--c-faint)">Mencari…</p>{/if}
      <div class="max-h-52 space-y-1 overflow-y-auto">
        {#each jamaahResults as j}
          <button type="button" onclick={() => (picked = j)} class="flex w-full items-center gap-2 rounded-lg p-2 text-left hover:bg-slate-50">
            <Avatar name={j.nama || j.name} size={28} />
            <div><p class="text-sm font-medium" style="color:var(--c-ink)">{j.nama || j.name}</p><p class="text-xs" style="color:var(--c-faint)">{j.no_identitas || j.no_hp || ''}</p></div>
          </button>
        {/each}
      </div>
    {/if}

    {#if picked}
      <div class="flex flex-col gap-1"><label for="v-prov" class="text-xs font-medium text-slate-700">Provider Visa</label><input id="v-prov" type="text" bind:value={createForm.provider} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="v-ref" class="text-xs font-medium text-slate-700">No. Referensi</label><input id="v-ref" type="text" bind:value={createForm.reference_no} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="v-notes" class="text-xs font-medium text-slate-700">Catatan</label><input id="v-notes" type="text" bind:value={createForm.notes} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex gap-2 pt-2">
        <button type="button" onclick={() => (showCreate = false)} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
        <button type="button" onclick={saveDraft} disabled={saving} class="flex flex-1 items-center justify-center gap-2 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{#if saving}<Loader2 class="h-4 w-4 animate-spin" />{/if}Simpan Draf</button>
      </div>
    {/if}
  </div>
</SlideDrawer>

<!-- Approve drawer -->
<SlideDrawer open={showApprove} title="Setujui Visa" width="420px" onClose={() => (showApprove = false)}>
  <div class="space-y-4 p-6">
    <p class="text-sm" style="color:var(--c-muted)">Masukkan nomor visa & masa berlaku untuk {approveTarget?.jamaah_name}.</p>
    <div class="flex flex-col gap-1"><label for="ap-ref" class="text-xs font-medium text-slate-700">No. Visa</label><input id="ap-ref" type="text" bind:value={approveForm.reference_no} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="ap-exp" class="text-xs font-medium text-slate-700">Berlaku Sampai</label><input id="ap-exp" type="date" bind:value={approveForm.expiry_date} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => (showApprove = false)} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={confirmApprove} class="flex-1 rounded-xl py-2.5 text-sm font-semibold text-white" style="background:var(--c-success)">Setujui</button>
    </div>
  </div>
</SlideDrawer>

<style>
  .visa-act { padding: 4px 10px; font-size: 11px; font-weight: 700; border-radius: 8px; border: 1px solid var(--c-line); color: var(--c-ink-soft); transition: all .15s; }
  .visa-act:hover { border-color: var(--c-primary); color: var(--c-primary); }
  .visa-act-ok { background: var(--c-success-soft); color: var(--c-success); border-color: transparent; }
  .visa-act-no { background: var(--c-danger-soft); color: var(--c-danger); border-color: transparent; }
</style>
