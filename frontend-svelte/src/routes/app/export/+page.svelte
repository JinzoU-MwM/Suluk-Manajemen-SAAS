<script>
  import Seo from "$lib/components/Seo.svelte";
  import ExportPage from "$lib/pages/ExportPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Export - Suluk" description="Export data ke Excel Siskopatuh." path="/app/export" robots="noindex,nofollow" />

<ExportPage onNavigate={navigate} user={session.user} />
