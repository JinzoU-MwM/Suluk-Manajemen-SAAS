<script>
  import { Lock, Crown, Check, Monitor, LogOut } from "lucide-svelte";
  import "../mobile.css";

  let { user = null, plan = "", onUpgrade = () => {}, onExit = () => {}, onLogout = () => {} } = $props();

  const perks = [
    "Kelola jamaah & pembayaran dari ponsel",
    "AI Scanner dokumen pakai kamera",
    "Approval refund & diskon di mana saja",
  ];
  let planLabel = $derived(plan ? plan.charAt(0).toUpperCase() + plan.slice(1) : "Gratis");
</script>

<!-- Locked-access modal: shown after a non-Pro user logs into the app -->
<div class="pgm-backdrop">
  <div class="pgm-card m-pop">
    <div class="pgm-badge"><Lock size={26} /></div>
    <div class="pgm-title">Akses Aplikasi Terkunci</div>
    <p class="pgm-sub">
      Paket <b>{planLabel}</b> Anda belum termasuk Aplikasi Mobile Suluk.
      Upgrade ke <b>Pro</b> untuk membuka akses penuh di ponsel.
    </p>

    <div class="pgm-perks">
      {#each perks as p}
        <div class="pgm-perk"><span class="pgm-check"><Check size={12} /></span>{p}</div>
      {/each}
    </div>

    <button type="button" class="m-btn m-btn-primary" onclick={() => onUpgrade()}><Crown size={18} />Upgrade ke Pro</button>
    <button type="button" class="m-btn m-btn-ghost" style="margin-top:10px" onclick={() => onExit()}><Monitor size={18} />Buka Versi Desktop</button>
    <button type="button" class="pgm-logout" onclick={() => onLogout()}><LogOut size={15} />Keluar</button>
  </div>
</div>

<style>
  .pgm-backdrop {
    position: fixed; inset: 0; z-index: 60; height: 100dvh;
    display: flex; align-items: center; justify-content: center; padding: 22px;
    background: linear-gradient(165deg, rgba(15,61,46,.96), rgba(27,127,90,.94));
    backdrop-filter: blur(6px); -webkit-backdrop-filter: blur(6px);
    font-family: var(--font-ui); color: var(--c-ink);
  }
  .pgm-card {
    width: 100%; max-width: 360px; background: var(--c-surface);
    border-radius: 24px; padding: 28px 24px calc(env(safe-area-inset-bottom) + 24px);
    text-align: center; box-shadow: 0 30px 70px -20px rgba(0,0,0,.5);
  }
  .pgm-badge {
    width: 60px; height: 60px; border-radius: 18px; margin: 0 auto 16px;
    background: var(--c-accent-soft); color: var(--c-accent);
    display: flex; align-items: center; justify-content: center;
  }
  .pgm-title { font-family: var(--font-display, Georgia, serif); font-size: 22px; font-weight: 800; letter-spacing: -0.01em; }
  .pgm-sub { font-size: 14px; color: var(--c-muted); line-height: 1.6; margin: 8px 0 18px; }
  .pgm-perks { display: flex; flex-direction: column; gap: 9px; text-align: left; margin-bottom: 22px; }
  .pgm-perk { display: flex; align-items: center; gap: 10px; font-size: 13.5px; font-weight: 500; color: var(--c-ink); }
  .pgm-check {
    flex-shrink: 0; width: 20px; height: 20px; border-radius: 50%; background: var(--c-primary-soft); color: var(--c-primary);
    display: flex; align-items: center; justify-content: center;
  }
  .pgm-logout {
    width: 100%; margin-top: 14px; display: flex; align-items: center; justify-content: center; gap: 6px;
    font-size: 14px; font-weight: 600; color: var(--c-danger); background: none; border: none; padding: 6px;
  }
</style>
