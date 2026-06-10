<script>
  import { Search } from "lucide-svelte";
  let { value = $bindable(""), placeholder = "Cari…" } = $props();
</script>

<div class="m-search">
  <Search size={17} />
  <input bind:value {placeholder} />
</div>
