<script>
  import Seo from "$lib/components/Seo.svelte";
  import Unduh from "$lib/pages/marketing/Unduh.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb } from "$lib/seo/schema.js";

  const schema = breadcrumb([
    { name: "Beranda", path: "/" },
    { name: "Unduh Aplikasi", path: "/unduh" },
  ]);
</script>

<Seo
  title="Unduh Aplikasi Suluk — Kelola Travel Umrah dari Ponsel"
  description="Unduh aplikasi mobile Suluk untuk Android. Scan dokumen, catat pembayaran, setujui refund, dan pantau jamaah langsung dari genggaman."
  path="/unduh"
  {schema}
/>

<Unduh onNav={gotoMarketing} onGoToApp={gotoApp} />
