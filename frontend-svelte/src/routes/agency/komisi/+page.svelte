<script>
  import { onMount } from "svelte";
  import { ApiService } from "$lib/services/api.js";
  import { formatRupiah } from "$lib/utils/formatting.js";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import Seo from "$lib/components/Seo.svelte";

  let commissions = $state([]);
  let loading = $state(true);
  let error = $state("");

  onMount(async () => {
    try {
      const data = await ApiService.myCommissions();
      commissions = data.commissions || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  function tierLabel(t) { return t > 1 ? `Tier ${t}` : "Langsung"; }
  let totalPending = $derived(commissions.filter((c) => c.status === "pending").reduce((s, c) => s + (c.commission_amount || 0), 0));
</script>

<Seo title="Komisi Saya - Suluk" path="/agency/komisi" robots="noindex,nofollow" />

<h1 class="mb-1 text-xl font-extrabold" style="color:var(--c-ink)">Komisi Saya</h1>
<p class="mb-5 text-sm" style="color:var(--c-muted)">Tertunda: <span class="font-bold" style="color:var(--c-warning)">{formatRupiah(totalPending)}</span></p>

{#if loading}
  <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
{:else if error}
  <div class="rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{error}</div>
{:else if commissions.length === 0}
  <div class="py-16 text-center" style="color:var(--c-faint)">Belum ada komisi</div>
{:else}
  <div class="overflow-hidden rounded-2xl" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
    <table class="w-full" style="font-size:13.5px">
      <thead style="background:var(--c-bg-2)">
        <tr class="text-left text-[11px] font-bold uppercase tracking-wider" style="color:var(--c-faint)">
          <th class="px-4 py-3">Jamaah</th>
          <th class="hidden px-4 py-3 sm:table-cell">Paket</th>
          <th class="px-4 py-3">Tier</th>
          <th class="px-4 py-3 text-right">Jumlah</th>
          <th class="px-4 py-3">Status</th>
        </tr>
      </thead>
      <tbody>
        {#each commissions as c}
          <tr style="border-top:1px solid var(--c-line-soft)">
            <td class="px-4 py-3 font-semibold" style="color:var(--c-ink)">{c.jamaah_name || "-"}</td>
            <td class="hidden px-4 py-3 sm:table-cell" style="color:var(--c-muted)">{c.package_name || "-"}</td>
            <td class="px-4 py-3">
              <span class="rounded-full px-2 py-0.5 text-[11px] font-bold" style="background:{c.tier_level > 1 ? 'var(--c-info-soft, #e0f2fe)' : 'var(--c-primary-tint)'};color:{c.tier_level > 1 ? 'var(--c-info)' : 'var(--c-primary)'}">{tierLabel(c.tier_level)}</span>
            </td>
            <td class="px-4 py-3 text-right font-bold tabular" style="font-variant-numeric:tabular-nums;color:var(--c-ink)">{formatRupiah(c.commission_amount)}</td>
            <td class="px-4 py-3"><StatusBadge status={c.status} size="xs" /></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}
