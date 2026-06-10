<script>
  import "../app.css";
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Toast from "$lib/components/Toast.svelte";

  let { children } = $props();

  // Map a legacy hash URL (suluk.site/#/...) to its new clean path, so links
  // shared before the SvelteKit migration still resolve.
  function legacyHashToPath(h) {
    const STATIC = {
      "#/tentang": "/tentang",
      "#/kontak": "/kontak",
      "#/privasi": "/privasi",
      "#/ketentuan": "/ketentuan",
      "#/software-travel-umrah": "/software-travel-umrah",
      "#/fitur/invoice-umrah": "/fitur/invoice-umrah",
      "#/fitur/crm-jamaah": "/fitur/crm-jamaah",
      "#/fitur/laporan-keuangan": "/fitur/laporan-keuangan",
      "#/fitur/e-kontrak": "/fitur/e-kontrak",
      "#/fitur/penggajian": "/fitur/penggajian",
      "#/unduh": "/unduh",
      "#/app": "/mobile",
      "#/super-admin": "/super-admin",
      "#/dashboard": "/app",
    };
    if (STATIC[h]) return STATIC[h];
    let m;
    if ((m = h.match(/^#\/m\/([a-f0-9]+)$/i))) return `/m/${m[1]}`;
    if ((m = h.match(/^#\/reg\/([a-zA-Z0-9_-]+)$/i))) return `/reg/${m[1]}`;
    if ((m = h.match(/^#\/paket\/([a-zA-Z0-9_-]+)$/i))) return `/paket/${m[1]}`;
    if ((m = h.match(/^#\/kontrak\/([a-zA-Z0-9_-]+)$/i))) return `/kontrak/${m[1]}`;
    return null;
  }

  function redirectLegacyHash() {
    const h = window.location.hash;
    if (!h || h === "#" || h === "#/") return;
    const dest = legacyHashToPath(h);
    if (dest) goto(dest, { replaceState: true });
  }

  onMount(() => {
    // Clean up any leftover dark mode from previous versions.
    document.documentElement.classList.remove("dark");
    try {
      localStorage.removeItem("darkMode");
    } catch {}

    redirectLegacyHash();
    window.addEventListener("hashchange", redirectLegacyHash);
    return () => window.removeEventListener("hashchange", redirectLegacyHash);
  });
</script>

{@render children()}

<!-- Global toast — the mobile shell has its own MToast, so suppress it there. -->
{#if !$page.url.pathname.startsWith("/mobile")}
  <Toast />
{/if}
