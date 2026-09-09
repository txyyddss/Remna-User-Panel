import { readFileSync, writeFileSync } from 'node:fs'
import { log } from 'node:console'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parse } from 'yaml'
import { fromJSONSchema } from 'zod'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../../api/openapi.yaml')
const documents = new Map()
const definitions = {}
const references = new Map()
const load = file => {
  if (!documents.has(file)) documents.set(file, parse(readFileSync(file, 'utf8'), { merge: true }))
  return documents.get(file)
}
function dereference(value, file) {
  while (value?.$ref) {
    const [path, fragment = ''] = value.$ref.split('#')
    file = path ? resolve(dirname(file), path) : file
    value = fragment.split('/').filter(Boolean).reduce((current, key) => current[key.replaceAll('~1', '/').replaceAll('~0', '~')], load(file))
  }
  return [value, file]
}
function schema(value, file) {
  if (typeof value === 'boolean') return value
  if (value.$ref) {
    const key = `${file}:${value.$ref}`
    if (!references.has(key)) {
      const name = `s${references.size}`
      references.set(key, name)
      const [target, targetFile] = dereference(value, file)
      if (!target) throw new Error(`Unresolved schema ${key}`)
      definitions[name] = schema(target, targetFile)
    }
    return { $ref: `#/$defs/${references.get(key)}` }
  }
  const output = {}
  for (const key of ['type', 'required', 'enum', 'const']) if (key in value) output[key] = value[key]
  if (value.properties) output.properties = Object.fromEntries(Object.entries(value.properties).map(([key, child]) => [key, schema(child, file)]))
  if (value.items) output.items = schema(value.items, file)
  for (const key of ['oneOf', 'anyOf', 'allOf']) if (value[key]) output[key] = value[key].map(child => schema(child, file))
  // New optional fields are compatible; value restrictions remain server-owned.
  if (value.type === 'object' || value.properties) output.additionalProperties = typeof value.additionalProperties === 'object' ? schema(value.additionalProperties, file) : true
  return output
}
const operations = []
for (const [path, reference] of Object.entries(load(root).paths)) {
  const [item, file] = dereference(reference, root)
  for (const method of ['get', 'post']) {
    const response = Object.entries(item[method]?.responses ?? {}).find(([status]) => status.startsWith('2'))?.[1]
    if (!response) continue
    const [resolved, responseFile] = dereference(response, file)
    const body = resolved.content?.['application/json']?.schema
    if (body) operations.push({ method: method.toUpperCase(), path, schema: schema(body, responseFile) })
  }
}
for (const operation of operations) fromJSONSchema({ ...operation.schema, $defs: definitions })
const output = resolve(dirname(fileURLToPath(import.meta.url)), '../src/api/contracts/responseSchemas.json')
writeFileSync(output, JSON.stringify({ definitions, operations }) + '\n')
log(`Generated ${operations.length} response shape contracts.`)
