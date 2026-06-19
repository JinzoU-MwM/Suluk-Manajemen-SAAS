import { error } from "@sveltejs/kit";
import { getGuide } from "$lib/content/help/index.js";

// Resolusi panduan area "agency". Slug tak dikenal → 404.
export function load({ params }) {
  const guide = getGuide("agency", params.slug);
  if (!guide) {
    error(404, "Panduan tidak ditemukan");
  }
  return { guide };
}
