# Kloter (Group) Cancel Cascade — Design

**Tanggal:** 2026-07-02
**Status:** Approved (pending spec review)
**Latar belakang:** Finding C5 dari `docs/superpowers/reviews/2026-07-01-refund-cancel-module-review.md`: membatalkan kloter (`TransitionDeparture` ke `batal`) cuma `UPDATE groups SET departure_status='batal'` — tidak ada cascade, tidak ada event, tidak ada worklist. Semua invoice jamaah dalam kloter itu (bisa puluhan orang) tetap `unpaid`/`paid` selamanya, lebih parah dari kasus per-jamaah karena tidak ada sinyal apa pun bahwa ada yang perlu ditindaklanjuti.

---

## 1. Goal & decisions

Saat kloter dibatalkan (`draft`/`siap` → `batal`), setiap jamaah anggota kloter itu diproses lewat cascade yang PERSIS SAMA dengan yang sudah ada untuk kasus gagal-berangkat individual (`cascadeGagalBerangkat`, dibangun 2026-07-01): cancel sisa tagihan yang belum dibayar + ajukan refund `pending` untuk yang sudah dibayar. Tidak ada infrastruktur baru — kloter cancel cuma memanggil fungsi yang sama untuk setiap anggota.

| Keputusan | Pilihan |
|---|---|
| Reuse | Panggil `cascadeGagalBerangkat(ctx, jamaahID, packageID, authToken)` yang sudah ada, sekali per anggota kloter (`jamaahID = member.MemberID`, `packageID = group.PackageID`). Tidak menulis ulang logic cancel/refund. |
| Worklist | Tidak ada UI baru. Refund `pending` yang tercipta dari cascade sudah otomatis muncul di menu Pembatalan yang ada — itulah worklist-nya (sama seperti kasus individual). |
| Concurrency | Anggota diproses dengan concurrency terbatas (worker pool kecil, stdlib `sync.WaitGroup` + semaphore channel — tidak menambah dependency baru) supaya kloter besar tidak membuat satu HTTP request jadi puluhan detik menunggu panggilan sekuensial. |
| Kapan cascade jalan | Hanya kalau `group.PackageID != nil` (kloter `draft` tanpa paket belum pernah punya invoice apa pun — tidak ada yang perlu di-cancel/refund) dan `member_count > 0`. |
| Uang bergerak kapan | Sama seperti kasus individual: **tidak pernah otomatis**. Cascade cuma cancel invoice (reverse sisa piutang) + bikin refund `pending`. Finance tetap approve/process/complete manual. |
| Response ke client | `TransitionDeparture` sekarang balikin `{"group": ..., "cascade": GroupCascadeSummary}` (mirip pola `UpdatePipelineStatus`), berisi hitungan agregat (bukan detail per-anggota) — cukup untuk toast ringkasan di frontend. |
| Frontend | Tombol "Batalkan" di `GroupsPage.svelte` sekarang minta konfirmasi dulu (belum ada sama sekali sebelumnya) karena konsekuensinya sekarang jauh lebih besar (cascade finansial ke semua anggota), dan menampilkan toast ringkasan hasil cascade. |

**Di luar cakupan:** Tidak ada dashboard baru untuk memantau progres cascade kloter besar secara real-time — toast ringkasan setelah request selesai sudah cukup untuk MVP ini. Reaktivasi kloter (`batal → draft`) tidak memicu apa pun (tidak ada "un-cascade" — kalau staff salah batalkan lalu reaktivasi, refund yang sudah terlanjur `pending` tetap ada di menu Pembatalan dan harus ditolak manual oleh finance; ini sama dengan risiko yang sudah diterima di kasus individual per spec 2026-07-01).

## 2. Architecture

```
Frontend: klik "Batalkan" di GroupsPage.svelte → confirm() → transitionGroupDeparture(groupId, 'batal')
  → PATCH /groups/:groupId/departure/status  {status: "batal"}

JamaahService.TransitionDeparture (unchanged: gate check, repo.TransitionDeparture persist) →
  if to == DepartureBatal && g.PackageID != nil && g.MemberCount > 0:
      members = repo.ListGroupMembers(groupID)
      summary = cascadeGroupCancelled(ctx, members, *g.PackageID, authToken)   // bounded-concurrency loop over cascadeGagalBerangkat

  response: { group: <updated group>, cascade: GroupCascadeSummary }
```

`cascadeGroupCancelled` cuma memanggil `cascadeGagalBerangkat` per anggota (fungsi yang sudah ada, tidak diubah) dan mengagregasi hasilnya — tidak ada logic cancel/refund baru yang ditulis.

## 3. Components & interfaces

**`internal/jamaah/service/service.go`:**
- Tipe baru `GroupCascadeSummary{MembersProcessed, InvoicesCancelled, RefundsInitiated int}`.
- Fungsi baru `cascadeGroupCancelled(ctx, members []model.GroupMember, packageID uuid.UUID, authToken string) GroupCascadeSummary` — worker pool kecil (cap 5 konkuren) yang memanggil `cascadeGagalBerangkat` per anggota dan menjumlahkan hasilnya lewat mutex.
- `TransitionDeparture`: setelah `repo.TransitionDeparture` sukses dan `to == DepartureBatal`, panggil `cascadeGroupCancelled` kalau `PackageID != nil && MemberCount > 0`. Return tambahan `GroupCascadeSummary`.

**`internal/jamaah/handler/departure.go`:**
- `TransitionDeparture`: teruskan `c.Get("Authorization")`; response jadi `{"group": g, "cascade": summary}`.

**Frontend `frontend-svelte/src/lib/pages/GroupsPage.svelte`:**
- `transitionDep(status)`: kalau `status === 'batal'`, `confirm()` dulu dengan pesan yang menyebutkan jumlah anggota; tampilkan toast ringkasan cascade dari response.

## 4. Definition of Done
- Membatalkan kloter yang punya paket & anggota memicu cascade cancel+refund untuk setiap anggota, memakai fungsi yang sudah ada dan sudah diuji (`cascadeGagalBerangkat`).
- Refund yang tercipta muncul di menu Pembatalan seperti biasa — tidak ada lagi kloter yang dibatalkan tanpa jejak sama sekali.
- Kloter besar tidak membuat request timeout karena cascade dijalankan dengan concurrency terbatas.
- `go build ./...` + `go test ./...` hijau; `npm run build` hijau.
