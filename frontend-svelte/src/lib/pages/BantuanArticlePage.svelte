<script>
  // Halaman detail satu panduan, dipakai ulang oleh ketiga area. `guide` sudah
  // di-resolve di +page.js (404 ditangani di sana), jadi komponen ini hanya
  // merender. `related` di-resolve dalam area yang sama agar tidak bocor.
  import { ArrowLeft } from "lucide-svelte";
  import GuideBody from "$lib/components/help/GuideBody.svelte";
  import GuideCard from "$lib/components/help/GuideCard.svelte";
  import { getGuide } from "$lib/content/help/index.js";

  /**
   * @type {{
   *   area: import('$lib/content/help/index.js').HelpArea,
   *   basePath: string,
   *   guide: import('$lib/content/help/index.js').HelpGuide,
   * }}
   */
  let { area, basePath, guide } = $props();

  let related = $derived(
    (guide.related ?? [])
      .map((slug) => getGuide(area, slug))
      .filter((g) => g !== undefined),
  );

  const MONTHS = [
    "Januari", "Februari", "Maret", "April", "Mei", "Juni",
    "Juli", "Agustus", "September", "Oktober", "November", "Desember",
  ];

  function formatTanggal(iso) {
    const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso ?? "");
    if (!match) return "";
    const [, y, m, d] = match;
    const month = MONTHS[Number(m) - 1];
    if (!month) return "";
    return `${Number(d)} ${month} ${y}`;
  }

  let updated = $derived(formatTanggal(guide.updatedAt));
</script>

<article class="artikel">
  <a class="artikel-back" href={basePath}>
    <ArrowLeft size={16} aria-hidden="true" />
    Kembali ke Pusat Bantuan
  </a>

  <p class="artikel-cat">{guide.category}</p>
  <h1 class="artikel-title">{guide.title}</h1>
  {#if updated}
    <p class="artikel-meta">Diperbarui {updated}</p>
  {/if}

  <GuideBody body={guide.body} />

  {#if related.length > 0}
    <section class="artikel-related" aria-labelledby="related-heading">
      <h2 id="related-heading">Panduan terkait</h2>
      <div class="artikel-related-list">
        {#each related as item (item.slug)}
          <GuideCard guide={item} href={`${basePath}/${item.slug}`} />
        {/each}
      </div>
    </section>
  {/if}
</article>

<style>
  .artikel {
    max-width: 720px;
  }
  .artikel-back {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13.5px;
    font-weight: 600;
    color: var(--c-primary);
    text-decoration: none;
    margin-bottom: 18px;
  }
  .artikel-back:hover {
    text-decoration: underline;
  }
  .artikel-cat {
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--c-primary);
    margin: 0 0 6px;
  }
  .artikel-title {
    font-family: var(--font-serif, "Playfair Display", Georgia, serif);
    font-size: 28px;
    font-weight: 800;
    line-height: 1.2;
    color: var(--c-ink);
    margin: 0 0 6px;
  }
  .artikel-meta {
    font-size: 13px;
    color: var(--c-faint);
    margin: 0 0 20px;
  }
  .artikel-related {
    margin-top: 36px;
    padding-top: 24px;
    border-top: 1px solid var(--c-line);
  }
  .artikel-related h2 {
    font-family: var(--font-serif, "Playfair Display", Georgia, serif);
    font-size: 18px;
    font-weight: 700;
    color: var(--c-ink);
    margin: 0 0 14px;
  }
  .artikel-related-list {
    display: grid;
    gap: 12px;
  }
</style>
