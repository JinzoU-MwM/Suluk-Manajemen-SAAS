<!--
  BrandLogo.svelte — Suluk brand lockup (real double-S mark + "ERP for Travel").
  variant: 'default' (dark text, for light surfaces) | 'light' (white mark + text, for dark/green surfaces)
-->
<script>
  let { size = 'default', iconOnly = false, variant = 'default' } = $props();

  const markSizes = {
    small: 'h-7',
    default: 'h-9',
    large: 'h-11',
  };
  const titleSizes = {
    small: 'text-lg',
    default: 'text-xl',
    large: 'text-3xl',
  };

  const mark = $derived(variant === 'light' ? '/brand/suluk-mark-white.png' : '/brand/suluk-mark.png');
  const titleColor = $derived(variant === 'light' ? 'text-white' : 'text-primary-800');
</script>

{#if iconOnly}
  <img src="/brand/suluk-icon.png" alt="Suluk" class="{markSizes[size]} w-auto" />
{:else}
  <div class="flex items-center gap-2.5">
    <img src={mark} alt="Suluk" class="{markSizes[size]} w-auto" />
    <div class="flex flex-col leading-none">
      <span class="font-serif font-bold {titleSizes[size]} {titleColor}">Suluk</span>
      <span class="mt-0.5 text-[9px] font-semibold uppercase tracking-[0.22em] text-gold-500">ERP for Travel</span>
    </div>
  </div>
{/if}
