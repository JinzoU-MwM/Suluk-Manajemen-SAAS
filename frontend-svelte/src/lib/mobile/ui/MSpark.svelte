<script>
  let { data = [], color = "#fff", height = 40 } = $props();
  let max = $derived(Math.max(1, ...data));
</script>

<div style="display:flex;align-items:flex-end;gap:4px;height:{height}px">
  {#each data as v, i}
    <div
      class="m-spark-bar"
      style="flex:1;height:{(v / max) * 100}%;background:{color};opacity:{0.35 +
        (i / Math.max(1, data.length)) * 0.65};border-radius:3px;min-height:4px;animation-delay:{i * 0.06}s"
    ></div>
  {/each}
</div>
