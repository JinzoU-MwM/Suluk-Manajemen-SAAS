<script>
  import { X } from "lucide-svelte";
  let { open = false, onClose, title = "", children, footer = null } = $props();
  $effect(() => {
    if (!open) return;
    const onKey = (e) => {
      if (e.key === "Escape") onClose?.();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });
</script>

{#if open}
  <div class="m-sheet-wrap">
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div class="m-sheet-bd" onclick={onClose}></div>
    <div class="m-sheet">
      <div class="m-sheet-grip"></div>
      <div class="m-sheet-head">
        <div class="m-sheet-title">{title}</div>
        <button type="button" class="m-sheet-x" onclick={onClose}><X size={17} /></button>
      </div>
      <div class="m-sheet-body">{@render children?.()}</div>
      {#if footer}<div style="padding:14px 20px 0;display:flex;gap:10px">{@render footer()}</div>{/if}
    </div>
  </div>
{/if}
