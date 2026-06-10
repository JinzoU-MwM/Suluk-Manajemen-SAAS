<script>
  import Seo from "$lib/components/Seo.svelte";
  import InventoryPage from "$lib/pages/InventoryPage.svelte";
  import { goto } from "$app/navigation";
  import { session } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Inventory - Suluk" description="Manajemen inventory perlengkapan jamaah." path="/app/stock" robots="noindex,nofollow" />

<InventoryPage isOpen={true} onClose={() => goto("/app")} groups={session.groups} isPro={session.isPro} />
