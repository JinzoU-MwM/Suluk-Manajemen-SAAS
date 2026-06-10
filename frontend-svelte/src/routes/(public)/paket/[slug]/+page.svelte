<script>
  import { page } from "$app/stores";
  import Seo from "$lib/components/Seo.svelte";
  import PublicPackagePage from "$lib/pages/PublicPackagePage.svelte";

  let slug = $derived($page.params.slug);
</script>

<Seo title="Paket Umrah - Suluk" description="Lihat detail paket umrah, harga per tipe kamar, dan sisa kursi." path="/paket/{slug}" />

<PublicPackagePage {slug} />
