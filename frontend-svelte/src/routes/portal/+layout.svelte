<script>
  // Jemaah self-service portal shell. Guards access: only an authenticated
  // user with role "jamaah" may enter; staff → /app, agents → /agency, the
  // logged-out → /login.
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { LayoutDashboard, FileText, BadgeCheck, Wallet, UserCircle, LogOut, Menu } from "lucide-svelte";
  import { session, logoutAndRedirect } from "$lib/stores/session.svelte.js";

  let { children } = $props();
  let ready = $state(false);
  let sidebarOpen = $state(false);

  const NAV = [
    { href: "/portal", label: "Beranda", icon: LayoutDashboard },
    { href: "/portal/dokumen", label: "Dokumen", icon: FileText },
    { href: "/portal/visa", label: "Visa", icon: BadgeCheck },
    { href: "/portal/profil", label: "Profil", icon: UserCircle },
  ];
  let path = $derived($page.url.pathname);

  onMount(async () => {
    session.hydrateFromCache();
    const ok = await session.loadSession();
    if (!ok) { goto("/login", { replaceState: true }); return; }
    const role = session.user?.role;
    if (role !== "jamaah") { goto(role === "agent" ? "/agency" : "/app", { replaceState: true }); return; }
    ready = true;
  });
</script>

{#if !ready}
  <div class="flex min-h-screen items-center justify-center text-slate-500">Memuat…</div>
{:else}
  <div class="portal-shell">
    <aside class="portal-side" class:portal-side-open={sidebarOpen}>
      <div class="portal-brand">
        <span class="portal-logo">S</span>
        <div>
          <p class="text-sm font-extrabold" style="color:var(--c-ink)">Portal Jemaah</p>
          <p class="truncate text-[11px]" style="color:var(--c-faint)">{session.user?.name || ""}</p>
        </div>
      </div>
      <nav class="portal-nav">
        {#each NAV as item}
          <a href={item.href} class="portal-link" class:portal-link-on={item.href === "/portal" ? path === "/portal" : path.startsWith(item.href)} onclick={() => (sidebarOpen = false)}>
            <item.icon class="h-4 w-4" />
            {item.label}
          </a>
        {/each}
      </nav>
      <button type="button" class="portal-logout" onclick={logoutAndRedirect}><LogOut class="h-4 w-4" /> Keluar</button>
    </aside>

    <div class="portal-main">
      <header class="portal-topbar">
        <button type="button" class="portal-menu" onclick={() => (sidebarOpen = !sidebarOpen)} aria-label="Menu"><Menu class="h-5 w-5" /></button>
        <span class="text-sm font-bold" style="color:var(--c-ink)">Portal Jemaah</span>
      </header>
      <main class="portal-content">{@render children()}</main>
    </div>
  </div>
{/if}

<style>
  .portal-shell { display: flex; min-height: 100vh; background: var(--c-bg); }
  .portal-side { display: flex; flex-direction: column; gap: 8px; width: 240px; flex-shrink: 0; padding: 16px 12px; background: var(--c-surface); border-right: 1px solid var(--c-line); }
  .portal-brand { display: flex; align-items: center; gap: 10px; padding: 8px 8px 16px; }
  .portal-logo { display: flex; align-items: center; justify-content: center; width: 36px; height: 36px; border-radius: 10px; background: var(--c-primary); color: #fff; font-weight: 800; }
  .portal-nav { display: flex; flex-direction: column; gap: 2px; flex: 1; }
  .portal-link { display: flex; align-items: center; gap: 10px; padding: 10px 12px; border-radius: var(--radius); font-size: 13.5px; font-weight: 600; color: var(--c-muted); transition: all 0.15s; }
  .portal-link:hover { background: var(--c-bg-2); color: var(--c-ink); }
  .portal-link-on { background: var(--c-primary-tint); color: var(--c-primary); }
  .portal-logout { display: flex; align-items: center; gap: 10px; padding: 10px 12px; border-radius: var(--radius); font-size: 13.5px; font-weight: 600; color: var(--c-danger); transition: background 0.15s; }
  .portal-logout:hover { background: var(--c-danger-soft); }
  .portal-main { flex: 1; min-width: 0; display: flex; flex-direction: column; }
  .portal-topbar { display: none; align-items: center; gap: 12px; padding: 12px 16px; background: var(--c-surface); border-bottom: 1px solid var(--c-line); }
  .portal-menu { color: var(--c-muted); }
  .portal-content { flex: 1; min-width: 0; padding: 24px; overflow-y: auto; }
  @media (max-width: 768px) {
    .portal-side { position: fixed; inset: 0 auto 0 0; z-index: 40; transform: translateX(-100%); transition: transform 0.2s; }
    .portal-side-open { transform: translateX(0); }
    .portal-topbar { display: flex; }
  }
</style>
