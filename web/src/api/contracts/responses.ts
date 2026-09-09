import { fromJSONSchema, type ZodType } from 'zod'
import generated from './responseSchemas.json'

const operations = generated.operations.map(operation => ({
  ...operation,
  pattern: new RegExp(`^${operation.path.replace(/\{[^}]+\}/g, '[^/]+')}$`),
})).sort((a, b) => (a.path.match(/\{/g)?.length ?? 0) - (b.path.match(/\{/g)?.length ?? 0))
const validators = new Map<string, ZodType>()

/** Check shapes against the current API contract, never against an older snapshot. */
export function isResponseCompatible(path: string, method: string, value: unknown): boolean {
  const operation = operations.find(candidate => candidate.method === method && candidate.pattern.test(path))
  if (!operation) return true
  const key = `${method}:${operation.path}`
  let validator = validators.get(key)
  if (!validator) {
    validator = fromJSONSchema({ ...operation.schema, $defs: generated.definitions } as unknown as Parameters<typeof fromJSONSchema>[0])
    validators.set(key, validator)
  }
  return validator.safeParse(value).success
}
