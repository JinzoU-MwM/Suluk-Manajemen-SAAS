<script>
  import { onMount } from 'svelte';
  import {
    Plus, Search, LayoutGrid, List,
    Phone, ChevronRight, UserCircle, Loader2,
    Users, CheckCircle, CreditCard, Clock,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import EmptyState from '../components/EmptyState.svelte';
  import Pager from '../components/Pager.svelte';
  import Avatar from '../components/Avatar.svelte';
  import { showToast, mapError } from '../services/toast.svelte.js';
  import { formatRupiah as formatIDR } from '../utils/formatting.js';
  import { ApiService } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  // ── State ──────────────────────────────────────────────
  let jamaah = $state([]);
  let isLoading = $state(true);
  let error = $state('');
  let searchQuery = $state('');
  let filterStatus = $state('all');
  let viewMode = $state('table'); // 'table' | 'kanban'
  let drawerOpen = $state(false);
  let selectedJamaah = $state(null);
  let activeTab = $state('profil');

  // Pagination (server-side via /jamaah/crm meta)
  const PAGE_SIZE = 25;
  let page = $state(1);
  let total = $state(0);
  let pkgMap = new Map();
  let searchDebounce;

  // Create form
  let showCreate = $state(false);
  let saving = $state(false);
  const emptyForm = { nama: '', no_hp: '', email: '', gender: '', no_identitas: '', no_paspor: '', alamat: '', lead_source: 'walk_in' };
  let form = $state({ ...emptyForm });

  const PIPELINE_STATUSES = [
    { id: 'all',       label: 'Semua' },
    { id: 'prospek',   label: 'Prospek' },
    { id: 'survey',    label: 'Survey' },
    { id: 'booking',   label: 'Booking' },
    { id: 'dp',        label: 'DP' },
    { id: 'cicilan',   label: 'Cicilan' },
    { id: 'lunas',     label: 'Lunas' },
    { id: 'berangkat', label: 'Berangkat' },
    { id: 'batal',     label: 'Batal' },
  ];

  const DRAWER_TABS = [
    { id: 'profil',   label: 'Profil' },
    { id: 'invoice',  label: 'Invoice' },
    { id: 'catatan',  label: 'Catatan' },
  ];

  const LEAD_SOURCES = [
    { id: 'walk_in',  label: 'Walk-in' },
    { id: 'referral', label: 'Referral' },
    { id: 'online',   label: 'Online' },
    { id: 'agent',    label: 'Agen' },
  ];

  // Client-side status filter over the current page (search + paging are server-side)
  let filtered = $derived(
    filterStatus === 'all' ? jamaah : jamaah.filter(j => j.pipeline_status === filterStatus)
  );

  // Summary tiles (Suluk design). Counts are over the loaded page; total is server-side.
  let statTiles = $derived([
    { label: 'Total Jamaah', value: total, icon: Users, accent: '#1B7F5A' },
    { label: 'Lunas', value: jamaah.filter(j => ['lunas', 'berangkat'].includes(j.pipeline_status)).length, icon: CheckCircle, accent: '#1B7F5A' },
    { label: 'Masih Cicilan', value: jamaah.filter(j => ['cicilan', 'dp'].includes(j.pipeline_status)).length, icon: CreditCard, accent: '#2563a8' },
    { label: 'Prospek', value: jamaah.filter(j => ['prospek', 'survey', 'booking'].includes(j.pipeline_status)).length, icon: Clock, accent: '#C99A2E' },
  ]);

  onMount(loadJamaah);

  async function loadJamaah() {
    isLoading = true;
    error = '';
    try {
      const [crm, pkgs] = await Promise.all([
        ApiService.listCRM({ search: searchQuery, page, pageSize: PAGE_SIZE }),
        ApiService.listPackages({ pageSize: 200 }).catch(() => ([])),
      ]);
      const pkgList = pkgs?.packages || pkgs?.data || pkgs || [];
      pkgMap = new Map(pkgList.map(p => [p.id, p.name]));
      total = crm.meta?.total || 0;
      jamaah = (crm.data || []).map(r => ({
        id: r.id,
        name: r.nama,
        phone: r.no_hp || '',
        nik: r.no_identitas || '',
        passport_no: r.no_paspor || '',
        email: r.email || '',
        gender: r.gender || '',
        package_name: r.package_id ? (pkgMap.get(r.package_id) || 'Paket') : null,
        room_type: r.room_type || null,
        pipeline_status: r.pipeline_status || 'prospek',
        total_invoice: r.total_amount || 0,
        paid: r.total_paid || 0,
        sisa_tagihan: r.total_remaining || 0,
      }));
    } catch (e) {
      error = e.message;
      showToast(mapError(e.message), 'error');
    } finally {
      isLoading = false;
    }
  }

  function onSearchInput() {
    clearTimeout(searchDebounce);
    searchDebounce = setTimeout(() => { page = 1; loadJamaah(); }, 350);
  }

  function gotoPage(p) {
    page = p;
    loadJamaah();
  }

  function openCreate() {
    form = { ...emptyForm };
    showCreate = true;
  }

  async function saveProfile() {
    if (!form.nama.trim()) {
      showToast('Nama wajib diisi', 'warning');
      return;
    }
    saving = true;
    try {
      await ApiService.createProfile(form);
      showToast('Jamaah berhasil ditambahkan', 'success');
      showCreate = false;
      page = 1;
      await loadJamaah();
    } catch (e) {
      showToast(mapError(e.message), 'error');
    } finally {
      saving = false;
    }
  }

  function openDetail(j) {
    selectedJamaah = j;
    activeTab = 'profil';
    drawerOpen = true;
  }

  function openWhatsApp(phone) {
    const clean = (phone || '').replace(/\D/g, '');
    if (!clean) { showToast('Nomor HP belum diisi', 'warning'); return; }
    const number = clean.startsWith('0') ? '62' + clean.slice(1) : clean;
    window.open(`https://wa.me/${number}`, '_blank');
  }
</script>

{#snippet statTile(t)}
  {@const TIcon = t.icon}
  <div class="rounded-2xl border border-slate-200/70 bg-white p-4 shadow-sm">
    <div class="mb-2 flex h-10 w-10 items-center justify-center rounded-xl" style="background:{t.accent}18;color:{t.accent}">
      <TIcon class="h-5 w-5" />
    </div>
    <p class="tabular text-2xl font-extrabold tracking-tight text-[#10211c]" style="font-variant-numeric:tabular-nums">{t.value}</p>
    <p class="mt-0.5 text-[13px] font-medium text-slate-500">{t.label}</p>
  </div>
{/snippet}

<div class="flex h-screen flex-col">
  <!-- Header -->
  <div class="flex-shrink-0 border-b border-slate-100 bg-white px-6 py-5">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-serif text-xl font-bold text-slate-800">CRM & Jamaah</h1>
        <p class="mt-0.5 text-sm text-slate-500">{total} jamaah</p>
      </div>
      <div class="flex items-center gap-2">
        <!-- View toggle -->
        <div class="flex rounded-xl border border-slate-200 p-0.5">
          <button
            type="button"
            onclick={() => (viewMode = 'table')}
            class="flex h-8 w-8 items-center justify-center rounded-lg transition-colors {viewMode === 'table' ? 'bg-primary-600 text-white' : 'text-slate-400 hover:text-slate-600'}"
            title="Tampilan tabel"
          >
            <List class="h-4 w-4" />
          </button>
          <button
            type="button"
            onclick={() => (viewMode = 'kanban')}
            class="flex h-8 w-8 items-center justify-center rounded-lg transition-colors {viewMode === 'kanban' ? 'bg-primary-600 text-white' : 'text-slate-400 hover:text-slate-600'}"
            title="Tampilan kanban"
          >
            <LayoutGrid class="h-4 w-4" />
          </button>
        </div>
        <button
          type="button"
          onclick={openCreate}
          class="flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm shadow-primary-600/30 transition-all hover:bg-primary-700"
        >
          <Plus class="h-4 w-4" />
          Tambah Jamaah
        </button>
      </div>
    </div>

    <!-- Search + filter -->
    <div class="mt-4 flex gap-3">
      <div class="relative flex-1">
        <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <input
          type="text"
          bind:value={searchQuery}
          oninput={onSearchInput}
          placeholder="Cari nama, HP, NIK, paspor..."
          class="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-9 pr-3 text-sm outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>
      <div class="flex gap-1 overflow-x-auto">
        {#each PIPELINE_STATUSES as s}
          <button
            type="button"
            onclick={() => (filterStatus = s.id)}
            class="flex-shrink-0 rounded-lg px-3 py-2 text-xs font-semibold transition-all
              {filterStatus === s.id ? 'bg-primary-600 text-white' : 'text-slate-500 hover:bg-slate-100'}"
          >
            {s.label}
          </button>
        {/each}
      </div>
    </div>
  </div>

  <!-- Table view -->
  {#if viewMode === 'table'}
    <div class="flex-1 overflow-auto">
      <div class="grid grid-cols-2 gap-4 p-6 pb-0 sm:grid-cols-4">
        {#each statTiles as t}{@render statTile(t)}{/each}
      </div>
      {#if isLoading}
        <div class="space-y-3 p-6">
          {#each [1,2,3,4] as _}
            <div class="h-14 animate-pulse rounded-xl bg-slate-100"></div>
          {/each}
        </div>
      {:else if error}
        <div class="m-6 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">{mapError(error)}</div>
      {:else if filtered.length === 0}
        <EmptyState
          icon={UserCircle}
          title={searchQuery || filterStatus !== 'all' ? 'Tidak ada jamaah yang cocok' : 'Belum ada jamaah'}
          text={searchQuery || filterStatus !== 'all' ? 'Coba ubah kata kunci atau filter.' : 'Klik “Tambah Jamaah” untuk menambah data, atau import dari AI Scanner.'}
        />
      {:else}
        <table class="w-full">
          <thead class="sticky top-0 bg-slate-50">
            <tr class="text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
              <th class="px-6 py-3">Jamaah</th>
              <th class="hidden px-4 py-3 md:table-cell">Paket</th>
              <th class="hidden px-4 py-3 lg:table-cell">Pembayaran</th>
              <th class="px-4 py-3">Status</th>
              <th class="px-4 py-3 text-right">Sisa Tagihan</th>
              <th class="px-4 py-3"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            {#each filtered as j}
              <tr class="group bg-white transition-colors hover:bg-primary-50/30">
                <td class="px-6 py-3.5">
                  <div class="flex items-center gap-3">
                    <Avatar name={j.name} size={38} />
                    <div class="min-w-0">
                      <p class="truncate text-sm font-bold text-[#10211c]">{j.name}</p>
                      <p class="truncate text-xs text-slate-400">{j.nik || j.phone || '—'}</p>
                    </div>
                  </div>
                </td>
                <td class="hidden px-4 py-3.5 md:table-cell">
                  <p class="text-sm font-medium text-slate-600">{j.package_name || '—'}</p>
                  {#if j.room_type}<p class="text-xs text-slate-400">{j.room_type}</p>{/if}
                </td>
                <td class="hidden px-4 py-3.5 lg:table-cell">
                  {#if j.total_invoice > 0}
                    <div class="min-w-[140px]">
                      <div class="mb-1 flex justify-between text-xs">
                        <span class="tabular font-bold text-[#10211c]" style="font-variant-numeric:tabular-nums">{formatIDR(j.paid)}</span>
                        <span class="tabular text-slate-400" style="font-variant-numeric:tabular-nums">{formatIDR(j.total_invoice)}</span>
                      </div>
                      <div class="h-1.5 overflow-hidden rounded-full bg-slate-100">
                        <div class="h-full rounded-full {j.sisa_tagihan <= 0 ? 'bg-primary-600' : 'bg-gold-500'}" style:width={`${Math.min(100, Math.round((j.paid / j.total_invoice) * 100))}%`}></div>
                      </div>
                    </div>
                  {:else}
                    <span class="text-xs text-slate-400">—</span>
                  {/if}
                </td>
                <td class="px-4 py-3.5">
                  <StatusBadge status={j.pipeline_status} size="xs" />
                </td>
                <td class="px-4 py-3.5 text-right">
                  {#if j.sisa_tagihan > 0}
                    <span class="text-sm font-semibold text-red-600">{formatIDR(j.sisa_tagihan)}</span>
                  {:else if j.total_invoice > 0}
                    <span class="text-sm font-semibold text-primary-600">Lunas</span>
                  {:else}
                    <span class="text-sm text-slate-400">—</span>
                  {/if}
                </td>
                <td class="px-4 py-3.5 text-right">
                  <button
                    type="button"
                    onclick={() => openDetail(j)}
                    class="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-semibold text-primary-600 transition-colors hover:bg-primary-50"
                  >
                    Detail
                    <ChevronRight class="h-3 w-3" />
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
        <div class="px-6">
          <Pager {page} pageSize={PAGE_SIZE} {total} onchange={gotoPage} />
        </div>
      {/if}
    </div>

  <!-- Kanban view -->
  {:else}
    <div class="flex-1 overflow-x-auto">
      <div class="flex h-full gap-3 p-4">
        {#each PIPELINE_STATUSES.filter(s => s.id !== 'all') as col}
          {@const colItems = jamaah.filter(j => j.pipeline_status === col.id)}
          <div class="flex h-full w-64 flex-shrink-0 flex-col rounded-2xl bg-slate-100 p-3">
            <div class="mb-3 flex items-center justify-between">
              <span class="text-xs font-bold uppercase tracking-wider text-slate-500">{col.label}</span>
              <span class="rounded-full bg-white px-2 py-0.5 text-[10px] font-bold text-slate-500">{colItems.length}</span>
            </div>
            <div class="flex-1 space-y-2 overflow-y-auto">
              {#each colItems as j}
                <button
                  type="button"
                  onclick={() => openDetail(j)}
                  class="w-full rounded-xl bg-white p-3 text-left shadow-sm transition-all hover:shadow-md"
                >
                  <p class="text-sm font-semibold text-slate-800">{j.name}</p>
                  {#if j.package_name}
                    <p class="mt-0.5 text-[11px] text-slate-500">{j.package_name}</p>
                  {/if}
                  {#if j.sisa_tagihan > 0}
                    <p class="mt-2 text-[11px] font-semibold text-red-500">Sisa {formatIDR(j.sisa_tagihan)}</p>
                  {/if}
                </button>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<!-- Tambah Jamaah Drawer -->
<SlideDrawer open={showCreate} title="Tambah Jamaah" width="480px" onClose={() => (showCreate = false)}>
  <div class="space-y-4 p-6">
    <div class="flex flex-col gap-1">
      <label for="c-nama" class="text-xs font-medium text-slate-700">Nama Lengkap <span class="text-red-500">*</span></label>
      <input id="c-nama" type="text" bind:value={form.nama} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="c-hp" class="text-xs font-medium text-slate-700">No. HP</label>
        <input id="c-hp" type="tel" bind:value={form.no_hp} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
      </div>
      <div class="flex flex-col gap-1">
        <label for="c-gender" class="text-xs font-medium text-slate-700">Jenis Kelamin</label>
        <select id="c-gender" bind:value={form.gender} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400">
          <option value="">-</option>
          <option value="L">Laki-laki</option>
          <option value="P">Perempuan</option>
        </select>
      </div>
    </div>
    <div class="flex flex-col gap-1">
      <label for="c-email" class="text-xs font-medium text-slate-700">Email</label>
      <input id="c-email" type="email" bind:value={form.email} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="c-nik" class="text-xs font-medium text-slate-700">NIK</label>
        <input id="c-nik" type="text" bind:value={form.no_identitas} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
      </div>
      <div class="flex flex-col gap-1">
        <label for="c-paspor" class="text-xs font-medium text-slate-700">No. Paspor</label>
        <input id="c-paspor" type="text" bind:value={form.no_paspor} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
      </div>
    </div>
    <div class="flex flex-col gap-1">
      <label for="c-alamat" class="text-xs font-medium text-slate-700">Alamat</label>
      <input id="c-alamat" type="text" bind:value={form.alamat} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
    </div>
    <div class="flex flex-col gap-1">
      <label for="c-lead" class="text-xs font-medium text-slate-700">Sumber</label>
      <select id="c-lead" bind:value={form.lead_source} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400">
        {#each LEAD_SOURCES as s}<option value={s.id}>{s.label}</option>{/each}
      </select>
    </div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => (showCreate = false)} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600 hover:bg-slate-50">Batal</button>
      <button type="button" onclick={saveProfile} disabled={saving} class="flex flex-1 items-center justify-center gap-2 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">
        {#if saving}<Loader2 class="h-4 w-4 animate-spin" />{/if}
        Simpan
      </button>
    </div>
  </div>
</SlideDrawer>

<!-- Jamaah Detail Drawer -->
<SlideDrawer
  open={drawerOpen}
  title={selectedJamaah?.name || ''}
  width="560px"
  onClose={() => (drawerOpen = false)}
>
  {#if selectedJamaah}
    <div class="flex flex-col h-full">
      <!-- Status row -->
      <div class="flex items-center justify-between border-b border-slate-100 px-6 py-3">
        <StatusBadge status={selectedJamaah.pipeline_status} />
        <div class="flex gap-2">
          <button
            type="button"
            onclick={() => openWhatsApp(selectedJamaah.phone)}
            class="flex items-center gap-1.5 rounded-lg bg-emerald-50 px-3 py-1.5 text-xs font-semibold text-emerald-700 transition-colors hover:bg-emerald-100"
          >
            <Phone class="h-3.5 w-3.5" />
            WhatsApp
          </button>
          <button
            type="button"
            onclick={() => { onNavigate?.('invoices'); drawerOpen = false; }}
            class="flex items-center gap-1.5 rounded-lg bg-primary-50 px-3 py-1.5 text-xs font-semibold text-primary-700 transition-colors hover:bg-primary-100"
          >
            Invoice
          </button>
        </div>
      </div>

      <!-- Tabs -->
      <div class="flex border-b border-slate-100">
        {#each DRAWER_TABS as tab}
          <button
            type="button"
            onclick={() => (activeTab = tab.id)}
            class="flex-1 py-3 text-xs font-semibold transition-colors
              {activeTab === tab.id
                ? 'border-b-2 border-primary-600 text-primary-700'
                : 'text-slate-400 hover:text-slate-600'}"
          >
            {tab.label}
          </button>
        {/each}
      </div>

      <!-- Tab content -->
      <div class="flex-1 overflow-y-auto p-6">
        {#if activeTab === 'profil'}
          <div class="space-y-4">
            <div class="rounded-xl border border-slate-100 p-4 space-y-2.5">
              <h4 class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Identitas</h4>
              {@render InfoRow("Nama Lengkap", selectedJamaah.name)}
              {@render InfoRow("NIK", selectedJamaah.nik || '—')}
              {@render InfoRow("No. Paspor", selectedJamaah.passport_no || '—')}
              {@render InfoRow("Jenis Kelamin", selectedJamaah.gender === 'L' ? 'Laki-laki' : selectedJamaah.gender === 'P' ? 'Perempuan' : '—')}
            </div>
            <div class="rounded-xl border border-slate-100 p-4 space-y-2.5">
              <h4 class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Kontak</h4>
              {@render InfoRow("HP / WhatsApp", selectedJamaah.phone || '—')}
              {@render InfoRow("Email", selectedJamaah.email || '—')}
            </div>
            <div class="rounded-xl border border-slate-100 p-4 space-y-2.5">
              <h4 class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Paket</h4>
              {@render InfoRow("Paket", selectedJamaah.package_name || 'Belum dipilih')}
              {@render InfoRow("Tipe Kamar", selectedJamaah.room_type || '—')}
            </div>
          </div>
        {:else if activeTab === 'invoice'}
          <div class="space-y-3">
            <div class="rounded-xl bg-slate-50 p-4">
              <p class="text-xs text-slate-400">Total Tagihan</p>
              <p class="text-xl font-bold text-slate-800">{formatIDR(selectedJamaah.total_invoice)}</p>
            </div>
            <div class="rounded-xl bg-emerald-50 p-4">
              <p class="text-xs text-emerald-600">Sudah Dibayar</p>
              <p class="text-xl font-bold text-emerald-700">{formatIDR(selectedJamaah.paid)}</p>
            </div>
            {#if selectedJamaah.sisa_tagihan > 0}
              <div class="rounded-xl bg-red-50 p-4">
                <p class="text-xs text-red-500">Sisa Tagihan</p>
                <p class="text-xl font-bold text-red-600">{formatIDR(selectedJamaah.sisa_tagihan)}</p>
              </div>
            {/if}
            <button
              type="button"
              onclick={() => { onNavigate?.('invoices'); drawerOpen = false; }}
              class="w-full rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-700"
            >
              Lihat Invoice Lengkap
            </button>
          </div>
        {:else if activeTab === 'catatan'}
          <p class="rounded-xl bg-slate-50 p-4 text-sm text-slate-500">Catatan internal akan tersedia di pembaruan berikutnya.</p>
        {/if}
      </div>
    </div>
  {/if}
</SlideDrawer>

{#snippet InfoRow(label, value)}
  <div class="flex items-start justify-between gap-4 text-sm">
    <span class="flex-shrink-0 text-slate-400">{label}</span>
    <span class="text-right font-medium text-slate-700">{value}</span>
  </div>
{/snippet}
