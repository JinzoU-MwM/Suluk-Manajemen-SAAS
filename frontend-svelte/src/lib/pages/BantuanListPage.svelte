<script>
  // Halaman daftar Pusat Bantuan, dipakai ulang oleh ketiga area (app/portal/
  // agency). `area` menentukan konten — helper hanya membaca array area itu,
  // sehingga panduan antar-area tidak pernah bercampur. Gaya mengikuti idiom
  // halaman aplikasi: PageHeader + utilitas Tailwind skala brand.
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

  // /app tidak punya padding konten (pages mengisinya sendiri); /portal & /agency
  // sudah dipadati 24px oleh layout-nya.
  let outerPad = $derived(area === "app" ? "p-4 lg:p-8" : "p-1 lg:p-2");

  let query = $state("");
  let searching = $derived(query.trim() !== "");
  let results = $derived(searching ? searchGuides(area, query) : []);
  let categories = $derived(getCategories(area));

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

<div class="mx-auto flex w-full max-w-3xl flex-col gap-6 {outerPad}">
  <PageHeader kicker="Bantuan" title="Pusat Bantuan" subtitle={SUBTITLE[area] ?? ""} />

  <HelpSearch bind:value={query} />

  {#if searching}
    <p class="-mt-1 text-[13px] text-slate-500" aria-live="polite">
      {results.length} hasil untuk “{query.trim()}”
    </p>
    {#if results.length > 0}
      <div class="flex flex-col gap-2.5">
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
      <section class="flex flex-col gap-3">
        <h2 class="font-serif text-lg font-bold text-slate-800">{category}</h2>
        <div class="flex flex-col gap-2.5">
          {#each categories[category] as guide (guide.slug)}
            <GuideCard {guide} href={hrefFor(guide.slug)} />
          {/each}
        </div>
      </section>
    {/each}
  {/if}
</div>
