<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("dokumen-persyaratan-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Dokumen Persyaratan Umrah", path: "/panduan/dokumen-persyaratan-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/dokumen-persyaratan-umrah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Dokumen Persyaratan <em>Umrah</em> yang Wajib Disiapkan</h1>
    <p class="gp-lead">
      Dokumen yang tidak lengkap adalah penyebab umum tertundanya keberangkatan. Berikut daftar
      dokumen yang umumnya wajib disiapkan setiap jamaah umrah.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Daftar Dokumen</div>
    <h2 class="gp-h2">Yang umumnya diminta</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span><b>Paspor</b> berlaku minimal 6 bulan, nama minimal dua kata sesuai ketentuan.</span></li>
        <li><span class="ck">✓</span><span><b>KTP dan Kartu Keluarga</b> untuk verifikasi identitas dan mahram.</span></li>
        <li><span class="ck">✓</span><span><b>Pasfoto</b> berlatar putih sesuai spesifikasi yang berlaku.</span></li>
        <li><span class="ck">✓</span><span><b>Sertifikat vaksin meningitis</b> (dan vaksin lain bila disyaratkan).</span></li>
        <li><span class="ck">✓</span><span><b>Bukti hubungan mahram</b> (buku nikah/akta) untuk jamaah perempuan sesuai ketentuan.</span></li>
        <li><span class="ck">✓</span><span>Dokumen tambahan sesuai kebijakan terbaru (mis. terkait kesehatan).</span></li>
      </ul>
      <p>
        <em>Persyaratan dapat berubah mengikuti regulasi Arab Saudi dan Kementerian Agama —
        konfirmasikan yang terbaru sebelum setiap keberangkatan.</em>
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Pengelolaan</div>
    <h2 class="gp-h2">Agar tidak ada dokumen yang tertinggal</h2>
    <div class="gp-prose">
      <p>
        Kumpulkan dokumen sejak pendaftaran. Dengan
        <a href="/panduan/cara-input-data-jamaah-siskopatuh">scan KTP/KK/paspor</a>, data
        jamaah terisi otomatis dan dokumen tersimpan digital per orang. Sistem dapat menandai
        jamaah yang dokumennya belum lengkap, sehingga tim langsung tahu siapa yang perlu
        ditindaklanjuti — jauh sebelum mendekati hari keberangkatan.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Simpan dan pantau dokumen jamaah secara digital</h2>
    <p>Scan sekali, tersimpan rapi per jamaah, dan ketahuan mana yang belum lengkap.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
