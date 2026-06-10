<script>
  import Seo from "$lib/components/Seo.svelte";
  import Privacy from "$lib/pages/marketing/Privacy.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb } from "$lib/seo/schema.js";

  const schema = breadcrumb([
    { name: "Beranda", path: "/" },
    { name: "Kebijakan Privasi", path: "/privasi" },
  ]);
</script>

<Seo
  title="Kebijakan Privasi - Suluk"
  description="Kebijakan privasi Suluk: bagaimana kami mengumpulkan, menggunakan, dan melindungi data Anda dan data jamaah."
  path="/privasi"
  {schema}
/>

<Privacy onNav={gotoMarketing} onGoToApp={gotoApp} />
