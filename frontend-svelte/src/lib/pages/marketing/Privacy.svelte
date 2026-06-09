<script>
  let { onNav = () => {}, onGoToApp = () => {} } = $props();
</script>

<section class="gp-hero">
  <div class="gp-wrap">
    <p class="gp-badge">Legal</p>
    <h1 class="gp-h1">Kebijakan Privasi</h1>
    <p class="gp-lead">Kami menghargai kepercayaan Anda. Halaman ini menjelaskan bagaimana Suluk mengumpulkan, menggunakan, dan melindungi data Anda.</p>
  </div>
</section>

<section class="gp-legal">
  <div class="gp-wrap">
    <p class="updated">Terakhir diperbarui: 8 Juni 2026</p>

    <h2>1. Pengantar</h2>
    <p>Kebijakan ini berlaku untuk seluruh layanan Suluk yang dioperasikan oleh PT Suluk Mitra Barokah. Dengan menggunakan layanan kami, Anda menyetujui praktik pengelolaan data yang dijelaskan di sini.</p>

    <h2>2. Data yang Kami Kumpulkan</h2>
    <h3>Data Akun</h3>
    <ul>
      <li>Nama, email, nomor telepon, dan nama travel/perusahaan.</li>
      <li>Informasi langganan dan riwayat pembayaran.</li>
    </ul>
    <h3>Data Operasional</h3>
    <ul>
      <li>Data jamaah yang Anda input (identitas, dokumen, pembayaran).</li>
      <li>Data paket, invoice, kontrak, dan transaksi keuangan.</li>
    </ul>
    <h3>Data Teknis</h3>
    <ul>
      <li>Alamat IP, jenis perangkat, dan aktivitas penggunaan untuk keamanan dan peningkatan layanan.</li>
    </ul>

    <h2>3. Bagaimana Kami Menggunakan Data</h2>
    <p>Data digunakan untuk menyediakan dan meningkatkan layanan, memproses pembayaran, memberikan dukungan, serta menjaga keamanan akun. Kami tidak menjual data Anda kepada pihak ketiga.</p>

    <h2>4. Penyimpanan &amp; Keamanan</h2>
    <ul>
      <li>Data dienkripsi saat transit dan saat disimpan.</li>
      <li>Pencadangan (backup) rutin untuk mencegah kehilangan data.</li>
      <li>Audit log untuk melacak akses terhadap data sensitif.</li>
      <li>Akses dibatasi sesuai peran pengguna dalam akun travel Anda.</li>
    </ul>

    <h2>5. Hak Pengguna</h2>
    <ul>
      <li>Mengakses dan memperbarui data Anda kapan saja.</li>
      <li>Meminta koreksi data yang tidak akurat.</li>
      <li>Meminta penghapusan akun dan data terkait.</li>
      <li>Mengekspor data jamaah dan transaksi Anda.</li>
    </ul>

    <h2>6. Retensi Data</h2>
    <p>Data akun aktif disimpan selama langganan berjalan. Setelah pembatalan, data disimpan hingga 30 hari sebelum dihapus permanen, kecuali diwajibkan lain oleh hukum.</p>

    <h2>7. Perubahan Kebijakan</h2>
    <p>Kami dapat memperbarui kebijakan ini sewaktu-waktu. Perubahan signifikan akan diberitahukan melalui email atau notifikasi dalam aplikasi.</p>

    <h2>8. Kontak</h2>
    <p>Pertanyaan terkait privasi dapat dikirim ke <a href="mailto:privasi@suluk.id" style="color:var(--c-primary);font-weight:700">privasi@suluk.id</a> atau melalui halaman <button type="button" class="gp-link-btn" style="color:var(--c-primary);font-weight:700" onclick={() => onNav('contact')}>Hubungi Kami</button>.</p>
  </div>
</section>
