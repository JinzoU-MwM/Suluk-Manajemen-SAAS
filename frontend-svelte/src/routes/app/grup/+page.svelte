<script>
  import Seo from "$lib/components/Seo.svelte";
  import GroupsPage from "$lib/pages/GroupsPage.svelte";
  import { navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Grup & Hotel - Suluk" description="Kelola grup keberangkatan dan persiapan hotel." path="/app/grup" robots="noindex,nofollow" />

<GroupsPage onNavigate={navigate} />
