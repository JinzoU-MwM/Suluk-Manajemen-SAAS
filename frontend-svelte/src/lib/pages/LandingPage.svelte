<script>
  import { PLANS } from "../config/pricing.js";
  import LandingNav from "../components/LandingNav.svelte";

  let { onGoToLogin = () => {}, onGoToRegister = () => {}, onNavigate = () => {} } = $props();

  function scrollTop(e) {
    e?.preventDefault();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  // WhatsApp contact for the Enterprise "Hubungi Sales" CTA.
  const SALES_WA = "https://wa.me/6285159980404?text=" +
    encodeURIComponent("Halo, saya tertarik dengan paket Enterprise Suluk.");

  function priceText(p) {
    if (!p.purchasable) return p.priceLabel; // "Gratis" / "Custom"
    return p.priceLabel;
  }

  function handlePlanCta(p) {
    if (p.key === "enterprise") {
      window.open(SALES_WA, "_blank");
    } else {
      onGoToRegister();
    }
  }

  const modules = [
    { title: "Data Jamaah", desc: "Kelola data calon jamaah, dokumen, dan status pembayaran lengkap.", color: "#0f7a5a", svg: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>' },
    { title: "AI Scanner", desc: "Pindai KTP, KK, dan paspor — data terisi otomatis dengan OCR.", color: "#2563c9", svg: '<path d="M3 7V5a2 2 0 0 1 2-2h2"/><path d="M17 3h2a2 2 0 0 1 2 2v2"/><path d="M21 17v2a2 2 0 0 1-2 2h-2"/><path d="M7 21H5a2 2 0 0 1-2-2v-2"/><line x1="7" y1="12" x2="17" y2="12"/>' },
    { title: "Paket Perjalanan", desc: "Atur paket umrah & haji, tier harga, kuota, dan publikasi.", color: "#c79a3e", svg: '<path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22" x2="12" y2="12"/>' },
    { title: "CRM & Pipeline", desc: "Kelola lead dari iklan hingga closing dengan pipeline visual.", color: "#7a5ae0", svg: '<rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>' },
    { title: "Invoice & Pembayaran", desc: "Buat tagihan, catat cicilan, dan pantau tunggakan otomatis.", color: "#0f7a5a", svg: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="9" y1="13" x2="15" y2="13"/><line x1="9" y1="17" x2="13" y2="17"/>' },
    { title: "Keuangan", desc: "Laporan laba rugi, arus kas, dan rincian biaya operasional.", color: "#15564a", svg: '<path d="M19 7V5a2 2 0 0 0-2-2H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v4a1 1 0 0 1-1 1h-3a2 2 0 0 1 0-4h4"/><path d="M3 5v14a2 2 0 0 0 2 2h15a1 1 0 0 0 1-1v-4"/>' },
    { title: "Rooming List", desc: "Atur penempatan kamar hotel dengan seret-lepas yang intuitif.", color: "#b87708", svg: '<path d="M2 4v16"/><path d="M2 8h18a2 2 0 0 1 2 2v10"/><path d="M2 17h20"/><path d="M6 8v9"/>' },
    { title: "Kontrak Digital", desc: "Template akad & tanda tangan digital langsung dari WhatsApp.", color: "#a9842f", svg: '<path d="M3 17.5V21h3.5l11-11-3.5-3.5-11 11z"/><path d="M14 6.5 17.5 10"/><path d="M3 21h18"/>' },
    { title: "Vendor & Agen", desc: "Pantau maskapai, hotel, muassasah, serta komisi agen mitra.", color: "#2563c9", svg: '<path d="M14 18V6a2 2 0 0 0-2-2H4a2 2 0 0 0-2 2v11a1 1 0 0 0 1 1h2"/><path d="M15 18H9"/><path d="M19 18h2a1 1 0 0 0 1-1v-3.65a1 1 0 0 0-.22-.62l-3.48-4.35A1 1 0 0 0 17.52 8H14"/><circle cx="17" cy="18" r="2"/><circle cx="7" cy="18" r="2"/>' },
  ];

  const testimonials = [
    { quote: "Suluk benar-benar mengubah cara kami bekerja. Input jamaah yang dulu makan waktu berjam-jam, sekarang cukup scan KTP. Tim admin kami jauh lebih produktif.", name: "Hj. Nurul Hidayah", role: "Owner, Cahaya Iman Tour — Jakarta", color: "#0f7a5a" },
    { quote: "Laporan keuangan dan tunggakan jamaah langsung kelihatan real-time. Saya bisa ambil keputusan lebih cepat tanpa harus menunggu rekap manual.", name: "H. Lukman Hakim", role: "Direktur, Mitra Suci Wisata — Surabaya", color: "#c79a3e" },
    { quote: "Fitur rooming list dan kontrak digitalnya sangat membantu saat musim haji. Semua rapi, jamaah pun merasa lebih profesional dilayani.", name: "Rizki Pratama", role: "Manajer, Safar Amanah — Medan", color: "#2563c9" },
  ];

  function initials(name) {
    return name.replace(/^(H\.|Hj\.)\s*/, "").split(" ").slice(0, 2).map((w) => w[0]).join("");
  }
</script>

<div class="lp">
  <!-- NAV -->
  <LandingNav />

  <!-- HERO -->
  <header class="lp-hero" id="top">
    <div class="lp-container lp-hero-grid">
      <div>
        <span class="lp-kicker">
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/></svg>
          Platform #1 Travel Umrah &amp; Haji di Indonesia
        </span>
        <h1 class="lp-h1">Kelola travel umrah &amp; haji Anda dalam <em>satu sistem</em></h1>
        <p class="lp-lead">Dari pemindaian dokumen jamaah dengan AI, manajemen paket, keuangan, hingga kontrak digital — semuanya otomatis, rapi, dan terintegrasi.</p>
        <div class="lp-hero-cta">
          <button class="lp-btn lp-btn-primary lp-btn-lg" onclick={() => onGoToRegister()}>
            Coba Gratis 14 Hari
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="M12 5l7 7-7 7"/></svg>
          </button>
          <button class="lp-btn lp-btn-ghost lp-btn-lg" onclick={() => onGoToLogin()}>Lihat Demo Langsung</button>
        </div>
        <div class="lp-hero-note">
          <span><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--c-primary)" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg> Tanpa kartu kredit</span>
          <span><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--c-primary)" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg> Setup 5 menit</span>
          <span><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--c-primary)" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg> Bahasa Indonesia</span>
        </div>
      </div>

      <div class="lp-mock">
        <div class="lp-mock-card">
          <div class="lp-mock-bar">
            <span class="lp-dot" style="background:#f0625c"></span>
            <span class="lp-dot" style="background:#f5bd4f"></span>
            <span class="lp-dot" style="background:#62c554"></span>
            <span style="margin-left:10px;font-size:12px;font-weight:700;color:var(--c-muted)">Dashboard — Suluk</span>
          </div>
          <div class="lp-mock-body">
            <div class="lp-mock-stats">
              <div class="lp-mock-stat"><div class="v">1.284</div><div class="l">Jamaah Aktif</div></div>
              <div class="lp-mock-stat"><div class="v" style="color:var(--c-primary)">Rp 4,6 M</div><div class="l">Pendapatan</div></div>
              <div class="lp-mock-stat"><div class="v">5</div><div class="l">Paket Aktif</div></div>
            </div>
            <div class="lp-mock-chart">
              <i style="height:42%"></i><i style="height:55%"></i><i style="height:48%"></i>
              <i style="height:70%"></i><i style="height:62%"></i><i style="height:88%"></i>
              <i style="height:78%"></i><i style="height:95%"></i>
            </div>
          </div>
        </div>
        <div class="lp-float lp-float-1">
          <span class="ic" style="background:var(--c-info-soft);color:var(--c-info)">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 7V5a2 2 0 0 1 2-2h2"/><path d="M17 3h2a2 2 0 0 1 2 2v2"/><path d="M21 17v2a2 2 0 0 1-2 2h-2"/><path d="M7 21H5a2 2 0 0 1-2-2v-2"/><line x1="7" y1="12" x2="17" y2="12"/></svg>
          </span>
          <span><div style="font-size:13px;font-weight:800">KTP terbaca</div><div class="l-sub" style="font-size:11.5px;color:var(--c-muted)">Akurasi 94% · 2 detik</div></span>
        </div>
        <div class="lp-float lp-float-2">
          <span class="ic" style="background:var(--c-success-soft);color:var(--c-success)">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M9 12l2 2 4-4"/></svg>
          </span>
          <span><div style="font-size:13px;font-weight:800">Pembayaran lunas</div><div class="l-sub" style="font-size:11.5px;color:var(--c-muted)">Rp 28,5 jt diterima</div></span>
        </div>
      </div>
    </div>
  </header>

  <!-- TRUST -->
  <section class="lp-trust">
    <div class="lp-container lp-trust-grid">
      <div class="lp-trust-item"><div class="v">500+</div><div class="l">Travel terdaftar</div></div>
      <div class="lp-trust-item"><div class="v">120rb+</div><div class="l">Jamaah dikelola</div></div>
      <div class="lp-trust-item"><div class="v">99,9%</div><div class="l">Uptime sistem</div></div>
      <div class="lp-trust-item"><div class="v">4,9/5</div><div class="l">Rating pengguna</div></div>
    </div>
  </section>

  <!-- AI FEATURE -->
  <section class="lp-sec lp-feature" id="fitur">
    <div class="lp-container lp-feature-grid">
      <div>
        <div class="lp-eyebrow">Fitur Unggulan</div>
        <h2 class="lp-h2">Pindai KTP &amp; Paspor dalam hitungan detik</h2>
        <p class="lp-sec-lead">Tidak perlu lagi mengetik data jamaah satu per satu. Cukup foto dokumen, AI Suluk mengekstrak seluruh data secara otomatis dan akurat.</p>
        <ul class="lp-feature-list">
          <li><span class="ck"><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg></span><span><b>OCR cerdas untuk KTP, KK &amp; Paspor</b><p>Ekstraksi NIK, nama, tanggal lahir, nomor paspor — langsung tervalidasi.</p></span></li>
          <li><span class="ck"><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg></span><span><b>Hemat 90% waktu input data</b><p>Daftarkan puluhan jamaah dalam menit, bukan jam.</p></span></li>
          <li><span class="ck"><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg></span><span><b>Minim kesalahan manusia</b><p>Data konsisten untuk manifest, visa, dan kontrak.</p></span></li>
        </ul>
      </div>
      <div class="lp-scan-card">
        <div class="lp-scan-top">
          <div class="lp-scan-line"></div>
          <div style="text-align:center;color:rgba(255,255,255,.55)">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="5" width="20" height="14" rx="2"/><line x1="2" y1="10" x2="22" y2="10"/></svg>
            <div style="font-size:12px;margin-top:10px;font-weight:600;letter-spacing:.05em">KARTU TANDA PENDUDUK</div>
          </div>
        </div>
        <div class="lp-scan-fields">
          <div class="lp-scan-field"><div class="l">NIK</div><div class="v">3273014509680001</div></div>
          <div class="lp-scan-field"><div class="l">Jenis Kelamin</div><div class="v">Laki-laki</div></div>
          <div class="lp-scan-field full"><div class="l">Nama Lengkap</div><div class="v">Muhammad Faisal Akbar</div></div>
          <div class="lp-scan-field"><div class="l">Tgl Lahir</div><div class="v">05-09-1997</div></div>
          <div class="lp-scan-field"><div class="l">Kota</div><div class="v">Medan</div></div>
        </div>
      </div>
    </div>
  </section>

  <!-- MODULES -->
  <section class="lp-sec" id="modul">
    <div class="lp-container">
      <div class="lp-sec-head">
        <div class="lp-eyebrow">Modul Lengkap</div>
        <h2 class="lp-h2">Semua kebutuhan operasional travel, di satu tempat</h2>
        <p class="lp-sec-lead">Dari calon jamaah hingga keberangkatan dan laporan keuangan — Suluk menyatukan seluruh alur kerja bisnis travel Anda.</p>
      </div>
      <div class="lp-modules">
        {#each modules as m}
          <div class="lp-mod">
            <div class="ic" style="background:{m.color}1c;color:{m.color}">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">{@html m.svg}</svg>
            </div>
            <h3>{m.title}</h3>
            <p>{m.desc}</p>
          </div>
        {/each}
      </div>
    </div>
  </section>

  <!-- HOW IT WORKS -->
  <section class="lp-sec" id="cara" style="background:var(--c-bg)">
    <div class="lp-container">
      <div class="lp-sec-head">
        <div class="lp-eyebrow">Cara Kerja</div>
        <h2 class="lp-h2">Mulai dalam 4 langkah mudah</h2>
      </div>
      <div class="lp-steps">
        <div class="lp-step"><div class="num">1</div><h3>Daftar &amp; atur travel</h3><p>Buat akun, lengkapi profil travel dan tim Anda dalam beberapa menit.</p></div>
        <div class="lp-step"><div class="num">2</div><h3>Input jamaah</h3><p>Scan KTP &amp; paspor dengan AI atau buka link pendaftaran mandiri.</p></div>
        <div class="lp-step"><div class="num">3</div><h3>Kelola operasional</h3><p>Atur paket, invoice, rooming, kontrak, dan keuangan dari satu dasbor.</p></div>
        <div class="lp-step"><div class="num">4</div><h3>Pantau pertumbuhan</h3><p>Laporan real-time untuk pendapatan, jamaah, dan performa agen.</p></div>
      </div>
    </div>
  </section>

  <!-- USE CASES -->
  <section class="lp-sec" id="usecase">
    <div class="lp-container">
      <div class="lp-sec-head">
        <div class="lp-eyebrow">Cocok Untuk Siapa</div>
        <h2 class="lp-h2">Tumbuh bersama skala travel Anda</h2>
        <p class="lp-sec-lead">Apa pun ukuran travel Anda, Suluk menyesuaikan diri — dari yang baru mulai hingga yang mengelola ribuan jamaah per tahun.</p>
      </div>
      <div class="lp-usecases">
        <div class="lp-uc">
          <div class="lp-uc-top" style="--uc:#0f7a5a">
            <span class="lp-uc-ic"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 21h18"/><path d="M5 21V7l8-4v18"/><path d="M19 21V11l-6-4"/></svg></span>
            <h3>Travel Kecil</h3>
            <p class="lp-uc-scale">1–3 keberangkatan / tahun</p>
          </div>
          <p class="lp-uc-desc">Baru memulai dan ingin langsung rapi tanpa biaya besar.</p>
          <div class="lp-uc-label">Tantangan</div>
          <p class="lp-uc-txt">Data jamaah masih di Excel, sulit pantau pembayaran.</p>
          <div class="lp-uc-label">Fitur utama</div>
          <p class="lp-uc-txt">Data Jamaah, AI Scanner, Paket dasar.</p>
          <div class="lp-uc-rec">Rekomendasi: <strong>Paket Pemula (Gratis)</strong></div>
        </div>
        <div class="lp-uc">
          <div class="lp-uc-top" style="--uc:#c79a3e">
            <span class="lp-uc-ic"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="2" width="16" height="20" rx="2"/><path d="M9 22v-4h6v4"/><path d="M8 6h.01M16 6h.01M12 6h.01M12 10h.01M16 10h.01M8 10h.01"/></svg></span>
            <h3>Travel Menengah</h3>
            <p class="lp-uc-scale">4–12 keberangkatan / tahun</p>
          </div>
          <p class="lp-uc-desc">Operasional makin kompleks, butuh otomatisasi penuh.</p>
          <div class="lp-uc-label">Tantangan</div>
          <p class="lp-uc-txt">Banyak invoice, kontrak, dan rooming yang harus dikelola.</p>
          <div class="lp-uc-label">Fitur utama</div>
          <p class="lp-uc-txt">Invoice, CRM, Keuangan, Kontrak, Rooming.</p>
          <div class="lp-uc-rec">Rekomendasi: <strong>Paket Pro</strong></div>
        </div>
        <div class="lp-uc">
          <div class="lp-uc-top" style="--uc:#2563c9">
            <span class="lp-uc-ic"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 21h18"/><path d="M5 21V5a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v16"/><path d="M19 21V9a2 2 0 0 0-2-2h-2"/><path d="M9 7h2M9 11h2M9 15h2"/></svg></span>
            <h3>Travel Besar</h3>
            <p class="lp-uc-scale">12+ keberangkatan / tahun</p>
          </div>
          <p class="lp-uc-desc">Multi-cabang dengan ratusan jamaah dan tim besar.</p>
          <div class="lp-uc-label">Tantangan</div>
          <p class="lp-uc-txt">Konsolidasi cabang, kontrol akses, laporan menyeluruh.</p>
          <div class="lp-uc-label">Fitur utama</div>
          <p class="lp-uc-txt">Semua modul, multi-cabang, payroll, API.</p>
          <div class="lp-uc-rec">Rekomendasi: <strong>Paket Enterprise</strong></div>
        </div>
      </div>
    </div>
  </section>

  <!-- PRICING -->
  <section class="lp-sec" id="harga">
    <div class="lp-container">
      <div class="lp-sec-head">
        <div class="lp-eyebrow">Harga</div>
        <h2 class="lp-h2">Paket yang tumbuh bersama travel Anda</h2>
        <p class="lp-sec-lead">Mulai gratis, tingkatkan saat bisnis berkembang. Tanpa biaya tersembunyi.</p>
      </div>
      <div class="lp-pricing">
        {#each PLANS as p}
          <div class="lp-price {p.popular ? 'pop' : ''}">
            {#if p.popular}<span class="lp-price-badge">Paling Populer</span>{/if}
            <h3>{p.name}</h3>
            <p class="desc">{p.desc}</p>
            <div>
              <span class="amt">{priceText(p)}</span>
              {#if p.purchasable}<span class="per">/ bulan</span>{/if}
            </div>
            <ul>
              {#each p.features as f}
                <li><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg> {f}</li>
              {/each}
            </ul>
            <button
              class="lp-btn {p.popular ? 'lp-btn-primary' : 'lp-btn-ghost'}"
              onclick={() => handlePlanCta(p)}
            >{p.cta}</button>
          </div>
        {/each}
      </div>
    </div>
  </section>

  <!-- TESTIMONIALS -->
  <section class="lp-sec" id="testimoni" style="background:var(--c-bg)">
    <div class="lp-container">
      <div class="lp-sec-head">
        <div class="lp-eyebrow">Testimoni</div>
        <h2 class="lp-h2">Dipercaya travel di seluruh Indonesia</h2>
      </div>
      <div class="lp-testi">
        {#each testimonials as t}
          <div class="lp-tcard">
            <div class="lp-stars">
              {#each Array(5) as _}<svg width="17" height="17" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M11.5 2.3a.5.5 0 0 1 .9 0l2.3 4.6 5.1.7a.5.5 0 0 1 .3.9l-3.7 3.6.9 5.1a.5.5 0 0 1-.8.5L12 15.9l-4.6 2.4a.5.5 0 0 1-.8-.5l.9-5.1L3.8 9.1a.5.5 0 0 1 .3-.9l5.1-.7z"/></svg>{/each}
            </div>
            <blockquote>“{t.quote}”</blockquote>
            <div class="lp-tauthor">
              <div class="lp-tavatar" style="background:{t.color}22;color:{t.color}">{initials(t.name)}</div>
              <div><div class="n">{t.name}</div><div class="r">{t.role}</div></div>
            </div>
          </div>
        {/each}
      </div>
    </div>
  </section>

  <!-- APLIKASI MOBILE -->
  <section class="lp-sec" id="aplikasi">
    <div class="lp-container">
      <div class="lp-app-band">
        <div class="lp-app-copy">
          <span class="lp-app-tag">APLIKASI MOBILE · FITUR PRO</span>
          <h2>Kelola travel dari genggaman</h2>
          <p>Scan KTP/paspor pakai kamera, catat pembayaran, setujui refund &amp; diskon, dan pantau jamaah — langsung dari ponsel, di mana saja. Tersedia untuk paket Pro ke atas.</p>
          <div class="lp-app-actions">
            <a class="lp-btn lp-btn-white lp-btn-lg" href="/unduh">⬇ Unduh Aplikasi</a>
            <a class="lp-btn lp-btn-lg lp-app-ghost" href="/mobile">Buka Web App →</a>
          </div>
        </div>
        <div class="lp-app-art" aria-hidden="true">
          <div class="lp-phone">
            <div class="lp-phone-notch"></div>
            <div class="lp-phone-screen">
              <img src="/brand/suluk-mark-white.png" alt="" />
              <div class="lp-phone-name">Suluk</div>
              <div class="lp-phone-tag">ERP FOR TRAVEL</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- CTA -->
  <section class="lp-cta-sec">
    <div class="lp-container">
      <div class="lp-cta-box">
        <h2>Siap memodernkan travel Anda?</h2>
        <p>Bergabung dengan 500+ travel yang sudah mengelola jamaah, keuangan, dan operasional lebih mudah bersama Suluk.</p>
        <div class="lp-cta-actions">
          <button class="lp-btn lp-btn-white lp-btn-lg" onclick={() => onGoToRegister()}>Coba Gratis 14 Hari</button>
          <button class="lp-btn lp-btn-lg" onclick={() => onGoToLogin()} style="background:rgba(255,255,255,.14);color:#fff">Jadwalkan Demo</button>
        </div>
      </div>
    </div>
  </section>

  <!-- FOOTER -->
  <footer class="lp-footer">
    <div class="lp-container">
      <div class="lp-footer-grid">
        <div class="lp-footer-about">
          <a class="lp-brand" href="#top" style="margin-bottom:0" onclick={(e) => scrollTop(e)}>
            <span class="lp-brand-mark"><img src="/brand/suluk-mark-white.png" alt="Suluk" style="height:42px;width:auto;display:block" /></span>
            <span><span class="lp-brand-name" style="color:#fff">Suluk</span><div class="lp-brand-sub">ERP FOR TRAVEL</div></span>
          </a>
          <p>Sistem manajemen all-in-one untuk bisnis travel umrah &amp; haji di Indonesia. Otomatis, terintegrasi, dan dibuat untuk pertumbuhan.</p>
        </div>
        <div><h4>Fitur</h4><ul><li><button class="lp-foot-link" onclick={() => onNavigate("software")}>Software Travel</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("fitur-invoice")}>Invoice</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("fitur-crm")}>CRM Jamaah</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("fitur-keuangan")}>Laporan Keuangan</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("fitur-kontrak")}>E-Kontrak</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("fitur-payroll")}>Payroll</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("unduh")}>Aplikasi Mobile</button></li></ul></div>
        <div><h4>Perusahaan</h4><ul><li><button class="lp-foot-link" onclick={() => onNavigate("about")}>Tentang</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("contact")}>Kontak</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("privacy")}>Kebijakan Privasi</button></li><li><button class="lp-foot-link" onclick={() => onNavigate("terms")}>Syarat &amp; Ketentuan</button></li></ul></div>
        <div><h4>Akun</h4><ul><li><button class="lp-foot-link" onclick={() => onGoToLogin()}>Masuk</button></li><li><button class="lp-foot-link" onclick={() => onGoToRegister()}>Daftar Gratis</button></li></ul></div>
      </div>
      <div class="lp-footer-bottom">
        <div>© 2026 Suluk — PT Suluk Mitra Barokah. Hak cipta dilindungi.</div>
        <div>Dibuat dengan ❤️ untuk travel umrah &amp; haji Indonesia</div>
      </div>
    </div>
  </footer>
</div>

<style>
  .lp { font-family: var(--font-ui); color: var(--c-ink); background: var(--c-surface); overflow-x: hidden; }
  .lp :global(*) { box-sizing: border-box; }
  .lp-container { max-width: 1180px; margin: 0 auto; padding: 0 24px; }
  .lp :global(img) { max-width: 100%; display: block; }
  .lp section { position: relative; }

  .lp-btn { display: inline-flex; align-items: center; justify-content: center; gap: 9px; font-weight: 700; border-radius: 12px; cursor: pointer; transition: transform .12s, box-shadow .15s, filter .15s; white-space: nowrap; text-align: center; line-height: 1; border: none; }
  .lp-btn-primary { background: var(--c-primary); color: #fff; padding: 14px 24px; font-size: 15px; box-shadow: 0 8px 24px -8px var(--c-primary); }
  .lp-btn-primary:hover { transform: translateY(-2px); filter: brightness(1.05); }
  .lp-btn-ghost { background: transparent; color: var(--c-ink); padding: 14px 22px; font-size: 15px; border: 1px solid var(--c-line); }
  .lp-btn-ghost:hover { background: var(--c-bg); border-color: var(--c-faint); }
  .lp-btn-white { background: #fff; color: var(--c-primary-deep); padding: 14px 26px; font-size: 15px; }
  .lp-btn-white:hover { transform: translateY(-2px); }
  .lp-btn-lg { padding: 16px 30px; font-size: 16px; }

  /* Brand mark styles — reused by the footer brand link. (The top navbar now
     lives in the shared LandingNav component.) */
  .lp-brand { display: flex; align-items: center; gap: 11px; flex-shrink: 0; }
  .lp-brand-mark { display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
  .lp-brand-name { font-family: var(--font-display); font-size: 23px; font-weight: 800; line-height: 1; }
  .lp-brand-sub { font-size: 9.5px; font-weight: 700; letter-spacing: .17em; color: var(--c-accent); margin-top: 4px; }

  .lp-hero { padding: 80px 0 90px; background: radial-gradient(1200px 500px at 80% -10%, var(--c-primary-tint), transparent 60%), radial-gradient(900px 500px at 0% 0%, var(--c-accent-soft), transparent 55%); }
  .lp-hero-grid { display: grid; grid-template-columns: 1.05fr 1fr; gap: 56px; align-items: center; }
  .lp-kicker { display: inline-flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 700; color: var(--c-primary-deep); background: var(--c-primary-soft); padding: 7px 14px; border-radius: 999px; margin-bottom: 22px; }
  .lp-h1 { font-family: var(--font-display); font-size: 54px; line-height: 1.08; font-weight: 800; letter-spacing: -.02em; margin: 0 0 20px; text-wrap: balance; }
  .lp-h1 em { font-style: italic; color: var(--c-primary); }
  .lp-lead { font-size: 18px; line-height: 1.6; color: var(--c-muted); max-width: 520px; margin: 0 0 30px; }
  .lp-hero-cta { display: flex; gap: 14px; flex-wrap: wrap; }
  .lp-hero-note { margin-top: 18px; font-size: 13.5px; color: var(--c-faint); display: flex; align-items: center; gap: 16px; flex-wrap: wrap; }
  .lp-hero-note span { display: inline-flex; align-items: center; gap: 6px; }

  .lp-mock { position: relative; }
  .lp-mock-card { background: var(--c-surface); border: 1px solid var(--c-line); border-radius: 20px; box-shadow: 0 30px 70px -28px rgba(16,33,28,.35); overflow: hidden; }
  .lp-mock-bar { display: flex; align-items: center; gap: 7px; padding: 13px 16px; border-bottom: 1px solid var(--c-line-soft); background: var(--c-bg); }
  .lp-dot { width: 11px; height: 11px; border-radius: 999px; }
  .lp-mock-body { padding: 18px; }
  .lp-mock-stats { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 12px; margin-bottom: 16px; }
  .lp-mock-stat { background: var(--c-bg); border-radius: 12px; padding: 13px; }
  .lp-mock-stat .v { font-size: 20px; font-weight: 800; letter-spacing: -.02em; }
  .lp-mock-stat .l { font-size: 11px; color: var(--c-muted); margin-top: 3px; }
  .lp-mock-chart { height: 130px; background: var(--c-primary-tint); border-radius: 12px; padding: 14px; display: flex; align-items: flex-end; gap: 9px; }
  .lp-mock-chart i { flex: 1; background: var(--c-primary); border-radius: 5px 5px 2px 2px; display: block; opacity: .85; }
  .lp-float { position: absolute; background: var(--c-surface); border: 1px solid var(--c-line); border-radius: 14px; box-shadow: 0 18px 44px -18px rgba(16,33,28,.3); padding: 13px 15px; display: flex; align-items: center; gap: 11px; }
  .lp-float-1 { top: -22px; right: -18px; }
  .lp-float-2 { bottom: -24px; left: -22px; }
  .lp-float .ic { width: 38px; height: 38px; border-radius: 10px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
  .lp-float > span { white-space: nowrap; }

  .lp-trust { padding: 30px 0; border-top: 1px solid var(--c-line); border-bottom: 1px solid var(--c-line); background: var(--c-bg); }
  .lp-trust-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 24px; }
  .lp-trust-item { text-align: center; }
  .lp-trust-item .v { font-size: 32px; font-weight: 800; letter-spacing: -.02em; color: var(--c-primary-deep); }
  .lp-trust-item .l { font-size: 13.5px; color: var(--c-muted); margin-top: 4px; }

  .lp-sec { padding: 88px 0; scroll-margin-top: 84px; }
  .lp-sec-head { text-align: center; max-width: 640px; margin: 0 auto 52px; }
  .lp-eyebrow { font-size: 13px; font-weight: 700; letter-spacing: .1em; text-transform: uppercase; color: var(--c-primary); margin-bottom: 14px; }
  .lp-h2 { font-family: var(--font-display); font-size: 40px; line-height: 1.12; font-weight: 800; letter-spacing: -.02em; margin: 0 0 16px; text-wrap: balance; }
  .lp-sec-lead { font-size: 17px; line-height: 1.6; color: var(--c-muted); }

  .lp-feature { background: var(--c-bg); }
  .lp-feature-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 56px; align-items: center; }
  .lp-feature-list { margin: 26px 0 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 16px; }
  .lp-feature-list li { display: flex; gap: 13px; align-items: flex-start; }
  .lp-feature-list .ck { width: 26px; height: 26px; border-radius: 8px; background: var(--c-primary-soft); color: var(--c-primary-deep); display: flex; align-items: center; justify-content: center; flex-shrink: 0; margin-top: 1px; }
  .lp-feature-list b { font-weight: 700; }
  .lp-feature-list p { margin: 3px 0 0; color: var(--c-muted); font-size: 14px; line-height: 1.5; }
  .lp-scan-card { background: var(--c-surface); border: 1px solid var(--c-line); border-radius: 20px; box-shadow: var(--shadow-lg); overflow: hidden; }
  .lp-scan-top { background: linear-gradient(135deg, #1a3d33, #0c2b22); padding: 30px; position: relative; min-height: 200px; display: flex; align-items: center; justify-content: center; }
  .lp-scan-line { position: absolute; left: 24px; right: 24px; height: 3px; background: var(--c-accent); box-shadow: 0 0 16px 3px var(--c-accent); border-radius: 2px; animation: lpScan 2.6s ease-in-out infinite; }
  @keyframes lpScan { 0%,100% { top: 30px; } 50% { top: 170px; } }
  .lp-scan-fields { padding: 18px; display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
  .lp-scan-field { background: var(--c-bg); border-radius: 10px; padding: 11px 13px; }
  .lp-scan-field .l { font-size: 10px; font-weight: 700; letter-spacing: .04em; text-transform: uppercase; color: var(--c-faint); }
  .lp-scan-field .v { font-size: 13px; font-weight: 700; margin-top: 3px; }
  .lp-scan-field.full { grid-column: 1 / -1; }

  .lp-modules { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
  .lp-mod { padding: 26px; border: 1px solid var(--c-line); border-radius: 18px; background: var(--c-surface); transition: transform .18s, box-shadow .18s, border-color .18s; }
  .lp-mod:hover { transform: translateY(-4px); box-shadow: 0 20px 44px -22px rgba(16,33,28,.28); border-color: var(--c-primary-soft); }
  .lp-mod .ic { width: 50px; height: 50px; border-radius: 14px; display: flex; align-items: center; justify-content: center; margin-bottom: 16px; }
  .lp-mod h3 { font-size: 17px; font-weight: 800; margin: 0 0 8px; }
  .lp-mod p { font-size: 14px; line-height: 1.55; color: var(--c-muted); margin: 0; }

  .lp-steps { display: grid; grid-template-columns: repeat(4, 1fr); gap: 24px; }
  .lp-step { text-align: left; }
  .lp-step .num { width: 46px; height: 46px; border-radius: 14px; background: var(--c-primary); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 19px; font-weight: 800; margin-bottom: 16px; }
  .lp-step h3 { font-size: 17px; font-weight: 800; margin: 0 0 8px; }
  .lp-step p { font-size: 14px; line-height: 1.55; color: var(--c-muted); margin: 0; }

  .lp-pricing { display: grid; grid-template-columns: repeat(5, 1fr); gap: 16px; align-items: stretch; }
  .lp-price { border: 1px solid var(--c-line); border-radius: 20px; padding: 24px; background: var(--c-surface); display: flex; flex-direction: column; }
  .lp-price.pop { border: 2px solid var(--c-primary); box-shadow: 0 28px 60px -28px var(--c-primary); position: relative; }
  .lp-price-badge { position: absolute; top: -13px; left: 50%; transform: translateX(-50%); background: var(--c-primary); color: #fff; font-size: 12px; font-weight: 700; padding: 5px 14px; border-radius: 999px; }
  .lp-price h3 { font-size: 19px; font-weight: 800; margin: 0 0 6px; }
  .lp-price .desc { font-size: 13.5px; color: var(--c-muted); margin: 0 0 20px; min-height: 38px; }
  .lp-price .amt { font-size: 40px; font-weight: 800; letter-spacing: -.03em; line-height: 1; }
  .lp-price .per { font-size: 14px; color: var(--c-muted); font-weight: 600; }
  .lp-price ul { list-style: none; padding: 0; margin: 24px 0 26px; display: flex; flex-direction: column; gap: 13px; flex: 1; }
  .lp-price li { display: flex; gap: 11px; align-items: flex-start; font-size: 14px; color: var(--c-ink-soft); }
  .lp-price li svg { color: var(--c-primary); flex-shrink: 0; margin-top: 2px; }

  .lp-testi { display: grid; grid-template-columns: repeat(3, 1fr); gap: 22px; }
  .lp-tcard { background: var(--c-surface); border: 1px solid var(--c-line); border-radius: 18px; padding: 28px; }
  .lp-stars { display: flex; gap: 3px; color: var(--c-accent); margin-bottom: 16px; }
  .lp-tcard blockquote { margin: 0 0 22px; font-size: 15.5px; line-height: 1.62; color: var(--c-ink); }
  .lp-tauthor { display: flex; align-items: center; gap: 12px; }
  .lp-tavatar { width: 46px; height: 46px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 16px; flex-shrink: 0; }
  .lp-tauthor .n { font-weight: 700; font-size: 14.5px; }
  .lp-tauthor .r { font-size: 13px; color: var(--c-muted); margin-top: 2px; }

  .lp-usecases { display: grid; grid-template-columns: repeat(3, 1fr); gap: 22px; }
  .lp-uc { border: 1px solid var(--c-line); border-radius: 20px; padding: 26px; background: var(--c-surface); transition: box-shadow .18s, transform .18s; }
  .lp-uc:hover { box-shadow: 0 20px 46px -24px rgba(16,33,28,.28); transform: translateY(-4px); }
  .lp-uc-top { padding-bottom: 16px; margin-bottom: 16px; border-bottom: 1px solid var(--c-line-soft); }
  .lp-uc-ic { width: 48px; height: 48px; border-radius: 13px; background: color-mix(in srgb, var(--uc) 14%, #fff); color: var(--uc); display: flex; align-items: center; justify-content: center; margin-bottom: 14px; }
  .lp-uc-top h3 { font-size: 19px; font-weight: 800; margin: 0; }
  .lp-uc-scale { font-size: 13px; color: var(--c-muted); margin: 4px 0 0; font-weight: 600; }
  .lp-uc-desc { font-size: 14.5px; color: var(--c-ink-soft); line-height: 1.55; margin: 0 0 16px; }
  .lp-uc-label { font-size: 11px; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; color: var(--c-faint); margin-bottom: 4px; }
  .lp-uc-txt { font-size: 13.5px; color: var(--c-muted); line-height: 1.5; margin: 0 0 14px; }
  .lp-uc-rec { font-size: 13.5px; color: var(--c-ink-soft); padding-top: 14px; border-top: 1px solid var(--c-line-soft); }
  .lp-uc-rec strong { color: var(--c-primary-deep); }

  .lp-cta-sec { padding: 30px 0 100px; }
  .lp-cta-box { background: linear-gradient(135deg, var(--c-primary-deep), var(--c-primary)); border-radius: 28px; padding: 64px 48px; text-align: center; color: #fff; position: relative; overflow: hidden; }
  .lp-cta-box::after { content: ''; position: absolute; width: 360px; height: 360px; border-radius: 50%; background: rgba(255,255,255,.07); right: -90px; top: -120px; }
  .lp-cta-box h2 { font-family: var(--font-display); font-size: 40px; font-weight: 800; margin: 0 0 16px; letter-spacing: -.02em; position: relative; }
  .lp-cta-box p { font-size: 17px; opacity: .9; max-width: 520px; margin: 0 auto 30px; position: relative; }
  .lp-cta-actions { display: flex; gap: 14px; justify-content: center; flex-wrap: wrap; position: relative; }

  .lp-footer { background: var(--c-sidebar-bg); color: #cfe0d9; padding: 60px 0 30px; }
  .lp-footer-grid { display: grid; grid-template-columns: 1.6fr 1fr 1fr 1fr; gap: 40px; padding-bottom: 40px; border-bottom: 1px solid rgba(255,255,255,.1); }
  .lp-footer h4 { font-size: 13px; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; color: #fff; margin: 0 0 18px; }
  .lp-footer ul { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 12px; }
  .lp-footer a, .lp-foot-link { font-size: 14px; color: #a9c2b8; transition: color .15s; background: none; border: none; padding: 0; cursor: pointer; text-align: left; font-family: inherit; }
  .lp-footer a:hover, .lp-foot-link:hover { color: #fff; }
  .lp-footer-about p { font-size: 14px; line-height: 1.6; color: #a9c2b8; margin: 16px 0 0; max-width: 300px; }
  .lp-footer-bottom { padding-top: 26px; display: flex; justify-content: space-between; align-items: center; gap: 16px; font-size: 13px; color: #88a298; flex-wrap: wrap; }

  @media (max-width: 1024px) {
    .lp-h1 { font-size: 44px; }
    .lp-hero-grid { gap: 40px; }
    .lp-modules { grid-template-columns: repeat(2, 1fr); }
    .lp-steps { grid-template-columns: repeat(2, 1fr); row-gap: 36px; }
    .lp-footer-grid { grid-template-columns: 1fr 1fr; row-gap: 36px; }
  }

  @media (max-width: 1180px) and (min-width: 761px) {
    .lp-pricing { grid-template-columns: repeat(3, 1fr); }
  }

  @media (max-width: 760px) {
    .lp-container { padding: 0 18px; }
    .lp-hero { padding: 48px 0 56px; }
    .lp-hero-grid { grid-template-columns: 1fr; gap: 54px; }
    .lp-h1 { font-size: 36px; }
    .lp-lead { font-size: 16px; }
    .lp-hero-cta { flex-direction: column; }
    .lp-hero-cta .lp-btn { width: 100%; }
    .lp-float-1 { top: -16px; right: 0; padding: 10px 12px; }
    .lp-float-2 { bottom: -16px; left: 0; padding: 10px 12px; }
    .lp-float .l-sub { display: none; }
    .lp-trust-grid { grid-template-columns: 1fr 1fr; row-gap: 24px; }
    .lp-trust-item .v { font-size: 27px; }
    .lp-sec { padding: 60px 0; }
    .lp-h2, .lp-cta-box h2 { font-size: 30px; }
    .lp-sec-head { margin-bottom: 38px; }
    .lp-feature-grid { grid-template-columns: 1fr; gap: 38px; }
    .lp-modules { grid-template-columns: 1fr; }
    .lp-steps { grid-template-columns: 1fr; }
    .lp-usecases { grid-template-columns: 1fr; }
    .lp-pricing { grid-template-columns: 1fr; }
    .lp-price.pop { order: -1; }
    .lp-testi { grid-template-columns: 1fr; }
    .lp-cta-box { padding: 44px 24px; }
    .lp-cta-box p { font-size: 15.5px; }
    .lp-cta-actions .lp-btn { width: 100%; }
    .lp-footer-grid { grid-template-columns: 1fr; gap: 30px; }
    .lp-footer-bottom { flex-direction: column; align-items: flex-start; }
    .lp-app-band { grid-template-columns: 1fr; padding: 32px 24px; text-align: center; }
    .lp-app-actions { justify-content: center; }
    .lp-app-art { display: none; }
  }

  /* Aplikasi Mobile section */
  a.lp-btn { text-decoration: none; }
  .lp-app-band {
    display: grid; grid-template-columns: 1.2fr .8fr; gap: 28px; align-items: center;
    background: linear-gradient(135deg, var(--c-primary-deep), var(--c-primary));
    border-radius: 28px; padding: 44px 48px; color: #fff; overflow: hidden;
  }
  .lp-app-tag { font-size: 12px; font-weight: 800; letter-spacing: .14em; color: var(--c-accent); }
  .lp-app-band h2 { font-family: var(--font-display, Georgia, serif); font-size: 30px; font-weight: 800; margin: 12px 0 10px; letter-spacing: -.02em; }
  .lp-app-band p { font-size: 15.5px; line-height: 1.65; opacity: .92; max-width: 460px; }
  .lp-app-actions { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 22px; }
  .lp-app-ghost { background: rgba(255,255,255,.14); color: #fff; }
  .lp-app-ghost:hover { background: rgba(255,255,255,.22); }
  .lp-app-art { display: flex; justify-content: center; }
  .lp-phone {
    width: 168px; height: 320px; border-radius: 30px; background: #0a2a1f;
    border: 6px solid rgba(255,255,255,.16); position: relative; box-shadow: 0 24px 60px -20px rgba(0,0,0,.5);
    display: flex; align-items: center; justify-content: center;
  }
  .lp-phone-notch { position: absolute; top: 10px; left: 50%; transform: translateX(-50%); width: 54px; height: 7px; border-radius: 999px; background: rgba(255,255,255,.25); }
  .lp-phone-screen { text-align: center; }
  .lp-phone-screen img { height: 60px; width: auto; }
  .lp-phone-name { font-family: var(--font-display, Georgia, serif); font-size: 22px; font-weight: 800; color: #fff; margin-top: 12px; }
  .lp-phone-tag { font-size: 9px; font-weight: 700; letter-spacing: .18em; color: var(--c-accent); margin-top: 3px; }
</style>
