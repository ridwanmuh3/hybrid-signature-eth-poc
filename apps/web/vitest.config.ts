import { defineConfig, mergeConfig } from "vitest/config";
import viteConfig from "./vite.config";
import { playwright } from "@vitest/browser-playwright";

export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      browser: {
        provider: playwright(),
        enabled: true,
        headless: true,
      },
    },
  })
);
