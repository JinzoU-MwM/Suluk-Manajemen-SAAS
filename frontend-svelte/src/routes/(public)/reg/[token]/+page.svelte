<script>
  import { page } from "$app/stores";
  import Seo from "$lib/components/Seo.svelte";
  import PublicRegistrationPage from "$lib/pages/PublicRegistrationPage.svelte";

  let token = $derived($page.params.token);
</script>

<Seo title="Form Pendaftaran Jamaah - Suluk" description="Form pendaftaran jamaah online." path="/reg/{token}" robots="noindex,nofollow" />

<PublicRegistrationPage {token} />
