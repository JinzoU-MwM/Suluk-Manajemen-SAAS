# Accounting Module (Engine + Reporting) — Full Code Review

**Tanggal:** 2026-07-02
**Cakupan:** Audit menyeluruh (bukan diff) — mesin akuntansi inti (`internal/accounting/service/{consumer,service,coa}.go`, `internal/shared/outbox/*.go`, `internal/shared/events/*.go`), plus layer insight/reporting (`insights.go`, `insights_cache.go`) dan frontend (`accountingApi.js`, `AkuntansiPage.svelte`, `Akuntansi.svelte` mobile).
**Status:** Dokumentasi temuan untuk sesi perbaikan berikutnya. Belum ada fix diterapkan.

**Catatan:** modul ini adalah "penerima" dari semua modul lain (refund, payroll, tabungan, vendor) — bug di produsen event sudah didokumentasikan di review masing-masing modul. Dokumen ini fokus ke MESIN-nya sendiri: apakah post-once/idempotency-nya benar-benar aman, apakah ada proteksi di level database, dan apakah laporan yang dihasilkan akurat.

---

## Ringkasan

| Level | Jumlah | Tema |
|---|---|---|
| 🟠 High | 5 | Tidak ada constraint balance di level DB, lock relay outbox tidak benar-benar exclusive, window dedup NATS terbatas, org-scope tidak dipaksa skema, error asli ketimbun jadi "not found" |
| 🟡 Medium | 4 | Insight kas/piutang cuma baca akun COA seed, date-range bisa terbalik diam-diam, cache narasi AI tidak teruji end-to-end, kegagalan Gemini disembunyikan |

**Yang justru SUDAH benar** (dikonfirmasi, bukan temuan): post-once idempotency (`processed_events` + unique index `journals(org_id, source_event_id)`) berjalan atomik dalam satu transaksi bersama insert jurnal — ini fondasi paling penting dan sudah solid.

---

## 🟠 High

### A1. Debit == Kredit cuma dicek di kode Go, TIDAK ADA constraint/trigger di level database
**File:** `internal/accounting/repository/repository.go:127-134` (cek di Go, sebelum transaksi dibuka), `migrations/accounting/001_initial_schema.up.sql:40-53` (skema `journal_lines`)
Constraint yang ada di skema cuma per-baris (`debit>=0 AND credit>=0`, `(debit=0) != (credit=0)`) — tidak ada yang memvalidasi SUM(debit)=SUM(credit) per `journal_id`. Endpoint koreksi manual di masa depan, script backfill, atau bug di template posting baru yang lupa mereplikasi cek ini bisa memasukkan jurnal tidak seimbang langsung ke database, dan Postgres akan menerimanya tanpa protes — merusak trial balance secara diam-diam.

### A2. Lock `FOR UPDATE SKIP LOCKED` di outbox relay TIDAK benar-benar exclusive
**File:** `internal/shared/outbox/outbox.go:65-67` (`FetchUnpublished`), `:126-151` (`processBatch`)
Query `FOR UPDATE SKIP LOCKED` dijalankan lewat `pool.Query()` langsung, bukan di dalam `BEGIN`/`COMMIT` eksplisit — jadi Postgres memperlakukannya sebagai auto-commit single-statement, lock-nya lepas begitu SELECT selesai, JAUH sebelum publish+mark-published dijalankan. Hari ini aman karena cuma 1 instance relay per service — tapi kalau service di-scale ke >1 replica (tidak ada apa pun yang mencegahnya), dua replica bisa fetch batch yang sama dan publish event yang sama dua kali. Bug laten, menunggu horizontal scaling.

### A3. Publish-lalu-mark-published tidak atomik, window dedup NATS cuma 10 menit
**File:** `internal/shared/outbox/outbox.go:142-149`, `internal/shared/events/bus.go:53`
Kalau proses crash di antara `bus.Publish` dan `store.MarkPublished`, baris yang sama akan dipublish ulang setelah restart — dan itu jadi event yang genuinely duplikat di level broker kalau downtime relay lebih dari 10 menit (window `Duplicates` NATS JetStream). Accounting-service sendiri terlindungi (idempotency-nya independen dari transport, lihat catatan di atas), TAPI subscriber lain di stream yang sama (`inventory-deduct`, `jamaah-scoring`) tidak punya proteksi setara — berisiko double-processing.

### A4. `journal_lines.org_id` tidak terikat FK/constraint ke `journals.org_id`, query laporan tidak filter ulang
**File:** `migrations/accounting/001_initial_schema.up.sql:40-53`, `repository.go:303-311,336-344` (`TrialBalance`/`AccountActivity`)
Tidak ada kebocoran aktif hari ini (karena `Post()` selalu menulis `org_id` yang sama ke kedua tabel), tapi tidak ada proteksi skema sama sekali kalau ada jalur tulis baru (endpoint void/reversal, script koreksi) yang suatu saat menulis `journal_lines` dengan `org_id` yang tidak cocok dengan `journal_id`-nya — laporan trial balance/income statement org lain bisa tercampur diam-diam.

### A5. `GetJournal` menelan semua jenis error jadi "journal not found"
**File:** `internal/accounting/repository/repository.go:242-249`, `handler.go:101-104`
Kegagalan koneksi DB, timeout, atau error lain semuanya dipetakan ke 404 generik tanpa logging. Saat terjadi outage/database failover, semua request `GetJournal` melapor "tidak ditemukan" — menyembunyikan masalah infrastruktur nyata dari monitoring, dan bisa menyesatkan orang yang investigasi "kenapa jurnal ini hilang" padahal sebenarnya cuma DB lagi down.

---

## 🟡 Medium (Insights & Reporting)

### A6. Insight "Kas"/"Piutang" cuma baca kode akun seed yang di-hardcode (1101/1102/1201), abaikan akun baru
**File:** `internal/accounting/service/insights.go:39`
Kalau org bikin akun bank kedua (misal kode `1103`), `TrialBalance`/`BalanceSheet` yang sebenarnya sudah benar menghitungnya — tapi dashboard Insight tetap cuma jumlahkan `1101+1102`, mengabaikan `1103` sepenuhnya. Metrik "Kas + Bank" dan deteksi anomali "saldo kas/bank negatif" jadi understated/salah untuk org yang sudah menambah akun.

### A7. `dateRange()` bisa hasilkan rentang terbalik secara diam-diam, bukan error
**File:** `internal/accounting/handler/handler.go:182`
Kalau caller cuma kirim `to` tanpa `from`, default `from` dihitung dari awal bulan BERJALAN — independen dari `to` yang eksplisit. Kalau `to` di bulan lampau, hasilnya `from > to`, query `BETWEEN` tidak match apa pun, dan endpoint balikan 200 dengan semua angka nol alih-alih error yang jelas — laporan periode lama bisa terlihat "kosong" padahal datanya ada.

### A8. Test cache narasi AI tidak pernah menguji end-to-end bahwa perubahan data mengubah cache key
**File:** `internal/accounting/service/insights_cache_test.go:10`
Test yang ada cuma menguji mekanika cache (hit/miss/TTL) dengan key buatan tangan — tidak pernah memanggil `GenerateInsights` betulan sebelum/sesudah simulasi posting jurnal. Kalau prompt generator direfactor dan tanpa sengaja menghilangkan satu metrik dari string yang di-hash, dua kondisi keuangan berbeda bisa menghasilkan cache key sama dan menyajikan narasi basi sampai 6 jam — tidak ada test yang akan menangkap regresi ini.

### A9. Kegagalan panggilan Gemini disembunyikan total dari pengguna
**File:** `internal/accounting/service/insights.go:116`
Error cuma di-log (bahkan hilang total kalau logger nil) — `AIAvailable` tetap `true`, `AINarrative` tetap kosong. Di UI, kartu AI cuma menghilang tanpa penjelasan apa pun (bukan pesan error, bukan fallback message) sampai panggilan berikutnya berhasil.

---

## Tidak ada masalah (dikonfirmasi bersih)
- Post-once idempotency (`processed_events` + `uq_journal_source_event`) atomik & benar — pondasi terkuat dari seluruh mesin.
- Semua titik `outbox.Insert` yang dicek (invoice, vendor, payroll, agent, tabungan, jamaah) memakai `tx` yang sama dengan business-write-nya, bukan pool terpisah — atomicity insert-with-business-write terjaga.
- Tidak ada mismatch route/method di `accountingApi.js`, dan `createAccountingApi` sudah ter-wire dengan benar.
- Tidak ada laporan yang menjumlahkan dari satu halaman paginated saja — semua total pakai endpoint agregat khusus (`/reports/neraca`, `/reports/laba-rugi`).

## Rekomendasi urutan perbaikan
1. **A1** — tambahkan `CHECK`/trigger di level DB untuk memaksa SUM(debit)=SUM(credit) per journal, sebagai pengaman terakhir kalau app-layer check pernah gagal.
2. **A2, A3** — bungkus fetch-lock outbox relay dalam transaksi eksplisit yang benar-benar dipegang sampai publish+mark selesai; audit apakah consumer lain (inventory, jamaah-scoring) butuh idempotency setara accounting-service.
3. **A5** — pisahkan error "genuinely not found" dari error infrastruktur di `GetJournal`, jangan telan jadi 404 buta.
4. **A6, A7** — fix cepat: insight baca SEMUA akun kategori Kas/Piutang (bukan kode hardcode), dan `dateRange()` harus error kalau `from>to` alih-alih diam-diam balikan kosong.
5. **A4, A8, A9** — backlog hardening.
