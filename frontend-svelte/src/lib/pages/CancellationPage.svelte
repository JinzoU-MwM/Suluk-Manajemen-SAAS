<script>
  import { onMount } from 'svelte';
  import {
    RotateCcw, ShieldAlert, CheckCheck, Wallet, Ban, Plus,
    Pencil, Trash2, XCircle, CheckCircle,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import Avatar from '../components/Avatar.svelte';
  import Card from '../components/ui/Card.svelte';
  import Button from '../components/ui/Button.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { ApiService } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  let refunds = $state([]);
  let total = $state(0);
  let policies = $state([]);
  let invoices = $state([]);
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

  const FILTER_TABS = [
    { value: 'all', label: 'Semua' },
    { value: 'pending', label: 'Pending' },
    { value: 'approved', label: 'Disetujui' },
    { value: 'processed', label: 'Diproses' },
    { value: 'completed', label: 'Selesai' },
    { value: 'rejected', label: 'Ditolak' },
  ];

  async function loadData() {
    loading = true;
    try {
      const [refundData, policyData, invoiceData] = await Promise.all([
        ApiService.listRefunds({ status: statusFilter === 'all' ? '' : statusFilter }),
        ApiService.listPolicies(),
        ApiService.listInvoices({ pageSize: 100 }).catch(() => null),
      ]);
      refunds = refundData.refunds || [];
      total = refundData.total || 0;
      policies = policyData.policies || [];
      const invList = invoiceData?.invoices || invoiceData?.data || (Array.isArray(invoiceData) ? invoiceData : []) || [];
      invoices = invList.filter((i) => i.status !== 'batal' && (i.amount_paid ?? 0) > 0);
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      loading = false;
    }
  }

  function setFilter(v) {
    statusFilter = v;
    loadData();
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

  function openNewRefund() {
    refundForm = { invoice_id: '', amount: 0, refund_pct: 100, reason: '', notes: '' };
    showNewRefundDrawer = true;
  }

  async function selectInvoiceForRefund(id) {
    refundForm.invoice_id = id;
    const inv = invoices.find((i) => i.id === id);
    refundForm.amount = inv?.amount_paid ?? 0;
    const policy = await ApiService.getApplicablePolicy(id);
    refundForm.refund_pct = policy?.refund_pct ?? 100;
  }

  async function saveNewRefund() {
    if (!refundForm.invoice_id || !refundForm.amount) {
      showToast('Pilih invoice dan pastikan ada nominal yang dibayar', 'error');
      return;
    }
    savingRefund = true;
    try {
      await ApiService.initiateRefund(refundForm.invoice_id, {
        amount: Number(refundForm.amount),
        refund_pct: refundForm.refund_pct === '' ? 100 : Number(refundForm.refund_pct),
        reason: refundForm.reason,
        notes: refundForm.notes,
      });
      showToast('Pengajuan refund dibuat');
      showNewRefundDrawer = false;
      await loadData();
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      savingRefund = false;
    }
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

<div class="p-4 lg:p-8">
  <PageHeader
    kicker="Pembatalan"
    title="Pembatalan & Refund"
    subtitle="Kelola pengajuan pembatalan jamaah dan proses pengembalian dana sesuai kebijakan."
  >
    {#snippet actions()}
      <Button variant="ghost" icon={Plus} onclick={openNewPolicy}>Kebijakan</Button>
      <Button variant="primary" icon={Plus} onclick={openNewRefund}>Ajukan Refund</Button>
    {/snippet}
  </PageHeader>

  {#if loading}
    <div class="flex items-center justify-center py-16"><div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600 border-t-transparent"></div></div>
  {:else}
    <!-- Summary -->
    {@const s = summary()}
    <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard icon={Wallet} label="Total Refund" value={formatIDR(s.totalAmount)} accent="var(--c-danger)" />
      <StatCard icon={XCircle} label="Menunggu Approval" value={String(s.pending)} accent="var(--c-warning)" />
      <StatCard icon={CheckCheck} label="Disetujui" value={String(s.approved)} accent="var(--c-info)" />
      <StatCard icon={CheckCircle} label="Selesai" value={String(s.completed)} accent="var(--c-success)" />
    </div>

    <!-- Policies bar -->
    {#if policies.length}
      <div class="mb-4 flex flex-wrap gap-2">
        {#each policies as p}
          <span
            class="inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-semibold"
            style="background:var(--c-bg-2);color:var(--c-muted)"
          >
            {p.name}: {p.refund_pct}% @ {p.days_before}h
            <button type="button" onclick={() => editPolicy(p)} class="ml-0.5" style="color:var(--c-faint)" aria-label="Edit kebijakan"><Pencil class="h-3 w-3" /></button>
          </span>
        {/each}
      </div>
    {/if}

    <!-- Filter -->
    <div class="mb-4 flex flex-wrap items-center gap-3">
      <FilterTabs tabs={FILTER_TABS} value={statusFilter} onChange={setFilter} />
    </div>

    <!-- Refund Table -->
    <Card pad={false} style="padding:8px 4px">
      {#if refunds.length === 0}
        <div class="flex flex-col items-center justify-center py-16" style="color:var(--c-faint)">
          <ShieldAlert class="mb-2 h-10 w-10" />
          <p class="text-sm">Belum ada data refund</p>
        </div>
      {:else}
        <div class="overflow-x-auto">
          <table class="w-full" style="border-collapse:collapse;font-size:13.5px">
            <thead>
              <tr>
                <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Jamaah</th>
                <th style="text-align:right;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Jumlah</th>
                <th style="text-align:right;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">%</th>
                <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Alasan</th>
                <th style="text-align:center;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Status</th>
                <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Tanggal</th>
                <th style="text-align:right;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {#each refunds as r}
                <tr
                  class="cursor-pointer transition-colors"
                  style="transition:background .12s"
                  onclick={() => openRefundDetail(r)}
                  onmouseenter={(e) => e.currentTarget.style.background = 'var(--c-primary-tint)'}
                  onmouseleave={(e) => e.currentTarget.style.background = ''}
                >
                  <td style="padding:14px 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle">
                    <div class="flex items-center gap-3">
                      <Avatar name={r.jamaah_name || r.invoice_id || '?'} size={36} />
                      <div class="min-w-0">
                        <p class="truncate text-sm font-bold" style="color:var(--c-ink)">{r.jamaah_name || 'Jamaah'}</p>
                        <p class="truncate font-mono text-xs" style="color:var(--c-faint)">{r.invoice_id?.substring(0, 8)}...</p>
                      </div>
                    </div>
                  </td>
                  <td class="tabular" style="padding:14px 16px;border-bottom:1px solid var(--c-line-soft);text-align:right;font-weight:800;color:var(--c-ink);font-variant-numeric:tabular-nums;white-space:nowrap">{formatIDR(r.amount)}</td>
                  <td class="tabular" style="padding:14px 16px;border-bottom:1px solid var(--c-line-soft);text-align:right;color:var(--c-ink-soft);font-variant-numeric:tabular-nums;white-space:nowrap">{r.refund_pct}%</td>
                  <td style="padding:14px 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;max-width:160px;color:var(--c-muted);font-size:12.5px" class="truncate">{r.reason || '-'}</td>
                  <td style="padding:14px 16px;border-bottom:1px solid var(--c-line-soft);text-align:center;white-space:nowrap"><StatusBadge status={r.status} size="xs" /></td>
                  <td style="padding:14px 16px;border-bottom:1px solid var(--c-line-soft);color:var(--c-muted);font-size:12.5px;white-space:nowrap">{formatDate(r.created_at)}</td>
                  <td style="padding:14px 16px;border-bottom:1px solid var(--c-line-soft);text-align:right;white-space:nowrap">
                    <div role="presentation" class="flex items-center justify-end gap-1.5" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
                      {#if r.status === 'pending'}
                        <button type="button" onclick={() => refundAction(r.id, 'approve')} class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-bold" style="background:var(--c-success-soft);color:var(--c-success)"><CheckCheck class="h-3 w-3" />Setuju</button>
                        <button type="button" onclick={() => refundAction(r.id, 'reject')} class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-bold" style="background:var(--c-danger-soft);color:var(--c-danger)"><Ban class="h-3 w-3" />Tolak</button>
                      {/if}
                      {#if r.status === 'approved'}
                        <button type="button" onclick={() => refundAction(r.id, 'process')} class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-bold" style="background:var(--c-info-soft);color:var(--c-info)"><RotateCcw class="h-3 w-3" />Proses</button>
                      {/if}
                      {#if r.status === 'processed'}
                        <button type="button" onclick={() => refundAction(r.id, 'complete')} class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-bold" style="background:var(--c-success-soft);color:var(--c-success)"><CheckCheck class="h-3 w-3" />Selesai</button>
                      {/if}
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </Card>
  {/if}
</div>

<!-- Policy Drawer -->
<SlideDrawer open={showPolicyDrawer} onClose={() => showPolicyDrawer = false} title={editingPolicy ? 'Edit Kebijakan Refund' : 'Tambah Kebijakan Refund'} width="480px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1">
      <label for="pol-name" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Nama Kebijakan</label>
      <input id="pol-name" type="text" bind:value={policyForm.name} placeholder="Contoh: 30 hari sebelum keberangkatan" class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)" />
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="pol-days" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Hari Sebelum</label>
        <input id="pol-days" type="number" bind:value={policyForm.days_before} class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)" />
      </div>
      <div class="flex flex-col gap-1">
        <label for="pol-pct" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Persentase (%)</label>
        <input id="pol-pct" type="number" bind:value={policyForm.refund_pct} min="0" max="100" step="0.01" class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)" />
      </div>
    </div>
    <div class="flex flex-col gap-1">
      <label for="pol-desc" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Deskripsi</label>
      <textarea id="pol-desc" bind:value={policyForm.description} rows="3" placeholder="Optional..." class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)"></textarea>
    </div>
    <div class="flex gap-2 pt-2">
      <Button variant="ghost" full onclick={() => showPolicyDrawer = false}>Batal</Button>
      <Button variant="primary" full disabled={savingPolicy} onclick={savePolicy}>{savingPolicy ? 'Menyimpan...' : editingPolicy ? 'Update' : 'Simpan'}</Button>
    </div>
    {#if editingPolicy}
      <Button variant="danger" full icon={Trash2} onclick={() => deletePolicy(editingPolicy.id)}>Hapus Kebijakan</Button>
    {/if}
  </div>
</SlideDrawer>

<!-- Refund Detail Drawer -->
<SlideDrawer open={showRefundDrawer} onClose={() => showRefundDrawer = false} title="Detail Refund" width="480px">
  {#if selectedRefund}
    <div class="flex flex-col gap-4 p-4">
      <div class="space-y-3 rounded-xl p-4" style="border:1px solid var(--c-line-soft)">
        <div class="flex items-center justify-between">
          <span class="text-xs" style="color:var(--c-muted)">Status</span>
          <StatusBadge status={selectedRefund.status} size="sm" />
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs" style="color:var(--c-muted)">Jumlah</span>
          <span class="text-sm font-bold" style="color:var(--c-ink);font-variant-numeric:tabular-nums">{formatIDR(selectedRefund.amount)}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs" style="color:var(--c-muted)">Persentase</span>
          <span class="text-sm font-semibold" style="color:var(--c-ink-soft)">{selectedRefund.refund_pct}%</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs" style="color:var(--c-muted)">Invoice</span>
          <span class="font-mono text-xs" style="color:var(--c-muted)">{selectedRefund.invoice_id}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs" style="color:var(--c-muted)">Alasan</span>
          <span class="text-sm" style="color:var(--c-ink-soft)">{selectedRefund.reason || '-'}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs" style="color:var(--c-muted)">Catatan</span>
          <span class="text-sm" style="color:var(--c-ink-soft)">{selectedRefund.notes || '-'}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs" style="color:var(--c-muted)">Dibuat</span>
          <span class="text-sm" style="color:var(--c-muted)">{formatDate(selectedRefund.created_at)}</span>
        </div>
        {#if selectedRefund.approved_at}
          <div class="flex items-center justify-between">
            <span class="text-xs" style="color:var(--c-muted)">Disetujui</span>
            <span class="text-sm" style="color:var(--c-muted)">{formatDate(selectedRefund.approved_at)}</span>
          </div>
        {/if}
        {#if selectedRefund.processed_at}
          <div class="flex items-center justify-between">
            <span class="text-xs" style="color:var(--c-muted)">Diproses</span>
            <span class="text-sm" style="color:var(--c-muted)">{formatDate(selectedRefund.processed_at)}</span>
          </div>
        {/if}
      </div>
      <div class="flex flex-wrap gap-2">
        {#if selectedRefund.status === 'pending'}
          <Button variant="primary" full onclick={() => { refundAction(selectedRefund.id, 'approve'); showRefundDrawer = false; }}>Setujui</Button>
          <Button variant="danger" full onclick={() => { refundAction(selectedRefund.id, 'reject'); showRefundDrawer = false; }}>Tolak</Button>
        {/if}
        {#if selectedRefund.status === 'approved'}
          <Button variant="primary" full onclick={() => { refundAction(selectedRefund.id, 'process'); showRefundDrawer = false; }}>Proses Refund</Button>
        {/if}
        {#if selectedRefund.status === 'processed'}
          <Button variant="primary" full onclick={() => { refundAction(selectedRefund.id, 'complete'); showRefundDrawer = false; }}>Tandai Selesai</Button>
        {/if}
      </div>
    </div>
  {/if}
</SlideDrawer>

<!-- New Refund Drawer -->
<SlideDrawer open={showNewRefundDrawer} onClose={() => showNewRefundDrawer = false} title="Ajukan Refund Baru" width="480px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1">
      <label for="ref-invoice" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Invoice</label>
      <select id="ref-invoice" value={refundForm.invoice_id} onchange={(e) => selectInvoiceForRefund(e.target.value)} class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)">
        <option value="">Pilih invoice...</option>
        {#each invoices as inv}
          <option value={inv.id}>{(inv.jamaah_name || inv.invoice_number || inv.id)} · {formatIDR(inv.amount_paid)}</option>
        {/each}
      </select>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="ref-amount" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Jumlah Refund (Rp)</label>
        <input id="ref-amount" type="number" bind:value={refundForm.amount} class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)" />
      </div>
      <div class="flex flex-col gap-1">
        <label for="ref-pct" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Persentase (%)</label>
        <input id="ref-pct" type="number" bind:value={refundForm.refund_pct} min="0" max="100" step="0.01" class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)" />
      </div>
    </div>
    <div class="flex flex-col gap-1">
      <label for="ref-reason" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Alasan</label>
      <textarea id="ref-reason" bind:value={refundForm.reason} rows="2" placeholder="Contoh: Jamaah gagal berangkat" class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)"></textarea>
    </div>
    <div class="flex flex-col gap-1">
      <label for="ref-notes" class="text-xs font-semibold" style="color:var(--c-ink-soft)">Catatan (opsional)</label>
      <textarea id="ref-notes" bind:value={refundForm.notes} rows="2" class="rounded-xl px-3 py-2 text-sm outline-none focus:border-primary-400" style="border:1px solid var(--c-line)"></textarea>
    </div>
    <div class="flex gap-2 pt-2">
      <Button variant="ghost" full onclick={() => showNewRefundDrawer = false}>Batal</Button>
      <Button variant="primary" full disabled={savingRefund} onclick={saveNewRefund}>{savingRefund ? 'Menyimpan...' : 'Ajukan'}</Button>
    </div>
  </div>
</SlideDrawer>
