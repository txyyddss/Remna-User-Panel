import { existsSync, readFileSync, readdirSync } from 'node:fs'
import { dirname, extname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import ts from 'typescript'

const workspace = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
const failures = []
const ignored = new Set([
  resolve(workspace, 'internal/webui/dist'),
])

function pathLabel(path) {
  return relative(workspace, path).replaceAll('\\', '/') || '.'
}

function isIgnored(path) {
  return [...ignored].some((entry) => path === entry || path.startsWith(`${entry}\\`) || path.startsWith(`${entry}/`))
}

function directories(root) {
  const output = []
  const visit = (directory) => {
    if (isIgnored(directory)) return
    output.push(directory)
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      if (entry.isDirectory() && !entry.name.startsWith('.')) visit(join(directory, entry.name))
    }
  }
  visit(resolve(workspace, root))
  return output
}

function files(root) {
  return directories(root).flatMap((directory) => readdirSync(directory, { withFileTypes: true })
    .filter((entry) => entry.isFile())
    .map((entry) => join(directory, entry.name)))
}

function lineCount(path) {
  return readFileSync(path, 'utf8').split(/\r?\n/).length
}

function checkLines(root, extensions, maximum) {
  for (const path of files(root)) {
    if (extensions.has(extname(path)) && lineCount(path) > maximum) {
      failures.push(`${pathLabel(path)} exceeds ${maximum} lines (${lineCount(path)})`)
    }
  }
}

function checkFolders(root) {
  for (const directory of directories(root)) {
    const directFiles = readdirSync(directory, { withFileTypes: true }).filter((entry) => entry.isFile())
    const readmePath = join(directory, 'README.md')
    if (!existsSync(readmePath)) {
      failures.push(`${pathLabel(directory)} is missing README.md`)
      continue
    }
    if (directFiles.length < 2) failures.push(`${pathLabel(directory)} contains fewer than two files`)
    const readme = readFileSync(readmePath, 'utf8')
    for (const entry of directFiles.filter(({ name }) => name !== 'README.md' && !name.startsWith('.'))) {
      if (!readme.includes(`\`${entry.name}\``)) {
        failures.push(`${pathLabel(readmePath)} does not describe ${entry.name}`)
      }
    }
  }
}

function flatten(value, prefix = '', output = new Map()) {
  if (typeof value === 'string') output.set(prefix, value)
  else if (value && typeof value === 'object' && !Array.isArray(value)) {
    for (const [key, child] of Object.entries(value)) flatten(child, prefix ? `${prefix}.${key}` : key, output)
  }
  return output
}

function placeholders(value) {
  return [...value.matchAll(/\{(\w+)\}/g)].map((match) => match[1]).sort().join(',')
}

function readLocale(locale, domains) {
  const messages = {}
  for (const domain of domains) {
    const path = resolve(workspace, `web/locales/${locale}/${domain}.json`)
    if (!existsSync(path)) failures.push(`${pathLabel(path)} is missing`)
    else Object.assign(messages, JSON.parse(readFileSync(path, 'utf8')))
  }
  return flatten(messages)
}

function checkLocales() {
  const manifest = JSON.parse(readFileSync(resolve(workspace, 'web/locales/manifest.json'), 'utf8'))
  const [primary, ...translations] = manifest.locales
  const expected = readLocale(primary, manifest.domains)
  for (const locale of translations) {
    const actual = readLocale(locale, manifest.domains)
    for (const [key, value] of expected) {
      if (!actual.has(key)) failures.push(`locale ${locale} is missing ${key}`)
      else if (placeholders(value) !== placeholders(actual.get(key))) failures.push(`locale ${locale} has mismatched placeholders at ${key}`)
    }
    for (const key of actual.keys()) if (!expected.has(key)) failures.push(`locale ${locale} has extra key ${key}`)
  }
}

const allowedTechnicalPhrases = new Set(['noopener noreferrer'])

function looksLikeUserCopy(value) {
  const normalized = value.replace(/\s+/g, ' ').trim()
  if (!normalized || allowedTechnicalPhrases.has(normalized)) return false
  return /[\u3400-\u9fff]{2}/u.test(normalized) || /[A-Za-z]{2,}\s+[A-Za-z]{2,}/.test(normalized)
}

function checkScriptCopy(path, source, lineOffset = 0) {
  const sourceFile = ts.createSourceFile(path, source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
  const visit = (node) => {
    let value
    if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) value = node.text
    else if (ts.isTemplateExpression(node)) value = node.head.text + node.templateSpans.map((span) => span.literal.text).join('')
    if (value !== undefined && looksLikeUserCopy(value)) {
      const position = sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile))
      failures.push(`${pathLabel(path)}:${position.line + lineOffset + 1} contains hardcoded script copy: ${value.trim()}`)
    }
    ts.forEachChild(node, visit)
  }
  visit(sourceFile)
}

function checkFrontendPolicy() {
  const forbidden = /@phosphor-icons\/vue|from\s+['"]reka-ui['"]|<(?:button|input|select|textarea|table)\b/i
  const visibleText = />\s*([^<>{}\n]*[A-Za-z\u3400-\u9fff][^<>{}\n]*)\s*</g
  const staticCopy = /\s(?:placeholder|title|aria-label|alt)=(?![:])['"][^'"]*[A-Za-z\u3400-\u9fff][^'"]*['"]/i
  for (const path of files('web/src')) {
    const extension = extname(path).toLowerCase()
    const iconName = /(?:^|[-_.])(icon|logo|flag|mark|glyph|symbol)(?:[-_.]|$)/i.test(pathLabel(path))
    if (['.svg', '.ico'].includes(extension) || (iconName && ['.png', '.webp', '.jpg', '.jpeg', '.gif', '.avif'].includes(extension))) {
      failures.push(`${pathLabel(path)} is a local icon asset`)
    }
    if (!['.vue', '.ts'].includes(extname(path))) continue
    const source = readFileSync(path, 'utf8')
    if (forbidden.test(source)) failures.push(`${pathLabel(path)} uses a forbidden native control or direct icon library`)
    if (extname(path) === '.ts' && !path.endsWith('.test.ts') && !path.endsWith('generated.ts')) checkScriptCopy(path, source)
    if (extname(path) === '.vue') {
      const template = source.match(/<template>([\s\S]*?)<\/template>/)?.[1] ?? ''
      for (const match of template.matchAll(visibleText)) {
        if (match[1].trim()) failures.push(`${pathLabel(path)} contains hardcoded visible text: ${match[1].trim()}`)
      }
      if (staticCopy.test(template)) failures.push(`${pathLabel(path)} contains a hardcoded accessibility or form label`)
      for (const match of source.matchAll(/<script\b[^>]*>([\s\S]*?)<\/script>/g)) {
        const scriptStart = (match.index ?? 0) + match[0].indexOf(match[1])
        const lineOffset = source.slice(0, scriptStart).split(/\r?\n/).length - 1
        checkScriptCopy(path, match[1], lineOffset)
      }
    }
  }
  const indexPath = resolve(workspace, 'web/index.html')
  const title = readFileSync(indexPath, 'utf8').match(/<title>([\s\S]*?)<\/title>/i)?.[1].trim()
  if (title) failures.push(`${pathLabel(indexPath)} contains a hardcoded document title: ${title}`)
}

checkLines('web/src', new Set(['.vue', '.ts', '.css']), 200)
checkLines('web/locales', new Set(['.json']), 200)
checkLines('web/scripts', new Set(['.mjs']), 200)
checkLines('internal', new Set(['.go']), 300)
checkLines('cmd', new Set(['.go']), 300)
checkLines('api', new Set(['.yaml', '.yml']), 350)
for (const root of ['web/src', 'web/locales', 'web/scripts', 'internal', 'cmd', 'api']) checkFolders(root)
checkLocales()
checkFrontendPolicy()

if (failures.length) {
  globalThis.console.error(failures.map((failure) => `- ${failure}`).join('\n'))
  globalThis.process.exitCode = 1
} else {
  globalThis.console.log('Structure, localization, and icon policy checks passed.')
}
