<script>
  // Halaman daftar Pusat Bantuan, dipakai ulang oleh ketiga area (app/portal/
  // agency). `area` menentukan konten — helper hanya membaca array area itu,
  // sehingga panduan antar-area tidak pernah bercampur.
  import { Search } from "lucide-svelte";
  import PageHeader from "$lib/components/PageHeader.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import HelpSearch from "$lib/components/help/HelpSearch.svelte";
  import GuideCard from "$lib/components/help/GuideCard.svelte";
  import { getCategories, searchGuides } from "$lib/content/help/index.js";

  /** @type {{ area: import('$lib/content/help/index.js').HelpArea, basePath: string }} */
  let { area, basePath } = $props();

  const SUBTITLE = {
    app: "Panduan langkah demi langkah mengelola jamaah, paket, keuangan, dan keberangkatan.",
    portal: "Panduan menyiapkan dokumen, memantau visa, dan melengkapi data keberangkatan Anda.",
    agency: "Panduan mengelola lead, memantau jaringan, dan mencairkan komisi Anda.",
  };

  let query = $state("");
  let searching = $derived(query.trim() !== "");
  let results = $derived(searching ? searchGuides(area, query) : []);
  let categories = $derived(getCategories(area));

  // Tampilkan "Memulai" lebih dulu, sisanya menurut abjad.
  let categoryNames = $derived(
    Object.keys(categories).sort(
      (a, b) =>
        (a === "Memulai" ? 0 : 1) - (b === "Memulai" ? 0 : 1) ||
        a.localeCompare(b, "id"),
    ),
  );

  function hrefFor(slug) {
    return `${basePath}/${slug}`;
  }
</script>

<div class="bantuan">
  <PageHeader kicker="Bantuan" title="Pusat Bantuan" subtitle={SUBTITLE[area] ?? ""} />

  <div class="bantuan-search">
    <HelpSearch bind:value={query} />
  </div>

  {#if searching}
    <p class="bantuan-count" aria-live="polite">
      {results.length} hasil untuk “{query.trim()}”
    </p>
    {#if results.length > 0}
      <div class="bantuan-list">
        {#each results as guide (guide.slug)}
          <GuideCard {guide} href={hrefFor(guide.slug)} showCategory />
        {/each}
      </div>
    {:else}
      <EmptyState
        icon={Search}
        title="Tidak ada hasil"
        text="Coba kata kunci lain — misalnya nama menu atau langkah yang Anda cari."
      />
    {/if}
  {:else}
    {#each categoryNames as category (category)}
      <section class="bantuan-group">
        <h2 class="bantuan-group-title">{category}</h2>
        <div class="bantuan-list">
          {#each categories[category] as guide (guide.slug)}
            <GuideCard {guide} href={hrefFor(guide.slug)} />
          {/each}
        </div>
      </section>
    {/each}
  {/if}
</div>

<style>
  .bantuan {
    max-width: 820px;
  }
  .bantuan-search {
    margin-bottom: 22px;
  }
  .bantuan-count {
    font-size: 13.5px;
    color: var(--c-muted);
    margin: 0 0 14px;
  }
  .bantuan-group {
    margin-bottom: 26px;
  }
  .bantuan-group-title {
    font-family: var(--font-serif, "Playfair Display", Georgia, serif);
    font-size: 17px;
    font-weight: 700;
    color: var(--c-ink);
    margin: 0 0 12px;
  }
  .bantuan-list {
    display: grid;
    gap: 12px;
  }
</style>
