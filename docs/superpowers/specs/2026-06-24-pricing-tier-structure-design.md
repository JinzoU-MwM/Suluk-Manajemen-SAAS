# Pricing Tier Structure Redesign — Design

**Tanggal:** 2026-06-24
**Status:** Draft (menunggu review)
**Fokus:** Struktur tier (jumlah paket, segmentasi, pembeda antar-tier, mekanik trial/fallback/overage)

---

## 1. Konteks & Tujuan

- **Stage:** Pre-launch — belum ada pelanggan bayar; paket ini buat dipakai pas launching.
- **Tujuan utama:** Akuisisi / pertumbuhan (banyakin agen yang adopsi dulu; ARPU nomor 2).
- **Target beachhead:** Agen umrah **menengah established** — punya kantor + beberapa staff, butuh modul lengkap (keuangan, kontrak, multi-user), mau bayar.
- **Hero tier:** **Pro** — jadi default yang ditonjolin.

### Prinsip desain
1. **Sesederhana mungkin** — pre-launch, belum ada data WTP; jangan launching dengan banyak tier yang dikarang sendiri. Tambah tier setelah ada data.
2. **Free trial > freemium** untuk segmen established — kasih rasain value penuh dulu, baru bayar.
3. **Pro sebagai anchor** — pola Good-Better-Best, Pro di tengah ("Paling Populer").

### Di luar scope (workstream terpisah)
- **Penyetelan angka harga** — harga ditahan di 149k/299k/599k sebagai placeholder. Flag: untuk target established, **Pro 299k kemungkinan kemurahan**; tuning harga dibahas terpisah.
- **Retensi mendalam** — auto-renew, dunning, kartu tersimpan. Di struktur ini cukup reminder perpanjangan.

---

## 2. Baseline (kondisi sekarang)

- **5 tier:** Gratis / Starter / Pro / Bisnis / Enterprise.
- **Billing:** prepaid, non-recurring (Pakasir), tanpa kartu tersimpan, perpanjang manual. Org auto-turun ke Gratis pas expired.
- **Trial:** Pro 14 hari, **opt-in** (`POST /api/v1/subscription/activate-trial`), sekali per org (`trial_used`).
- **Enforcement limit:** `internal/jamaah/service/limits.go` tanya ke auth `/subscription/status`, cache 45 detik, fail-open. Limit cuma ngeblok **penambahan** data baru di atas cap (data lama tetap kelihatan).
- **Gate modul:** `IsProOrHigher()` di `internal/shared/plan/plan.go` ngebuka CRM/keuangan/kontrak untuk Pro ke atas.

**File terkait:** `internal/shared/plan/plan.go`, `frontend-svelte/src/lib/config/pricing.js`, `internal/jamaah/service/limits.go`, `internal/auth/service/subscription.go`, `internal/auth/handler/subscription.go`, `migrations/auth/006_subscriptions.up.sql`, `010_plan_tiers.up.sql`, `016_subscription_cancel.up.sql`, `frontend-svelte/src/lib/components/UpgradeModal.svelte`, `frontend-svelte/src/lib/pages/ProfilePage.svelte`.

---

## 3. Struktur target — Opsi A (Good-Better-Best)

### 3.1 Lineup & peran tiap tier

Yang **dipajang** di pricing page (3 kartu, kiri → kanan):

| Posisi | Paket | Buat siapa | Peran strategis |
|---|---|---|---|
| 1 | **Starter** | Agen kecil/baru, atau yang belum butuh modul berat | Pintu masuk friksi-rendah ("ada opsi terjangkau") |
| 2 | **Pro** ⭐ *Paling Populer* | **Agen menengah established (target)** | **HERO** — full suite, default |
| 3 | **Bisnis** | Agen multi-cabang / multi-PT | Anchor atas + jalur ekspansi |

Yang **tidak** jadi kartu berharga:
- **Gratis** → *fallback diam*. Tempat mendarat setelah trial/expired biar data jamaah gak ilang. Tampil kecil ("atau mulai gratis"), bukan CTA utama.
- **Enterprise** → tombol **"Hubungi Sales"** (WhatsApp), bukan harga.

### 3.2 Matriks pembeda antar-tier

Prinsip: **Starter buat *menjalankan*, Pro buat *menumbuhkan*.** Modul serius (keuangan, kontrak, CRM) ditaruh di Pro biar agen established ketarik ke hero.

| | **Gratis** *(fallback)* | **Starter** | **Pro** ⭐ | **Bisnis** |
|---|---|---|---|---|
| Buat | parkir data | agen kecil / coba serius | **agen menengah** | multi-cabang/PT |
| Harga (placeholder) | 0 | 149k/bln | 299k/bln | 599k/bln |
| Jamaah | 50 | 500 | ∞ | ∞ |
| Staff/user | 1 | 3 | 10 | 25 |
| Grup keberangkatan | 2 | 10 | ∞ | ∞ |
| Data jamaah + cicilan/pembayaran | dasar | ✓ | ✓ | ✓ |
| AI Scanner (paspor/visa) | 5/bln | 100/bln + top-up | ∞ (fair-use) | ∞ (fair-use) |
| Laporan | — | dasar | advanced | advanced |
| CRM (pipeline leads/marketing) | — | — | ✓ | ✓ |
| Keuangan/Akuntansi (+AI copilot) | — | — | ✓ | ✓ |
| Kontrak | — | — | ✓ | ✓ |
| Multi-cabang | — | — | — | ✓ |
| Multi-PT (multi-entity) | — | — | — | ✓ |
| Support | self-serve | email | email + chat prioritas | priority + onboarding |

**Pemicu upgrade (sengaja dibikin jelas):**
- **→ Pro** pas: butuh keuangan/kontrak/CRM, ATAU jamaah > 500, ATAU staff > 3, ATAU kuota scan rutin mentok.
- **→ Bisnis** pas: buka cabang/PT ke-2, ATAU staff > 10.

### 3.3 Overage AI Scanner di Starter — top-up prepaid (Cara A)

Kuota scan = **soft cap**, bukan tembok keras. Cocok 100% sama billing prepaid (gak butuh kartu/auto-charge).

- Kuota 100 habis → **gak diblok**, muncul prompt: *"Kuota scan bulan ini habis. Beli tambahan?"*
- Beli **paket 100 scan = Rp 49.000** (one-time via Pakasir, sama flow kayak beli paket). Kuota nambah untuk bulan berjalan, reset tiap awal bulan.
- **Matematika upsell:** Pro itu +Rp150k dari Starter. Top-up Rp49k/100-scan → ~300 scan ekstra/bulan ≈ Rp147k ≈ *"mending langsung Pro unlimited"*. Jadi overage nangkep yang sesekali rame, dan otomatis ngedorong yang rutin-rame ke Pro.
- **Margin:** biaya Aivene per scan ~Rp 30-80 (gemini-2.5-flash-lite), jadi top-up ~Rp490/scan margin gede tapi tetap murah buat agen.
- **Pro & Bisnis "unlimited"** tetap dikasih **fair-use soft cap** (mis. ~2.000/bln): kasih alert, jangan blokir, biar biaya Aivene aman dari abuse.

### 3.4 Trial & fallback lifecycle

```
   Daftar
     │
     ▼
[Trial Pro 14 hari]  ← otomatis, tanpa bayar/kartu
     │
     ├──── beli paket ───▶ [Starter / Pro / Bisnis] ──(prepaid abis)──┐
     │                                                                 │
     └──── gak upgrade ──▶ [Gratis — fallback diam] ◀───────────────────┘
```

**1. Reverse trial (otomatis pas daftar)**
- Daftar = langsung dapet **Pro full 14 hari**, tanpa diminta, tanpa kartu.
- Banner countdown: *"X hari Pro tersisa — upgrade biar fitur gak keputus."*
- Sekali per org (`trial_used`). **Perubahan dari sekarang:** pemicu trial dari "klik tombol" → jadi **auto pas org dibuat**.

**2. Abis trial / paket expired → turun ke Gratis**
- Gratis = 50 jamaah, 1 user, scan 5/bln, modul dasar.
- **Data di atas cap = read-only, BUKAN dihapus.** Org tetap bisa LIHAT semua data (mis. 300 jamaah dari trial), tapi nambah/edit di atas cap → harus upgrade. → data aman + tekanan re-konversi.
- Ini sejalan sama enforcement sekarang (limit cuma ngeblok penambahan baru), jadi PR engineering kecil.

**3. Renewal (prepaid, manual)**
- Sebelum expired: reminder **H-7 / H-3 / H-1** ("perpanjang biar gak turun ke Gratis").
- Retensi lebih dalam (auto-renew/dunning) = workstream terpisah.

---

## 4. Reuse vs Baru (titik sentuh engineering)

**Di-reuse (udah ada):**
- Tabel `subscriptions` (plan, status, expires_at, trial_used, cancel_at_period_end).
- Auto-downgrade ke Gratis pas expired.
- Enforcement limit fail-open di `limits.go` (cuma ngeblok penambahan di atas cap).
- Aktivasi paket via webhook internal dari invoice-service (Pakasir).
- Definisi plan terpusat di `plan.go` (backend) + mirror `pricing.js` (frontend).

**Baru / berubah:**
1. **Definisi plan** (`plan.go` + `pricing.js`): set ulang limit per tier sesuai matriks 3.2; gate CRM/keuangan/kontrak tetap Pro+; pricing page dari 5 kartu → **3 kartu** (Gratis & Enterprise jadi elemen sekunder).
2. **Reverse trial:** trigger trial otomatis pas org dibuat (bukan opt-in). Logika `ActivateTrial` di-reuse, pemicunya pindah ke flow registrasi/org-creation.
3. **Scan metering + top-up:** counter scan per-org per-bulan (reset awal bulan); SKU "100 scan tambahan" lewat Pakasir; prompt top-up pas kuota habis; fair-use cap untuk Pro/Bisnis (alert, bukan blokir).
4. **Read-only over-cap:** pastikan UX "lihat tapi gak bisa nambah" jelas (banner/CTA upgrade) — perilaku enforcement mostly udah ada.
5. **Reminder perpanjangan:** notifikasi H-7/H-3/H-1. Kanal default **email** (infra Resend/SMTP udah ada); WhatsApp opsional, ngikut workstream retensi. Pemicunya bagian dari struktur ini.

---

## 5. Risiko & catatan

- **Starter "tipis":** karena CRM/keuangan/kontrak naik ke Pro, Starter jadi tipis. Itu memang disengaja (pintu masuk + decoy), tapi pantau biar gak nyedot orang yang harusnya Pro.
- **Pro kemungkinan underpriced** untuk target established — flag buat workstream harga.
- **Kompleksitas top-up:** metering + SKU baru nambah PR, tapi ngilangin friksi tembok-keras (net positif buat akuisisi/retensi).
- **Pre-launch = asumsi:** semua angka (limit, kuota, harga) adalah hipotesis; siap di-tweak begitu ada data pemakaian nyata.

---

## 6. Definition of Done (untuk implementasi nanti)

- Pricing page nampilin 3 kartu (Starter / Pro★ / Bisnis), Pro ditandai "Paling Populer"; Gratis & Enterprise sekunder.
- Limit per tier sesuai matriks 3.2 dienforce di backend.
- Org baru otomatis masuk trial Pro 14 hari; abis itu turun ke Gratis dengan data read-only di atas cap.
- Starter: kuota scan 100/bln dengan opsi top-up Rp49k/100-scan via Pakasir; Pro/Bisnis fair-use cap.
- Reminder perpanjangan H-7/H-3/H-1 aktif.
