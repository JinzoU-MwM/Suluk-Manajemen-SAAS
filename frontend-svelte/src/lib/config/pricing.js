// Single source of truth for subscription tiers on the frontend.
// Mirrors internal/shared/plan/plan.go — keep the two in sync.

export const UNLIMITED = -1;

// Scan top-up SKU (mirror of plan.go ScanTopupPrice/ScanTopupScans).
export const SCAN_TOPUP_PRICE = 49000;
export const SCAN_TOPUP_SCANS = 100;

/** Ordered tier catalog (low → high). Prices in IDR (no decimals). */
export const PLANS = [
    {
        key: 'gratis',
        name: 'Gratis',
        rank: 0,
        monthlyPrice: 0,
        annualPrice: 0,
        priceLabel: 'Gratis',
        annualLabel: 'Gratis',
        maxJamaah: 50,
        maxGroups: 2,
        maxUsers: 1,
        maxScansPerMonth: 5,
        purchasable: false,
        popular: false,
        desc: 'Untuk travel baru yang ingin mulai rapi.',
        cta: 'Mulai Gratis',
        features: ['Hingga 50 jamaah', 'Data jamaah & grup', 'Manajemen paket dasar', '1 pengguna'],
    },
    {
        key: 'starter',
        name: 'Starter',
        rank: 1,
        monthlyPrice: 149000,
        annualPrice: 1490000,
        priceLabel: 'Rp 149rb',
        annualLabel: 'Rp 1,49jt',
        maxJamaah: 500,
        maxGroups: 10,
        maxUsers: 3,
        maxScansPerMonth: 100,
        purchasable: true,
        popular: false,
        desc: 'Untuk travel kecil yang mau mulai rapi.',
        cta: 'Pilih Starter',
        features: ['Hingga 500 jamaah', 'Pembayaran & cicilan jamaah', 'AI Scanner 100 dok/bulan', 'Hingga 3 pengguna', 'Laporan dasar'],
    },
    {
        key: 'pro',
        name: 'Pro',
        rank: 2,
        monthlyPrice: 299000,
        annualPrice: 2990000,
        priceLabel: 'Rp 299rb',
        annualLabel: 'Rp 2,99jt',
        maxJamaah: UNLIMITED,
        maxGroups: UNLIMITED,
        maxUsers: 10,
        maxScansPerMonth: UNLIMITED,
        purchasable: true,
        popular: true,
        desc: 'Paling lengkap untuk operasional penuh.',
        cta: 'Coba Pro 14 Hari',
        features: ['Jamaah tak terbatas', 'Semua modul (CRM, Keuangan, Kontrak)', 'AI Scanner tanpa batas', 'Hingga 10 pengguna', 'Laporan & ekspor lanjutan'],
    },
    {
        key: 'bisnis',
        name: 'Bisnis',
        rank: 3,
        monthlyPrice: 599000,
        annualPrice: 5990000,
        priceLabel: 'Rp 599rb',
        annualLabel: 'Rp 5,99jt',
        maxJamaah: UNLIMITED,
        maxGroups: UNLIMITED,
        maxUsers: 25,
        maxScansPerMonth: UNLIMITED,
        purchasable: true,
        popular: false,
        desc: 'Untuk travel besar & multi-cabang.',
        cta: 'Pilih Bisnis',
        features: ['Semua fitur Pro', 'Multi-cabang & multi-PT', 'Hingga 25 pengguna', 'Dukungan prioritas'],
    },
    {
        key: 'enterprise',
        name: 'Enterprise',
        rank: 4,
        monthlyPrice: 0,
        annualPrice: 0,
        priceLabel: 'Custom',
        annualLabel: 'Custom',
        maxJamaah: UNLIMITED,
        maxGroups: UNLIMITED,
        maxUsers: UNLIMITED,
        maxScansPerMonth: UNLIMITED,
        purchasable: false,
        popular: false,
        desc: 'Solusi khusus untuk grup usaha besar.',
        cta: 'Hubungi Sales',
        features: ['Semua fitur Bisnis', 'Akses API & integrasi', 'Pengguna tak terbatas', 'Dukungan prioritas 24/7'],
    },
];

/** Rank lookup, including legacy aliases (free → gratis, business → bisnis). */
export const TIER_RANK = {
    gratis: 0,
    free: 0,
    starter: 1,
    pro: 2,
    bisnis: 3,
    business: 3,
    enterprise: 4,
};

const PRO_RANK = TIER_RANK.pro;

/** Normalize a plan string to a current tier key. */
export function normalizePlan(plan) {
    const k = (plan || '').toString().trim().toLowerCase();
    if (k === 'free') return 'gratis';
    if (k === 'business') return 'bisnis';
    return PLANS.some((p) => p.key === k) ? k : 'gratis';
}

/** Numeric rank for comparisons; unknown → 0 (gratis). */
export function rankOf(plan) {
    const k = (plan || '').toString().trim().toLowerCase();
    return TIER_RANK[k] ?? 0;
}

/** True when the plan unlocks the advanced modules (pro and above). */
export function isProOrHigher(plan) {
    return rankOf(plan) >= PRO_RANK;
}

/** True when key's tier is at least min's tier. */
export function atLeast(plan, min) {
    return rankOf(plan) >= rankOf(min);
}

/** Tier metadata for a plan key (normalized); always returns a tier. */
export function planMeta(plan) {
    const k = normalizePlan(plan);
    return PLANS.find((p) => p.key === k) || PLANS[0];
}

/** "Rp 1.234.567" */
export function formatIDR(n) {
    return 'Rp ' + Number(n || 0).toLocaleString('id-ID');
}

/** Human limit label: unlimited → "tak terbatas". */
export function limitLabel(n) {
    return n === UNLIMITED ? 'tak terbatas' : String(n);
}
