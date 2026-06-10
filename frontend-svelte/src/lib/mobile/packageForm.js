// Shared package create/edit form schema + payload normalizer (Paket + PaketDetail).

export const PACKAGE_FIELDS = [
  { key: "name", label: "Nama Paket", required: true, placeholder: "Umrah Reguler 9 Hari" },
  { key: "package_type", label: "Jenis", type: "select", options: [{ value: "umrah", label: "Umrah" }, { value: "haji", label: "Haji" }] },
  { key: "departure_date", label: "Tanggal Berangkat", type: "date" },
  { key: "duration_days", label: "Durasi (hari)", type: "number" },
  { key: "total_seats", label: "Kuota Kursi", type: "number" },
  { key: "airline", label: "Maskapai" },
  { key: "hotel_makkah_name", label: "Hotel Mekkah" },
  { key: "hotel_madinah_name", label: "Hotel Madinah" },
];

export function packagePayload(data) {
  const out = { ...data };
  if (out.duration_days !== "" && out.duration_days != null) out.duration_days = Number(out.duration_days) || 0;
  if (out.total_seats !== "" && out.total_seats != null) out.total_seats = Number(out.total_seats) || 0;
  // drop empty optionals so the backend keeps its defaults
  for (const k of Object.keys(out)) if (out[k] === "" || out[k] == null) delete out[k];
  return out;
}
