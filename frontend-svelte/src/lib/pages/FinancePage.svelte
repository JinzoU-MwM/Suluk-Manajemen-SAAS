<script>
  import { onMount } from 'svelte';
  import { TrendingUp, TrendingDown, DollarSign, AlertCircle, Download } from 'lucide-svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { formatRupiah as formatIDR } from '../utils/formatting.js';
  import { ApiService } from '../services/api';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import EmptyState from '../components/EmptyState.svelte';
  import Card from '../components/ui/Card.svelte';
  import Button from '../components/ui/Button.svelte';
  import Badge from '../components/ui/Badge.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
  import ProgressBar from '../components/ui/ProgressBar.svelte';

  let { onNavigate, user = null } = $props();

  let activeTab = $state('pl');
  let isLoading = $state(true);
  let dashboard = $state(null);
  let selectedTripId = $state(null);
  let plDetail = $state(null);
  let plLoading = $state(false);

  const TABS = [
    { value: 'pl',       label: 'P&L per Paket' },
    { value: 'aging',    label: 'Piutang per Paket' },
    { value: 'cashflow', label: 'Tren Pendapatan' },
  ];

  // Palette for expense-breakdown bars (mirrors the design's Rincian Biaya colors).
  const COST_COLORS = ['#2563c9', 'var(--c-accent)', '#7a5ae0', 'var(--c-primary)', '#a9842f', '#15564a', '#b87708'];

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
  let plProfit = $derived(plDetail?.gross_profit ?? 0);
  let plPositive = $derived(plProfit >= 0);

  // Expense-breakdown bars from the per-package cost breakdown (actual amounts).
  let costBars = $derived(
    (plDetail?.cost_breakdown ?? []).map((c, i) => ({
      label: c.label || c.category,
      value: c.actual_amount ?? 0,
      projected: c.projected_amount ?? 0,
      variance: c.variance_amount ?? 0,
      color: COST_COLORS[i % COST_COLORS.length],
    })),
  );
  let maxCost = $derived(Math.max(...costBars.map((c) => c.value), 1));

  let chartPoints = $derived(
    revenueChart.map((d) => ({
      label: `${d.month}${d.year ? ' ' + d.year : ''}`,
      shortLabel: d.month,
      value: d.total ?? 0,
    })),
  );
</script>

<div class="finance-page min-h-screen bg-[var(--c-bg)] p-6 lg:p-8">
  <PageHeader
    kicker="Keuangan"
    title="Laporan Keuangan"
    subtitle="Laporan laba rugi, piutang, dan tren pendapatan bisnis Anda."
  >
    {#snippet actions()}
      <Button variant="ghost" icon={Download} onclick={() => onNavigate?.('export')}>
        Unduh Laporan
      </Button>
    {/snippet}
  </PageHeader>

  {#if dashboard?.partial}
    <div
      class="mb-5 flex items-center gap-2 rounded-[var(--radius)] border px-4 py-2.5 text-sm"
      style="border-color:var(--c-warning);background:var(--c-warning-soft);color:var(--c-warning)"
    >
      <AlertCircle class="h-4 w-4 flex-shrink-0" />
      Sebagian data belum lengkap{dashboard.degraded_sources?.length
        ? ` (sumber: ${dashboard.degraded_sources.join(', ')})`
        : ''} — angka mungkin belum final.
    </div>
  {/if}

  <!-- Summary cards -->
  {#if isLoading}
    <div class="mb-6 grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4">
      {#each [1, 2, 3, 4] as _}
        <div class="h-[120px] animate-pulse rounded-2xl bg-[var(--c-bg-2)]"></div>
      {/each}
    </div>
  {:else if stats}
    <div class="mb-6 grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard
        icon={TrendingUp}
        label="Pendapatan"
        value={formatIDR(stats.pemasukan)}
        accent="var(--c-success)"
        delta={stats.pemasukan_change != null ? `${Math.abs(Math.round(stats.pemasukan_change))}%` : null}
        deltaUp={(stats.pemasukan_change ?? 0) >= 0}
      />
      <StatCard
        icon={TrendingDown}
        label="Biaya Operasional"
        value={formatIDR(stats.pengeluaran)}
        accent="var(--c-warning)"
      />
      <StatCard
        icon={DollarSign}
        label="Laba Kotor"
        value={formatIDR(stats.gross_profit)}
        accent="var(--c-primary)"
        sub={`Margin ${pct(stats.gross_profit, stats.pemasukan)}%`}
      />
      <StatCard
        icon={AlertCircle}
        label="Total Piutang"
        value={formatIDR(stats.piutang)}
        accent="var(--c-danger)"
      />
    </div>
  {/if}

  <!-- Tabs -->
  <div class="mb-5">
    <FilterTabs tabs={TABS} value={activeTab} onChange={(v) => (activeTab = v)} />
  </div>

  <!-- Tab Content -->
  {#if activeTab === 'pl'}
    {#if packages.length === 0}
      <Card>
        <EmptyState
          icon={DollarSign}
          title="Belum ada paket aktif"
          text="Belum ada paket aktif untuk dianalisa laba ruginya."
        />
      </Card>
    {:else}
      <div class="mb-5 max-w-md">
        <label for="select-trip" class="mb-1.5 block text-[11.5px] font-bold uppercase tracking-[0.04em] text-[var(--c-faint)]">
          Pilih Paket
        </label>
        <select
          id="select-trip"
          bind:value={selectedTripId}
          class="w-full rounded-[var(--radius)] border bg-[var(--c-surface)] px-3.5 py-2.5 text-sm text-[var(--c-ink)] outline-none transition-colors"
          style="border-color:var(--c-line)"
        >
          {#each packages as p}
            <option value={p.id}>{p.name}</option>
          {/each}
        </select>
      </div>

      {#if plLoading}
        <div class="grid gap-5 lg:grid-cols-2">
          <div class="h-48 animate-pulse rounded-2xl bg-[var(--c-bg-2)]"></div>
          <div class="h-48 animate-pulse rounded-2xl bg-[var(--c-bg-2)]"></div>
        </div>
      {:else if plDetail}
        <div class="grid items-start gap-5 lg:grid-cols-2">
          <!-- Pendapatan -->
          <Card>
            <h3 class="mb-4 text-[15.5px] font-extrabold text-[var(--c-ink)]">Pendapatan</h3>
            <div class="space-y-2.5 text-sm">
              {@render PLRow('Total Tagihan Jamaah', plDetail.total_revenue ?? 0)}
              {@render PLRow('Sudah Terkumpul', plDetail.revenue_collected ?? 0)}
              <div class="border-t pt-2.5" style="border-color:var(--c-line)">
                {@render PLRow('Outstanding (belum dibayar)', plDetail.revenue_outstanding ?? 0, true, true)}
              </div>
            </div>
          </Card>

          <!-- Pengeluaran -->
          <Card>
            <h3 class="mb-4 text-[15.5px] font-extrabold text-[var(--c-ink)]">Pengeluaran</h3>
            <div class="space-y-2.5 text-sm">
              {@render PLRow('Biaya Operasional', plDetail.total_op_expenses ?? 0, true)}
              {@render PLRow('Biaya Vendor', plDetail.total_vendor_costs ?? 0, true)}
              <div class="border-t pt-2.5" style="border-color:var(--c-line)">
                {@render PLRow('Total Pengeluaran', plTotalExpense, true, true)}
              </div>
            </div>
          </Card>

          <!-- Cost breakdown (expense ProgressBars + projected vs actual table) -->
          {#if costBars.length}
            <Card class="lg:col-span-2">
              <h3 class="mb-5 text-[15.5px] font-extrabold text-[var(--c-ink)]">Rincian Biaya per Kategori</h3>

              <!-- Expense breakdown bars -->
              <div class="mb-6 flex flex-col gap-3.5">
                {#each costBars as c}
                  <div>
                    <div class="mb-1.5 flex items-center justify-between text-[13px]">
                      <span class="font-semibold capitalize text-[var(--c-ink-soft)]">{c.label}</span>
                      <span class="font-bold text-[var(--c-ink)]" style="font-variant-numeric:tabular-nums">{formatIDR(c.value)}</span>
                    </div>
                    <ProgressBar value={c.value} max={maxCost} color={c.color} />
                  </div>
                {/each}
              </div>

              <!-- Projected vs actual table -->
              <div class="overflow-x-auto">
                <table class="w-full text-[13.5px]">
                  <thead>
                    <tr>
                      <th class="finance-th text-left">Kategori</th>
                      <th class="finance-th text-right">Proyeksi</th>
                      <th class="finance-th text-right">Aktual</th>
                      <th class="finance-th text-right">Selisih</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each costBars as c}
                      <tr>
                        <td class="finance-td capitalize" style="color:var(--c-ink-soft)">{c.label}</td>
                        <td class="finance-td text-right" style="color:var(--c-muted);font-variant-numeric:tabular-nums">{formatIDR(c.projected)}</td>
                        <td class="finance-td text-right" style="color:var(--c-ink);font-variant-numeric:tabular-nums">{formatIDR(c.value)}</td>
                        <td
                          class="finance-td text-right font-semibold"
                          style="color:var({c.variance > 0 ? '--c-danger' : '--c-success'});font-variant-numeric:tabular-nums"
                        >{formatIDR(c.variance)}</td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </Card>
          {/if}

          <!-- Result -->
          <Card
            class="lg:col-span-2"
            style="background:var({plPositive ? '--c-success-soft' : '--c-danger-soft'});border-color:var({plPositive ? '--c-success' : '--c-danger'})"
          >
            <div class="flex flex-wrap items-center justify-between gap-4">
              <div>
                <div class="mb-1 flex items-center gap-2">
                  <p class="text-sm font-bold" style="color:var({plPositive ? '--c-success' : '--c-danger'})">Laba Kotor</p>
                  <Badge tone={plPositive ? 'success' : 'danger'} label={plPositive ? 'Untung' : 'Rugi'} />
                </div>
                <p class="font-serif text-3xl font-extrabold" style="color:var({plPositive ? '--c-primary-deep' : '--c-danger'});font-variant-numeric:tabular-nums">{formatIDR(plProfit)}</p>
                {#if plDetail.net_profit != null}
                  <p class="mt-1 text-xs text-[var(--c-muted)]">Laba Bersih: {formatIDR(plDetail.net_profit)}</p>
                {/if}
              </div>
              <div class="text-right">
                <p class="text-sm text-[var(--c-muted)]">Margin</p>
                <p class="text-2xl font-extrabold text-[var(--c-ink)]" style="font-variant-numeric:tabular-nums">{pct(plProfit, plDetail.total_revenue ?? 0)}%</p>
                {#if plDetail.projected?.profit != null}
                  <p class="mt-1 text-xs text-[var(--c-faint)]">Proyeksi: {formatIDR(plDetail.projected.profit)}</p>
                {/if}
              </div>
            </div>
            {#if plDetail.data_notes?.length}
              <p class="mt-3 text-[11px] text-[var(--c-faint)]">{plDetail.data_notes.join(' · ')}</p>
            {/if}
          </Card>
        </div>
      {/if}
    {/if}

  {:else if activeTab === 'aging'}
    <!-- Piutang per Paket -->
    {#if outstandingPackages.length === 0}
      <Card>
        <EmptyState
          icon={TrendingUp}
          title="Tidak ada piutang berjalan"
          text="Semua tagihan paket aktif sudah lunas. Kerja bagus!"
        />
      </Card>
    {:else}
      <Card pad={false} class="overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-[13.5px]">
            <thead>
              <tr>
                <th class="finance-th finance-th--pad text-left">Paket</th>
                <th class="finance-th finance-th--pad hidden text-right lg:table-cell">Total Tagihan</th>
                <th class="finance-th finance-th--pad hidden text-right md:table-cell">Terkumpul</th>
                <th class="finance-th finance-th--pad text-right">Sisa Piutang</th>
                <th class="finance-th finance-th--pad text-right">Progress</th>
              </tr>
            </thead>
            <tbody>
              {#each outstandingPackages as p}
                {@const progress = Math.round(p.payment_pct ?? pct(p.paid ?? 0, p.revenue ?? 0))}
                <tr class="finance-row">
                  <td class="finance-td finance-td--pad font-semibold" style="color:var(--c-ink)">{p.name}</td>
                  <td class="finance-td finance-td--pad hidden text-right lg:table-cell" style="color:var(--c-ink-soft);font-variant-numeric:tabular-nums">{formatIDR(p.revenue ?? 0)}</td>
                  <td class="finance-td finance-td--pad hidden text-right md:table-cell" style="color:var(--c-muted);font-variant-numeric:tabular-nums">{formatIDR(p.paid ?? 0)}</td>
                  <td class="finance-td finance-td--pad text-right font-bold" style="color:var(--c-danger);font-variant-numeric:tabular-nums">{formatIDR(p.remaining ?? 0)}</td>
                  <td class="finance-td finance-td--pad">
                    <div class="flex items-center justify-end gap-2.5">
                      <div class="hidden w-24 sm:block">
                        <ProgressBar value={progress} max={100} />
                      </div>
                      <span class="w-9 text-right text-[13px] font-semibold" style="color:var(--c-ink-soft);font-variant-numeric:tabular-nums">{progress}%</span>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </Card>
    {/if}

  {:else if activeTab === 'cashflow'}
    <!-- Revenue trend (real monthly revenue) -->
    <Card>
      <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
        <h3 class="text-[15.5px] font-extrabold text-[var(--c-ink)]">Arus Pendapatan</h3>
        <span class="flex items-center gap-1.5 text-[12.5px] text-[var(--c-muted)]">
          <span class="h-2.5 w-2.5 rounded-[3px]" style="background:var(--c-primary)"></span>
          Pendapatan Bulanan
        </span>
      </div>

      {#if chartPoints.length === 0}
        <EmptyState icon={TrendingUp} title="Belum ada data pendapatan" text="Data tren pendapatan akan muncul di sini setelah ada transaksi." />
      {:else if chartPoints.length === 1}
        <!-- Single data point: a bar is clearer than a line. -->
        <div>
          <div class="mb-1.5 flex items-center justify-between text-xs text-[var(--c-muted)]">
            <span class="font-medium">{chartPoints[0].label}</span>
            <span style="font-variant-numeric:tabular-nums">{formatIDR(chartPoints[0].value)}</span>
          </div>
          <ProgressBar value={chartPoints[0].value} max={maxRevenue} height={10} />
        </div>
      {:else}
        {@render areaChart(chartPoints)}
      {/if}
    </Card>
  {/if}
</div>

{#snippet areaChart(points)}
  {@const n = points.length}
  {@const max = Math.max(...points.map((p) => p.value), 1) * 1.15}
  {@const W = 640}
  {@const H = 230}
  {@const padX = 8}
  {@const padT = 16}
  {@const padB = 28}
  {@const xs = (i) => (n <= 1 ? W / 2 : padX + (i / (n - 1)) * (W - 2 * padX))}
  {@const ys = (v) => padT + (1 - v / max) * (H - padT - padB)}
  {@const line = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${xs(i).toFixed(1)},${ys(p.value).toFixed(1)}`).join(' ')}
  {@const area = `${line} L${xs(n - 1).toFixed(1)},${(H - padB).toFixed(1)} L${xs(0).toFixed(1)},${(H - padB).toFixed(1)} Z`}
  <svg viewBox={`0 0 ${W} ${H}`} class="w-full" style="height:230px;overflow:visible" preserveAspectRatio="none" role="img" aria-label="Tren pendapatan">
    <defs>
      <linearGradient id="sulukFinArea" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="var(--c-primary)" stop-opacity="0.22" />
        <stop offset="100%" stop-color="var(--c-primary)" stop-opacity="0" />
      </linearGradient>
    </defs>
    {#each [0.25, 0.5, 0.75, 1] as g}
      <line x1={padX} x2={W - padX} y1={padT + g * (H - padT - padB)} y2={padT + g * (H - padT - padB)} stroke="var(--c-line-soft)" stroke-width="1" />
    {/each}
    <path d={area} fill="url(#sulukFinArea)" />
    <path d={line} fill="none" stroke="var(--c-primary)" stroke-width="2.5" vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" />
    {#each points as p, i}
      <circle cx={xs(i)} cy={ys(p.value)} r="4" fill="var(--c-surface)" stroke="var(--c-primary)" stroke-width="2.5" vector-effect="non-scaling-stroke" />
    {/each}
  </svg>
  <div class="mt-1.5 flex justify-between px-1 text-[11px] font-medium text-[var(--c-faint)]">
    {#each points as p}<span class="truncate">{p.shortLabel}</span>{/each}
  </div>
{/snippet}

{#snippet PLRow(label, value, isNegative, bold)}
  <div class="flex items-center justify-between">
    <span class="capitalize {bold ? 'font-semibold text-[var(--c-ink)]' : 'text-[var(--c-muted)]'}">{label}</span>
    <span
      class="{bold ? 'text-base font-bold' : 'font-medium'}"
      style="color:var({isNegative ? '--c-danger' : '--c-ink'});font-variant-numeric:tabular-nums"
    >
      {isNegative ? '− ' : ''}{formatIDR(Math.abs(value))}
    </span>
  </div>
{/snippet}

<style>
  .finance-th {
    padding: 0 16px 12px;
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--c-faint);
    white-space: nowrap;
    border-bottom: 1px solid var(--c-line);
  }
  .finance-th--pad {
    padding-top: 16px;
  }
  .finance-td {
    padding: 13px 16px;
    border-bottom: 1px solid var(--c-line-soft);
    vertical-align: middle;
    white-space: nowrap;
  }
  .finance-td--pad {
    padding: 14px 16px;
  }
  .finance-row {
    transition: background 0.12s;
  }
  .finance-row:hover {
    background: var(--c-primary-tint);
  }
  .finance-page :global(select:focus) {
    border-color: var(--c-primary) !important;
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }
</style>
