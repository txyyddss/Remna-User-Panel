import { nextTick, onScopeDispose } from 'vue'
import type { NavigationMenuItem } from '@nuxt/ui'
import type { Router } from 'vue-router'

function navigationPaths(items: NavigationMenuItem[]): string[] {
  return items.flatMap(item => [
    ...(typeof item.to === 'string' ? [item.to] : []),
    ...navigationPaths(item.children ?? []),
  ])
}

/** Slide a captured viewport so outgoing scroll and fixed navigation stay stable. */
export function usePageTransition(router: Router, items: () => NavigationMenuItem[]): void {
  let pending: ViewTransition | undefined
  let releaseRender: (() => void) | undefined
  let destination = ''
  let version = 0
  let disposed = false

  const removeGuard = router.beforeResolve(async (to, from) => {
    const request = ++version
    await pending?.finished.catch(() => undefined)
    if (disposed || request !== version) return false
    if (to.path === from.path || to.meta.immersive || from.meta.immersive
      || !document.startViewTransition || window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

    const paths = navigationPaths(items())
    const rank = (path: string) => paths.indexOf(path.startsWith('/admin/users/') ? '/admin/users' : path)
    const backwards = rank(to.path) >= 0 && rank(to.path) < rank(from.path)
    document.documentElement.dataset.pageSlide = backwards ? 'backward' : 'forward'
    destination = to.fullPath
    const rendered = new Promise<void>(resolve => { releaseRender = resolve })
    await new Promise<void>(resolve => {
      pending = document.startViewTransition(async () => {
        resolve()
        await rendered
      })
      // Hidden documents and unsupported snapshot states still complete navigation.
      void pending.ready.catch(resolve)
      void pending.finished.catch(() => undefined).then(() => {
        delete document.documentElement.dataset.pageSlide
      })
    })
  })

  const removeAfter = router.afterEach((to, _from, failure) => {
    if (to.fullPath !== destination) return
    if (failure) pending?.skipTransition()
    void nextTick().then(() => releaseRender?.())
  })
  const removeError = router.onError(() => {
    releaseRender?.()
    pending?.skipTransition()
  })
  onScopeDispose(() => {
    disposed = true
    removeGuard()
    removeAfter()
    removeError()
    releaseRender?.()
    pending?.skipTransition()
    delete document.documentElement.dataset.pageSlide
  })
}
