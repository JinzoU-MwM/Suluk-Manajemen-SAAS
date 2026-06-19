<script>
  // Kartu satu panduan di daftar Pusat Bantuan. Gaya mengikuti item daftar
  // standar aplikasi (rounded-xl, shadow-sm, ring slate, hover ring primary).
  import { ChevronRight } from "lucide-svelte";

  /**
   * @type {{
   *   guide: import('$lib/content/help/index.js').HelpGuide,
   *   href: string,
   *   showCategory?: boolean,
   * }}
   */
  let { guide, href, showCategory = false } = $props();
</script>

<a
  {href}
  class="group flex items-center justify-between gap-4 rounded-xl bg-white p-4 shadow-sm ring-1 ring-slate-200/60 transition-all hover:ring-primary-200 focus:outline-none focus:ring-2 focus:ring-primary-400"
>
  <span class="min-w-0 flex-1">
    {#if showCategory}
      <span class="text-[11px] font-semibold uppercase tracking-wide text-primary-600">
        {guide.category}
      </span>
    {/if}
    <span class="block text-[15px] font-semibold text-slate-800">{guide.title}</span>
    <span class="mt-0.5 block text-[13px] leading-relaxed text-slate-500">{guide.summary}</span>
  </span>
  <ChevronRight
    class="h-4 w-4 shrink-0 text-slate-300 transition-transform group-hover:translate-x-0.5"
    aria-hidden="true"
  />
</a>
