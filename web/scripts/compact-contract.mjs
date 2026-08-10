import { readFileSync, writeFileSync } from 'node:fs'

const path = new globalThis.URL('../src/api/generated.ts', import.meta.url)
const source = readFileSync(path, 'utf8')
const compact = source
  .replace(/\/\*[\s\S]*?\*\//g, '')
  .replace(/^\s*\/\/.*$/gm, '')
  .replace(/\s+/g, ' ')
  .trim()

writeFileSync(path, `${compact}\n`)
