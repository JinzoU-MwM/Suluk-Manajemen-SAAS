<script>
  let { variant = "primary", icon = null, onclick = null, style = "", disabled = false, children } = $props();
  const Icon = icon;
</script>

<button type="button" class="m-btn m-btn-{variant}" {onclick} {style} {disabled}>
  {#if Icon}<Icon size={18} />{/if}{@render children?.()}
</button>
