<script>
  // Sidebar — matches the Claude design app shell (app.jsx Sidebar): deep-green
  // rail, S-mark logo + "ERP FOR TRAVEL", grouped nav, gold active accent bar,
  // AI badge, and the "Paket Pro" upgrade card.
  import {
    LayoutDashboard, Users, UserCheck, ScanLine, Package, Briefcase,
    Building2, Receipt, BedDouble, FileSignature, Map, ClipboardList,
    Wallet, Truck, Banknote, Boxes, XCircle, Sparkles, X, BookOpen,
    PiggyBank, Calculator,
  } from "lucide-svelte";

  let {
    currentPage = "dashboard",
    onPageChange,
    user = null,
    isPro = false,
    onUpgrade = () => {},
    open = false, // mobile drawer
    onClose = () => {},
  } = $props();

  const NAV = [
    { group: "Utama", items: [
      { id: "dashboard", label: "Dashboard", icon: LayoutDashboard },
      { id: "jamaah", label: "Data Jamaah", icon: Users },
      { id: "grup", label: "Grup", icon: UserCheck },
      { id: "scanner", label: "AI Scanner", icon: ScanLine, badge: "AI" },
    ] },
    { group: "Penjualan", items: [
      { id: "packages", label: "Paket", icon: Package },
      { id: "crm", label: "CRM", icon: Briefcase },
      { id: "agents", label: "Agen & Mitra", icon: Building2 },
      { id: "invoices", label: "Invoice", icon: Receipt },
    ] },
    { group: "Operasional", items: [
      { id: "rooming", label: "Rooming List", icon: BedDouble },
      { id: "contracts", label: "Kontrak", icon: FileSignature },
      { id: "itinerary", label: "Itinerary", icon: Map },
      { id: "manifest", label: "Manifest", icon: ClipboardList },
    ] },
    { group: "Keuangan", items: [
      { id: "finance", label: "Keuangan", icon: Wallet },
      { id: "kasir", label: "Kasir (POS)", icon: Calculator },
      { id: "tabungan", label: "Tabungan", icon: PiggyBank },
      { id: "akuntansi", label: "Akuntansi", icon: BookOpen },
      { id: "vendors", label: "Vendor", icon: Truck },
      { id: "payroll", label: "Payroll", icon: Banknote },
      { id: "inventory", label: "Inventaris", icon: Boxes },
      { id: "cancellation", label: "Pembatalan", icon: XCircle },
    ] },
  ];

  function go(id) {
    onPageChange?.(id);
    onClose();
  }
</script>

{#if open}
  <button class="sb-backdrop" aria-label="Tutup menu" onclick={() => onClose()}></button>
{/if}

<aside class="sb {open ? 'sb-open' : ''}">
  <div class="sb-logo">
    <img src="/brand/suluk-mark-white.png" alt="Suluk" style="height:42px;width:auto;display:block" />
    <div>
      <div class="font-serif" style="font-size:22px;font-weight:800;color:#fff;line-height:1;letter-spacing:.01em">Suluk</div>
      <div style="font-size:9.5px;font-weight:700;color:var(--c-accent);letter-spacing:.16em;margin-top:4px">ERP FOR TRAVEL</div>
    </div>
    <button class="sb-close" aria-label="Tutup" onclick={() => onClose()}><X size={20} /></button>
  </div>

  <nav class="sb-nav">
    {#each NAV as grp}
      <div style="margin-bottom:18px">
        <div class="sb-group">{grp.group}</div>
        {#each grp.items as item}
          {@const active = currentPage === item.id}
          {@const Icon = item.icon}
          <button class="sb-item {active ? 'active' : ''}" onclick={() => go(item.id)}>
            {#if active}<span class="sb-accent"></span>{/if}
            <Icon size={19} />
            <span style="flex:1;text-align:left">{item.label}</span>
            {#if item.badge}<span class="sb-badge">{item.badge}</span>{/if}
          </button>
        {/each}
      </div>
    {/each}
  </nav>

  {#if !isPro}
    <div class="sb-foot">
      <div class="sb-pro">
        <Sparkles size={70} class="sb-pro-deco" />
        <div style="font-size:13.5px;font-weight:800;color:#fff;position:relative">Paket Pro</div>
        <div style="font-size:11.5px;color:rgba(255,255,255,.82);margin-top:4px;line-height:1.4;position:relative">Buka modul tak terbatas &amp; laporan lanjutan.</div>
        <button class="sb-pro-btn" onclick={() => onUpgrade()}>Upgrade Sekarang</button>
      </div>
    </div>
  {/if}
</aside>

<style>
  .sb {
    width: 256px;
    flex-shrink: 0;
    background: var(--c-sidebar-bg);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    height: 100vh;
    position: sticky;
    top: 0;
  }
  .sb-logo { padding: 22px 20px 18px; display: flex; align-items: center; gap: 12px; }
  .sb-close { display: none; margin-left: auto; color: rgba(255, 255, 255, 0.6); background: none; border: none; cursor: pointer; }
  .sb-nav { flex: 1; overflow-y: auto; padding: 4px 12px 12px; }
  .sb-group {
    font-size: 10.5px; font-weight: 700; letter-spacing: 0.13em;
    color: rgba(255, 255, 255, 0.42); padding: 0 12px 8px; text-transform: uppercase;
  }
  .sb-item {
    width: 100%; display: flex; align-items: center; gap: 12px; padding: 9px 12px;
    border-radius: var(--radius); margin-bottom: 2px; color: #cfe0d9;
    font-weight: 500; font-size: 14px; transition: background 0.14s, color 0.14s;
    position: relative; background: none; border: none; cursor: pointer;
  }
  .sb-item:hover { background: rgba(255, 255, 255, 0.05); color: #fff; }
  .sb-item.active { background: rgba(255, 255, 255, 0.08); color: #fff; font-weight: 700; }
  .sb-accent {
    position: absolute; left: -12px; top: 50%; transform: translateY(-50%);
    width: 3px; height: 20px; background: var(--c-accent); border-radius: 999px;
  }
  .sb-badge {
    font-size: 9.5px; font-weight: 800; letter-spacing: 0.05em; color: #1c1814;
    background: var(--c-accent); padding: 2px 6px; border-radius: 5px;
  }
  .sb-foot { padding: 14px; }
  .sb-pro {
    background: linear-gradient(135deg, var(--c-primary), var(--c-primary-deep));
    border-radius: var(--radius-lg); padding: 16px; position: relative; overflow: hidden;
  }
  :global(.sb-pro-deco) { position: absolute; right: -14px; bottom: -14px; color: rgba(255, 255, 255, 0.12); }
  .sb-pro-btn {
    margin-top: 12px; width: 100%; padding: 8px; font-size: 12.5px; font-weight: 700;
    color: var(--c-primary-deep); background: #fff; border: none; border-radius: var(--radius-sm);
    position: relative; cursor: pointer;
  }
  .sb-backdrop { display: none; }

  @media (max-width: 900px) {
    .sb {
      position: fixed; top: 0; left: 0; z-index: 120; height: 100vh;
      transform: translateX(-100%); transition: transform 0.25s ease;
      box-shadow: var(--shadow-lg);
    }
    .sb.sb-open { transform: translateX(0); }
    .sb-close { display: flex; }
    .sb-backdrop {
      display: block; position: fixed; inset: 0; z-index: 110;
      background: rgba(16, 33, 28, 0.4); border: none; cursor: pointer;
    }
  }
</style>
