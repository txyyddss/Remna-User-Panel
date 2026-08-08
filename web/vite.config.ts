import { fileURLToPath, URL } from 'node:url'
import { readFileSync } from 'node:fs'
import { writeFile } from 'node:fs/promises'

import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

function localeLeafKeys(value: unknown, prefix = ''): string[] {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return prefix ? [prefix] : []
  return Object.entries(value).flatMap(([key, child]) => localeLeafKeys(child, prefix ? `${prefix}.${key}` : key))
}

function localeLeaves(value: unknown, prefix = '', output: Record<string, string> = {}): Record<string, string> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    if (typeof value === 'string') output[prefix] = value
    return output
  }
  for (const [key, child] of Object.entries(value)) localeLeaves(child, prefix ? `${prefix}.${key}` : key, output)
  return output
}

function validateLocaleParity(): void {
  const read = (name: string) => JSON.parse(readFileSync(new URL(`./locales/${name}.json`, import.meta.url), 'utf8')) as unknown
  const english = new Set(localeLeafKeys(read('en')))
  const englishLeaves = localeLeaves(read('en'))
  for (const name of ['zh-CN']) {
    const actual = new Set(localeLeafKeys(read(name)))
    const missing = [...english].filter((key) => !actual.has(key))
    const extra = [...actual].filter((key) => !english.has(key))
    if (missing.length || extra.length) throw new Error(`Locale ${name} does not match en.json. Missing: ${missing.join(', ')}. Extra: ${extra.join(', ')}`)
    const translatedLeaves = localeLeaves(read(name))
    for (const key of english) {
      const expected = [...englishLeaves[key].matchAll(/\{(\w+)\}/g)].map((match) => match[1]).sort().join(',')
      const actualPlaceholders = [...translatedLeaves[key].matchAll(/\{(\w+)\}/g)].map((match) => match[1]).sort().join(',')
      if (expected !== actualPlaceholders) throw new Error(`Locale ${name} placeholder mismatch at ${key}`)
    }
  }
}

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    {
      name: 'validate-locale-parity',
      buildStart: validateLocaleParity,
    },
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
    setupFiles: ['./src/test/storage-setup.ts', './src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/composables/**/*.ts', 'src/router/**/*.ts', 'src/stores/**/*.ts', 'src/utils/**/*.ts'],
    },
  },
})
