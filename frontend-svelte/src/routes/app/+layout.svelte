<script>
  // Authenticated app shell — sidebar + topbar + content, plus the global upgrade
  // modal and support chat. Replaces App.svelte's "showSidebar" branch. Guards
  // access: verifies the session on mount and bounces to /login if it fails.
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Sidebar from "$lib/components/Sidebar.svelte";
  import Topbar from "$lib/components/Topbar.svelte";
  import UpgradeModal from "$lib/components/UpgradeModal.svelte";
  import SupportChatBubble from "$lib/components/SupportChatBubble.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";

  let { children } = $props();
  let sidebarOpen = $state(false);
  let ready = $state(false);

  // Derive the legacy page-key from the path for Sidebar/Topbar highlighting.
  function pathToPageKey(path) {
    if (path === "/app" || path === "/app/") return "dashboard";
    const seg = path.replace(/^\/app\/?/, "").split("/")[0];
    return seg || "dashboard";
  }
  let currentPage = $derived(pathToPageKey($page.url.pathname));

  onMount(async () => {
    session.hydrateFromCache();
    const ok = await session.loadSession();
    if (!ok) {
      goto("/login", { replaceState: true });
      return;
    }
    // External agents belong in the B2B portal, not the staff app.
    if (session.user?.role === "agent") {
      goto("/agency", { replaceState: true });
      return;
    }
    // Jemaah users belong in the self-service portal.
    if (session.user?.role === "jamaah") {
      goto("/portal", { replaceState: true });
      return;
    }
    // Return from a Pakasir payment redirect: re-fetch the (possibly upgraded)
    // subscription and clean the URL.
    const params = new URLSearchParams(window.location.search);
    if (params.get("payment") === "success") {
      await session.refresh();
      history.replaceState({}, "", window.location.pathname);
    }
    ready = true;
  });
</script>

{#if !ready}
  <div class="min-h-screen flex items-center justify-center text-slate-500">Memuat…</div>
{:else}
  <div class="suluk-shell">
    <Sidebar
      {currentPage}
      onPageChange={navigate}
      user={session.user}
      isPro={session.isPro}
      open={sidebarOpen}
      onClose={() => (sidebarOpen = false)}
      onUpgrade={() => session.openUpgrade()}
    />

    <div class="suluk-main">
      <Topbar
        {currentPage}
        user={session.user}
        onMenu={() => (sidebarOpen = !sidebarOpen)}
        onProfile={() => navigate("profile")}
      />
      <div class="suluk-content">
        {@render children()}
      </div>
    </div>
  </div>
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
<SupportChatBubble />
