<script>
  import { onMount } from 'svelte';
  import {
    Plus, Search, Receipt, AlertCircle, CheckCircle,
    Clock, ChevronRight, Upload, Download, Loader2,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import IDRInput from '../components/IDRInput.svelte';
  import EmptyState from '../components/EmptyState.svelte';
  import Pager from '../components/Pager.svelte';
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

  const INVOICE_STATUSES = [
    { id: 'all',          label: 'Semua' },
    { id: 'belum bayar',  label: 'Belum Bayar' },
    { id: 'sebagian',     label: 'Sebagian' },
    { id: 'lunas',        label: 'Lunas' },
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
    totalActive: invoices.filter(i => i.status !== 'lunas').reduce((s, i) => s + i.remaining, 0),
    overdueCount: invoices.filter(i => i.is_overdue).length,
    cashToday: invoices.reduce((s, i) => s + i.payments.filter(p => p.date === new Date().toISOString().slice(0,10)).reduce((a, p) => a + p.amount, 0), 0),
  });

  onMount(loadInvoices);

  async function loadInvoices() {
    isLoading = true;
    try {
      const [invoiceData, jamaahData, packageData] = await Promise.all([
        invoiceApi.listInvoices({ status: filterStatus === 'all' ? '' : filterStatus }),
        ApiService.listJamaah({ pageSize: 1000 }).catch(() => ({ jamaah: [] })),
        ApiService.listPackages({ pageSize: 1000 }).catch(() => ({ packages: [] })),
      ]);

      const rawInvoices = invoiceData.invoices || invoiceData || [];
      const jamaahList = jamaahData.jamaah || jamaahData.data || jamaahData || [];
      const packageList = packageData.packages || packageData.data || packageData || [];

      allJamaah = jamaahList;
      allPackages = packageList;
      const jamaahMap = new Map(jamaahList.map(j => [j.id, j.nama || j.nama_paspor || j.name || 'Tanpa Nama']));
      const packageMap = new Map(packageList.map(p => [p.id, p.name || 'Tanpa Nama']));

      invoices = rawInvoices.map(inv => ({
        ...inv,
        invoice_no: inv.invoice_number || inv.invoice_no,
        total: inv.total_amount ?? inv.total,
        paid: inv.amount_paid ?? inv.paid,
        remaining: inv.amount_remaining ?? inv.remaining,
        jamaah_name: inv.jamaah_name || jamaahMap.get(inv.jamaah_id) || 'Jamaah',
        package_name: inv.package_name || packageMap.get(inv.package_id) || 'Paket Umroh',
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
</script>

<div class="flex h-screen flex-col">
  <!-- Header -->
  <div class="flex-shrink-0 border-b border-slate-100 bg-white px-6 py-5">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold text-slate-800">Invoice & Pembayaran</h1>
        <p class="mt-0.5 text-sm text-slate-500">{filtered.length} invoice</p>
      </div>
      <button
        type="button"
        onclick={openCreate}
        class="flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm shadow-primary-600/30 transition-all hover:bg-primary-700"
      >
        <Plus class="h-4 w-4" />
        Buat Invoice
      </button>
    </div>

    <!-- Summary cards -->
    <div class="mt-4 grid grid-cols-3 gap-3">
      <div class="rounded-xl bg-red-50 p-3">
        <p class="text-[11px] font-semibold text-red-400">Total Piutang Aktif</p>
        <p class="mt-0.5 text-base font-bold text-red-700">{formatIDR(summaryStats.totalActive)}</p>
      </div>
      <div class="rounded-xl bg-amber-50 p-3">
        <p class="text-[11px] font-semibold text-amber-400">Overdue</p>
        <p class="mt-0.5 text-base font-bold text-amber-700">{summaryStats.overdueCount} jamaah</p>
      </div>
      <div class="rounded-xl bg-emerald-50 p-3">
        <p class="text-[11px] font-semibold text-emerald-400">Kas Hari Ini</p>
        <p class="mt-0.5 text-base font-bold text-emerald-700">{formatIDR(summaryStats.cashToday)}</p>
      </div>
    </div>

    <!-- Search + filter -->
    <div class="mt-4 flex gap-3">
      <div class="relative flex-1 min-w-0">
        <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Cari nama atau nomor invoice..."
          class="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-9 pr-3 text-sm outline-none focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>
      <div class="flex gap-1">
        {#each INVOICE_STATUSES as s}
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

  <!-- Invoice table -->
  <div class="flex-1 overflow-auto">
    {#if isLoading}
      <div class="space-y-3 p-6">
        {#each [1,2,3,4] as _}
          <div class="h-16 animate-pulse rounded-xl bg-slate-100"></div>
        {/each}
      </div>
    {:else if filtered.length === 0}
      <EmptyState
        icon={Receipt}
        title={searchQuery || filterStatus !== 'all' ? 'Tidak ada invoice yang cocok' : 'Belum ada invoice'}
        text={searchQuery || filterStatus !== 'all' ? 'Coba ubah kata kunci atau filter status.' : 'Invoice akan muncul di sini setelah dibuat dari paket jamaah.'}
      />
    {:else}
      <table class="w-full">
        <thead class="sticky top-0 bg-slate-50">
          <tr class="text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
            <th class="hidden px-6 py-3 md:table-cell">No. Invoice</th>
            <th class="px-4 py-3 sm:px-6">Jamaah</th>
            <th class="hidden px-4 py-3 text-right lg:table-cell">Total</th>
            <th class="hidden px-4 py-3 text-right lg:table-cell">Terbayar</th>
            <th class="px-4 py-3 text-right">Sisa</th>
            <th class="px-4 py-3">Status</th>
            <th class="hidden px-4 py-3 md:table-cell">Jatuh Tempo</th>
            <th class="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          {#each paged as inv}
            <tr class="group bg-white transition-colors hover:bg-primary-50/30 {inv.is_overdue ? 'bg-red-50/30' : ''}">
              <td class="hidden px-6 py-3.5 md:table-cell">
                <div class="flex items-center gap-2">
                  {#if inv.is_overdue}
                    <AlertCircle class="h-4 w-4 flex-shrink-0 text-red-500" />
                  {/if}
                  <span class="text-sm font-mono font-semibold text-slate-700">{inv.invoice_no}</span>
                </div>
              </td>
              <td class="px-4 py-3.5 sm:px-6">
                <div class="flex items-center gap-2">
                  {#if inv.is_overdue}
                    <AlertCircle class="h-4 w-4 flex-shrink-0 text-red-500 md:hidden" />
                  {/if}
                  <div class="min-w-0">
                    <p class="truncate text-sm font-semibold text-slate-800">{inv.jamaah_name}</p>
                    <p class="truncate text-xs text-slate-400">{inv.package_name} · {inv.room_type}</p>
                  </div>
                </div>
              </td>
              <td class="hidden px-4 py-3.5 text-right text-sm font-medium text-slate-600 lg:table-cell">{formatIDR(inv.total)}</td>
              <td class="hidden px-4 py-3.5 text-right text-sm font-medium text-emerald-600 lg:table-cell">{formatIDR(inv.paid)}</td>
              <td class="px-4 py-3.5 text-right text-sm font-bold {inv.remaining > 0 ? 'text-red-600' : 'text-emerald-600'}">
                {inv.remaining > 0 ? formatIDR(inv.remaining) : 'Lunas'}
              </td>
              <td class="px-4 py-3.5">
                <StatusBadge status={inv.status} size="xs" />
              </td>
              <td class="hidden px-4 py-3.5 md:table-cell">
                <span class="text-xs {inv.is_overdue ? 'font-semibold text-red-600' : 'text-slate-500'}">
                  {formatDate(inv.due_date)}
                </span>
              </td>
              <td class="px-4 py-3.5 text-right">
                <button
                  type="button"
                  onclick={() => openDetail(inv)}
                  class="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-semibold text-primary-600 transition-colors hover:bg-primary-50"
                >
                  Detail <ChevronRight class="h-3 w-3" />
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
      <div class="px-6">
        <Pager {page} pageSize={PAGE_SIZE} total={filtered.length} onchange={(p) => (page = p)} />
      </div>
    {/if}
  </div>
</div>

<!-- Invoice Detail Drawer -->
<SlideDrawer
  open={drawerOpen}
  title="Invoice {selectedInvoice?.invoice_no || ''}"
  width="580px"
  onClose={() => (drawerOpen = false)}
>
  {#if selectedInvoice}
    <div class="p-6 space-y-5">
      <!-- Header info -->
      <div class="rounded-xl bg-slate-50 p-4 space-y-2">
        <div class="flex items-center justify-between">
          <StatusBadge status={selectedInvoice.status} />
          {#if selectedInvoice.is_overdue}
            <span class="flex items-center gap-1 text-xs font-semibold text-red-600">
              <AlertCircle class="h-3.5 w-3.5" /> Overdue
            </span>
          {/if}
        </div>
        <p class="text-base font-bold text-slate-800">{selectedInvoice.jamaah_name}</p>
        <p class="text-sm text-slate-500">{selectedInvoice.package_name} · {selectedInvoice.room_type}</p>
      </div>

      <!-- Billing summary -->
      <div class="rounded-xl border border-slate-100 p-4 space-y-3">
        <h4 class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Ringkasan Tagihan</h4>
        <div class="space-y-2 text-sm">
          <div class="flex justify-between">
            <span class="text-slate-500">Total Tagihan</span>
            <span class="font-semibold text-slate-800">{formatIDR(selectedInvoice.total)}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-500">Sudah Dibayar</span>
            <span class="font-semibold text-emerald-600">{formatIDR(selectedInvoice.paid)}</span>
          </div>
          <div class="border-t border-slate-100 pt-2 flex justify-between">
            <span class="font-semibold text-slate-700">Sisa Tagihan</span>
            <span class="font-bold {selectedInvoice.remaining > 0 ? 'text-red-600' : 'text-emerald-600'}">
              {selectedInvoice.remaining > 0 ? formatIDR(selectedInvoice.remaining) : 'LUNAS'}
            </span>
          </div>
        </div>
        <!-- Progress bar -->
        <div class="h-2 overflow-hidden rounded-full bg-slate-100">
          <div
            class="h-full rounded-full bg-emerald-400"
            style="width: {Math.round((selectedInvoice.paid / selectedInvoice.total) * 100)}%"
          ></div>
        </div>
        <p class="text-center text-[11px] text-slate-400">
          {Math.round((selectedInvoice.paid / selectedInvoice.total) * 100)}% lunas
        </p>
      </div>

      <!-- Payment history -->
      <div class="rounded-xl border border-slate-100 p-4">
        <h4 class="mb-3 text-[10px] font-bold uppercase tracking-wider text-slate-400">Riwayat Pembayaran</h4>
        {#if selectedInvoice.payments.length === 0}
          <p class="text-sm text-slate-400">Belum ada pembayaran.</p>
        {:else}
          <div class="space-y-2">
            {#each selectedInvoice.payments as pay}
              <div class="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2.5">
                <div>
                  <p class="text-xs font-semibold text-slate-700">{formatDate(pay.date)}</p>
                  <p class="text-[11px] text-slate-400">{pay.method} {pay.ref ? '· ' + pay.ref : ''}</p>
                </div>
                <span class="text-sm font-bold text-emerald-600">{formatIDR(pay.amount)}</span>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Actions -->
      {#if !showPaymentModal}
        <div class="flex gap-3">
          <button
            type="button"
            onclick={() => downloadInvoicePDF(selectedInvoice)}
            class="flex flex-1 items-center justify-center gap-2 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600 hover:bg-slate-50"
          >
            <Download class="h-4 w-4" />
            PDF Invoice
          </button>
          {#if selectedInvoice.remaining > 0}
            <button
              type="button"
              onclick={() => (showPaymentModal = true)}
              class="flex flex-1 items-center justify-center gap-2 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700"
            >
              <Plus class="h-4 w-4" />
              Rekam Pembayaran
            </button>
          {/if}
        </div>

      <!-- Record payment form -->
      {:else}
        <div class="rounded-xl border border-primary-200 bg-primary-50 p-4 space-y-3">
          <h4 class="text-sm font-bold text-primary-800">Rekam Pembayaran Baru</h4>

          <IDRInput label="Nominal" bind:value={payAmount} required />

          <div class="flex flex-col gap-1">
            <label for="pay-method" class="text-sm font-medium text-slate-700">Metode</label>
            <select
              id="pay-method"
              bind:value={payMethod}
              class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
            >
              {#each PAY_METHODS as m}
                <option value={m.id}>{m.label}</option>
              {/each}
            </select>
          </div>

          <div class="flex flex-col gap-1">
            <label for="pay-date" class="text-sm font-medium text-slate-700">Tanggal</label>
            <input
              id="pay-date"
              type="date"
              bind:value={payDate}
              class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
            />
          </div>

          <div class="flex flex-col gap-1">
            <label for="pay-ref" class="text-sm font-medium text-slate-700">No. Referensi (opsional)</label>
            <input
              id="pay-ref"
              type="text"
              bind:value={payRef}
              placeholder="Nomor bukti transfer..."
              class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
            />
          </div>

          <div class="flex gap-2">
            <button
              type="button"
              onclick={() => (showPaymentModal = false)}
              class="flex-1 rounded-xl border border-slate-200 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-50"
            >
              Batal
            </button>
            <button
              type="button"
              onclick={submitPayment}
              class="flex-1 rounded-xl bg-primary-600 py-2 text-sm font-semibold text-white hover:bg-primary-700"
            >
              Simpan
            </button>
          </div>
        </div>
      {/if}
    </div>
  {/if}
</SlideDrawer>

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
