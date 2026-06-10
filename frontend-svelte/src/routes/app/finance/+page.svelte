<script>
  import Seo from "$lib/components/Seo.svelte";
  import FinancePage from "$lib/pages/FinancePage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Keuangan - Suluk" description="Laporan keuangan operasional." path="/app/finance" robots="noindex,nofollow" />

<FinancePage onNavigate={navigate} user={session.user} />
