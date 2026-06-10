<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("cara-mendapatkan-jamaah-umrah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Mendapatkan Jamaah Umrah", path: "/panduan/cara-mendapatkan-jamaah-umrah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/cara-mendapatkan-jamaah-umrah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Cara Mendapatkan <em>Jamaah Umrah</em>: Strategi Marketing Travel</h1>
    <p class="gp-lead">
      Pendaftar yang stabil adalah nyawa bisnis travel. Berikut kanal-kanal yang terbukti
      efektif untuk menjaring jamaah umrah — plus cara agar tidak ada prospek yang terlewat.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Kanal</div>
    <h2 class="gp-h2">Dari mana jamaah datang</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span><b>Referral alumni —</b> jamaah yang puas adalah pemasar terbaik. Minta testimoni dan rekomendasi.</span></li>
        <li><span class="ck">✓</span><span><b>Media sosial —</b> konten edukasi umrah, dokumentasi keberangkatan, dan live manasik membangun kepercayaan.</span></li>
        <li><span class="ck">✓</span><span><b>Komunitas pengajian &amp; masjid —</b> kerja sama dengan majelis taklim dan tokoh agama setempat.</span></li>
        <li><span class="ck">✓</span><span><b>Jaringan agen/mitra —</b> perluas jangkauan lewat agen dengan skema komisi yang jelas.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Konversi</div>
    <h2 class="gp-h2">Jangan biarkan prospek hilang</h2>
    <div class="gp-prose">
      <p>
        Banyak travel kehilangan calon jamaah bukan karena kurang leads, tetapi karena
        follow-up yang berantakan — prospek ditanya di WA lalu lupa ditindaklanjuti. Gunakan
        <a href="/fitur/crm-jamaah">pipeline CRM</a> untuk mencatat setiap prospek, sumbernya,
        dan tahapannya (baru, tertarik, DP, lunas), sehingga tim tahu siapa yang harus
        dihubungi hari ini. Saat closing, prospek otomatis menjadi data jamaah dan
        <a href="/fitur/invoice-umrah">tagihannya langsung dibuat</a>.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Kepercayaan</div>
    <h2 class="gp-h2">Profesionalisme menjual</h2>
    <div class="gp-prose">
      <p>
        Umrah menyangkut uang dan ibadah — jamaah memilih travel yang terpercaya. Proses yang
        rapi (pendaftaran online, kontrak digital, invoice jelas, manifest tertib) memberi
        kesan profesional yang meningkatkan konversi dan rujukan.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Kelola prospek hingga closing dalam satu pipeline</h2>
    <p>Tidak ada lagi calon jamaah yang lupa di-follow-up.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
