# Vendor Module — Full Code Review

**Tanggal:** 2026-07-02
**Cakupan:** Audit menyeluruh (bukan diff) — `internal/vendor_svc/{handler,service,repository,model}/*.go`, migrations, `cmd/vendor-service/main.go`, frontend (`vendorApi.js`, `VendorsPage.svelte`, `Vendor.svelte` mobile).
**Status:** Dokumentasi temuan untuk sesi perbaikan berikutnya. Belum ada fix diterapkan.

---

## Ringkasan

| Level | Jumlah | Tema |
|---|---|---|
| 🔴 Critical | 3 | Pembayaran vendor tidak pernah masuk buku besar (dan TIDAK ADA cara manual untuk koreksi), overpayment tidak divalidasi, operasi tidak atomik |
| 🟠 High | 4 | Edit/hapus tagihan tidak reverse jurnal, validasi cross-org fail-open, path param diabaikan, tidak idempotent |
| 🟡 Medium | 5 | Rounding drift, 3 bug frontend (form salah, dropdown tipe vendor salah, badge utang selalu 0), endpoint yatim |

---

## 🔴 Critical

### V1. Pembayaran tagihan vendor TIDAK PERNAH posting ke buku besar — dan tidak ada cara manual untuk memperbaikinya
**File:** `internal/vendor_svc/service/service.go:276-326` (`CreatePayment`)
Fungsi ini update `paid_amount`/`status` di database vendor_svc sendiri, tapi tidak pernah panggil `outbox.Insert`. Tidak ada tipe event `vendor.payment.*` di `events.go` sama sekali, dan tidak ada case untuk itu di `posting.go` — bahkan kalau ada yang coba kirim event-nya nanti, sistem akan skip (`ErrNoTemplate`) karena template GL-nya memang belum pernah dibuat. Akun `Hutang Vendor` di neraca terus overstated tanpa batas seiring tagihan benar-benar dibayar. **Diperparah:** tidak ada fitur jurnal manual di seluruh aplikasi (`cmd/accounting-service/main.go` cuma expose `GET /journals`, tidak ada `POST`) — bookkeeper yang sadar selisih ini saat rekonsiliasi akhir bulan sama sekali tidak punya cara memperbaikinya lewat produk, harus lewat kode.

### V2. Tidak ada validasi overpayment — tagihan yang sudah lunas bisa "dibayar" lagi
**File:** `internal/vendor_svc/service/service.go:276-313`
`CreatePayment` tidak pernah bandingkan `req.Amount` dengan sisa tagihan (`amount_idr - paid_amount`), dan tidak ada `CHECK` constraint di skema. Tagihan Rp 10 juta yang sudah lunas masih bisa menerima pembayaran Rp 10 juta lagi — `paid_amount` jadi Rp 20 juta, status tetap "lunas" (di-clamp), tapi ringkasan utang (`GetDebtSummary`) bisa jadi negatif per tagihan.

### V3. `CreatePayment`/`DeletePayment` tidak atomik — bisa crash di tengah, uang tercatat tapi status tagihan tidak update
**File:** `internal/vendor_svc/service/service.go:311-323,339-351`
Insert pembayaran dan update status tagihan adalah 2-3 round-trip DB terpisah, tidak dibungkus transaksi (beda dengan `CreateBillTx` yang sudah benar). Crash di antara keduanya membuat pembayaran tercatat permanen tapi tagihan tetap tampil belum lunas — tidak ada mekanisme reconcile.

---

## 🟠 High

### V4. Edit/hapus tagihan vendor tidak pernah membalik jurnal yang sudah terpost
**File:** `internal/vendor_svc/service/service.go:189-239` (`UpdateBill`, `DeleteBill`)
Setelah tagihan dibuat dan jurnal `Dr Beban/Cr Hutang Vendor` terpost, mengedit nominal/kurs atau menghapus tagihan sama sekali tidak menyentuh outbox — jurnal lama tetap ada selamanya, sekalipun tagihannya sendiri sudah berubah atau hilang dari `vendor_bills`. Kesalahan input kurs yang dikoreksi tidak pernah terkoreksi di buku besar; tagihan yang dihapus meninggalkan jurnal yatim permanen.

### V5. Validasi cross-org package fail-open kalau env var kosong
**File:** `internal/vendor_svc/service/service.go:390-397` (`validatePackage`)
Kalau `PACKAGE_SERVICE_ADDR` kosong (misalnya salah konfigurasi saat deploy), validasi "tagihan ini harus milik paket dari org yang sama" langsung di-skip diam-diam — tidak ada startup check yang gagal keras untuk kasus ini (beda dengan pengecekan JWT key yang sudah benar `Fatal` kalau hilang).

### V6. Path param `:billId` diabaikan — endpoint pembayaran cuma percaya body request
**File:** `internal/vendor_svc/handler/handler.go:363-403`
`POST /bills/:billId/payments` tidak pernah baca `c.Params("billId")` — cuma pakai `req.VendorBillID` dari body. Kalau body dan URL tidak sinkron (form/cache basi), pembayaran bisa tercatat ke tagihan yang berbeda dari yang ditampilkan di layar.

### V7. Tidak ada idempotency di `CreatePayment`
**File:** `internal/vendor_svc/model/model.go:125-134`
Retry jaringan atau klik ganda di tombol "Catat Pembayaran" bikin 2 baris pembayaran terpisah untuk 1 transaksi nyata.

---

## 🟡 Medium

### V8. Selisih pembulatan antara Go (`float64`) dan Postgres (`NUMERIC`) untuk `amount_idr`
**File:** `internal/vendor_svc/service/service.go:170`
Payload event pakai `int64(float64(amount)*rate)` (truncate), kolom DB pakai cast `NUMERIC::BIGINT` (round-to-nearest) — beda aturan pembulatan bisa bikin selisih 1 rupiah per tagihan berkurs pecahan, terakumulasi seiring waktu.

### V9. (Frontend web) Form "Buat Tagihan" tandai paket sebagai opsional padahal backend wajibkan
**File:** `frontend-svelte/src/lib/pages/VendorsPage.svelte:236-239,244-245,892-907`
Label "Trip (opsional)" dan dropdown-nya disembunyikan total kalau org belum punya paket — tapi backend selalu reject tanpa `package_id`. Org baru tanpa paket dijamin gagal setiap kali coba buat tagihan vendor.

### V10. (Frontend mobile) Dropdown tipe vendor tawarkan pilihan yang tidak valid di backend
**File:** `frontend-svelte/src/lib/mobile/screens/Vendor.svelte:18-23`
Opsi "Muassasah", "Transportasi", "Visa", "Handling" semua BUKAN nilai valid backend (yang benar: `maskapai, hotel, transport, perlengkapan, katering, lainnya`) — dan "perlengkapan" yang valid malah tidak ada di daftar mobile. Web sudah benar; ini regresi khusus mobile.

### V11. (Frontend mobile) Badge "utang" per vendor selalu tampil Rp 0
**File:** `frontend-svelte/src/lib/mobile/screens/Vendor.svelte:36,40-43`
Kode berasumsi response `getDebtSummary()` punya breakdown per-vendor (`ds.by_vendor`/`ds.vendors`) — padahal backend cuma balikin satu objek agregat org-wide. Setiap baris vendor dan kartu "Total utang" di mobile selalu menampilkan Rp 0 / badge hijau, menyembunyikan utang riil.

### V12. Endpoint `listPaymentsByVendor` ada di backend tapi tidak pernah dibungkus di frontend client
**File:** `cmd/vendor-service/main.go:118`, `frontend-svelte/src/lib/services/apiDomains/vendorApi.js`
Bukan bug aktif (belum ada yang memanggilnya), tapi jebakan untuk developer masa depan yang berasumsi `ApiService.listPaymentsByVendor` sudah ada (mengikuti pola nama `listPaymentsByBill` yang memang ada) — akan crash undefined kalau langsung dipakai tanpa dicek dulu.

---

## Rekomendasi urutan perbaikan
1. **V1** — paling mendesak dan paling besar dampaknya: tambah `EventVendorPaymentPosted` + case di `posting.go` (`Dr Hutang Vendor / Cr Kas|Bank`, pola sama seperti `EventPaymentReceived`), dan panggil `outbox.Insert` dari `CreatePayment` di dalam transaksi (seperti `CreateBillTx`).
2. **V4** — sekalian tambahkan reversing event untuk `UpdateBill`/`DeleteBill` memakai infrastruktur yang sama dari fix V1.
3. **V2, V3** — validasi sisa tagihan sebelum insert pembayaran + bungkus `CreatePayment`/`DeletePayment` dalam satu transaksi.
4. **V9, V10** — fix cepat frontend: samakan daftar tipe vendor mobile dengan web, jadikan `package_id` wajib terlihat di form.
5. **V11** — fix response-shape assumption di mobile Vendor.svelte.
6. Sisanya (V5, V6, V7, V8, V12) — backlog hardening, bukan hotfix mendesak.
