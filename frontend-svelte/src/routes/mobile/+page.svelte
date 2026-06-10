<script>
  // Mobile shell (the Capacitor wrapper / mobile PWA). In mobile mode all auth
  // flows stay inside this shell — login and Pro-gate render here, not the
  // desktop dashboard. Replaces App.svelte's "mobile" branch.
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import Seo from "$lib/components/Seo.svelte";
  import MobileApp from "$lib/mobile/MobileApp.svelte";
  import MobileLogin from "$lib/mobile/screens/MobileLogin.svelte";
  import MobileProGate from "$lib/mobile/screens/MobileProGate.svelte";
  import UpgradeModal from "$lib/components/UpgradeModal.svelte";
  import { session } from "$lib/stores/session.svelte.js";

  let ready = $state(false);

  function handleLoginSuccess(userData) {
    session.setUserAndLoad(userData);
  }

  function exitToDesktop() {
    session.setMobileMode(false);
    goto("/app");
  }

  async function handleLogout() {
    await session.logout();
    // Stay in the mobile shell — it now renders MobileLogin.
  }

  onMount(async () => {
    session.setMobileMode(true);
    session.hydrateFromCache();
    await session.loadSession();
    ready = true;
  });
</script>

<Seo title="Suluk App" description="Aplikasi mobile Suluk." path="/mobile" robots="noindex,nofollow" />

{#if !ready}
  <div class="min-h-screen flex items-center justify-center" style="background:#0F3D2E;color:#fff">Memuat…</div>
{:else if session.user}
  {#if !session.subLoaded}
    <div class="min-h-screen flex items-center justify-center" style="background:#0F3D2E;color:#fff">Memuat…</div>
  {:else if !session.isPro}
    <MobileProGate
      user={session.user}
      plan={session.subscription?.plan ?? ""}
      onUpgrade={() => session.openUpgrade()}
      onExit={exitToDesktop}
      onLogout={handleLogout}
    />
  {:else}
    <MobileApp user={session.user} onExit={exitToDesktop} onLogout={handleLogout} />
  {/if}
{:else}
  <MobileLogin onLoginSuccess={handleLoginSuccess} />
{/if}

<UpgradeModal
  show={session.upgradeOpen}
  onClose={() => session.closeUpgrade()}
  onSuccess={async (newSub) => {
    session.subscription = newSub;
    session.closeUpgrade();
    await session.refresh();
  }}
/>
