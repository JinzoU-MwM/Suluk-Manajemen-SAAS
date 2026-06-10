<script>
  import Seo from "$lib/components/Seo.svelte";
  import ProGate from "$lib/components/ProGate.svelte";
  import TeamPage from "$lib/pages/TeamPage.svelte";
  import { session } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Tim - Suluk" description="Manajemen tim operasional." path="/app/team" robots="noindex,nofollow" />

<ProGate
  name="Manajemen Tim"
  desc="Kelola tim Mutawwif dan staf operasional Anda. Bagikan akses manifest digital yang aman."
  highlights={[
    "Undang anggota tim via email",
    "Atur hak akses per role (Admin / Mutawwif)",
    "Manifest digital aman dengan PIN untuk Mutawwif",
  ]}
>
  <TeamPage isPro={session.isPro} />
</ProGate>
