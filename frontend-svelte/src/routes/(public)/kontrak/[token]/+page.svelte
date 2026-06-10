<script>
  import { page } from "$app/stores";
  import Seo from "$lib/components/Seo.svelte";
  import PublicContractSigningPage from "$lib/pages/PublicContractSigningPage.svelte";

  let token = $derived($page.params.token);
</script>

<Seo title="Tanda Tangan Kontrak - Suluk" description="Baca dan tanda tangani kontrak perjalanan secara digital dari HP." path="/kontrak/{token}" robots="noindex,nofollow" />

<PublicContractSigningPage {token} />
