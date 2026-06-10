<script>
  import Seo from "$lib/components/Seo.svelte";
  import PackagesPage from "$lib/pages/PackagesPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Paket - Suluk" description="Kelola paket umrah & haji." path="/app/packages" robots="noindex,nofollow" />

<PackagesPage onNavigate={navigate} user={session.user} />
