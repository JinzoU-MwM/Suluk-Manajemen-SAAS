<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("tips-memilih-software-travel-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Tips Memilih Software Travel Umrah", path: "/panduan/tips-memilih-software-travel-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/tips-memilih-software-travel-umrah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">7 Tips Memilih <em>Software Travel Umrah</em> yang Tepat</h1>
    <p class="gp-lead">
      Software yang tepat menghemat puluhan jam kerja admin dan mencegah kesalahan mahal.
      Berikut tujuh kriteria yang sebaiknya Anda nilai sebelum memutuskan.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Kriteria</div>
    <h2 class="gp-h2">Yang wajib ada</h2>
    <div class="gp-prose">
      <ol>
        <li><b>Dibuat khusus untuk umrah/haji</b> — memahami istilah dan alur PPIU, bukan software akuntansi umum.</li>
        <li><b>Dukungan Siskopatuh</b> — bisa <a href="/panduan/cara-input-data-jamaah-siskopatuh">ekspor Excel 32 kolom</a> tanpa menata ulang manual.</li>
        <li><b>OCR dokumen</b> — scan KTP/KK/paspor mengisi data otomatis, mengurangi salah ketik.</li>
        <li><b>Keuangan terintegrasi</b> — <a href="/fitur/invoice-umrah">invoice</a> &amp; <a href="/fitur/laporan-keuangan">laporan</a> nyambung, profit per trip terpantau.</li>
        <li><b>E-kontrak &amp; CRM</b> — <a href="/fitur/e-kontrak">akad digital</a> dan pipeline prospek dalam satu alur.</li>
        <li><b>Aplikasi mobile</b> — tim lapangan &amp; muthawwif bisa kerja dari <a href="/unduh">ponsel</a>.</li>
        <li><b>Keamanan &amp; dukungan</b> — data terenkripsi, ada backup, dan tim support yang responsif.</li>
      </ol>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Tips</div>
    <h2 class="gp-h2">Cara mengevaluasi sebelum berlangganan</h2>
    <div class="gp-prose">
      <p>
        Manfaatkan masa uji coba gratis untuk menguji alur kerja nyata: scan beberapa dokumen,
        buat satu paket, terbitkan invoice, dan susun rooming satu grup. Jika dalam percobaan
        singkat saja sudah terasa lebih cepat, itu pertanda baik. Bandingkan juga model harga —
        per jamaah, per bulan, atau per fitur — dengan skala travel Anda.
      </p>
      <p>
        Suluk memenuhi ketujuh kriteria di atas dalam satu platform. Pelajari selengkapnya di
        halaman <a href="/software-travel-umrah">software travel umrah</a>.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Uji sendiri, gratis 14 hari</h2>
    <p>Tanpa kartu kredit. Rasakan alur kerja lengkap sebelum memutuskan.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis 14 Hari →</a>
  </div></div></section>
</article>
