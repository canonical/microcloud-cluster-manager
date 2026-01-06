import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  css: {
    preprocessorOptions: {
      scss: {
        silenceDeprecations: ["global-builtin", "import", "if-function"],
      },
    },
  },
  plugins: [tsconfigPaths(), react()],
  base: process.env.VITE_BASE_URL || "/",
  server: {
    port: 3000,
    proxy: {
      "/ui/assets": {
        target: "https://ma.lxd-cm.local:8414/",
        rewrite: (path) => path.replace(/^\/ui/, ""),
        secure: false,
      },
      // NOTE: the following paths will be directed to the cluster ingress for local development
      // Here we do not want to change the host header for oidc flow
      "/1.0": {
        target: "https://ma.lxd-cm.local:30000/",
        secure: false,
      },
      "/oidc": {
        target: "https://ma.lxd-cm.local:30000/",
        secure: false,
      },
    },
    allowedHosts: ["ma.lxd-cm.local"],
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
