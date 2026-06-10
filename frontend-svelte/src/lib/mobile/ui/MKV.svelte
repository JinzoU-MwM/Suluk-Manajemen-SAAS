<script>
  // items: [{ k, v, full }]
  let { items = [] } = $props();
</script>

<div style="display:grid;grid-template-columns:1fr 1fr;gap:15px 16px">
  {#each items as it}
    <div style={it.full ? "grid-column:1 / -1" : ""}>
      <div style="font-size:11px;font-weight:700;letter-spacing:.03em;text-transform:uppercase;color:var(--c-faint);margin-bottom:4px">
        {it.k}
      </div>
      <div style="font-size:14.5px;font-weight:600;color:var(--c-ink)">{it.v}</div>
    </div>
  {/each}
</div>
