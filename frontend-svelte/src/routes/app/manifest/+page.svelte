<script>
  import Seo from "$lib/components/Seo.svelte";
  import ProGate from "$lib/components/ProGate.svelte";
  import ManifestPage from "$lib/pages/ManifestPage.svelte";
</script>

<Seo title="Manifest Digital - Suluk" description="Bagikan manifest digital ber-PIN untuk mutawwif." path="/app/manifest" robots="noindex,nofollow" />

<ProGate
  name="Manifest Digital"
  desc="Bagikan manifest digital aman untuk mutawwif, lengkap dengan PIN dan link khusus per grup."
  highlights={[
    "Link manifest ber-PIN untuk setiap grup",
    "Data jamaah dan rooming tampil di HP mutawwif",
    "Akses bisa dibatasi masa berlakunya",
  ]}
>
  <ManifestPage />
</ProGate>
