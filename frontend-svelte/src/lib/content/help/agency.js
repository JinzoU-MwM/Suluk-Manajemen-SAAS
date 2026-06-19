// Konten Pusat Bantuan untuk area AGEN (/agency).
// Audiens: mitra agen penjualan — fokus pada lead, jaringan, dan komisi.
// Hanya dibaca oleh helper di index.js saat area === "agency".
//
// TODO(domain): mekanisme pencairan komisi & struktur jaringan mengikuti istilah
// umum; verifikasi label menu (Lead Saya, Komisi Saya, Jaringan) bila UI berubah.

/** @type {import('./index.js').HelpGuide[]} */
export const AGENCY_GUIDES = [
  {
    slug: "mengenal-dashboard-agen",
    title: "Mengenal Dashboard Agen",
    category: "Memulai",
    order: 1,
    summary:
      "Orientasi singkat Portal Agen: tempat memantau lead, jaringan, dan komisi Anda dalam satu tampilan.",
    keywords: ["dashboard", "agen", "mulai", "beranda", "portal agen"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Portal Agen membantu Anda mengelola calon jamaah (lead) yang Anda bawa, memantau jaringan, dan melihat komisi yang Anda peroleh." },
      { type: "h2", text: "Menu utama" },
      {
        type: "ul",
        items: [
          "Lead Saya — daftar calon jamaah yang Anda daftarkan.",
          "Jaringan — struktur agen dan jamaah di bawah Anda.",
          "Komisi Saya — perolehan komisi dan status pencairan.",
        ],
      },
    ],
    related: ["mengelola-leads", "melihat-komisi"],
  },
  {
    slug: "mengelola-leads",
    title: "Menambah & Mengelola Lead",
    category: "Leads & Jamaah",
    order: 1,
    summary:
      "Cara mendaftarkan calon jamaah sebagai lead dan menindaklanjutinya hingga menjadi jamaah yang berangkat.",
    keywords: ["lead", "prospek", "calon jamaah", "daftar", "follow up", "tindak lanjut"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Lead adalah calon jamaah yang Anda bawa. Mencatatnya di portal memastikan tidak ada prospek yang terlewat dan komisi tercatat dengan benar." },
      { type: "h2", text: "Menambahkan lead" },
      {
        type: "ol",
        items: [
          "Buka menu Lead Saya, lalu klik Tambah Lead.",
          "Isi nama dan kontak calon jamaah.",
          "Catat paket yang diminati bila sudah ada.",
          "Simpan, lalu tindak lanjuti secara berkala hingga calon jamaah mendaftar.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Perbarui status setiap lead agar mudah melihat mana yang perlu dihubungi kembali.",
      },
    ],
    related: ["memantau-jaringan", "melihat-komisi"],
  },
  {
    slug: "memantau-jaringan",
    title: "Memantau Jaringan Agen",
    category: "Leads & Jamaah",
    order: 2,
    summary:
      "Melihat struktur agen dan jamaah di bawah Anda untuk memahami sumber perolehan dan komisi.",
    keywords: ["jaringan", "network", "downline", "struktur", "tim"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Menu Jaringan menampilkan agen dan jamaah yang terhubung dengan akun Anda, membantu memahami dari mana komisi berasal." },
      {
        type: "callout",
        variant: "info",
        text: "Gunakan tampilan ini untuk membina agen di bawah Anda agar perolehan jaringan tumbuh sehat.",
      },
    ],
    related: ["mengelola-leads", "melihat-komisi"],
  },
  {
    slug: "melihat-komisi",
    title: "Melihat & Mencairkan Komisi",
    category: "Komisi",
    order: 1,
    summary:
      "Memahami perhitungan komisi Anda dan langkah mengajukan pencairan saat komisi sudah dapat ditarik.",
    keywords: ["komisi", "pencairan", "penarikan", "withdraw", "saldo", "bonus"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Menu Komisi Saya menampilkan komisi dari setiap jamaah yang berangkat melalui Anda, beserta status pencairannya." },
      { type: "h2", text: "Mengajukan pencairan" },
      {
        type: "ol",
        items: [
          "Buka Komisi Saya dan pastikan ada saldo yang dapat dicairkan.",
          "Ajukan pencairan dan pilih rekening tujuan.",
          "Tunggu proses verifikasi dari travel.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Komisi umumnya dapat dicairkan setelah jamaah memenuhi syarat tertentu (mis. pelunasan atau keberangkatan). Periksa ketentuan dari travel Anda.",
      },
    ],
    related: ["mengelola-leads", "memantau-jaringan"],
  },
  {
    slug: "profil-agen",
    title: "Melihat Profil Agen",
    category: "Akun",
    order: 1,
    summary:
      "Lihat data akun Anda — kontak, upline, dan rekening bank yang dipakai untuk penerimaan komisi.",
    keywords: ["profil", "akun", "rekening", "bank", "data diri", "upline"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Halaman Profil menampilkan data akun Anda sebagaimana tercatat di sistem: nama, telepon, email, alamat, upline, serta rekening bank." },
      {
        type: "callout",
        variant: "warning",
        text: "Data profil bersifat hanya-baca. Bila ada yang perlu diubah, hubungi kantor untuk pembaruan.",
      },
      {
        type: "callout",
        variant: "info",
        text: "Pastikan data rekening bank sudah benar karena dipakai untuk pencairan komisi Anda.",
      },
    ],
    related: ["melihat-komisi", "memantau-jaringan"],
  },
];
