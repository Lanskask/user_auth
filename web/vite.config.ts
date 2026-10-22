/// <reference types="vitest/config" />
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// The Go server embeds ./static, so that is where the production build goes.
export default defineConfig({
  plugins: [react()],
  build: { outDir: "static", emptyOutDir: false },
  server: {
    // Same-origin in development: the browser only ever talks to Vite, which forwards /api to the Go server.
    proxy: { "/api": "http://localhost:8080" },
  },
  test: {
    environment: "jsdom",
    globals: true,
    include: ["src/**/*.test.{ts,tsx}"],
    setupFiles: ["src/test/setup.ts"],
  },
});
