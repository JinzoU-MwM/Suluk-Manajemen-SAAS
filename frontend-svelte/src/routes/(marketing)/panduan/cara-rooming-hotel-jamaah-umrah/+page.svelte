<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("cara-rooming-hotel-jamaah-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Rooming Hotel Jamaah", path: "/panduan/cara-rooming-hotel-jamaah-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/cara-rooming-hotel-jamaah-umrah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Cara Mengatur <em>Rooming Hotel</em> Jamaah Umrah</h1>
    <p class="gp-lead">
      Pembagian kamar hotel di Makkah dan Madinah sering jadi sumber komplain jamaah. Berikut
      cara menyusun rooming list yang adil, sesuai syariat, dan cepat.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Prinsip</div>
    <h2 class="gp-h2">Aturan dasar pembagian kamar</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Pisahkan jamaah laki-laki dan perempuan ke kamar berbeda.</span></li>
        <li><span class="ck">✓</span><span>Satukan pasangan suami-istri dan satu keluarga jika memungkinkan.</span></li>
        <li><span class="ck">✓</span><span>Isi kamar sesuai tipe paket: Quad (4), Triple (3), atau Double (2).</span></li>
        <li><span class="ck">✓</span><span>Pertimbangkan usia dan kondisi kesehatan — lansia dekat dengan pendamping.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Manual vs Otomatis</div>
    <h2 class="gp-h2">Mengapa rooming manual melelahkan</h2>
    <div class="gp-prose">
      <p>
        Menyusun rooming 45 jamaah di spreadsheet berarti memindahkan nama satu per satu
        sambil menjaga aturan gender dan keluarga — dan setiap perubahan jamaah memaksa Anda
        menata ulang. Fitur auto-rooming pada <a href="/software-travel-umrah">software travel umrah</a>
        mengisi kamar otomatis sesuai aturan tersebut, lalu menyisakan drag &amp; drop hanya
        untuk penyesuaian akhir.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Langkah</div>
    <h2 class="gp-h2">Menyusun rooming dalam 3 langkah</h2>
    <div class="gp-prose">
      <ol>
        <li>Pastikan data gender &amp; relasi keluarga jamaah sudah lengkap.</li>
        <li>Jalankan auto-rooming sesuai tipe kamar paket — sistem mengisi otomatis.</li>
        <li>Sesuaikan kasus khusus (lansia, permintaan sekamar) dengan drag &amp; drop, lalu bagikan ke manifest.</li>
      </ol>
      <p>
        Setelah final, rooming tampil otomatis di
        <a href="/panduan/manifest-digital-mutawwif">manifest digital mutawwif</a> — lihat juga
        <a href="/panduan/checklist-keberangkatan-umrah">checklist keberangkatan</a> untuk langkah berikutnya.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Atur rooming dalam hitungan menit</h2>
    <p>Auto-rooming yang memisahkan gender dan menyatukan keluarga — cukup satu klik.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
