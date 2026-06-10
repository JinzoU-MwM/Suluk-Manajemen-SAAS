<script>
  import Seo from "$lib/components/Seo.svelte";
  import ScannerPage from "$lib/pages/ScannerPage.svelte";
  import { session, logoutAndRedirect } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Scanner Dokumen - Suluk" description="Scan KTP/KK, paspor, dan visa untuk ekstraksi data otomatis." path="/app/scanner" robots="noindex,nofollow" />

<ScannerPage
  user={session.user}
  subscription={session.subscription}
  onLogout={logoutAndRedirect}
  onSubscriptionChange={() => session.refresh()}
/>
