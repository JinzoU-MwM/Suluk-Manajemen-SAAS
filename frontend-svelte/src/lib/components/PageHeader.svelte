<!--
  PageHeader.svelte — Suluk design page header used across admin pages:
  green uppercase kicker + Playfair title + muted subtitle + right-aligned actions.
-->
<script>
  let { kicker = "", title = "", subtitle = "", actions = null } = $props();
</script>

<div class="mb-6 flex flex-wrap items-end justify-between gap-4">
  <div class="min-w-0">
    {#if kicker}
      <p class="mb-1.5 text-xs font-bold uppercase tracking-[0.08em] text-primary-600">{kicker}</p>
    {/if}
    <h1 class="font-serif text-[25px] font-bold tracking-tight text-[#10211c]">{title}</h1>
    {#if subtitle}
      <p class="mt-1.5 max-w-2xl text-sm text-slate-500">{subtitle}</p>
    {/if}
  </div>
  {#if actions}
    <div class="flex flex-shrink-0 items-center gap-2.5">{@render actions()}</div>
  {/if}
</div>
