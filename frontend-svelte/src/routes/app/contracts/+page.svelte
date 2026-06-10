<script>
  import Seo from "$lib/components/Seo.svelte";
  import ProGate from "$lib/components/ProGate.svelte";
  import ContractsPage from "$lib/pages/ContractsPage.svelte";
  import { session } from "$lib/stores/session.svelte.js";
</script>

<Seo title="E-Kontrak - Suluk" description="Kelola template kontrak digital untuk jamaah." path="/app/contracts" robots="noindex,nofollow" />

<ProGate
  name="E-Kontrak Digital"
  desc="Kelola template kontrak digital dan siapkan fondasi penandatanganan jamaah."
  highlights={[
    "Template kontrak dengan variabel otomatis",
    "Preview isi kontrak sebelum dikirim",
    "Fondasi untuk monitoring tanda tangan",
  ]}
>
  <ContractsPage user={session.user} />
</ProGate>
