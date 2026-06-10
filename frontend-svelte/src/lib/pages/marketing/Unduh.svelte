<script>
  let { onNav = () => {}, onGoToApp = () => {} } = $props();
  const APK_URL = "/downloads/suluk.apk";
  const iosSteps = [
    "Buka suluk.site/#/app di Safari",
    "Ketuk tombol Bagikan (kotak dengan panah ke atas)",
    "Pilih “Tambah ke Layar Utama”",
    "Buka Suluk dari ikon di layar utama Anda",
  ];
</script>

<section class="gp-hero">
  <div class="gp-wrap">
    <p class="gp-badge">Aplikasi Mobile · Fitur Pro</p>
    <h1 class="gp-h1">Suluk dalam Genggaman</h1>
    <p class="gp-lead">
      Kelola jamaah, catat pembayaran, scan dokumen, dan setujui pengajuan langsung dari ponsel —
      di mana saja. Aplikasi staf Suluk tersedia untuk pelanggan paket <strong>Pro ke atas</strong>.
    </p>
  </div>
</section>

<section class="gp-sec" style="border-top:none">
  <div class="gp-wrap">
    <div class="gp-contact-grid">
      <!-- Android -->
      <div class="gp-contact-card">
        <div class="ic" style="background:var(--c-success-soft);color:var(--c-success)">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>
        </div>
        <h3>Android</h3>
        <p>Unduh & pasang langsung (APK). Android 5.0+.</p>
        <a class="dl-btn" href={APK_URL} download="Suluk.apk">⬇ Unduh APK Android</a>
        <p class="dl-note">Jika muncul peringatan, izinkan "Instal dari sumber ini".</p>
      </div>

      <!-- iOS / PWA -->
      <div class="gp-contact-card">
        <div class="ic" style="background:var(--c-info-soft);color:var(--c-info)">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2"/><path d="M9 6h6"/></svg>
        </div>
        <h3>iPhone & iPad</h3>
        <p>Pasang sebagai aplikasi langsung dari Safari:</p>
        <ol class="dl-steps">
          {#each iosSteps as s, i}<li><span>{i + 1}</span>{s}</li>{/each}
        </ol>
        <a class="dl-btn dl-btn-soft" href="/mobile">Buka Aplikasi Web →</a>
      </div>

      <!-- Stores -->
      <div class="gp-contact-card">
        <div class="ic" style="background:var(--c-accent-soft);color:var(--c-accent)">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
        </div>
        <h3>Play Store & App Store</h3>
        <p>Versi toko resmi sedang dalam proses peluncuran.</p>
        <span class="dl-badge">Segera hadir</span>
      </div>
    </div>

    <div class="dl-pro">
      <div class="dl-pro-tag">FITUR PRO</div>
      <p>Aplikasi mobile aktif untuk paket <strong>Pro, Bisnis, dan Enterprise</strong>. Akun gratis/starter dapat mencoba lewat upgrade.</p>
    </div>
  </div>
</section>

<section class="gp-cta">
  <div class="gp-wrap">
    <div class="gp-cta-box">
      <h2>Belum punya akses?</h2>
      <p>Upgrade ke Pro dan bawa operasional travel Anda ke mana saja.</p>
      <button class="gp-btn" onclick={() => onGoToApp()}>Mulai / Upgrade ke Pro →</button>
    </div>
  </div>
</section>

<style>
  .dl-btn {
    display: inline-flex; align-items: center; gap: 8px; margin-top: 12px;
    background: var(--c-primary); color: #fff; font-weight: 700; font-size: 14.5px;
    padding: 11px 18px; border-radius: 12px; text-decoration: none; transition: filter .15s;
  }
  .dl-btn:hover { filter: brightness(1.06); }
  .dl-btn-soft { background: var(--c-primary-soft); color: var(--c-primary-deep); }
  .dl-note { font-size: 12px; color: var(--c-faint); margin-top: 8px; }
  .dl-steps { list-style: none; padding: 0; margin: 12px 0 14px; display: flex; flex-direction: column; gap: 8px; }
  .dl-steps li { display: flex; align-items: flex-start; gap: 9px; font-size: 13.5px; color: var(--c-ink-soft); line-height: 1.4; }
  .dl-steps li span {
    flex-shrink: 0; width: 20px; height: 20px; border-radius: 50%; background: var(--c-info-soft);
    color: var(--c-info); font-size: 11px; font-weight: 800; display: flex; align-items: center; justify-content: center; margin-top: 1px;
  }
  .dl-badge {
    display: inline-block; margin-top: 12px; background: var(--c-accent-soft); color: var(--c-accent);
    font-weight: 700; font-size: 13px; padding: 8px 14px; border-radius: 999px;
  }
  .dl-pro {
    margin-top: 30px; background: var(--c-primary-tint); border: 1px solid var(--c-primary-soft);
    border-radius: 18px; padding: 22px 24px;
  }
  .dl-pro-tag { font-size: 11px; font-weight: 800; letter-spacing: .12em; color: var(--c-accent); margin-bottom: 6px; }
  .dl-pro p { margin: 0; font-size: 15px; color: var(--c-ink-soft); line-height: 1.6; }
</style>
