<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("cara-membuat-paket-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Cara Membuat Paket Umrah", path: "/panduan/cara-membuat-paket-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/cara-membuat-paket-umrah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Cara Membuat <em>Paket Umrah</em> yang Menarik dan Menguntungkan</h1>
    <p class="gp-lead">
      Paket yang dirancang baik menyeimbangkan nilai bagi jamaah dan margin bagi travel.
      Berikut komponen yang perlu Anda susun dan cara menentukan harganya.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Komponen</div>
    <h2 class="gp-h2">Susun isi paket</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span><b>Durasi &amp; tanggal —</b> 9 hari, 12 hari, atau paket plus (Turki/Aqsa).</span></li>
        <li><span class="ck">✓</span><span><b>Hotel —</b> bintang dan jarak ke Masjidil Haram/Nabawi; sebutkan nama hotel bila bisa.</span></li>
        <li><span class="ck">✓</span><span><b>Maskapai &amp; rute —</b> langsung atau transit memengaruhi harga dan kenyamanan.</span></li>
        <li><span class="ck">✓</span><span><b>Tipe kamar —</b> Quad/Triple/Double dengan harga berjenjang.</span></li>
        <li><span class="ck">✓</span><span><b>Layanan termasuk —</b> visa, manasik, perlengkapan, muthawwif, transportasi, katering.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Harga</div>
    <h2 class="gp-h2">Tetapkan harga berjenjang</h2>
    <div class="gp-prose">
      <p>
        Hitung biaya per jamaah untuk tiap tipe kamar, tambahkan alokasi biaya tetap dan
        margin, lalu tetapkan harga per tier. Panduan lengkapnya ada di
        <a href="/panduan/menghitung-biaya-profit-paket-umrah">cara menghitung biaya dan profit
        paket umrah</a>. Pastikan setiap tier tetap untung — jangan sampai kamar Quad yang
        paling laku justru menggerus margin.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Publikasi</div>
    <h2 class="gp-h2">Tampilkan paket secara profesional</h2>
    <div class="gp-prose">
      <p>
        Sajikan paket dengan detail jelas (fasilitas, harga per tipe, sisa kuota) di halaman
        publik yang bisa dibagikan ke calon jamaah. Dengan
        <a href="/software-travel-umrah">platform operasional</a>, paket yang Anda buat bisa
        langsung dipublikasikan dan dihubungkan ke pendaftaran serta invoice.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Buat, hargai, dan publikasikan paket dalam satu tempat</h2>
    <p>Dari komponen biaya hingga halaman paket publik yang siap dibagikan.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
