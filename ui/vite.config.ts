import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  plugins: [tsconfigPaths(), react()],
  base: process.env.VITE_BASE_URL || "/",
  server: {
    port: 3000,
    proxy: {
      "/ui/assets": {
        target: "https://0.0.0.0:8414/",
        rewrite: (path) => path.replace(/^\/ui/, ""),
        secure: false,
      },
    },
  },
  build: {
    outDir: "./build/ui",
    minify: "esbuild",
  },
  experimental: {
    renderBuiltUrl(filename: string) {
      return "/ui/" + filename;
    },
  },
});
