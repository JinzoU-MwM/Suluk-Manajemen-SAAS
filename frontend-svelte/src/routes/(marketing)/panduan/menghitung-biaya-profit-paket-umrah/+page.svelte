<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("menghitung-biaya-profit-paket-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Biaya & Profit Paket Umrah", path: "/panduan/menghitung-biaya-profit-paket-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/menghitung-biaya-profit-paket-umrah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Cara Menghitung <em>Biaya dan Profit</em> Paket Umrah</h1>
    <p class="gp-lead">
      Menentukan harga paket terlalu rendah menggerus margin; terlalu tinggi membuat kalah
      saing. Berikut cara menyusun harga paket umrah berbasis data, bukan kira-kira.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Komponen Biaya</div>
    <h2 class="gp-h2">Rincian biaya per jamaah</h2>
    <div class="gp-prose">
      <p>Mulai dari biaya langsung yang menempel pada setiap jamaah:</p>
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Tiket pesawat (PP) dan airport handling.</span></li>
        <li><span class="ck">✓</span><span>Akomodasi hotel Makkah &amp; Madinah (dibagi per tipe kamar).</span></li>
        <li><span class="ck">✓</span><span>Visa, muassasah, transportasi darat, dan katering.</span></li>
        <li><span class="ck">✓</span><span>Perlengkapan jamaah (koper, ihram, ID card) dan manasik.</span></li>
        <li><span class="ck">✓</span><span>Honor muthawwif &amp; tim — dibagi ke jumlah jamaah per trip.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Margin</div>
    <h2 class="gp-h2">Menetapkan harga jual</h2>
    <div class="gp-prose">
      <p>
        Jumlahkan biaya langsung per jamaah, tambahkan alokasi biaya tetap (gaji kantor,
        marketing, operasional) dibagi target jamaah, lalu tambahkan margin keuntungan. Tip
        penting: biaya tipe kamar berbeda (Quad lebih murah dari Double), jadi tetapkan harga
        berjenjang per tipe agar margin tiap tier tetap sehat.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Pantau</div>
    <h2 class="gp-h2">Lacak profit aktual per trip</h2>
    <div class="gp-prose">
      <p>
        Rencana hanya separuh cerita — yang menentukan adalah realisasi. Dengan
        <a href="/fitur/invoice-umrah">invoice</a> mencatat pemasukan dan vendor mencatat
        pengeluaran, <a href="/fitur/laporan-keuangan">laporan keuangan</a> menampilkan laba
        rugi per trip secara real-time, sehingga Anda tahu paket mana yang benar-benar
        menguntungkan dan mana yang perlu dikoreksi.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Pantau profit setiap trip secara real-time</h2>
    <p>Dari harga paket hingga laba rugi — semua otomatis dalam satu sistem.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
