<!-- FilterTabs — Suluk design segmented tabs (mirrors components.jsx FilterTabs). -->
<script>
  // tabs: array of { value, label, count? } or string
  let { tabs = [], value = "", onChange = () => {} } = $props();
  function norm(t) {
    return typeof t === "string" ? { value: t, label: t, count: null } : { count: null, ...t };
  }
</script>

<div style="display:inline-flex;gap:4px;background:var(--c-bg-2);padding:4px;border-radius:var(--radius)">
  {#each tabs.map(norm) as t}
    {@const active = t.value === value}
    <button
      type="button"
      onclick={() => onChange(t.value)}
      style="padding:7px 14px;font-size:13px;font-weight:600;border-radius:var(--radius-sm);display:inline-flex;align-items:center;gap:7px;transition:all .15s;color:var({active ? '--c-ink' : '--c-muted'});background:{active ? 'var(--c-surface)' : 'transparent'};box-shadow:{active ? 'var(--shadow-sm)' : 'none'}"
    >
      {t.label}
      {#if t.count != null}
        <span style="font-size:11px;font-weight:700;color:var({active ? '--c-primary' : '--c-faint'})">{t.count}</span>
      {/if}
    </button>
  {/each}
</div>
