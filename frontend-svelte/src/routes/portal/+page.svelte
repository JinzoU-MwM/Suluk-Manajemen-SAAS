<script>
  import { onMount } from "svelte";
  import { Wallet, CheckCircle, Package, Plane } from "lucide-svelte";
  import { ApiService } from "$lib/services/api.js";
  import { formatRupiah } from "$lib/utils/formatting.js";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import Seo from "$lib/components/Seo.svelte";

  let profile = $state(null);
  let regs = $state([]);
  let payments = $state(null);
  let loading = $state(true);
  let error = $state("");

  onMount(async () => {
    try {
      const [p, r, pay] = await Promise.all([
        ApiService.portalMe(),
        ApiService.portalRegistrations().then((d) => d.registrations || []).catch(() => []),
        ApiService.portalPayments().catch(() => null),
      ]);
      profile = p; regs = r; payments = pay;
    } catch (e) { error = e.message; } finally { loading = false; }
  });

  let tiles = $derived([
    { label: "Total Tagihan", value: formatRupiah(payments?.total_amount || 0), icon: Package, accent: "var(--c-primary)" },
    { label: "Sudah Dibayar", value: formatRupiah(payments?.total_paid || 0), icon: CheckCircle, accent: "var(--c-success)" },
    { label: "Sisa Tagihan", value: formatRupiah(payments?.total_remaining || 0), icon: Wallet, accent: "var(--c-warning)" },
  ]);
</script>

<Seo title="Beranda - Portal Jemaah" path="/portal" robots="noindex,nofollow" />

<h1 class="mb-1 text-xl font-extrabold" style="color:var(--c-ink)">Assalamualaikum, {profile?.nama || "Jemaah"} 👋</h1>
<p class="mb-5 text-sm" style="color:var(--c-muted)">Pantau status pendaftaran, pembayaran, dan dokumen Anda.</p>

{#if loading}
  <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
{:else if error}
  <div class="rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{error}</div>
{:else}
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
    {#each tiles as t}
      <div class="rounded-2xl p-4" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
        <div class="mb-2 flex h-9 w-9 items-center justify-center rounded-xl" style="background:{t.accent}1a;color:{t.accent}"><t.icon class="h-4.5 w-4.5" /></div>
        <p class="text-lg font-extrabold tabular" style="font-variant-numeric:tabular-nums;color:var(--c-ink)">{t.value}</p>
        <p class="text-xs" style="color:var(--c-faint)">{t.label}</p>
      </div>
    {/each}
  </div>

  <h2 class="mb-2 mt-6 text-sm font-bold uppercase tracking-wider" style="color:var(--c-faint)">Pendaftaran Paket</h2>
  {#if regs.length === 0}
    <div class="rounded-2xl p-6 text-center text-sm" style="background:var(--c-surface);border:1px solid var(--c-line);color:var(--c-faint)">Belum ada pendaftaran paket.</div>
  {:else}
    <div class="space-y-2">
      {#each regs as r}
        <div class="flex items-center justify-between rounded-2xl p-4" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
          <div class="flex items-center gap-3">
            <Plane class="h-5 w-5" style="color:var(--c-primary)" />
            <div>
              <p class="text-sm font-bold" style="color:var(--c-ink)">{r.room_type || "Paket"}</p>
              {#if r.berangkat_date}<p class="text-xs" style="color:var(--c-faint)">Berangkat dijadwalkan</p>{/if}
            </div>
          </div>
          <StatusBadge status={r.pipeline_status} size="xs" />
        </div>
      {/each}
    </div>
  {/if}
{/if}
