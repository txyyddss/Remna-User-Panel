import { createUrl, type RequestOptions } from '../http'
import { responseCacheKey } from './policy'
import { readResponseCache } from './session'
import type { ShallowRef } from 'vue'

// Only rendering loaders opt into snapshots; request() still resolves with live data.
export function restoreCached<T>(path: string, apply: (value: T) => void, options: RequestOptions = {}): boolean {
  const key = responseCacheKey(createUrl(path, options.query), options)
  if (!key) return false
  const value = readResponseCache<T>(key)
  if (value === undefined) return false
  apply(value)
  return true
}

export function restoreRef<T>(path: string, target: ShallowRef<T>, options: RequestOptions = {}): boolean {
  return restoreCached<T>(path, value => { target.value = value }, options)
}

export function restoreItems<T>(path: string, target: ShallowRef<T[]>, options: RequestOptions = {}): boolean {
  return restoreCached<{ items: T[] }>(path, value => { target.value = value.items }, options)
}
