<script>
  import { page } from "$app/stores";
  import Seo from "$lib/components/Seo.svelte";
  import MutawwifManifest from "$lib/pages/MutawwifManifest.svelte";

  let token = $derived($page.params.token);
</script>

<Seo title="Manifest Mutawwif - Suluk" description="Manifest jamaah untuk operasional mutawwif." path="/m/{token}" robots="noindex,nofollow" />

<MutawwifManifest {token} />
