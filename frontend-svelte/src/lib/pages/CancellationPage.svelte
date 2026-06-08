<script>
  import { onMount } from 'svelte';
  import {
    RotateCcw, ShieldAlert, CheckCheck, Clock, Ban, Plus,
    Pencil, Trash2, ChevronRight,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { ApiService } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  let refunds = $state([]);
  let total = $state(0);
  let policies = $state([]);
  let loading = $state(true);
  let statusFilter = $state('all');

  let showPolicyDrawer = $state(false);
  let editingPolicy = $state(null);
  let policyForm = $state({ name: '', days_before: 30, refund_pct: 100, description: '' });
  let savingPolicy = $state(false);

  let showRefundDrawer = $state(false);
  let selectedRefund = $state(null);

  let showNewRefundDrawer = $state(false);
  let refundForm = $state({ invoice_id: '', amount: 0, refund_pct: 100, reason: '', notes: '' });
  let savingRefund = $state(false);

  const STATUS_LABELS = {
    pending: 'Pending',
    approved: 'Disetujui',
    processed: 'Diproses',
    completed: 'Selesai',
    rejected: 'Ditolak',
  };

  const STATUS_COLORS = {
    pending: 'amber',
    approved: 'blue',
    processed: 'indigo',
    completed: 'emerald',
    rejected: 'red',
  };

  async function loadData() {
    loading = true;
    try {
      const [refundData, policyData] = await Promise.all([
        ApiService.listRefunds({ status: statusFilter === 'all' ? '' : statusFilter }),
        ApiService.listPolicies(),
      ]);
      refunds = refundData.refunds || [];
      total = refundData.total || 0;
      policies = policyData.policies || [];
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      loading = false;
    }
  }

  onMount(() => { loadData(); });

  function formatIDR(n) { return n ? `Rp ${Number(n).toLocaleString('id-ID')}` : 'Rp 0'; }
  function formatDate(d) { return d ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '-'; }

  function summary() {
    const totalAmount = refunds.reduce((s, r) => s + (r.amount || 0), 0);
    const pending = refunds.filter(r => r.status === 'pending').length;
    const approved = refunds.filter(r => r.status === 'approved').length;
    const completed = refunds.filter(r => r.status === 'completed').length;
    return { totalAmount, pending, approved, completed };
  }

  function openNewPolicy() {
    editingPolicy = null;
    policyForm = { name: '', days_before: 30, refund_pct: 100, description: '' };
    showPolicyDrawer = true;
  }

  function editPolicy(p) {
    editingPolicy = p;
    policyForm = { name: p.name, days_before: p.days_before, refund_pct: p.refund_pct, description: p.description };
    showPolicyDrawer = true;
  }

  async function savePolicy() {
    savingPolicy = true;
    try {
      if (editingPolicy) {
        await ApiService.updatePolicy(editingPolicy.id, policyForm);
        showToast('Kebijakan diperbarui');
      } else {
        await ApiService.createPolicy(policyForm);
        showToast('Kebijakan ditambahkan');
      }
      showPolicyDrawer = false;
      await loadData();
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      savingPolicy = false;
    }
  }

  async function deletePolicy(id) {
    if (!confirm('Hapus kebijakan ini?')) return;
    try {
      await ApiService.deletePolicy(id);
      showToast('Kebijakan dihapus');
      await loadData();
    } catch (e) {
      showToast(e.message, 'error');
    }
  }

  function openRefundDetail(r) {
    selectedRefund = r;
    showRefundDrawer = true;
  }

  async function refundAction(id, action) {
    try {
      if (action === 'approve') await ApiService.approveRefund(id);
      else if (action === 'process') await ApiService.processRefund(id);
      else if (action === 'complete') await ApiService.completeRefund(id);
      else if (action === 'reject') await ApiService.rejectRefund(id);
      showToast(`Refund ${action === 'approve' ? 'disetujui' : action === 'reject' ? 'ditolak' : action === 'process' ? 'diproses' : 'selesai'}`);
      await loadData();
    } catch (e) {
      showToast(e.message, 'error');
    }
  }
</script>

<div class="flex flex-col gap-6 p-4 lg:p-8">
  <header class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
    <div>
      <h1 class="text-xl font-bold text-slate-900">Pembatalan & Refund</h1>
      <p class="text-sm text-slate-500">Kelola pengembalian dana dan kebijakan refund.</p>
    </div>
    <div class="flex gap-2">
      <button type="button" onclick={openNewPolicy} class="flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 hover:bg-slate-50">
        <Plus class="h-4 w-4" /> Kebijakan
      </button>
    </div>
  </header>

  {#if loading}
    <div class="flex items-center justify-center py-16"><div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600 border-t-transparent"></div></div>
  {:else}
    <!-- Summary -->
    {@const s = summary()}
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div class="rounded-2xl border border-slate-100 bg-white p-4 shadow-sm">
        <p class="text-2xl font-bold text-slate-800">{formatIDR(s.totalAmount)}</p>
        <p class="text-xs text-slate-500">Total Refund</p>
      </div>
      <div class="rounded-2xl border border-slate-100 bg-white p-4 shadow-sm">
        <p class="text-2xl font-bold text-amber-600">{s.pending}</p>
        <p class="text-xs text-slate-500">Menunggu Approval</p>
      </div>
      <div class="rounded-2xl border border-slate-100 bg-white p-4 shadow-sm">
        <p class="text-2xl font-bold text-blue-600">{s.approved}</p>
        <p class="text-xs text-slate-500">Disetujui</p>
      </div>
      <div class="rounded-2xl border border-slate-100 bg-white p-4 shadow-sm">
        <p class="text-2xl font-bold text-emerald-600">{s.completed}</p>
        <p class="text-xs text-slate-500">Selesai</p>
      </div>
    </div>

    <!-- Policies bar -->
    <div class="flex flex-wrap gap-2">
      {#each policies as p}
        <span class="inline-flex items-center gap-1 rounded-lg bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600">
          {p.name}: {p.refund_pct}% @ {p.days_before}h
          <button type="button" onclick={() => editPolicy(p)} class="ml-1 text-slate-400 hover:text-slate-600"><Pencil class="h-3 w-3" /></button>
        </span>
      {/each}
    </div>

    <!-- Filter -->
    <div class="flex items-center gap-2">
      <select bind:value={statusFilter} onchange={loadData} class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-medium text-slate-700 outline-none">
        <option value="all">Semua Status</option>
        <option value="pending">Pending</option>
        <option value="approved">Disetujui</option>
        <option value="processed">Diproses</option>
        <option value="completed">Selesai</option>
        <option value="rejected">Ditolak</option>
      </select>
    </div>

    <!-- Refund Table -->
    <div class="overflow-hidden rounded-2xl border border-slate-100 bg-white shadow-sm">
      {#if refunds.length === 0}
        <div class="flex flex-col items-center justify-center py-16 text-slate-400">
          <ShieldAlert class="h-10 w-10 mb-2" />
          <p class="text-sm">Belum ada data refund</p>
        </div>
      {:else}
        <table class="w-full text-sm">
          <thead class="border-b border-slate-100 bg-slate-50">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Invoice</th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Jumlah</th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">%</th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Alasan</th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Status</th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Tanggal</th>
              <th class="px-4 py-3 text-right text-xs font-semibold text-slate-500">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-50">
            {#each refunds as r}
              <tr class="hover:bg-slate-50 cursor-pointer" onclick={() => openRefundDetail(r)}>
                <td class="px-4 py-3 font-mono text-xs text-slate-500">{r.invoice_id?.substring(0, 8)}...</td>
                <td class="px-4 py-3 font-semibold text-slate-800">{formatIDR(r.amount)}</td>
                <td class="px-4 py-3 text-slate-600">{r.refund_pct}%</td>
                <td class="px-4 py-3 max-w-[150px] truncate text-xs text-slate-500">{r.reason || '-'}</td>
                <td class="px-4 py-3"><StatusBadge status={r.status} size="xs" /></td>
                <td class="px-4 py-3 text-xs text-slate-500">{formatDate(r.created_at)}</td>
                <td class="px-4 py-3 text-right">
                  <div role="presentation" class="flex items-center justify-end gap-1" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
                    {#if r.status === 'pending'}
                      <button type="button" onclick={() => refundAction(r.id, 'approve')} class="rounded-lg bg-emerald-100 px-2.5 py-1 text-xs font-semibold text-emerald-700 hover:bg-emerald-200"><CheckCheck class="h-3 w-3 inline mr-1" />Setuju</button>
                      <button type="button" onclick={() => refundAction(r.id, 'reject')} class="rounded-lg bg-red-100 px-2.5 py-1 text-xs font-semibold text-red-700 hover:bg-red-200"><Ban class="h-3 w-3 inline mr-1" />Tolak</button>
                    {/if}
                    {#if r.status === 'approved'}
                      <button type="button" onclick={() => refundAction(r.id, 'process')} class="rounded-lg bg-indigo-100 px-2.5 py-1 text-xs font-semibold text-indigo-700 hover:bg-indigo-200"><RotateCcw class="h-3 w-3 inline mr-1" />Proses</button>
                    {/if}
                    {#if r.status === 'processed'}
                      <button type="button" onclick={() => refundAction(r.id, 'complete')} class="rounded-lg bg-emerald-100 px-2.5 py-1 text-xs font-semibold text-emerald-700 hover:bg-emerald-200"><CheckCheck class="h-3 w-3 inline mr-1" />Selesai</button>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>
  {/if}
</div>

<!-- Policy Drawer -->
<SlideDrawer open={showPolicyDrawer} onClose={() => showPolicyDrawer = false} title={editingPolicy ? 'Edit Kebijakan Refund' : 'Tambah Kebijakan Refund'} width="480px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1">
      <label for="pol-name" class="text-xs font-medium text-slate-700">Nama Kebijakan</label>
      <input id="pol-name" type="text" bind:value={policyForm.name} placeholder="Contoh: 30 hari sebelum keberangkatan" class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" />
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="pol-days" class="text-xs font-medium text-slate-700">Hari Sebelum</label>
        <input id="pol-days" type="number" bind:value={policyForm.days_before} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" />
      </div>
      <div class="flex flex-col gap-1">
        <label for="pol-pct" class="text-xs font-medium text-slate-700">Persentase (%)</label>
        <input id="pol-pct" type="number" bind:value={policyForm.refund_pct} min="0" max="100" step="0.01" class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" />
      </div>
    </div>
    <div class="flex flex-col gap-1">
      <label for="pol-desc" class="text-xs font-medium text-slate-700">Deskripsi</label>
      <textarea id="pol-desc" bind:value={policyForm.description} rows="3" placeholder="Optional..." class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400"></textarea>
    </div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showPolicyDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600 hover:bg-slate-50">Batal</button>
      <button type="button" onclick={savePolicy} disabled={savingPolicy} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingPolicy ? 'Menyimpan...' : editingPolicy ? 'Update' : 'Simpan'}</button>
    </div>
    {#if editingPolicy}
      <button type="button" onclick={() => deletePolicy(editingPolicy.id)} class="rounded-xl border border-red-200 py-2.5 text-sm font-semibold text-red-600 hover:bg-red-50"><Trash2 class="h-4 w-4 inline mr-1" />Hapus Kebijakan</button>
    {/if}
  </div>
</SlideDrawer>

<!-- Refund Detail Drawer -->
<SlideDrawer open={showRefundDrawer} onClose={() => showRefundDrawer = false} title="Detail Refund" width="480px">
  {#if selectedRefund}
    <div class="flex flex-col gap-4 p-4">
      <div class="rounded-xl border border-slate-100 p-4 space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-500">Status</span>
          <StatusBadge status={selectedRefund.status} size="sm" />
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-500">Jumlah</span>
          <span class="text-sm font-bold text-slate-800">{formatIDR(selectedRefund.amount)}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-500">Persentase</span>
          <span class="text-sm font-semibold text-slate-700">{selectedRefund.refund_pct}%</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-500">Invoice</span>
          <span class="text-xs font-mono text-slate-500">{selectedRefund.invoice_id}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-500">Alasan</span>
          <span class="text-sm text-slate-700">{selectedRefund.reason || '-'}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-500">Catatan</span>
          <span class="text-sm text-slate-700">{selectedRefund.notes || '-'}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-500">Dibuat</span>
          <span class="text-sm text-slate-600">{formatDate(selectedRefund.created_at)}</span>
        </div>
        {#if selectedRefund.approved_at}
          <div class="flex items-center justify-between">
            <span class="text-xs text-slate-500">Disetujui</span>
            <span class="text-sm text-slate-600">{formatDate(selectedRefund.approved_at)}</span>
          </div>
        {/if}
        {#if selectedRefund.processed_at}
          <div class="flex items-center justify-between">
            <span class="text-xs text-slate-500">Diproses</span>
            <span class="text-sm text-slate-600">{formatDate(selectedRefund.processed_at)}</span>
          </div>
        {/if}
      </div>
      <div class="flex flex-wrap gap-2">
        {#if selectedRefund.status === 'pending'}
          <button type="button" onclick={() => { refundAction(selectedRefund.id, 'approve'); showRefundDrawer = false; }} class="flex-1 rounded-xl bg-emerald-600 py-2.5 text-sm font-semibold text-white hover:bg-emerald-700">Setujui</button>
          <button type="button" onclick={() => { refundAction(selectedRefund.id, 'reject'); showRefundDrawer = false; }} class="flex-1 rounded-xl bg-red-600 py-2.5 text-sm font-semibold text-white hover:bg-red-700">Tolak</button>
        {/if}
        {#if selectedRefund.status === 'approved'}
          <button type="button" onclick={() => { refundAction(selectedRefund.id, 'process'); showRefundDrawer = false; }} class="w-full rounded-xl bg-indigo-600 py-2.5 text-sm font-semibold text-white hover:bg-indigo-700">Proses Refund</button>
        {/if}
        {#if selectedRefund.status === 'processed'}
          <button type="button" onclick={() => { refundAction(selectedRefund.id, 'complete'); showRefundDrawer = false; }} class="w-full rounded-xl bg-emerald-600 py-2.5 text-sm font-semibold text-white hover:bg-emerald-700">Tandai Selesai</button>
        {/if}
      </div>
    </div>
  {/if}
</SlideDrawer>
