# Payroll Module — Full Code Review

**Tanggal:** 2026-07-02
**Cakupan:** Audit menyeluruh (bukan diff) — `internal/payroll/{handler,service,repository,model}/*.go`, migrations, `cmd/payroll-service/main.go`, frontend (`payrollApi.js`, `PayrollPage.svelte`, `Payroll.svelte` mobile).
**Status:** Dokumentasi temuan untuk sesi perbaikan berikutnya. Belum ada fix diterapkan.

---

## Ringkasan

| Level | Jumlah | Tema |
|---|---|---|
| 🔴 Critical | 4 | Double-posting payroll, kasbon tidak pernah masuk buku besar, poison-message NATS, desain accrue-then-pay tidak pernah selesai dibangun |
| 🟠 High | 4 | Phantom success di finalize, period bisa dimanipulasi, kasbon-slip linkage mati, foreign-key hilang |
| 🟡 Medium | 3 | SQL injection foot-gun, tombol "Bayar" mobile selalu gagal tapi tetap ditandai lunas, angka KPI salah |

---

## 🔴 Critical

### P1. Payroll slip bisa dobel-posting — race condition tanpa unique constraint
**File:** `internal/payroll/service/service.go:112-118`, `internal/payroll/repository/repository.go:73-78` (`SlipExistsForPeriod`)
Cek "sudah ada slip untuk periode ini?" dan insert slip baru dilakukan di dua request/transaksi terpisah (check-then-act klasik), dan tidak ada `UNIQUE (org_id, employee_id, period)` di migration manapun. Dua klik ganda pada "Run Payroll" bisa menghasilkan 2 baris `salary_slips` + 2 event `payroll.posted` untuk 1 karyawan/1 bulan yang sama — dobel Beban Gaji/Kas/Hutang Pajak di GL.

### P2. Kasbon (cash advance) TIDAK PERNAH posting ke buku besar — issue maupun repay
**File:** `internal/payroll/repository/repository.go:140-147` (`CreateAdvance`), `:170-201` (`RepayAdvance`), `internal/payroll/service/service.go:193,203-209`
Tidak ada `outbox.Insert` di kedua fungsi. Tidak ada tipe event `EventAdvanceIssued`/`Repaid` di `events.go`, tidak ada case di `posting.go`, bahkan tidak ada akun "Piutang Karyawan" di COA standar (`internal/accounting/service/coa.go`). Uang keluar (kasbon) dan uang masuk (pelunasan kasbon) sama-sama tidak pernah tercatat — saldo Kas di GL terus-menerus overstated tanpa batas.

### P3. Slip gaji dengan net negatif bikin pesan "racun" (poison message) yang retry selamanya di NATS
**File:** `internal/payroll/service/service.go:131-138` (cap yang cuma proteksi gross, bukan net), `internal/accounting/service/posting.go:189` (reject `p.Net < 0`), `internal/accounting/service/consumer.go:14-17` (NAK on error → redelivery)
Contoh: gaji pokok kecil + potongan absen besar bisa menghasilkan `net` negatif meski `gross` masih positif (cap yang ada cuma melindungi `gross`, bukan `net`). Slip tetap tersimpan & tampil di UI sebagai "berhasil", tapi event `payroll.posted`-nya ditolak permanen oleh posting.go dan terus di-redeliver oleh NATS — jurnal untuk karyawan ini TIDAK PERNAH terpost, dan pesan ini mengganggu consumer group selamanya sampai ada intervensi manual.

### P4. Event `payroll.posted` ditembak saat slip masih DRAFT, bukan saat benar-benar dibayar — desain accrue-then-pay tidak pernah selesai dibangun
**File:** `internal/payroll/service/service.go:162` (event fired di `CreateSalarySlip`, status masih `'draft'` per `outbox_tx.go:24`), `FinalizeSlip` (`service.go:172-174`) tidak emit apa pun, akun `AccHutangGaji` ("Hutang Gaji") ada di COA (`coa.go:14,45`) tapi tidak pernah dipakai di posting manapun.
Setiap slip DRAFT langsung mendebit `AccKas` di GL — seolah uang sudah keluar — padahal belum tentu dibayar (tidak ada fitur void/cancel slip sama sekali). Akun Hutang Gaji yang menganggur kuat mengindikasikan desain aslinya seharusnya: draft → `Dr Beban Gaji / Cr Hutang Gaji`, finalize/bayar → `Dr Hutang Gaji / Cr Kas` — tapi cuma separuh jalan yang terpasang.

---

## 🟠 High

### P5. `UpdateSalarySlipStatus` blind UPDATE — finalize bisa "berhasil" padahal tidak ada baris yang berubah
**File:** `internal/payroll/repository/repository.go:135-138`
Tidak ada `AND status='draft'` di WHERE clause, tidak ada pengecekan `RowsAffected()`. `PUT /slips/{id-salah/org-lain}/finalize` tetap balikin 200 sukses walau 0 baris ke-update — sama seperti pola "phantom success" yang sudah pernah ditemukan & diperbaiki di modul refund.

### P6. Format `period` bebas string, bypass gampang untuk guard duplikat
**File:** `internal/payroll/model/model.go:97`, `service.go:112`, `repository.go:75`
`SlipExistsForPeriod` cocokkan `period` secara exact-string. `"2026-07"` dan `"2026-7"` (sama-sama ≤7 karakter, lolos validasi kolom `VARCHAR(7)`) dianggap periode berbeda — bikin slip dobel untuk bulan yang sama tanpa perlu race condition sama sekali.

### P7. `advance_repayments.salary_slip_id` tidak divalidasi (tidak ada FK, tidak ada cek org/employee)
**File:** `internal/payroll/repository/repository.go:170,191`, migration terkait
Client bisa kirim UUID slip mana pun (bahkan dari org lain / tidak ada) sebagai `salary_slip_id` saat melunasi kasbon — tersimpan tanpa validasi, merusak audit trail rekonsiliasi kasbon vs payroll run.

### P8. `SalarySlip.AdvanceDeduction` field mati — potongan kasbon tidak pernah masuk perhitungan net
**File:** `internal/payroll/model/model.go:33`, `service.go:140-151`
Field ini dibind di setiap query SQL tapi tidak pernah di-assign di kode aplikasi — selalu 0. Pelunasan kasbon lewat slip tertentu (`salary_slip_id` opsional di `RepayAdvance`) tidak pernah memengaruhi `gross`/`net` slip itu.

---

## 🟡 Medium

### P9. `UpdateEmployee` — SQL injection foot-gun laten (belum exploitable, tapi rapuh)
**File:** `internal/payroll/repository/repository.go:80-93`
Kolom UPDATE dibangun via `fmt.Sprintf("%s", k)` dari key map Go tanpa allow-list. Aman hari ini karena caller selalu hardcode key literal — tapi refactor apa pun ke depan yang menurunkan key dari nama field JSON request (pola yang sudah dipakai di tempat lain di codebase ini) akan membuka SQL injection langsung di tabel gaji/BPJS/pajak karyawan.

### P10. (Frontend, mobile) Tombol "Bayar" SELALU gagal di backend, tapi UI tetap tandai "✓ Dibayar"
**File:** `frontend-svelte/src/lib/mobile/screens/Payroll.svelte:53-61`
`bayar(e)` set `paid[e.id]=true` duluan (optimistic), baru panggil `createSalarySlip({employee_id: e.id})` — tanpa `period`, yang wajib di backend (`handler.go:104-106`, selalu 400). Error di-catch dan disembunyikan (toast "Slip dicatat (lokal)"), baris tetap permanen menampilkan "Dibayar" walau tidak ada slip yang pernah tercipta di server — silent lie ke staff.

### P11. (Frontend, mobile) KPI "Total gaji" baca field yang tidak ada, jatuh ke perhitungan lokal yang salah
**File:** `frontend-svelte/src/lib/mobile/screens/Payroll.svelte:51`
`summary?.total_payroll ?? summary?.total ?? ...` — backend cuma balikin `monthly_payroll` (bukan `total_payroll`/`total`), jadi selalu jatuh ke fallback `employees.reduce(...)` yang menjumlahkan base+allowance SEMUA karyawan (bukan payroll run periode berjalan, tidak memperhitungkan potongan/pajak). Desktop `PayrollPage.svelte` sudah benar (`summary.monthly_payroll`) — cuma mobile yang salah.

---

## Tidak ada masalah (dikonfirmasi bersih)
- Tidak ada mismatch HTTP method/route di `payrollApi.js` vs backend (15 method dicek, semua cocok).
- `createPayrollApi` sudah ter-wire dengan benar ke `ApiService`.
- Payload event `payroll.posted` (`gross`/`tax`/`net`) cocok persis dengan yang dibaca `posting.go` — tidak ada bug field-hilang seperti kasus refund.
- `RepayAdvance`'s balance check pakai CAS + `FOR UPDATE` yang benar (pola yang sama dengan yang sudah diperbaiki di modul refund).

## Rekomendasi urutan perbaikan
1. **P1, P6** — tambah `UNIQUE(org_id, employee_id, period)` + normalisasi format period sebelum dibandingkan.
2. **P4** — putuskan desain: apakah payroll.posted harus pindah ke saat finalize (accrue-then-pay dengan `AccHutangGaji`), atau draft memang dimaksudkan langsung final (kalau begitu, hapus akun Hutang Gaji yang menganggur & dokumentasikan keputusannya).
3. **P2** — tambahkan event + GL template untuk kasbon (issue & repay), mirip pola `EventPaymentReceived`.
4. **P3** — floor `net` di validasi sebelum insert slip (jangan cuma cap `gross`), supaya tidak ada slip yang bisa lolos ke poison-message state.
5. **P10** — perbaiki mobile Payroll.svelte supaya kirim `period`, dan jangan optimistic-mark sebelum request sukses.
6. Sisanya (P5, P7, P8, P9, P11) — backlog, bukan hotfix mendesak.
