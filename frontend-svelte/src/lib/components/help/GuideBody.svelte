<script>
  // Merender body panduan (HelpBlock[]) menjadi HTML semantik. Tidak ada HTML
  // mentah — teks selalu masuk sebagai text node, jadi aman dari injeksi.
  import Callout from "./Callout.svelte";

  /** @type {{ body: import('$lib/content/help/index.js').HelpBlock[] }} */
  let { body = [] } = $props();
</script>

<div class="guide-body">
  {#each body as block}
    {#if block.type === "h2"}
      <h2>{block.text}</h2>
    {:else if block.type === "p"}
      <p>{block.text}</p>
    {:else if block.type === "ul"}
      <ul>
        {#each block.items ?? [] as item}<li>{item}</li>{/each}
      </ul>
    {:else if block.type === "ol"}
      <ol>
        {#each block.items ?? [] as item}<li>{item}</li>{/each}
      </ol>
    {:else if block.type === "callout"}
      <Callout variant={block.variant} text={block.text} />
    {/if}
  {/each}
</div>

<style>
  .guide-body {
    color: var(--c-ink-soft);
    font-size: 15px;
    line-height: 1.7;
  }
  .guide-body h2 {
    font-family: var(--font-serif, "Playfair Display", Georgia, serif);
    font-size: 19px;
    font-weight: 700;
    color: var(--c-ink);
    margin: 26px 0 10px;
  }
  .guide-body p {
    margin: 0 0 14px;
  }
  .guide-body ul,
  .guide-body ol {
    margin: 0 0 14px;
    padding-left: 22px;
  }
  .guide-body li {
    margin-bottom: 7px;
  }
  .guide-body ol {
    list-style: decimal;
  }
  .guide-body ul {
    list-style: disc;
  }
</style>
