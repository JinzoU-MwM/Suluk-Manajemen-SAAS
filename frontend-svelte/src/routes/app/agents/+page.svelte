<script>
  import Seo from "$lib/components/Seo.svelte";
  import AgentsPage from "$lib/pages/AgentsPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Agen - Suluk" description="Kelola agen mitra & komisi." path="/app/agents" robots="noindex,nofollow" />

<AgentsPage onNavigate={navigate} user={session.user} />
