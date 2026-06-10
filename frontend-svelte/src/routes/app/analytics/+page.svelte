<script>
  import Seo from "$lib/components/Seo.svelte";
  import ProGate from "$lib/components/ProGate.svelte";
  import AnalyticsPage from "$lib/pages/AnalyticsPage.svelte";
</script>

<Seo title="Analytics - Suluk" description="Statistik operasional jamaah dan grup." path="/app/analytics" robots="noindex,nofollow" />

<ProGate
  name="Analytics Operasional"
  desc="Pantau tren jamaah, kelengkapan perlengkapan, gender, dan performa grup dalam satu tempat."
  highlights={[
    "Ringkasan data jamaah dan grup",
    "Trend jamaah per bulan",
    "Komposisi gender dan kelengkapan perlengkapan",
  ]}
>
  <AnalyticsPage />
</ProGate>
