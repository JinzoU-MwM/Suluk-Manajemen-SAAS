<script>
  import { onMount } from 'svelte';
  import {
    Plus, Receipt, AlertCircle, CheckCircle, Clock,
    Download, Printer, Wallet, Check, X, Loader2, Search,
  } from 'lucide-svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import Avatar from '../components/Avatar.svelte';
  import EmptyState from '../components/EmptyState.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import IDRInput from '../components/IDRInput.svelte';
  import Pager from '../components/Pager.svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import Card from '../components/ui/Card.svelte';
  import Button from '../components/ui/Button.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
  import { showToast, mapError } from '../services/toast.svelte.js';
  import { formatRupiah as formatIDR, formatDate } from '../utils/formatting.js';
  import { invoiceApi } from '../services/apiDomains/invoiceApi.js';
  import { ApiService, authHeaders } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  // ── State ──────────────────────────────────────────────
  let invoices = $state([]);
  let isLoading = $state(true);
  let searchQuery = $state('');
  let filterStatus = $state('all');
  let drawerOpen = $state(false);
  let selectedInvoice = $state(null);
  let showPaymentModal = $state(false);

  // New payment form
  let payAmount = $state(0);
  let payMethod = $state('transfer');
  let payDate = $state(new Date().toISOString().slice(0, 10));
  let payRef = $state('');
  let payNote = $state('');

  // ── Create-invoice modal ───────────────────────────────
  let showCreate = $state(false);
  let creating = $state(false);
  let allJamaah = $state([]);
  let allPackages = $state([]);
  let cJamaahId = $state('');
  let cJamaahSearch = $state('');
  let cPackageId = $state('');
  let cRoomType = $state('quad');
  let cPrice = $state(0);
  let cScheme = $state('dp_lunas');
  let cDiscount = $state(0);
  let cDueDate = $state('');
  let cNotes = $state('');

  const PAYMENT_SCHEMES = [
    { id: 'dp_lunas', label: 'DP lalu Lunas' },
    { id: 'cicilan',  label: 'Cicilan' },
    { id: 'full',     label: 'Bayar Penuh' },
  ];
  const ROOM_TYPES = ['quad', 'triple', 'double', 'single'];

  let selectedPkg = $derived(allPackages.find(p => String(p.id) === String(cPackageId)) || null);
  let pkgTiers = $derived(selectedPkg?.pricing_tiers || []);
  let roomOptions = $derived(pkgTiers.length ? pkgTiers.map(t => t.room_type) : ROOM_TYPES);
  let jamaahFiltered = $derived(
    !cJamaahSearch ? allJamaah : allJamaah.filter(j =>
      (j.nama || j.nama_paspor || '').toLowerCase().includes(cJamaahSearch.toLowerCase()))
  );

  // Keep room type valid for the chosen package, and auto-fill price from its tier.
  $effect(() => {
    if (pkgTiers.length && !pkgTiers.some(t => t.room_type === cRoomType)) {
      cRoomType = pkgTiers[0].room_type;
    }
  });
  $effect(() => {
    const tier = pkgTiers.find(t => t.room_type === cRoomType);
    if (tier) cPrice = tier.price;
  });

  function openCreate() {
    cJamaahId = ''; cJamaahSearch = ''; cPackageId = '';
    cRoomType = 'quad'; cPrice = 0; cScheme = 'dp_lunas';
    cDiscount = 0; cDueDate = ''; cNotes = '';
    showCreate = true;
    // Lazy-load the picker lists only when the create drawer is opened, instead
    // of pulling them on every page mount.
    loadPickerLists();
  }

  async function loadPickerLists() {
    if (allJamaah.length && allPackages.length) return; // already loaded this session
    const [jamaahData, packageData] = await Promise.all([
      ApiService.listJamaah({ pageSize: 1000 }).catch(() => ({ jamaah: [] })),
      ApiService.listPackages({ pageSize: 1000 }).catch(() => ({ packages: [] })),
    ]);
    allJamaah = jamaahData.jamaah || jamaahData.data || jamaahData || [];
    allPackages = packageData.packages || packageData.data || packageData || [];
  }

  async function submitCreate() {
    if (!cJamaahId) { showToast('Pilih jamaah terlebih dahulu', 'warning'); return; }
    if (!cPackageId) { showToast('Pilih paket terlebih dahulu', 'warning'); return; }
    if (!cPrice || cPrice <= 0) { showToast('Harga paket belum terisi', 'warning'); return; }
    creating = true;
    try {
      // Ensure the jamaah is registered to the package, then invoice that registration.
      let reg = await ApiService.getRegistration(cJamaahId, cPackageId);
      if (!reg) {
        reg = await ApiService.registerToPackage(cJamaahId, {
          package_id: cPackageId,
          room_type: cRoomType,
          price_snapshot: cPrice,
          discount_amount: cDiscount || 0,
        });
      }
      await invoiceApi.createInvoice({
        jamaah_id: cJamaahId,
        package_id: cPackageId,
        registration_id: reg.id,
        room_type: cRoomType,
        price_snapshot: cPrice,
        discount_amount: cDiscount || 0,
        payment_scheme: cScheme,
        due_date: cDueDate || undefined,
        notes: cNotes || undefined,
      });
      showToast('Invoice berhasil dibuat', 'success');
      showCreate = false;
      await loadInvoices();
    } catch (e) {
      showToast(mapError(e.message), 'error');
    } finally {
      creating = false;
    }
  }

  // FilterTabs values use the real invoice status keys (plus "all").
  const INVOICE_STATUSES = [
    { value: 'all',          label: 'Semua' },
    { value: 'belum bayar',  label: 'Belum Bayar' },
    { value: 'sebagian',     label: 'Sebagian' },
    { value: 'lunas',        label: 'Lunas' },
  ];

  const PAY_METHODS = [
    { id: 'transfer', label: 'Transfer Bank' },
    { id: 'qris',     label: 'QRIS' },
    { id: 'tunai',    label: 'Tunai' },
    { id: 'cek',      label: 'Cek/Giro' },
  ];

  let filtered = $derived(
    invoices.filter(inv => {
      const matchStatus = filterStatus === 'all' || inv.status === filterStatus;
      const q = searchQuery.toLowerCase();
      const matchSearch = !q || inv.jamaah_name.toLowerCase().includes(q) || inv.invoice_no.toLowerCase().includes(q);
      return matchStatus && matchSearch;
    })
  );

  // Pagination (client-side over the filtered list)
  const PAGE_SIZE = 25;
  let page = $state(1);
  let paged = $derived(filtered.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE));
  // Reset to first page whenever the filter or search changes.
  $effect(() => {
    filterStatus; searchQuery;
    page = 1;
  });

  // Summary stats
  let summaryStats = $derived({
    totalTagih: invoices.reduce((s, i) => s + (i.total || 0), 0),
    totalBayar: invoices.reduce((s, i) => s + (i.paid || 0), 0),
    outstanding: invoices.reduce((s, i) => s + (i.remaining || 0), 0),
    overdueCount: invoices.filter(i => i.is_overdue).length,
  });

  onMount(loadInvoices);

  async function loadInvoices() {
    isLoading = true;
    try {
      // Only fetch invoices on mount — the list already carries jamaah_name /
      // package_name from the backend, so we no longer pull up to 1000 jamaah +
      // 1000 packages here just to build fallback name maps.
      const invoiceData = await invoiceApi.listInvoices({ status: filterStatus === 'all' ? '' : filterStatus });
      const rawInvoices = invoiceData.invoices || invoiceData || [];

      invoices = rawInvoices.map(inv => ({
        ...inv,
        invoice_no: inv.invoice_number || inv.invoice_no,
        total: inv.total_amount ?? inv.total,
        paid: inv.amount_paid ?? inv.paid,
        remaining: inv.amount_remaining ?? inv.remaining,
        jamaah_name: inv.jamaah_name || 'Jamaah',
        package_name: inv.package_name || 'Paket Umroh',
      }));
    } catch (e) {
      showToast('Gagal memuat invoice: ' + e.message, 'error');
      invoices = [];
    } finally {
      isLoading = false;
    }
  }

  function openDetail(inv) {
    selectedInvoice = inv;
    drawerOpen = true;
    showPaymentModal = false;
  }

  function openPayment() {
    payAmount = selectedInvoice?.remaining || 0;
    payRef = '';
    payNote = '';
    payMethod = 'transfer';
    payDate = new Date().toISOString().slice(0, 10);
    showPaymentModal = true;
  }

  async function submitPayment() {
    if (!payAmount || payAmount <= 0) {
      showToast('Masukkan nominal pembayaran', 'warning');
      return;
    }
    if (!selectedInvoice) return;
    try {
      await invoiceApi.recordPayment(selectedInvoice.id, {
        amount: payAmount,
        payment_method: payMethod,
        paid_at: payDate,
        reference_number: payRef,
        notes: payNote,
      });
      showToast('Pembayaran berhasil dicatat', 'success');
      showPaymentModal = false;
      drawerOpen = false;
      payAmount = 0;
      payRef = '';
      payNote = '';
      await loadInvoices();
    } catch (e) {
      showToast(e.message || 'Gagal mencatat pembayaran', 'error');
    }
  }

  async function downloadInvoicePDF(inv) {
    if (!inv) return;
    try {
      showToast('Menyiapkan kwitansi PDF...');
      const url = `/api/invoices/${inv.id}/pdf`;
      const res = await fetch(url, { headers: authHeaders() });
      if (!res.ok) {
        const err = await res.text();
        showToast(err || 'Gagal mengunduh PDF', 'error');
        return;
      }
      const blob = await res.blob();
      const blobUrl = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = blobUrl;
      a.download = `invoice_${inv.invoice_no || inv.invoice_number || inv.id}.pdf`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(blobUrl);
      showToast('Kwitansi PDF berhasil diunduh');
    } catch (e) {
      showToast('Gagal mengunduh: ' + e.message, 'error');
    }
  }

  function pct(inv) {
    if (!inv || !inv.total) return 0;
    return Math.round((inv.paid / inv.total) * 100);
  }
</script>

<div class="suluk-page">
  <PageHeader
    kicker="Penjualan"
    title="Invoice & Pembayaran"
    subtitle="Buat tagihan, catat pembayaran, dan pantau tunggakan jamaah."
  >
    {#snippet actions()}
      <Button variant="ghost" icon={Download} onclick={() => selectedInvoice ? downloadInvoicePDF(selectedInvoice) : showToast('Pilih invoice untuk diunduh', 'warning')}>Ekspor</Button>
      <Button variant="primary" icon={Plus} onclick={openCreate}>Buat Invoice</Button>
    {/snippet}
  </PageHeader>

  <!-- Summary cards -->
  <div class="suluk-stat-grid">
    <StatCard icon={Receipt} label="Total Tagihan" value={formatIDR(summaryStats.totalTagih)} accent="var(--c-primary)" />
    <StatCard icon={CheckCircle} label="Sudah Diterima" value={formatIDR(summaryStats.totalBayar)} accent="var(--c-success)" />
    <StatCard icon={Clock} label="Outstanding" value={formatIDR(summaryStats.outstanding)} accent="var(--c-warning)" />
    <StatCard icon={AlertCircle} label="Jatuh Tempo" value={`${summaryStats.overdueCount}`} accent="var(--c-danger)" sub="perlu ditindaklanjuti" />
  </div>

  <!-- Table card -->
  <Card pad={false}>
    <div class="suluk-toolbar">
      <FilterTabs tabs={INVOICE_STATUSES} value={filterStatus} onChange={(v) => (filterStatus = v)} />
      <div class="suluk-search">
        <Search size={17} class="suluk-search-icon" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Cari invoice…"
        />
      </div>
    </div>

    {#if isLoading}
      <div class="suluk-skeletons">
        {#each [1, 2, 3, 4, 5] as _}
          <div class="suluk-skeleton-row"></div>
        {/each}
      </div>
    {:else if filtered.length === 0}
      <EmptyState
        icon={Receipt}
        title={searchQuery || filterStatus !== 'all' ? 'Tidak ada invoice yang cocok' : 'Belum ada invoice'}
        text={searchQuery || filterStatus !== 'all' ? 'Coba ubah kata kunci atau filter status.' : 'Invoice akan muncul di sini setelah dibuat dari paket jamaah.'}
      />
    {:else}
      <div class="suluk-table-wrap">
        <table class="suluk-table">
          <thead>
            <tr>
              <th>Invoice#</th>
              <th>Jamaah</th>
              <th class="num">Jumlah</th>
              <th class="num">Dibayar</th>
              <th>Jatuh Tempo</th>
              <th class="center">Status</th>
            </tr>
          </thead>
          <tbody>
            {#each paged as inv}
              <tr class="suluk-row" onclick={() => openDetail(inv)}>
                <td>
                  <div class="cell-stack">
                    <span class="inv-id">{inv.invoice_no}</span>
                    <span class="inv-sub">{formatDate(inv.created_at || inv.issue_date)}</span>
                  </div>
                </td>
                <td>
                  <div class="cell-jamaah">
                    <Avatar name={inv.jamaah_name} size={34} />
                    <div class="cell-stack min-w-0">
                      <span class="jamaah-name">{inv.jamaah_name}</span>
                      <span class="inv-sub">{inv.package_name} · {inv.room_type}</span>
                    </div>
                  </div>
                </td>
                <td class="num">
                  <span class="amount">{formatIDR(inv.total)}</span>
                </td>
                <td class="num">
                  <span class="amount" style={inv.paid > 0 ? 'color:var(--c-success)' : 'color:var(--c-faint)'}>{formatIDR(inv.paid)}</span>
                </td>
                <td>
                  <span class="due {inv.is_overdue ? 'overdue' : ''}">
                    {#if inv.is_overdue}<AlertCircle size={13} />{/if}
                    {formatDate(inv.due_date)}
                  </span>
                </td>
                <td class="center">
                  <StatusBadge status={inv.status} />
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="px-4">
        <Pager {page} pageSize={PAGE_SIZE} total={filtered.length} onchange={(p) => (page = p)} />
      </div>
    {/if}
  </Card>
</div>

<!-- Invoice Detail Drawer -->
<SlideDrawer
  open={drawerOpen}
  title="Invoice {selectedInvoice?.invoice_no || ''}"
  width="480px"
  onClose={() => (drawerOpen = false)}
>
  {#if selectedInvoice}
    <div class="drawer-body">
      <!-- Status + issue date -->
      <div class="drawer-status-row">
        <StatusBadge status={selectedInvoice.status} />
        <span class="issue-date">Terbit {formatDate(selectedInvoice.created_at || selectedInvoice.issue_date)}</span>
      </div>

      <!-- Jamaah -->
      <div class="drawer-jamaah">
        <Avatar name={selectedInvoice.jamaah_name} size={42} />
        <div class="min-w-0">
          <p class="dj-name">{selectedInvoice.jamaah_name}</p>
          <p class="dj-sub">{selectedInvoice.package_name} · {selectedInvoice.room_type}</p>
        </div>
      </div>

      <!-- Billing summary -->
      <div class="summary-box">
        <div class="summary-line">
          <span>Total tagihan</span>
          <span class="amount">{formatIDR(selectedInvoice.total)}</span>
        </div>
        <div class="summary-line">
          <span>Sudah dibayar</span>
          <span class="amount" style="color:var(--c-success)">{formatIDR(selectedInvoice.paid)}</span>
        </div>
        <div class="summary-line total">
          <span>Sisa</span>
          <span class="amount" style={selectedInvoice.remaining > 0 ? 'color:var(--c-danger)' : 'color:var(--c-success)'}>
            {selectedInvoice.remaining > 0 ? formatIDR(selectedInvoice.remaining) : 'LUNAS'}
          </span>
        </div>
        <div class="progress-track">
          <div class="progress-fill" style="width:{pct(selectedInvoice)}%"></div>
        </div>
        <p class="progress-label">{pct(selectedInvoice)}% lunas</p>
      </div>

      <!-- Payment history -->
      <div>
        <h4 class="section-label">Riwayat Pembayaran</h4>
        {#if (selectedInvoice.payments?.length ?? 0) === 0}
          <p class="empty-line">Belum ada pembayaran tercatat.</p>
        {:else}
          <div class="pay-list">
            {#each selectedInvoice.payments as pay}
              <div class="pay-item">
                <div class="pay-icon"><Check size={16} /></div>
                <div class="min-w-0 flex-1">
                  <p class="pay-amount">{formatIDR(pay.amount)}</p>
                  <p class="pay-meta">{pay.method}{pay.ref ? ' · ' + pay.ref : ''} · {formatDate(pay.date)}</p>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer actions -->
      <div class="drawer-footer">
        <Button variant="ghost" icon={Printer} full onclick={() => downloadInvoicePDF(selectedInvoice)}>Cetak</Button>
        {#if selectedInvoice.remaining > 0}
          <Button variant="primary" icon={Wallet} full onclick={openPayment}>Catat Pembayaran</Button>
        {/if}
      </div>
    </div>
  {/if}
</SlideDrawer>

<!-- Record Payment Modal -->
{#if showPaymentModal && selectedInvoice}
  <button type="button" class="modal-backdrop" aria-label="Tutup" onclick={() => (showPaymentModal = false)}></button>
  <div class="modal-panel" role="dialog" aria-modal="true" aria-label="Catat Pembayaran">
    <div class="modal-head">
      <h3 class="modal-title">Catat Pembayaran</h3>
      <button type="button" class="modal-close" onclick={() => (showPaymentModal = false)} aria-label="Tutup">
        <X size={18} />
      </button>
    </div>

    <div class="modal-body">
      <p class="modal-hint">
        Untuk <strong>{selectedInvoice.jamaah_name}</strong> · sisa
        <strong>{formatIDR(selectedInvoice.remaining)}</strong>
      </p>

      <IDRInput label="Jumlah Pembayaran" bind:value={payAmount} required />

      <div class="field">
        <label for="pay-date">Tanggal</label>
        <input id="pay-date" type="date" bind:value={payDate} />
      </div>

      <div class="field">
        <label for="pay-method">Metode</label>
        <select id="pay-method" bind:value={payMethod}>
          {#each PAY_METHODS as m}
            <option value={m.id}>{m.label}</option>
          {/each}
        </select>
      </div>

      <div class="field">
        <label for="pay-ref">No. Referensi (opsional)</label>
        <input id="pay-ref" type="text" bind:value={payRef} placeholder="Nomor bukti transfer…" />
      </div>

      <div class="field">
        <label for="pay-note">Catatan (opsional)</label>
        <input id="pay-note" type="text" bind:value={payNote} placeholder="Catatan pembayaran…" />
      </div>
    </div>

    <div class="modal-foot">
      <Button variant="ghost" onclick={() => (showPaymentModal = false)}>Batal</Button>
      <Button variant="primary" icon={Check} onclick={submitPayment}>Simpan</Button>
    </div>
  </div>
{/if}

<!-- Buat Invoice Drawer -->
<SlideDrawer open={showCreate} title="Buat Invoice" width="520px" onClose={() => (showCreate = false)}>
  <div class="space-y-4 p-6">
    <!-- Jamaah -->
    <div class="flex flex-col gap-1">
      <label for="ci-jamaah" class="text-sm font-medium text-slate-700">Jamaah <span class="text-red-500">*</span></label>
      <input
        type="text"
        bind:value={cJamaahSearch}
        placeholder="Cari nama jamaah..."
        class="mb-1 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-primary-400"
      />
      <select id="ci-jamaah" bind:value={cJamaahId} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400">
        <option value="">— Pilih jamaah —</option>
        {#each jamaahFiltered as j}
          <option value={j.id}>{j.nama || j.nama_paspor || 'Tanpa nama'}</option>
        {/each}
      </select>
      {#if allJamaah.length === 0}
        <p class="text-xs text-amber-600">Belum ada jamaah. Tambahkan jamaah di menu CRM dulu.</p>
      {/if}
    </div>

    <!-- Package -->
    <div class="flex flex-col gap-1">
      <label for="ci-pkg" class="text-sm font-medium text-slate-700">Paket <span class="text-red-500">*</span></label>
      <select id="ci-pkg" bind:value={cPackageId} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400">
        <option value="">— Pilih paket —</option>
        {#each allPackages as p}
          <option value={p.id}>{p.name}</option>
        {/each}
      </select>
    </div>

    <!-- Room type + price -->
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="ci-room" class="text-sm font-medium text-slate-700">Tipe Kamar</label>
        <select id="ci-room" bind:value={cRoomType} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400">
          {#each roomOptions as rt}
            <option value={rt}>{rt}</option>
          {/each}
        </select>
      </div>
      <IDRInput label="Harga" bind:value={cPrice} />
    </div>

    <!-- Payment scheme -->
    <div class="flex flex-col gap-1">
      <label for="ci-scheme" class="text-sm font-medium text-slate-700">Skema Pembayaran</label>
      <select id="ci-scheme" bind:value={cScheme} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400">
        {#each PAYMENT_SCHEMES as s}
          <option value={s.id}>{s.label}</option>
        {/each}
      </select>
    </div>

    <IDRInput label="Diskon (opsional)" bind:value={cDiscount} />

    <div class="flex flex-col gap-1">
      <label for="ci-due" class="text-sm font-medium text-slate-700">Jatuh Tempo (opsional)</label>
      <input id="ci-due" type="date" bind:value={cDueDate} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
    </div>

    <div class="flex flex-col gap-1">
      <label for="ci-notes" class="text-sm font-medium text-slate-700">Catatan (opsional)</label>
      <input id="ci-notes" type="text" bind:value={cNotes} class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
    </div>

    <div class="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3 text-sm">
      <span class="text-slate-500">Total Tagihan</span>
      <span class="font-bold text-slate-800">{formatIDR(Math.max(0, (cPrice || 0) - (cDiscount || 0)))}</span>
    </div>

    <div class="flex gap-2 pt-1">
      <button type="button" onclick={() => (showCreate = false)} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600 hover:bg-slate-50">Batal</button>
      <button type="button" onclick={submitCreate} disabled={creating} class="flex flex-1 items-center justify-center gap-2 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">
        {#if creating}<Loader2 class="h-4 w-4 animate-spin" />{/if}
        Buat Invoice
      </button>
    </div>
  </div>
</SlideDrawer>

<style>
  .suluk-page {
    padding: var(--space-6);
    max-width: 1400px;
    margin: 0 auto;
  }
  .suluk-stat-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--space-4);
    margin-bottom: var(--space-5);
  }

  /* Toolbar (FilterTabs + SearchBar) */
  .suluk-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
    padding: 16px 20px;
    flex-wrap: wrap;
  }
  .suluk-search {
    position: relative;
    width: 280px;
    max-width: 100%;
  }
  :global(.suluk-search .suluk-search-icon) {
    position: absolute;
    left: 13px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--c-faint);
    pointer-events: none;
  }
  .suluk-search input {
    width: 100%;
    padding: 10px 14px 10px 38px;
    font-size: 13.5px;
    color: var(--c-ink);
    background: var(--c-surface);
    border: 1px solid var(--c-line);
    border-radius: var(--radius);
    outline: none;
    transition: border-color .15s, box-shadow .15s;
  }
  .suluk-search input:focus {
    border-color: var(--c-primary);
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }

  /* Table */
  .suluk-table-wrap {
    overflow-x: auto;
    padding: 0 4px 4px;
  }
  .suluk-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13.5px;
  }
  .suluk-table th {
    text-align: left;
    padding: 0 16px 12px;
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: .05em;
    text-transform: uppercase;
    color: var(--c-faint);
    white-space: nowrap;
    border-bottom: 1px solid var(--c-line);
  }
  .suluk-table th.num { text-align: right; }
  .suluk-table th.center { text-align: center; }
  .suluk-table td {
    padding: 13px 16px;
    border-bottom: 1px solid var(--c-line-soft);
    color: var(--c-ink-soft);
    white-space: nowrap;
    vertical-align: middle;
  }
  .suluk-table td.num { text-align: right; }
  .suluk-table td.center { text-align: center; }
  .suluk-row { cursor: pointer; transition: background .12s; }
  .suluk-row:hover { background: var(--c-bg); }
  .suluk-table tbody tr:last-child td { border-bottom: none; }

  .cell-stack { display: flex; flex-direction: column; gap: 2px; }
  .cell-jamaah { display: flex; align-items: center; gap: 11px; }
  .inv-id { font-weight: 800; color: var(--c-ink); }
  .inv-sub { font-size: 12px; color: var(--c-faint); overflow: hidden; text-overflow: ellipsis; }
  .jamaah-name { font-weight: 600; color: var(--c-ink); overflow: hidden; text-overflow: ellipsis; }
  .amount { font-weight: 700; font-variant-numeric: tabular-nums; color: var(--c-ink); }
  .due {
    display: inline-flex; align-items: center; gap: 5px;
    font-weight: 600; color: var(--c-ink-soft);
  }
  .due.overdue { color: var(--c-danger); }
  .min-w-0 { min-width: 0; }

  /* Skeletons */
  .suluk-skeletons { display: flex; flex-direction: column; gap: 10px; padding: 16px 20px; }
  .suluk-skeleton-row {
    height: 52px;
    border-radius: var(--radius);
    background: var(--c-bg-2);
    animation: suluk-pulse 1.4s ease-in-out infinite;
  }
  @keyframes suluk-pulse { 0%,100% { opacity: 1; } 50% { opacity: .5; } }

  /* Drawer body */
  .drawer-body { display: flex; flex-direction: column; gap: 20px; padding: 24px; }
  .drawer-status-row { display: flex; justify-content: space-between; align-items: center; }
  .issue-date { font-size: 13px; color: var(--c-muted); }
  .drawer-jamaah { display: flex; align-items: center; gap: 12px; }
  .dj-name { font-size: 16px; font-weight: 800; color: var(--c-ink); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .dj-sub { font-size: 13px; color: var(--c-muted); margin-top: 2px; }

  .summary-box {
    background: var(--c-bg);
    border-radius: var(--radius);
    padding: 18px;
  }
  .summary-line {
    display: flex; justify-content: space-between; align-items: center;
    padding: 6px 0; font-size: 13.5px; color: var(--c-muted);
  }
  .summary-line.total {
    margin-top: 6px; padding-top: 12px;
    border-top: 1px solid var(--c-line);
    font-size: 15px; color: var(--c-ink);
  }
  .summary-line.total span:first-child { font-weight: 700; }
  .summary-line.total .amount { font-weight: 800; }
  .progress-track {
    height: 7px; background: var(--c-bg-2); border-radius: 999px;
    overflow: hidden; margin-top: 14px;
  }
  .progress-fill {
    height: 100%; background: var(--c-success); border-radius: 999px;
    transition: width .6s cubic-bezier(.2,.7,.3,1);
  }
  .progress-label { text-align: center; font-size: 11.5px; color: var(--c-faint); margin-top: 8px; }

  .section-label {
    font-size: 12.5px; font-weight: 700; color: var(--c-faint);
    text-transform: uppercase; letter-spacing: .04em; margin-bottom: 10px;
  }
  .empty-line { font-size: 13px; color: var(--c-faint); padding: 8px 0; }
  .pay-list { display: flex; flex-direction: column; gap: 8px; }
  .pay-item {
    display: flex; align-items: center; gap: 12px;
    padding: 12px 14px; background: var(--c-success-soft);
    border-radius: var(--radius-sm);
  }
  .pay-icon {
    width: 32px; height: 32px; border-radius: 8px;
    background: var(--c-success); color: #fff;
    display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  }
  .pay-amount { font-size: 13.5px; font-weight: 700; color: var(--c-ink); }
  .pay-meta { font-size: 12px; color: var(--c-muted); margin-top: 2px; }

  .drawer-footer { display: flex; gap: 10px; padding-top: 4px; }

  /* Modal */
  .modal-backdrop {
    position: fixed; inset: 0; z-index: 95;
    background: rgba(16,33,28,.4);
    backdrop-filter: blur(2px);
    border: none; cursor: pointer;
    animation: suluk-fade .2s ease;
  }
  .modal-panel {
    position: fixed; z-index: 96;
    top: 50%; left: 50%; transform: translate(-50%, -50%);
    width: 460px; max-width: 94vw;
    background: var(--c-surface);
    border-radius: var(--radius-xl);
    box-shadow: var(--shadow-lg);
    overflow: hidden;
    animation: suluk-scale .25s cubic-bezier(.2,.7,.3,1) both;
  }
  @keyframes suluk-fade { from { opacity: 0; } to { opacity: 1; } }
  @keyframes suluk-scale {
    from { opacity: 0; transform: translate(-50%, -50%) scale(.96); }
    to { opacity: 1; transform: translate(-50%, -50%) scale(1); }
  }
  .modal-head {
    padding: 22px 26px 0;
    display: flex; justify-content: space-between; align-items: flex-start;
  }
  .modal-title { font-size: 18px; font-weight: 800; color: var(--c-ink); }
  .modal-close {
    display: flex; align-items: center; justify-content: center;
    width: 32px; height: 32px; border-radius: 8px;
    color: var(--c-faint); transition: background .15s, color .15s;
  }
  .modal-close:hover { background: var(--c-bg-2); color: var(--c-ink); }
  .modal-body {
    padding: 14px 26px 24px;
    display: flex; flex-direction: column; gap: 16px;
  }
  .modal-hint { font-size: 13.5px; color: var(--c-muted); }
  .modal-hint strong { color: var(--c-ink); font-variant-numeric: tabular-nums; }
  .modal-foot {
    padding: 16px 26px;
    border-top: 1px solid var(--c-line);
    display: flex; gap: 10px; justify-content: flex-end;
    background: var(--c-bg);
  }
  .field { display: flex; flex-direction: column; gap: 6px; }
  .field label { font-size: 12.5px; font-weight: 700; color: var(--c-ink-soft); }
  .field input,
  .field select {
    width: 100%; padding: 11px 13px; font-size: 13.5px;
    color: var(--c-ink); background: var(--c-surface);
    border: 1px solid var(--c-line); border-radius: var(--radius);
    outline: none; transition: border-color .15s, box-shadow .15s;
  }
  .field input:focus,
  .field select:focus {
    border-color: var(--c-primary);
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }

  @media (max-width: 640px) {
    .suluk-page { padding: var(--space-4); }
    .suluk-toolbar { gap: 12px; }
    .suluk-search { width: 100%; }
  }
</style>
