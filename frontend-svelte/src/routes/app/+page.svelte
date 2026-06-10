<script>
  import Seo from "$lib/components/Seo.svelte";
  import Dashboard from "$lib/pages/Dashboard.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Dashboard - Suluk" description="Dashboard operasional jamaah." path="/app" robots="noindex,nofollow" />

<Dashboard
  user={session.user}
  subscription={session.subscription}
  onNavigate={navigate}
  onUpgrade={() => session.openUpgrade()}
/>
