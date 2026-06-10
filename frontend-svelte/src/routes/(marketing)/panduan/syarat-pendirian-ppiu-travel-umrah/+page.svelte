<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("syarat-pendirian-ppiu-travel-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Syarat Pendirian PPIU", path: "/panduan/syarat-pendirian-ppiu-travel-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/syarat-pendirian-ppiu-travel-umrah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Syarat dan Cara Mendirikan <em>PPIU</em> (Travel Umrah Resmi)</h1>
    <p class="gp-lead">
      Untuk memberangkatkan jamaah umrah secara legal, travel wajib berstatus PPIU
      (Penyelenggara Perjalanan Ibadah Umrah) berizin Kementerian Agama. Berikut gambaran
      syarat dan tahapannya.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Pengertian</div>
    <h2 class="gp-h2">Apa itu PPIU?</h2>
    <div class="gp-prose">
      <p>
        PPIU adalah biro perjalanan yang memperoleh izin resmi dari Kementerian Agama untuk
        menyelenggarakan perjalanan ibadah umrah. Tanpa izin ini, sebuah travel tidak boleh
        memberangkatkan jamaah dan berisiko terkena sanksi. Izin PPIU melindungi jamaah dan
        memberi kredibilitas pada travel Anda.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Syarat Umum</div>
    <h2 class="gp-h2">Persyaratan yang biasanya diminta</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Berbentuk badan usaha (PT) dengan akta dan pengesahan yang sah.</span></li>
        <li><span class="ck">✓</span><span>Izin usaha melalui sistem OSS (Online Single Submission) dengan KBLI biro perjalanan.</span></li>
        <li><span class="ck">✓</span><span>Memiliki kantor tetap dan sarana yang memadai.</span></li>
        <li><span class="ck">✓</span><span>Modal dan jaminan (mis. bank garansi) sesuai ketentuan yang berlaku.</span></li>
        <li><span class="ck">✓</span><span>Susunan pengurus, komisaris, dan tenaga ahli/pembimbing ibadah.</span></li>
        <li><span class="ck">✓</span><span>Komitmen memenuhi standar pelayanan dan pelaporan ke Siskopatuh.</span></li>
      </ul>
      <p>
        <em>Catatan: persyaratan dan nominal modal/jaminan dapat berubah mengikuti regulasi
        terbaru. Selalu verifikasi ke kanwil Kementerian Agama atau portal resmi sebelum
        memulai proses.</em>
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Tahapan</div>
    <h2 class="gp-h2">Langkah pengajuan</h2>
    <div class="gp-prose">
      <ol>
        <li>Dirikan PT dan urus legalitas dasar (akta, NPWP, NIB via OSS).</li>
        <li>Lengkapi sarana: kantor, struktur organisasi, dan jaminan sesuai ketentuan.</li>
        <li>Ajukan permohonan izin PPIU ke Kementerian Agama dan lalui verifikasi.</li>
        <li>Setelah izin terbit, daftarkan dan laporkan operasional melalui Siskopatuh.</li>
      </ol>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Setelah Berizin</div>
    <h2 class="gp-h2">Siapkan sistem operasional sejak awal</h2>
    <div class="gp-prose">
      <p>
        Izin hanyalah awal. Begitu beroperasi, Anda harus mengelola data jamaah,
        <a href="/panduan/cara-input-data-jamaah-siskopatuh">pelaporan Siskopatuh</a>,
        keuangan, dan dokumen secara rapi dan dapat diaudit. Menyiapkan
        <a href="/software-travel-umrah">sistem operasional travel umrah</a> sejak awal membuat
        travel baru langsung tertib dan profesional.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Mulai tertib sejak hari pertama</h2>
    <p>Kelola jamaah, keuangan, dan dokumen dalam satu sistem — gratis untuk travel baru.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
