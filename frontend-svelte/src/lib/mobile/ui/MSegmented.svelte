<script>
  let { tabs = [], value, onChange } = $props();
  const val = (t) => (typeof t === "string" ? t : t.value);
  const lab = (t) => (typeof t === "string" ? t : t.label);
</script>

<div class="m-seg">
  {#each tabs as t}
    <button type="button" class={val(t) === value ? "on" : ""} onclick={() => onChange(val(t))}>{lab(t)}</button>
  {/each}
</div>
