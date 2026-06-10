<script>
  // Full-screen super-admin panel (no sidebar). Verifies the session and that the
  // user is a super admin before rendering; otherwise bounces to /login or /app.
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import Seo from "$lib/components/Seo.svelte";
  import SuperAdminDashboard from "$lib/pages/SuperAdminDashboard.svelte";
  import { session } from "$lib/stores/session.svelte.js";

  let checking = $state(true);
  let allowed = $state(false);

  onMount(async () => {
    session.hydrateFromCache();
    const ok = await session.loadSession();
    if (!ok) {
      goto("/login", { replaceState: true });
      return;
    }
    if (!session.isSuperAdmin) {
      goto("/app", { replaceState: true });
      return;
    }
    allowed = true;
    checking = false;
  });
</script>

<Seo title="Super Admin - Suluk" description="Panel super admin." path="/super-admin" robots="noindex,nofollow" />

{#if checking}
  <div class="min-h-screen flex items-center justify-center">
    <div class="text-center">
      <div class="animate-spin rounded-full h-12 w-12 border-4 border-emerald-500 border-t-transparent mx-auto"></div>
      <p class="mt-4 text-slate-600">Verifying access...</p>
    </div>
  </div>
{:else if allowed}
  <SuperAdminDashboard />
{/if}
