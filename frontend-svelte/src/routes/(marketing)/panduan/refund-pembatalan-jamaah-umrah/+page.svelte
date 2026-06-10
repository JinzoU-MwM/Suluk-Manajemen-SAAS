<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("refund-pembatalan-jamaah-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Refund & Pembatalan Jamaah", path: "/panduan/refund-pembatalan-jamaah-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/refund-pembatalan-jamaah-umrah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Cara Mengelola <em>Refund dan Pembatalan</em> Jamaah Umrah</h1>
    <p class="gp-lead">
      Pembatalan tidak terhindarkan. Yang membedakan travel profesional adalah kebijakan yang
      adil, transparan, dan tercatat rapi — sehingga tidak merugikan jamaah maupun travel.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Kebijakan</div>
    <h2 class="gp-h2">Tetapkan kebijakan pembatalan di awal</h2>
    <div class="gp-prose">
      <p>
        Cantumkan ketentuan pembatalan dalam <a href="/fitur/e-kontrak">kontrak</a> sejak
        pendaftaran: biaya yang sudah dikeluarkan travel (tiket, visa, hotel, perlengkapan)
        umumnya tidak bisa dikembalikan penuh. Jelaskan potongan secara wajar berdasarkan biaya
        riil yang sudah terpakai, bukan angka sepihak — ini menjaga kepercayaan dan menghindari
        sengketa.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Proses</div>
    <h2 class="gp-h2">Proses refund yang transparan</h2>
    <div class="gp-prose">
      <ol>
        <li>Terima permohonan pembatalan dan catat tanggalnya (makin dekat keberangkatan, makin besar biaya riil yang sudah terpakai).</li>
        <li>Hitung potongan berdasarkan komponen yang sudah dibayarkan ke vendor.</li>
        <li>Sampaikan rincian ke jamaah, lalu proses pengembalian sisa dana.</li>
        <li>Catat seluruh transaksi agar laporan keuangan tetap akurat.</li>
      </ol>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Pencatatan</div>
    <h2 class="gp-h2">Jaga laporan tetap akurat</h2>
    <div class="gp-prose">
      <p>
        Pembatalan memengaruhi kuota grup, rooming, dan keuangan. Mencatat refund secara
        terstruktur memastikan <a href="/fitur/laporan-keuangan">laporan laba rugi</a> per trip
        tetap benar dan kursi yang kosong bisa segera ditawarkan ke calon jamaah lain.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Kelola pembatalan dengan adil dan tercatat</h2>
    <p>Dari permohonan hingga refund dan dampaknya ke laporan — semua terekam.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
