<script>
  import Seo from "$lib/components/Seo.svelte";
  import CRMPage from "$lib/pages/CRMPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="CRM - Suluk" description="Pipeline prospek jamaah." path="/app/crm" robots="noindex,nofollow" />

<CRMPage onNavigate={navigate} user={session.user} />
