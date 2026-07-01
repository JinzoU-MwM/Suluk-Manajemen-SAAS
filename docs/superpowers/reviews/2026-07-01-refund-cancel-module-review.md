# Refund / Invoice-Cancel / Kloter-Cancel Module — Full Code Review

**Tanggal:** 2026-07-01
**Cakupan:** Bukan diff review — audit menyeluruh kode yang SUDAH ada (termasuk kode lama, di luar scope kerja hari ini). Modul yang dicakup: `internal/invoice/{repository,service,handler}/refund.go` + cancel-invoice path, `internal/jamaah/service/service.go` (cascade gagal-berangkat), `internal/jamaah/repository/group.go` (kloter/pembatalan grup), semua frontend API client (`frontend-svelte/src/lib/services/apiDomains/*.js`), dan layar mobile terkait invoice/refund/approval.
**Status:** Dokumentasi temuan untuk sesi perbaikan berikutnya. Belum ada fix yang diterapkan dari daftar ini (kecuali yang sudah dikerjakan sebelumnya di sesi yang sama: bug PATCH/POST cascade, `ApiService.listInvoices` wiring, migration backfill, falsy-zero `refund_pct`, gofmt).

---

## Ringkasan eksekutif

30 temuan, dikelompokkan berikut. Yang paling mendesak: **ada beberapa jalur konkret di mana uang bisa direfund dua kali atau invoice yang sudah lunas dibatalkan tanpa jejak refund sama sekali** — ini bukan risiko teoritis, tapi urutan aksi yang jelas (staff re-drag setelah error, race retry di HTTP client, dsb).

| Level | Jumlah | Tema |
|---|---|---|
| 🔴 Critical (uang bisa salah / hilang / dobel) | 7 | Double-refund, invoice lunas dibatalkan tanpa refund, kloter cancel senyap |
| 🟠 High (race condition / data integrity / gap signifikan) | 9 | TOCTOU race, retry-amplifikasi error, kebijakan refund dekoratif, invoice ambigu |
| 🟡 Medium (bug nyata, dampak lebih sempit) | 10 | payment_method tidak muncul di API read, dead code berbahaya, error leak, index kurang |
| ⚪ Low (housekeeping) | 4 | Dead code aman, field tidak terpakai |

---

## 🔴 Critical

### C1. Refund bisa dobel: `InitiateRefund` tidak cek total refund lain yang masih terbuka
**File:** `internal/invoice/service/refund.go:25`
Cuma cek `req.Amount <= inv.AmountPaid` — tidak menjumlahkan refund `pending`/`approved`/`processed` lain yang sudah ada untuk invoice yang sama (karena `amount_paid` baru dikurangi saat `CompleteRefund`, bukan saat initiate).
**Skenario:** Invoice `amount_paid=4.000.000`. Refund #1 diajukan 3.500.000 (lolos). Refund #2 diajukan 3.500.000 (lolos juga, karena `amount_paid` belum berubah). Kedua-duanya di-approve+process+complete → 7.000.000 direfund dari yang cuma dibayar 4.000.000, `CompleteRefund` cuma nge-clamp `newPaid` ke 0, tidak error.

### C2. Invoice lunas bisa dibatalkan tanpa refund sama sekali
**File:** `internal/invoice/service/service.go:188`
`CancelInvoice`/`CancelInvoiceTx` tidak ada guard untuk status `lunas`. Kalau `amount_remaining==0`, tidak ada outbox event dibuat (karena logic-nya cuma reverse sisa piutang), tapi invoice tetap berpindah ke `batal` — `amount_paid` yang sudah dibayar penuh tidak pernah disentuh, tidak ada refund record, tidak ada jurnal.

### C3. Race retry di `POST /invoices/:id/refund` bisa bikin refund dobel dalam SATU kali panggilan cascade
**File:** `internal/invoice/service/refund.go:20-48`, `internal/shared/httpclient/httpclient.go:107-167`
`InitiateRefund` = INSERT langsung, tidak idempotent, tidak ada idempotency key. `httpclient` retry otomatis kalau koneksi putus/response gagal dibaca — kalau response-nya hilang setelah server sudah commit INSERT, client retry dan bikin baris refund kedua, tanpa staff melakukan apa pun.

### C4. Jalur konkret: staff re-drag setelah error palsu → cascade jalan dua kali → refund dobel
**File:** `internal/jamaah/service/service.go:419-528`, `internal/jamaah/handler/handler.go:196-203`
`cascadeGagalBerangkat` jalan setelah status pipeline commit. Kalau `GetRegistration` di akhir gagal (DB blip sesaat), handler return 500 dan `CascadeResult` yang sudah berhasil dibuang. Frontend rollback UI + toast error generik. Staff yang lihat "gagal" secara alami akan re-drag kartu yang sama → cascade jalan lagi → refund kedua terbuat untuk invoice yang (kalau `amount_remaining==0` di percobaan pertama) statusnya bahkan belum sempat jadi `batal`, jadi masih lolos filter cascade kedua.

### C5. Pembatalan kloter/grup: TIDAK ADA cascade sama sekali, dan tidak ada jejak apa pun
**File:** `internal/jamaah/repository/group.go:125-157` (`TransitionDeparture`)
Membatalkan satu kloter cuma `UPDATE groups SET departure_status='batal'`. Tidak ada event yang di-emit untuk transisi `batal` (beda dengan `siap`/`berangkat` yang emit event). Semua invoice jamaah dalam grup itu (bisa puluhan orang) tetap `unpaid`/`paid`, tidak ada satu pun yang otomatis di-cancel/direfund, dan **tidak ada worklist/sinyal apa pun** yang menunjukkan ada yang perlu ditindaklanjuti — ini lebih parah dari kasus per-jamaah (yang setidaknya menghasilkan refund `pending` yang terlihat di menu Pembatalan).

### C6. Mobile "Ajukan Refund" (Pembatalan.svelte) SELALU gagal — field `amount` tidak pernah dikirim
**File:** `frontend-svelte/src/lib/mobile/screens/Pembatalan.svelte:40-47`
Form cuma kumpulkan `invoice_id`, `reason`, `refund_pct` — `amount` tidak pernah dikirim di body. Backend (`internal/invoice/handler/refund.go:68-70`) reject dengan 400 kalau `Amount <= 0`. **Ini artinya alur refund manual di mobile — yang di investigasi awal sesi ini disebut sebagai "yang sudah berfungsi" dibanding web — sebenarnya juga rusak total**, cuma gagalnya beda cara (400 di server, bukan crash JS).

### C7. TOCTOU race: cancel invoice baca `amount_remaining` di luar transaksi, bisa posting jurnal basi
**File:** `internal/invoice/service/service.go:189`
`CancelInvoice` baca invoice via `GetInvoiceByID` tanpa lock, baru transaksi cancel jalan belakangan. Kalau ada pembayaran masuk (lewat jalur lain yang terkunci dengan benar) di antara baca dan commit, event `invoice.cancelled` yang dibuat tetap pakai nilai `remaining` yang sudah basi → jurnal akuntansi salah untuk invoice yang sebenarnya sudah lunas duluan.

---

## 🟠 High

### H1. `ErrAlreadyCancelled` dipetakan ke HTTP 500, bukan 4xx → retry otomatis memperkuat laporan error palsu
**File:** `internal/invoice/handler/handler.go:159-161`
Semua error dari `CancelInvoice` — termasuk "sudah dibatalkan" — jadi 500. Karena `httpclient` retry pada status ≥500, satu kali blip transient di percobaan pertama (yang sebenarnya BERHASIL cancel) bikin percobaan retry kedua gagal dengan "already cancelled" yang juga dilaporkan sebagai 500 → cascade melaporkan `InvoiceCancelled=false` padahal invoice-nya genuinely sudah ter-cancel.

### H2. Kartu CRM tetap bisa di-drag saat request sedang berjalan → race pipeline-status tanpa lock
**File:** `frontend-svelte/src/lib/pages/CRMPage.svelte:311-348,536-538`; `internal/jamaah/repository/repository.go:237-270`
Tidak ada guard "busy"/disabled selama PATCH in-flight. Staff bisa drag kartu ke Batal, lalu langsung drag lagi ke status lain sebelum request pertama selesai → dua PATCH konkuren, repo layer cuma `UPDATE` polos tanpa lock/versi. Cascade dari request pertama tetap jalan (irreversible) apa pun hasil race-nya di DB.

### H3. Kebijakan refund (`RefundPolicy`) 100% dekoratif — tidak pernah dipakai untuk hitung persentase
**File:** `internal/invoice/service/refund.go:20-48`, `frontend-svelte/src/lib/pages/CancellationPage.svelte:154-158`
Admin bisa buat kebijakan "H-30: 50%, H-7: 20%" di UI, tapi `InitiateRefund` ambil `refund_pct` mentah dari body request — tidak pernah lookup kebijakan. Form persentase selalu default 100% dan tidak auto-terisi dari kebijakan yang relevan. Tidak ada `policy_id` di tabel `refunds` sama sekali, jadi bahkan audit "refund ini pakai kebijakan apa" tidak mungkin dilakukan.

### H4. Jamaah bisa punya 2 invoice non-batal untuk paket yang sama → cascade bisa pilih invoice yang salah
**File:** `internal/jamaah/service/service.go:428-437`, `internal/jamaah/service/service.go:532-545` (`RemoveFromPackage`)
`RemoveFromPackage` hapus registrasi tanpa pernah memanggil invoice-service — invoice lama (mungkin sudah dibayar sebagian) jadi menggantung. Kalau jamaah didaftarkan ulang ke paket yang sama, ada 2 invoice untuk 1 kombinasi jamaah+paket. Cascade ambil "match pertama" (`created_at DESC`, jadi invoice TERBARU) — kalau invoice terbaru itu yang belum dibayar, invoice lama yang justru sudah dibayar tidak pernah disentuh/direfund.

### H5. Vendor bill payments tidak pernah posting jurnal (ditemukan sebagai efek samping audit posting.go)
**File:** `internal/vendor_svc/service/service.go:276` (`CreatePayment`)
Tidak ada event `vendor.payment.made` atau semacamnya di `events.go`/`posting.go` — bayar tagihan vendor tidak pernah menghasilkan jurnal Dr Hutang Vendor / Cr Kas. Di luar cakupan modul refund, tapi relevan untuk modul vendor besok.

### H6. `RefundHandler` bocorkan pesan error mentah ke client
**File:** `internal/invoice/handler/refund.go:38` (dan beberapa titik lain: 146,155,169,188,201)
Tidak pakai `response.Internal` (yang sudah ada dan dipakai handler lain) — error DB mentah (bisa berisi nama tabel/kolom/host internal) langsung dikirim ke client di setiap kegagalan.

### H7. `GetPayments` tidak filter `org_id` sama sekali
**File:** `internal/invoice/repository/repository.go:224`
Aman hari ini HANYA karena `InitiateRefund` sudah validasi org lewat `GetInvoiceByID` sebelumnya — tapi fungsi ini sendiri tidak defense-in-depth. Refactor apa pun ke depan yang menghapus/menukar urutan validasi itu bisa membocorkan `bank_name`/`account_number`/`reference_number` pembayaran org lain.

### H8. Refund card di mobile (Approval & Pembatalan) selalu tampilkan "Jamaah"/"Refund" generik
**File:** `frontend-svelte/src/lib/mobile/screens/Approval.svelte:19-25`, `Pembatalan.svelte:63-64`
`model.Refund` tidak punya field nama jamaah/invoice/paket — query refund cuma select kolom `refunds` sendiri. Dengan 2+ refund pending, approver tidak bisa membedakan mana-mana sebelum approve/reject.

### H9. Rollback race di CRM: `persistStage` bisa timpa state baru dengan state basi
**File:** `frontend-svelte/src/lib/pages/CRMPage.svelte:333-348`
Kalau ada 2 request `persistStage` konkuren (lihat H2) dan yang lebih lama gagal belakangan, `catch` block-nya pakai `prev` miliknya sendiri (state SEBELUM drag manapun) untuk rollback — menimpa hasil request yang lebih baru yang sudah sukses & sesuai server.

---

## 🟡 Medium

### M1. `payment_method` tidak pernah di-SELECT di 3 endpoint read refund
**File:** `internal/invoice/repository/refund.go:45` (`ListRefunds`), `:74` (`GetRefund`), `:195` (`GetRefundsByInvoice`)
Kolom tersimpan benar (`CreateRefund` insert, `CompleteRefund` baca internal dengan benar), tapi `GET /refunds`, `GET /refunds/:id`, `GET /invoices/:id/refunds` selalu balikin `"payment_method": ""`. Finance tidak bisa lihat dari akun mana refund akan keluar sebelum approve.

### M2. `CancelInvoiceWithRefund` dead code, berbahaya kalau dipakai
**File:** `internal/invoice/repository/refund.go:311`
Tidak pernah dipanggil. Semantiknya beda total dari alur asli (skip tabel `refunds`, skip approval, skip event akuntansi). Kalau ada yang nemu dan pakai karena namanya cocok, refund akan lolos tanpa approval dan tanpa jejak akuntansi.

### M3. `InvoiceRepo.CancelInvoice` dead code (real path-nya `CancelInvoiceTx`)
**File:** `internal/invoice/repository/repository.go:98`
Tidak dipanggil dari mana pun — service pakai `CancelInvoiceTx` (di `outbox.go`). Fungsi lama ini tidak emit outbox event.

### M4. `UpdateInvoiceStatus` generic setter, sama sekali tidak dipakai
**File:** `internal/invoice/repository/repository.go:109`

### M5. Filter tab "Jatuh Tempo" di Invoice.svelte (mobile) tidak pernah menampilkan apa pun
**File:** `frontend-svelte/src/lib/mobile/screens/Invoice.svelte:15,18`
Status `jatuh_tempo` tidak ada di enum backend (`belum_bayar|sebagian|lunas|batal`) — overdue itu konsep turunan (status + due_date), tidak pernah dimaterialisasi jadi status sendiri.

### M6. Nama jamaah selalu kosong ("Jamaah") di daftar invoice mobile
**File:** `frontend-svelte/src/lib/mobile/screens/Invoice.svelte:52`
`model.Invoice` tidak punya field nama, query list tidak join ke tabel jamaah.

### M7. KPI "Total tagihan"/"Diterima" di Invoice.svelte cuma jumlahkan 50 invoice pertama
**File:** `frontend-svelte/src/lib/mobile/screens/Invoice.svelte:32-33`
Ada endpoint `GET /invoices/summary` yang sudah benar (`GetSummary`) tapi tidak dipakai di sini — untuk org dengan >50 invoice, angka KPI understated.

### M8. Prefill invoice dari halaman profil jamaah (Bayar.svelte) tidak pernah match
**File:** `frontend-svelte/src/lib/mobile/screens/Bayar.svelte:29-31`
Bergantung pada `jamaah_name` di objek invoice yang selalu undefined (sama akar penyebab dengan M6).

### M9. `invoiceApi.js` `cancelInvoice()` — bug POST/PATCH yang SAMA seperti yang diperbaiki di cascade, tapi versi lama, belum diperbaiki
**File:** `frontend-svelte/src/lib/services/apiDomains/invoiceApi.js:52-59`
Kirim `method: 'POST'` ke `/invoices/:id/cancel`, padahal route-nya PATCH-only. Tombol cancel invoice manapun di web/mobile yang lewat client ini pasti 404 diam-diam.

### M10. Tidak ada composite index `(org_id, created_at)` di tabel `refunds`
**File:** `migrations/invoice/002_refund.up.sql:31`
Index yang ada cuma single-column (`org_id`, `invoice_id`, `status`). Query list refund org-wide (`ORDER BY created_at DESC` + filter `org_id`) tidak punya index gabungan yang pas — makin banyak data, makin lambat.

---

## ⚪ Low

- `RefundActionRequest` (`{Notes string}`) didefinisikan tapi tidak dipakai di handler manapun — kalau ada yang mau tambahkan catatan approval, kelihatannya sudah ada wiring-nya padahal belum.
- `CancelInvoiceWithRefund` juga tidak emit outbox event (bagian dari M2, dicatat terpisah untuk kelengkapan).
- Race-condition di `ApproveRefund`/`ProcessRefund`/`RejectRefund`/`CompleteRefund` sebenarnya **aman** (CAS pattern + `FOR UPDATE` sudah benar) — dikonfirmasi lewat audit, bukan temuan, tapi dicatat supaya tidak dicurigai ulang.
- `internal/accounting/service/posting.go` vs semua event producer lain (payment, payroll, vendor bill, savings, commission, POS) — **tidak ada bug field-hilang lain** seperti kasus `payment_method` refund. Sudah diaudit menyeluruh, bersih.

---

## Rekomendasi urutan perbaikan (besok)

1. **C1, C3, C4** — refund dobel: tambahkan idempotency key di `POST /invoices/:id/refund` (atau constraint unique di level aplikasi: 1 refund `pending/approved/processed` aktif per invoice), dan validasi `InitiateRefund` terhadap SUM semua refund terbuka, bukan cuma `amount_paid`.
2. **C2, C7** — cancel invoice lunas: tambahkan guard `status != 'lunas'` (seperti yang sudah ada di `CancelInvoiceWithRefund` yang dead), dan pindahkan baca `amount_remaining` ke dalam transaksi yang sama dengan `FOR UPDATE`.
3. **C6** — perbaiki `Pembatalan.svelte` (mobile) supaya kirim `amount` (dari `amount_paid` invoice terpilih, sama seperti pola yang sudah dipakai di `CancellationPage.svelte` web).
4. **C5** — kloter cancel: minimal, emit event/buat worklist ketika grup dibatalkan, supaya tidak senyap total.
5. **H1** — ubah `ErrAlreadyCancelled` jadi mapped ke 409/400, bukan 500, supaya tidak memicu retry yang salah arah.
6. **M1, M9** — dua fix cepat & aman: tambah `payment_method` ke 3 query SELECT; ganti `POST` jadi `PATCH` di `invoiceApi.js:cancelInvoice`.
7. Sisanya (H2/H9 race di CRM, H3 refund policy enforcement, H4 invoice ambigu, H6/H7 hardening, M2-M5/M10 cleanup) — realistis untuk backlog terpisah, bukan hotfix.
