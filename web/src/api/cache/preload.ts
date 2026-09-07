import type { Session } from '../types'
import type { DatabaseTable } from '../features'
import { request } from '../http'
import { getLocale } from '@/i18n'
import { adminResources, memberResources, type PreloadResource } from './preloadRoutes'

let active: AbortController | null = null

export function stopSessionPreload(): void {
  active?.abort()
  active = null
}

export function preloadSession(session: Session): void {
  stopSessionPreload()
  const controller = new AbortController()
  active = controller
  const queue = memberResources(session, getLocale())
  if (session.user.role === 'admin') queue.push(...adminResources())
  const seen = new Set<string>()

  async function load(resource: PreloadResource): Promise<void> {
    const identity = JSON.stringify(resource)
    if (seen.has(identity)) return
    seen.add(identity)
    const timeout = new AbortController()
    const abort = () => timeout.abort()
    controller.signal.addEventListener('abort', abort, { once: true })
    const timer = globalThis.setTimeout(abort, 20_000)
    try {
      const value = await request(resource.path, { ...resource.options, signal: timeout.signal })
      if (resource.path === '/api/v1/admin/database/tables') {
        const first = (value as { items: DatabaseTable[] }).items[0]
        if (first) queue.push({
          path: `/api/v1/admin/database/tables/${encodeURIComponent(first.name)}/query`,
          options: { method: 'POST', body: { filters: [], limit: 50 } },
        })
      }
    } catch (error) {
      if ((error as { status?: number } | null)?.status === 401) controller.abort()
      // Optional background reads never replace the foreground page's error handling.
    } finally {
      globalThis.clearTimeout(timer)
      controller.signal.removeEventListener('abort', abort)
    }
  }

  async function worker(): Promise<void> {
    while (!controller.signal.aborted && active === controller) {
      const resource = queue.shift()
      if (!resource) return
      await load(resource)
    }
  }

  // Start after bootstrap can mount the requested page. Three reads run at a time.
  globalThis.setTimeout(() => {
    if (active !== controller || controller.signal.aborted) return
    void Promise.allSettled(Array.from({ length: 3 }, worker))
  }, 0)
}
