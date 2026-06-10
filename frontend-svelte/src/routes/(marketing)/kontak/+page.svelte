<script>
  import Seo from "$lib/components/Seo.svelte";
  import Contact from "$lib/pages/marketing/Contact.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb } from "$lib/seo/schema.js";

  const schema = breadcrumb([
    { name: "Beranda", path: "/" },
    { name: "Kontak", path: "/kontak" },
  ]);
</script>

<Seo
  title="Hubungi Suluk — Kontak & Dukungan"
  description="Hubungi tim Suluk untuk demo, pertanyaan paket, atau dukungan teknis software travel umrah & haji."
  path="/kontak"
  {schema}
/>

<Contact onNav={gotoMarketing} onGoToApp={gotoApp} />
