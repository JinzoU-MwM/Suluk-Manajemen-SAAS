<script>
  // Shared public navbar for the homepage + all marketing/landing pages. Uses
  // real route links (crawlable, consistent, good internal linking for SEO)
  // instead of the two divergent navbars that existed before the SvelteKit move.
  import { page } from "$app/stores";

  let mobileOpen = $state(false);

  const FEATURES = [
    { href: "/fitur/invoice-umrah", label: "Invoice Otomatis" },
    { href: "/fitur/crm-jamaah", label: "CRM Jamaah" },
    { href: "/fitur/laporan-keuangan", label: "Laporan Keuangan" },
    { href: "/fitur/e-kontrak", label: "E-Kontrak Digital" },
    { href: "/fitur/penggajian", label: "Penggajian" },
  ];

  let path = $derived($page.url.pathname);
  let featuresActive = $derived(path.startsWith("/fitur/"));
  const close = () => (mobileOpen = false);
</script>

<nav class="ln">
  <div class="ln-inner">
    <a class="ln-brand" href="/" onclick={close}>
      <span class="ln-mark"><img src="/brand/suluk-mark.png" alt="Suluk" /></span>
      <span class="ln-brand-text">
        <span class="ln-brand-name">Suluk</span>
        <span class="ln-brand-sub">ERP FOR TRAVEL</span>
      </span>
    </a>

    <div class="ln-links">
      <a href="/software-travel-umrah" class:active={path === "/software-travel-umrah"}>Software</a>

      <div class="ln-dd">
        <button type="button" class="ln-dd-btn" class:active={featuresActive} aria-haspopup="true">
          Fitur
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9" /></svg>
        </button>
        <div class="ln-dd-menu">
          {#each FEATURES as f}
            <a href={f.href} class:active={path === f.href}>{f.label}</a>
          {/each}
        </div>
      </div>

      <a href="/#harga">Harga</a>
      <a href="/unduh" class:active={path === "/unduh"}>Aplikasi</a>
      <a href="/panduan" class:active={path.startsWith("/panduan")}>Panduan</a>
      <a href="/tentang" class:active={path === "/tentang"}>Tentang</a>
    </div>

    <div class="ln-actions">
      <a class="ln-login" href="/login">Masuk</a>
      <a class="ln-cta" href="/daftar">Coba Gratis</a>
      <button class="ln-burger" aria-label="Menu" aria-expanded={mobileOpen} onclick={() => (mobileOpen = !mobileOpen)}>
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><line x1="4" y1="7" x2="20" y2="7" /><line x1="4" y1="12" x2="20" y2="12" /><line x1="4" y1="17" x2="20" y2="17" /></svg>
      </button>
    </div>
  </div>

  <div class="ln-mobile" class:open={mobileOpen}>
    <a href="/software-travel-umrah" onclick={close}>Software</a>
    <div class="ln-mobile-group">
      <span class="ln-mobile-label">Fitur</span>
      {#each FEATURES as f}<a href={f.href} onclick={close}>{f.label}</a>{/each}
    </div>
    <a href="/#harga" onclick={close}>Harga</a>
    <a href="/unduh" onclick={close}>Aplikasi</a>
    <a href="/panduan" onclick={close}>Panduan</a>
    <a href="/tentang" onclick={close}>Tentang</a>
    <a href="/kontak" onclick={close}>Kontak</a>
    <a class="ln-login-m" href="/login" onclick={close}>Masuk</a>
    <a class="ln-cta ln-cta-block" href="/daftar" onclick={close}>Coba Gratis 14 Hari</a>
  </div>
</nav>

<style>
  .ln {
    position: sticky;
    top: 0;
    z-index: 100;
    background: rgba(255, 255, 255, 0.82);
    backdrop-filter: saturate(180%) blur(14px);
    border-bottom: 1px solid var(--c-line, #e6e9e7);
  }
  .ln-inner {
    display: flex;
    align-items: center;
    gap: 32px;
    height: 70px;
    max-width: 1180px;
    margin: 0 auto;
    padding: 0 24px;
  }
  @media (max-width: 760px) {
    .ln-inner { padding: 0 18px; }
  }
  .ln-brand { display: flex; align-items: center; gap: 11px; flex-shrink: 0; text-decoration: none; }
  .ln-mark { display: flex; align-items: center; }
  .ln-mark img { height: 42px; width: auto; display: block; }
  .ln-brand-text { display: flex; flex-direction: column; }
  .ln-brand-name { font-family: var(--font-display, "Playfair Display", serif); font-size: 23px; font-weight: 800; line-height: 1; color: var(--c-ink, #14201b); }
  .ln-brand-sub { font-size: 9.5px; font-weight: 700; letter-spacing: 0.17em; color: var(--c-accent, #c79a3e); margin-top: 4px; }

  .ln-links { display: flex; align-items: center; gap: 28px; flex: 1; }
  .ln-links > a,
  .ln-dd-btn {
    font-size: 14.5px;
    font-weight: 600;
    color: var(--c-ink-soft, #4a5a52);
    text-decoration: none;
    transition: color 0.15s;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    font-family: inherit;
    display: inline-flex;
    align-items: center;
    gap: 5px;
  }
  .ln-links > a:hover,
  .ln-dd-btn:hover,
  .ln-links > a.active,
  .ln-dd-btn.active { color: var(--c-primary, #0f7a5a); }
  .ln-links > a.active,
  .ln-dd-btn.active { font-weight: 700; }

  /* Dropdown */
  .ln-dd { position: relative; }
  .ln-dd-menu {
    position: absolute;
    top: calc(100% + 14px);
    left: -14px;
    min-width: 210px;
    background: var(--c-surface, #fff);
    border: 1px solid var(--c-line, #e6e9e7);
    border-radius: 14px;
    box-shadow: var(--shadow-lg, 0 18px 40px rgba(15, 61, 46, 0.16));
    padding: 8px;
    opacity: 0;
    visibility: hidden;
    transform: translateY(-6px);
    transition: opacity 0.16s, transform 0.16s, visibility 0.16s;
  }
  .ln-dd:hover .ln-dd-menu,
  .ln-dd:focus-within .ln-dd-menu { opacity: 1; visibility: visible; transform: translateY(0); }
  .ln-dd-menu a {
    display: block;
    padding: 10px 12px;
    border-radius: 9px;
    font-size: 14px;
    font-weight: 600;
    color: var(--c-ink, #14201b);
    text-decoration: none;
  }
  .ln-dd-menu a:hover { background: var(--c-bg, #f5f7f6); color: var(--c-primary, #0f7a5a); }
  .ln-dd-menu a.active { color: var(--c-primary, #0f7a5a); }

  .ln-actions { display: flex; align-items: center; gap: 12px; }
  .ln-login {
    font-size: 14.5px;
    font-weight: 700;
    color: var(--c-ink, #14201b);
    padding: 10px 8px;
    text-decoration: none;
  }
  .ln-login:hover { color: var(--c-primary, #0f7a5a); }
  .ln-cta {
    font-size: 14.5px;
    font-weight: 700;
    color: #fff;
    background: var(--c-primary, #0f7a5a);
    padding: 10px 18px;
    border-radius: 10px;
    text-decoration: none;
    transition: background 0.15s, transform 0.1s;
  }
  .ln-cta:hover { background: var(--c-primary-dark, #0c5f46); }
  .ln-cta:active { transform: translateY(1px); }

  .ln-burger {
    display: none;
    width: 44px;
    height: 44px;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    color: var(--c-ink, #14201b);
    background: none;
    border: none;
    cursor: pointer;
  }
  .ln-burger:hover { background: var(--c-bg, #f5f7f6); }

  .ln-mobile { display: none; }

  @media (max-width: 900px) {
    .ln-links { display: none; }
    .ln-actions .ln-login,
    .ln-actions .ln-cta { display: none; }
    .ln-burger { display: flex; }
    .ln-inner { justify-content: space-between; gap: 12px; }
    .ln-mobile {
      position: fixed;
      inset: 70px 0 auto;
      background: var(--c-surface, #fff);
      border-bottom: 1px solid var(--c-line, #e6e9e7);
      box-shadow: var(--shadow-lg, 0 18px 40px rgba(15, 61, 46, 0.16));
      padding: 14px 18px 22px;
      z-index: 99;
      max-height: calc(100vh - 70px);
      overflow-y: auto;
    }
    .ln-mobile.open { display: block; }
    .ln-mobile > a,
    .ln-mobile-group > a {
      display: block;
      padding: 12px 8px;
      font-size: 16px;
      font-weight: 600;
      color: var(--c-ink, #14201b);
      text-decoration: none;
      border-bottom: 1px solid var(--c-line-soft, #eef1ef);
    }
    .ln-mobile-group { padding: 6px 0; }
    .ln-mobile-label {
      display: block;
      padding: 12px 8px 4px;
      font-size: 11px;
      font-weight: 800;
      letter-spacing: 0.12em;
      text-transform: uppercase;
      color: var(--c-accent, #c79a3e);
    }
    .ln-mobile-group > a { padding-left: 20px; font-size: 15px; }
    .ln-login-m { font-weight: 700; }
    .ln-cta-block {
      display: block;
      width: 100%;
      text-align: center;
      margin-top: 16px;
      border-bottom: none;
    }
  }
</style>
