<script>
  let { onNav = () => {}, onGoToApp = () => {} } = $props();
</script>

<main class="gp-404">
  <div class="big">404</div>
  <h1 style="font-family:var(--font-display);font-size:30px;font-weight:800;margin:14px 0 10px">Halaman Tidak Ditemukan</h1>
  <p style="font-size:16px;color:var(--c-muted);max-width:420px;margin:0 0 26px;line-height:1.6">Maaf, halaman yang Anda cari mungkin sudah dipindahkan atau tidak tersedia.</p>
  <div class="gp-cta-row" style="justify-content:center">
    <button type="button" class="gp-btn gp-btn-primary" onclick={() => onGoToApp()}>Kembali ke Beranda</button>
    <button type="button" class="gp-btn gp-btn-ghost" onclick={() => onNav('contact')}>Hubungi Kami</button>
  </div>
</main>
