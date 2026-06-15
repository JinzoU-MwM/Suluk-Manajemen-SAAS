<script>
  // B2B agent portal shell — its own minimal sidebar, separate from the staff
  // app. Guards access: only an authenticated user with role "agent" may enter;
  // staff are bounced to /app, the logged-out to /login.
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { LayoutDashboard, Wallet, Network, UserCircle, LogOut, Menu, Users } from "lucide-svelte";
  import { session, logoutAndRedirect } from "$lib/stores/session.svelte.js";

  let { children } = $props();
  let ready = $state(false);
  let sidebarOpen = $state(false);

  const NAV = [
    { href: "/agency", label: "Dashboard", icon: LayoutDashboard },
    { href: "/agency/leads", label: "Lead Saya", icon: Users },
    { href: "/agency/komisi", label: "Komisi Saya", icon: Wallet },
    { href: "/agency/jaringan", label: "Jaringan", icon: Network },
    { href: "/agency/profil", label: "Profil", icon: UserCircle },
  ];

  let path = $derived($page.url.pathname);

  onMount(async () => {
    session.hydrateFromCache();
    const ok = await session.loadSession();
    if (!ok) {
      goto("/login", { replaceState: true });
      return;
    }
    if (session.user?.role !== "agent") {
      goto("/app", { replaceState: true });
      return;
    }
    ready = true;
  });
</script>

{#if !ready}
  <div class="flex min-h-screen items-center justify-center text-slate-500">Memuat…</div>
{:else}
  <div class="agency-shell">
    <aside class="agency-side" class:agency-side-open={sidebarOpen}>
      <div class="agency-brand">
        <span class="agency-logo">S</span>
        <div>
          <p class="text-sm font-extrabold" style="color:var(--c-ink)">Portal Agen</p>
          <p class="truncate text-[11px]" style="color:var(--c-faint)">{session.user?.name || ""}</p>
        </div>
      </div>
      <nav class="agency-nav">
        {#each NAV as item}
          <a
            href={item.href}
            class="agency-link"
            class:agency-link-on={item.href === "/agency" ? path === "/agency" : path.startsWith(item.href)}
            onclick={() => (sidebarOpen = false)}
          >
            <item.icon class="h-4 w-4" />
            {item.label}
          </a>
        {/each}
      </nav>
      <button type="button" class="agency-logout" onclick={logoutAndRedirect}>
        <LogOut class="h-4 w-4" /> Keluar
      </button>
    </aside>

    <div class="agency-main">
      <header class="agency-topbar">
        <button type="button" class="agency-menu" onclick={() => (sidebarOpen = !sidebarOpen)} aria-label="Menu"><Menu class="h-5 w-5" /></button>
        <span class="text-sm font-bold" style="color:var(--c-ink)">Portal Agen</span>
      </header>
      <main class="agency-content">
        {@render children()}
      </main>
    </div>
  </div>
{/if}

<style>
  .agency-shell { display: flex; min-height: 100vh; background: var(--c-bg); }
  .agency-side {
    display: flex; flex-direction: column; gap: 8px;
    width: 240px; flex-shrink: 0; padding: 16px 12px;
    background: var(--c-surface); border-right: 1px solid var(--c-line);
  }
  .agency-brand { display: flex; align-items: center; gap: 10px; padding: 8px 8px 16px; }
  .agency-logo {
    display: flex; align-items: center; justify-content: center;
    width: 36px; height: 36px; border-radius: 10px;
    background: var(--c-primary); color: #fff; font-weight: 800;
  }
  .agency-nav { display: flex; flex-direction: column; gap: 2px; flex: 1; }
  .agency-link {
    display: flex; align-items: center; gap: 10px;
    padding: 10px 12px; border-radius: var(--radius);
    font-size: 13.5px; font-weight: 600; color: var(--c-muted);
    transition: all 0.15s;
  }
  .agency-link:hover { background: var(--c-bg-2); color: var(--c-ink); }
  .agency-link-on { background: var(--c-primary-tint); color: var(--c-primary); }
  .agency-logout {
    display: flex; align-items: center; gap: 10px;
    padding: 10px 12px; border-radius: var(--radius);
    font-size: 13.5px; font-weight: 600; color: var(--c-danger);
    transition: background 0.15s;
  }
  .agency-logout:hover { background: var(--c-danger-soft); }
  .agency-main { flex: 1; min-width: 0; display: flex; flex-direction: column; }
  .agency-topbar {
    display: none; align-items: center; gap: 12px;
    padding: 12px 16px; background: var(--c-surface); border-bottom: 1px solid var(--c-line);
  }
  .agency-menu { color: var(--c-muted); }
  .agency-content { flex: 1; min-width: 0; padding: 24px; overflow-y: auto; }
  @media (max-width: 768px) {
    .agency-side {
      position: fixed; inset: 0 auto 0 0; z-index: 40;
      transform: translateX(-100%); transition: transform 0.2s;
    }
    .agency-side-open { transform: translateX(0); }
    .agency-topbar { display: flex; }
  }
</style>
