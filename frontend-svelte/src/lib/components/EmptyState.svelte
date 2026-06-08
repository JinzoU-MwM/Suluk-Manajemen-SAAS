<!--
  EmptyState.svelte — friendly "no data" block built on the app.css .empty-state classes.
  Props: icon (lucide component), title, text. Optional `action` snippet for a CTA.
-->
<script>
  let { icon: Icon = null, title = "", text = "", action = null } = $props();
</script>

<div class="empty-state">
  {#if Icon}
    <Icon class="empty-state-icon" />
  {/if}
  <p class="empty-state-title">{title}</p>
  {#if text}
    <p class="empty-state-text">{text}</p>
  {/if}
  {#if action}
    <div class="mt-4">{@render action()}</div>
  {/if}
</div>
