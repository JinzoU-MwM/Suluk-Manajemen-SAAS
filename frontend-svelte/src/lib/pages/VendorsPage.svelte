<script>
  import { onMount } from 'svelte';
  import {
    Building2, Plus, Search, AlertCircle, ChevronRight, X,
    Pencil, Trash2, Banknote, Calendar, Filter, CreditCard,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import IDRInput from '../components/IDRInput.svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { ApiService } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  const VENDOR_TYPES = [
    { id: 'all', label: 'Semua' },
    { id: 'maskapai', label: 'Maskapai' },
    { id: 'hotel', label: 'Hotel' },
    { id: 'transport', label: 'Transport' },
    { id: 'perlengkapan', label: 'Perlengkapan' },
    { id: 'katering', label: 'Katering' },
    { id: 'lainnya', label: 'Lainnya' },
  ];

  const BILL_STATUSES = [
    { id: 'all', label: 'Semua' },
    { id: 'belum_bayar', label: 'Belum Bayar' },
    { id: 'sebagian', label: 'Sebagian' },
    { id: 'lunas', label: 'Lunas' },
  ];

  // ── State ──────────────────────────────────────────────
  let vendors = $state([]);
  let isLoading = $state(true);
  let searchQuery = $state('');
  let filterType = $state('all');

  // Summary
  let debtSummary = $state(null);
  let overdueCount = $state(0);

  // Vendor form (create/edit)
  let showVendorForm = $state(false);
  let editingVendor = $state(null);
  let vendorForm = $state({
    name: '', type: 'lainnya', npwp: '', address: '',
    pic_name: '', pic_phone: '', pic_email: '',
    bank_name: '', bank_account_number: '', bank_account_name: '',
    notes: '',
  });
  let savingVendor = $state(false);

  // Vendor detail drawer
  let showDetail = $state(false);
  let selectedVendor = $state(null);
  let detailTab = $state('info');

  // Bills for selected vendor
  let vendorBills = $state([]);
  let billsLoading = $state(false);
  let billFilter = $state('all');

  // Bill form
  let showBillForm = $state(false);
  let billForm = $state({
    vendor_id: '', package_id: '', description: '',
    amount: 0, currency: 'IDR', exchange_rate: 1, due_date: '',
  });
  let savingBill = $state(false);

  // Expanded bill detail
  let expandedBillId = $state(null);
  let billPayments = $state([]);
  let paymentsLoading = $state(false);

  // Payment form
  let showPaymentForm = $state(false);
  let payBillId = $state('');
  let payForm = $state({ amount: 0, payment_date: '', source_account: '', notes: '' });
  let savingPayment = $state(false);

  // Packages list for bill form
  let packages = $state([]);
  let packagesLoading = $state(false);

  // ── Derived ──────────────────────────────────────────────
  let filteredVendors = $derived(
    vendors.filter(v => {
      if (filterType !== 'all' && v.type !== filterType) return false;
      if (!searchQuery) return true;
      const q = searchQuery.toLowerCase();
      return v.name.toLowerCase().includes(q)
        || (v.pic_name && v.pic_name.toLowerCase().includes(q));
    })
  );

  let filteredBills = $derived(
    vendorBills.filter(b => billFilter === 'all' || b.status === billFilter)
  );

  // ── Setup ────────────────────────────────────────────────
  onMount(loadAll);

  async function loadAll() {
    isLoading = true;
    try {
      await Promise.all([loadVendors(), loadDebtSummary(), loadOverdueCount()]);
    } catch (e) {
      showToast(e.message || 'Gagal memuat data vendor', 'error');
    } finally {
      isLoading = false;
    }
  }

  async function loadVendors() {
    const data = await ApiService.listVendors({ pageSize: 200 });
    vendors = Array.isArray(data) ? data : (data?.data ?? []);
  }

  async function loadDebtSummary() {
    try {
      debtSummary = await ApiService.getDebtSummary();
    } catch { /* non-critical */ }
  }

  async function loadOverdueCount() {
    try {
      const bills = await ApiService.getOverdueBills();
      overdueCount = Array.isArray(bills) ? bills.length : 0;
    } catch { /* non-critical */ }
  }

  // ── Vendor form ──────────────────────────────────────────
  function openCreateVendor() {
    editingVendor = null;
    vendorForm = {
      name: '', type: 'lainnya', npwp: '', address: '',
      pic_name: '', pic_phone: '', pic_email: '',
      bank_name: '', bank_account_number: '', bank_account_name: '',
      notes: '',
    };
    showVendorForm = true;
  }

  function openEditVendor(v) {
    editingVendor = v;
    vendorForm = {
      name: v.name, type: v.type, npwp: v.npwp || '', address: v.address || '',
      pic_name: v.pic_name || '', pic_phone: v.pic_phone || '', pic_email: v.pic_email || '',
      bank_name: v.bank_name || '', bank_account_number: v.bank_account_number || '',
      bank_account_name: v.bank_account_name || '', notes: v.notes || '',
    };
    showVendorForm = true;
  }

  async function saveVendor() {
    if (!vendorForm.name) { showToast('Nama vendor wajib diisi', 'warning'); return; }
    savingVendor = true;
    try {
      if (editingVendor) {
        await ApiService.updateVendor(editingVendor.id, vendorForm);
        showToast('Vendor berhasil diupdate', 'success');
      } else {
        await ApiService.createVendor(vendorForm);
        showToast('Vendor berhasil dibuat', 'success');
      }
      showVendorForm = false;
      await loadVendors();
    } catch (e) {
      showToast(e.message || 'Gagal menyimpan vendor', 'error');
    } finally {
      savingVendor = false;
    }
  }

  async function deleteVendor(v) {
    if (!confirm(`Hapus vendor "${v.name}"? Tindakan ini tidak dapat dibatalkan.`)) return;
    try {
      await ApiService.deleteVendor(v.id);
      showToast('Vendor berhasil dihapus', 'success');
      if (selectedVendor?.id === v.id) { showDetail = false; selectedVendor = null; }
      await loadAll();
    } catch (e) {
      showToast(e.message || 'Gagal menghapus vendor', 'error');
    }
  }

  // ── Vendor detail ────────────────────────────────────────
  async function openDetail(v) {
    selectedVendor = v;
    detailTab = 'info';
    showDetail = true;
    vendorBills = [];
    billFilter = 'all';
    expandedBillId = null;
    showPaymentForm = false;
    await loadVendorBills(v.id);
  }

  async function loadVendorBills(vendorId) {
    billsLoading = true;
    try {
      const data = await ApiService.listBills({ vendorId, pageSize: 100 });
      vendorBills = Array.isArray(data) ? data : (data?.data ?? []);
    } catch (e) {
      showToast(e.message || 'Gagal memuat tagihan', 'error');
    } finally {
      billsLoading = false;
    }
  }

  function getVendorDebt(v) {
    if (!debtSummary) return '—';
    return formatIDR(debtSummary.total_outstanding_idr || 0);
  }

  // ── Bill form ────────────────────────────────────────────
  async function openCreateBill(vendorId) {
    if (packages.length === 0) {
      packagesLoading = true;
      try {
        const data = await ApiService.listPackages({ pageSize: 200 });
        packages = Array.isArray(data) ? data : (data?.data ?? []);
      } catch { /* packages not critical */ } finally {
        packagesLoading = false;
      }
    }
    billForm = {
      vendor_id: vendorId, package_id: '', description: '',
      amount: 0, currency: 'IDR', exchange_rate: 1, due_date: '',
    };
    showBillForm = true;
  }

  async function saveBill() {
    if (!billForm.description) { showToast('Deskripsi tagihan wajib diisi', 'warning'); return; }
    if (!billForm.amount || billForm.amount < 1) { showToast('Nominal tagihan wajib diisi', 'warning'); return; }
    savingBill = true;
    try {
      await ApiService.createBill(billForm);
      showToast('Tagihan berhasil dibuat', 'success');
      showBillForm = false;
      if (selectedVendor) await loadVendorBills(selectedVendor.id);
    } catch (e) {
      showToast(e.message || 'Gagal membuat tagihan', 'error');
    } finally {
      savingBill = false;
    }
  }

  // ── Bill detail / payments ───────────────────────────────
  async function toggleBillDetail(billId) {
    if (expandedBillId === billId) { expandedBillId = null; return; }
    expandedBillId = billId;
    paymentsLoading = true;
    try {
      const data = await ApiService.listPaymentsByBill(billId);
      billPayments = Array.isArray(data) ? data : (data?.data ?? []);
    } catch (e) {
      billPayments = [];
      showToast(e.message || 'Gagal memuat riwayat pembayaran', 'error');
    } finally {
      paymentsLoading = false;
    }
  }

  function openPayForm(billId) {
    payBillId = billId;
    payForm = { amount: 0, payment_date: new Date().toISOString().slice(0, 10), source_account: '', notes: '' };
    showPaymentForm = true;
  }

  async function savePayment() {
    if (!payForm.amount || payForm.amount < 1) { showToast('Nominal pembayaran wajib diisi', 'warning'); return; }
    savingPayment = true;
    try {
      await ApiService.createPayment({
        vendor_bill_id: payBillId,
        amount: payForm.amount,
        payment_date: payForm.payment_date,
        source_account: payForm.source_account,
        notes: payForm.notes,
      });
      showToast('Pembayaran berhasil dicatat', 'success');
      showPaymentForm = false;
      if (selectedVendor) await loadVendorBills(selectedVendor.id);
      if (expandedBillId === payBillId) {
        const data = await ApiService.listPaymentsByBill(payBillId);
        billPayments = Array.isArray(data) ? data : (data?.data ?? []);
      }
    } catch (e) {
      showToast(e.message || 'Gagal mencatat pembayaran', 'error');
    } finally {
      savingPayment = false;
    }
  }

  // ── Utilities ─────────────────────────────────────────────
  function formatIDR(num) {
    return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(num || 0);
  }

  function formatDate(d) {
    if (!d) return '—';
    return new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
  }

  const typeLabel = {
    maskapai: 'Maskapai', hotel: 'Hotel', transport: 'Transport',
    perlengkapan: 'Perlengkapan', katering: 'Katering', lainnya: 'Lainnya',
  };

  const typeColor = {
    maskapai: 'bg-blue-50 text-blue-700',
    hotel: 'bg-emerald-50 text-emerald-700',
    transport: 'bg-amber-50 text-amber-700',
    perlengkapan: 'bg-purple-50 text-purple-700',
    katering: 'bg-rose-50 text-rose-700',
    lainnya: 'bg-slate-100 text-slate-600',
  };
</script>

<div class="flex h-screen flex-col">
  <!-- Header -->
  <div class="flex-shrink-0 border-b border-slate-100 bg-white px-6 py-5">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-serif text-xl font-bold text-slate-800">Vendor & Biaya Ops</h1>
        <p class="mt-0.5 text-sm text-slate-500">Kelola vendor dan catat pengeluaran operasional per trip</p>
      </div>
      {#if !isLoading}
        <button
          type="button"
          onclick={openCreateVendor}
          class="flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm shadow-primary-600/30 transition-all hover:bg-primary-700"
        >
          <Plus class="h-4 w-4" /> Tambah Vendor
        </button>
      {/if}
    </div>

    <!-- Summary cards -->
    <div class="mt-4 grid grid-cols-3 gap-3">
      <div class="rounded-xl bg-red-50 p-3">
        <p class="text-[11px] font-semibold text-red-400">Total Outstanding</p>
        <p class="mt-0.5 text-base font-bold text-red-700">
          {#if isLoading}
            <span class="inline-block h-5 w-24 animate-pulse rounded bg-red-200"></span>
          {:else}
            {formatIDR(debtSummary?.total_outstanding_idr || 0)}
          {/if}
        </p>
      </div>
      <div class="rounded-xl bg-amber-50 p-3">
        <p class="text-[11px] font-semibold text-amber-400">Overdue</p>
        <p class="mt-0.5 text-base font-bold text-amber-700">
          {#if isLoading}
            <span class="inline-block h-5 w-16 animate-pulse rounded bg-amber-200"></span>
          {:else}
            {overdueCount} tagihan
          {/if}
        </p>
      </div>
      <div class="rounded-xl bg-blue-50 p-3">
        <p class="text-[11px] font-semibold text-blue-400">Total Vendor</p>
        <p class="mt-0.5 text-base font-bold text-blue-700">
          {#if isLoading}
            <span class="inline-block h-5 w-12 animate-pulse rounded bg-blue-200"></span>
          {:else}
            {vendors.length}
          {/if}
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
          placeholder="Cari nama vendor atau PIC..."
          class="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-9 pr-3 text-sm outline-none focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>
      <div class="flex gap-1 overflow-x-auto">
        {#each VENDOR_TYPES as t}
          <button
            type="button"
            onclick={() => (filterType = t.id)}
            class="flex-shrink-0 rounded-lg px-3 py-2 text-xs font-semibold transition-all
              {filterType === t.id ? 'bg-primary-600 text-white' : 'text-slate-500 hover:bg-slate-100'}"
          >
            {t.label}
          </button>
        {/each}
      </div>
    </div>
  </div>

  <!-- Vendor table -->
  <div class="flex-1 overflow-auto">
    {#if isLoading}
      <div class="space-y-3 p-6">
        {#each [1,2,3,4,5] as _}
          <div class="h-16 animate-pulse rounded-xl bg-slate-100"></div>
        {/each}
      </div>
    {:else if filteredVendors.length === 0}
      <div class="flex flex-col items-center justify-center py-24 text-slate-400">
        <Building2 class="mb-3 h-12 w-12 opacity-30" />
        <p class="font-medium">Tidak ada vendor</p>
        <p class="mt-1 text-sm">Klik "Tambah Vendor" untuk menambahkan vendor pertama</p>
      </div>
    {:else}
      <table class="w-full min-w-[700px]">
        <thead class="sticky top-0 bg-slate-50">
          <tr class="text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
            <th class="px-6 py-3">Vendor</th>
            <th class="px-4 py-3">Tipe</th>
            <th class="px-4 py-3">PIC</th>
            <th class="px-4 py-3">Kontak</th>
            <th class="px-4 py-3"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-50">
          {#each filteredVendors as v}
            <tr
              class="group bg-white transition-colors hover:bg-primary-50/30 cursor-pointer"
              onclick={() => openDetail(v)}
            >
              <td class="px-6 py-3.5">
                <p class="text-sm font-semibold text-slate-800">{v.name}</p>
                {#if v.npwp}
                  <p class="text-xs text-slate-400">NPWP: {v.npwp}</p>
                {/if}
              </td>
              <td class="px-4 py-3.5">
                <span class="inline-block rounded-md px-2 py-0.5 text-xs font-semibold {typeColor[v.type] || typeColor.lainnya}">
                  {typeLabel[v.type] || v.type}
                </span>
              </td>
              <td class="px-4 py-3.5">
                <p class="text-sm text-slate-700">{v.pic_name || '—'}</p>
              </td>
              <td class="px-4 py-3.5">
                <p class="text-sm text-slate-600">{v.pic_phone || '—'}</p>
                {#if v.pic_email}
                  <p class="text-xs text-slate-400">{v.pic_email}</p>
                {/if}
              </td>
              <td class="px-4 py-3.5 text-right">
                <button
                  type="button"
                  onclick={(e) => { e.stopPropagation(); openDetail(v); }}
                  class="flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-semibold text-primary-600 transition-colors hover:bg-primary-50"
                >
                  Detail <ChevronRight class="h-3 w-3" />
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<!-- Create / Edit Vendor Drawer -->
<SlideDrawer
  open={showVendorForm}
  title={editingVendor ? 'Edit Vendor' : 'Tambah Vendor'}
  width="520px"
  onClose={() => (showVendorForm = false)}
>
  <div class="p-6 space-y-4">
    <div class="flex flex-col gap-1">
      <label for="v-name" class="text-sm font-medium text-slate-700">Nama Vendor <span class="text-red-500">*</span></label>
      <input
        id="v-name"
        type="text" bind:value={vendorForm.name}
        class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
      />
    </div>

    <div class="flex flex-col gap-1">
      <label for="v-type" class="text-sm font-medium text-slate-700">Tipe</label>
      <select
        id="v-type"
        bind:value={vendorForm.type}
        class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
      >
        {#each VENDOR_TYPES.slice(1) as t}
          <option value={t.id}>{t.label}</option>
        {/each}
      </select>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="v-npwp" class="text-sm font-medium text-slate-700">NPWP</label>
        <input
          id="v-npwp"
          type="text" bind:value={vendorForm.npwp}
          class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
        />
      </div>
      <div class="flex flex-col gap-1">
        <label for="v-pic-phone" class="text-sm font-medium text-slate-700">Telepon PIC</label>
        <input
          id="v-pic-phone"
          type="text" bind:value={vendorForm.pic_phone}
          class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
        />
      </div>
    </div>

    <div class="flex flex-col gap-1">
      <label for="v-address" class="text-sm font-medium text-slate-700">Alamat</label>
      <textarea
        id="v-address"
        bind:value={vendorForm.address}
        rows="2"
        class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400 resize-none"
      ></textarea>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="v-pic-name" class="text-sm font-medium text-slate-700">Nama PIC</label>
        <input
          id="v-pic-name"
          type="text" bind:value={vendorForm.pic_name}
          class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
        />
      </div>
      <div class="flex flex-col gap-1">
        <label for="v-pic-email" class="text-sm font-medium text-slate-700">Email PIC</label>
        <input
          id="v-pic-email"
          type="email" bind:value={vendorForm.pic_email}
          class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
        />
      </div>
    </div>

    <div class="rounded-xl bg-slate-50 p-4 space-y-3">
      <h4 class="text-xs font-bold uppercase tracking-wider text-slate-400">Informasi Bank</h4>
      <div class="flex flex-col gap-1">
        <label for="v-bank-name" class="text-sm font-medium text-slate-700">Nama Bank</label>
        <input
          id="v-bank-name"
          type="text" bind:value={vendorForm.bank_name}
          class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
        />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div class="flex flex-col gap-1">
          <label for="v-bank-account-number" class="text-sm font-medium text-slate-700">No. Rekening</label>
          <input
            id="v-bank-account-number"
            type="text" bind:value={vendorForm.bank_account_number}
            class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
          />
        </div>
        <div class="flex flex-col gap-1">
          <label for="v-bank-account-name" class="text-sm font-medium text-slate-700">Atas Nama</label>
          <input
            id="v-bank-account-name"
            type="text" bind:value={vendorForm.bank_account_name}
            class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
          />
        </div>
      </div>
    </div>

    <div class="flex flex-col gap-1">
      <label for="v-notes" class="text-sm font-medium text-slate-700">Catatan</label>
      <textarea
        id="v-notes"
        bind:value={vendorForm.notes}
        rows="2"
        class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400 resize-none"
      ></textarea>
    </div>

    <div class="flex gap-2 pt-2">
      <button
        type="button"
        onclick={() => (showVendorForm = false)}
        class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600 hover:bg-slate-50"
      >
        Batal
      </button>
      <button
        type="button"
        onclick={saveVendor}
        disabled={savingVendor}
        class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
      >
        {savingVendor ? 'Menyimpan...' : 'Simpan'}
      </button>
    </div>
  </div>
</SlideDrawer>

<!-- Vendor Detail Drawer -->
<SlideDrawer
  open={showDetail}
  title={selectedVendor?.name || ''}
  width="580px"
  onClose={() => { showDetail = false; selectedVendor = null; }}
>
  {#if selectedVendor}
    <div class="p-6">
      <!-- Tabs -->
      <div class="mb-5 flex gap-1">
        <button
          type="button"
          onclick={() => (detailTab = 'info')}
          class="rounded-lg px-4 py-1.5 text-xs font-semibold transition-all
            {detailTab === 'info' ? 'bg-primary-600 text-white' : 'text-slate-500 hover:bg-slate-100'}"
        >
          Info Vendor
        </button>
        <button
          type="button"
          onclick={() => (detailTab = 'bills')}
          class="rounded-lg px-4 py-1.5 text-xs font-semibold transition-all
            {detailTab === 'bills' ? 'bg-primary-600 text-white' : 'text-slate-500 hover:bg-slate-100'}"
        >
          Tagihan ({vendorBills.length})
        </button>
      </div>

      {#if detailTab === 'info'}
        <!-- Vendor Info -->
        <div class="space-y-4">
          <div class="rounded-xl bg-slate-50 p-4 space-y-2">
            <div class="flex items-center justify-between">
              <span class="inline-block rounded-md px-2 py-0.5 text-xs font-semibold {typeColor[selectedVendor.type] || typeColor.lainnya}">
                {typeLabel[selectedVendor.type] || selectedVendor.type}
              </span>
              <div class="flex gap-1">
                <button
                  type="button"
                  onclick={() => { showDetail = false; openEditVendor(selectedVendor); }}
                  class="flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-semibold text-primary-600 hover:bg-primary-50"
                >
                  <Pencil class="h-3.5 w-3.5" /> Edit
                </button>
                <button
                  type="button"
                  onclick={() => deleteVendor(selectedVendor)}
                  class="flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-semibold text-red-600 hover:bg-red-50"
                >
                  <Trash2 class="h-3.5 w-3.5" /> Hapus
                </button>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
              {#if selectedVendor.npwp}
                <div><span class="text-slate-400">NPWP</span><br /><span class="font-medium text-slate-700">{selectedVendor.npwp}</span></div>
              {/if}
              {#if selectedVendor.pic_name}
                <div><span class="text-slate-400">PIC</span><br /><span class="font-medium text-slate-700">{selectedVendor.pic_name}</span></div>
              {/if}
              {#if selectedVendor.pic_phone}
                <div><span class="text-slate-400">Telepon</span><br /><span class="font-medium text-slate-700">{selectedVendor.pic_phone}</span></div>
              {/if}
              {#if selectedVendor.pic_email}
                <div><span class="text-slate-400">Email</span><br /><span class="font-medium text-slate-700">{selectedVendor.pic_email}</span></div>
              {/if}
            </div>

            {#if selectedVendor.address}
              <div class="text-sm">
                <span class="text-slate-400">Alamat</span><br />
                <span class="font-medium text-slate-700">{selectedVendor.address}</span>
              </div>
            {/if}
          </div>

          {#if selectedVendor.bank_name || selectedVendor.bank_account_number}
            <div class="rounded-xl border border-slate-100 p-4 space-y-2">
              <h4 class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Informasi Bank</h4>
              <div class="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                {#if selectedVendor.bank_name}
                  <div><span class="text-slate-400">Bank</span><br /><span class="font-medium text-slate-700">{selectedVendor.bank_name}</span></div>
                {/if}
                {#if selectedVendor.bank_account_number}
                  <div><span class="text-slate-400">No. Rekening</span><br /><span class="font-medium text-slate-700">{selectedVendor.bank_account_number}</span></div>
                {/if}
                {#if selectedVendor.bank_account_name}
                  <div><span class="text-slate-400">Atas Nama</span><br /><span class="font-medium text-slate-700">{selectedVendor.bank_account_name}</span></div>
                {/if}
              </div>
            </div>
          {/if}

          {#if selectedVendor.notes}
            <div class="rounded-xl border border-slate-100 p-4">
              <h4 class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Catatan</h4>
              <p class="mt-1 text-sm text-slate-600">{selectedVendor.notes}</p>
            </div>
          {/if}
        </div>

      {:else if detailTab === 'bills'}
        <!-- Bills List -->
        <div>
          <div class="mb-4 flex items-center justify-between">
            <div class="flex gap-1">
              {#each BILL_STATUSES as s}
                <button
                  type="button"
                  onclick={() => (billFilter = s.id)}
                  class="rounded-lg px-3 py-1.5 text-[11px] font-semibold transition-all
                    {billFilter === s.id ? 'bg-primary-600 text-white' : 'text-slate-500 hover:bg-slate-100'}"
                >
                  {s.label}
                </button>
              {/each}
            </div>
            <button
              type="button"
              onclick={() => openCreateBill(selectedVendor.id)}
              class="flex items-center gap-1 rounded-lg bg-primary-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-primary-700"
            >
              <Plus class="h-3.5 w-3.5" /> Tagihan
            </button>
          </div>

          {#if billsLoading}
            <div class="space-y-2">
              {#each [1,2] as _}
                <div class="h-20 animate-pulse rounded-xl bg-slate-100"></div>
              {/each}
            </div>
          {:else if filteredBills.length === 0}
            <div class="flex flex-col items-center justify-center py-12 text-slate-400">
              <Banknote class="mb-2 h-8 w-8 opacity-30" />
              <p class="text-sm font-medium">Belum ada tagihan</p>
            </div>
          {:else}
            <div class="space-y-2">
              {#each filteredBills as bill}
                <div
                  role="button"
                  tabindex="0"
                  class="rounded-xl border {expandedBillId === bill.id ? 'border-primary-200 bg-primary-50/50' : 'border-slate-100 bg-white'} overflow-hidden transition-all cursor-pointer outline-none focus:ring-2 focus:ring-primary-400"
                  onclick={() => toggleBillDetail(bill.id)}
                  onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggleBillDetail(bill.id); } }}
                >
                  <div class="p-4">
                    <div class="flex items-start justify-between">
                      <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2">
                          <span class="text-sm font-semibold text-slate-800">{bill.description}</span>
                          <StatusBadge status={bill.status} size="xs" />
                        </div>
                        <div class="mt-1 flex items-center gap-3 text-xs text-slate-400">
                          {#if bill.due_date}
                            <span class="flex items-center gap-1"><Calendar class="h-3 w-3" /> Jatuh tempo {formatDate(bill.due_date)}</span>
                          {/if}
                          {#if bill.vendor_name}
                            <span>{bill.vendor_name}</span>
                          {/if}
                        </div>
                      </div>
                      <div class="text-right flex-shrink-0 ml-4">
                        <p class="text-sm font-bold text-slate-800">{formatIDR(bill.amount_idr || bill.amount)}</p>
                        <p class="text-xs text-slate-400">
                          Terbayar {formatIDR(bill.paid_amount)} · Sisa <span class="font-semibold text-red-600">{formatIDR((bill.amount_idr || bill.amount) - bill.paid_amount)}</span>
                        </p>
                      </div>
                    </div>
                    <!-- Progress -->
                    <div class="mt-2 h-1.5 overflow-hidden rounded-full bg-slate-100">
                      <div
                        class="h-full rounded-full {bill.status === 'lunas' ? 'bg-emerald-400' : 'bg-amber-400'}"
                        style="width: {Math.min(100, Math.round((bill.paid_amount / Math.max(bill.amount_idr || bill.amount, 1)) * 100))}%"
                      ></div>
                    </div>
                  </div>

                  <!-- Expanded: Payments -->
                  {#if expandedBillId === bill.id}
                    <div class="border-t border-primary-100 px-4 py-3 space-y-3">
                      {#if paymentsLoading}
                        <div class="h-12 animate-pulse rounded-lg bg-slate-100"></div>
                      {:else}
                        <h5 class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Riwayat Pembayaran</h5>
                        {#if billPayments.length === 0}
                          <p class="text-xs text-slate-400">Belum ada pembayaran</p>
                        {:else}
                          <div class="space-y-1.5">
                            {#each billPayments as p}
                              <div class="flex items-center justify-between rounded-lg bg-white px-3 py-2">
                                <div>
                                  <p class="text-xs font-semibold text-slate-700">{formatDate(p.payment_date)}</p>
                                  {#if p.source_account}
                                    <p class="text-[11px] text-slate-400">Dari: {p.source_account}</p>
                                  {/if}
                                </div>
                                <div class="text-right">
                                  <p class="text-sm font-bold text-emerald-600">{formatIDR(p.amount_idr || p.amount)}</p>
                                </div>
                              </div>
                            {/each}
                          </div>
                        {/if}
                      {/if}

                      {#if bill.status !== 'lunas'}
                        {#if !showPaymentForm || payBillId !== bill.id}
                          <button
                            type="button"
                            onclick={(e) => { e.stopPropagation(); openPayForm(bill.id); }}
                            class="flex w-full items-center justify-center gap-2 rounded-xl border border-primary-200 py-2 text-xs font-semibold text-primary-700 hover:bg-primary-100"
                          >
                            <Plus class="h-3.5 w-3.5" /> Rekam Pembayaran
                          </button>
                        {/if}

                        {#if showPaymentForm && payBillId === bill.id}
                          <div role="presentation" class="rounded-xl bg-white p-3 border border-primary-100 space-y-2" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
                            <h5 class="text-xs font-bold text-primary-800">Pembayaran Baru</h5>
                            <IDRInput label="Nominal" bind:value={payForm.amount} required />
                            <div class="flex flex-col gap-1">
                              <label for="pay-bill-date" class="text-xs font-medium text-slate-700">Tanggal Bayar</label>
                              <input
                                id="pay-bill-date"
                                type="date" bind:value={payForm.payment_date}
                                class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-primary-400"
                              />
                            </div>
                            <div class="flex flex-col gap-1">
                              <label for="pay-bill-source" class="text-xs font-medium text-slate-700">Sumber Dana</label>
                              <input
                                id="pay-bill-source"
                                type="text" bind:value={payForm.source_account}
                                placeholder="Kas / Rekening..."
                                class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-primary-400"
                              />
                            </div>
                            <div class="flex gap-2 pt-1">
                              <button
                                type="button"
                                onclick={(e) => { e.stopPropagation(); showPaymentForm = false; }}
                                class="flex-1 rounded-lg border border-slate-200 py-1.5 text-xs font-semibold text-slate-600 hover:bg-slate-50"
                              >
                                Batal
                              </button>
                              <button
                                type="button"
                                onclick={(e) => { e.stopPropagation(); savePayment(); }}
                                disabled={savingPayment}
                                class="flex-1 rounded-lg bg-primary-600 py-1.5 text-xs font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
                              >
                                {savingPayment ? '...' : 'Simpan'}
                              </button>
                            </div>
                          </div>
                        {/if}
                      {/if}
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  {/if}
</SlideDrawer>

<!-- Create Bill Drawer -->
<SlideDrawer
  open={showBillForm}
  title="Tambah Tagihan"
  width="480px"
  onClose={() => (showBillForm = false)}
>
  <div class="p-6 space-y-4">
    <div class="flex flex-col gap-1">
      <label for="b-desc" class="text-sm font-medium text-slate-700">Deskripsi <span class="text-red-500">*</span></label>
      <input
        id="b-desc"
        type="text" bind:value={billForm.description}
        placeholder="e.g. Tiket pesawat Umroh Ramadan"
        class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
      />
    </div>

    <IDRInput label="Nominal" bind:value={billForm.amount} required />

    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="b-curr" class="text-sm font-medium text-slate-700">Mata Uang</label>
        <select
          id="b-curr"
          bind:value={billForm.currency}
          class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
        >
          <option value="IDR">IDR</option>
          <option value="USD">USD</option>
          <option value="SAR">SAR</option>
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label for="b-rate" class="text-sm font-medium text-slate-700">Kurs</label>
        <input
          id="b-rate"
          type="number" bind:value={billForm.exchange_rate}
          step="0.01"
          class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
        />
      </div>
    </div>

    <div class="flex flex-col gap-1">
      <label for="b-due" class="text-sm font-medium text-slate-700">Jatuh Tempo</label>
      <input
        id="b-due"
        type="date" bind:value={billForm.due_date}
        class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
      />
    </div>

    {#if packages.length > 0}
      <div class="flex flex-col gap-1">
        <label for="b-pkg" class="text-sm font-medium text-slate-700">Trip (opsional)</label>
        <select
          id="b-pkg"
          bind:value={billForm.package_id}
          class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
        >
          <option value="">— Pilih Trip —</option>
          {#each packages as pkg}
            <option value={pkg.id}>{pkg.name}</option>
          {/each}
        </select>
      </div>
    {/if}

    <div class="flex gap-2 pt-2">
      <button
        type="button"
        onclick={() => (showBillForm = false)}
        class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600 hover:bg-slate-50"
      >
        Batal
      </button>
      <button
        type="button"
        onclick={saveBill}
        disabled={savingBill}
        class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50"
      >
        {savingBill ? 'Menyimpan...' : 'Simpan'}
      </button>
    </div>
  </div>
</SlideDrawer>
