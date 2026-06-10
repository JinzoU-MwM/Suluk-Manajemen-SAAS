<script>
  import Seo from "$lib/components/Seo.svelte";
  import VendorsPage from "$lib/pages/VendorsPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Vendor - Suluk" description="Kelola vendor & utang vendor." path="/app/vendors" robots="noindex,nofollow" />

<VendorsPage onNavigate={navigate} user={session.user} />
