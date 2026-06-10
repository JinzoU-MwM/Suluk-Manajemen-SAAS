<script>
  import Seo from "$lib/components/Seo.svelte";
  import ProGate from "$lib/components/ProGate.svelte";
  import RoomingPage from "$lib/pages/RoomingPage.svelte";
  import { goto } from "$app/navigation";
  import { session } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Rooming - Suluk" description="Pengaturan rooming jamaah." path="/app/rooming" robots="noindex,nofollow" />

<ProGate
  name="Smart Auto-Rooming"
  desc="Bagi kamar hotel secara otomatis. Gender dipisahkan, keluarga disatukan - cukup 1 klik."
  highlights={[
    "Algoritma cerdas mengisi kamar Quad/Triple/Double",
    "Laki-laki & perempuan otomatis dipisahkan",
    "Drag & drop untuk penyesuaian manual",
  ]}
>
  <RoomingPage isOpen={true} onClose={() => goto("/app")} groups={session.groups} isPro={session.isPro} />
</ProGate>
