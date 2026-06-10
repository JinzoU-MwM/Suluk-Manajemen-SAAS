<script>
  import Seo from "$lib/components/Seo.svelte";
  import DocumentsPage from "$lib/pages/DocumentsPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Dokumen - Suluk" description="Kelola dokumen jamaah." path="/app/documents" robots="noindex,nofollow" />

<DocumentsPage onNavigate={navigate} user={session.user} />
