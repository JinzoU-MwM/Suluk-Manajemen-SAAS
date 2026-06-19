<script>
  // Kotak pencarian panduan dengan debounce. Menulis ke prop `value` (bindable)
  // ~200 ms setelah pengguna berhenti mengetik agar daftar tidak berkedip.
  // Gaya mengikuti input standar aplikasi (Tailwind, skala brand).
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

<div class="relative" role="search">
  <label class="sr-only" for="help-search-input">{label}</label>
  <Search
    class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400"
    aria-hidden="true"
  />
  <input
    id="help-search-input"
    type="search"
    {placeholder}
    value={raw}
    oninput={onInput}
    autocomplete="off"
    class="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-9 pr-9 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100 [&::-webkit-search-cancel-button]:hidden"
  />
  {#if raw}
    <button
      type="button"
      onclick={clear}
      aria-label="Hapus pencarian"
      class="absolute right-2.5 top-1/2 flex -translate-y-1/2 items-center justify-center rounded-full p-1 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
    >
      <X size={15} aria-hidden="true" />
    </button>
  {/if}
</div>

<style>
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
