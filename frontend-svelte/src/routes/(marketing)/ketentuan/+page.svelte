<script>
  import Seo from "$lib/components/Seo.svelte";
  import Terms from "$lib/pages/marketing/Terms.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb } from "$lib/seo/schema.js";

  const schema = breadcrumb([
    { name: "Beranda", path: "/" },
    { name: "Syarat & Ketentuan", path: "/ketentuan" },
  ]);
</script>

<Seo
  title="Syarat & Ketentuan - Suluk"
  description="Syarat dan ketentuan penggunaan layanan Suluk untuk travel umrah & haji."
  path="/ketentuan"
  {schema}
/>

<Terms onNav={gotoMarketing} onGoToApp={gotoApp} />
