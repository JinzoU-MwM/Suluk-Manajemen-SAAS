<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb, article } from "$lib/seo/schema.js";
  import { getPanduan } from "$lib/content/panduan.js";

  const meta = getPanduan("mengelola-cicilan-pembayaran-jamaah");
  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Panduan", path: "/panduan" },
      { name: "Mengelola Cicilan Jamaah", path: "/panduan/mengelola-cicilan-pembayaran-jamaah" },
    ]),
    article({ title: meta.title, description: meta.description, slug: meta.slug, datePublished: meta.date }),
  ];
</script>

<Seo title={`${meta.title} | Suluk`} description={meta.description} path="/panduan/mengelola-cicilan-pembayaran-jamaah" {schema} />

<article>
  <section class="gp-hero"><div class="gp-wrap">
    <p class="gp-badge"><a href="/panduan" style="color:inherit;text-decoration:none">← Panduan</a></p>
    <h1 class="gp-h1">Cara Mengelola Pembayaran <em>Cicilan</em> Jamaah Umrah</h1>
    <p class="gp-lead">
      Sebagian besar jamaah membayar bertahap. Tanpa pencatatan yang rapi, travel mudah
      kehilangan jejak siapa yang sudah lunas dan siapa yang menunggak. Begini mengelolanya.
    </p>
    <p class="gp-meta">{meta.readMinutes} menit baca · Diperbarui {meta.date}</p>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Skema</div>
    <h2 class="gp-h2">Tentukan skema pembayaran yang jelas</h2>
    <div class="gp-prose">
      <p>
        Tetapkan besaran DP, jumlah dan jadwal cicilan, serta batas pelunasan (biasanya
        beberapa minggu sebelum keberangkatan). Komunikasikan skema ini di awal dan tuangkan
        dalam <a href="/fitur/e-kontrak">kontrak</a> agar tidak ada salah paham.
      </p>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Pencatatan</div>
    <h2 class="gp-h2">Catat setiap pembayaran ke invoice</h2>
    <div class="gp-prose">
      <ul class="gp-benefits">
        <li><span class="ck">✓</span><span>Setiap cicilan dicatat ke <a href="/fitur/invoice-umrah">invoice</a> jamaah, bukan ke buku terpisah.</span></li>
        <li><span class="ck">✓</span><span>Sisa tagihan dan status (lunas/kurang) terlihat otomatis di dashboard.</span></li>
        <li><span class="ck">✓</span><span>Kwitansi bisa langsung dikirim ke jamaah via tautan atau WhatsApp.</span></li>
      </ul>
    </div>
  </div></section>

  <section class="gp-sec"><div class="gp-wrap">
    <div class="gp-eyebrow">Tunggakan</div>
    <h2 class="gp-h2">Pantau dan ingatkan jatuh tempo</h2>
    <div class="gp-prose">
      <p>
        Dengan piutang yang terpantau real-time, Anda bisa mengingatkan jamaah yang mendekati
        jatuh tempo sebelum keberangkatan terganggu. Ini juga menjaga arus kas — uang yang
        seharusnya masuk tidak menggantung tanpa ketahuan. Lihat dampaknya pada
        <a href="/panduan/menghitung-biaya-profit-paket-umrah">profit per trip</a>.
      </p>
    </div>
  </div></section>

  <section class="gp-cta"><div class="gp-wrap"><div class="gp-cta-box">
    <h2>Pantau cicilan &amp; tunggakan otomatis</h2>
    <p>Tahu persis siapa sudah lunas dan siapa yang perlu diingatkan.</p>
    <a class="gp-btn" href="/daftar">Coba Gratis →</a>
  </div></div></section>
</article>
