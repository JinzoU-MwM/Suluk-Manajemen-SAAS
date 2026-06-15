<script>
  import { onMount } from "svelte";
  import { ApiService } from "$lib/services/api.js";
  import { formatRupiah } from "$lib/utils/formatting.js";
  import Seo from "$lib/components/Seo.svelte";

  let nodes = $state([]);
  let loading = $state(true);
  let error = $state("");

  onMount(async () => {
    try {
      const data = await ApiService.myDownline();
      nodes = data.downline || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });
</script>

<Seo title="Jaringan - Suluk" path="/agency/jaringan" robots="noindex,nofollow" />

<h1 class="mb-1 text-xl font-extrabold" style="color:var(--c-ink)">Jaringan Saya</h1>
<p class="mb-5 text-sm" style="color:var(--c-muted)">{nodes.length} agen di bawah Anda.</p>

{#if loading}
  <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
{:else if error}
  <div class="rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{error}</div>
{:else if nodes.length === 0}
  <div class="py-16 text-center" style="color:var(--c-faint)">Belum ada agen di jaringan Anda</div>
{:else}
  <div class="space-y-1.5">
    {#each nodes as n}
      <div class="flex items-center justify-between rounded-xl px-3 py-2.5" style="background:var(--c-surface);border:1px solid var(--c-line);margin-left:{(n.depth - 1) * 18}px">
        <div class="flex items-center gap-2">
          <span style="color:var(--c-faint)">↳</span>
          <span class="text-sm font-semibold" style="color:var(--c-ink)">{n.name}</span>
          {#if !n.is_active}<span class="rounded-full px-1.5 py-0.5 text-[10px]" style="background:var(--c-bg-2);color:var(--c-faint)">nonaktif</span>{/if}
        </div>
        <span class="text-xs font-semibold" style="color:var(--c-muted)">{n.total_jamaah} jamaah · {formatRupiah(n.total_commissions)}</span>
      </div>
    {/each}
  </div>
{/if}
