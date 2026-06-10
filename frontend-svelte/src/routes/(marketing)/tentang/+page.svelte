<script>
  import Seo from "$lib/components/Seo.svelte";
  import About from "$lib/pages/marketing/About.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb } from "$lib/seo/schema.js";

  const schema = breadcrumb([
    { name: "Beranda", path: "/" },
    { name: "Tentang", path: "/tentang" },
  ]);
</script>

<Seo
  title="Tentang Suluk — Software Travel Umrah & Haji Indonesia"
  description="Suluk adalah platform operasional all-in-one untuk bisnis travel umrah & haji di Indonesia: data jamaah, paket, keuangan, dan dokumen dalam satu sistem."
  path="/tentang"
  {schema}
/>

<About onNav={gotoMarketing} onGoToApp={gotoApp} />
