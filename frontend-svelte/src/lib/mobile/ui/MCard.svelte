<script>
  let { pad = false, onclick = null, class: cls = "", style = "", children } = $props();
</script>

<svelte:element
  this={onclick ? "button" : "div"}
  type="button"
  class="m-card {pad ? 'm-card-pad' : ''} {cls}"
  style={style}
  onclick={onclick}
>
  {@render children?.()}
</svelte:element>
