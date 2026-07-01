# Tabungan (Savings) Module — Full Code Review

**Tanggal:** 2026-07-02
**Cakupan:** Audit menyeluruh (bukan diff) — `internal/tabungan/{handler,service,repository,model}/*.go`, migrations, `cmd/tabungan-service/main.go`, jalur cross-service ke invoice-service (`Convert`/settle), frontend (`tabunganApi.js`, `TabunganPage.svelte`, `Tabungan.svelte` mobile).
**Status:** Dokumentasi temuan untuk sesi perbaikan berikutnya. Belum ada fix diterapkan.

**Catatan arsitektur penting:** tabungan-service dan invoice-service adalah proses + database terpisah, dijembatani cuma lewat satu HTTP call sinkron (`POST /invoices/internal/settle`). Tidak ada distributed transaction, tidak ada saga/kompensasi, dan yang paling parah — **tidak ada idempotency key** yang dibawa lintas panggilan itu. Sebagian besar temuan Critical berakar dari sini.

---

## Ringkasan

| Level | Jumlah | Tema |
|---|---|---|
| 🔴 Critical | 4 | Konversi tabungan bisa "hilang" separuh (invoice lunas tapi saldo tidak berkurang penuh), tabungan jamaah A bisa melunasi invoice jamaah B, setoran bisa dobel |
| 🟠 High | 3 | Bug Kas/Bank yang sama seperti kasus refund, query tanpa org-scope, akar penyebab race |
| — Frontend | 3 | Tombol setor (deposit) MATI TOTAL di web & mobile, fitur convert tidak punya UI sama sekali |

---

## 🔴 Critical

### T1. `Convert()` — konsistensi antar-service rusak di bawah race, kode sendiri sudah mengakui ini ("CONVERSION INCONSISTENCY") tapi cuma di-log
**File:** `internal/tabungan/service/service.go:90-131`
Baca saldo tabungan tanpa lock, panggil invoice-service untuk kurangi sisa tagihan invoice, BARU SETELAH itu coba kurangi saldo tabungan di transaksi terpisah. Kalau staff klik convert dua kali (atau ada retry), invoice-service (yang sudah benar pakai `FOR UPDATE`) memproses kedua request secara berurutan dan invoice bisa jadi lunas penuh — tapi di sisi tabungan, request kedua akan gagal (`ErrInsufficient`, saldo sudah 0 dari request pertama). Hasilnya: invoice tercatat lunas penuh, tapi tabungan cuma berkurang sebagian — selisihnya tidak pernah tercatat di GL (`Dr Hutang Tabungan/Cr Piutang` untuk porsi itu tidak pernah terpost) dan tidak ada mekanisme reconcile apa pun.

### T2. Retry setelah response hilang bisa menyisakan saldo tabungan "hantu" yang bisa dibelanjakan dua kali
**File:** `internal/tabungan/service/service.go:106-113,143-155`
Kalau response dari `settleInvoice` hilang (network blip) padahal server sudah commit, request di-retry dengan jumlah yang sama — tapi karena `remaining` invoice sudah berkurang dari percobaan pertama, jumlah yang benar-benar dipotong dari invoice di percobaan kedua lebih kecil. Sisa saldo tabungan yang "seharusnya" sudah terpakai tapi ternyata masih ada di akun bisa dikonversi LAGI ke invoice lain — efeknya double-spend dana yang sebenarnya sudah dianggap terpakai oleh pembukuan.

### T3. Tidak ada validasi kepemilikan jamaah — tabungan jamaah A bisa melunasi invoice jamaah B
**File:** `internal/tabungan/service/service.go:90-113` dibandingkan `internal/invoice/repository/settle.go:14-56`
`Convert()` hanya cek `org_id` di kedua sisi (akun tabungan & invoice) — tidak pernah membandingkan `jamaah_id` pemilik akun tabungan dengan `jamaah_id` pemilik invoice. Salah ketik ID invoice (atau human error/insider) bisa membuat tabungan satu jamaah dipakai melunasi invoice jamaah lain, tanpa error, tanpa peringatan, tanpa jejak audit — baru ketahuan kalau jamaah yang tabungannya "hilang" komplain.

### T4. Setoran (deposit) tidak idempotent — retry/klik ganda dobel-kredit saldo
**File:** `internal/tabungan/handler/handler.go:73-100`, `service.go:63-84`, `repository.go:97-127`
Tidak ada idempotency key, tidak ada unique constraint di `(account_id, reference)`. Response yang hilang lalu di-retry menghasilkan 2 baris setoran + saldo bertambah 2x dari 1 kali setoran tunai yang sebenarnya — pola bug yang sama persis yang sudah diproteksi dengan benar di alur approve/process/reject refund (CAS + `RowsAffected()==0`), tapi tidak ada di sini.

---

## 🟠 High

### T5. Bug Kas/Bank yang SAMA seperti kasus refund — `payment_method` dikirim tapi tidak pernah dibaca
**File:** `internal/accounting/service/posting.go:52-54,226-242` (struct `savingsPayload` cuma punya field `amount`), `internal/tabungan/service/service.go:75` (emitter yang benar-benar kirim `payment_method`)
Setoran tabungan via transfer bank tetap selalu diposting ke `AccKas`, tidak pernah ke `AccBank` — karena field-nya memang dikirim oleh tabungan-service tapi `posting.go` tidak punya slot untuk membacanya. Test yang ada (`posting_test.go`) juga tidak pernah menguji kasus `payment_method` non-tunai untuk savings, jadi regresi ini tidak akan pernah tertangkap CI.

### T6. `listDeposits` tidak filter `org_id` — aman sekarang, rapuh ke depan
**File:** `internal/tabungan/repository/repository.go:77-93`
Fungsi ini cuma query by `account_id`, mengandalkan caller (`GetAccount`) untuk validasi org duluan. Kode/handler baru mana pun yang memanggilnya langsung tanpa validasi ulang bisa membocorkan histori setoran org lain.

### T7. Jumlah yang dikirim ke invoice-service diputuskan SEBELUM lock tabungan diambil — akar penyebab T1
**File:** `internal/tabungan/repository/repository.go:132-149` vs `service.go:98-101`
`FOR UPDATE` di `ConvertTx` sudah benar untuk melindungi tabungan sendiri dari saldo negatif — tapi karena keputusan "berapa yang mau di-convert" dibuat sebelum lock itu diambil, dan efek di sisi invoice-service sudah permanen sebelum lock tabungan dicek, tidak ada cara untuk rollback sisi invoice kalau sisi tabungan ternyata gagal.

---

## Frontend — bug, bukan cuma cleanup

### T8. Tombol "Setor" (deposit) MATI TOTAL — di web maupun mobile — karena mismatch "active" vs "aktif"
**File:** `frontend-svelte/src/lib/pages/TabunganPage.svelte:139,161`, `frontend-svelte/src/lib/mobile/screens/Tabungan.svelte:96`
Frontend cek `status === "active"` (Inggris), backend selalu balikin `"aktif"` (Indonesia — `model.go:11`, default kolom DB). Tombol setor tidak pernah muncul di web, kartu tidak pernah bisa di-tap di mobile — fitur setor yang backend-nya sudah benar dan siap pakai, sama sekali tidak bisa diakses dari UI manapun.

### T9. Badge status selalu tampilkan teks mentah "aktif", bukan label yang dimaksud
**File:** `frontend-svelte/src/lib/mobile/screens/Tabungan.svelte:12`, `TabunganPage.svelte:32`
`STATUS_LABEL` map pakai key Inggris (`active`), tidak pernah match `"aktif"` dari backend — jatuh ke fallback raw string.

### T10. Fitur "Convert" (tabungan → pelunasan invoice) sudah lengkap di backend & API client, tapi TIDAK ADA satu pun tombol/modal di UI manapun untuk memicunya
**File:** `frontend-svelte/src/lib/services/apiDomains/tabunganApi.js:29` (`convertTabungan`, terwire dengan benar ke `ApiService`)
`grep` seluruh frontend cuma menemukan definisinya, nol pemanggilan dari komponen manapun. Kombinasi dengan T8 (tombol setor mati) berarti alur tabungan end-to-end (setor → convert ke invoice) praktis tidak bisa dijalankan sama sekali dari aplikasi hari ini.

---

## Tidak ada masalah (dikonfirmasi bersih)
- `ConvertTx`/`DepositTx` atomik penuh DI DALAM database tabungan sendiri (insert + update saldo + outbox dalam satu transaksi) — masalahnya cuma di lintas-service.
- `SettleFromCredit` di invoice-service benar mengunci baris & meng-cap `applied = min(amount, remaining)` — tidak bisa over-apply per satu panggilan.
- Constraint `CHECK (balance >= 0)` di level DB benar-benar mencegah saldo negatif dalam satu transaksi.
- Semua route utama (`ListAccounts`, `GetAccount`, `CreateAccount`, dll) sudah org-scoped dengan benar dan digating `owner/admin/finance`.
- Tidak ada mismatch HTTP method/route di `tabunganApi.js` — semua 5 method cocok dengan backend, dan `createTabunganApi` sudah ter-wire ke `ApiService`.

## Rekomendasi urutan perbaikan
1. **T3** — paling mendesak dari sisi risiko kepercayaan: tambah validasi `jamaah_id` akun tabungan == `jamaah_id` invoice sebelum convert.
2. **T1, T2, T7** — butuh desain ulang alur convert: idempotency key dikirim ke invoice-service, atau ubah urutan (lock+decide saldo tabungan DULU, baru panggil invoice-service, dengan kemampuan compensating-cancel kalau gagal).
3. **T4** — tambah idempotency key/dedup di endpoint deposit.
4. **T5** — fix cepat & aman: tambah field `payment_method` di `savingsPayload` + branching Kas/Bank yang sama seperti payment/refund (persis pola yang sudah ada, tinggal disalin).
5. **T8, T9** — fix cepat & aman: samakan string status (pakai `"aktif"` di frontend, atau sentralisasi konstanta status).
6. **T10** — perlu keputusan produk: tambahkan UI untuk convert, atau sengaja disembunyikan (kalau begitu, kenapa endpoint & client-nya sudah lengkap dibangun?).
