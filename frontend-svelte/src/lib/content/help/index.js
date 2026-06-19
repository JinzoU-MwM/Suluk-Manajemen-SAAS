// Pusat Bantuan — indeks konten & helper.
//
// Isolasi peran: setiap helper menerima `area` dan HANYA membaca array area itu.
// /portal tidak akan pernah menampilkan panduan /app, dst.
//
// Catatan rendering: `body` disimpan sebagai blok terstruktur (HelpBlock[]), bukan
// string markdown — repo belum memakai mdsvex, jadi blok dirender oleh komponen
// Svelte (GuideBody.svelte). Tidak ada HTML mentah, sehingga aman tanpa sanitasi.

import { APP_GUIDES } from "./app.js";
import { PORTAL_GUIDES } from "./portal.js";
import { AGENCY_GUIDES } from "./agency.js";

/**
 * @typedef {"app" | "portal" | "agency"} HelpArea
 */

/**
 * Satu blok konten di dalam body sebuah panduan.
 * @typedef {Object} HelpBlock
 * @property {"p" | "h2" | "ul" | "ol" | "callout"} type - Jenis blok.
 * @property {string}   [text]    - Teks untuk p / h2 / callout.
 * @property {string[]} [items]   - Butir untuk ul / ol.
 * @property {"info" | "tip" | "warning"} [variant] - Gaya untuk callout.
 */

/**
 * @typedef {Object} HelpGuide
 * @property {string}      slug      - URL-safe, unik dalam satu area.
 * @property {string}      title     - Judul (Bahasa Indonesia).
 * @property {string}      category  - Kategori untuk pengelompokan (mis. "Memulai").
 * @property {string}      summary   - 1–2 kalimat ringkasan untuk daftar & pencarian.
 * @property {string[]}    keywords  - Kata kunci tambahan untuk pencarian (sinonim).
 * @property {number}      [order]   - Urutan dalam kategori (kecil = atas).
 * @property {HelpBlock[]} body      - Isi panduan sebagai blok terstruktur.
 * @property {string[]}    [related] - slug panduan terkait (dalam area yang sama).
 * @property {string}      [updatedAt] - Tanggal ISO, opsional.
 */

/** @type {Record<HelpArea, HelpGuide[]>} */
const AREAS = {
  app: APP_GUIDES,
  portal: PORTAL_GUIDES,
  agency: AGENCY_GUIDES,
};

/** Daftar kunci area yang valid. @type {HelpArea[]} */
export const AREA_KEYS = ["app", "portal", "agency"];

/**
 * Ambil array mentah satu area (selalu array, kosong bila area tak dikenal).
 * @param {HelpArea} area
 * @returns {HelpGuide[]}
 */
function areaGuides(area) {
  return AREAS[area] ?? [];
}

/** Normalisasi string untuk pencarian: string aman, di-trim, huruf kecil. */
function normalize(value) {
  return (value ?? "").toString().trim().toLowerCase();
}

/** Gabungkan teks seluruh blok body menjadi satu string yang bisa dicari. */
function bodyText(body) {
  return (body ?? [])
    .map((block) => [block.text, ...(block.items ?? [])].filter(Boolean).join(" "))
    .join(" ");
}

/**
 * Seluruh panduan satu area, terurut berdasarkan kategori lalu `order` lalu judul.
 * @param {HelpArea} area
 * @returns {HelpGuide[]}
 */
export function getGuides(area) {
  return [...areaGuides(area)].sort(
    (a, b) =>
      a.category.localeCompare(b.category, "id") ||
      (a.order ?? 0) - (b.order ?? 0) ||
      a.title.localeCompare(b.title, "id"),
  );
}

/**
 * Panduan satu area, dikelompokkan per kategori (urutan kategori & isi terjaga).
 * @param {HelpArea} area
 * @returns {Record<string, HelpGuide[]>}
 */
export function getCategories(area) {
  /** @type {Record<string, HelpGuide[]>} */
  const grouped = {};
  for (const guide of getGuides(area)) {
    (grouped[guide.category] ??= []).push(guide);
  }
  return grouped;
}

/**
 * Cari satu panduan berdasarkan slug, hanya di dalam area tersebut.
 * @param {HelpArea} area
 * @param {string} slug
 * @returns {HelpGuide | undefined}
 */
export function getGuide(area, slug) {
  return areaGuides(area).find((guide) => guide.slug === slug);
}

/**
 * Pencarian dalam satu area. Cocokkan setiap kata (AND) pada title/keywords/
 * summary/body, lalu urutkan berdasarkan bobot bidang (judul & kata kunci di atas
 * isi). Query kosong → seluruh panduan terurut (lihat getGuides).
 * @param {HelpArea} area
 * @param {string} query
 * @returns {HelpGuide[]}
 */
export function searchGuides(area, query) {
  const q = normalize(query);
  if (!q) return getGuides(area);

  const terms = q.split(/\s+/).filter(Boolean);

  /** @type {{ guide: HelpGuide, score: number }[]} */
  const scored = [];

  for (const guide of areaGuides(area)) {
    const fields = {
      title: normalize(guide.title),
      keywords: normalize((guide.keywords ?? []).join(" ")),
      summary: normalize(guide.summary),
      body: normalize(bodyText(guide.body)),
    };

    let score = 0;
    let matchedAll = true;

    for (const term of terms) {
      let weight = 0;
      if (fields.title.includes(term)) weight = 10;
      else if (fields.keywords.includes(term)) weight = 6;
      else if (fields.summary.includes(term)) weight = 3;
      else if (fields.body.includes(term)) weight = 1;

      if (weight === 0) {
        matchedAll = false;
        break;
      }
      score += weight;
    }

    if (matchedAll) scored.push({ guide, score });
  }

  return scored
    .sort((a, b) => b.score - a.score || a.guide.title.localeCompare(b.guide.title, "id"))
    .map((entry) => entry.guide);
}

// Pemetaan segmen rute → slug panduan, agar tombol bantuan kontekstual ("?") di
// setiap halaman dapat menautkan ke panduan yang relevan. Kunci "" = halaman
// indeks area. Modul tanpa panduan khusus sengaja tidak didaftarkan (tombolnya
// akan menaut ke daftar Pusat Bantuan).
/** @type {Record<HelpArea, Record<string, string>>} */
export const MODULE_GUIDE = {
  app: {
    "": "mengenal-dashboard",
    jamaah: "mengelola-data-jamaah",
    scanner: "scan-dokumen-ai",
    packages: "membuat-paket-umrah",
    crm: "crm-pipeline-jamaah",
    agents: "kelola-agen-mitra",
    contracts: "e-kontrak-jamaah",
    grup: "membuat-grup-keberangkatan",
    visa: "kelola-visa-dokumen",
    itinerary: "menyusun-itinerary",
    pembimbing: "kelola-pembimbing",
    rooming: "menyusun-manifest-rooming",
    manifest: "menyusun-manifest-rooming",
    invoices: "invoice-dan-pembayaran",
    kasir: "invoice-dan-pembayaran",
    finance: "laporan-keuangan",
    cancellation: "pembatalan-refund",
    tabungan: "tabungan-umrah-jamaah",
    akuntansi: "akuntansi-pembukuan",
    payroll: "payroll-penggajian",
    vendors: "kelola-vendor",
    inventory: "inventaris-perlengkapan",
    stock: "inventaris-perlengkapan",
    documents: "kelola-dokumen-jamaah",
    analytics: "analitik-statistik",
    export: "ekspor-laporan",
    team: "tim-organisasi",
  },
  portal: {
    "": "masuk-dan-beranda",
    dokumen: "mengunggah-dokumen",
    visa: "memantau-status-visa",
    profil: "melengkapi-profil",
  },
  agency: {
    "": "mengenal-dashboard-agen",
    leads: "mengelola-leads",
    komisi: "melihat-komisi",
    jaringan: "memantau-jaringan",
    profil: "profil-agen",
  },
};

/**
 * Slug panduan untuk satu segmen rute area (mis. "jamaah" pada /app/jamaah).
 * Mengembalikan undefined bila tidak ada pemetaan atau panduannya tidak ada.
 * @param {HelpArea} area
 * @param {string} segment - segmen pertama setelah prefix area; "" untuk indeks.
 * @returns {string | undefined}
 */
export function getGuideSlugForRoute(area, segment) {
  const slug = MODULE_GUIDE[area]?.[segment ?? ""];
  return slug && getGuide(area, slug) ? slug : undefined;
}
