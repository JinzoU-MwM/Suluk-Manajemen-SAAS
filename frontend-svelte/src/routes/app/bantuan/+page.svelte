<script>
  import Seo from "$lib/components/Seo.svelte";
  import BantuanListPage from "$lib/pages/BantuanListPage.svelte";
</script>

<Seo
  title="Pusat Bantuan - Suluk"
  description="Panduan penggunaan aplikasi Suluk untuk staf travel: jamaah, paket, keuangan, dan keberangkatan."
  path="/app/bantuan"
  robots="noindex,nofollow"
/>

<BantuanListPage area="app" basePath="/app/bantuan" />
