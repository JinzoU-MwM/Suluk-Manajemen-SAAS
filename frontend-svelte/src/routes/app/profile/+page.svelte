<script>
  import Seo from "$lib/components/Seo.svelte";
  import ProfilePage from "$lib/pages/ProfilePage.svelte";
  import { session, logoutAndRedirect } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Profil - Suluk" description="Pengaturan akun Suluk." path="/app/profile" robots="noindex,nofollow" />

<ProfilePage
  onLogout={logoutAndRedirect}
  user={session.user}
  subscription={session.subscription}
  trialAvailable={session.trialAvailable}
  groups={session.groups}
  onUpgradeRequest={() => session.openUpgrade()}
/>
