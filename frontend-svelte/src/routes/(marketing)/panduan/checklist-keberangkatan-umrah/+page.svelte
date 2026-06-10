<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("checklist-keberangkatan-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Checklist Keberangkatan Umrah", path: "/panduan/checklist-keberangkatan-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/checklist-keberangkatan-umrah" {schema} />

<article>
  <section class="gp-hero">
    <div class="gp-wrap">
      <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
      <h1 class="gp-h1">Checklist Persiapan <em>Keberangkatan</em> Jamaah Umrah</h1>
      <p class="gp-lead">
        H-30 hingga hari keberangkatan adalah masa paling krusial. Gunakan daftar periksa
        ini agar tidak ada dokumen, perlengkapan, atau jamaah yang terlewat.
      </p>
      <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
    </div>
  </section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">H-30</div>
    <h2 class="gp-h2">Dokumen &amp; administrasi</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Paspor seluruh jamaah masih berlaku minimal 6 bulan.</span></li>
        <li><span class="ck">✓</span><span>Visa umrah terbit dan dicocokkan dengan nama paspor.</span></li>
        <li><span class="ck">✓</span><span>Data jamaah terlapor di Siskopatuh — lihat <a href="/panduan/cara-input-data-jamaah-siskopatuh">cara input data jamaah</a>.</span></li>
        <li><span class="ck">✓</span><span>Status pembayaran lunas terpantau lewat modul <a href="/fitur/invoice-umrah">invoice</a>.</span></li>
        <li><span class="ck">✓</span><span>Kontrak/akad jamaah sudah ditandatangani secara <a href="/fitur/e-kontrak">digital</a>.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">H-14</div>
    <h2 class="gp-h2">Manifest, grup &amp; rooming</h2>
    <div class="gp-prose">
      <p>
        Finalisasi pembagian grup dan kamar hotel. Pastikan pasangan dan keluarga
        disatukan, serta gender dipisahkan sesuai aturan. Bagikan manifest digital ber-PIN
        kepada muthawwif agar mereka punya data jamaah dan rooming lengkap di ponsel.
      </p>
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Rooming list hotel Makkah &amp; Madinah final.</span></li>
        <li><span class="ck">✓</span><span>Manifest digital dibagikan ke muthawwif setiap grup.</span></li>
        <li><span class="ck">✓</span><span>Jadwal/itinerary harian tersusun dan dapat diakses jamaah.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">H-7</div>
    <h2 class="gp-h2">Perlengkapan &amp; manasik</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Koper, kain ihram, mukena, ID card, dan buku doa terdistribusi — tercatat siapa sudah menerima.</span></li>
        <li><span class="ck">✓</span><span>Manasik umrah terlaksana dan kehadiran jamaah tercatat.</span></li>
        <li><span class="ck">✓</span><span>Kontak darurat keluarga dan koordinator lapangan terkumpul.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Hari-H</div>
    <h2 class="gp-h2">Di bandara</h2>
    <div class="gp-prose">
      <p>
        Lakukan presensi jamaah saat berkumpul, cocokkan dengan manifest, dan pastikan tidak
        ada yang tertinggal sebelum check-in. Dengan manifest digital, koordinator cukup
        menandai kehadiran dari ponsel — lebih cepat daripada memanggil satu per satu dari
        kertas.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap">
    <div class="gp-cta-box">
      <h2>Kelola seluruh checklist dalam satu sistem</h2>
      <p>Dari dokumen hingga rooming dan manifest — semua terpantau di satu dashboard.</p>
      <a class="gp-btn" href="/daftar">Coba Gratis 14 Hari →</a>
    </div>
  </div></section>
</article>
