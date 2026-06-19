<script>
  // Kotak pencarian panduan dengan debounce. Menulis ke prop `value` (bindable)
  // ~200 ms setelah pengguna berhenti mengetik agar daftar tidak berkedip.
  import { Search, X } from "lucide-svelte";

  let {
    value = $bindable(""),
    placeholder = "Cari panduan…",
    label = "Cari panduan",
  } = $props();

  let raw = $state(value);
  /** @type {ReturnType<typeof setTimeout> | undefined} */
  let timer;

  function onInput(event) {
    raw = event.currentTarget.value;
    clearTimeout(timer);
    timer = setTimeout(() => {
      value = raw;
    }, 200);
  }

  function clear() {
    clearTimeout(timer);
    raw = "";
    value = "";
  }
</script>

<div class="help-search" role="search">
  <label class="sr-only" for="help-search-input">{label}</label>
  <Search class="help-search-ico" size={18} aria-hidden="true" />
  <input
    id="help-search-input"
    type="search"
    {placeholder}
    value={raw}
    oninput={onInput}
    autocomplete="off"
  />
  {#if raw}
    <button type="button" class="help-search-clear" onclick={clear} aria-label="Hapus pencarian">
      <X size={16} aria-hidden="true" />
    </button>
  {/if}
</div>

<style>
  .help-search {
    position: relative;
    display: flex;
    align-items: center;
  }
  .help-search :global(.help-search-ico) {
    position: absolute;
    left: 14px;
    color: var(--c-faint);
    pointer-events: none;
  }
  .help-search input {
    width: 100%;
    padding: 11px 40px 11px 42px;
    font-size: 14.5px;
    color: var(--c-ink);
    background: var(--c-surface);
    border: 1px solid var(--c-line);
    border-radius: var(--radius-lg);
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .help-search input::placeholder {
    color: var(--c-faint);
  }
  .help-search input:focus {
    outline: none;
    border-color: var(--c-primary);
    box-shadow: 0 0 0 3px var(--c-primary-tint);
  }
  /* Sembunyikan ikon clear bawaan input[type=search] agar tak dobel. */
  .help-search input::-webkit-search-cancel-button {
    -webkit-appearance: none;
    appearance: none;
  }
  .help-search-clear {
    position: absolute;
    right: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 5px;
    color: var(--c-muted);
    background: none;
    border: none;
    border-radius: 999px;
    cursor: pointer;
  }
  .help-search-clear:hover {
    background: var(--c-bg-2);
    color: var(--c-ink);
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
</style>
