<script>
  let { onNav = () => {}, onGoToApp = () => {} } = $props();
</script>

<section class="gp-hero">
  <div class="gp-wrap">
    <p class="gp-badge">Dukungan</p>
    <h1 class="gp-h1">Hubungi Kami</h1>
    <p class="gp-lead">Ada pertanyaan atau butuh bantuan? Tim kami siap membantu travel Anda berkembang bersama Suluk.</p>
  </div>
</section>

<section class="gp-sec" style="border-top:none">
  <div class="gp-wrap">
    <div class="gp-contact-grid">
      <div class="gp-contact-card">
        <div class="ic" style="background:var(--c-success-soft);color:var(--c-success)">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 21l1.65-3.8a9 9 0 1 1 3.4 2.9L3 21"/><path d="M9 10a.5.5 0 0 0 1 0V9a.5.5 0 0 0-1 0v1a5 5 0 0 0 5 5h1a.5.5 0 0 0 0-1h-1a.5.5 0 0 0 0 1"/></svg>
        </div>
        <h3>WhatsApp Business</h3>
        <p>Senin–Jumat, 09.00–17.00 WIB</p>
        <a class="val" href="https://wa.me/6281100000000">+62 811-0000-0000</a>
      </div>
      <div class="gp-contact-card">
        <div class="ic" style="background:var(--c-info-soft);color:var(--c-info)">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="M22 7l-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>
        </div>
        <h3>Email</h3>
        <p>Respons kurang dari 24 jam</p>
        <a class="val" href="mailto:support@suluk.id">support@suluk.id</a>
      </div>
      <div class="gp-contact-card">
        <div class="ic" style="background:var(--c-accent-soft);color:var(--c-accent)">
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
        </div>
        <h3>Pusat Bantuan</h3>
        <p>Temukan jawaban cepat di FAQ kami.</p>
        <button class="val" onclick={() => onNav('about')} style="background:none;border:none;padding:0;cursor:pointer;font:inherit;color:inherit">Lihat FAQ →</button>
      </div>
    </div>
    <div style="margin-top:32px;background:var(--c-bg);border:1px solid var(--c-line);border-radius:18px;padding:26px">
      <div style="font-size:13px;font-weight:700;color:var(--c-faint);text-transform:uppercase;letter-spacing:.05em;margin-bottom:6px">Kantor</div>
      <p style="margin:0;font-size:15.5px;color:var(--c-ink-soft);line-height:1.6">PT Suluk Mitra Barokah<br/>Jl. Asia Afrika No. 100, Bandung, Jawa Barat 40111</p>
    </div>
  </div>
</section>

<section class="gp-cta">
  <div class="gp-wrap">
    <div class="gp-cta-box">
      <h2>Coba Suluk Sekarang</h2>
      <p>Mulai gratis 14 hari sambil tim kami siap membantu setiap langkah Anda.</p>
      <button class="gp-btn" onclick={() => onGoToApp()}>Mulai Trial Gratis →</button>
    </div>
  </div>
</section>
