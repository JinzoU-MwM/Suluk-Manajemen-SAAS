<script>
  import Seo from "$lib/components/Seo.svelte";
  import ProGate from "$lib/components/ProGate.svelte";
  import InventoryPage from "$lib/pages/InventoryPage.svelte";
  import { goto } from "$app/navigation";
  import { session } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Inventory - Suluk" description="Manajemen inventory perlengkapan jamaah." path="/app/inventory" robots="noindex,nofollow" />

<ProGate
  name="Inventory Checklist"
  desc="Kelola stok koper, ihram, mukena, dan perlengkapan jamaah. Distribusi tercatat rapi, tidak ada yang terlewat."
  highlights={[
    "Kelola stok semua item di satu tempat",
    "Forecast kebutuhan otomatis berdasarkan data jamaah",
    "Tracking distribusi per jamaah - siapa sudah terima",
  ]}
>
  <InventoryPage isOpen={true} onClose={() => goto("/app")} groups={session.groups} isPro={session.isPro} />
</ProGate>
