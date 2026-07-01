# Auth, Package, Contract, Mutawwif, Agent (B2B), AI-OCR — Code Review

**Tanggal:** 2026-07-02
**Cakupan:** Review ringkas (1 agent per modul, bukan swarm) menyeluruh (bukan diff) untuk 6 modul non-finansial. Lebih ringkas dari review modul finansial (refund, payroll, tabungan, vendor, accounting — lihat dokumen terpisah), tapi tetap menemukan beberapa isu keamanan serius.
**Status:** Dokumentasi temuan untuk sesi perbaikan berikutnya. Belum ada fix diterapkan.

---

## Ringkasan lintas-modul

| Modul | Critical | High | Medium | Tema utama |
|---|---|---|---|---|
| Auth | 2 | 4 | 2 | **Account takeover via Google SSO**, logout & ganti password tidak benar-benar revoke sesi |
| AI-OCR | 2 | 2 | 3 | **Excel formula injection**, kuota scan tidak pernah ditegakkan (billing exposure) |
| Agent (B2B) | 2 | 2 | 2 | Komisi bisa dobel, komisi tidak pernah di-reverse saat sale dibatalkan |
| Package | 1 | 3 | 2 | Kamar bisa oversold lewat edit harga, pembuatan paket dari mobile 100% gagal |
| Contract | 1 | 2 | 3 | Kontrak bisa ditandatangani dengan placeholder `{{...}}` belum terisi, IP audit selalu salah |
| Mutawwif | 1 | 2 | 2 | Pembimbing bisa double-booking, hapus pembimbing diam-diam hapus semua penugasan aktif |

---

## 🔐 Auth Module

### AUTH-1 (Critical). Google SSO auto-link tanpa verifikasi password — celah account takeover
**File:** `internal/auth/service/service.go:233-246`
Login Google untuk email yang SUDAH punya akun password langsung di-link & login otomatis cuma berdasarkan `email_verified=true` dari Google — tidak pernah minta password akun lama. Kalau penyerang berhasil kuasai email yang sama secara Google-verified (email lama yang di-reclaim, alias tim yang dipakai ulang, dll), dia bisa ambil alih akun Suluk siapa pun tanpa pernah tahu passwordnya.

### AUTH-2 (Critical). Logout & ganti password TIDAK benar-benar mencabut sesi
**File:** `frontend-svelte/src/lib/services/apiDomains/authSubscriptionApi.js:44-51` (logout tidak kirim header yang dibutuhkan backend), `internal/auth/service/service.go:737-753` (`ChangePassword` tidak panggil `DeleteRefreshTokensByUser`, padahal fungsi yang sama persis dipakai di `DeleteAccount` satu fungsi setelahnya)
Klik "Logout" cuma hapus localStorage di browser — refresh token di database tetap hidup sampai 7 hari. Ganti password (yang biasanya dilakukan justru karena curiga akun dibobol) juga tidak mencabut sesi/token yang sudah terbit — penyerang yang sudah punya refresh token tetap bisa terus dapat access token baru.

### AUTH-3 (High). Halaman Tim (`/app/team`) selalu crash — fitur manajemen tim tidak bisa dipakai
**File:** `internal/auth/handler/handler.go:167-174`, `frontend-svelte/src/lib/pages/TeamPage.svelte:304`
`GetOrganization` balikin objek organisasi tanpa field `members`, tapi frontend langsung akses `team.members.length` — TypeError, halaman blank untuk semua org.

### AUTH-4 (High). Owner bisa hapus akun sendiri dan bikin organisasi yatim selamanya
**File:** `internal/auth/service/service.go:759-769`
`DeleteAccount` tidak ada guard "jangan biarkan owner terakhir menghapus diri sendiri" (padahal `RemoveTeamMember` sudah punya guard ini). Tidak ada fitur transfer kepemilikan — begitu owner terakhir hapus akun, organisasi itu tidak bisa lagi diurus siapa pun (billing, undang admin baru, dll).

### AUTH-5 (High). Race condition di seat-limit team member (plan cap bisa dilanggar)
**File:** `internal/auth/service/service.go:559-580,582-661`
Cek-lalu-tulis tanpa lock — dua admin klik "tambah anggota" bersamaan bisa membuat org melebihi kuota seat paketnya tanpa proteksi database.

### AUTH-6 (High). Update subscription blind-write tanpa CAS — race bisa membalikkan pembayaran yang baru masuk
**File:** `internal/auth/repository/subscription.go:42-47`
Webhook pembayaran yang mengaktifkan Pro bisa "ketimpa" oleh proses auto-expire yang baca data basi hampir bersamaan, membalikkan status kembali ke expired.

### AUTH-7 (Medium). Perpanjangan langganan sebelum expired membuang sisa hari yang belum terpakai
**File:** `internal/auth/service/subscription.go:120-161`
`ActivatePlan` selalu hitung `expires_at = sekarang + 1 bulan`, bukan `expiry lama + 1 bulan` — pelanggan yang perpanjang lebih awal rugi hari yang tersisa.

### AUTH-8 (Medium). `checkSeatLimit` dan billing race di atas — sudah tercakup di AUTH-5.

---

## 📄 AI-OCR Module

### AIOCR-1 (Critical). Kuota scan bulanan TIDAK PERNAH ditegakkan — cuma tampilan, org bisa scan tanpa batas
**File:** `internal/aiocr/service/process_sync.go:47`
Endpoint `/process-documents` (yang benar-benar dipakai UI) tidak pernah cek kuota sebelum panggil AI provider berbayar — kuota cuma dihitung SETELAH sukses, cuma untuk ditampilkan. Org paket Gratis (5 scan/bulan) bisa scan ribuan dokumen, tagihan AI provider membengkak tanpa batas.

### AIOCR-2 (Critical). Excel formula injection lewat teks hasil OCR
**File:** `internal/aiocr/service/export.go:104`
Teks hasil OCR (yang bisa direkayasa lewat gambar yang diupload) ditulis mentah-mentah ke sel `.xlsx` tanpa sanitasi karakter `=`/`+`/`-`/`@`. Dokumen yang direkayasa (misal nama berisi formula `=cmd|'/c calc.exe'!A1`) bisa mengeksekusi kode saat file Excel dibuka oleh staff atau diupload ulang ke portal Siskopatuh resmi.

### AIOCR-3 (High). Cache/dedup OCR sudah ada di skema database tapi tidak pernah dipakai
**File:** `internal/aiocr/service/process_sync.go:44-46`
Submit dokumen yang sama dua kali (double-click, retry) selalu memicu panggilan AI berbayar baru + potong kuota baru, padahal ada tabel `ai_cache` + kolom `file_hash` yang dirancang khusus untuk ini tapi tidak pernah dibaca/ditulis.

### AIOCR-4 (High). Endpoint topup scan internal tidak validasi jumlah terhadap nilai kanonik
**File:** `internal/aiocr/handler/handler.go:61-88`
Cuma cek `scans > 0` — satu-satunya proteksi dari over-credit adalah shared secret `INTERNAL_API_KEY`. Bug/misconfig/kompromi kunci itu bisa mengkredit jumlah scan berapa pun ke org mana pun.

### AIOCR-5 (Medium). Fitur export Siskopatuh Excel sepenuhnya dead code — tidak pernah di-mount ke route manapun
**File:** `internal/aiocr/handler/handler.go:266-295`, `cmd/ai-ocr-service/main.go`

### AIOCR-6 (Medium). `GetStatus` selalu lapor provider "gemini" walau yang aktif "opencode"
**File:** `internal/aiocr/handler/handler.go:258-264`

### AIOCR-7 (Medium). Gagal insert satu file dalam scan job di-skip diam-diam, job tetap "berhasil" dengan file hilang
**File:** `internal/aiocr/service/service.go:120-122`

---

## 🤝 Agent (B2B) Module

### AGENT-1 (Critical). Komisi bisa tercatat dobel — tidak ada idempotency/unique constraint
**File:** `internal/agent/service/service.go:143-184`, `migrations/agent/001_initial_schema.up.sql`
Retry/klik ganda pada "Catat Komisi" membuat baris komisi baru (plus seluruh cascade upline) tanpa cek apakah invoice yang sama sudah pernah dikomisi — agent & upline-nya dibayar dua kali untuk satu penjualan.

### AGENT-2 (Critical). Komisi TIDAK PERNAH di-reverse saat invoice sumbernya dibatalkan/direfund
**File:** seluruh `internal/agent/service/service.go` (tidak ada method reversal), `internal/accounting/service/posting.go:126,148`
Agent tetap menyimpan 100% komisi untuk penjualan yang sudah dibatalkan — tidak ada event `commission.reversed`, tidak ada kode yang membatalkan komisi.

### AGENT-3 (High). Membayar komisi (`PayCommission`) tidak pernah posting ke buku besar
**File:** `internal/agent/repository/repository.go:207-219`
Pola yang sama persis dengan bug pembayaran vendor — status berubah jadi "paid" di database agent, tapi Hutang Komisi di GL tidak pernah didebit, Kas tidak pernah dikredit.

### AGENT-4 (High). Tier upline yang non-aktif membuat komisi tier itu hilang total, bukan naik ke tier di atasnya
**File:** `internal/agent/service/cascade.go:75-87`
Kalau upline tier-2 nonaktif, komisi tier-2 hilang tanpa jejak — bukan otomatis naik ke tier-3 yang aktif.

### AGENT-5, AGENT-6 (Medium) — level hierarki basi setelah re-parent, pagination page=0 bikin 500. Backlog, bukan mendesak.

---

## 📦 Package Module

### PKG-1 (High, hampir Critical). Edit harga kamar bisa reset counter reservasi → kamar oversold
**File:** `frontend-svelte/src/lib/pages/PackagesPage.svelte:342-426`, `internal/package/repository/repository.go:409-418`
Mengosongkan sementara field harga suatu tipe kamar (misal saat edit) menghapus baris pricing-tier-nya (termasuk counter `reserved_seats`). Isi lagi harganya nanti = tier baru dengan counter mulai dari 0 — kamar yang sebenarnya sudah penuh jadi "bisa dibooking lagi" untuk jamaah baru, padahal yang lama masih ada.

### PKG-2 (High). Pembuatan paket dari mobile app 100% selalu gagal
**File:** `frontend-svelte/src/lib/mobile/packageForm.js:3-12`
Kirim `package_type` yang tidak cocok enum backend, dan tidak pernah kirim `return_date` yang wajib (NOT NULL). Setiap percobaan buat paket dari mobile pasti gagal.

### PKG-3 (High). Kuota bisa diset lebih kecil dari jumlah yang sudah direservasi, tanpa peringatan
**File:** `internal/package/handler/handler.go:323-361`
Beda dengan update paket (yang sudah ada guard-nya), update pricing-tier tidak cek `quota_seats` baru terhadap `reserved_seats` yang sudah ada — tipe kamar bisa jadi permanen tidak bisa dibooking lagi akibat salah ketik.

### PKG-4 (Medium). Harga early-bird tidak pernah kedaluwarsa otomatis
**File:** `internal/package/model/model.go:105-106`

### PKG-5 (Medium). Slug paket unik lintas-org (bukan per-org) — nama paket bocor antar tenant tak terkait
**File:** `internal/package/repository/repository.go:341-349`

---

## ✍️ Contract Module

### CONTRACT-1 (High, dekat Critical, isu legal). Kontrak bisa ditandatangani dengan placeholder `{{...}}` yang belum terisi
**File:** `internal/contract/service/service.go:133`
Tidak ada pengecekan bahwa semua `{{variabel}}` di template sudah tergantikan sebelum kontrak dikirim & bisa ditandatangani — jamaah bisa menandatangani dokumen legal yang secara harfiah masih berisi teks `{{ketentuan_refund}}`.

### CONTRACT-2 (High). Alamat IP yang tercatat untuk audit tanda tangan SELALU salah
**File:** `internal/contract/handler/handler.go:196`
`ProxyHeader`/`TrustedProxies` tidak pernah dikonfigurasi — IP yang tersimpan selalu IP internal gateway, bukan IP jamaah yang menandatangani. Fitur "verifikasi integritas via IP" yang diiklankan ke pengguna sebenarnya tidak berfungsi.

### CONTRACT-3 (Medium). Drawer "Generate Kontrak" terisi data contoh (nama, no paspor, harga palsu) yang bisa ikut terkirim kalau staff tidak sadar
**File:** `frontend-svelte/src/lib/pages/ContractsPage.svelte:63-75,127-137,331-335`

### CONTRACT-4 (Medium). Kontrak pendek tidak bisa ditandatangani sama sekali — bug deteksi scroll
**File:** `frontend-svelte/src/lib/pages/PublicContractSigningPage.svelte:60-66,300`
Tombol "Tanda Tangani" terkunci selamanya kalau isi kontrak muat tanpa perlu di-scroll (event scroll tidak pernah terpicu).

### CONTRACT-5 (Medium). `CreateInstance` tidak validasi `jamaah_id`/`package_id` milik org yang sama

---

## 🕌 Mutawwif Module

### MUT-1 (High). Tidak ada cek bentrok jadwal — pembimbing bisa ditugaskan ke 2 kloter yang berangkat bersamaan
**File:** `internal/mutawwif/service/service.go:110-132`
Tidak ada tanggal sama sekali di tabel penugasan — bentrok jadwal tidak terdeteksi secara struktural, baru ketahuan di lapangan.

### MUT-2 (High). Hapus data pembimbing diam-diam menghapus SEMUA penugasannya (termasuk kloter yang sudah berangkat)
**File:** cascade delete di migration, `PembimbingPage.svelte:73-77`
Dialog konfirmasi tidak menyebutkan penugasan aktif yang akan ikut terhapus — kloter yang sedang jalan bisa kehilangan pembimbingnya di sistem tanpa notifikasi ke siapa pun.

### MUT-3 (Medium). `group_id` saat assign tidak divalidasi kepemilikan org — celah IDOR ringan

### MUT-4 (Medium). Role penugasan (leader/co_leader/kesehatan) tidak divalidasi — string bebas apa pun diterima

---

## Yang sudah dikonfirmasi BERSIH (dicek, bukan temuan)
- Tidak ada mismatch HTTP method/route di modul manapun dari 6 ini (auth, package, contract, mutawwif, agent, ai-ocr) — semua apiDomains sudah ter-wire dengan benar ke `ApiService`.
- AI-OCR: tidak ada kebocoran lintas-org di query manapun; webhook topup scan sudah idempotent (unique constraint + atomic claim).
- Package: `ReserveSeat`/`ReleaseSeat` sendiri race-safe (conditional UPDATE dalam transaksi) — oversell yang ditemukan (PKG-1) lewat jalur lain (hapus-buat-ulang tier), bukan race di reservasi itu sendiri.

## Rekomendasi urutan perbaikan (lintas 6 modul ini)
1. **AUTH-1** (account takeover) dan **AIOCR-2** (formula injection) — paling mendesak dari sisi keamanan, perbaiki duluan sebelum yang lain.
2. **AIOCR-1** (kuota tidak ditegakkan) — eksposur biaya AI provider tanpa batas, perbaiki secepatnya.
3. **AUTH-2, AUTH-3, AUTH-4** — logout/ganti-password tidak revoke sesi, halaman Tim selalu crash, owner bisa yatimkan org sendiri.
4. **PKG-1, PKG-2** — oversold kamar & pembuatan paket mobile yang 100% gagal, keduanya mengganggu operasional harian.
5. **AGENT-1, AGENT-2** — komisi dobel & tidak pernah di-reverse, sama seriusnya dengan temuan vendor/payroll.
6. **CONTRACT-1, CONTRACT-2** — isu legal/compliance (placeholder belum terisi, IP audit palsu).
7. Sisanya — backlog, bukan hotfix mendesak, tapi tetap dicatat supaya tidak hilang.
