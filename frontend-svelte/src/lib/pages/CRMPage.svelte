<script>
  import { onMount } from 'svelte';
  import {
    Plus, Search, LayoutGrid, List,
    Phone, ChevronRight, UserCircle, Loader2,
    Users, CheckCircle, CreditCard, Clock, Package as PackageIcon,
    Flame, BarChart3, ArrowDownUp, X,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import EmptyState from '../components/EmptyState.svelte';
  import Pager from '../components/Pager.svelte';
  import Avatar from '../components/Avatar.svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import Button from '../components/ui/Button.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
  import ProgressBar from '../components/ui/ProgressBar.svelte';
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
  let viewMode = $state('kanban'); // 'table' | 'kanban'
  let drawerOpen = $state(false);
  let selectedJamaah = $state(null);
  let activeTab = $state('profil');

  // Kanban drag-and-drop (optimistic; persisted via PATCH .../status)
  let dragId = $state(null);
  let dragFrom = $state(null);
  let overCol = $state(null);

  // Lead-temperature filter + score sort (server-side) and funnel analytics
  let tempFilter = $state(''); // '' | 'hot' | 'warm' | 'cold'
  let sortByScore = $state(false);
  let showFunnel = $state(false);
  let funnel = $state(null);
  let funnelLoading = $state(false);

  // "Mark as lost" flow (prompts a reason when dropping into Batal)
  let showLostModal = $state(false);
  let lostReason = $state('');
  let pendingBatal = $state(null);

  const TEMP_FILTERS = [
    { id: '',     label: 'Semua' },
    { id: 'hot',  label: '🔥 Hot' },
    { id: 'warm', label: 'Warm' },
    { id: 'cold', label: 'Cold' },
  ];

  const LOST_REASONS = [
    { id: 'harga',      label: 'Harga terlalu mahal' },
    { id: 'jadwal',     label: 'Jadwal tidak cocok' },
    { id: 'kompetitor', label: 'Pilih travel lain' },
    { id: 'dana',       label: 'Kendala dana' },
    { id: 'tidak_jadi', label: 'Batal berangkat' },
    { id: 'lainnya',    label: 'Lainnya' },
  ];

  function tempColor(t) {
    return t === 'hot' ? 'var(--c-danger)' : t === 'warm' ? 'var(--c-warning)' : 'var(--c-muted)';
  }
  function tempBg(t) {
    return t === 'hot' ? 'var(--c-danger-soft)' : t === 'warm' ? 'var(--c-warning-soft, #fef3c7)' : 'var(--c-bg-2)';
  }

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
    { id: 'all',       label: 'Semua',     color: 'var(--c-muted)' },
    { id: 'prospek',   label: 'Prospek',   color: '#2563c9' },
    { id: 'survey',    label: 'Survey',    color: 'var(--c-accent)' },
    { id: 'booking',   label: 'Booking',   color: '#7a5ae0' },
    { id: 'dp',        label: 'DP',        color: 'var(--c-warning)' },
    { id: 'cicilan',   label: 'Cicilan',   color: 'var(--c-info)' },
    { id: 'lunas',     label: 'Lunas',     color: 'var(--c-success)' },
    { id: 'berangkat', label: 'Berangkat', color: '#0f7a5a' },
    { id: 'batal',     label: 'Batal',     color: 'var(--c-danger)' },
  ];

  // Kanban columns = pipeline stages excluding the "all" filter pseudo-stage.
  let kanbanCols = $derived(PIPELINE_STATUSES.filter(s => s.id !== 'all'));

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
        ApiService.listCRM({
          search: searchQuery, page, pageSize: PAGE_SIZE,
          temp: tempFilter, sort: sortByScore ? 'score' : '',
        }),
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
        package_id: r.package_id || null,
        package_name: r.package_id ? (pkgMap.get(r.package_id) || 'Paket') : null,
        room_type: r.room_type || null,
        pipeline_status: r.pipeline_status || 'prospek',
        total_invoice: r.total_amount || 0,
        paid: r.total_paid || 0,
        sisa_tagihan: r.total_remaining || 0,
        lead_score: r.lead_score ?? null,
        lead_temp: r.lead_temp || 'cold',
        days_in_stage: r.days_in_stage || 0,
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

  function cardsFor(colId) {
    return jamaah.filter(j => j.pipeline_status === colId);
  }

  // ── Temperature filter / score sort / funnel ──
  function setTempFilter(t) {
    if (tempFilter === t) return;
    tempFilter = t;
    page = 1;
    loadJamaah();
  }

  function toggleSort() {
    sortByScore = !sortByScore;
    page = 1;
    loadJamaah();
  }

  async function toggleFunnel() {
    showFunnel = !showFunnel;
    if (showFunnel && !funnel) await loadFunnel();
  }

  async function loadFunnel() {
    funnelLoading = true;
    try {
      funnel = await ApiService.getPipelineFunnel();
    } catch (e) {
      showToast(mapError(e.message), 'error');
      showFunnel = false;
    } finally {
      funnelLoading = false;
    }
  }

  function stageLabel(id) {
    return PIPELINE_STATUSES.find(s => s.id === id)?.label || id;
  }
  function stageColor(id) {
    return PIPELINE_STATUSES.find(s => s.id === id)?.color || 'var(--c-muted)';
  }

  // ── Kanban drag-and-drop (optimistic; persisted to server) ──
  function onDragStart(j) {
    dragId = j.id;
    dragFrom = j.pipeline_status;
  }

  function onDragEnd() {
    dragId = null;
    dragFrom = null;
    overCol = null;
  }

  function onDrop(toCol) {
    const id = dragId, from = dragFrom;
    onDragEnd();
    if (!id || from === toCol) return;
    const idx = jamaah.findIndex(j => j.id === id);
    if (idx === -1) return;
    const j = jamaah[idx];
    if (!j.package_id) {
      showToast('Lead belum terdaftar di paket — daftarkan ke paket dulu sebelum memindah tahap.', 'warning');
      return;
    }
    if (toCol === 'batal') {
      // Capture why the lead was lost before committing the move.
      pendingBatal = { id, idx, packageId: j.package_id, from };
      lostReason = '';
      showLostModal = true;
      return;
    }
    persistStage(idx, j.package_id, toCol, '');
  }

  async function persistStage(idx, packageId, toCol, lost_reason) {
    const prev = jamaah[idx];
    jamaah[idx] = { ...prev, pipeline_status: toCol }; // optimistic
    try {
      await ApiService.updatePipelineStatus(prev.id, packageId, { pipeline_status: toCol, lost_reason });
      showToast(`Lead dipindahkan ke ${stageLabel(toCol)}`, 'success');
    } catch (e) {
      jamaah[idx] = prev; // rollback
      showToast(mapError(e.message), 'error');
    }
  }

  function confirmBatal() {
    if (!pendingBatal) return;
    const { idx, packageId } = pendingBatal;
    showLostModal = false;
    persistStage(idx, packageId, 'batal', lostReason || '');
    pendingBatal = null;
  }

  function cancelBatal() {
    showLostModal = false;
    pendingBatal = null;
  }
</script>

<div class="flex h-full min-h-0 flex-col" style="background:var(--c-bg)">
  <div class="flex-shrink-0 px-6 pt-6">
    <PageHeader
      title="CRM — Pipeline Penjualan"
      subtitle="Seret kartu jamaah antar tahap untuk memperbarui status. Pantau progres pembayaran tiap lead."
    >
      {#snippet actions()}
        <!-- View toggle -->
        <div style="display:inline-flex;gap:4px;background:var(--c-bg-2);padding:4px;border-radius:var(--radius)">
          <button
            type="button"
            onclick={() => (viewMode = 'kanban')}
            class="crm-toggle"
            class:crm-toggle-on={viewMode === 'kanban'}
            title="Tampilan kanban"
          >
            <LayoutGrid class="h-4 w-4" />
          </button>
          <button
            type="button"
            onclick={() => (viewMode = 'table')}
            class="crm-toggle"
            class:crm-toggle-on={viewMode === 'table'}
            title="Tampilan tabel"
          >
            <List class="h-4 w-4" />
          </button>
        </div>
        <Button variant="primary" icon={Plus} onclick={openCreate}>Tambah Jamaah</Button>
      {/snippet}
    </PageHeader>

    <!-- Stat tiles -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {#each statTiles as t}
        <StatCard icon={t.icon} label={t.label} value={String(t.value)} accent={t.accent} />
      {/each}
    </div>

    <!-- Search + stage filter -->
    <div class="mt-5 flex flex-wrap items-center gap-3">
      <div class="relative min-w-[220px] flex-1">
        <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" style="color:var(--c-faint)" />
        <input
          type="text"
          bind:value={searchQuery}
          oninput={onSearchInput}
          placeholder="Cari nama, HP, NIK, paspor..."
          class="crm-search"
        />
      </div>
      <div class="overflow-x-auto">
        <FilterTabs
          tabs={PIPELINE_STATUSES.map(s => ({ value: s.id, label: s.label }))}
          value={filterStatus}
          onChange={(v) => (filterStatus = v)}
        />
      </div>
    </div>

    <!-- Lead-temperature filter + score sort + funnel toggle -->
    <div class="mt-3 flex flex-wrap items-center gap-2">
      <div style="display:inline-flex;gap:4px;background:var(--c-bg-2);padding:4px;border-radius:var(--radius)">
        {#each TEMP_FILTERS as t}
          <button
            type="button"
            onclick={() => setTempFilter(t.id)}
            class="crm-chip"
            class:crm-chip-on={tempFilter === t.id}
          >{t.label}</button>
        {/each}
      </div>
      <button
        type="button"
        onclick={toggleSort}
        class="crm-pillbtn"
        class:crm-pillbtn-on={sortByScore}
        title="Urutkan berdasarkan skor lead tertinggi"
      >
        <ArrowDownUp class="h-3.5 w-3.5" />
        Skor tertinggi
      </button>
      <button
        type="button"
        onclick={toggleFunnel}
        class="crm-pillbtn"
        class:crm-pillbtn-on={showFunnel}
        title="Analitik funnel pipeline"
      >
        <BarChart3 class="h-3.5 w-3.5" />
        Funnel
      </button>
    </div>

    {#if showFunnel}
      <div class="mt-3 rounded-2xl p-4" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
        {#if funnelLoading}
          <div class="flex items-center gap-2 text-sm" style="color:var(--c-faint)">
            <Loader2 class="h-4 w-4 animate-spin" /> Memuat analitik…
          </div>
        {:else if funnel}
          <div class="mb-3 flex items-center gap-2">
            <BarChart3 class="h-4 w-4" style="color:var(--c-primary)" />
            <h3 class="text-sm font-bold" style="color:var(--c-ink)">Funnel Pipeline</h3>
            <span class="text-xs" style="color:var(--c-faint)">{funnel.total} registrasi</span>
          </div>
          <div class="grid grid-cols-2 gap-2 md:grid-cols-3 lg:grid-cols-5">
            {#each funnel.stages as st}
              <div class="rounded-xl p-3" style="background:var(--c-bg-2);border-left:3px solid {stageColor(st.stage)}">
                <p class="text-[11px] font-bold uppercase tracking-wide" style="color:var(--c-faint)">{stageLabel(st.stage)}</p>
                <p class="mt-0.5 text-lg font-extrabold" style="color:var(--c-ink)">{st.count}</p>
                {#if st.total_value > 0}
                  <p class="text-[11px] tabular" style="font-variant-numeric:tabular-nums;color:var(--c-muted)">{formatIDR(st.total_value)}</p>
                {/if}
                {#if st.count > 0}
                  <p class="mt-1 text-[10px]" style="color:var(--c-faint)">⌀ {Math.round(st.avg_days_in_stage)} hari di tahap</p>
                {/if}
              </div>
            {/each}
          </div>
          {#if funnel.sources?.length}
            <div class="mt-3 flex flex-wrap items-center gap-2">
              <span class="text-[11px] font-bold uppercase tracking-wide" style="color:var(--c-faint)">Sumber:</span>
              {#each funnel.sources as s}
                <span class="rounded-full px-2.5 py-0.5 text-[11px] font-semibold" style="background:var(--c-bg-2);color:var(--c-muted)">{s.source} · {s.count}</span>
              {/each}
            </div>
          {/if}
        {/if}
      </div>
    {/if}
  </div>

  <!-- Body -->
  {#if isLoading}
    <div class="flex-1 overflow-auto p-6">
      <div class="space-y-3">
        {#each [1,2,3,4] as _}
          <div class="h-14 animate-pulse rounded-2xl" style="background:var(--c-bg-2)"></div>
        {/each}
      </div>
    </div>
  {:else if error}
    <div class="m-6 rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{mapError(error)}</div>

  <!-- Kanban view -->
  {:else if viewMode === 'kanban'}
    <div class="min-h-0 flex-1 overflow-x-auto px-6 py-5">
      <div class="flex h-full items-start gap-3.5">
        {#each kanbanCols as col}
          {@const colItems = cardsFor(col.id)}
          <div
            role="list"
            class="flex h-full min-h-[120px] w-[260px] flex-shrink-0 flex-col rounded-2xl p-2.5 transition-colors"
            style="background:{overCol === col.id ? col.color + '10' : 'var(--c-bg-2)'};border:2px solid {overCol === col.id ? col.color : 'transparent'}"
            ondragover={(e) => { e.preventDefault(); overCol = col.id; }}
            ondragleave={() => { if (overCol === col.id) overCol = null; }}
            ondrop={() => onDrop(col.id)}
          >
            <!-- Column header -->
            <div class="flex items-center gap-2 px-1.5 pb-3 pt-1.5">
              <span class="h-2.5 w-2.5 flex-shrink-0 rounded-full" style="background:{col.color}"></span>
              <span class="text-[13.5px] font-extrabold" style="color:var(--c-ink)">{col.label}</span>
              <span class="rounded-full px-2 py-0.5 text-[11px] font-bold" style="background:var(--c-surface);color:var(--c-faint)">{colItems.length}</span>
            </div>

            <!-- Cards -->
            <div class="flex min-h-[60px] flex-1 flex-col gap-2.5 overflow-y-auto">
              {#each colItems as j (j.id)}
                <div
                  role="button"
                  tabindex="0"
                  draggable="true"
                  ondragstart={() => onDragStart(j)}
                  ondragend={onDragEnd}
                  onclick={() => openDetail(j)}
                  onkeydown={(e) => { if (e.key === 'Enter') openDetail(j); }}
                  class="cursor-grab rounded-xl p-3 text-left transition-shadow active:cursor-grabbing"
                  style="background:var(--c-surface);border:1px solid var(--c-line);border-left:3px solid {col.color};box-shadow:var(--shadow-sm);opacity:{dragId === j.id ? 0.4 : 1}"
                >
                  <div class="mb-2 flex items-center gap-2.5">
                    <Avatar name={j.name} size={30} />
                    <div class="min-w-0 flex-1">
                      <p class="truncate text-[13px] font-bold" style="color:var(--c-ink)">{j.name}</p>
                      <p class="truncate text-[11px]" style="color:var(--c-faint)">{j.nik || j.phone || '—'}</p>
                    </div>
                    {@render scoreBadge(j)}
                  </div>

                  {#if j.package_name}
                    <div class="mb-1 flex items-center gap-1.5 text-[12px]" style="color:var(--c-muted)">
                      <PackageIcon class="h-3 w-3 flex-shrink-0" />
                      <span class="truncate">{j.package_name}</span>
                    </div>
                  {/if}

                  {#if j.days_in_stage > 0}
                    <p class="mb-1.5 text-[10px]" style="color:var(--c-faint)">{j.days_in_stage} hari di tahap</p>
                  {/if}

                  {#if j.total_invoice > 0}
                    <div class="mb-1 flex items-center justify-between text-[11px]">
                      <span class="tabular font-bold" style="font-variant-numeric:tabular-nums;color:var(--c-ink)">{formatIDR(j.paid)}</span>
                      <span class="tabular" style="font-variant-numeric:tabular-nums;color:var(--c-faint)">{formatIDR(j.total_invoice)}</span>
                    </div>
                    <ProgressBar
                      value={j.paid}
                      max={j.total_invoice}
                      height={6}
                      color={j.sisa_tagihan <= 0 ? 'var(--c-success)' : 'var(--c-accent)'}
                    />
                    {#if j.sisa_tagihan > 0}
                      <p class="mt-2 text-[11px] font-semibold" style="color:var(--c-danger)">Sisa {formatIDR(j.sisa_tagihan)}</p>
                    {:else}
                      <p class="mt-2 text-[11px] font-semibold" style="color:var(--c-success)">Lunas</p>
                    {/if}
                  {/if}
                </div>
              {/each}
              {#if colItems.length === 0}
                <p class="px-1 py-2 text-[11px]" style="color:var(--c-faint)">Belum ada</p>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </div>

  <!-- Table view -->
  {:else}
    <div class="min-h-0 flex-1 overflow-auto px-6 py-5">
      {#if filtered.length === 0}
        <EmptyState
          icon={UserCircle}
          title={searchQuery || filterStatus !== 'all' ? 'Tidak ada jamaah yang cocok' : 'Belum ada jamaah'}
          text={searchQuery || filterStatus !== 'all' ? 'Coba ubah kata kunci atau filter.' : 'Klik “Tambah Jamaah” untuk menambah data, atau import dari AI Scanner.'}
        />
      {:else}
        <div class="overflow-hidden rounded-2xl" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
          <table class="w-full">
            <thead style="background:var(--c-bg-2)">
              <tr class="text-left text-xs font-semibold uppercase tracking-wider" style="color:var(--c-faint)">
                <th class="px-6 py-3">Jamaah</th>
                <th class="hidden px-4 py-3 md:table-cell">Paket</th>
                <th class="hidden px-4 py-3 lg:table-cell">Pembayaran</th>
                <th class="hidden px-4 py-3 sm:table-cell">Skor</th>
                <th class="px-4 py-3">Status</th>
                <th class="px-4 py-3 text-right">Sisa Tagihan</th>
                <th class="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody style="--tw-divide-opacity:1">
              {#each filtered as j (j.id)}
                <tr class="group transition-colors" style="border-top:1px solid var(--c-line-soft)">
                  <td class="px-6 py-3.5">
                    <div class="flex items-center gap-3">
                      <Avatar name={j.name} size={38} />
                      <div class="min-w-0">
                        <p class="truncate text-sm font-bold" style="color:var(--c-ink)">{j.name}</p>
                        <p class="truncate text-xs" style="color:var(--c-faint)">{j.nik || j.phone || '—'}</p>
                      </div>
                    </div>
                  </td>
                  <td class="hidden px-4 py-3.5 md:table-cell">
                    <p class="text-sm font-medium" style="color:var(--c-ink-soft)">{j.package_name || '—'}</p>
                    {#if j.room_type}<p class="text-xs" style="color:var(--c-faint)">{j.room_type}</p>{/if}
                  </td>
                  <td class="hidden px-4 py-3.5 lg:table-cell">
                    {#if j.total_invoice > 0}
                      <div class="min-w-[140px]">
                        <div class="mb-1 flex justify-between text-xs">
                          <span class="tabular font-bold" style="font-variant-numeric:tabular-nums;color:var(--c-ink)">{formatIDR(j.paid)}</span>
                          <span class="tabular" style="font-variant-numeric:tabular-nums;color:var(--c-faint)">{formatIDR(j.total_invoice)}</span>
                        </div>
                        <ProgressBar
                          value={j.paid}
                          max={j.total_invoice}
                          height={6}
                          color={j.sisa_tagihan <= 0 ? 'var(--c-success)' : 'var(--c-accent)'}
                        />
                      </div>
                    {:else}
                      <span class="text-xs" style="color:var(--c-faint)">—</span>
                    {/if}
                  </td>
                  <td class="hidden px-4 py-3.5 sm:table-cell">
                    {#if j.lead_score != null}
                      {@render scoreBadge(j)}
                      {#if j.days_in_stage > 0}<p class="mt-1 text-[10px]" style="color:var(--c-faint)">{j.days_in_stage} hr</p>{/if}
                    {:else}
                      <span class="text-xs" style="color:var(--c-faint)">—</span>
                    {/if}
                  </td>
                  <td class="px-4 py-3.5">
                    <StatusBadge status={j.pipeline_status} size="xs" />
                  </td>
                  <td class="px-4 py-3.5 text-right">
                    {#if j.sisa_tagihan > 0}
                      <span class="text-sm font-semibold" style="color:var(--c-danger)">{formatIDR(j.sisa_tagihan)}</span>
                    {:else if j.total_invoice > 0}
                      <span class="text-sm font-semibold" style="color:var(--c-success)">Lunas</span>
                    {:else}
                      <span class="text-sm" style="color:var(--c-faint)">—</span>
                    {/if}
                  </td>
                  <td class="px-4 py-3.5 text-right">
                    <button
                      type="button"
                      onclick={() => openDetail(j)}
                      class="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors"
                      style="color:var(--c-primary)"
                    >
                      Detail
                      <ChevronRight class="h-3 w-3" />
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
        <Pager {page} pageSize={PAGE_SIZE} {total} onchange={gotoPage} />
      {/if}
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

{#snippet scoreBadge(j)}
  {#if j.lead_score != null}
    <span
      class="inline-flex flex-shrink-0 items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-extrabold"
      style="background:{tempBg(j.lead_temp)};color:{tempColor(j.lead_temp)}"
      title="Skor lead: {j.lead_score}/100 ({j.lead_temp})"
    >
      {#if j.lead_temp === 'hot'}<Flame class="h-3 w-3" />{:else}<span class="h-1.5 w-1.5 rounded-full" style="background:{tempColor(j.lead_temp)}"></span>{/if}
      {j.lead_score}
    </span>
  {/if}
{/snippet}

<!-- Mark-as-lost reason modal -->
{#if showLostModal}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4"
    style="background:rgba(0,0,0,0.4)"
    role="button"
    tabindex="-1"
    onclick={cancelBatal}
    onkeydown={(e) => { if (e.key === 'Escape') cancelBatal(); }}
  >
    <div
      class="w-full max-w-sm rounded-2xl p-5"
      style="background:var(--c-surface);box-shadow:var(--shadow-lg, 0 10px 40px rgba(0,0,0,0.2))"
      role="dialog"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={() => {}}
    >
      <div class="mb-3 flex items-center justify-between">
        <h3 class="text-sm font-bold" style="color:var(--c-ink)">Tandai lead sebagai Batal</h3>
        <button type="button" onclick={cancelBatal} style="color:var(--c-faint)"><X class="h-4 w-4" /></button>
      </div>
      <p class="mb-3 text-xs" style="color:var(--c-muted)">Pilih alasan agar bisa dianalisis nanti.</p>
      <div class="flex flex-col gap-1.5">
        {#each LOST_REASONS as r}
          <button
            type="button"
            onclick={() => (lostReason = r.label)}
            class="rounded-xl px-3 py-2 text-left text-sm transition-colors"
            style="border:1px solid {lostReason === r.label ? 'var(--c-primary)' : 'var(--c-line)'};background:{lostReason === r.label ? 'var(--c-primary-tint)' : 'transparent'};color:var(--c-ink)"
          >{r.label}</button>
        {/each}
      </div>
      <div class="mt-4 flex gap-2">
        <button type="button" onclick={cancelBatal} class="flex-1 rounded-xl border py-2.5 text-sm font-semibold" style="border-color:var(--c-line);color:var(--c-muted)">Batal</button>
        <button type="button" onclick={confirmBatal} class="flex-1 rounded-xl py-2.5 text-sm font-semibold text-white" style="background:var(--c-danger)">Tandai Batal</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .crm-toggle {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    color: var(--c-muted);
    transition: all 0.15s;
  }
  .crm-toggle:hover { color: var(--c-ink); }
  .crm-toggle-on {
    background: var(--c-surface);
    color: var(--c-primary);
    box-shadow: var(--shadow-sm);
  }
  .crm-search {
    width: 100%;
    padding: 10px 12px 10px 36px;
    font-size: 14px;
    border-radius: var(--radius);
    border: 1px solid var(--c-line);
    background: var(--c-surface);
    color: var(--c-ink);
    outline: none;
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .crm-search::placeholder { color: var(--c-faint); }
  .crm-search:focus {
    border-color: var(--c-primary);
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }
  tbody tr.group:hover { background: var(--c-primary-tint); }

  .crm-chip {
    padding: 5px 12px;
    font-size: 12.5px;
    font-weight: 700;
    border-radius: var(--radius-sm);
    color: var(--c-muted);
    white-space: nowrap;
    transition: all 0.15s;
  }
  .crm-chip:hover { color: var(--c-ink); }
  .crm-chip-on {
    background: var(--c-surface);
    color: var(--c-primary);
    box-shadow: var(--shadow-sm);
  }
  .crm-pillbtn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 7px 12px;
    font-size: 12.5px;
    font-weight: 700;
    border-radius: var(--radius);
    border: 1px solid var(--c-line);
    background: var(--c-surface);
    color: var(--c-muted);
    white-space: nowrap;
    transition: all 0.15s;
  }
  .crm-pillbtn:hover { color: var(--c-ink); border-color: var(--c-primary); }
  .crm-pillbtn-on {
    background: var(--c-primary-tint);
    color: var(--c-primary);
    border-color: var(--c-primary);
  }
</style>
