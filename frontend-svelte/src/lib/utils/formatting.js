/**
 * Centralized Indonesian locale formatting helpers.
 *
 * These were previously duplicated across ~17 files; import from here instead.
 * Canonical implementations match the originals in Dashboard.svelte.
 */

const rupiahFmt = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
});

const numberFmt = new Intl.NumberFormat('id-ID');

/** Format a number as Rupiah, e.g. 1500000 -> "Rp1.500.000". Null-safe. */
export function formatRupiah(value) {
    if (value == null || isNaN(value)) return 'Rp0';
    return rupiahFmt.format(value);
}

/** Group-separated number, e.g. 1500000 -> "1.500.000". Null-safe. */
export function formatNumber(value) {
    if (value == null || isNaN(value)) return '0';
    return numberFmt.format(value);
}

/** Short date, e.g. "8 Jun 2026". Returns "-" for empty input. */
export function formatDate(value) {
    if (!value) return '-';
    const d = new Date(value);
    if (isNaN(d.getTime())) return '-';
    return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
}

/** Short date + time, e.g. "08/06/2026 14:30". Returns "-" for empty input. */
export function formatDateTime(value) {
    if (!value) return '-';
    const d = new Date(value);
    if (isNaN(d.getTime())) return '-';
    return d.toLocaleString('id-ID', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    });
}

/** Clamp a percentage to a 0-100 integer. Null-safe (returns 0). */
export function formatPct(value) {
    if (value == null || isNaN(value)) return 0;
    return Math.round(Math.min(100, Math.max(0, value)));
}
