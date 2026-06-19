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

  const inlineCls =
    "flex h-10 w-10 items-center justify-center rounded-[var(--radius)] text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-400";
  const floatCls =
    "fixed bottom-5 right-5 z-[60] flex h-12 w-12 items-center justify-center rounded-full bg-white text-primary-600 shadow-lg ring-1 ring-slate-200 transition-colors hover:bg-primary-600 hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-primary-400";
</script>

{#if !onHelpPage}
  <a {href} class={floating ? floatCls : inlineCls} title={label} aria-label={label}>
    <HelpCircle size={floating ? 22 : 20} aria-hidden="true" />
  </a>
{/if}
