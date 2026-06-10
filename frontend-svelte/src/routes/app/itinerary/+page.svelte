<script>
  import Seo from "$lib/components/Seo.svelte";
  import ProGate from "$lib/components/ProGate.svelte";
  import ItineraryPage from "$lib/pages/ItineraryPage.svelte";
  import { session } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Itinerary - Suluk" description="Jadwal perjalanan jamaah." path="/app/itinerary" robots="noindex,nofollow" />

<ProGate
  name="Jadwal Perjalanan"
  desc="Atur jadwal perjalanan umrah dari hari ke hari. Jamaah dan Mutawwif bisa melihat agenda via manifest digital."
  highlights={[
    "Buat jadwal harian lengkap dari keberangkatan sampai pulang",
    "Otomatis tampil di manifest digital Mutawwif",
    "Template jadwal untuk reuse di grup berikutnya",
  ]}
>
  <ItineraryPage groups={session.groups} isPro={session.isPro} />
</ProGate>
