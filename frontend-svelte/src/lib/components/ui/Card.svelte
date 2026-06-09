<!-- Card — Suluk design primitive (mirrors components.jsx Card). -->
<script>
  let { pad = true, hover = false, onclick = null, class: cls = "", style = "", children } = $props();
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  class="suluk-card {hover ? 'suluk-card-hover' : ''} {cls}"
  style="background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius-lg);box-shadow:var(--shadow-sm);{pad ? 'padding:var(--pad);' : ''}{onclick ? 'cursor:pointer;' : ''}{style}"
  role={onclick ? "button" : undefined}
  tabindex={onclick ? 0 : undefined}
  onclick={onclick}
  onkeydown={onclick ? (e) => { if (e.key === "Enter") onclick(e); } : undefined}
>
  {@render children?.()}
</div>
