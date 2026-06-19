<script>
  // Topbar — matches the Claude design app.jsx Topbar: page title + search + bell + user.
  import { Menu, Search, Bell, ChevronDown } from "lucide-svelte";
  import Avatar from "./Avatar.svelte";
  import HelpHint from "$lib/components/help/HelpHint.svelte";

  let { currentPage = "dashboard", user = null, onMenu = () => {}, onProfile = () => {} } = $props();

  const TITLES = {
    dashboard: "Dashboard", jamaah: "Data Jamaah", grup: "Grup", scanner: "AI Scanner",
    packages: "Paket", crm: "CRM", agents: "Agen & Mitra", invoices: "Invoice",
    rooming: "Rooming List", contracts: "Kontrak", itinerary: "Itinerary", manifest: "Manifest",
    finance: "Keuangan", vendors: "Vendor", payroll: "Payroll", inventory: "Inventaris",
    stock: "Inventaris", cancellation: "Pembatalan", profile: "Profil Saya", team: "Tim & Organisasi",
    documents: "Dokumen", analytics: "Analytics", export: "Export Laporan",
    bantuan: "Pusat Bantuan",
  };

  let title = $derived(TITLES[currentPage] || "Dashboard");
  let name = $derived(user?.name || "Pengguna");
  let role = $derived(
    (user?.role ? user.role.charAt(0).toUpperCase() + user.role.slice(1) : "Staff") +
      (user?.org_name ? ` · ${user.org_name}` : ""),
  );
</script>

<header class="tb">
  <button class="tb-menu" aria-label="Menu" onclick={() => onMenu()}><Menu size={22} /></button>

  <div style="flex:1;min-width:0">
    <div class="tb-title">{title}</div>
  </div>

  <div class="tb-search">
    <Search size={17} class="tb-search-ic" />
    <input placeholder="Cari jamaah, paket, invoice…" />
  </div>

  <HelpHint area="app" />

  <button class="tb-bell" aria-label="Notifikasi">
    <Bell size={20} />
    <span class="tb-dot"></span>
  </button>

  <div class="tb-divider"></div>

  <button class="tb-user" title="Buka profil" onclick={() => onProfile()}>
    <Avatar {name} size={38} />
    <div class="tb-user-meta">
      <div style="font-size:13.5px;font-weight:700;color:var(--c-ink)">{name}</div>
      <div style="font-size:11.5px;color:var(--c-faint)">{role}</div>
    </div>
    <ChevronDown size={16} class="tb-chev" />
  </button>
</header>

<style>
  .tb {
    height: 66px; flex-shrink: 0; background: var(--c-surface);
    border-bottom: 1px solid var(--c-line); display: flex; align-items: center;
    gap: 16px; padding: 0 26px; position: sticky; top: 0; z-index: 50;
  }
  .tb-menu { display: none; width: 40px; height: 40px; margin-left: -6px; border-radius: var(--radius); align-items: center; justify-content: center; color: var(--c-ink); background: none; border: none; cursor: pointer; }
  .tb-title { font-size: 16.5px; font-weight: 800; color: var(--c-ink); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .tb-search { position: relative; width: 300px; }
  :global(.tb-search-ic) { position: absolute; left: 13px; top: 50%; transform: translateY(-50%); color: var(--c-faint); }
  .tb-search input {
    width: 100%; padding: 9px 14px 9px 38px; font-size: 13.5px; background: var(--c-bg);
    border: 1px solid var(--c-line); border-radius: var(--radius); outline: none; color: var(--c-ink);
  }
  .tb-search input:focus { border-color: var(--c-primary); background: var(--c-surface); }
  .tb-bell { position: relative; width: 40px; height: 40px; border-radius: var(--radius); display: flex; align-items: center; justify-content: center; color: var(--c-muted); background: none; border: none; cursor: pointer; }
  .tb-dot { position: absolute; top: 7px; right: 8px; width: 8px; height: 8px; border-radius: 999px; background: var(--c-danger); border: 2px solid var(--c-surface); }
  .tb-divider { width: 1px; height: 28px; background: var(--c-line); }
  .tb-user { display: flex; align-items: center; gap: 10px; cursor: pointer; background: transparent; padding: 4px 6px; border-radius: var(--radius); border: none; transition: background 0.14s; }
  .tb-user:hover { background: var(--c-bg); }
  :global(.tb-chev) { color: var(--c-faint); }

  @media (max-width: 900px) {
    .tb { gap: 10px; padding: 0 16px; }
    .tb-menu { display: flex; }
    .tb-search { display: none; }
    .tb-user-meta, :global(.tb-chev) { display: none; }
  }
</style>
