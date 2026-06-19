import { error } from "@sveltejs/kit";
import { getGuide } from "$lib/content/help/index.js";

// Resolusi panduan area "app". Slug tak dikenal → 404 (ditangani SvelteKit).
export function load({ params }) {
  const guide = getGuide("app", params.slug);
  if (!guide) {
    error(404, "Panduan tidak ditemukan");
  }
  return { guide };
}
