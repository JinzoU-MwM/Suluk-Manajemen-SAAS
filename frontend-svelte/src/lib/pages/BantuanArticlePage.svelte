<script>
  // Halaman detail satu panduan, dipakai ulang oleh ketiga area. `guide` sudah
  // di-resolve di +page.js (404 ditangani di sana), jadi komponen ini hanya
  // merender. `related` di-resolve dalam area yang sama agar tidak bocor.
  // Gaya mengikuti idiom aplikasi (Tailwind, skala brand, kartu konten).
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

  let outerPad = $derived(area === "app" ? "p-4 lg:p-8" : "p-1 lg:p-2");

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

<div class="mx-auto flex w-full max-w-3xl flex-col {outerPad}">
  <a
    href={basePath}
    class="inline-flex w-fit items-center gap-1.5 text-sm font-semibold text-primary-600 transition-colors hover:text-primary-700"
  >
    <ArrowLeft size={16} aria-hidden="true" />
    Kembali ke Pusat Bantuan
  </a>

  <p class="mt-5 text-[11px] font-semibold uppercase tracking-wide text-primary-600">{guide.category}</p>
  <h1 class="mt-1 font-serif text-[26px] font-bold leading-tight text-slate-800">{guide.title}</h1>
  {#if updated}
    <p class="mt-1.5 text-[13px] text-slate-400">Diperbarui {updated}</p>
  {/if}

  <div class="mt-6 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm lg:p-7">
    <GuideBody body={guide.body} />
  </div>

  {#if related.length > 0}
    <section class="mt-8" aria-labelledby="related-heading">
      <h2 id="related-heading" class="mb-3 font-serif text-lg font-bold text-slate-800">
        Panduan terkait
      </h2>
      <div class="flex flex-col gap-2.5">
        {#each related as item (item.slug)}
          <GuideCard guide={item} href={`${basePath}/${item.slug}`} />
        {/each}
      </div>
    </section>
  {/if}
</div>
