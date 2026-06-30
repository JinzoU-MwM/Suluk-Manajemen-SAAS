<script>
  import Seo from "$lib/components/Seo.svelte";
  import Pricing from "$lib/pages/marketing/Pricing.svelte";
  import { gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb } from "$lib/seo/schema.js";

  const schema = breadcrumb([
    { name: "Beranda", path: "/" },
    { name: "Harga", path: "/harga" },
  ]);
</script>

<Seo
  title="Harga Suluk — Paket Langganan Software Travel Umrah & Haji"
  description="Harga transparan Suluk: paket Gratis, Starter, Pro, dan Bisnis. Bayar tahunan dan hemat 2 bulan. Coba Pro gratis 14 hari, tanpa kartu kredit."
  path="/harga"
  {schema}
/>

<Pricing onGoToApp={gotoApp} />
