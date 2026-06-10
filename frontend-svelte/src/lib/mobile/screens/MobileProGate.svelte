<script>
  import { Crown, Check, Monitor, LogOut } from "lucide-svelte";
  import "../mobile.css";

  let { user = null, plan = "", onUpgrade = () => {}, onExit = () => {}, onLogout = () => {} } = $props();

  const perks = [
    "Kelola jamaah & pembayaran dari ponsel",
    "AI Scanner dokumen pakai kamera",
    "Approval refund & diskon di mana saja",
    "CRM, paket, keuangan — lengkap",
  ];
</script>

<div class="pg-wrap">
  <div class="pg-hero">
    <div class="pg-crown"><Crown size={30} /></div>
    <div class="pg-title">Aplikasi Mobile Suluk</div>
    <div class="pg-sub">Fitur eksklusif paket <b>Pro</b> ke atas</div>
  </div>

  <div class="pg-sheet m-slide">
    <div class="pg-plan">Paket Anda saat ini: <b>{plan ? plan.charAt(0).toUpperCase() + plan.slice(1) : "Gratis"}</b></div>
    <p class="pg-lead">Upgrade ke Pro untuk membuka aplikasi staf di ponsel — kelola travel Anda di mana saja.</p>

    <div class="pg-perks">
      {#each perks as p}
        <div class="pg-perk"><span class="pg-check"><Check size={13} /></span>{p}</div>
      {/each}
    </div>

    <button type="button" class="m-btn m-btn-primary" onclick={() => onUpgrade()}><Crown size={18} />Upgrade ke Pro</button>
    <button type="button" class="m-btn m-btn-ghost" style="margin-top:10px" onclick={() => onExit()}><Monitor size={18} />Buka Versi Desktop</button>
    <button type="button" class="pg-logout" onclick={() => onLogout()}><LogOut size={15} />Keluar</button>
  </div>
</div>

<style>
  .pg-wrap {
    position: fixed; inset: 0; height: 100dvh; z-index: 50; display: flex; flex-direction: column;
    overflow-y: auto; background: linear-gradient(165deg, var(--c-primary-deep), var(--c-primary));
    font-family: var(--font-ui); color: var(--c-ink);
  }
  .pg-hero { flex-shrink: 0; text-align: center; color: #fff; padding: calc(env(safe-area-inset-top) + 64px) 24px 34px; }
  .pg-crown {
    width: 64px; height: 64px; border-radius: 20px; margin: 0 auto 16px;
    background: rgba(255,255,255,.14); display: flex; align-items: center; justify-content: center; color: var(--c-accent);
  }
  .pg-title { font-family: var(--font-display, Georgia, serif); font-size: 25px; font-weight: 800; letter-spacing: -0.01em; }
  .pg-sub { font-size: 13.5px; opacity: .9; margin-top: 5px; }
  .pg-sheet {
    flex: 1; background: var(--c-surface); border-radius: 26px 26px 0 0;
    padding: 26px 22px calc(env(safe-area-inset-bottom) + 28px); box-shadow: 0 -8px 30px rgba(0,0,0,.12); min-height: 50%;
  }
  .pg-plan { font-size: 13px; color: var(--c-muted); background: var(--c-bg-2); padding: 9px 14px; border-radius: 10px; text-align: center; }
  .pg-lead { font-size: 14.5px; color: var(--c-ink-soft); line-height: 1.6; margin: 16px 0 18px; }
  .pg-perks { display: flex; flex-direction: column; gap: 11px; margin-bottom: 22px; }
  .pg-perk { display: flex; align-items: center; gap: 11px; font-size: 14px; font-weight: 500; color: var(--c-ink); }
  .pg-check {
    flex-shrink: 0; width: 22px; height: 22px; border-radius: 50%; background: var(--c-primary-soft); color: var(--c-primary);
    display: flex; align-items: center; justify-content: center;
  }
  .pg-logout {
    width: 100%; margin-top: 16px; display: flex; align-items: center; justify-content: center; gap: 6px;
    font-size: 14px; font-weight: 600; color: var(--c-danger); background: none; border: none; padding: 8px;
  }
</style>
