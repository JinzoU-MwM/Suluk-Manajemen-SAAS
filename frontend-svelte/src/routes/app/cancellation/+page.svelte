<script>
  import Seo from "$lib/components/Seo.svelte";
  import CancellationPage from "$lib/pages/CancellationPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Pembatalan - Suluk" description="Kelola pembatalan & refund jamaah." path="/app/cancellation" robots="noindex,nofollow" />

<CancellationPage onNavigate={navigate} user={session.user} />
