<script>
  // Tombol "?" kontekstual: menautkan ke panduan Pusat Bantuan untuk halaman
  // yang sedang dibuka (berdasarkan segmen rute). Bila halaman belum punya
  // panduan khusus, tombol menaut ke daftar Pusat Bantuan area tersebut.
  // Disembunyikan saat sudah berada di dalam Pusat Bantuan.
  import { page } from "$app/stores";
  import { HelpCircle } from "lucide-svelte";
  import { getGuideSlugForRoute } from "$lib/content/help/index.js";

  /** @type {{ area: import('$lib/content/help/index.js').HelpArea, floating?: boolean }} */
  let { area, floating = false } = $props();

  let areaBase = $derived(`/${area}`);
  let helpBase = $derived(`/${area}/bantuan`);

  // Segmen pertama setelah prefix area; "" untuk halaman indeks area.
  let segment = $derived.by(() => {
    const path = $page.url.pathname;
    const rest = path.startsWith(areaBase) ? path.slice(areaBase.length) : "";
    return rest.replace(/^\//, "").split("/")[0] || "";
  });

  let onHelpPage = $derived(segment === "bantuan");
  let slug = $derived(getGuideSlugForRoute(area, segment));
  let href = $derived(slug ? `${helpBase}/${slug}` : helpBase);
  let label = $derived(slug ? "Buka panduan halaman ini" : "Buka Pusat Bantuan");
</script>

{#if !onHelpPage}
  <a
    {href}
    class="help-hint {floating ? 'help-hint-float' : 'help-hint-inline'}"
    title={label}
    aria-label={label}
  >
    <HelpCircle size={floating ? 22 : 20} aria-hidden="true" />
  </a>
{/if}

<style>
  .help-hint {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--c-muted);
    border-radius: var(--radius);
    transition: background 0.14s, color 0.14s, border-color 0.14s;
  }
  .help-hint:focus-visible {
    outline: 2px solid var(--c-primary);
    outline-offset: 2px;
  }
  .help-hint-inline {
    width: 40px;
    height: 40px;
  }
  .help-hint-inline:hover {
    background: var(--c-bg);
    color: var(--c-ink);
  }
  .help-hint-float {
    position: fixed;
    right: 20px;
    bottom: 20px;
    z-index: 60;
    width: 46px;
    height: 46px;
    border-radius: 999px;
    color: var(--c-primary);
    background: var(--c-surface);
    border: 1px solid var(--c-line);
    box-shadow: 0 6px 20px rgba(15, 61, 46, 0.14);
  }
  .help-hint-float:hover {
    color: #fff;
    background: var(--c-primary);
    border-color: var(--c-primary);
  }
</style>
