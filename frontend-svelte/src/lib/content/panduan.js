// Index of the /panduan (guides) articles. Drives the listing page, the sitemap,
// and each article's own metadata. Add an entry + a matching route folder under
// src/routes/(marketing)/panduan/<slug>/ to publish a new guide.
export const PANDUAN = [
  {
    slug: "cara-input-data-jamaah-siskopatuh",
    title: "Cara Input Data Jamaah ke Siskopatuh Lebih Cepat",
    description:
      "Panduan mempercepat input data jamaah ke format Siskopatuh: dari scan KTP/KK otomatis hingga ekspor Excel 32 kolom tanpa salah ketik.",
    date: "2026-06-10",
    readMinutes: 6,
  },
  {
    slug: "checklist-keberangkatan-umrah",
    title: "Checklist Persiapan Keberangkatan Jamaah Umrah",
    description:
      "Daftar periksa lengkap persiapan keberangkatan jamaah umrah: dokumen, manifest, rooming, perlengkapan, hingga manasik.",
    date: "2026-06-10",
    readMinutes: 7,
  },
  {
    slug: "cara-rooming-hotel-jamaah-umrah",
    title: "Cara Mengatur Rooming Hotel Jamaah Umrah",
    description:
      "Cara membagi kamar hotel jamaah umrah dengan rapi: pisahkan gender, satukan keluarga, dan isi kamar Quad/Triple/Double secara otomatis.",
    date: "2026-06-10",
    readMinutes: 5,
  },
  {
    slug: "manifest-digital-mutawwif",
    title: "Manifest Digital Mutawwif: Apa Itu dan Cara Membuatnya",
    description:
      "Mengenal manifest digital untuk mutawwif: data jamaah, rooming, dan itinerary dalam satu tautan ber-PIN yang aman, langsung di ponsel.",
    date: "2026-06-10",
    readMinutes: 5,
  },
  {
    slug: "menghitung-biaya-profit-paket-umrah",
    title: "Cara Menghitung Biaya dan Profit Paket Umrah",
    description:
      "Panduan menyusun harga paket umrah: komponen biaya per jamaah, margin, dan cara memantau laba rugi per trip secara real-time.",
    date: "2026-06-10",
    readMinutes: 7,
  },
  {
    slug: "tips-memilih-software-travel-umrah",
    title: "7 Tips Memilih Software Travel Umrah yang Tepat",
    description:
      "Kriteria penting memilih software travel umrah: dukungan Siskopatuh, OCR dokumen, keuangan, e-kontrak, hingga aplikasi mobile untuk tim lapangan.",
    date: "2026-06-10",
    readMinutes: 6,
  },
];

export function getPanduan(slug) {
  return PANDUAN.find((p) => p.slug === slug);
}
