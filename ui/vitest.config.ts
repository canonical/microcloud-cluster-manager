import { mergeConfig, defineConfig } from "vitest/config";
import viteConfig from "./vite.config";

export default mergeConfig(
  viteConfig,
  defineConfig({
    build: {
      sourcemap: "inline",
    },
    test: {
      environment: "jsdom",
      globals: true,
      include: ["./src/**/*.spec.{ts,tsx}"],
      server: {
        deps: {
          inline: ["@canonical/react-ds-global"],
        },
      },
      coverage: {
        provider: "istanbul",
        reportsDirectory: "coverage/unit",
      },
    },
  }),
);
