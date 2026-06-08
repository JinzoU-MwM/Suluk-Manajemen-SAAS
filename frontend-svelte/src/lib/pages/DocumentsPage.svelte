<script>
  import { onMount } from 'svelte';
  import {
    ClipboardList, Search, AlertCircle, CheckCircle, Upload,
    ChevronRight, FileText,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { ApiService } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  const DOC_TYPES = [
    { id: 'ktp', label: 'KTP' },
    { id: 'kk', label: 'Kartu Keluarga' },
    { id: 'paspor', label: 'Paspor' },
    { id: 'pas_foto', label: 'Pas Foto' },
    { id: 'icv', label: 'ICV' },
    { id: 'visa', label: 'Visa' },
    { id: 'formulir', label: 'Formulir' },
    { id: 'akta_nikah', label: 'Akta Nikah' },
    { id: 'akta_lahir', label: 'Akta Lahir' },
    { id: 'surat_mahram', label: 'Surat Mahram' },
    { id: 'surat_rekomendasi', label: 'Surat Rekomendasi' },
    { id: 'other', label: 'Lainnya' },
  ];

    const FILTERS = [
    { id: 'all', label: 'Semua' },
    { id: 'incomplete', label: 'Belum Lengkap' },
    { id: 'complete', label: 'Lengkap' },
  ];

  let jamaahList = $state([]);
  let isLoading = $state(true);
  let searchQuery = $state('');
  let filterStatus = $state('all');

  // Summary
  let totalCount = $state(0);
  let incompleteCount = $state(0);
  let passportExpiring = $state(0);

  // Detail
  let drawerOpen = $state(false);
  let selectedJamaah = $state(null);
  let selectedDocs = $state([]);
  let docsLoading = $state(false);

  // Upload
  let showUploadForm = $state(false);
  let uploadDocType = $state('');
  let uploadFile = $state(null);
  let uploading = $state(false);

  let filtered = $derived(
    jamaahList.filter(j => {
      if (filterStatus === 'incomplete') {
        const completed = (j.documents || []).filter(d => d.status === 'selesai').length;
        if (completed >= DOC_TYPES.length) return false;
      }
      if (filterStatus === 'complete') {
        const completed = (j.documents || []).filter(d => d.status === 'selesai').length;
        if (completed < DOC_TYPES.length) return false;
      }
      if (!searchQuery) return true;
      const q = searchQuery.toLowerCase();
      return j.nama.toLowerCase().includes(q) || (j.no_paspor && j.no_paspor.toLowerCase().includes(q));
    })
  );

  onMount(loadData);

  async function loadData() {
    isLoading = true;
    try {
      const data = await ApiService.listJamaah({ pageSize: 200 });
      const rawList = Array.isArray(data) ? data : (data?.data ?? []);
      // Load documents for each jamaah
      const withDocs = await Promise.all(
        rawList.map(async (j) => {
          try {
            const docs = await ApiService.listDocuments(j.id);
            j.documents = Array.isArray(docs) ? docs : [];
          } catch {
            j.documents = [];
          }
          return j;
        })
      );
      jamaahList = withDocs;
      totalCount = withDocs.length;
      incompleteCount = withDocs.filter(j => {
        const completed = (j.documents || []).filter(d => d.status === 'selesai').length;
        return completed < DOC_TYPES.length;
      }).length;

      try {
        const alerts = await ApiService.getDashboardAlerts();
        passportExpiring = alerts?.passport_expiring_90?.length || 0;
      } catch { /* non-critical */ }
    } catch (e) {
      showToast(e.message || 'Gagal memuat data dokumen', 'error');
    } finally {
      isLoading = false;
    }
  }

  async function openDetail(j) {
    selectedJamaah = j;
    drawerOpen = true;
    showUploadForm = false;
    docsLoading = true;
    try {
      const docs = await ApiService.listDocuments(j.id);
      selectedDocs = Array.isArray(docs) ? docs : [];
    } catch {
      selectedDocs = [];
    } finally {
      docsLoading = false;
    }
  }

  function getDocStatus(docType) {
    return selectedDocs.find(d => d.doc_type === docType);
  }

  function getDocLabel(docType) {
    return DOC_TYPES.find(d => d.id === docType)?.label || docType;
  }

  function completedCount(docs) {
    return (docs || []).filter(d => d.status === 'selesai').length;
  }

  async function handleUpload() {
    if (!uploadDocType) { showToast('Pilih jenis dokumen', 'warning'); return; }
    if (!uploadFile) { showToast('Pilih file', 'warning'); return; }
    uploading = true;
    try {
      await ApiService.uploadDocument(selectedJamaah.id, { doc_type: uploadDocType }, uploadFile);
      showToast('Dokumen berhasil diupload', 'success');
      showUploadForm = false;
      uploadFile = null;
      uploadDocType = '';
      const docs = await ApiService.listDocuments(selectedJamaah.id);
      selectedDocs = Array.isArray(docs) ? docs : [];
    } catch (e) {
      showToast(e.message || 'Gagal upload dokumen', 'error');
    } finally {
      uploading = false;
    }
  }

  async function updateStatus(docId, status) {
    try {
      await ApiService.updateDocumentStatus(selectedJamaah.id, docId, { status });
      showToast('Status dokumen diperbarui', 'success');
      const docs = await ApiService.listDocuments(selectedJamaah.id);
      selectedDocs = Array.isArray(docs) ? docs : [];
    } catch (e) {
      showToast(e.message || 'Gagal update status', 'error');
    }
  }

  function formatDate(d) {
    if (!d) return '—';
    return new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
  }
</script>

<div class="flex h-screen flex-col">
  <!-- Header -->
  <div class="flex-shrink-0 border-b border-slate-100 bg-white px-6 py-5">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-800">Dokumen & Paspor</h1>
        <p class="mt-0.5 text-sm text-slate-500">Checklist kelengkapan dokumen jamaah</p>
      </div>
    </div>

    <!-- Summary -->
    <div class="mt-4 grid grid-cols-3 gap-3">
      <div class="rounded-xl bg-blue-50 p-3">
        <p class="text-[11px] font-semibold text-blue-400">Total Jamaah</p>
        <p class="mt-0.5 text-base font-bold text-blue-700">
          {#if isLoading}<span class="inline-block h-5 w-12 animate-pulse rounded bg-blue-200"></span>{:else}{totalCount}{/if}
        </p>
      </div>
      <div class="rounded-xl bg-amber-50 p-3">
        <p class="text-[11px] font-semibold text-amber-400">Dokumen Belum Lengkap</p>
        <p class="mt-0.5 text-base font-bold text-amber-700">
          {#if isLoading}<span class="inline-block h-5 w-16 animate-pulse rounded bg-amber-200"></span>{:else}{incompleteCount}{/if}
        </p>
      </div>
      <div class="rounded-xl bg-red-50 p-3">
        <p class="text-[11px] font-semibold text-red-400">Paspor Akan Habis</p>
        <p class="mt-0.5 text-base font-bold text-red-700">
          {#if isLoading}<span class="inline-block h-5 w-12 animate-pulse rounded bg-red-200"></span>{:else}{passportExpiring}{/if}
        </p>
      </div>
    </div>

    <!-- Search + filter -->
    <div class="mt-4 flex gap-3">
      <div class="relative flex-1 min-w-0">
        <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Cari nama jamaah atau no. paspor..."
          class="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-9 pr-3 text-sm outline-none focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>
      <div class="flex gap-1">
        {#each FILTERS as f}
          <button
            type="button"
            onclick={() => (filterStatus = f.id)}
            class="rounded-lg px-3 py-2 text-xs font-semibold transition-all
              {filterStatus === f.id ? 'bg-primary-600 text-white' : 'text-slate-500 hover:bg-slate-100'}"
          >
            {f.label}
          </button>
        {/each}
      </div>
    </div>
  </div>

  <!-- List -->
  <div class="flex-1 overflow-auto">
    {#if isLoading}
      <div class="space-y-3 p-6">
        {#each [1,2,3,4] as _}
          <div class="h-20 animate-pulse rounded-xl bg-slate-100"></div>
        {/each}
      </div>
    {:else if filtered.length === 0}
      <div class="flex flex-col items-center justify-center py-24 text-slate-400">
        <ClipboardList class="mb-3 h-12 w-12 opacity-30" />
        <p class="font-medium">Belum ada data jamaah</p>
      </div>
    {:else}
      <div class="space-y-2 p-6">
        {#each filtered as j}
          <button
            type="button"
            class="w-full text-left flex items-center justify-between rounded-xl bg-white p-4 shadow-sm ring-1 ring-slate-200/60 cursor-pointer transition-all hover:ring-primary-200 focus:outline-none focus:ring-2 focus:ring-primary-400"
            onclick={() => openDetail(j)}
          >
            <div class="flex items-center gap-4 min-w-0">
              <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-primary-50 text-sm font-bold text-primary-600">
                {j.nama?.charAt(0) || '?'}
              </div>
              <div class="min-w-0">
                <p class="text-sm font-semibold text-slate-800 truncate">{j.nama}</p>
                <p class="text-xs text-slate-400">
                  {completedCount(j.documents)}/{DOC_TYPES.length} dokumen
                  {#if j.no_paspor}
                    · Paspor: {j.no_paspor}
                  {/if}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-3 flex-shrink-0">
              {#if completedCount(j.documents) >= DOC_TYPES.length}
                <span class="flex items-center gap-1 text-xs font-semibold text-emerald-600">
                  <CheckCircle class="h-3.5 w-3.5" /> Lengkap
                </span>
              {:else}
                <span class="flex items-center gap-1 text-xs font-semibold text-amber-600">
                  <AlertCircle class="h-3.5 w-3.5" /> {DOC_TYPES.length - completedCount(j.documents)} kurang
                </span>
              {/if}
              <ChevronRight class="h-4 w-4 text-slate-300" />
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Detail Drawer -->
<SlideDrawer
  open={drawerOpen}
  title="Dokumen: {selectedJamaah?.nama || ''}"
  width="520px"
  onClose={() => { drawerOpen = false; selectedJamaah = null; }}
>
  {#if selectedJamaah}
    <div class="p-6 space-y-4">
      <!-- Jamaah info -->
      <div class="rounded-xl bg-slate-50 p-4 space-y-1">
        <p class="text-sm font-bold text-slate-800">{selectedJamaah.nama}</p>
        {#if selectedJamaah.no_paspor}
          <p class="text-xs text-slate-500">Paspor: {selectedJamaah.no_paspor}</p>
        {/if}
        {#if selectedJamaah.tanggal_paspor}
          <p class="text-xs text-slate-500">Terbit: {formatDate(selectedJamaah.tanggal_paspor)}</p>
        {/if}
      </div>

      <!-- Document checklist -->
      <div>
        <div class="mb-3 flex items-center justify-between">
          <h3 class="text-xs font-bold uppercase tracking-wider text-slate-400">Checklist Dokumen</h3>
          <button
            type="button"
            onclick={() => { showUploadForm = !showUploadForm; uploadDocType = ''; uploadFile = null; }}
            class="flex items-center gap-1 rounded-lg bg-primary-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-primary-700"
          >
            <Upload class="h-3.5 w-3.5" /> Upload
          </button>
        </div>

        {#if docsLoading}
          <div class="space-y-2">
            {#each [1,2,3] as _}
              <div class="h-12 animate-pulse rounded-lg bg-slate-100"></div>
            {/each}
          </div>
        {:else}
          <div class="space-y-1">
            {#each DOC_TYPES as dt}
              {@const doc = getDocStatus(dt.id)}
              <div class="flex items-center justify-between rounded-lg px-3 py-2.5 {doc ? 'bg-white' : 'bg-slate-50'} border border-slate-100">
                <div class="flex items-center gap-2.5 min-w-0">
                  <FileText class="h-4 w-4 flex-shrink-0 text-slate-400" />
                  <div class="min-w-0">
                    <p class="text-sm font-medium text-slate-700">{dt.label}</p>
                    {#if doc?.file_name}
                      <p class="text-xs text-slate-400 truncate">{doc.file_name}</p>
                    {/if}
                  </div>
                </div>
                <div class="flex items-center gap-2 flex-shrink-0">
                  {#if doc}
                    <select
                      value={doc.status}
                      onchange={(e) => updateStatus(doc.id, /** @type {HTMLSelectElement} */(e.target).value)}
                      class="rounded-lg border border-slate-200 px-2 py-1 text-xs outline-none focus:border-primary-400"
                    >
                      <option value="belum_diterima">Belum</option>
                      <option value="diterima">Diterima</option>
                      <option value="diproses">Diproses</option>
                      <option value="selesai">Selesai</option>
                    </select>
                    {#if doc.status === 'selesai'}
                      <CheckCircle class="h-4 w-4 text-emerald-500" />
                    {/if}
                  {:else}
                    <span class="text-xs text-slate-400">—</span>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Upload form -->
      {#if showUploadForm}
        <div class="rounded-xl border border-primary-200 bg-primary-50 p-4 space-y-3">
          <h4 class="text-xs font-bold text-primary-800">Upload Dokumen Baru</h4>

          <div class="flex flex-col gap-1">
            <label for="upload-doc-type" class="text-xs font-medium text-slate-700">Jenis Dokumen</label>
            <select
              id="upload-doc-type"
              bind:value={uploadDocType}
              class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-primary-400"
            >
              <option value="">Pilih jenis...</option>
              {#each DOC_TYPES as dt}
                <option value={dt.id}>{dt.label}</option>
              {/each}
            </select>
          </div>

          <div class="flex flex-col gap-1">
            <label for="upload-file-input" class="text-xs font-medium text-slate-700">File</label>
            <input
              id="upload-file-input"
              type="file"
              accept=".pdf,.jpg,.jpeg,.png"
              onchange={(e) => { const el = /** @type {HTMLInputElement} */ (e.target); uploadFile = el.files[0] || null; }}
              class="text-sm file:mr-3 file:rounded-lg file:border-0 file:bg-primary-100 file:px-3 file:py-1.5 file:text-xs file:font-semibold file:text-primary-700 hover:file:bg-primary-200"
            />
          </div>

          <div class="flex gap-2">
            <button
              type="button"
              onclick={() => { showUploadForm = false; uploadFile = null; uploadDocType = ''; }}
              class="flex-1 rounded-lg border border-slate-200 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-50"
            >
              Batal
            </button>
            <button
              type="button"
              onclick={handleUpload}
              disabled={uploading}
              class="flex-1 rounded-lg bg-primary-600 py-2 text-xs font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
            >
              {uploading ? 'Mengupload...' : 'Upload'}
            </button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</SlideDrawer>
