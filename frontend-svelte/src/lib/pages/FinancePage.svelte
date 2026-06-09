<script>
  import { onMount } from 'svelte';
  import { TrendingUp, TrendingDown, DollarSign, AlertCircle, Download } from 'lucide-svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { formatRupiah as formatIDR } from '../utils/formatting.js';
  import { ApiService } from '../services/api';
  import StatCard from '../components/StatCard.svelte';

  let { onNavigate, user = null } = $props();

  let activeTab = $state('pl');
  let isLoading = $state(true);
  let dashboard = $state(null);
  let selectedTripId = $state(null);
  let plDetail = $state(null);
  let plLoading = $state(false);

  const TABS = [
    { id: 'pl',       label: 'P&L per Paket' },
    { id: 'aging',    label: 'Piutang per Paket' },
    { id: 'cashflow', label: 'Tren Pendapatan' },
  ];

  let stats = $derived(buildStats(dashboard));
  let packages = $derived(dashboard?.active_packages ?? []);
  let revenueChart = $derived(dashboard?.revenue_chart ?? []);
  let maxRevenue = $derived(Math.max(...revenueChart.map((d) => d.total ?? 0), 1));
  let outstandingPackages = $derived(packages.filter((p) => (p.remaining ?? 0) > 0));

  function buildStats(d) {
    if (!d) return null;
    const s = d.summary ?? {};
    const revenue = s.total_revenue ?? 0;
    const gross = s.gross_profit_month ?? 0;
    return {
      pemasukan: revenue,
      // Operational expenses for the period = revenue − gross profit.
      pengeluaran: Math.max(revenue - gross, 0),
      gross_profit: gross,
      piutang: s.total_piutang ?? 0,
      pemasukan_change: s.revenue_growth_pct ?? null,
    };
  }

  onMount(loadData);

  async function loadData() {
    isLoading = true;
    try {
      dashboard = await ApiService.getOwnerDashboard();
      const pkgs = dashboard?.active_packages ?? [];
      if (pkgs.length && !selectedTripId) selectedTripId = pkgs[0].id;
    } catch {
      showToast('Gagal memuat laporan keuangan', 'error');
    } finally {
      isLoading = false;
    }
  }

  // Load the detailed per-package P&L whenever the selection changes.
  $effect(() => {
    const id = selectedTripId;
    if (!id) {
      plDetail = null;
      return;
    }
    let cancelled = false;
    plLoading = true;
    ApiService.getPnL(id)
      .then((d) => {
        if (!cancelled) plDetail = d;
      })
      .catch(() => {
        if (!cancelled) {
          plDetail = null;
          showToast('Gagal memuat P&L paket', 'error');
        }
      })
      .finally(() => {
        if (!cancelled) plLoading = false;
      });
    return () => {
      cancelled = true;
    };
  });

  function pct(a, b) {
    return b > 0 ? Math.round((a / b) * 100) : 0;
  }

  let plTotalExpense = $derived(
    plDetail ? (plDetail.total_op_expenses ?? 0) + (plDetail.total_vendor_costs ?? 0) : 0,
  );
</script>

<div class="flex h-screen flex-col">
  <!-- Header -->
  <div class="flex-shrink-0 border-b border-slate-100 bg-white px-6 py-5">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="font-serif text-xl font-bold text-slate-800">Laporan Keuangan</h1>
        <p class="mt-0.5 text-sm text-slate-500">Pantau P&L, piutang, dan tren pendapatan bisnis Anda</p>
      </div>
      <div class="flex items-center gap-3">
        <button
          type="button"
          onclick={() => onNavigate?.('export')}
          class="flex items-center gap-2 rounded-xl border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-50"
        >
          <Download class="h-4 w-4" />
          Export Excel
        </button>
      </div>
    </div>

    {#if dashboard?.partial}
      <div class="mt-3 flex items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm text-amber-700">
        <AlertCircle class="h-4 w-4 flex-shrink-0" />
        Sebagian data belum lengkap{dashboard.degraded_sources?.length ? ` (sumber: ${dashboard.degraded_sources.join(', ')})` : ''} — angka mungkin belum final.
      </div>
    {/if}

    <!-- Summary cards -->
    {#if isLoading}
      <div class="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {#each [1, 2, 3, 4] as _}
          <div class="h-20 animate-pulse rounded-xl bg-slate-100"></div>
        {/each}
      </div>
    {:else if stats}
      <div class="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
        <StatCard icon={TrendingUp} label="Total Pemasukan" value={formatIDR(stats.pemasukan)} accent="#1B7F5A"
          delta={stats.pemasukan_change != null ? `${Math.abs(Math.round(stats.pemasukan_change))}%` : null} deltaUp={(stats.pemasukan_change ?? 0) >= 0} />
        <StatCard icon={TrendingDown} label="Biaya Operasional" value={formatIDR(stats.pengeluaran)} accent="#b87708" />
        <StatCard icon={DollarSign} label="Laba Kotor" value={formatIDR(stats.gross_profit)} accent="#2563a8" />
        <StatCard icon={AlertCircle} label="Total Piutang" value={formatIDR(stats.piutang)} accent="#c0392b" />
      </div>
    {/if}

    <!-- Tabs -->
    <div class="mt-4 flex gap-1">
      {#each TABS as tab}
        <button
          type="button"
          onclick={() => (activeTab = tab.id)}
          class="rounded-lg px-4 py-1.5 text-xs font-semibold transition-all
            {activeTab === tab.id ? 'bg-primary-600 text-white' : 'text-slate-500 hover:bg-slate-100'}"
        >
          {tab.label}
        </button>
      {/each}
    </div>
  </div>

  <!-- Tab Content -->
  <div class="flex-1 overflow-y-auto bg-slate-50 p-6">

    {#if activeTab === 'pl'}
      {#if packages.length === 0}
        <div class="rounded-2xl bg-white p-6 text-sm text-slate-400 shadow-sm ring-1 ring-slate-200/60">
          Belum ada paket aktif untuk dianalisa.
        </div>
      {:else}
        <div class="mb-4">
          <label for="select-trip" class="mb-1 block text-sm font-medium text-slate-600">Pilih Paket</label>
          <select
            id="select-trip"
            bind:value={selectedTripId}
            class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400"
          >
            {#each packages as p}
              <option value={p.id}>{p.name}</option>
            {/each}
          </select>
        </div>

        {#if plLoading}
          <div class="grid gap-4 lg:grid-cols-2">
            <div class="h-48 animate-pulse rounded-2xl bg-slate-100"></div>
            <div class="h-48 animate-pulse rounded-2xl bg-slate-100"></div>
          </div>
        {:else if plDetail}
          <div class="grid gap-4 lg:grid-cols-2">
            <!-- Pendapatan -->
            <div class="rounded-2xl bg-white p-5 shadow-sm ring-1 ring-slate-200/60">
              <h3 class="mb-4 text-sm font-bold text-slate-700">Pendapatan</h3>
              <div class="space-y-2 text-sm">
                {@render PLRow('Total Tagihan Jamaah', plDetail.total_revenue ?? 0)}
                {@render PLRow('Sudah Terkumpul', plDetail.revenue_collected ?? 0)}
                <div class="border-t border-slate-100 pt-2">
                  {@render PLRow('Outstanding (belum dibayar)', plDetail.revenue_outstanding ?? 0, true, true)}
                </div>
              </div>
            </div>

            <!-- Pengeluaran -->
            <div class="rounded-2xl bg-white p-5 shadow-sm ring-1 ring-slate-200/60">
              <h3 class="mb-4 text-sm font-bold text-slate-700">Pengeluaran</h3>
              <div class="space-y-2 text-sm">
                {@render PLRow('Biaya Operasional', plDetail.total_op_expenses ?? 0, true)}
                {@render PLRow('Biaya Vendor', plDetail.total_vendor_costs ?? 0, true)}
                <div class="border-t border-slate-100 pt-2">
                  {@render PLRow('Total Pengeluaran', plTotalExpense, true, true)}
                </div>
              </div>
            </div>

            <!-- Cost breakdown (projected vs actual) -->
            {#if plDetail.cost_breakdown?.length}
              <div class="lg:col-span-2 rounded-2xl bg-white p-5 shadow-sm ring-1 ring-slate-200/60">
                <h3 class="mb-4 text-sm font-bold text-slate-700">Rincian Biaya per Kategori</h3>
                <table class="w-full text-sm">
                  <thead>
                    <tr class="text-left text-[11.5px] font-semibold uppercase tracking-wide text-slate-400">
                      <th class="py-2">Kategori</th>
                      <th class="py-2 text-right">Proyeksi</th>
                      <th class="py-2 text-right">Aktual</th>
                      <th class="py-2 text-right">Selisih</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each plDetail.cost_breakdown as c}
                      <tr class="border-b border-slate-100">
                        <td class="py-2.5 capitalize text-slate-600">{c.label || c.category}</td>
                        <td class="py-2.5 text-right text-slate-500" style="font-variant-numeric:tabular-nums">{formatIDR(c.projected_amount ?? 0)}</td>
                        <td class="py-2.5 text-right text-slate-800" style="font-variant-numeric:tabular-nums">{formatIDR(c.actual_amount ?? 0)}</td>
                        <td class="py-2.5 text-right {(c.variance_amount ?? 0) > 0 ? 'text-red-600' : 'text-emerald-600'}" style="font-variant-numeric:tabular-nums">{formatIDR(c.variance_amount ?? 0)}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}

            <!-- Result -->
            <div class="lg:col-span-2 rounded-2xl p-5 shadow-sm ring-1 {(plDetail.gross_profit ?? 0) >= 0 ? 'bg-emerald-50 ring-emerald-200' : 'bg-red-50 ring-red-200'}">
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm font-semibold {(plDetail.gross_profit ?? 0) >= 0 ? 'text-emerald-600' : 'text-red-600'}">Laba Kotor</p>
                  <p class="text-3xl font-bold {(plDetail.gross_profit ?? 0) >= 0 ? 'text-emerald-700' : 'text-red-700'}">{formatIDR(plDetail.gross_profit ?? 0)}</p>
                  {#if plDetail.net_profit != null}
                    <p class="mt-1 text-xs text-slate-500">Laba Bersih: {formatIDR(plDetail.net_profit)}</p>
                  {/if}
                </div>
                <div class="text-right">
                  <p class="text-sm text-slate-500">Margin</p>
                  <p class="text-2xl font-bold text-slate-700">{pct(plDetail.gross_profit ?? 0, plDetail.total_revenue ?? 0)}%</p>
                  {#if plDetail.projected?.profit != null}
                    <p class="mt-1 text-xs text-slate-400">Proyeksi: {formatIDR(plDetail.projected.profit)}</p>
                  {/if}
                </div>
              </div>
              {#if plDetail.data_notes?.length}
                <p class="mt-3 text-[11px] text-slate-400">{plDetail.data_notes.join(' · ')}</p>
              {/if}
            </div>
          </div>
        {/if}
      {/if}

    {:else if activeTab === 'aging'}
      <!-- Piutang per Paket -->
      {#if outstandingPackages.length === 0}
        <div class="rounded-2xl bg-white p-6 text-sm text-slate-400 shadow-sm ring-1 ring-slate-200/60">
          Tidak ada piutang berjalan. 🎉
        </div>
      {:else}
        <div class="overflow-x-auto rounded-2xl bg-white shadow-sm ring-1 ring-slate-200/60">
          <table class="w-full">
            <thead class="bg-slate-50">
              <tr class="text-left text-xs font-semibold uppercase tracking-wider text-slate-400">
                <th class="px-5 py-3">Paket</th>
                <th class="hidden px-4 py-3 text-right lg:table-cell">Total Tagihan</th>
                <th class="hidden px-4 py-3 text-right md:table-cell">Terkumpul</th>
                <th class="px-4 py-3 text-right">Sisa Piutang</th>
                <th class="px-4 py-3 text-right">Progress</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50">
              {#each outstandingPackages as p}
                <tr class="hover:bg-slate-50">
                  <td class="px-5 py-3 text-sm font-semibold text-slate-800">{p.name}</td>
                  <td class="hidden px-4 py-3 text-right text-sm text-slate-600 lg:table-cell" style="font-variant-numeric:tabular-nums">{formatIDR(p.revenue ?? 0)}</td>
                  <td class="hidden px-4 py-3 text-right text-sm text-slate-500 md:table-cell" style="font-variant-numeric:tabular-nums">{formatIDR(p.paid ?? 0)}</td>
                  <td class="px-4 py-3 text-right text-sm font-bold text-red-600" style="font-variant-numeric:tabular-nums">{formatIDR(p.remaining ?? 0)}</td>
                  <td class="px-4 py-3 text-right text-sm text-slate-600">{Math.round(p.payment_pct ?? pct(p.paid ?? 0, p.revenue ?? 0))}%</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}

    {:else if activeTab === 'cashflow'}
      <!-- Revenue trend (real monthly revenue) -->
      <div class="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-slate-200/60">
        <h3 class="mb-6 text-sm font-bold text-slate-700">Pendapatan Bulanan</h3>
        {#if revenueChart.length === 0}
          <p class="text-sm text-slate-400">Belum ada data pendapatan.</p>
        {:else}
          <div class="space-y-4">
            {#each revenueChart as d}
              <div>
                <div class="mb-1 flex items-center justify-between text-xs text-slate-500">
                  <span class="font-medium">{d.month} {d.year ?? ''}</span>
                  <span style="font-variant-numeric:tabular-nums">{formatIDR(d.total ?? 0)}</span>
                </div>
                <div class="h-6">
                  <div class="rounded bg-primary-500" style="width: {pct(d.total ?? 0, maxRevenue)}%; min-width: 2px; height: 100%"></div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>


{#snippet PLRow(label, value, isNegative, bold)}
  <div class="flex items-center justify-between">
    <span class="capitalize text-slate-500 {bold ? 'font-semibold text-slate-700' : ''}">{label}</span>
    <span class="{isNegative ? 'text-red-600' : 'text-slate-800'} {bold ? 'text-base font-bold' : 'font-medium'}" style="font-variant-numeric:tabular-nums">
      {isNegative ? '− ' : ''}{formatIDR(Math.abs(value))}
    </span>
  </div>
{/snippet}
