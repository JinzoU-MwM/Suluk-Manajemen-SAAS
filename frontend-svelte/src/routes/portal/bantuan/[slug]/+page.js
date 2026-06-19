import { error } from "@sveltejs/kit";
import { getGuide } from "$lib/content/help/index.js";

// Resolusi panduan area "portal". Slug tak dikenal → 404.
export function load({ params }) {
  const guide = getGuide("portal", params.slug);
  if (!guide) {
    error(404, "Panduan tidak ditemukan");
  }
  return { guide };
}
