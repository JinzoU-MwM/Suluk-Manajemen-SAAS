<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("manifest-digital-mutawwif");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Manifest Digital Mutawwif", path: "/panduan/manifest-digital-mutawwif" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/manifest-digital-mutawwif" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Manifest Digital <em>Mutawwif</em>: Apa Itu &amp; Cara Membuatnya</h1>
    <p class="gp-lead">
      Mutawwif tak perlu lagi membawa tumpukan kertas manifest. Manifest digital menyatukan
      data jamaah, rooming, dan jadwal dalam satu tautan aman yang bisa dibuka dari ponsel.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Pengertian</div>
    <h2 class="gp-h2">Apa itu manifest digital?</h2>
    <div class="gp-prose">
      <p>
        Manifest digital adalah versi online dari daftar jamaah per grup keberangkatan yang
        dibagikan ke mutawwif/muthawwif melalui tautan ber-PIN. Alih-alih lembaran kertas yang
        mudah hilang dan cepat usang, mutawwif membuka data terbaru langsung dari HP — termasuk
        kontak, rooming, dan itinerary.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Manfaat</div>
    <h2 class="gp-h2">Kenapa lebih baik dari kertas</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Selalu sinkron — perubahan rooming atau jadwal langsung terlihat mutawwif.</span></li>
        <li><span class="ck">✓</span><span>Aman — akses dilindungi PIN dan bisa dibatasi masa berlakunya.</span></li>
        <li><span class="ck">✓</span><span>Presensi cepat — tandai kehadiran jamaah dari ponsel, bukan memanggil dari kertas.</span></li>
        <li><span class="ck">✓</span><span>Privasi terjaga — bagikan hanya per grup, bukan seluruh database.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Cara membuat</div>
    <h2 class="gp-h2">Membuat manifest dalam 3 langkah</h2>
    <div class="gp-prose">
      <ol>
        <li>Lengkapi data jamaah dan <a href="/panduan/cara-rooming-hotel-jamaah-umrah">rooming</a> per grup.</li>
        <li>Buat tautan manifest grup tersebut dan atur PIN aksesnya.</li>
        <li>Kirim tautan + PIN ke mutawwif via WhatsApp — selesai.</li>
      </ol>
      <p>
        Manifest digital adalah fitur inti dari <a href="/software-travel-umrah">platform operasional Suluk</a>,
        dan mutawwif bisa membukanya lewat <a href="/unduh">aplikasi mobile</a>.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Bagikan manifest digital ke tim Anda</h2>
    <p>Data jamaah, rooming, dan jadwal dalam satu tautan aman ber-PIN.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
