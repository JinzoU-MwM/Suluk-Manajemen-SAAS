<script>
  import Seo from "$lib/components/Seo.svelte";
  import InvoicesPage from "$lib/pages/InvoicesPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Invoice - Suluk" description="Tagihan & pembayaran jamaah." path="/app/invoices" robots="noindex,nofollow" />

<InvoicesPage onNavigate={navigate} user={session.user} />
