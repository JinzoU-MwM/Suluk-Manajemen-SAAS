<script>
  import { onMount } from "svelte";
  import { ApiService } from "$lib/services/api.js";
  import { formatDate } from "$lib/utils/formatting.js";
  import Seo from "$lib/components/Seo.svelte";

  let visa = $state(null);
  let loading = $state(true);
  let error = $state("");

  const STATUS = {
    draft: ["Draf", "var(--c-muted)"],
    submitted: ["Diajukan", "var(--c-info)"],
    approved: ["Disetujui", "var(--c-success)"],
    rejected: ["Ditolak", "var(--c-danger)"],
    expired: ["Kedaluwarsa", "var(--c-warning)"],
  };

  onMount(async () => {
    try { visa = await ApiService.portalVisa(); }
    catch (e) { error = e.message; } finally { loading = false; }
  });
</script>

<Seo title="Visa - Portal Jemaah" path="/portal/visa" robots="noindex,nofollow" />

<h1 class="mb-5 text-xl font-extrabold" style="color:var(--c-ink)">Status Visa</h1>

{#if loading}
  <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
{:else if error}
  <div class="rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{error}</div>
{:else if !visa}
  <div class="py-16 text-center" style="color:var(--c-faint)">Belum ada pengajuan visa.</div>
{:else}
  {@const meta = STATUS[visa.status] || STATUS.draft}
  <div class="max-w-md rounded-2xl p-5" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
    <span class="rounded-full px-3 py-1 text-xs font-bold" style="background:{meta[1]}1a;color:{meta[1]}">{meta[0]}</span>
    <div class="mt-4 space-y-2.5 text-sm">
      <div class="flex justify-between"><span style="color:var(--c-faint)">Provider</span><span style="color:var(--c-ink)">{visa.provider || "—"}</span></div>
      <div class="flex justify-between"><span style="color:var(--c-faint)">No. Visa</span><span style="color:var(--c-ink)">{visa.reference_no || "—"}</span></div>
      {#if visa.expiry_date}<div class="flex justify-between"><span style="color:var(--c-faint)">Berlaku s/d</span><span style="color:var(--c-ink)">{formatDate(visa.expiry_date)}</span></div>{/if}
      {#if visa.status === "rejected" && visa.reject_reason}<div class="flex justify-between"><span style="color:var(--c-faint)">Alasan</span><span style="color:var(--c-danger)">{visa.reject_reason}</span></div>{/if}
    </div>
  </div>
{/if}
