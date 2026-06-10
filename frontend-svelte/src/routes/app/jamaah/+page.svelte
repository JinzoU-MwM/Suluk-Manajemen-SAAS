<script>
  import Seo from "$lib/components/Seo.svelte";
  import DataJamaahPage from "$lib/pages/DataJamaahPage.svelte";
  import { navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Data Jamaah - Suluk" description="Kelola data jamaah per grup keberangkatan." path="/app/jamaah" robots="noindex,nofollow" />

<DataJamaahPage onNavigate={navigate} />
