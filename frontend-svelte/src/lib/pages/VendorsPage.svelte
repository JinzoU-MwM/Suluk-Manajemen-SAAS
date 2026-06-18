<script>
  import { onMount } from 'svelte';
  import {
    Building2, Plus, Search, AlertCircle, ChevronRight,
    Pencil, Trash2, Banknote, Calendar,
    Truck, Wallet, Boxes,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import IDRInput from '../components/IDRInput.svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import EmptyState from '../components/EmptyState.svelte';
  import Badge from '../components/ui/Badge.svelte';
  import Button from '../components/ui/Button.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
  import ProgressBar from '../components/ui/ProgressBar.svelte';
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

  // type → Badge tone (success|info|warning|danger|muted)
  const typeTone = {
    maskapai: 'info',
    hotel: 'success',
    transport: 'warning',
    perlengkapan: 'info',
    katering: 'danger',
    lainnya: 'muted',
  };

  const typeIcon = {
    maskapai: Truck, hotel: Building2, transport: Truck,
    perlengkapan: Boxes, katering: Boxes, lainnya: Building2,
  };

  // FilterTabs expects { value, label, count? }
  let typeTabs = $derived(VENDOR_TYPES.map(t => ({ value: t.id, label: t.label })));
  let billTabs = $derived(BILL_STATUSES.map(s => ({ value: s.id, label: s.label })));
</script>

<div class="flex h-full flex-col" style="background:var(--c-bg)">
  <!-- Header -->
  <div class="flex-shrink-0 px-4 pt-6 lg:px-8" style="background:var(--c-bg)">
    <PageHeader
      kicker="Operasional"
      title="Vendor &amp; Pemasok"
      subtitle="Kelola maskapai, hotel, transport, dan pemasok lain beserta tagihan biaya operasional per trip."
    >
      {#snippet actions()}
        {#if !isLoading}
          <Button variant="primary" icon={Plus} onclick={openCreateVendor}>Tambah Vendor</Button>
        {/if}
      {/snippet}
    </PageHeader>

    <!-- Summary cards (Suluk design) -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard icon={Truck} label="Total Vendor" value={`${vendors.length}`} accent="var(--c-primary)" />
      <StatCard icon={Wallet} label="Total Outstanding" value={formatIDR(debtSummary?.total_outstanding_idr || 0)} accent="var(--c-danger)" />
      <StatCard icon={AlertCircle} label="Overdue" value={`${overdueCount} tagihan`} accent="var(--c-accent)" />
      <StatCard icon={Boxes} label="Kategori" value={`${VENDOR_TYPES.length - 1}`} accent="var(--c-info)" />
    </div>

    <!-- Search + filter -->
    <div class="mt-5 flex flex-wrap items-center justify-between gap-3">
      <FilterTabs tabs={typeTabs} value={filterType} onChange={(v) => (filterType = v)} />
      <div class="relative min-w-[220px] flex-1 sm:max-w-xs">
        <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" style="color:var(--c-faint)" />
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Cari nama vendor atau PIC…"
          class="w-full rounded-xl py-2.5 pl-9 pr-3 text-sm outline-none transition-colors"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        />
      </div>
    </div>
  </div>

  <!-- Vendor table -->
  <div class="flex-1 overflow-auto px-4 py-6 lg:px-8">
    {#if isLoading}
      <div class="space-y-3">
        {#each [1,2,3,4,5] as _}
          <div class="h-16 animate-pulse rounded-2xl" style="background:var(--c-bg-2)"></div>
        {/each}
      </div>
    {:else if filteredVendors.length === 0}
      <div class="rounded-2xl" style="border:1px solid var(--c-line);background:var(--c-surface);box-shadow:var(--shadow-sm)">
        <EmptyState
          icon={Building2}
          title="Tidak ada vendor"
          text={'Klik "Tambah Vendor" untuk menambahkan vendor pertama.'}
        />
      </div>
    {:else}
      <div class="overflow-hidden rounded-2xl" style="border:1px solid var(--c-line);background:var(--c-surface);box-shadow:var(--shadow-sm)">
        <div class="overflow-x-auto">
          <table class="w-full min-w-[720px] text-sm">
            <thead>
              <tr class="text-left" style="border-bottom:1px solid var(--c-line-soft)">
                <th class="px-6 py-3.5 text-[11.5px] font-semibold uppercase tracking-wide" style="color:var(--c-faint)">Vendor</th>
                <th class="px-4 py-3.5 text-[11.5px] font-semibold uppercase tracking-wide" style="color:var(--c-faint)">Kategori</th>
                <th class="px-4 py-3.5 text-[11.5px] font-semibold uppercase tracking-wide" style="color:var(--c-faint)">PIC</th>
                <th class="px-4 py-3.5 text-[11.5px] font-semibold uppercase tracking-wide" style="color:var(--c-faint)">Kontak</th>
                <th class="px-6 py-3.5 text-right text-[11.5px] font-semibold uppercase tracking-wide" style="color:var(--c-faint)">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {#each filteredVendors as v}
                {@const VIcon = typeIcon[v.type] || Building2}
                <tr
                  class="suluk-row cursor-pointer transition-colors"
                  style="border-bottom:1px solid var(--c-line-soft)"
                  onclick={() => openDetail(v)}
                >
                  <td class="px-6 py-3.5">
                    <div class="flex items-center gap-3">
                      <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl" style="background:var(--c-accent-soft);color:var(--c-accent)">
                        <VIcon class="h-[19px] w-[19px]" />
                      </div>
                      <div class="min-w-0">
                        <p class="truncate text-sm font-bold" style="color:var(--c-ink)">{v.name}</p>
                        {#if v.npwp}
                          <p class="truncate text-xs" style="color:var(--c-faint)">NPWP: {v.npwp}</p>
                        {/if}
                      </div>
                    </div>
                  </td>
                  <td class="px-4 py-3.5">
                    <Badge tone={typeTone[v.type] || 'muted'} label={typeLabel[v.type] || v.type} />
                  </td>
                  <td class="px-4 py-3.5">
                    <p class="text-sm" style="color:var(--c-ink-soft)">{v.pic_name || '—'}</p>
                  </td>
                  <td class="px-4 py-3.5">
                    <p class="text-sm tabular" style="color:var(--c-muted)">{v.pic_phone || '—'}</p>
                    {#if v.pic_email}
                      <p class="text-xs" style="color:var(--c-faint)">{v.pic_email}</p>
                    {/if}
                  </td>
                  <td class="px-6 py-3.5 text-right">
                    <button
                      type="button"
                      onclick={(e) => { e.stopPropagation(); openDetail(v); }}
                      class="inline-flex items-center gap-1 rounded-lg px-3 py-1.5 text-xs font-semibold transition-colors"
                      style="color:var(--c-primary)"
                    >
                      Detail <ChevronRight class="h-3 w-3" />
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
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
      <label for="v-name" class="text-sm font-medium" style="color:var(--c-ink-soft)">Nama Vendor <span style="color:var(--c-danger)">*</span></label>
      <input
        id="v-name"
        type="text" bind:value={vendorForm.name}
        class="rounded-xl px-3 py-2.5 text-sm outline-none"
        style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
      />
    </div>

    <div class="flex flex-col gap-1">
      <label for="v-type" class="text-sm font-medium" style="color:var(--c-ink-soft)">Tipe</label>
      <select
        id="v-type"
        bind:value={vendorForm.type}
        class="rounded-xl px-3 py-2.5 text-sm outline-none"
        style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
      >
        {#each VENDOR_TYPES.slice(1) as t}
          <option value={t.id}>{t.label}</option>
        {/each}
      </select>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="v-npwp" class="text-sm font-medium" style="color:var(--c-ink-soft)">NPWP</label>
        <input
          id="v-npwp"
          type="text" bind:value={vendorForm.npwp}
          class="rounded-xl px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        />
      </div>
      <div class="flex flex-col gap-1">
        <label for="v-pic-phone" class="text-sm font-medium" style="color:var(--c-ink-soft)">Telepon PIC</label>
        <input
          id="v-pic-phone"
          type="text" bind:value={vendorForm.pic_phone}
          class="rounded-xl px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        />
      </div>
    </div>

    <div class="flex flex-col gap-1">
      <label for="v-address" class="text-sm font-medium" style="color:var(--c-ink-soft)">Alamat</label>
      <textarea
        id="v-address"
        bind:value={vendorForm.address}
        rows="2"
        class="resize-none rounded-xl px-3 py-2.5 text-sm outline-none"
        style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
      ></textarea>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="v-pic-name" class="text-sm font-medium" style="color:var(--c-ink-soft)">Nama PIC</label>
        <input
          id="v-pic-name"
          type="text" bind:value={vendorForm.pic_name}
          class="rounded-xl px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        />
      </div>
      <div class="flex flex-col gap-1">
        <label for="v-pic-email" class="text-sm font-medium" style="color:var(--c-ink-soft)">Email PIC</label>
        <input
          id="v-pic-email"
          type="email" bind:value={vendorForm.pic_email}
          class="rounded-xl px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        />
      </div>
    </div>

    <div class="space-y-3 rounded-xl p-4" style="background:var(--c-bg-2)">
      <h4 class="text-xs font-bold uppercase tracking-wider" style="color:var(--c-faint)">Informasi Bank</h4>
      <div class="flex flex-col gap-1">
        <label for="v-bank-name" class="text-sm font-medium" style="color:var(--c-ink-soft)">Nama Bank</label>
        <input
          id="v-bank-name"
          type="text" bind:value={vendorForm.bank_name}
          class="rounded-xl px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        />
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div class="flex flex-col gap-1">
          <label for="v-bank-account-number" class="text-sm font-medium" style="color:var(--c-ink-soft)">No. Rekening</label>
          <input
            id="v-bank-account-number"
            type="text" bind:value={vendorForm.bank_account_number}
            class="rounded-xl px-3 py-2.5 text-sm outline-none"
            style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
          />
        </div>
        <div class="flex flex-col gap-1">
          <label for="v-bank-account-name" class="text-sm font-medium" style="color:var(--c-ink-soft)">Atas Nama</label>
          <input
            id="v-bank-account-name"
            type="text" bind:value={vendorForm.bank_account_name}
            class="rounded-xl px-3 py-2.5 text-sm outline-none"
            style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
          />
        </div>
      </div>
    </div>

    <div class="flex flex-col gap-1">
      <label for="v-notes" class="text-sm font-medium" style="color:var(--c-ink-soft)">Catatan</label>
      <textarea
        id="v-notes"
        bind:value={vendorForm.notes}
        rows="2"
        class="resize-none rounded-xl px-3 py-2.5 text-sm outline-none"
        style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
      ></textarea>
    </div>

    <div class="flex gap-2 pt-2">
      <Button variant="ghost" full onclick={() => (showVendorForm = false)}>Batal</Button>
      <Button variant="primary" full disabled={savingVendor} onclick={saveVendor}>
        {savingVendor ? 'Menyimpan…' : 'Simpan'}
      </Button>
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
      <div class="mb-5">
        <FilterTabs
          tabs={[{ value: 'info', label: 'Info Vendor' }, { value: 'bills', label: 'Tagihan', count: vendorBills.length }]}
          value={detailTab}
          onChange={(v) => (detailTab = v)}
        />
      </div>

      {#if detailTab === 'info'}
        <!-- Vendor Info -->
        <div class="space-y-4">
          <div class="space-y-2 rounded-xl p-4" style="background:var(--c-bg-2)">
            <div class="flex items-center justify-between">
              <Badge tone={typeTone[selectedVendor.type] || 'muted'} label={typeLabel[selectedVendor.type] || selectedVendor.type} />
              <div class="flex gap-1.5">
                <Button variant="ghost" size="sm" icon={Pencil} onclick={() => { showDetail = false; openEditVendor(selectedVendor); }}>Edit</Button>
                <Button variant="danger" size="sm" icon={Trash2} onclick={() => deleteVendor(selectedVendor)}>Hapus</Button>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
              {#if selectedVendor.npwp}
                <div><span style="color:var(--c-faint)">NPWP</span><br /><span class="font-medium" style="color:var(--c-ink-soft)">{selectedVendor.npwp}</span></div>
              {/if}
              {#if selectedVendor.pic_name}
                <div><span style="color:var(--c-faint)">PIC</span><br /><span class="font-medium" style="color:var(--c-ink-soft)">{selectedVendor.pic_name}</span></div>
              {/if}
              {#if selectedVendor.pic_phone}
                <div><span style="color:var(--c-faint)">Telepon</span><br /><span class="font-medium" style="color:var(--c-ink-soft)">{selectedVendor.pic_phone}</span></div>
              {/if}
              {#if selectedVendor.pic_email}
                <div><span style="color:var(--c-faint)">Email</span><br /><span class="font-medium" style="color:var(--c-ink-soft)">{selectedVendor.pic_email}</span></div>
              {/if}
            </div>

            {#if selectedVendor.address}
              <div class="text-sm">
                <span style="color:var(--c-faint)">Alamat</span><br />
                <span class="font-medium" style="color:var(--c-ink-soft)">{selectedVendor.address}</span>
              </div>
            {/if}
          </div>

          {#if selectedVendor.bank_name || selectedVendor.bank_account_number}
            <div class="space-y-2 rounded-xl p-4" style="border:1px solid var(--c-line-soft)">
              <h4 class="text-[10px] font-bold uppercase tracking-wider" style="color:var(--c-faint)">Informasi Bank</h4>
              <div class="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                {#if selectedVendor.bank_name}
                  <div><span style="color:var(--c-faint)">Bank</span><br /><span class="font-medium" style="color:var(--c-ink-soft)">{selectedVendor.bank_name}</span></div>
                {/if}
                {#if selectedVendor.bank_account_number}
                  <div><span style="color:var(--c-faint)">No. Rekening</span><br /><span class="font-medium" style="color:var(--c-ink-soft)">{selectedVendor.bank_account_number}</span></div>
                {/if}
                {#if selectedVendor.bank_account_name}
                  <div><span style="color:var(--c-faint)">Atas Nama</span><br /><span class="font-medium" style="color:var(--c-ink-soft)">{selectedVendor.bank_account_name}</span></div>
                {/if}
              </div>
            </div>
          {/if}

          {#if selectedVendor.notes}
            <div class="rounded-xl p-4" style="border:1px solid var(--c-line-soft)">
              <h4 class="text-[10px] font-bold uppercase tracking-wider" style="color:var(--c-faint)">Catatan</h4>
              <p class="mt-1 text-sm" style="color:var(--c-muted)">{selectedVendor.notes}</p>
            </div>
          {/if}
        </div>

      {:else if detailTab === 'bills'}
        <!-- Bills List -->
        <div>
          <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
            <FilterTabs tabs={billTabs} value={billFilter} onChange={(v) => (billFilter = v)} />
            <Button variant="primary" size="sm" icon={Plus} onclick={() => openCreateBill(selectedVendor.id)}>Tagihan</Button>
          </div>

          {#if billsLoading}
            <div class="space-y-2">
              {#each [1,2] as _}
                <div class="h-20 animate-pulse rounded-xl" style="background:var(--c-bg-2)"></div>
              {/each}
            </div>
          {:else if filteredBills.length === 0}
            <EmptyState icon={Banknote} title="Belum ada tagihan" />
          {:else}
            <div class="space-y-2">
              {#each filteredBills as bill}
                {@const total = bill.amount_idr || bill.amount}
                <div
                  role="button"
                  tabindex="0"
                  class="overflow-hidden rounded-xl outline-none transition-all cursor-pointer"
                  style="border:1px solid {expandedBillId === bill.id ? 'var(--c-primary-soft)' : 'var(--c-line-soft)'};background:{expandedBillId === bill.id ? 'var(--c-primary-tint)' : 'var(--c-surface)'}"
                  onclick={() => toggleBillDetail(bill.id)}
                  onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggleBillDetail(bill.id); } }}
                >
                  <div class="p-4">
                    <div class="flex items-start justify-between">
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-2">
                          <span class="text-sm font-semibold" style="color:var(--c-ink)">{bill.description}</span>
                          <StatusBadge status={bill.status} size="xs" />
                        </div>
                        <div class="mt-1 flex items-center gap-3 text-xs" style="color:var(--c-faint)">
                          {#if bill.due_date}
                            <span class="flex items-center gap-1"><Calendar class="h-3 w-3" /> Jatuh tempo {formatDate(bill.due_date)}</span>
                          {/if}
                          {#if bill.vendor_name}
                            <span>{bill.vendor_name}</span>
                          {/if}
                        </div>
                      </div>
                      <div class="ml-4 flex-shrink-0 text-right">
                        <p class="text-sm font-bold tabular" style="color:var(--c-ink)">{formatIDR(total)}</p>
                        <p class="text-xs" style="color:var(--c-faint)">
                          Terbayar {formatIDR(bill.paid_amount)} · Sisa <span class="font-semibold" style="color:var(--c-danger)">{formatIDR(total - bill.paid_amount)}</span>
                        </p>
                      </div>
                    </div>
                    <!-- Progress -->
                    <div class="mt-2.5">
                      <ProgressBar
                        value={bill.paid_amount}
                        max={Math.max(total, 1)}
                        color={bill.status === 'lunas' ? 'var(--c-success)' : 'var(--c-accent)'}
                      />
                    </div>
                  </div>

                  <!-- Expanded: Payments -->
                  {#if expandedBillId === bill.id}
                    <div class="space-y-3 px-4 py-3" style="border-top:1px solid var(--c-primary-soft)">
                      {#if paymentsLoading}
                        <div class="h-12 animate-pulse rounded-lg" style="background:var(--c-bg-2)"></div>
                      {:else}
                        <h5 class="text-[10px] font-bold uppercase tracking-wider" style="color:var(--c-faint)">Riwayat Pembayaran</h5>
                        {#if billPayments.length === 0}
                          <p class="text-xs" style="color:var(--c-faint)">Belum ada pembayaran</p>
                        {:else}
                          <div class="space-y-1.5">
                            {#each billPayments as p}
                              <div class="flex items-center justify-between rounded-lg px-3 py-2" style="background:var(--c-surface)">
                                <div>
                                  <p class="text-xs font-semibold" style="color:var(--c-ink-soft)">{formatDate(p.payment_date)}</p>
                                  {#if p.source_account}
                                    <p class="text-[11px]" style="color:var(--c-faint)">Dari: {p.source_account}</p>
                                  {/if}
                                </div>
                                <div class="text-right">
                                  <p class="text-sm font-bold tabular" style="color:var(--c-success)">{formatIDR(p.amount_idr || p.amount)}</p>
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
                            class="flex w-full items-center justify-center gap-2 rounded-xl py-2 text-xs font-semibold transition-colors"
                            style="border:1px solid var(--c-primary-soft);color:var(--c-primary-deep)"
                          >
                            <Plus class="h-3.5 w-3.5" /> Rekam Pembayaran
                          </button>
                        {/if}

                        {#if showPaymentForm && payBillId === bill.id}
                          <div role="presentation" class="space-y-2 rounded-xl p-3" style="background:var(--c-surface);border:1px solid var(--c-primary-soft)" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
                            <h5 class="text-xs font-bold" style="color:var(--c-primary-deep)">Pembayaran Baru</h5>
                            <IDRInput label="Nominal" bind:value={payForm.amount} required />
                            <div class="flex flex-col gap-1">
                              <label for="pay-bill-date" class="text-xs font-medium" style="color:var(--c-ink-soft)">Tanggal Bayar</label>
                              <input
                                id="pay-bill-date"
                                type="date" bind:value={payForm.payment_date}
                                class="rounded-xl px-3 py-2 text-sm outline-none"
                                style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
                              />
                            </div>
                            <div class="flex flex-col gap-1">
                              <label for="pay-bill-source" class="text-xs font-medium" style="color:var(--c-ink-soft)">Sumber Dana</label>
                              <input
                                id="pay-bill-source"
                                type="text" bind:value={payForm.source_account}
                                placeholder="Kas / Rekening…"
                                class="rounded-xl px-3 py-2 text-sm outline-none"
                                style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
                              />
                            </div>
                            <div class="flex gap-2 pt-1">
                              <button
                                type="button"
                                onclick={(e) => { e.stopPropagation(); showPaymentForm = false; }}
                                class="flex-1 rounded-lg py-1.5 text-xs font-semibold"
                                style="border:1px solid var(--c-line);color:var(--c-muted)"
                              >
                                Batal
                              </button>
                              <button
                                type="button"
                                onclick={(e) => { e.stopPropagation(); savePayment(); }}
                                disabled={savingPayment}
                                class="flex-1 rounded-lg py-1.5 text-xs font-semibold text-white disabled:opacity-50"
                                style="background:var(--c-primary)"
                              >
                                {savingPayment ? '…' : 'Simpan'}
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
      <label for="b-desc" class="text-sm font-medium" style="color:var(--c-ink-soft)">Deskripsi <span style="color:var(--c-danger)">*</span></label>
      <input
        id="b-desc"
        type="text" bind:value={billForm.description}
        placeholder="e.g. Tiket pesawat Umroh Ramadan"
        class="rounded-xl px-3 py-2.5 text-sm outline-none"
        style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
      />
    </div>

    <IDRInput label="Nominal" bind:value={billForm.amount} required />

    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="b-curr" class="text-sm font-medium" style="color:var(--c-ink-soft)">Mata Uang</label>
        <select
          id="b-curr"
          bind:value={billForm.currency}
          class="rounded-xl px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        >
          <option value="IDR">IDR</option>
          <option value="USD">USD</option>
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label for="b-rate" class="text-sm font-medium" style="color:var(--c-ink-soft)">Kurs</label>
        <input
          id="b-rate"
          type="number" bind:value={billForm.exchange_rate}
          step="0.01"
          class="rounded-xl px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        />
      </div>
    </div>

    <div class="flex flex-col gap-1">
      <label for="b-due" class="text-sm font-medium" style="color:var(--c-ink-soft)">Jatuh Tempo</label>
      <input
        id="b-due"
        type="date" bind:value={billForm.due_date}
        class="rounded-xl px-3 py-2.5 text-sm outline-none"
        style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
      />
    </div>

    {#if packages.length > 0}
      <div class="flex flex-col gap-1">
        <label for="b-pkg" class="text-sm font-medium" style="color:var(--c-ink-soft)">Trip (opsional)</label>
        <select
          id="b-pkg"
          bind:value={billForm.package_id}
          class="rounded-xl px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);background:var(--c-surface);color:var(--c-ink)"
        >
          <option value="">— Pilih Trip —</option>
          {#each packages as pkg}
            <option value={pkg.id}>{pkg.name}</option>
          {/each}
        </select>
      </div>
    {/if}

    <div class="flex gap-2 pt-2">
      <Button variant="ghost" full onclick={() => (showBillForm = false)}>Batal</Button>
      <Button variant="primary" full disabled={savingBill} onclick={saveBill}>
        {savingBill ? 'Menyimpan…' : 'Simpan'}
      </Button>
    </div>
  </div>
</SlideDrawer>

<style>
  .suluk-row:hover {
    background: var(--c-primary-tint);
  }
  .tabular {
    font-variant-numeric: tabular-nums;
  }
</style>
