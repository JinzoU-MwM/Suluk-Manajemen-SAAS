// Konten Pusat Bantuan untuk area STAF (/app).
// Hanya dibaca oleh helper di index.js saat area === "app". Jangan impor file
// area lain dari sini — isolasi peran dijaga di level konten & rute.
//
// Tipe blok body: lihat HelpBlock di ./index.js
//   { type: "p",  text }                      paragraf
//   { type: "h2", text }                      sub-judul
//   { type: "ul" | "ol", items: string[] }    daftar (tak berurut / berurut)
//   { type: "callout", variant, text }        kotak info | tip | warning
//
// TODO(domain): label tombol/menu persis pada tiap modul belum diverifikasi ke
// UI final; teks di bawah memakai istilah umum yang sudah dipakai produk
// (Jamaah, Paket, Manifest, Rooming, Kasir, Invoice). Sesuaikan bila berubah.

/** @type {import('./index.js').HelpGuide[]} */
export const APP_GUIDES = [
  {
    slug: "mengenal-dashboard",
    title: "Mengenal Dashboard Suluk",
    category: "Memulai",
    order: 1,
    summary:
      "Pengenalan tata letak aplikasi: sidebar modul, topbar, dan tempat menemukan menu utama untuk operasional travel.",
    keywords: ["dashboard", "beranda", "menu", "sidebar", "navigasi", "mulai"],
    updatedAt: "2026-06-19",
    body: [
      {
        type: "p",
        text: "Setelah masuk, Anda akan melihat Dashboard. Sisi kiri berisi sidebar dengan seluruh modul, bagian atas berisi topbar untuk profil dan notifikasi, dan area tengah menampilkan ringkasan operasional travel Anda.",
      },
      { type: "h2", text: "Modul utama di sidebar" },
      {
        type: "ul",
        items: [
          "Jamaah — data calon jamaah beserta dokumennya.",
          "Paket — daftar paket umrah/haji beserta harga dan kuota.",
          "Grup & Manifest — pengelompokan keberangkatan dan manifest digital.",
          "Rooming — pembagian kamar hotel jamaah.",
          "Invoice & Kasir — penagihan, pembayaran, dan cicilan.",
          "Keuangan — laporan pemasukan, pengeluaran, dan laba per trip.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Gunakan kotak pencarian di Pusat Bantuan ini kapan saja untuk menemukan panduan modul tertentu dengan cepat.",
      },
    ],
    related: ["mengelola-data-jamaah", "membuat-paket-umrah"],
  },
  {
    slug: "mengelola-data-jamaah",
    title: "Menambah & Mengelola Data Jamaah",
    category: "Jamaah & Paket",
    order: 1,
    summary:
      "Cara menambahkan jamaah baru, mempercepat input lewat scan KTP/paspor, dan memperbarui data agar siap untuk Siskopatuh.",
    keywords: ["jamaah", "jemaah", "ktp", "paspor", "ocr", "scan", "siskopatuh", "data peserta"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Jamaah adalah pusat data seluruh calon jamaah Anda. Dari sini Anda menambah, mencari, dan memperbarui data peserta." },
      { type: "h2", text: "Menambahkan jamaah baru" },
      {
        type: "ol",
        items: [
          "Buka modul Jamaah dari sidebar, lalu klik Tambah Jamaah.",
          "Isi data diri sesuai dokumen resmi (nama sesuai paspor, NIK, tanggal lahir).",
          "Unggah berkas pendukung seperti KTP, KK, dan paspor.",
          "Simpan. Jamaah akan muncul di daftar dan dapat dimasukkan ke paket atau grup.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Manfaatkan fitur scan KTP/paspor untuk mengisi data otomatis sehingga mengurangi salah ketik saat menyiapkan ekspor Siskopatuh.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Pastikan nama jamaah ditulis persis seperti di paspor. Perbedaan ejaan dapat menghambat proses visa.",
      },
    ],
    related: ["membuat-paket-umrah", "menyusun-manifest-rooming"],
  },
  {
    slug: "membuat-paket-umrah",
    title: "Membuat Paket Umrah",
    category: "Jamaah & Paket",
    order: 2,
    summary:
      "Menyusun paket umrah lengkap dengan komponen biaya, harga jual, dan kuota kursi agar penjualan rapi dan laba terpantau.",
    keywords: ["paket", "harga", "kuota", "biaya", "profit", "margin", "trip", "keberangkatan"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Paket menentukan apa yang Anda jual: tanggal keberangkatan, maskapai, hotel, durasi, harga, dan jumlah kursi." },
      { type: "h2", text: "Langkah membuat paket" },
      {
        type: "ol",
        items: [
          "Buka modul Paket, klik Tambah Paket.",
          "Isi nama paket, tanggal keberangkatan, durasi, hotel, dan maskapai.",
          "Masukkan komponen biaya per jamaah lalu tentukan harga jual.",
          "Tetapkan kuota kursi, lalu simpan.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Setelah paket aktif, Anda dapat menautkan jamaah ke paket tersebut dan memantau laba-rugi per trip dari modul Keuangan.",
      },
    ],
    related: ["mengelola-data-jamaah", "invoice-dan-pembayaran"],
  },
  {
    slug: "menyusun-manifest-rooming",
    title: "Menyusun Manifest & Rooming",
    category: "Keberangkatan",
    order: 1,
    summary:
      "Membentuk grup keberangkatan, membagi kamar hotel (Quad/Triple/Double), dan menerbitkan manifest digital untuk mutawwif.",
    keywords: ["manifest", "rooming", "kamar", "grup", "keberangkatan", "mutawwif", "hotel"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Menjelang keberangkatan, jamaah dikelompokkan ke dalam grup, dibagi kamarnya, dan dirangkum dalam manifest." },
      { type: "h2", text: "Alur ringkas" },
      {
        type: "ol",
        items: [
          "Buat grup keberangkatan dan masukkan jamaah dari paket terkait.",
          "Buka Rooming untuk membagi kamar — pisahkan gender dan satukan keluarga.",
          "Periksa kelengkapan dokumen tiap jamaah di grup tersebut.",
          "Terbitkan manifest digital untuk mutawwif/petugas lapangan.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Manifest digital memuat data jamaah, rooming, dan itinerary dalam satu tautan sehingga petugas cukup membukanya dari ponsel.",
      },
    ],
    related: ["mengelola-data-jamaah", "membuat-paket-umrah"],
  },
  {
    slug: "invoice-dan-pembayaran",
    title: "Invoice, Pembayaran & Cicilan",
    category: "Keuangan",
    order: 1,
    summary:
      "Menerbitkan invoice untuk jamaah, mencatat pembayaran dan cicilan di Kasir, serta memantau sisa tagihan.",
    keywords: ["invoice", "tagihan", "pembayaran", "cicilan", "kasir", "dp", "pelunasan", "keuangan"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Setiap jamaah yang membeli paket dapat ditagih melalui invoice, lalu pembayarannya dicatat di Kasir hingga lunas." },
      { type: "h2", text: "Menerbitkan & menagih" },
      {
        type: "ol",
        items: [
          "Dari modul Invoice, buat invoice baru dan pilih jamaah serta paketnya.",
          "Tentukan total tagihan dan, bila perlu, skema cicilan/DP.",
          "Bagikan invoice ke jamaah.",
          "Saat jamaah membayar, catat pembayaran di Kasir; sisa tagihan akan berkurang otomatis.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Status pembayaran tiap jamaah terlihat di daftar invoice sehingga mudah memantau siapa yang belum melunasi.",
      },
    ],
    related: ["membuat-paket-umrah", "mengenal-dashboard"],
  },
  {
    slug: "scan-dokumen-ai",
    title: "Memindai Dokumen dengan AI Scanner",
    category: "Jamaah & Paket",
    order: 3,
    summary:
      "Ekstrak data jamaah otomatis dari KTP, Kartu Keluarga, atau paspor memakai AI OCR, lalu simpan langsung ke grup.",
    keywords: ["scanner", "ocr", "ai", "ktp", "paspor", "kartu keluarga", "scan", "ekstrak", "import"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "AI Scanner membaca foto dokumen identitas dan mengisi data jamaah secara otomatis, sehingga input lebih cepat dan minim salah ketik." },
      { type: "h2", text: "Langkah memindai" },
      {
        type: "ol",
        items: [
          "Pilih tipe dokumen: KTP, Kartu Keluarga, atau Paspor.",
          "Pilih grup tujuan bila ingin langsung menyimpan hasilnya (opsional).",
          "Unggah file dokumen (JPG, PNG, atau PDF, maksimal 10 MB).",
          "Tunggu proses OCR selesai — progres ditampilkan secara langsung.",
          "Tinjau hasil di tabel, perbaiki data yang ditandai, lalu Export Excel atau Simpan ke Grup.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Periksa peringatan validasi sebelum menyimpan — data yang janggal ditandai agar mudah dikoreksi.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Sebagian kuota OCR dan mode lanjutan hanya tersedia di paket Pro; pada paket gratis berlaku batas penggunaan.",
      },
    ],
    related: ["mengelola-data-jamaah", "membuat-grup-keberangkatan"],
  },
  {
    slug: "crm-pipeline-jamaah",
    title: "Mengelola Pipeline Jamaah (CRM)",
    category: "Penjualan & Agen",
    order: 1,
    summary:
      "Pantau calon jamaah dari Prospek hingga Berangkat dalam papan Kanban atau tabel, lengkap dengan progres pembayaran.",
    keywords: ["crm", "prospek", "lead", "pipeline", "kanban", "funnel", "penjualan", "calon jamaah"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "CRM menyusun calon jamaah dalam tahapan penjualan — dari Prospek, Booking, DP, Cicilan, Lunas, hingga Berangkat atau Batal — agar tidak ada prospek yang terlewat." },
      { type: "h2", text: "Memindahkan tahap" },
      {
        type: "ol",
        items: [
          "Klik Tambah Jamaah untuk memasukkan prospek baru (minimal nama).",
          "Gunakan pencarian (nama, HP, NIK, paspor) atau saring berdasarkan tahap.",
          "Di tampilan Kanban, seret kartu jamaah ke tahap berikutnya — perubahan tersimpan otomatis.",
          "Saat memindahkan ke Batal, pilih alasan kehilangan agar bisa dianalisis kemudian.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Saring lead berdasarkan suhu (Hot/Warm) untuk fokus follow-up, dan buka tampilan Funnel untuk melihat di tahap mana prospek menumpuk.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Jamaah harus terdaftar di sebuah paket sebelum dapat dipindahkan ke tahap berikutnya.",
      },
    ],
    related: ["mengelola-data-jamaah", "invoice-dan-pembayaran", "kelola-agen-mitra"],
  },
  {
    slug: "kelola-agen-mitra",
    title: "Mengelola Agen & Mitra",
    category: "Penjualan & Agen",
    order: 2,
    summary:
      "Kelola agen penjualan, atur komisi berjenjang (tier), dan catat pembayaran komisi.",
    keywords: ["agen", "mitra", "komisi", "tier", "upline", "downline", "jaringan", "referral"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Agen & Mitra mencatat agen penjualan beserta komisinya, termasuk struktur berjenjang (upline–downline)." },
      { type: "h2", text: "Menambah agen & komisi" },
      {
        type: "ol",
        items: [
          "Di tab Agen, klik Tambah Agen dan isi data (nama, kontak, rekening, tipe, rate komisi, upline).",
          "Atur rate komisi berjenjang lewat pengaturan Tier bila menggunakan struktur upline.",
          "Di tab Komisi, pantau komisi per jamaah dan tandai 'Bayar' saat komisi dilunasi.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Buatkan akun Portal untuk tiap agen agar mereka dapat memantau komisinya sendiri tanpa perlu bertanya.",
      },
    ],
    related: ["crm-pipeline-jamaah", "invoice-dan-pembayaran"],
  },
  {
    slug: "e-kontrak-jamaah",
    title: "Membuat E-Kontrak Jamaah",
    category: "Penjualan & Agen",
    order: 3,
    summary:
      "Susun template akad, terbitkan kontrak per jamaah dengan variabel otomatis, dan pantau status tanda tangan digital.",
    keywords: ["kontrak", "e-kontrak", "akad", "tanda tangan", "template", "ttd", "perjanjian"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Kontrak membantu Anda membuat akad perjalanan secara digital: sekali menyusun template, kontrak tiap jamaah terisi otomatis dari variabel." },
      { type: "h2", text: "Template & penerbitan" },
      {
        type: "ol",
        items: [
          "Buat template di tab Template Kontrak dan sisipkan variabel seperti nama jamaah, nama paket, dan harga.",
          "Klik Generate Kontrak, pilih template, lengkapi data jamaah, dan tentukan masa berlaku tautan.",
          "Bagikan tautan tanda tangan ke jamaah lewat WhatsApp atau email.",
          "Pantau status tiap kontrak: Terkirim, Ditandatangani, atau Expired.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Selalu Preview template sebelum menerbitkan untuk memastikan semua variabel terisi dengan benar.",
      },
    ],
    related: ["membuat-paket-umrah", "mengelola-data-jamaah"],
  },
  {
    slug: "membuat-grup-keberangkatan",
    title: "Membuat Grup Keberangkatan",
    category: "Keberangkatan",
    order: 2,
    summary:
      "Kelompokkan jamaah per kloter, tetapkan paket dan tanggal, tugaskan pembimbing, dan kelola status keberangkatan.",
    keywords: ["grup", "kloter", "keberangkatan", "rombongan", "pembimbing", "status"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Grup mengelompokkan jamaah ke dalam satu keberangkatan (kloter) beserta paket, tanggal, pembimbing, dan rooming-nya." },
      { type: "h2", text: "Langkah membuat grup" },
      {
        type: "ol",
        items: [
          "Klik Buat Grup Baru dan beri nama (mis. 'Umrah Maret 2026').",
          "Buka Keberangkatan pada kartu grup, lalu pilih paket dan tanggal berangkat.",
          "Tugaskan pembimbing beserta perannya (Ketua/Wakil/Kesehatan).",
          "Geser status mengikuti alur: Draf → Siap → Berangkat → Selesai.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Status Selesai dan Batal bersifat final dan tidak dapat diubah kembali.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Setelah status melewati Draf, paket dan tanggal keberangkatan tidak dapat diubah lagi.",
      },
    ],
    related: ["menyusun-manifest-rooming", "kelola-pembimbing", "kelola-visa-dokumen"],
  },
  {
    slug: "kelola-visa-dokumen",
    title: "Mengelola Visa & Dokumen",
    category: "Keberangkatan",
    order: 3,
    summary:
      "Lacak pengajuan visa jamaah dari draf hingga disetujui pada papan status, lengkap dengan nomor dan masa berlaku.",
    keywords: ["visa", "dokumen", "pengajuan", "status", "paspor", "kedaluwarsa"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Visa menampilkan pengajuan visa jamaah dalam papan berkolom sesuai statusnya, sehingga progres mudah dipantau." },
      { type: "h2", text: "Alur status" },
      {
        type: "ol",
        items: [
          "Cari dan pilih jamaah, lalu buat draf visa (provider, nomor referensi, catatan).",
          "Ajukan visa dari Draf ke Diajukan.",
          "Setujui dengan mengisi nomor visa dan tanggal berlaku, atau tolak dengan menyebut alasan.",
          "Bila ditolak atau kedaluwarsa, ajukan ulang dari kartu tersebut.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Papan terbagi lima kolom: Draf, Diajukan, Disetujui, Ditolak, dan Kedaluwarsa.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Untuk menyetujui, nomor visa dan tanggal berlaku keduanya wajib diisi.",
      },
    ],
    related: ["mengelola-data-jamaah", "membuat-grup-keberangkatan"],
  },
  {
    slug: "menyusun-itinerary",
    title: "Menyusun Itinerary Perjalanan",
    category: "Keberangkatan",
    order: 4,
    summary:
      "Buat jadwal perjalanan per grup — penerbangan, hotel, transportasi, dan aktivitas ibadah — dalam tampilan timeline.",
    keywords: ["itinerary", "jadwal", "rencana perjalanan", "timeline", "penerbangan", "hotel", "aktivitas"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Itinerary merangkum agenda perjalanan satu grup secara kronologis, dari penerbangan dan hotel hingga aktivitas ibadah." },
      { type: "h2", text: "Menambah agenda" },
      {
        type: "ol",
        items: [
          "Pilih grup keberangkatan dari dropdown.",
          "Klik Tambah, lalu pilih kategori (Penerbangan/Hotel/Transportasi/Aktivitas/Lainnya).",
          "Isi tanggal dan nama kegiatan (wajib); lengkapi jam, lokasi, dan catatan bila perlu.",
          "Simpan — agenda muncul pada timeline sesuai urutan waktunya.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Isi jam mulai dan selesai untuk kegiatan berdurasi pasti agar jamaah dan pembimbing mudah menyelaraskan jadwal.",
      },
    ],
    related: ["membuat-grup-keberangkatan", "kelola-pembimbing"],
  },
  {
    slug: "kelola-pembimbing",
    title: "Mengelola Pembimbing & Mutawwif",
    category: "Keberangkatan",
    order: 5,
    summary:
      "Kelola data pembimbing (mutawwif, tour leader, tim kesehatan) beserta lisensinya, dan pantau penugasan kloter.",
    keywords: ["pembimbing", "mutawwif", "tour leader", "tim kesehatan", "lisensi", "skp"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Pembimbing menyimpan daftar pembimbing ibadah beserta lisensinya dan berapa kloter yang mereka tangani." },
      { type: "h2", text: "Menambah pembimbing" },
      {
        type: "ol",
        items: [
          "Klik Tambah Pembimbing dan isi nama (wajib), telepon, dan tipe.",
          "Lengkapi email, nomor lisensi/SKP, dan tanggal masa berlaku lisensi.",
          "Simpan; pembimbing siap ditugaskan ke kloter.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Selalu isi masa berlaku lisensi — lisensi yang kedaluwarsa ditandai merah agar mudah diperbarui.",
      },
      {
        type: "callout",
        variant: "info",
        text: "Penugasan pembimbing ke kloter dilakukan dari modul Grup, bukan dari halaman ini.",
      },
    ],
    related: ["membuat-grup-keberangkatan", "menyusun-manifest-rooming"],
  },
  {
    slug: "laporan-keuangan",
    title: "Membaca Laporan Keuangan",
    category: "Keuangan",
    order: 2,
    summary:
      "Pantau pendapatan, biaya, laba kotor, dan piutang; analisis laba-rugi per paket dan tren pendapatan bulanan.",
    keywords: ["keuangan", "laporan", "laba rugi", "pendapatan", "biaya", "piutang", "margin", "profit"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Keuangan merangkum kesehatan bisnis Anda: total pendapatan, biaya operasional, laba kotor beserta marginnya, dan total piutang." },
      { type: "h2", text: "Yang bisa Anda lihat" },
      {
        type: "ul",
        items: [
          "Kartu ringkasan: Pendapatan, Biaya Operasional, Laba Kotor, dan Piutang.",
          "P&L per Paket — untung/rugi tiap paket perjalanan.",
          "Piutang per Paket — paket dengan sisa tagihan beserta progres pembayaran.",
          "Tren Pendapatan — grafik pemasukan bulanan.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Bandingkan kolom Proyeksi vs Aktual pada rincian biaya untuk menemukan kategori yang melenceng dari rencana.",
      },
      {
        type: "callout",
        variant: "info",
        text: "Angka bergantung pada kelengkapan data — pastikan seluruh biaya dan pembayaran sudah tercatat.",
      },
    ],
    related: ["invoice-dan-pembayaran", "kelola-vendor", "akuntansi-pembukuan"],
  },
  {
    slug: "pembatalan-refund",
    title: "Memproses Pembatalan & Refund",
    category: "Keuangan",
    order: 3,
    summary:
      "Kelola pengajuan pembatalan jamaah dan proses pengembalian dana sesuai kebijakan refund, dari pending hingga selesai.",
    keywords: ["pembatalan", "refund", "batal", "pengembalian dana", "kebijakan", "cancel"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Pembatalan memproses permintaan refund jamaah mengikuti kebijakan yang Anda tetapkan, sambil melacak setiap tahapnya." },
      { type: "h2", text: "Alur refund" },
      {
        type: "ol",
        items: [
          "Saring daftar berdasarkan status (Pending, Disetujui, Diproses, Selesai, Ditolak).",
          "Buka detail sebuah refund untuk melihat nominal, persentase, dan alasannya.",
          "Setujui atau tolak pengajuan yang masih Pending.",
          "Lanjutkan: Proses Refund, lalu Tandai Selesai saat dana sudah dikembalikan.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Kebijakan refund menetapkan persentase pengembalian berdasarkan berapa hari sebelum keberangkatan jamaah membatalkan.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Alur status berjalan satu arah (Pending → Disetujui → Diproses → Selesai) — periksa baik-baik sebelum melanjutkan tiap tahap.",
      },
    ],
    related: ["invoice-dan-pembayaran", "mengelola-data-jamaah"],
  },
  {
    slug: "tabungan-umrah-jamaah",
    title: "Mengelola Tabungan Umrah Jamaah",
    category: "Keuangan",
    order: 4,
    summary:
      "Buka rekening tabungan bertahap untuk jamaah menuju target paket, dan catat tiap setoran.",
    keywords: ["tabungan", "setoran", "nabung", "target", "cicilan tabungan", "umrah"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Tabungan membantu jamaah menyisihkan dana bertahap menuju target paket, dengan setiap setoran tercatat rapi." },
      { type: "h2", text: "Buka & setor" },
      {
        type: "ol",
        items: [
          "Klik Buka Tabungan, pilih jamaah, dan tentukan target (opsional).",
          "Untuk mencatat setoran, klik Setor pada tabungan yang dimaksud.",
          "Isi jumlah, metode pembayaran (Tunai/Transfer/QRIS), dan nomor referensi.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Isi target dan nomor referensi tiap setoran agar progres terpantau dan mudah diaudit.",
      },
      {
        type: "callout",
        variant: "info",
        text: "Setiap setoran otomatis tercatat di modul Akuntansi sebagai jurnal.",
      },
    ],
    related: ["mengelola-data-jamaah", "invoice-dan-pembayaran"],
  },
  {
    slug: "akuntansi-pembukuan",
    title: "Memahami Akuntansi & Pembukuan",
    category: "Keuangan",
    order: 5,
    summary:
      "Lihat neraca, laba-rugi, jurnal, dan bagan akun yang terbentuk otomatis dari transaksi operasional.",
    keywords: ["akuntansi", "pembukuan", "jurnal", "neraca", "laba rugi", "coa", "bagan akun"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Akuntansi menyusun pembukuan secara otomatis dari transaksi sehari-hari — invoice, pembayaran, vendor, payroll, komisi, dan tabungan — tanpa input jurnal manual." },
      { type: "h2", text: "Tab yang tersedia" },
      {
        type: "ul",
        items: [
          "Insight — ringkasan kas/bank, piutang, pendapatan, dan laba/rugi.",
          "Neraca — aset, liabilitas, dan ekuitas beserta status keseimbangannya.",
          "Laba Rugi — pendapatan dan beban suatu periode.",
          "Jurnal & Bagan Akun — daftar transaksi dan struktur akun.",
        ],
      },
      {
        type: "callout",
        variant: "warning",
        text: "Karena jurnal terbentuk otomatis, bila Neraca tidak seimbang berarti ada transaksi yang perlu diperiksa.",
      },
    ],
    related: ["laporan-keuangan", "invoice-dan-pembayaran"],
  },
  {
    slug: "payroll-penggajian",
    title: "Mengelola Payroll (Penggajian)",
    category: "Keuangan",
    order: 6,
    summary:
      "Kelola data karyawan, absensi, cuti, slip gaji, dan kasbon tim Anda.",
    keywords: ["payroll", "gaji", "penggajian", "slip gaji", "absensi", "cuti", "kasbon", "karyawan"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Payroll merangkum pengelolaan karyawan: dari absensi dan cuti hingga penerbitan slip gaji dan pencatatan kasbon." },
      { type: "h2", text: "Alur penggajian" },
      {
        type: "ol",
        items: [
          "Tambahkan karyawan beserta gaji pokok, tunjangan, dan potongan (BPJS, PPh21).",
          "Catat absensi per tanggal dan kelola pengajuan cuti.",
          "Buat slip gaji per periode, lalu finalisasi.",
          "Catat kasbon karyawan dan pembayarannya bila ada.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Slip gaji dihitung otomatis: gaji pokok + tunjangan dikurangi potongan menjadi gaji bersih.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Slip gaji yang sudah difinalisasi tidak dapat dibatalkan — pastikan datanya benar sebelum finalisasi.",
      },
    ],
    related: ["akuntansi-pembukuan"],
  },
  {
    slug: "kelola-vendor",
    title: "Mengelola Vendor & Tagihan",
    category: "Operasional & Aset",
    order: 1,
    summary:
      "Catat data vendor (hotel, maskapai, transport, katering), kelola tagihan operasional per trip, dan rekam pembayarannya.",
    keywords: ["vendor", "pemasok", "supplier", "tagihan", "hutang", "pembayaran", "hotel", "maskapai"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Vendor menyimpan data pemasok dan melacak tagihan operasional Anda kepada mereka hingga lunas." },
      { type: "h2", text: "Menambah vendor & tagihan" },
      {
        type: "ol",
        items: [
          "Klik Tambah Vendor dan isi nama, tipe, PIC, serta rekening bank.",
          "Buka detail vendor, lalu buat tagihan (deskripsi, nominal, jatuh tempo, trip terkait).",
          "Rekam pembayaran secara bertahap hingga sisa tagihan menjadi nol.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Pantau Total Outstanding dan tagihan Overdue di ringkasan untuk memprioritaskan pembayaran.",
      },
      {
        type: "callout",
        variant: "info",
        text: "Tagihan mendukung mata uang IDR maupun USD; isi kurs saat tagihan dibuat agar nilainya akurat.",
      },
    ],
    related: ["laporan-keuangan", "akuntansi-pembukuan"],
  },
  {
    slug: "inventaris-perlengkapan",
    title: "Mengelola Inventaris Perlengkapan",
    category: "Operasional & Aset",
    order: 2,
    summary:
      "Hitung kebutuhan perlengkapan per grup (koper, ihram, mukena, ukuran seragam) dan catat serah-terima ke jamaah lewat QR.",
    keywords: ["inventaris", "perlengkapan", "koper", "seragam", "ihram", "mukena", "stok", "qr", "serah terima"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Inventaris memperkirakan kebutuhan perlengkapan tiap grup dan membantu mencatat penyerahannya ke jamaah." },
      { type: "h2", text: "Langkah pengelolaan" },
      {
        type: "ol",
        items: [
          "Pilih grup keberangkatan untuk melihat perkiraan kebutuhan (jumlah jamaah, koper, ihram, mukena).",
          "Lengkapi ukuran baju dan Family ID tiap anggota.",
          "Pilih anggota lalu klik Tandai Terima saat perlengkapan diserahkan.",
          "Untuk serah-terima di lapangan, pindai QR jamaah pada checkpoint Perlengkapan atau Koper.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Gunakan tombol 'Semua' untuk menandai banyak anggota sekaligus saat distribusi massal.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Penandaan 'Terima' sulit dibatalkan dari layar — pastikan perlengkapan benar-benar sudah diserahkan.",
      },
    ],
    related: ["membuat-grup-keberangkatan", "menyusun-manifest-rooming"],
  },
  {
    slug: "kelola-dokumen-jamaah",
    title: "Melengkapi Dokumen Jamaah",
    category: "Jamaah & Paket",
    order: 4,
    summary:
      "Pantau kelengkapan berkas tiap jamaah dan masa berlaku paspor, lalu perbarui status tiap dokumen.",
    keywords: ["dokumen", "berkas", "paspor", "kelengkapan", "checklist", "ktp", "icv"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Dokumen menampilkan checklist kelengkapan berkas setiap jamaah beserta status paspornya, sehingga mudah memantau siapa yang belum lengkap." },
      { type: "h2", text: "Memperbarui dokumen" },
      {
        type: "ol",
        items: [
          "Cari jamaah (nama atau nomor paspor), atau saring berdasarkan status Lengkap/Belum Lengkap.",
          "Buka detail jamaah untuk melihat checklist jenis-jenis dokumennya.",
          "Unggah berkas yang belum ada, atau ubah status dokumen (Belum → Diterima → Diproses → Selesai).",
          "Pantau ringkasan di atas: total jamaah, dokumen belum lengkap, dan paspor yang akan habis.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Jamaah dianggap Lengkap hanya bila seluruh dokumen wajibnya berstatus Selesai.",
      },
      {
        type: "callout",
        variant: "warning",
        text: "Perubahan status langsung tersimpan tanpa tombol Simpan terpisah — pastikan pilihan Anda benar.",
      },
    ],
    related: ["mengelola-data-jamaah", "kelola-visa-dokumen", "scan-dokumen-ai"],
  },
  {
    slug: "analitik-statistik",
    title: "Membaca Analitik & Statistik",
    category: "Analitik & Laporan",
    order: 1,
    summary:
      "Lihat ringkasan operasional — jumlah jamaah, grup, tren bulanan, dan komposisi — untuk gambaran bisnis cepat.",
    keywords: ["analytics", "analitik", "statistik", "tren", "dashboard", "grafik", "laporan"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Halaman Analytics merangkum angka-angka operasional Anda dalam beberapa kartu dan grafik agar mudah dibaca sekilas." },
      { type: "h2", text: "Yang ditampilkan" },
      {
        type: "ul",
        items: [
          "Kartu ringkasan: total jamaah, total grup, dan jamaah bulan ini.",
          "Grafik tren beberapa bulan terakhir.",
          "Komposisi jamaah beserta data yang belum lengkap.",
        ],
      },
      {
        type: "callout",
        variant: "tip",
        text: "Klik Refresh untuk memuat angka terbaru bila data baru saja berubah.",
      },
    ],
    related: ["laporan-keuangan", "mengenal-dashboard"],
  },
  {
    slug: "ekspor-laporan",
    title: "Mengekspor Laporan",
    category: "Analitik & Laporan",
    order: 2,
    summary:
      "Unduh laporan keuangan dalam Excel serta kwitansi dan slip gaji dalam PDF untuk arsip atau analisis.",
    keywords: ["export", "ekspor", "unduh", "excel", "pdf", "laporan", "kwitansi", "slip gaji"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Export memudahkan Anda mengunduh berkas laporan untuk kebutuhan arsip, audit, atau analisis di luar aplikasi." },
      { type: "h2", text: "Yang bisa diunduh" },
      {
        type: "ul",
        items: [
          "Laporan Keuangan — P&L dan Biaya Operasional dalam format Excel.",
          "Kwitansi invoice yang sudah lunas dalam format PDF.",
          "Slip gaji per karyawan dalam format PDF.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Tiap kategori menampilkan item terbaru; unduh per item sesuai kebutuhan.",
      },
    ],
    related: ["laporan-keuangan", "invoice-dan-pembayaran", "payroll-penggajian"],
  },
  {
    slug: "tim-organisasi",
    title: "Mengelola Tim & Organisasi",
    category: "Pengaturan & Tim",
    order: 1,
    summary:
      "Undang anggota tim, atur peran (Owner/Admin/Viewer), dan kelola hak akses organisasi Anda.",
    keywords: ["tim", "organisasi", "anggota", "peran", "role", "akses", "undang", "owner", "admin", "viewer"],
    updatedAt: "2026-06-19",
    body: [
      { type: "p", text: "Modul Tim & Organisasi mengatur siapa saja yang dapat masuk ke akun travel Anda dan sebatas apa kewenangannya." },
      { type: "h2", text: "Mengundang anggota" },
      {
        type: "ol",
        items: [
          "Buat organisasi terlebih dahulu bila belum ada.",
          "Undang anggota dengan memasukkan email dan memilih peran (Admin atau Viewer).",
          "Bagikan kode undangan kepada anggota agar mereka dapat bergabung.",
          "Ubah peran atau keluarkan anggota dari daftar bila diperlukan.",
        ],
      },
      {
        type: "callout",
        variant: "info",
        text: "Viewer hanya dapat melihat data, Admin dapat mengelola, dan sebagian pengaturan hanya tersedia bagi Owner.",
      },
    ],
    related: ["mengenal-dashboard"],
  },
];
