<script>
  import { onMount } from 'svelte';
  import { FileSpreadsheet, FileText, Download, Receipt, TrendingUp, AlertTriangle } from 'lucide-svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { authHeaders } from '../services/apiCore.js';
  import { ApiService } from '../services/api.js';

  let { onNavigate, user } = $props();

  let slips = $state([]);
  let invoices = $state([]);
  let loading = $state(true);

  async function loadData() {
    try {
      const [iData, sData] = await Promise.all([
        ApiService.listInvoices?.({ status: 'lunas' }).catch(() => ({ invoices: [] })),
        ApiService.listSalarySlips?.('').catch(() => ({ slips: [] })),
      ]);
      invoices = iData?.invoices?.slice(0, 10) || [];
      slips = sData?.slips?.slice(0, 10) || [];
    } catch (e) { /* ignore */ }
    loading = false;
  }

  onMount(() => { loadData(); });

  async function download(url, filename) {
    try {
      const res = await fetch(url, { headers: authHeaders() });
      if (!res.ok) {
        const err = await res.text();
        showToast(err || 'Gagal mengunduh', 'error');
        return;
      }
      const blob = await res.blob();
      const blobUrl = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = blobUrl;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(blobUrl);
      showToast('Mengunduh ' + filename);
    } catch (e) {
      showToast('Gagal mengunduh: ' + e.message, 'error');
    }
  }
</script>

<div class="flex flex-col gap-6 p-4 lg:p-8">
  <header>
    <h1 class="font-serif text-xl font-bold text-slate-900">Export Laporan</h1>
    <p class="text-sm text-slate-500">Unduh laporan dalam format Excel dan PDF.</p>
  </header>

  <div class="grid gap-4 md:grid-cols-2">
    <!-- Finance Excel -->
    <div class="rounded-2xl border border-slate-100 bg-white p-5 shadow-sm">
      <div class="flex items-center gap-3 mb-4">
        <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-50"><TrendingUp class="h-5 w-5 text-blue-600" /></div>
        <div><h3 class="font-semibold text-slate-800">Laporan Keuangan</h3><p class="text-xs text-slate-500">Format Excel (.xlsx)</p></div>
      </div>
      <div class="flex flex-col gap-2">
        <button type="button" onclick={() => download(`${ApiService.getPnLExportUrl?.() || '/api/finance/export/pnl'}`, 'pnl_report.xlsx')} class="flex items-center justify-between rounded-xl border border-slate-200 px-4 py-2.5 text-sm hover:bg-slate-50"><span class="flex items-center gap-2"><FileSpreadsheet class="h-4 w-4 text-emerald-600" /> P&L Report</span><Download class="h-4 w-4 text-slate-400" /></button>
        <button type="button" onclick={() => download(`${ApiService.getExpensesExportUrl?.() || '/api/finance/export/expenses'}`, 'expenses.xlsx')} class="flex items-center justify-between rounded-xl border border-slate-200 px-4 py-2.5 text-sm hover:bg-slate-50"><span class="flex items-center gap-2"><FileSpreadsheet class="h-4 w-4 text-emerald-600" /> Biaya Operasional</span><Download class="h-4 w-4 text-slate-400" /></button>
      </div>
    </div>

    <!-- Invoice PDF -->
    <div class="rounded-2xl border border-slate-100 bg-white p-5 shadow-sm">
      <div class="flex items-center gap-3 mb-4">
        <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-50"><Receipt class="h-5 w-5 text-amber-600" /></div>
        <div><h3 class="font-semibold text-slate-800">Kwitansi PDF</h3><p class="text-xs text-slate-500">Format PDF</p></div>
      </div>
      <div class="flex flex-col gap-2 max-h-64 overflow-y-auto">
        {#if invoices.length === 0}
          <p class="text-sm text-slate-400 py-4 text-center">Belum ada invoice lunas</p>
        {:else}
          {#each invoices as inv}
            <button type="button" onclick={() => download(`/api/invoices/${inv.id}/pdf`, `invoice_${inv.invoice_number}.pdf`)} class="flex items-center justify-between rounded-xl border border-slate-200 px-4 py-2.5 text-sm hover:bg-slate-50">
              <span class="flex items-center gap-2"><FileText class="h-4 w-4 text-red-500" />{inv.invoice_number}</span><Download class="h-4 w-4 text-slate-400" />
            </button>
          {/each}
        {/if}
      </div>
    </div>

    <!-- Payroll PDF -->
    <div class="rounded-2xl border border-slate-100 bg-white p-5 shadow-sm">
      <div class="flex items-center gap-3 mb-4">
        <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-50"><FileText class="h-5 w-5 text-emerald-600" /></div>
        <div><h3 class="font-semibold text-slate-800">Slip Gaji PDF</h3><p class="text-xs text-slate-500">Format PDF</p></div>
      </div>
      <div class="flex flex-col gap-2 max-h-64 overflow-y-auto">
        {#if slips.length === 0}
          <p class="text-sm text-slate-400 py-4 text-center">Belum ada slip gaji</p>
        {:else}
          {#each slips as s}
            <button type="button" onclick={() => download(`/api/payroll/slips/${s.id}/pdf`, 'slip_gaji.pdf')} class="flex items-center justify-between rounded-xl border border-slate-200 px-4 py-2.5 text-sm hover:bg-slate-50">
              <span class="flex items-center gap-2"><FileText class="h-4 w-4 text-emerald-500" />{s.employee_name} — {s.period}</span><Download class="h-4 w-4 text-slate-400" />
            </button>
          {/each}
        {/if}
      </div>
    </div>
  </div>
</div>
