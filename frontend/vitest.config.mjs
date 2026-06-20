import path from "node:path";
import { fileURLToPath } from "node:url";

import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

const root = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      $lib: path.resolve(root, "src/lib"),
      $components: path.resolve(root, "src/lib/components"),
    },
    conditions: ["browser"],
  },
  test: {
    environment: "happy-dom",
    fileParallelism: false,
    maxWorkers: 1,
    pool: "forks",
    restoreMocks: true,
  },
});
