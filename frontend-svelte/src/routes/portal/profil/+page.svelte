<script>
  import { onMount } from "svelte";
  import { ApiService } from "$lib/services/api.js";
  import Seo from "$lib/components/Seo.svelte";

  let p = $state(null);
  let loading = $state(true);
  let error = $state("");

  onMount(async () => {
    try { p = await ApiService.portalMe(); }
    catch (e) { error = e.message; } finally { loading = false; }
  });

  function row(label, value) { return { label, value: value || "—" }; }
  let fields = $derived(
    p ? [
      row("Nama", p.nama),
      row("NIK", p.no_identitas),
      row("No. Paspor", p.no_paspor),
      row("Jenis Kelamin", p.gender === "L" ? "Laki-laki" : p.gender === "P" ? "Perempuan" : ""),
      row("No. HP", p.no_hp),
      row("Email", p.email),
      row("Alamat", p.alamat),
    ] : [],
  );
</script>

<Seo title="Profil - Portal Jemaah" path="/portal/profil" robots="noindex,nofollow" />

<h1 class="mb-5 text-xl font-extrabold" style="color:var(--c-ink)">Profil Saya</h1>

{#if loading}
  <div class="py-16 text-center" style="color:var(--c-faint)">Memuat…</div>
{:else if error}
  <div class="rounded-2xl p-4 text-sm" style="border:1px solid var(--c-danger);background:var(--c-danger-soft);color:var(--c-danger)">{error}</div>
{:else if p}
  <div class="max-w-xl rounded-2xl p-5" style="background:var(--c-surface);border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
    <div class="space-y-3">
      {#each fields as f}
        <div class="flex items-start justify-between gap-4 border-b pb-2.5 text-sm" style="border-color:var(--c-line-soft)">
          <span style="color:var(--c-faint)">{f.label}</span>
          <span class="text-right font-medium" style="color:var(--c-ink)">{f.value}</span>
        </div>
      {/each}
      <p class="pt-1 text-xs" style="color:var(--c-faint)">Untuk perubahan data, hubungi travel Anda.</p>
    </div>
  </div>
{/if}
