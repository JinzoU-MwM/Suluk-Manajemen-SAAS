import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),

  kit: {
    // Static output so the existing Dockerfile.frontend (copies /app/dist -> nginx)
    // and Capacitor (webDir: "dist") keep working unchanged.
    adapter: adapter({
      pages: "dist",
      assets: "dist",
      // SPA fallback: any non-prerendered route (the authed app, mobile shell,
      // public token deep-links) is served this shell and client-routed.
      fallback: "200.html",
      precompress: false,
      strict: false,
    }),
    // Static assets live in public/ (sw.js, manifest.json, brand/, icons, the
    // legacy static landing .html, sitemap/robots) — keep that dir as-is.
    files: {
      assets: "public",
    },
    prerender: {
      // Marketing pages link to the SPA app (/login, /app...) which is ssr=false;
      // don't fail the prerender crawl when it can't render those.
      handleHttpError: "warn",
      handleMissingId: "warn",
    },
  },
};

export default config;
