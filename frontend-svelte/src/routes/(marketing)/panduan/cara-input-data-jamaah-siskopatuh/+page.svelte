<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("cara-input-data-jamaah-siskopatuh");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Input Data Jamaah Siskopatuh", path: "/panduan/cara-input-data-jamaah-siskopatuh" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/cara-input-data-jamaah-siskopatuh" {schema} />

<article>
  <section class="gp-hero">
    <div class="gp-wrap">
      <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
      <h1 class="gp-h1">Cara Input Data Jamaah ke <em>Siskopatuh</em> Lebih Cepat</h1>
      <p class="gp-lead">
        Input data jamaah satu per satu ke Excel Siskopatuh 32 kolom adalah salah satu
        pekerjaan paling memakan waktu di travel umrah. Berikut cara memangkasnya dari
        berjam-jam menjadi hitungan menit.
      </p>
      <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
    </div>
  </section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Masalah</div>
    <h2 class="gp-h2">Mengapa input manual rawan error</h2>
    <div class="gp-prose">
      <p>
        Format Siskopatuh menuntut 32 kolom data per jamaah — mulai dari NIK, nama sesuai
        paspor, tanggal lahir, hingga data keluarga. Mengetik ulang dari KTP dan KK secara
        manual tidak hanya lambat, tetapi juga sangat rawan salah ketik: satu digit NIK
        keliru bisa membuat data ditolak saat pelaporan.
      </p>
      <p>
        Ketika satu grup berisi 45 jamaah dan setiap orang punya beberapa dokumen, pekerjaan
        ini bisa menyita waktu admin seharian penuh — waktu yang seharusnya dipakai untuk
        melayani calon jamaah baru.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Langkah 1</div>
    <h2 class="gp-h2">Scan KTP &amp; KK, biarkan AI yang mengetik</h2>
    <div class="gp-prose">
      <p>
        Alih-alih mengetik manual, foto atau unggah KTP dan KK jamaah. AI Scanner akan
        mengekstrak NIK, nama, tempat &amp; tanggal lahir, alamat, dan data keluarga secara
        otomatis ke field yang sesuai. Pelajari detailnya di halaman
        <a href="/software-travel-umrah">software travel umrah</a> kami.
      </p>
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Satu foto KTP mengisi belasan field sekaligus.</span></li>
        <li><span class="ck">✓</span><span>KK mengisi data hubungan keluarga tanpa ketik ulang.</span></li>
        <li><span class="ck">✓</span><span>Bisa dari ponsel langsung lewat <a href="/unduh">aplikasi mobile</a> saat jamaah datang.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Langkah 2</div>
    <h2 class="gp-h2">Verifikasi dan lengkapi data paspor</h2>
    <div class="gp-prose">
      <p>
        Setelah AI mengisi data dasar, lakukan verifikasi cepat: pastikan nama sudah sesuai
        ejaan paspor (bukan KTP), karena Siskopatuh dan visa mengacu pada paspor. Lengkapi
        nomor paspor, masa berlaku, dan data vaksinasi. Sistem menandai field yang masih
        kosong sehingga tidak ada data yang terlewat.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Langkah 3</div>
    <h2 class="gp-h2">Ekspor ke Excel Siskopatuh 32 kolom</h2>
    <div class="gp-prose">
      <p>
        Saat data sudah lengkap, ekspor seluruh jamaah dalam satu grup ke file Excel dengan
        susunan 32 kolom yang siap diunggah ke Siskopatuh — tanpa menyusun ulang kolom secara
        manual. Pekerjaan yang biasanya seharian kini selesai dalam beberapa menit.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap">
    <div class="gp-cta-box">
      <h2>Coba sendiri tanpa biaya</h2>
      <p>Scan 5 dokumen gratis dan rasakan input data jamaah 10x lebih cepat.</p>
      <a class="gp-btn" href="/daftar">Coba Gratis →</a>
    </div>
  </div></section>
</article>
