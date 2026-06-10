import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";

// https://vite.dev/config/
export default defineConfig({
  plugins: [sveltekit()],
  test: {
    environment: "node",
    include: ["src/**/*.test.js"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html"],
    },
  },
  server: {
    host: true,
    proxy: {
      "/api/packages": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, "/api/v1"),
      },
      "/api/contracts": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, "/api/v1"),
      },
      "/public/contracts": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
      },
      "/api": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, "/api/v1"),
      },
      "/public/packages": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true,
      },
    },
  },
});
