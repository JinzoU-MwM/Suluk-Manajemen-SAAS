# Refund untuk Jamaah Gagal Berangkat — Design

**Tanggal:** 2026-07-01
**Status:** Approved (pending spec review)
**Latar belakang:** Investigasi menemukan refund sudah punya backend lengkap (double-entry accounting, state machine pending→approved→processed→completed) tapi 4 gap membuatnya tidak terpakai/salah catat di dunia nyata: (1) web app tidak punya tombol untuk mengajukan refund baru, (2) tidak ada satu aksi yang menghubungkan "jamaah gagal berangkat" di CRM dengan cancel invoice + refund, (3) bug: refund selalu dicatat sebagai keluar dari akun Bank walau pembayaran aslinya tunai, (4) endpoint refund tidak terdokumentasi di `API_REFERENCE.md`.

---

## 1. Goal & decisions

Staff bisa menandai jamaah "Batal berangkat" di CRM dan sistem otomatis (a) membatalkan sisa tagihan yang belum dibayar dan (b) mengajukan refund untuk yang sudah dibayar — tanpa staff harus tahu 3 langkah manual terpisah. Finance tetap approve/process/complete refund seperti biasa lewat menu Pembatalan (uang tidak pernah bergerak otomatis). Staff juga bisa mengajukan refund manual langsung dari web (selama ini cuma bisa dari mobile). Refund tunai tidak lagi salah dicatat sebagai Bank.

Keputusan:

| Keputusan | Pilihan |
|---|---|
| Trigger cascade | Hanya saat `pipeline_status → batal` **dengan** lost-reason code `tidak_jadi` ("Batal berangkat"). Lost reason lain (harga, jadwal, kompetitor, dana, lainnya) tidak memicu apa pun secara finansial — itu lead yang belum tentu punya invoice/pembayaran. |
| Cara trigger dikenali | Frontend CRM mengirim `lost_reason_code` (id stabil, mis. `tidak_jadi`) terpisah dari `lost_reason` (label tampilan, bisa berubah). Backend mencocokkan pada code, bukan teks label. |
| Cakupan cascade | Cancel invoice (return sisa tagihan yang belum dibayar) **dan** initiate refund (ajukan `pending` refund untuk yang sudah dibayar) — dua-duanya best-effort, independen satu sama lain. |
| Persentase refund cascade | Selalu 100% dari `amount_paid`. **Tidak** mencoba mencocokkan ke `RefundPolicy` (butuh tanggal keberangkatan + logic pemilihan kebijakan — di luar cakupan). Finance bisa Tolak dan staff re-initiate manual (mobile/web) dengan persentase lebih kecil kalau kebijakan pembatalan mengharuskan potongan. |
| Otorisasi | Cascade memakai `Authorization` token milik user CRM yang melakukan aksi (bukan token service-level). Endpoint cancel/refund di invoice-service sudah digated `RequireRole(owner,admin,finance)` — kalau user CRM bukan role itu, cascade call 403 dan gagal secara *best-effort* (tidak menggagalkan update pipeline status). |
| Uang bergerak kapan | **Tidak pernah otomatis.** Cascade cuma membuat refund `pending`. Finance tetap harus approve → process → complete manual di menu Pembatalan — jurnal baru terpost di langkah "Tandai Selesai", sama seperti alur manual sekarang. |
| UI refund manual di web | Tambahkan tombol "Ajukan Refund" di `CancellationPage.svelte` (state `showNewRefundDrawer`/`refundForm` sudah ada tapi dead — tidak ada tombol/drawer yang memakainya). Pakai pola form yang sama seperti mobile: pilih invoice dari dropdown, amount auto-terisi dari `amount_paid`. |
| Payment-method bug | Refund otomatis mewarisi `payment_method` dari pembayaran **terakhir** pada invoice tersebut (`GetPayments` sudah urut `paid_at DESC`) — tanpa field baru di form manapun. Kalau tidak ada histori pembayaran, default `transfer_bank` (sama seperti default kolom `payments.payment_method`). |

**Di luar cakupan:** cascade untuk pembatalan kloter/grup (`group.cancelled` — event ini bahkan belum ada); pemilihan `RefundPolicy` otomatis berdasarkan hari-sebelum-keberangkatan; UI untuk memilih `payment_method` manual saat mengajukan refund; menghapus dead code `CancelInvoiceWithRefund` yang sudah ada di `internal/invoice/repository/refund.go:310` (tidak dipakai, dibiarkan — bukan bagian dari gap yang dilaporkan).

## 2. Global constraints

- Setiap langkah cascade (cancel invoice, initiate refund) adalah **best-effort dan independen**: kegagalan salah satu tidak boleh menggagalkan update `pipeline_status` yang sudah berhasil, dan tidak boleh menggagalkan langkah lainnya. Kegagalan dicatat lewat logger (`s.log`, kalau ada) dan dilaporkan balik ke frontend lewat field `cascade` di response supaya staff tahu harus lanjut manual di menu Pembatalan.
- Cascade **tidak boleh** memindahkan uang sendiri — hanya `CompleteRefund` (klik "Tandai Selesai" oleh finance) yang boleh memicu jurnal akuntansi. Ini konsisten dengan alur manual yang sudah ada.
- `jamaah-service` memanggil `invoice-service` lewat `httpclient` yang sudah ada (`s.httpc`, `s.invoiceAddr`) — pola yang sama persis dengan `crm.go:121` (`GetJSON` ke `/api/v1/invoices/balances`). Tidak menambah client/dependency baru.
- Perubahan skema (`refunds.payment_method`) harus backward-compatible: `NOT NULL DEFAULT 'transfer_bank'` supaya baris lama tidak perlu backfill manual.
- Commit message: **tanpa** baris AI co-author (ikut konvensi repo — lihat plan renewal-reminders).

## 3. Architecture & data flow

### 3a. Cascade: jamaah gagal berangkat → cancel invoice + refund

```
CRM (Kanban, drag ke kolom "Batal", pilih alasan "Batal berangkat")
  → PATCH /jamaah/:id/registrations/:pkgId/status
      body: { pipeline_status: "batal", lost_reason: "Batal berangkat", lost_reason_code: "tidak_jadi" }

JamaahService.UpdatePipelineStatus (unchanged logic) →
  repo.UpdatePipelineStatus (persist status + lost_reason + audit row, unchanged)

  if status == "batal" && lost_reason_code == "tidak_jadi":
      cascadeGagalBerangkat(jamaahID, packageID, authToken):
          invoices = GET invoice-service /api/v1/invoices/jamaah/:jamaahId   (via httpc.GetJSON)
          inv = first invoice where package_id == packageID && status != "batal"
          if inv == nil: return {} (nothing to do — no active invoice for this package)

          if inv.amount_remaining > 0:
              POST invoice-service /api/v1/invoices/:id/cancel {reason: "Jamaah gagal berangkat"}   (best-effort)
          if inv.amount_paid > 0:
              POST invoice-service /api/v1/invoices/:id/refund {amount: amount_paid, refund_pct: 100, reason: "Jamaah gagal berangkat"}  (best-effort)

          return CascadeResult{invoice_cancelled, refund_initiated, attempted}

  response: { registration: <updated registration>, cascade: CascadeResult }
```

Refund yang tercipta dari cascade masuk ke menu **Pembatalan** dengan status `pending`, sama seperti refund yang diajukan manual — finance memprosesnya lewat tombol Setujui/Proses/Tandai Selesai yang sudah ada (§3b di plan renewal-reminders-style: tidak ada perubahan di sisi accounting-service sama sekali).

### 3b. Refund manual dari web (gap UI)

`CancellationPage.svelte` sudah punya `showNewRefundDrawer`/`refundForm` di state tapi tidak ada tombol yang membukanya dan tidak ada `<SlideDrawer>` yang merender form itu. Tambahkan:
- Tombol "Ajukan Refund" di header aksi (sebelah tombol "Kebijakan").
- Drawer baru dengan `<select>` invoice (load dari `ApiService.listInvoices`, filter `status !== 'batal'`), amount auto-terisi dari invoice terpilih, submit ke `ApiService.initiateRefund` (client function-nya sudah ada, tinggal dipanggil).

### 3c. Bug fix: refund payment_method

```
Sebelum (bug):
  CompleteRefund → payload {amount, invoice_number}   (tanpa payment_method)
  posting.go → PaymentMethod == "" → selalu masuk AccBank

Sesudah:
  InitiateRefund → payments = GetPayments(invoiceID)   (sudah ORDER BY paid_at DESC)
                 → paymentMethod = payments[0].payment_method jika ada, else "transfer_bank"
                 → CreateRefund menyimpan payment_method ini di baris `refunds`
  CompleteRefund → SELECT ... payment_method FROM refunds ...
                 → payload {amount, invoice_number, payment_method}
  posting.go → cabang Kas/Bank sudah benar tanpa perlu diubah (logic-nya sudah ada, cuma input-nya yang hilang)
```

## 4. Components & interfaces

**Migration** `migrations/invoice/008_refund_payment_method.up.sql` / `.down.sql`:
```sql
ALTER TABLE refunds ADD COLUMN payment_method VARCHAR(30) NOT NULL DEFAULT 'transfer_bank';
```

**`internal/invoice/model/model.go`:**
- `Refund` struct: tambah `PaymentMethod string \`json:"payment_method" db:"payment_method"\``.

**`internal/invoice/repository/refund.go`:**
- `CreateRefund`: INSERT menyertakan `payment_method`.
- `CompleteRefund`: SELECT menyertakan `payment_method`; payload outbox menyertakan `"payment_method"`.

**`internal/invoice/service/refund.go`:**
- `InitiateRefund`: sebelum `CreateRefund`, panggil `s.repo.GetPayments(ctx, invoiceID)` (sudah ada), ambil `payments[0].PaymentMethod` sebagai default kalau ada.

**`internal/jamaah/model/model.go`:**
- `UpdatePipelineStatusRequest`: tambah `LostReasonCode string \`json:"lost_reason_code,omitempty" validate:"max=20"\``.

**`internal/jamaah/service/service.go`:**
- `UpdatePipelineStatus`: tambah parameter `lostReasonCode, authToken string`; return tambahan `CascadeResult`.
- Fungsi baru `cascadeGagalBerangkat(ctx, jamaahID, packageID uuid.UUID, authToken string) CascadeResult` dan tipe `CascadeResult{InvoiceCancelled, RefundInitiated, Attempted bool}`.

**`internal/jamaah/handler/handler.go`:**
- `UpdatePipelineStatus`: teruskan `req.LostReasonCode` dan `c.Get("Authorization")`; response jadi `{registration, cascade}`.

**Frontend `frontend-svelte/src/lib/services/apiDomains/jamaahApi.js`:**
- `updatePipelineStatus`: terima & kirim `lost_reason_code`.

**Frontend `frontend-svelte/src/lib/pages/CRMPage.svelte`:**
- State `lostReasonCode`, di-set bareng `lostReason` saat staff pilih alasan.
- `persistStage`/`confirmBatal`: kirim `lost_reason_code`; tampilkan toast berbeda tergantung hasil `cascade` (berhasil auto vs perlu manual).

**Frontend `frontend-svelte/src/lib/pages/CancellationPage.svelte`:**
- State `invoices` baru (di-load bareng `loadData()`).
- Fungsi `openNewRefund`, `selectInvoiceForRefund`, `saveNewRefund`.
- Tombol + `<SlideDrawer>` baru "Ajukan Refund Baru".

**`internal/accounting/service/posting_test.go`:**
- Dua case baru di `TestBuildPostingBalanced`: `EventRefundCompleted` dengan `payment_method: "tunai"` → `AccKas`, dan `"transfer_bank"` → `AccBank` (mendokumentasikan behavior yang benar; posting.go sendiri tidak berubah karena logic-nya sudah benar — bug-nya di layer yang tidak menyertakan `payment_method`).

**Docs** `API_REFERENCE.md`: tambah section `## Refunds` setelah `## Invoices`, mendaftar semua endpoint `/refunds/*` dan `/invoices/:id/refund` yang sudah live tapi belum terdokumentasi.

## 5. Edge cases

- **Jamaah tanpa invoice** (misal batal saat masih tahap `prospek`/`survey`, belum pernah register/invoice terbit): `GetInvoicesByJamaah` mengembalikan list kosong atau tidak ada yang cocok `package_id` → `cascadeGagalBerangkat` return `CascadeResult{}` (tidak melakukan apa-apa), tidak error.
- **Invoice sudah `batal` sebelumnya**: difilter keluar (`status != "batal"`) — cascade tidak mencoba cancel invoice yang sudah batal.
- **User CRM bukan finance/owner/admin**: kedua call cascade 403 → `CascadeResult{Attempted:true, InvoiceCancelled:false, RefundInitiated:false}` — frontend menampilkan toast "proses manual di menu Pembatalan", pipeline status tetap ter-update.
- **invoice-service down/timeout**: `httpc` sudah retry 2x built-in; kalau tetap gagal, sama seperti kasus di atas — logged, dilaporkan sebagai gagal, pipeline status tetap sukses.
- **Invoice `amount_remaining == 0` (lunas) tapi jamaah gagal berangkat**: hanya langkah refund yang jalan (skip cancel, karena tidak ada sisa tagihan untuk dibalik).
- **Invoice `amount_paid == 0` (belum bayar sama sekali)**: hanya langkah cancel yang jalan (skip refund, karena tidak ada nominal untuk direfund) — `req.Amount <= 0` sudah divalidasi ditolak di handler kalau sampai terpanggil, jadi cascade harus skip proaktif, bukan mengandalkan validasi itu.
- **Refund yang sama sudah pernah diajukan untuk invoice ini** (misal staff sempat drag-batal, undo, drag lagi): tidak ada guard idempotency eksplisit — bisa membuat refund `pending` duplikat. **Diterima sebagai risiko kecil**: finance melihat semua refund `pending` di menu Pembatalan dan akan menolak salah satunya; tidak dianggap cukup sering terjadi untuk menambah state tracking baru di iterasi ini.
- **Payment_method: invoice tanpa histori pembayaran sama sekali tapi `amount_paid > 0`**: tidak mungkin terjadi (amount_paid hanya naik lewat `RecordPayment` yang selalu insert baris `payments`) — kalau toh terjadi karena data lama yang inkonsisten, default `"transfer_bank"` dipakai.

## 6. Testing

- **`posting.go` refund Kas/Bank branching**: sudah ada logic-nya (dipakai juga oleh `payment.received`), ditambahkan 2 test case eksplisit untuk `refund.completed` di `TestBuildPostingBalanced` (tunai→Kas, transfer_bank→Bank) supaya regresi ke depan tertangkap tanpa perlu DB.
- **Repo/SQL layer** (`CreateRefund`, `CompleteRefund`, `GetPayments`, cascade HTTP calls): build-verified + integration, mengikuti konvensi repo ini (tidak ada test DB tersedia — lihat plan renewal-reminders §Self-Review "DB-test honesty").
- **Svelte UI** (drawer baru, cascade toast): build-verified (`npm run build` / dev server smoke test manual) — tidak ada unit test Svelte di repo ini untuk halaman lain juga.
- **Manual verification** (definisi selesai di bawah) menggantikan integration test end-to-end karena butuh Postgres + NATS hidup.

## 7. Definition of Done

- Staff bisa drag jamaah ke "Batal" dengan alasan "Batal berangkat" di CRM dan (kalau user itu owner/admin/finance) invoice otomatis ter-cancel (sisa tagihan) dan refund `pending` otomatis ter-ajukan untuk yang sudah dibayar — tanpa uang berpindah sampai finance klik "Tandai Selesai".
- Kalau user CRM bukan role finance, pipeline status tetap ter-update dan staff diberi tahu untuk memproses manual — tidak ada request yang gagal total.
- Web app (`CancellationPage.svelte`) punya tombol "Ajukan Refund" yang berfungsi, setara dengan yang sudah ada di mobile.
- Refund baru (baik dari cascade maupun manual) mencatat `payment_method` yang benar; refund untuk pembayaran tunai masuk ke akun Kas, bukan Bank, saat `CompleteRefund`.
- `API_REFERENCE.md` mendaftar semua endpoint `/refunds/*` dan `/invoices/:id/refund`.
- `go build ./...` + `go test ./...` hijau; `npm run build` (frontend-svelte) hijau.
