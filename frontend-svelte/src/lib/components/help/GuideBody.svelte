<script>
  // Merender body panduan (HelpBlock[]) menjadi HTML semantik dengan gaya
  // Tailwind aplikasi. Tidak ada HTML mentah — teks selalu masuk sebagai text
  // node, jadi aman dari injeksi.
  import Callout from "./Callout.svelte";

  /** @type {{ body: import('$lib/content/help/index.js').HelpBlock[] }} */
  let { body = [] } = $props();
</script>

<div class="text-[15px] leading-relaxed text-slate-600">
  {#each body as block}
    {#if block.type === "h2"}
      <h2 class="mb-2 mt-6 font-serif text-lg font-bold text-slate-800 first:mt-0">{block.text}</h2>
    {:else if block.type === "p"}
      <p class="mb-3.5">{block.text}</p>
    {:else if block.type === "ul"}
      <ul class="mb-3.5 list-disc space-y-1.5 pl-5">
        {#each block.items ?? [] as item}<li>{item}</li>{/each}
      </ul>
    {:else if block.type === "ol"}
      <ol class="mb-3.5 list-decimal space-y-1.5 pl-5">
        {#each block.items ?? [] as item}<li>{item}</li>{/each}
      </ol>
    {:else if block.type === "callout"}
      <Callout variant={block.variant} text={block.text} />
    {/if}
  {/each}
</div>
