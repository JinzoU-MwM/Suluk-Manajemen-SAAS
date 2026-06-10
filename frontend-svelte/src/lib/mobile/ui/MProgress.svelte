<script>
  let { value = 0, max = 100, color = "var(--c-primary)" } = $props();
  let pct = $derived(Math.min(100, max ? (value / max) * 100 : 0));
</script>

<div class="m-prog"><i style="width:{pct}%;background:{color}"></i></div>
