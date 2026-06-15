<script>
  import { onMount } from "svelte";
  import { ApiService } from "$lib/services/api.js";
  import StatusBadge from "$lib/components/StatusBadge.svelte";
  import Seo from "$lib/components/Seo.svelte";

  let leads = $state([]);
  let loading = $state(true);
  let error = $state("");

  onMount(async () => {
    try {
      const data = await ApiService.myLeads();
      leads = data.leads || [];
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  });

  function tempColor(t) {
    return t === "hot" ? "var(--c-danger)" : t === "warm" ? "var(--c-warning)" : "var(--c-muted)";
  }
  function tempBg(t) {
    return t === "hot" ? "var(--c-danger-soft)" : t === "warm" ? "var(--c-warning-soft, #fef3c7)" : "var(--c-bg-2)";
  }
</script>

<Seo title="Lead Saya - Suluk" path="/agency/leads" robots="noindex,nofollow" />

<h1 class="mb-1 text-xl font-extrabold" style="color:var(--c-ink)">Lead Saya</h1>
<p class="mb-5 text-sm" style="color:var(--c-muted)">Jamaah yang Anda dan jaringan Anda referensikan ({leads.length}).</p>

{#if loading}
  <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
{:else if error}
  <div class="rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{error}</div>
{:else if leads.length === 0}
  <div class="py-16 text-center" style="color:var(--c-faint)">Belum ada lead</div>
{:else}
  <div class="overflow-hidden rounded-2xl" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
    <table class="w-full" style="font-size:13.5px">
      <thead style="background:var(--c-bg-2)">
        <tr class="text-left text-[11px] font-bold uppercase tracking-wider" style="color:var(--c-faint)">
          <th class="px-4 py-3">Nama</th>
          <th class="hidden px-4 py-3 sm:table-cell">No. HP</th>
          <th class="px-4 py-3">Skor</th>
          <th class="px-4 py-3">Status</th>
        </tr>
      </thead>
      <tbody>
        {#each leads as l}
          <tr style="border-top:1px solid var(--c-line-soft)">
            <td class="px-4 py-3 font-semibold" style="color:var(--c-ink)">{l.nama}</td>
            <td class="hidden px-4 py-3 sm:table-cell" style="color:var(--c-muted)">{l.no_hp || "—"}</td>
            <td class="px-4 py-3">
              {#if l.lead_score != null}
                <span class="rounded-full px-2 py-0.5 text-[11px] font-bold" style="background:{tempBg(l.lead_temp)};color:{tempColor(l.lead_temp)}">{l.lead_score}</span>
              {:else}
                <span style="color:var(--c-faint)">—</span>
              {/if}
            </td>
            <td class="px-4 py-3">{#if l.pipeline_status}<StatusBadge status={l.pipeline_status} size="xs" />{:else}<span style="color:var(--c-faint)">—</span>{/if}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}
