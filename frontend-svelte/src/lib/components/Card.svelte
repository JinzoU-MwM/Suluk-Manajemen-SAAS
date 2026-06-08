<!--
  Card.svelte — consistent surface built on the app.css .card class.
  Optional `header` snippet renders in a .card-header; `padded` toggles body padding.
-->
<script>
  let { class: klass = "", padded = true, header = null, children, ...rest } = $props();
</script>

<div class={`card ${klass}`} {...rest}>
  {#if header}
    <div class="card-header">{@render header()}</div>
  {/if}
  <div class={padded ? "card-body" : ""}>
    {@render children?.()}
  </div>
</div>
