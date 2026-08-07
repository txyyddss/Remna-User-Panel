import { fileURLToPath, URL } from 'node:url'
import { writeFile } from 'node:fs/promises'

import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    {
      name: 'preserve-go-embed-placeholder',
      apply: 'build',
      closeBundle: () => writeFile(
        fileURLToPath(new URL('../internal/webui/dist/.placeholder', import.meta.url)),
        'Frontend assets are generated here by the Vite production build.\n',
      ),
    },
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    outDir: '../internal/webui/dist',
    emptyOutDir: true,
    sourcemap: false,
    assetsDir: 'assets',
  },
  server: {
    proxy: {
      '/api': 'http://127.0.0.1:8080',
      '/healthz': 'http://127.0.0.1:8080',
      '/readyz': 'http://127.0.0.1:8080',
    },
  },
  test: {
    environment: 'happy-dom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/composables/**/*.ts', 'src/router/**/*.ts', 'src/stores/**/*.ts', 'src/utils/**/*.ts'],
    },
  },
})
