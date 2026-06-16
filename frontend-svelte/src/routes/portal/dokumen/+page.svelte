<script>
  import { onMount } from "svelte";
  import { ApiService } from "$lib/services/api.js";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import Seo from "$lib/components/Seo.svelte";

  let docs = $state([]);
  let loading = $state(true);
  let error = $state("");

  onMount(async () => {
    try { const d = await ApiService.portalDocuments(); docs = d.documents || []; }
    catch (e) { error = e.message; } finally { loading = false; }
  });
</script>

<Seo title="Dokumen - Portal Jemaah" path="/portal/dokumen" robots="noindex,nofollow" />

<h1 class="mb-5 text-xl font-extrabold" style="color:var(--c-ink)">Dokumen Saya</h1>

{#if loading}
  <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
{:else if error}
  <div class="rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{error}</div>
{:else if docs.length === 0}
  <div class="py-16 text-center" style="color:var(--c-faint)">Belum ada dokumen.</div>
{:else}
  <div class="space-y-2">
    {#each docs as d}
      <div class="flex items-center justify-between rounded-xl p-3.5" style="background:var(--c-surface);border:1px solid var(--c-line)">
        <span class="text-sm font-semibold capitalize" style="color:var(--c-ink)">{(d.doc_type || "").replace(/_/g, " ")}</span>
        <StatusBadge status={d.status} size="xs" />
      </div>
    {/each}
  </div>
{/if}
