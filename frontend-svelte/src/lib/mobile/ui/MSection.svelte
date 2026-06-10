<script>
  let { label = "", action = null, style = "", children } = $props();
</script>

<div style="padding:0 18px;{style}">
  {#if label || action}
    <div style="display:flex;justify-content:space-between;align-items:flex-end">
      {#if label}<div class="m-label">{label}</div>{/if}
      {#if action}{@render action()}{/if}
    </div>
  {/if}
  {@render children?.()}
</div>
