// Suluk Mobile — shared formatters & helpers (ported from the design's data.jsx kit)

export const fmtRp = (n) => "Rp " + Math.round(Number(n) || 0).toLocaleString("id-ID");

export const fmtRpShort = (n) => {
  n = Number(n) || 0;
  if (n >= 1e9) return "Rp " + (n / 1e9).toFixed(2).replace(/\.?0+$/, "") + " M";
  if (n >= 1e6) return "Rp " + (n / 1e6).toFixed(1).replace(/\.0$/, "") + " jt";
  if (n >= 1e3) return "Rp " + Math.round(n / 1e3) + " rb";
  return "Rp " + n;
};

// status label -> [foreground, background]
export const STATUS_TONE = {
  Lunas: ["#1B7F5A", "#E8F4EF"], Dibayar: ["#1B7F5A", "#E8F4EF"], Ditandatangani: ["#1B7F5A", "#E8F4EF"],
  Published: ["#1B7F5A", "#E8F4EF"], Aktif: ["#1B7F5A", "#E8F4EF"], Aman: ["#1B7F5A", "#E8F4EF"], Disetujui: ["#1B7F5A", "#E8F4EF"],
  Cicilan: ["#2563a8", "#e6eef8"], Sebagian: ["#2563a8", "#e6eef8"], Terkirim: ["#2563a8", "#e6eef8"], Diproses: ["#2563a8", "#e6eef8"],
  DP: ["#b8860b", "#fbf0d8"], Verifikasi: ["#b8860b", "#fbf0d8"], "Menunggu TTD": ["#b8860b", "#fbf0d8"], Pending: ["#b8860b", "#fbf0d8"], Menipis: ["#b8860b", "#fbf0d8"], Menunggu: ["#b8860b", "#fbf0d8"],
  "Belum Bayar": ["#c0392b", "#fbe9e7"], "Jatuh Tempo": ["#c0392b", "#fbe9e7"], Kritis: ["#c0392b", "#fbe9e7"], Ditolak: ["#c0392b", "#fbe9e7"],
  Draft: ["#6b7d77", "#eef2f0"], Nonaktif: ["#6b7d77", "#eef2f0"],
};
export const statusTone = (s) => STATUS_TONE[s] || ["#6b7d77", "#eef2f0"];

const AV_COLORS = ["#1B7F5A", "#C99A2E", "#2563a8", "#a9842f", "#15564a", "#7a5ae0", "#b8860b"];
export function avatarColor(name) {
  const s = name || "";
  let hash = 0;
  for (let i = 0; i < s.length; i++) hash = s.charCodeAt(i) + ((hash << 5) - hash);
  return AV_COLORS[Math.abs(hash) % AV_COLORS.length];
}
export function initials(name) {
  return (name || "?")
    .split(" ")
    .filter((w) => !/^(H\.|Hj\.)$/.test(w))
    .slice(0, 2)
    .map((w) => w[0])
    .join("")
    .toUpperCase();
}
