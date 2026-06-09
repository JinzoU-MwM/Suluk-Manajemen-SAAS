<!-- ProgressBar — Suluk design primitive (mirrors components.jsx ProgressBar). -->
<script>
  let {
    value = 0,
    max = 100,
    color = "var(--c-primary)",
    height = 7,
    track = "var(--c-bg-2)",
  } = $props();
  let pct = $derived(Math.min(100, max > 0 ? (value / max) * 100 : 0));
</script>

<div style="height:{height}px;background:{track};border-radius:999px;overflow:hidden;width:100%">
  <div style="width:{pct}%;height:100%;background:{color};border-radius:999px;transition:width .6s cubic-bezier(.2,.7,.3,1)"></div>
</div>
