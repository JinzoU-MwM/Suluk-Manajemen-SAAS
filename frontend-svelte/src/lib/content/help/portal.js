// Konten Pusat Bantuan untuk area JAMAAH (/portal).
// Audiens: calon jamaah (pengguna awam) — gunakan bahasa sederhana dan menenangkan.
// Hanya dibaca oleh helper di index.js saat area === "portal".
//
// TODO(domain): langkah unggah dokumen & alur pembayaran mengikuti istilah umum;
// verifikasi label tombol portal (Dokumen, Visa, Profil) bila UI berubah.

/** @type {import('./index.js').HelpGuide[]} */
export const PORTAL_GUIDES = [
  {
    slug: "masuk-dan-beranda",
    title: "Masuk ke Portal & Memeriksa Status",
    category: "Memulai",
    order: 1,
    summary:
      "Cara masuk ke portal jamaah dan membaca status keberangkatan, dokumen, serta pembayaran Anda di halaman Beranda.",
    keywords: ["masuk", "login", "beranda", "status", "portal", "akun"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Portal Jemaah adalah tempat Anda memantau persiapan keberangkatan secara mandiri, kapan saja, dari ponsel maupun komputer." },
      { type: "h2", text: "Langkah masuk" },
      {
        type: "ol",
        items: [
          "Buka tautan portal yang diberikan travel Anda.",
          "Masukkan nomor/email dan kata sandi yang terdaftar.",
          "Anda akan diarahkan ke halaman Beranda berisi ringkasan status.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Di Beranda Anda dapat melihat sekilas: kelengkapan dokumen, status visa, dan sisa pembayaran.",
      },
    ],
    related: ["mengunggah-dokumen", "memantau-status-visa"],
  },
  {
    slug: "mengunggah-dokumen",
    title: "Mengunggah Dokumen (KTP, Paspor, Foto)",
    category: "Dokumen & Visa",
    order: 1,
    summary:
      "Panduan menyiapkan dan mengunggah berkas persyaratan umrah agar proses keberangkatan Anda tidak tertunda.",
    keywords: ["dokumen", "unggah", "upload", "ktp", "paspor", "foto", "berkas", "persyaratan"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Travel membutuhkan dokumen Anda untuk mengurus visa dan administrasi. Mengunggahnya lebih awal mempercepat keberangkatan." },
      { type: "h2", text: "Cara mengunggah" },
      {
        type: "ol",
        items: [
          "Buka menu Dokumen di portal.",
          "Pilih jenis dokumen yang diminta (mis. KTP, paspor, foto).",
          "Ambil foto yang jelas atau pilih berkas dari perangkat, lalu unggah.",
          "Pastikan statusnya berubah menjadi 'terunggah'.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Pastikan seluruh sisi dokumen terbaca jelas dan tidak buram agar tidak perlu mengulang unggahan.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Masa berlaku paspor minimal mengikuti ketentuan keberangkatan. Bila ragu, tanyakan ke travel Anda.",
      },
    ],
    related: ["memantau-status-visa", "melengkapi-profil"],
  },
  {
    slug: "memantau-status-visa",
    title: "Memantau Status Visa",
    category: "Dokumen & Visa",
    order: 2,
    summary:
      "Memahami arti setiap status visa di portal sehingga Anda tahu sejauh mana proses keberangkatan berjalan.",
    keywords: ["visa", "status", "proses", "keberangkatan"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Menu Visa menampilkan perkembangan pengurusan visa Anda. Statusnya diperbarui oleh travel seiring proses berjalan." },
      {
        type: "callout",
        variant: "info",
        text: "Jika status tidak berubah dalam waktu lama atau Anda merasa ada yang kurang, hubungi travel Anda untuk konfirmasi.",
      },
    ],
    related: ["mengunggah-dokumen", "masuk-dan-beranda"],
  },
  {
    slug: "melengkapi-profil",
    title: "Melengkapi & Memperbarui Profil",
    category: "Akun",
    order: 1,
    summary:
      "Menjaga data pribadi Anda tetap akurat — nama sesuai paspor, kontak, dan kontak darurat.",
    keywords: ["profil", "akun", "data diri", "kontak", "ubah"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Data profil yang akurat membantu travel menghubungi Anda dan menyiapkan dokumen tanpa kesalahan." },
      { type: "h2", text: "Yang perlu dipastikan" },
      {
        type: "ul",
        items: [
          "Nama lengkap ditulis persis seperti di paspor.",
          "Nomor telepon dan email aktif.",
          "Kontak darurat yang bisa dihubungi.",
        ],
      },
      {
        type: "callout",
        variant: "warning",
        text: "Perbedaan ejaan nama dengan paspor dapat menghambat penerbitan visa. Periksa kembali sebelum menyimpan.",
      },
    ],
    related: ["mengunggah-dokumen"],
  },
];
