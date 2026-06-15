<script>
  import { onMount } from "svelte";
  import { Wallet, CheckCircle, Users, Network } from "lucide-svelte";
  import { ApiService } from "$lib/services/api.js";
  import { formatRupiah } from "$lib/utils/formatting.js";
  import Seo from "$lib/components/Seo.svelte";

  let dash = $state(null);
  let loading = $state(true);
  let error = $state("");

  onMount(async () => {
    try {
      dash = await ApiService.myDashboard();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  let agent = $derived(dash?.agent || null);
  let tiles = $derived([
    { label: "Total Komisi", value: formatRupiah(agent?.total_commissions || 0), icon: Wallet, accent: "var(--c-primary)" },
    { label: "Sudah Dibayar", value: formatRupiah(agent?.total_paid || 0), icon: CheckCircle, accent: "var(--c-success)" },
    { label: "Outstanding", value: formatRupiah(agent?.total_outstanding || 0), icon: Wallet, accent: "var(--c-warning)" },
    { label: "Jaringan", value: `${dash?.downline_count ?? 0} agen`, icon: Network, accent: "var(--c-info)" },
  ]);
</script>

<Seo title="Dashboard Agen - Suluk" path="/agency" robots="noindex,nofollow" />

<h1 class="mb-1 text-xl font-extrabold" style="color:var(--c-ink)">Halo, {agent?.name || "Agen"} 👋</h1>
<p class="mb-5 text-sm" style="color:var(--c-muted)">Ringkasan komisi dan jaringan Anda.</p>

{#if loading}
  <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
{:else if error}
  <div class="rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{error}</div>
{:else}
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
    {#each tiles as t}
      <div class="rounded-2xl p-4" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
        <div class="mb-2 flex h-9 w-9 items-center justify-center rounded-xl" style="background:{t.accent}1a;color:{t.accent}">
          <t.icon class="h-4.5 w-4.5" />
        </div>
        <p class="text-lg font-extrabold tabular" style="font-variant-numeric:tabular-nums;color:var(--c-ink)">{t.value}</p>
        <p class="text-xs" style="color:var(--c-faint)">{t.label}</p>
      </div>
    {/each}
  </div>

  <div class="mt-5 grid gap-4 sm:grid-cols-2">
    <div class="rounded-2xl p-4" style="background:var(--c-surface);border:1px solid var(--c-line)">
      <div class="mb-1 flex items-center gap-2"><Users class="h-4 w-4" style="color:var(--c-primary)" /><h2 class="text-sm font-bold" style="color:var(--c-ink)">Jaringan langsung</h2></div>
      <p class="text-2xl font-extrabold" style="color:var(--c-ink)">{dash?.direct_count ?? 0}</p>
      <p class="text-xs" style="color:var(--c-faint)">agen di bawah Anda langsung (total {dash?.downline_count ?? 0} termasuk sub-jaringan)</p>
    </div>
    <div class="rounded-2xl p-4" style="background:var(--c-surface);border:1px solid var(--c-line)">
      <div class="mb-1 flex items-center gap-2"><Wallet class="h-4 w-4" style="color:var(--c-warning)" /><h2 class="text-sm font-bold" style="color:var(--c-ink)">Komisi tertunda</h2></div>
      <p class="text-2xl font-extrabold" style="color:var(--c-warning)">{formatRupiah(agent?.total_outstanding || 0)}</p>
      <p class="text-xs" style="color:var(--c-faint)">akan dibayarkan sesuai jadwal kantor</p>
    </div>
  </div>
{/if}
