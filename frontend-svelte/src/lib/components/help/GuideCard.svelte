<script>
  // Kartu satu panduan di daftar Pusat Bantuan. Tautan ke halaman detail.
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

<a class="guide-card" {href}>
  <span class="guide-card-body">
    {#if showCategory}
      <span class="guide-card-cat">{guide.category}</span>
    {/if}
    <span class="guide-card-title">{guide.title}</span>
    <span class="guide-card-sum">{guide.summary}</span>
  </span>
  <ChevronRight class="guide-card-chev" size={18} aria-hidden="true" />
</a>

<style>
  .guide-card {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 16px 18px;
    border: 1px solid var(--c-line);
    border-radius: var(--radius-lg);
    background: var(--c-surface);
    text-decoration: none;
    transition: border-color 0.15s, box-shadow 0.15s, transform 0.1s;
  }
  .guide-card:hover {
    border-color: color-mix(in srgb, var(--c-primary) 40%, var(--c-line));
    box-shadow: 0 6px 18px rgba(15, 61, 46, 0.08);
    transform: translateY(-1px);
  }
  .guide-card:focus-visible {
    outline: 2px solid var(--c-primary);
    outline-offset: 2px;
  }
  .guide-card-body {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
    flex: 1;
  }
  .guide-card-cat {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--c-primary);
  }
  .guide-card-title {
    font-size: 15.5px;
    font-weight: 700;
    color: var(--c-ink);
  }
  .guide-card-sum {
    font-size: 13.5px;
    line-height: 1.55;
    color: var(--c-muted);
  }
  .guide-card :global(.guide-card-chev) {
    flex-shrink: 0;
    color: var(--c-faint);
  }
</style>
