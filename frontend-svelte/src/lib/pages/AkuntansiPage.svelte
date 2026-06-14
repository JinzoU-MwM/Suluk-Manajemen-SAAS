<script>
  import { onMount } from "svelte";
  import { Scale, TrendingUp, ListTree, BookOpen } from "lucide-svelte";
  import PageHeader from "$lib/components/PageHeader.svelte";
  import Card from "$lib/components/ui/Card.svelte";
  import FilterTabs from "$lib/components/ui/FilterTabs.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { ApiService } from "$lib/services/api";
  import { showToast } from "$lib/services/toast.svelte.js";
  import { formatRupiah, formatDate } from "$lib/utils/formatting.js";

  const TABS = [
    { value: "neraca", label: "Neraca" },
    { value: "laba-rugi", label: "Laba Rugi" },
    { value: "jurnal", label: "Jurnal" },
    { value: "coa", label: "Bagan Akun" },
  ];

  const MODULE_LABEL = {
    invoice: "Invoice/Pembayaran",
    vendor: "Vendor",
    payroll: "Payroll",
    agent: "Komisi Agen",
    tabungan: "Tabungan",
    manual: "Manual",
    opening: "Saldo Awal",
  };

  let activeTab = $state("neraca");
  let isLoading = $state(true);
  let neraca = $state(null);
  let labaRugi = $state(null);
  let journals = $state([]);
  let coa = $state([]);

  async function loadTab(tab) {
    isLoading = true;
    try {
      if (tab === "neraca") neraca = await ApiService.getNeraca();
      else if (tab === "laba-rugi") labaRugi = await ApiService.getLabaRugi();
      else if (tab === "jurnal") {
        const r = await ApiService.listJournals({ page: 1, limit: 50 });
        journals = r.items || [];
      } else if (tab === "coa") coa = await ApiService.listCOA();
    } catch (e) {
      showToast(e?.message || "Gagal memuat data akuntansi", "error");
    } finally {
      isLoading = false;
    }
  }

  function setTab(v) {
    activeTab = v;
    loadTab(v);
  }

  onMount(() => loadTab(activeTab));

  const typeLabel = (t) =>
    ({ asset: "Aset", liability: "Liabilitas", equity: "Ekuitas", revenue: "Pendapatan", expense: "Beban" })[t] || t;
</script>

<div class="min-h-screen p-6 lg:p-8" style="background:var(--c-bg)">
  <PageHeader
    kicker="Keuangan"
    title="Akuntansi"
    subtitle="Buku besar double-entry — jurnal terbentuk otomatis dari setiap transaksi (invoice, pembayaran, vendor, payroll, komisi)."
  />

  <div class="mb-5">
    <FilterTabs tabs={TABS} value={activeTab} onChange={setTab} />
  </div>

  {#if isLoading}
    <div class="h-48 animate-pulse rounded-2xl" style="background:var(--c-bg-2,#eef2f0)"></div>

  {:else if activeTab === "neraca"}
    {#if !neraca}
      <Card><EmptyState icon={Scale} title="Belum ada data" text="Belum ada jurnal yang tercatat." /></Card>
    {:else}
      <div class="grid gap-5 lg:grid-cols-2">
        <Card>
          <h3 class="mb-3 font-serif text-lg font-bold">Aset</h3>
          {#each neraca.assets || [] as a}
            <div class="flex justify-between border-b py-2 text-sm" style="border-color:var(--c-line)">
              <span class="text-slate-600">{a.code} · {a.name}</span>
              <span class="font-medium tabular-nums">{formatRupiah(a.amount)}</span>
            </div>
          {/each}
          <div class="mt-3 flex justify-between font-bold">
            <span>Total Aset</span><span class="tabular-nums">{formatRupiah(neraca.total_assets)}</span>
          </div>
        </Card>
        <Card>
          <h3 class="mb-3 font-serif text-lg font-bold">Liabilitas & Ekuitas</h3>
          {#each [...(neraca.liabilities || []), ...(neraca.equity || [])] as l}
            <div class="flex justify-between border-b py-2 text-sm" style="border-color:var(--c-line)">
              <span class="text-slate-600">{l.code} · {l.name}</span>
              <span class="font-medium tabular-nums">{formatRupiah(l.amount)}</span>
            </div>
          {/each}
          <div class="mt-3 flex justify-between font-bold">
            <span>Total Liabilitas + Ekuitas</span>
            <span class="tabular-nums">{formatRupiah(neraca.total_liabilities + neraca.total_equity)}</span>
          </div>
        </Card>
      </div>
      <div class="mt-4 flex items-center gap-2 text-sm">
        <span class="rounded-full px-3 py-1 font-semibold"
          style="background:{neraca.balanced ? 'var(--c-primary-tint,#e7f5ee)' : '#fde8e8'};color:{neraca.balanced ? 'var(--c-primary,#0d7334)' : '#c0392b'}">
          {neraca.balanced ? "✓ Neraca seimbang" : "⚠ Neraca tidak seimbang"}
        </span>
        <span class="text-slate-500">per {neraca.as_of}</span>
      </div>
    {/if}

  {:else if activeTab === "laba-rugi"}
    {#if !labaRugi}
      <Card><EmptyState icon={TrendingUp} title="Belum ada data" text="Belum ada transaksi pada periode ini." /></Card>
    {:else}
      <div class="grid gap-5 lg:grid-cols-2">
        <Card>
          <h3 class="mb-3 font-serif text-lg font-bold">Pendapatan</h3>
          {#each labaRugi.revenue || [] as r}
            <div class="flex justify-between border-b py-2 text-sm" style="border-color:var(--c-line)">
              <span class="text-slate-600">{r.name}</span><span class="tabular-nums">{formatRupiah(r.amount)}</span>
            </div>
          {/each}
          <div class="mt-3 flex justify-between font-bold"><span>Total Pendapatan</span><span class="tabular-nums">{formatRupiah(labaRugi.total_revenue)}</span></div>
        </Card>
        <Card>
          <h3 class="mb-3 font-serif text-lg font-bold">Beban</h3>
          {#each labaRugi.expenses || [] as e}
            <div class="flex justify-between border-b py-2 text-sm" style="border-color:var(--c-line)">
              <span class="text-slate-600">{e.name}</span><span class="tabular-nums">{formatRupiah(e.amount)}</span>
            </div>
          {/each}
          <div class="mt-3 flex justify-between font-bold"><span>Total Beban</span><span class="tabular-nums">{formatRupiah(labaRugi.total_expense)}</span></div>
        </Card>
      </div>
      <Card class="mt-4">
        <div class="flex justify-between text-lg font-bold">
          <span>Laba (Rugi) Bersih</span>
          <span class="tabular-nums" style="color:{labaRugi.net_income >= 0 ? 'var(--c-success,#0d7334)' : 'var(--c-danger,#c0392b)'}">{formatRupiah(labaRugi.net_income)}</span>
        </div>
        <p class="mt-1 text-xs text-slate-500">Periode {labaRugi.from} s/d {labaRugi.to}</p>
      </Card>
    {/if}

  {:else if activeTab === "jurnal"}
    {#if journals.length === 0}
      <Card><EmptyState icon={ListTree} title="Belum ada jurnal" text="Jurnal akan muncul otomatis saat ada transaksi." /></Card>
    {:else}
      <Card pad={false} class="overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-slate-500" style="border-bottom:1px solid var(--c-line)">
                <th class="px-4 py-3 font-semibold">No. Jurnal</th>
                <th class="px-4 py-3 font-semibold">Tanggal</th>
                <th class="px-4 py-3 font-semibold">Sumber</th>
                <th class="px-4 py-3 font-semibold">Keterangan</th>
              </tr>
            </thead>
            <tbody>
              {#each journals as j}
                <tr style="border-top:1px solid var(--c-line)">
                  <td class="px-4 py-3 font-mono text-xs">{j.journal_no}</td>
                  <td class="px-4 py-3">{formatDate(j.journal_date)}</td>
                  <td class="px-4 py-3">{MODULE_LABEL[j.source_module] || j.source_module}</td>
                  <td class="px-4 py-3 text-slate-600">{j.description}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </Card>
    {/if}

  {:else if activeTab === "coa"}
    {#if coa.length === 0}
      <Card><EmptyState icon={BookOpen} title="Belum ada akun" text="Bagan akun akan terisi otomatis." /></Card>
    {:else}
      <Card pad={false} class="overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-slate-500" style="border-bottom:1px solid var(--c-line)">
                <th class="px-4 py-3 font-semibold">Kode</th>
                <th class="px-4 py-3 font-semibold">Nama Akun</th>
                <th class="px-4 py-3 font-semibold">Tipe</th>
                <th class="px-4 py-3 font-semibold">Saldo Normal</th>
              </tr>
            </thead>
            <tbody>
              {#each coa as a}
                <tr style="border-top:1px solid var(--c-line)">
                  <td class="px-4 py-3 font-mono text-xs">{a.code}</td>
                  <td class="px-4 py-3">{a.name}</td>
                  <td class="px-4 py-3">{typeLabel(a.type)}</td>
                  <td class="px-4 py-3 capitalize">{a.normal_balance === "debit" ? "Debit" : "Kredit"}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </Card>
    {/if}
  {/if}
</div>
